package pizzabakery

import "context"

// Order is a pizza order.
type Order struct {
	Toppings []string
}

// PlacedOrder is an order with a channel to receive the result.
type PlacedOrder struct {
	*Order
	Context context.Context
	Result  chan Pizza
}

// OrderBook is a source of pizza orders.
type OrderBook interface {
	// Receive blocks until an order is received or the context is done.
	// If there are no more orders to receive, Receive returns nil.
	Receive(context.Context) (*PlacedOrder, error)
}

// OrderChannel is an OrderBook that receives orders from a channel.
type OrderChannel chan *PlacedOrder

// Place places an order in the channel and returns the placed order.
func (oc OrderChannel) Place(ctx context.Context, order *Order) (<-chan Pizza, error) {
	placedOrder := &PlacedOrder{
		Context: ctx,
		Order:   order,
		Result:  make(chan Pizza),
	}
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case oc <- placedOrder:
		return placedOrder.Result, nil
	}
}

// Receive implements OrderBook.
func (oc OrderChannel) Receive(ctx context.Context) (*PlacedOrder, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case order, ok := <-oc:
		if !ok {
			return nil, nil
		}
		return order, nil
	}
}
