package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"wallet-api/internal/model"
	"wallet-api/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type mockWalletService struct {
	getBalance       func(ctx context.Context, walletID uuid.UUID) (*model.Wallet, error)
	executeOperation func(ctx context.Context, req model.WalletOperationRequest) (*model.WalletOperationResponse, error)
}

func (m *mockWalletService) GetBalance(ctx context.Context, walletID uuid.UUID) (*model.Wallet, error) {
	if m.getBalance != nil {
		return m.getBalance(ctx, walletID)
	}
	return nil, nil
}

func (m *mockWalletService) ExecuteOperation(ctx context.Context, req model.WalletOperationRequest) (*model.WalletOperationResponse, error) {
	if m.executeOperation != nil {
		return m.executeOperation(ctx, req)
	}
	return nil, nil
}

func TestWalletHandler_GetBalance(t *testing.T) {
	walletID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

	t.Run("200 OK with balance", func(t *testing.T) {
		svc := &mockWalletService{
			getBalance: func(context.Context, uuid.UUID) (*model.Wallet, error) {
				return &model.Wallet{ID: walletID, Balance: 200}, nil
			},
		}
		h := NewWalletHandler(svc)
		req := httptest.NewRequest(http.MethodGet, "/wallets/550e8400-e29b-41d4-a716-446655440000", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("WALLET_UUID", walletID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rec := httptest.NewRecorder()
		h.GetBalance(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("status = %d, want 200", rec.Code)
		}
		var body map[string]interface{}
		if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if body["balance"] != float64(200) {
			t.Errorf("balance = %v, want 200", body["balance"])
		}
	})

	t.Run("400 invalid UUID", func(t *testing.T) {
		h := NewWalletHandler(&mockWalletService{})
		req := httptest.NewRequest(http.MethodGet, "/wallets/not-a-uuid", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("WALLET_UUID", "not-a-uuid")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rec := httptest.NewRecorder()
		h.GetBalance(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want 400", rec.Code)
		}
	})

	t.Run("404 wallet not found", func(t *testing.T) {
		svc := &mockWalletService{
			getBalance: func(context.Context, uuid.UUID) (*model.Wallet, error) {
				return nil, service.ErrWalletNotFound
			},
		}
		h := NewWalletHandler(svc)
		req := httptest.NewRequest(http.MethodGet, "/wallets/550e8400-e29b-41d4-a716-446655440000", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("WALLET_UUID", walletID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rec := httptest.NewRecorder()
		h.GetBalance(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("status = %d, want 404", rec.Code)
		}
	})
}

func TestWalletHandler_ExecuteOperation(t *testing.T) {
	walletID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

	t.Run("200 OK DEPOSIT", func(t *testing.T) {
		svc := &mockWalletService{
			executeOperation: func(context.Context, model.WalletOperationRequest) (*model.WalletOperationResponse, error) {
				return &model.WalletOperationResponse{WalletID: walletID, Balance: 100}, nil
			},
		}
		h := NewWalletHandler(svc)
		body := map[string]interface{}{
			"walletId":      walletID.String(),
			"operationType": "DEPOSIT",
			"amount":        float64(100),
		}
		raw, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/wallet", bytes.NewReader(raw))
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()
		h.ExecuteOperation(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("status = %d, want 200", rec.Code)
		}
		var resp model.WalletOperationResponse
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if resp.Balance != 100 {
			t.Errorf("balance = %d, want 100", resp.Balance)
		}
	})

	t.Run("405 method not allowed", func(t *testing.T) {
		h := NewWalletHandler(&mockWalletService{})
		req := httptest.NewRequest(http.MethodGet, "/wallet", nil)
		rec := httptest.NewRecorder()
		h.ExecuteOperation(rec, req)
		if rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("status = %d, want 405", rec.Code)
		}
	})

	t.Run("400 invalid body", func(t *testing.T) {
		h := NewWalletHandler(&mockWalletService{})
		req := httptest.NewRequest(http.MethodPost, "/wallet", bytes.NewReader([]byte("not json")))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		h.ExecuteOperation(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want 400", rec.Code)
		}
	})

	t.Run("400 missing walletId", func(t *testing.T) {
		body := map[string]interface{}{
			"operationType": "DEPOSIT",
			"amount":        float64(10),
		}
		raw, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/wallet", bytes.NewReader(raw))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		h := NewWalletHandler(&mockWalletService{})
		h.ExecuteOperation(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want 400", rec.Code)
		}
	})

	t.Run("404 wallet not found", func(t *testing.T) {
		svc := &mockWalletService{
			executeOperation: func(context.Context, model.WalletOperationRequest) (*model.WalletOperationResponse, error) {
				return nil, service.ErrWalletNotFound
			},
		}
		h := NewWalletHandler(svc)
		body := map[string]interface{}{
			"walletId":      walletID.String(),
			"operationType": "WITHDRAW",
			"amount":        float64(10),
		}
		raw, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/wallet", bytes.NewReader(raw))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		h.ExecuteOperation(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Errorf("status = %d, want 404", rec.Code)
		}
	})

	t.Run("409 insufficient funds", func(t *testing.T) {
		svc := &mockWalletService{
			executeOperation: func(context.Context, model.WalletOperationRequest) (*model.WalletOperationResponse, error) {
				return nil, service.ErrInsufficientFunds
			},
		}
		h := NewWalletHandler(svc)
		body := map[string]interface{}{
			"walletId":      walletID.String(),
			"operationType": "WITHDRAW",
			"amount":        float64(1000),
		}
		raw, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/wallet", bytes.NewReader(raw))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		h.ExecuteOperation(rec, req)
		if rec.Code != http.StatusConflict {
			t.Errorf("status = %d, want 409", rec.Code)
		}
	})

	t.Run("400 invalid amount", func(t *testing.T) {
		svc := &mockWalletService{
			executeOperation: func(context.Context, model.WalletOperationRequest) (*model.WalletOperationResponse, error) {
				return nil, service.ErrInvalidAmount
			},
		}
		h := NewWalletHandler(svc)
		body := map[string]interface{}{
			"walletId":      walletID.String(),
			"operationType": "DEPOSIT",
			"amount":        float64(0),
		}
		raw, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/wallet", bytes.NewReader(raw))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		h.ExecuteOperation(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want 400", rec.Code)
		}
	})

	t.Run("400 invalid operation type", func(t *testing.T) {
		svc := &mockWalletService{
			executeOperation: func(context.Context, model.WalletOperationRequest) (*model.WalletOperationResponse, error) {
				return nil, service.ErrInvalidOperation
			},
		}
		h := NewWalletHandler(svc)
		body := map[string]interface{}{
			"walletId":      walletID.String(),
			"operationType": "TRANSFER",
			"amount":        float64(10),
		}
		raw, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/wallet", bytes.NewReader(raw))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		h.ExecuteOperation(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want 400", rec.Code)
		}
	})

	t.Run("500 internal error", func(t *testing.T) {
		svc := &mockWalletService{
			executeOperation: func(context.Context, model.WalletOperationRequest) (*model.WalletOperationResponse, error) {
				return nil, errors.New("database error")
			},
		}
		h := NewWalletHandler(svc)
		body := map[string]interface{}{
			"walletId":      walletID.String(),
			"operationType": "DEPOSIT",
			"amount":        float64(10),
		}
		raw, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/wallet", bytes.NewReader(raw))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		h.ExecuteOperation(rec, req)
		if rec.Code != http.StatusInternalServerError {
			t.Errorf("status = %d, want 500", rec.Code)
		}
	})
}
