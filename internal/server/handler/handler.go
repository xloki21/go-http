package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/xloki21/go-http/internal/model"
	"github.com/xloki21/go-http/pkg/source"
	"net/http"
	"sync/atomic"
	"time"

	mw "github.com/xloki21/go-http/internal/server/middlewares"
)

const (
	ApiV1 = "/api/v1/"
)
const (
	MaxUrlsPerRequest   = 7
	MaxOutgoingRequests = 4
	TimeoutPerRequest   = time.Second
)

func NewHandlers() *http.ServeMux {
	mux := new(http.ServeMux)
	mux.HandleFunc(ApiV1, mw.IncomingRequestCounter(ProcessRequest))
	return mux
}

func ProcessRequest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctxOps, cancelOPS := context.WithCancel(ctx)
	defer cancelOPS()
	requestProcessed := atomic.Bool{} // todo: make atomic?
	defer func() {
		requestProcessed.Store(true)
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				if !requestProcessed.Load() {
					http.Error(w, "Request cancelled", http.StatusBadRequest)
					cancelOPS()
				}
				return
			}
		}
	}()

	if r.Method == http.MethodPost {

		var urlList []model.URL
		decoder := json.NewDecoder(r.Body)
		// 1. Check format
		err := decoder.Decode(&urlList)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}
		// 2. Check payload size
		if len(urlList) == 0 {
			http.Error(w, "Empty URL List", http.StatusBadRequest)
			return
		}

		if len(urlList) > MaxUrlsPerRequest {
			http.Error(w, "URL List size exceeds limit", http.StatusBadRequest)
			return
		}

		// OK-case
		result, err := source.FetchURLList(ctxOps, urlList, MaxOutgoingRequests, TimeoutPerRequest)

		if err != nil {
			http.Error(w, fmt.Sprintf("Processing failed: %s", err), http.StatusBadRequest)
			return
		}
		body, err := json.Marshal(result)
		if err != nil {
			code := http.StatusInternalServerError
			http.Error(w, http.StatusText(code), code)
			return
		}

		// you must manually call w.WriteHeader before anything writes
		// to the response to set HTTP codes manually
		w.WriteHeader(http.StatusOK)
		_, err = w.Write(body)
		if err != nil {
			code := http.StatusInternalServerError
			http.Error(w, http.StatusText(code), code)
			return
		}

	} else {
		code := http.StatusMethodNotAllowed
		http.Error(w, http.StatusText(code), code)
		return
	}
}
