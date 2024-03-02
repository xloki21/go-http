package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/xloki21/go-http/internal/model"
	"github.com/xloki21/go-http/internal/server"
	"github.com/xloki21/go-http/internal/server/apperrors"
	"io"
	"net/http"
	"testing"
)

func TestProcessRequest(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	api := NewHandlers()
	srv := new(server.Server)

	go func() {
		if err := srv.Run("localhost", "8080", api); err != nil {
			fmt.Println(err)
		}
	}()

	defer func() {
		err := srv.Shutdown(context.Background())
		if err != nil {
			t.Errorf("ProcessRequest() error: %v", err)
		}
	}()

	host, port := "localhost", "8080"
	endpoint := fmt.Sprintf("http://%s:%s%s", host, port, ApiV1)

	type args struct {
		Method             string
		URLList            []model.URL
		TryToCancelRequest bool
	}
	tests := []struct {
		name  string
		args  args
		wants apperrors.AppError
	}{
		{
			name:  "GET Request",
			args:  args{Method: http.MethodGet, URLList: nil},
			wants: apperrors.MethodNotAllowed,
		},
		{
			name:  "POST Request with incorrect data (body=nil)",
			args:  args{Method: http.MethodPost, URLList: nil},
			wants: apperrors.EmptyBodyErr,
		},
		{
			name:  "POST Request with incorrect data (incorrect URL)",
			args:  args{Method: http.MethodPost, URLList: []model.URL{"https://1go.dev"}},
			wants: apperrors.URLNotFoundErr,
		},
		{
			name: "Post Request with incorrect data (size(URLList) > MaxUrlsPerRequest)",
			args: args{Method: http.MethodPost, URLList: func() []model.URL {
				var urls []model.URL
				for i := 0; i < MaxUrlsPerRequest+1; i++ {
					urls = append(urls, "https://go.dev/images/go-logo-white.svg")
				}
				return urls
			}()},
			wants: apperrors.TooBigURLListErr,
		},
		{
			name: "Post Request with correct data",
			args: args{Method: http.MethodPost, URLList: func() []model.URL {
				var urls []model.URL
				for i := 0; i < MaxUrlsPerRequest; i++ {
					urls = append(urls, "https://go.dev/images/go-logo-white.svg")
				}
				return urls
			}()},
			wants: apperrors.NilErr,
		},
		{
			name: "Post Request with processing timeout reached",
			args: args{
				Method: http.MethodPost,
				URLList: []model.URL{
					"http://images.cocodataset.org/zips/train2014.zip",
					"https://go.dev/images/go-logo-white.svg",
				},
			},
			wants: apperrors.TimeoutErr,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body []byte
			if len(tt.args.URLList) > 0 {
				body, _ = json.Marshal(tt.args.URLList)
			}
			request, err := http.NewRequestWithContext(ctx, tt.args.Method, endpoint,
				bytes.NewBuffer(body))

			resp, err := http.DefaultClient.Do(request)

			defer func(Body io.ReadCloser) {
				if err := Body.Close(); err != nil {
					t.Errorf("ProcessRequest() error: %v", err)
				}
			}(resp.Body)

			if err != nil {
				if errors.Is(err, context.Canceled) && !tt.args.TryToCancelRequest {
					t.Errorf("%v", err)
				}
			}
			if resp.StatusCode != tt.wants.Code {
				t.Errorf("ProcessRequest() getStatusCode = %v, wantsStatusCode = %v", resp.StatusCode, tt.wants.Code)
			}
		})
	}
}
