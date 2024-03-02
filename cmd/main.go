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
	host, port := "localhost", "8080"
	srv := new(server.Server)

	go func() {
		fmt.Printf("listening on http://%s:%s\n", host, port)
		if err := srv.Run("localhost", "8080", api); err != nil {
			fmt.Println(err)
		}
	}()

	quit := make(chan os.Signal, 1) // check: set channel size == 2?
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	if err := srv.Shutdown(context.Background()); err != nil {
		fmt.Println(err)
	}
}
