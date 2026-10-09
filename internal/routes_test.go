package internal_test

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gulmix/orders/internal"
	healthController "github.com/gulmix/orders/internal/controllers/http/health"
	orderController "github.com/gulmix/orders/internal/controllers/http/order"
	orderRepository "github.com/gulmix/orders/internal/repositories/order"
	"github.com/gulmix/orders/internal/requestid"
	orderService "github.com/gulmix/orders/internal/services/order"
)

// Тест уровня маршрутов. Он ничего не проверяет из вашего задания — он
// показывает, как тестировать HTTP-слой: через httptest, без
// поднятия настоящего сокета и без сети, поэтому такие тесты идут за
// миллисекунды и не флакают.
//
// Тесты на заказы пишете вы. Ориентир по объёму — в задании модуля.
func TestHealth(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	repo := orderRepository.NewMemory()
	svc := orderService.NewService(repo)
	orders := orderController.NewController(svc)

	srv := internal.InitRoutes(logger, healthController.NewController(), orders)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, ожидался %d", rec.Code, http.StatusOK)
	}

	var got map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("ответ не разобрался как JSON: %v", err)
	}
	if got["status"] != "ok" {
		t.Errorf(`status = %q, ожидался "ok"`, got["status"])
	}

	// Заголовок проставляет middleware — заодно проверяем, что обвязка на месте.
	if rec.Header().Get(requestid.Header) == "" {
		t.Errorf("заголовок %s пустой", requestid.Header)
	}
}

func TestOrder(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	repo := orderRepository.NewMemory()
	svc := orderService.NewService(repo)
	orders := orderController.NewController(svc)

	srv := internal.InitRoutes(logger, healthController.NewController(), orders)
	body := `{"customer_id": "cust_1", "items": [{"sku": "book", "qty": 2, "price_cents": 1500}]}`
	req := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	if rec.Code != 201 {
		t.Fatalf("invalid code %d, expected 201, body: %q", rec.Code, rec.Body.String())
	}
	var got struct {
		ID         string `json:"id"`
		TotalCents int64  `json:"total_cents"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("invalid response: %v", err)
	}
	if got.ID == "" {
		t.Fatalf("id order empty")
	}
	if got.TotalCents != 3000 {
		t.Errorf("total cents = %d, expected 3000", got.TotalCents)
	}
	if rec.Header().Get("Location") != "/orders/"+got.ID {
		t.Errorf("Location = %s, expected /orders/%s", rec.Header().Get("Location"), got.ID)
	}
	req1 := httptest.NewRequest(http.MethodGet, rec.Header().Get("Location"), nil)
	rec1 := httptest.NewRecorder()
	srv.ServeHTTP(rec1, req1)
	if rec1.Code != 200 {
		t.Fatalf("invalid code %d, expecred 200, body: %q", rec1.Code, rec1.Body.String())
	}
	var gotGet struct {
		ID         string `json:"id"`
		TotalCents int64  `json:"total_cents"`
	}
	if err := json.NewDecoder(rec1.Body).Decode(&gotGet); err != nil {
		t.Fatalf("invalid response: %v", err)
	}
	if gotGet.ID != got.ID {
		t.Fatalf("GET ID = %s not equal POST id = %s", gotGet.ID, got.ID)
	}
	if gotGet.TotalCents != got.TotalCents {
		t.Fatalf("GET total_cents = %d not equal POST total_cents = %d", gotGet.TotalCents, got.TotalCents)
	}
	req2 := httptest.NewRequest(http.MethodGet, "/orders/err", nil)
	rec2 := httptest.NewRecorder()
	srv.ServeHTTP(rec2, req2)
	if rec2.Code != 404 {
		t.Fatalf("invalid code %d, expected 404, body: %q", rec2.Code, rec2.Body.String())
	}
	req3 := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(`{"customer_id":`))
	req3.Header.Set("Content-Type", "application/json")
	rec3 := httptest.NewRecorder()
	srv.ServeHTTP(rec3, req3)
	if rec3.Code != 400 {
		t.Fatalf("invalid code %d, expected 400, body: %q", rec3.Code, rec3.Body.String())
	}
	req4 := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(body))
	req4.Header.Set("Content-Type", "text/plain")
	rec4 := httptest.NewRecorder()
	srv.ServeHTTP(rec4, req4)
	if rec4.Code != 400 {
		t.Fatalf("invalid code %d, expected 400, body: %q", rec4.Code, rec4.Body.String())
	}
	InvalidBody := `{"customer_id": "cust_1", "items": [{"sku": "book", "qty": 0, "price_cents": 1500}]}`
	req5 := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(InvalidBody))
	req5.Header.Set("Content-Type", "application/json")
	rec5 := httptest.NewRecorder()
	srv.ServeHTTP(rec5, req5)
	if rec5.Code != 422 {
		t.Fatalf("invalid code %d, expected 422, body: %q", rec5.Code, rec5.Body.String())
	}
	var gotErr struct {
		Error struct {
			Code      string `json:"code"`
			RequestID string `json:"request_id"`
		} `json:"error"`
	}

	if err := json.NewDecoder(rec5.Body).Decode(&gotErr); err != nil {
		t.Fatalf("take gotErr error: %v", err)
	}
	if gotErr.Error.Code != "validation_error" {
		t.Errorf("error.code = %q, expected %q", gotErr.Error.Code, "validation_error")
	}
	if gotErr.Error.RequestID == "" {
		t.Error("error.request_id empty")
	}
}
