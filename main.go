package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/joho/godotenv"
	"wallet-api/config"
	"wallet-api/internal/handler"
	"wallet-api/internal/repository"
	"wallet-api/internal/service"

	_ "github.com/lib/pq"
)

func main() {
	_ = godotenv.Load("config.env")
	_ = godotenv.Load(".env")
	if ex, err := os.Executable(); err == nil {
		_ = godotenv.Load(filepath.Join(filepath.Dir(ex), "config.env"))
	}

	cfg := config.Load()

	db, err := sql.Open("postgres", cfg.DSN)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()
	db.SetMaxOpenConns(cfg.DBMaxOpen)
	db.SetMaxIdleConns(cfg.DBMaxIdle)

	walletRepo := repository.NewWalletRepository(db)
	walletSvc := service.NewWalletService(walletRepo)
	walletHandler := handler.NewWalletHandler(walletSvc)
	router := handler.NewRouter(walletHandler)

	srv := &http.Server{
		Addr:         cfg.Addr,
		Handler:      router,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  60 * time.Second,
	}
	log.Printf("server starting on %s", cfg.Addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server: %v", err)
	}
}
