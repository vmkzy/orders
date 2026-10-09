// Package order — сервисный слой заказов: сценарии поверх домена и хранилища.
//
// Здесь решается, что вообще происходит: провалидировать, посчитать сумму,
// сходить в хранилище, вернуть доменную ошибку. Контроллер после этого
// занимается только переводом в HTTP.
package order

import (
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	model "github.com/gulmix/orders/internal/models/order"
	repository "github.com/gulmix/orders/internal/repositories/order"
)

var ErrNotFound = errors.New("order not found")

type Service struct {
	repo   *repository.Memory
	nextID atomic.Uint64
}

func NewService(repo *repository.Memory) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(customerID string, items []model.Item) (model.Order, error) {
	id := fmt.Sprintf("ord_%d", s.nextID.Add(1))
	timeOr := time.Now().UTC()
	order, err := model.New(id, customerID, items, timeOr)
	if err != nil {
		return model.Order{}, err
	}
	s.repo.Save(order)
	return order, nil
}

func (s *Service) Get(id string) (model.Order, error) {
	value, ok := s.repo.Get(id)
	if !ok {
		return model.Order{}, ErrNotFound
	}
	return value, nil
}
