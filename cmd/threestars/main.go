package main

import (
	"context"
	"log"
	"os"

	"github.com/johanstokking/pizza-bakery/internal/openai"
	"github.com/johanstokking/pizza-bakery/internal/pizzabakery"
	"github.com/johanstokking/pizza-bakery/internal/server"
)

func bake(orders pizzabakery.OrderChannel) {
	oven := &pizzabakery.GenAIOven{
		Generator: openai.NewImageGenerator(os.Getenv("OPENAI_API_KEY")),
	}

	for order := range orders {
		go func() {
			customer := pizzabakery.MustCustomerFromContext(order.Context)
			pizza, err := oven.BakePizza(order.Context, order.Order)
			if err == nil {
				select {
				case order.Result <- pizza:
				default:
					log.Printf("pizza is ready but customer %s is gone\n", customer)
				}
			}
		}()
	}
}

func main() {
	ctx := context.Background()

	orders := make(pizzabakery.OrderChannel, 10)
	go bake(orders)

	// Start the server.
	srv := server.New(ctx, ":8080", orders)
	srv.ListenAndServe()
}
