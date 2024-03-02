package handler

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/xloki21/go-http/internal/model"
	"github.com/xloki21/go-http/internal/server/apperrors"
	"github.com/xloki21/go-http/pkg/source"
	"net"
	"net/http"
	"net/url"
	"sync/atomic"
	"time"
)

const (
	ApiV1 = "/api/v1/"
)
const (
	MaxUrlsPerRequest   = 20
	MaxOutgoingRequests = 4
	TimeoutPerRequest   = time.Second
)

func NewHandlers() *http.ServeMux {
	mux := new(http.ServeMux)
	mux.HandleFunc(ApiV1, MWChain(ProcessRequest))
	return mux
}

func ProcessRequest(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	ctxOps, cancelOPS := context.WithCancel(ctx)
	defer cancelOPS()
	requestProcessed := atomic.Bool{}
	defer func() {
		requestProcessed.Store(true)
	}()

	go func() {
		<-ctx.Done()
		if !requestProcessed.Load() {
			http.Error(w, "Request cancelled", http.StatusBadRequest)
			cancelOPS()
		}
	}()
	if r.Method != http.MethodPost {
		return apperrors.MethodNotAllowed
	}

	var urlList []model.URL
	decoder := json.NewDecoder(r.Body)
	// 1. Check format
	err := decoder.Decode(&urlList)
	if err != nil {
		return apperrors.InvalidBodyErr
	}
	// 2. Check payload size
	if len(urlList) == 0 {
		return apperrors.EmptyBodyErr
	}

	if len(urlList) > MaxUrlsPerRequest {
		return apperrors.TooBigURLListErr
	}

	result, err := source.FetchURLList(ctxOps, urlList, MaxOutgoingRequests, TimeoutPerRequest)
	if err != nil {

		if errors.Is(err, context.DeadlineExceeded) {
			return apperrors.TimeoutErr
		}
		if errors.Is(err, context.Canceled) {
			return apperrors.RequestCancelledErr
		}

		if err, ok := err.(*url.Error); ok {
			if err, ok := err.Err.(*net.OpError); ok {
				if _, ok := err.Err.(*net.DNSError); ok {
					return apperrors.URLNotFoundErr
				}
			}
		}

		// common case
		return apperrors.BadRequestErr
	}
	body, err := json.Marshal(result)
	if err != nil {
		return apperrors.InternalErr
	}

	w.WriteHeader(http.StatusOK)
	if _, err = w.Write(body); err != nil {
		return apperrors.InternalErr
	}
	return nil
}
