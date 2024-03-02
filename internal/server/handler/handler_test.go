package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/xloki21/go-http/internal/server"
	"net/http"
	"strings"
	"testing"
)

func TestProcessRequest(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background()) // todo: user request context
	defer cancel()

	api := NewHandlers()

	srv := new(server.Server)
	go func() {
		// run test server
		if err := srv.Run("localhost", "8080", api); err != nil {
			fmt.Println(err)
		}
	}()

	defer func() {
		srv.Shutdown(context.Background())
	}()

	host, port := "localhost", "8080"
	endpoint := fmt.Sprintf("http://%s:%s%s", host, port, ApiV1)

	hugeRequest := strings.Split(strings.Repeat("https://go.dev/images/go-logo-white.svg ", MaxUrlsPerRequest+1), " ")
	wrongAddressRequest := []string{"https://1go.dev/images/go-logo-white.svg"}
	bHugeRequest, _ := json.Marshal(hugeRequest)
	bCorrectRequest, _ := json.Marshal(hugeRequest[:MaxUrlsPerRequest])
	bWrongAddressRequest, _ := json.Marshal(wrongAddressRequest)
	type args struct {
		r *http.Request
	}
	tests := []struct {
		name      string
		args      args
		wantsCode int
	}{
		{
			name: "Send GET Request",
			args: args{r: func() *http.Request {
				req, err := http.NewRequest(http.MethodGet, endpoint, nil)
				if err != nil {
					t.Errorf("%v", err)
				}
				return req
			}()},
			wantsCode: http.StatusMethodNotAllowed,
		},
		{
			name: "Send POST Request with nil body",
			args: args{r: func() *http.Request {
				req, err := http.NewRequest(http.MethodPost, endpoint, nil)
				if err != nil {
					t.Errorf("%v", err)
				}
				return req
			}()},
			wantsCode: http.StatusUnprocessableEntity,
		},
		{
			name: "Send POST Request with wrong URL address",
			args: args{r: func() *http.Request {
				req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewBuffer(bWrongAddressRequest))
				if err != nil {
					t.Errorf("%v", err)
				}
				return req
			}()},
			wantsCode: http.StatusBadRequest,
		},
		{
			name: "Send Post Request with URL list size exceeded limit",
			args: args{r: func() *http.Request {
				req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint,
					bytes.NewBuffer(bHugeRequest))
				if err != nil {
					t.Errorf("%v", err)
				}
				return req
			}()},
			wantsCode: http.StatusBadRequest,
		},
		{
			name: "Send Post Request with Correct URL List",
			args: args{r: func() *http.Request {
				req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint,
					bytes.NewBuffer(bCorrectRequest))

				if err != nil {
					t.Errorf("%v", err)
				}
				return req
			}()},
			wantsCode: http.StatusOK,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := http.DefaultClient.Do(tt.args.r)
			defer resp.Body.Close()
			if err != nil {
				t.Errorf("%v", err)
			}
			if resp.StatusCode != tt.wantsCode {
				t.Errorf("processRequest() getStatusCode = %v, wantsStatusCode = %v", resp.StatusCode, tt.wantsCode)
			}
		})
	}
}
