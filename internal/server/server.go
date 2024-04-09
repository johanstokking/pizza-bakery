package server

import (
	"bytes"
	"context"
	"embed"
	"errors"
	"image/png"
	"io/fs"
	"log"
	"net/http"
	"time"

	"github.com/johanstokking/pizza-bakery/internal/pizzabakery"
)

//go:embed static
var staticFS embed.FS

// MiddlewareFunc is a function that wraps an HTTP handler.
type MiddlewareFunc func(http.Handler) http.Handler

func applyMiddleware(h http.Handler, middleware ...MiddlewareFunc) http.Handler {
	for i := len(middleware) - 1; i >= 0; i-- {
		h = middleware[i](h)
	}
	return h
}

func handleOrder(ctx context.Context, orders pizzabakery.OrderChannel) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		toppings := make([]string, 0, len(r.PostForm["topping"]))
		for _, topping := range r.PostForm["topping"] {
			if topping == "" {
				continue
			}
			toppings = append(toppings, topping)
		}
		order := &pizzabakery.Order{
			Toppings: toppings,
		}

		reqCtx := r.Context()
		result, err := orders.Place(decoupleRequestContext(ctx, reqCtx), order)
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				http.Error(w, "too many requests", http.StatusTooManyRequests)
			} else if errors.Is(err, context.Canceled) {
				http.Error(w, "request canceled", http.StatusRequestTimeout)
			} else {
				http.Error(w, "internal", http.StatusInternalServerError)
			}
			return
		}

		select {
		case <-reqCtx.Done():
			log.Println("request canceled")
			http.Error(w, "timeout", http.StatusRequestTimeout)
		case pizza, ok := <-result:
			if !ok || pizza == nil {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			buf := bytes.NewBuffer(nil)
			if err := png.Encode(buf, pizza); err != nil {
				http.Error(w, "internal", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "image/png")
			http.ServeContent(w, r, "pizza.png", time.Now(), bytes.NewReader(buf.Bytes()))
		}
	})
}

// New returns a new HTTP server.
func New(ctx context.Context, addr string, orders pizzabakery.OrderChannel) *http.Server {
	mux := http.NewServeMux()

	// Handle static files
	static, _ := fs.Sub(staticFS, "static")
	mux.Handle("GET /", http.FileServer(http.FS(static)))

	// Handle order requests
	mux.Handle("POST /order", applyMiddleware(
		handleOrder(ctx, orders),
		// Middleware: first authenticate, then use the authenticated username as customer name
		Authenticate(AllowAnyone),
		authenticatedUserIsCustomer,
	))

	return &http.Server{
		Addr:    addr,
		Handler: mux,
	}
}
