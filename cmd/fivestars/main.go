package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/johanstokking/pizza-bakery/internal/openai"
	"github.com/johanstokking/pizza-bakery/internal/pizzabakery"
	"github.com/johanstokking/pizza-bakery/internal/server"
	"golang.org/x/sync/errgroup"
)

func main() {
	ctx := context.Background()

	// Prepare the oven and the order book.
	oven := &pizzabakery.GenAIOven{
		Generator: openai.NewImageGenerator(os.Getenv("OPENAI_API_KEY")),
	}
	orders := make(pizzabakery.OrderChannel, 10)

	// Create an errgroup derived from the background context.
	// The errgroup will be used to wait for all goroutines to finish.
	// If any worker returns an error, the wgCtx will be canceled for the other workers to stop.
	wg, wgCtx := errgroup.WithContext(ctx)

	// Subscribe to SIGTERM and SIGINT signals.
	// SIGTERM is typically sent by a process manager to request the process to gracefully terminate,
	// while SIGINT is sent by the user to request the process to terminate.
	stopCtx, _ := signal.NotifyContext(wgCtx, syscall.SIGTERM, syscall.SIGINT)

	// Start the server.
	srv := server.New(ctx, ":8080", orders)
	wg.Go(func() error {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		log.Println("server stopped")
		return nil
	})

	// Start 5 baking processes.
	for i := range 5 {
		wg.Go(func() error {
			defer log.Printf("baking process %d is done\n", i)
			return pizzabakery.Bake(wgCtx, orders, oven)
		})
	}

	// Gracefully shut down the server and close the order book when the stopCtx is done.
	wg.Go(func() error {
		<-stopCtx.Done()
		log.Println("shutting down server")
		srv.Shutdown(ctx)
		log.Println("closing orders")
		close(orders)
		return nil
	})

	log.Println("server is running on", srv.Addr)

	// Wait for all goroutines to finish.
	if err := wg.Wait(); err != nil {
		log.Printf("bakery stopped with error: %v\n", err)
		os.Exit(1)
	}
}
