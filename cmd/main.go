package main

import (
	"context"
	"fmt"
	"github.com/xloki21/go-http/internal/server"
	"github.com/xloki21/go-http/internal/server/handler"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	api := handler.NewHandlers()
	srv := server.Server{}

	go func() {
		fmt.Printf("listening on http://%s:%s\n", "localhost", "8080")
		if err := srv.Run("localhost", "8080", api); err != nil {
			fmt.Println(err)
		}
	}()

	quit := make(chan os.Signal, 1) // check: set channel size == 2?
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit
	ctx := context.Background()
	if err := srv.Shutdown(ctx); err != nil {
		fmt.Println(err)
	}
}
