package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/lucashmsilva/rinha-2025-api-go/internal/handlers"
	"github.com/lucashmsilva/rinha-2025-api-go/internal/infra/config"
	"github.com/lucashmsilva/rinha-2025-api-go/internal/infra/database"
	"github.com/lucashmsilva/rinha-2025-api-go/internal/service"
	"github.com/lucashmsilva/rinha-2025-api-go/internal/workers"
)

func main() {
	cfg := config.LoadConfig()
	db := database.LoadConnections(cfg.DbConnCfg)
	procService := service.NewProcessorService(cfg)
	healthChecker := workers.NewHealthChecker(db, procService)
	mux := http.NewServeMux()

	slog.SetLogLoggerLevel(slog.Level(cfg.LogLevel))

	if cfg.StartHealthChecker {
		healthChecker.StartHealthChecker()
	}

	mux.Handle("POST /payments", handlers.NewPaymentCreateHandler(db, procService).Handle())
	mux.Handle("GET /payments-summary", handlers.NewPaymentGetSummaryHandler(db).Handle())
	mux.Handle("POST /purge-payments", handlers.NewPaymentsPurgeHandler(db, procService).Handle())

	server := &http.Server{
		Addr:    fmt.Sprintf(":%v", cfg.Port),
		Handler: mux,
	}

	go func() {
		slog.Info("server started", "port", cfg.Port)
		server.ListenAndServe()
	}()

	shutdownServer(server)

	db.Conn.Close()
	slog.Info("bye")
}

func shutdownServer(server *http.Server) {
	slog.Info("listening for shutdown signals")
	rootCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-rootCtx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	slog.Info("server shutting down")
	server.Shutdown(shutdownCtx)
}
