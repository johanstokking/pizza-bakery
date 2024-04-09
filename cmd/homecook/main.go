package main

import (
	"context"
	"image/png"
	"log"
	"os"
	"time"

	"github.com/johanstokking/pizza-bakery/internal/openai"
	"github.com/johanstokking/pizza-bakery/internal/pizzabakery"
)

func main() {
	ctx := context.Background()

	oven := &pizzabakery.GenAIOven{
		Generator: openai.NewImageGenerator(os.Getenv("OPENAI_API_KEY")),
	}

	order := &pizzabakery.Order{
		Toppings: []string{
			"ricotta cheese",
			"pepperoni",
			"sausage",
		},
	}

	timeoutCtx, cancelTimeout := context.WithTimeout(ctx, 30*time.Second)
	defer cancelTimeout()

	customerCtx := pizzabakery.WithCustomer(timeoutCtx, "Johan")
	pizza, err := oven.BakePizza(customerCtx, order)
	if err != nil {
		log.Fatalln("failed to bake pizza:", err)
	}

	out, err := os.CreateTemp("", "pizza-*.png")
	if err != nil {
		log.Fatalln("failed to create temp file:", err)
	}
	defer out.Close()

	if err := png.Encode(out, pizza); err != nil {
		log.Fatalln("failed to save image:", err)
	}
	log.Println("saved pizza to", out.Name())
}
