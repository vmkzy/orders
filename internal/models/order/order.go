// Package order — доменная модель заказа.
//
// Здесь живут типы предметной области и правила, которые к ним относятся.
// Пакет ничего не знает про HTTP, JSON и базу данных: он не импортирует
// net/http и не содержит тегов сериализации. Это проверяется на приёмке.
package order

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"time"
)

type Item struct {
	SKU        string
	Qty        int64
	PriceCents int64
}
type Order struct {
	ID         string
	CustomerID string
	Status     string
	Items      []Item
	TotalCents int64
	CreatedAt  time.Time
}

var ErrInvalidOrder = errors.New("invalid order")

func New(id, customerID string, items []Item, createdAt time.Time) (Order, error) {
	customerID = strings.TrimSpace(customerID)
	if customerID == "" {
		return Order{}, fmt.Errorf("%w:customerID must be no empty", ErrInvalidOrder)
	}
	if len(items) == 0 {
		return Order{}, fmt.Errorf("%w:items len must be positive", ErrInvalidOrder)
	}
	for _, val := range items {
		if val.Qty <= 0 {
			return Order{}, fmt.Errorf("%w:qty must be positive", ErrInvalidOrder)
		}
		if val.PriceCents < 0 {
			return Order{}, fmt.Errorf("%w:priceCents must be >= 0", ErrInvalidOrder)
		}

	}
	var total int64
	for _, val := range items {
		mul, ok := safeMul(val.Qty, val.PriceCents)
		if !ok {
			return Order{}, fmt.Errorf("%w: total must be <= MaxInt64", ErrInvalidOrder)
		}

		if math.MaxInt64-mul < total {
			return Order{}, fmt.Errorf("%w: total must be <= MaxInt64", ErrInvalidOrder)
		}
		total += mul

	}
	clon := make([]Item, len(items))
	copy(clon, items)
	return Order{
		ID:         id,
		CustomerID: customerID,
		Status:     "created",
		Items:      clon,
		TotalCents: total,
		CreatedAt:  createdAt,
	}, nil
}
func safeMul(a, b int64) (int64, bool) {
	if a == 0 || b == 0 {
		return 0, true
	}
	result := a * b
	if result/b != a {
		return 0, false
	}

	return result, true
}
