package server

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

type Server struct {
	httpServer *http.Server
}

func (s *Server) Run(host, port string, handler http.Handler) error {
	s.httpServer = &http.Server{
		Addr:              fmt.Sprintf("%s:%s", host, port),
		Handler:           handler,
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      40 * time.Second,
		MaxHeaderBytes:    1 << 20, // 1MB
	}
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
