package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"wallet-api/internal/model"
	"wallet-api/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// WalletServicer is the service interface used by the handler (for testing with mocks).
type WalletServicer interface {
	GetBalance(ctx context.Context, walletID uuid.UUID) (*model.Wallet, error)
	ExecuteOperation(ctx context.Context, req model.WalletOperationRequest) (*model.WalletOperationResponse, error)
}

type WalletHandler struct {
	svc WalletServicer
}

func NewWalletHandler(svc WalletServicer) *WalletHandler {
	return &WalletHandler{svc: svc}
}

func (h *WalletHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	walletIDStr := chi.URLParam(r, "WALLET_UUID")
	walletID, err := uuid.Parse(walletIDStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid wallet UUID")
		return
	}

	wallet, err := h.svc.GetBalance(r.Context(), walletID)
	if err != nil {
		if errors.Is(err, service.ErrWalletNotFound) {
			writeJSONError(w, http.StatusNotFound, "wallet not found")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"walletId": wallet.ID,
		"balance":  wallet.Balance,
	})
}

func (h *WalletHandler) ExecuteOperation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req model.WalletOperationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.WalletID == uuid.Nil {
		writeJSONError(w, http.StatusBadRequest, "walletId is required")
		return
	}

	result, err := h.svc.ExecuteOperation(r.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrWalletNotFound):
			writeJSONError(w, http.StatusNotFound, "wallet not found")
		case errors.Is(err, service.ErrInsufficientFunds):
			writeJSONError(w, http.StatusConflict, "insufficient funds for withdrawal")
		case errors.Is(err, service.ErrInvalidAmount):
			writeJSONError(w, http.StatusBadRequest, "amount must be positive")
		case errors.Is(err, service.ErrInvalidOperation):
			writeJSONError(w, http.StatusBadRequest, "operationType must be DEPOSIT or WITHDRAW")
		default:
			writeJSONError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
