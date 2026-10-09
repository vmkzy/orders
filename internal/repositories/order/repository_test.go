package order_test

import (
	"fmt"
	"sync"
	"testing"

	model "github.com/gulmix/orders/internal/models/order"
	orderRepository "github.com/gulmix/orders/internal/repositories/order"
)

func TestRepository(t *testing.T) {
	repo := orderRepository.NewMemory()

	var wg sync.WaitGroup
	start := make(chan struct{})
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			id := fmt.Sprintf("ord_%d", i)
			order := model.Order{
				ID:         id,
				CustomerID: "cust_1",
				Status:     "created",
				Items:      []model.Item{{SKU: "book", Qty: 1, PriceCents: 10}},
				TotalCents: 10,
			}
			repo.Save(order)
			_, _ = repo.Get(order.ID)
		}(i)
	}
	close(start)
	wg.Wait()
	for i := 0; i < 100; i++ {
		id := fmt.Sprintf("ord_%d", i)
		got, ok := repo.Get(id)
		if !ok {
			t.Fatalf("order %s not found", id)
		}
		if got.ID != id || got.TotalCents != 10 {
			t.Errorf("order %s invalid data order: %+v", id, got)
		}
	}

}
