package pizzabakery

import "context"

// Customer is a customer of the pizza bakery.
type Customer string

type customerContextKeyType struct{}

var customerContextKey = customerContextKeyType{}

// WithCustomer returns a new context with the given customer.
func WithCustomer(ctx context.Context, c Customer) context.Context {
	return context.WithValue(ctx, customerContextKey, c)
}

// MustCustomerFromContext returns the customer from the context.
func MustCustomerFromContext(ctx context.Context) Customer {
	return ctx.Value(customerContextKey).(Customer)
}
