package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Addr         string
	DSN          string
	DBMaxOpen    int
	DBMaxIdle    int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

func Load() *Config {
	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		dsn = "postgres://localhost:5432/wallet?sslmode=disable"
	}
	dbMaxOpen := 200
	if v := os.Getenv("DB_MAX_OPEN_CONNS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			dbMaxOpen = n
		}
	}
	dbMaxIdle := 50
	if v := os.Getenv("DB_MAX_IDLE_CONNS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			dbMaxIdle = n
		}
	}
	readTimeout := 15 * time.Second
	if v := os.Getenv("HTTP_READ_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil && d > 0 {
			readTimeout = d
		}
	}
	writeTimeout := 15 * time.Second
	if v := os.Getenv("HTTP_WRITE_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil && d > 0 {
			writeTimeout = d
		}
	}
	return &Config{
		Addr:         addr,
		DSN:          dsn,
		DBMaxOpen:    dbMaxOpen,
		DBMaxIdle:    dbMaxIdle,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
	}
}
