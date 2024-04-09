package server

import (
	"context"
	"net/http"

	"github.com/johanstokking/pizza-bakery/internal/pizzabakery"
)

func authenticatedUserIsCustomer(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		username, _ := AuthenticatedUser(ctx)
		ctx = pizzabakery.WithCustomer(ctx, pizzabakery.Customer(username))
		handler.ServeHTTP(w, r.WithContext(ctx))
	})
}

// decoupleRequestContext decouples from the request context. This function returns a derived context from the parent
// that contains the customer from the request context.
func decoupleRequestContext(parent, request context.Context) context.Context {
	customer := pizzabakery.MustCustomerFromContext(request)
	return pizzabakery.WithCustomer(parent, customer)
}
