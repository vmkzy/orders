// Package order — HTTP-контроллер заказов: разбор запроса, вызов сервиса,
// перевод доменных ошибок в коды ответа.
//
// Единый формат ответа и коды ошибок — в соседнем пакете response.
package order

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"time"

	"github.com/gulmix/orders/internal/controllers/http/response"
	model "github.com/gulmix/orders/internal/models/order"
	service "github.com/gulmix/orders/internal/services/order"
)

type createRequest struct {
	CustomerID string    `json:"customer_id"`
	Items      []itemDTO `json:"items"`
}
type orderResponse struct {
	ID         string    `json:"id"`
	CustomerID string    `json:"customer_id"`
	Status     string    `json:"status"`
	Items      []itemDTO `json:"items"`
	TotalCents int64     `json:"total_cents"`
	CreatedAt  time.Time `json:"created_at"`
}
type itemDTO struct {
	SKU        string `json:"sku"`
	Qty        int64  `json:"qty"`
	PriceCents int64  `json:"price_cents"`
}
type Controller struct {
	service *service.Service
}

func NewController(service *service.Service) *Controller {
	return &Controller{service: service}
}

func (c *Controller) Create(w http.ResponseWriter, r *http.Request) {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		response.Error(w, r, http.StatusBadRequest, response.CodeBadRequest, "would json")
		return
	}
	var req createRequest
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	if err := decoder.Decode(&req); err != nil {
		response.Error(w, r, http.StatusBadRequest, response.CodeBadRequest, "error decode req")
		return
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		response.Error(w, r, http.StatusBadRequest, response.CodeBadRequest, "json must be one object")
		return
	}
	items := make([]model.Item, len(req.Items))
	for i, val := range req.Items {
		items[i] = model.Item{
			SKU:        val.SKU,
			Qty:        val.Qty,
			PriceCents: val.PriceCents,
		}
	}
	created, err := c.service.Create(req.CustomerID, items)
	if errors.Is(err, model.ErrInvalidOrder) {
		response.Error(w, r, http.StatusUnprocessableEntity, response.CodeValidation, "invalid info order")
		return
	} else if err != nil {
		response.Error(w, r, http.StatusInternalServerError, response.CodeInternal, "internal server error")
		return
	}
	var orderResp orderResponse
	w.Header().Set("Location", fmt.Sprintf("/orders/%s", created.ID))
	orderResp = modelToDTO(created)
	response.JSON(w, r, 201, orderResp)
}

func (c *Controller) Get(w http.ResponseWriter, r *http.Request) {
	strID := r.PathValue("id")
	order, err := c.service.Get(strID)
	if errors.Is(err, service.ErrNotFound) {
		response.Error(w, r, http.StatusNotFound, response.CodeNotFound, "order not found")
		return
	} else if err != nil {
		response.Error(w, r, http.StatusInternalServerError, response.CodeInternal, "internal server error")
		return
	}
	orderDTO := modelToDTO(order)
	response.JSON(w, r, 200, orderDTO)
}
func modelToDTO(o model.Order) orderResponse {
	items := make([]itemDTO, len(o.Items))
	for i, val := range o.Items {
		items[i] = itemDTO{
			SKU:        val.SKU,
			Qty:        val.Qty,
			PriceCents: val.PriceCents,
		}
	}
	return orderResponse{
		ID:         o.ID,
		CustomerID: o.CustomerID,
		Status:     o.Status,
		Items:      items,
		TotalCents: o.TotalCents,
		CreatedAt:  o.CreatedAt,
	}
}
