package handler

import (
	"github.com/xloki21/go-http/internal/server/apperrors"
	"net/http"
	"sync/atomic"
)

const MaxIncomingRequests = 100

var TotalRequestsInProcessing atomic.Int32

type HFuncWithError func(http.ResponseWriter, *http.Request) error

func DomainSpecificErrors(next HFuncWithError) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := next(w, r)
		if err != nil {
			errw := err.(apperrors.AppError)
			http.Error(w, errw.Message, errw.Code)
		}
	}
}

func MWChain(hFunc HFuncWithError) http.HandlerFunc {
	return DomainSpecificErrors(IncomingRequestCounter(hFunc))
}

func IncomingRequestCounter(next HFuncWithError) HFuncWithError {
	return func(w http.ResponseWriter, r *http.Request) error {
		TotalRequestsInProcessing.Add(1)
		defer TotalRequestsInProcessing.Add(-1)
		if TotalRequestsInProcessing.Load() >= MaxIncomingRequests {
			return apperrors.TooManyRequestsErr
		}
		return next(w, r)
	}
}
