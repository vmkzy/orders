// Package order — хранилище заказов.
//
// В Модуле 0 это карта в памяти. В Модуле 5 на её место встанет Postgres,
// и если границы проведены правильно, менять придётся только этот пакет.
package order

import (
	"sync"

	"github.com/gulmix/orders/internal/models/order"
)

func cloneOrder(o order.Order) order.Order {
	clone := make([]order.Item, len(o.Items))
	copy(clone, o.Items)
	o.Items = clone
	return o
}

type Memory struct {
	mu   sync.Mutex
	data map[string]order.Order
}

func NewMemory() *Memory {
	return &Memory{
		data: make(map[string]order.Order),
	}
}

func (s *Memory) Save(o order.Order) {
	s.mu.Lock()
	defer s.mu.Unlock()
	o = cloneOrder(o)
	s.data[o.ID] = o
}
func (s *Memory) Get(id string) (order.Order, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	value, ok := s.data[id]
	value = cloneOrder(value)
	return value, ok
}
