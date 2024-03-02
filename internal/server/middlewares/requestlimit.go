package mw

import (
	"net/http"
	"sync/atomic"
)

var TotalRequestsInProcessing atomic.Int32

const MaxIncomingRequests = 100

func IncomingRequestCounter(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		TotalRequestsInProcessing.Add(1)
		defer TotalRequestsInProcessing.Add(-1)
		if TotalRequestsInProcessing.Load() >= MaxIncomingRequests {
			http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
			return
		}

		next(w, r)
	}
}
