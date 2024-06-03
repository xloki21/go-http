package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/xloki21/go-http/internal/metrics"
	"github.com/xloki21/go-http/internal/server"
	"github.com/xloki21/go-http/internal/server/handler"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	api := handler.NewHandler()
	srv := server.Server{}
	logger := slog.Default()

	go func() {
		logger.Info("listening on", "address", fmt.Sprintf("%s:%s", "localhost", "8080"))
		if err := srv.Run("localhost", "8080", api); err != nil {
			if err != nil && !errors.Is(err, http.ErrServerClosed) {
				logger.Error("error during server shutdown")
			}
		}
	}()

	go func() {
		_ = metrics.Listen("127.0.0.1:8082")
	}()

	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)
	<-quit
	ctx := context.Background()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error(err.Error())
	}
}
