package order_test

import (
	"errors"
	"testing"
	"time"

	order "github.com/gulmix/orders/internal/models/order"
)

func TestModelOrder(t *testing.T) {
	createdAt := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	validItems := []order.Item{{SKU: "book", Qty: 1, PriceCents: 100}}
	tests := []struct {
		name       string
		customerID string
		items      []order.Item
	}{
		{
			name:       "empty customer",
			customerID: "",
			items:      validItems,
		},
		{
			name:       "empty items",
			customerID: "cust_1",
			items:      nil,
		},
		{
			name:       "qty equal 0",
			customerID: "cust_1",
			items:      []order.Item{{SKU: "book", Qty: 0, PriceCents: 1}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := order.New("ord_1", tt.customerID, tt.items, createdAt)
			if !errors.Is(err, order.ErrInvalidOrder) {
				t.Errorf("error: %v, expexted ErrInvalidOrder", err)
			}
		})
	}
}
