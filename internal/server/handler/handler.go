package handler

import (
	"net/http"
	"time"
)

const (
	MaxUrlsPerRequest   = 20
	MaxOutgoingRequests = 4
	TimeoutPerRequest   = time.Second
)

const (
	ApiV1 = "/api/v1"
	Fetch = ApiV1 + "/fetch"
)

type HFuncWithError func(http.ResponseWriter, *http.Request) error

func NewHandler() *http.ServeMux {
	mux := new(http.ServeMux)
	mux.HandleFunc(Fetch, MWChain(FetchHandlerFunc))
	return mux
}
