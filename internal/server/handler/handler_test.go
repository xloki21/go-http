package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/xloki21/go-http/internal/model"
	"github.com/xloki21/go-http/internal/server"
	"github.com/xloki21/go-http/internal/server/apperrors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestServerLoadTesting(t *testing.T) {

	wg := sync.WaitGroup{}

	api := NewHandlers()
	srv := new(server.Server)
	host, port := "localhost", "8080"
	go func() {
		if err := srv.Run(host, port, api); err != nil {
			fmt.Println(err)
		}
	}()

	defer func() {
		err := srv.Shutdown(context.Background())
		if err != nil {
			t.Errorf("ProcessRequest() error: %v", err)
		}
	}()

	URLList := func() []model.URL {
		var urls []model.URL
		for i := 0; i < MaxUrlsPerRequest; i++ {
			urls = append(urls, "https://go.dev/images/go-logo-white.svg")
		}
		return urls
	}()

	type args struct {
		NumberOfRequests int
	}
	tests := []struct {
		name    string
		args    args
		watsErr bool
	}{
		{
			name: "10% of loading",
			args: args{NumberOfRequests: 10},
		},
		{
			name: "50% of loading",
			args: args{NumberOfRequests: 50},
		},
		{
			name: "99% of loading",
			args: args{NumberOfRequests: 99},
		},
		{
			name:    "110% of loading",
			args:    args{NumberOfRequests: 110},
			watsErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body []byte
			body, err := json.Marshal(URLList)
			if err != nil {
				t.Errorf("ProcessRequest() error: %v", err)
			}
			for i := 0; i < tt.args.NumberOfRequests; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()

					request := httptest.NewRequest(http.MethodPost, ApiV1, bytes.NewBuffer(body))
					w := httptest.NewRecorder()

					api.ServeHTTP(w, request)
					if w.Code == apperrors.TooManyRequestsErr.Code {
						if !tt.watsErr {
							t.Fail()
							return
						}
					}
				}()
			}
			wg.Wait()
		})
	}
}

func TestProcessRequest(t *testing.T) {
	ctx := context.Background()
	api := NewHandlers()
	type Ctx struct {
		Context  context.Context
		CancelFn context.CancelFunc
	}

	type args struct {
		Method  string
		URLList []model.URL
		Ctx     Ctx
	}
	tests := []struct {
		name  string
		args  args
		wants apperrors.AppError
	}{
		{
			name:  "GET Request",
			args:  args{Method: http.MethodGet, URLList: nil, Ctx: Ctx{ctx, nil}},
			wants: apperrors.MethodNotAllowed,
		},
		{
			name:  "POST Request with incorrect data (body=nil)",
			args:  args{Method: http.MethodPost, URLList: nil, Ctx: Ctx{ctx, nil}},
			wants: apperrors.InvalidBodyErr,
		},
		{
			name:  "POST Request with incorrect data (incorrect URL)",
			args:  args{Method: http.MethodPost, URLList: []model.URL{"https://1go.dev"}, Ctx: Ctx{ctx, nil}},
			wants: apperrors.URLNotFoundErr,
		},
		{
			name: "Post Request with incorrect data (size(URLList) > MaxUrlsPerRequest)",
			args: args{
				Method: http.MethodPost,
				URLList: func() []model.URL {
					var urls []model.URL
					for i := 0; i < MaxUrlsPerRequest+1; i++ {
						urls = append(urls, "https://go.dev/images/go-logo-white.svg")
					}
					return urls
				}(),
				Ctx: Ctx{ctx, nil},
			},
			wants: apperrors.TooBigURLListErr,
		},
		{
			name: "Post Request with correct data",
			args: args{
				Method: http.MethodPost,
				URLList: func() []model.URL {
					var urls []model.URL
					for i := 0; i < MaxUrlsPerRequest; i++ {
						urls = append(urls, "https://go.dev/images/go-logo-white.svg")
					}
					return urls
				}(),
				Ctx: Ctx{ctx, nil},
			},
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
				Ctx: Ctx{ctx, nil},
			},
			wants: apperrors.TimeoutErr,
		},
		{
			name: "Post Request with cancellation",
			args: args{
				Method: http.MethodPost,
				URLList: func() []model.URL {
					var urls []model.URL
					for i := 0; i < MaxUrlsPerRequest; i++ {
						urls = append(urls, "https://go.dev/images/go-logo-white.svg")
					}
					return urls
				}(),
				Ctx: func() Ctx {
					ctxReq, cancelFn := context.WithCancel(ctx)
					return Ctx{Context: ctxReq, CancelFn: cancelFn}
				}(),
			},
			wants: apperrors.RequestCancelledErr,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body []byte
			if len(tt.args.URLList) > 0 {
				body, _ = json.Marshal(tt.args.URLList)
			}
			if tt.args.Ctx.CancelFn != nil {
				defer tt.args.Ctx.CancelFn()
			}
			request := httptest.NewRequest(tt.args.Method, ApiV1, bytes.NewBuffer(body))
			request = request.WithContext(tt.args.Ctx.Context)

			w := httptest.NewRecorder()
			if tt.args.Ctx.CancelFn != nil {
				wg := sync.WaitGroup{}
				go func() {
					wg.Add(1)
					defer wg.Done()
					time.Sleep(time.Millisecond * 750)
					defer tt.args.Ctx.CancelFn()

				}()
				wg.Wait()

			}

			api.ServeHTTP(w, request)

			if w.Code != tt.wants.Code {
				t.Errorf("ProcessRequest() getStatusCode = %v, wantsStatusCode = %v", w.Code, tt.wants.Code)
			}
			bodyMessage := strings.Trim(w.Body.String(), "\n")
			if w.Code != http.StatusOK && tt.wants.Message != bodyMessage {
				t.Errorf("ProcessRequest() getStatusCode = %v, wantsStatusCode = %v", w.Code, tt.wants.Code)

			}
			//}

		})
	}
}
