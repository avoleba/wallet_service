package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func NewRouter(wallet *WalletHandler) http.Handler {
	r := chi.NewRouter()

	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/wallet", wallet.ExecuteOperation)
		r.Get("/wallets/{WALLET_UUID}", wallet.GetBalance)
	})

	return r
}
