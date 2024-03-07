package handler

import (
	"net/http"
	"time"
)

const (
	maxUrlsPerRequest   = 20
	maxOutgoingRequests = 4
	timeoutPerRequest   = time.Second
)

const (
	apiV1 = "/api/v1"
	fetch = apiV1 + "/fetch"
)

type HFuncWithError func(http.ResponseWriter, *http.Request) error

func NewHandler() *http.ServeMux {
	mux := new(http.ServeMux)
	mux.HandleFunc(fetch, MWChain(FetchHandlerFunc))
	return mux
}
