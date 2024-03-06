package handler

import (
	"github.com/xloki21/go-http/internal/server/apperrors"
	"log"
	"net/http"
	"runtime/debug"
	"sync/atomic"
	"time"
)

const MaxIncomingRequests = 100

var TotalRequestsInProcessing atomic.Int32

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
	return DomainSpecificErrors(RequestThrottler(hFunc))
}

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, req)
		log.Printf("%s %s %s", req.Method, req.RequestURI, time.Since(start))
	})
}

func PanicRecovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				log.Println(string(debug.Stack()))
			}
		}()
		next.ServeHTTP(w, req)
	})
}

func RequestThrottler(next HFuncWithError) HFuncWithError {
	return func(w http.ResponseWriter, r *http.Request) error {
		TotalRequestsInProcessing.Add(1)
		defer TotalRequestsInProcessing.Add(-1)
		if TotalRequestsInProcessing.Load() >= MaxIncomingRequests {
			return apperrors.TooManyRequestsErr
		}
		return next(w, r)
	}
}
