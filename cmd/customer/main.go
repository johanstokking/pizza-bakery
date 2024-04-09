package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand/v2"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"

	"golang.org/x/sync/errgroup"
)

func main() {
	number := flag.Int("n", 5, "number of pizzas to bake")
	concurrency := flag.Int("c", 5, "number of pizzas to order concurrently")
	flag.Parse()

	semaphore := make(chan struct{}, *concurrency)

	wg, ctx := errgroup.WithContext(context.Background())
	for i := range *number {
		wg.Go(func() error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case semaphore <- struct{}{}:
			}
			defer func() {
				<-semaphore
			}()

			name := names[rand.IntN(len(names))]
			toppings := make([]string, rand.IntN(4)+2) // 2-5 toppings
			for i := range len(toppings) {
				toppings[i] = menu[rand.IntN(len(menu))]
			}
			log.Printf("ordering pizza %d for %s with toppings %s\n", i, name, toppings)

			formData := url.Values{
				"topping": toppings,
			}
			req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("http://%s@localhost:8080/order", name),
				strings.NewReader(formData.Encode()),
			)
			if err != nil {
				log.Printf("prepare request %d failed: %v\n", i, err)
				return nil
			}
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			res, err := http.DefaultClient.Do(req)
			if err != nil {
				log.Printf("order %d failed: %v\n", i, err)
				return nil
			}
			// Thou shalt close the body and read everything
			defer res.Body.Close()
			defer io.Copy(io.Discard, res.Body)

			tmpFile, err := os.CreateTemp("", "pizza-*.png")
			if err != nil {
				log.Printf("create temp file %d failed: %v\n", i, err)
				return nil
			}
			defer tmpFile.Close()

			if _, err := io.Copy(tmpFile, res.Body); err != nil {
				log.Printf("save pizza %d failed: %v\n", i, err)
				return nil
			}
			tmpFile.Close()
			log.Printf("pizza %d for %s saved to %s\n", i, name, tmpFile.Name())

			exec.Command("open", tmpFile.Name()).Run()
			return nil
		})
	}

	err := wg.Wait()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to bake pizzas: %v\n", err)
		os.Exit(1)
	}
}
