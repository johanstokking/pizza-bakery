package pizzabakery

import (
	"context"
	"fmt"
	"log"
)

// Bake bakes pizzas from orders.
func Bake(ctx context.Context, orders OrderBook, oven Oven) error {
	for {
		order, err := orders.Receive(ctx)
		if err != nil {
			return fmt.Errorf("pizzabakery: receive order: %w", err)
		}
		if order == nil {
			log.Println("no more orders")
			return nil
		}
		customer := MustCustomerFromContext(order.Context)

		pizza, err := oven.BakePizza(order.Context, order.Order)
		if err != nil {
			log.Printf("could not to bake pizza for %s: %v\n", customer, err)
			continue
		}
		select {
		case <-ctx.Done():
		case order.Result <- pizza:
		default:
			log.Printf("customer %s is no longer waiting, dispatch delivery\n", customer)
		}
	}
}
