package pizzabakery

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/johanstokking/pizza-bakery/internal"
)

// Oven bakes orders into pizzas or errors.
type Oven interface {
	BakePizza(context.Context, *Order) (Pizza, error)
}

// GenAIOven is an oven that uses an AI to bake pizzas.
type GenAIOven struct {
	// Generator is the AI that generates images of pizzas.
	Generator internal.ImageGenerator
}

// BakePizza implements Oven.
func (o *GenAIOven) BakePizza(ctx context.Context, order *Order) (Pizza, error) {
	// Build the prompt	for the AI.
	customer := MustCustomerFromContext(ctx)
	prompt := fmt.Sprintf(
		"A freshly baked pizza from the oven with toppings %s and melted cheese letters spelling out the name %s",
		strings.Join(order.Toppings, ", "), strings.ToUpper(string(customer)),
	)

	// Bake the pizza.
	log.Println("baking pizza for", customer)
	img, err := o.Generator.Generate(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("pizzabakery: generate AI pizza: %w", err)
	}
	log.Println("pizza is ready for", customer)
	return Pizza(img), nil
}
