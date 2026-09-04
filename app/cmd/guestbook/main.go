package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"guestbook/internal/config"
	"guestbook/internal/server"
	"guestbook/internal/simulator"
	"guestbook/internal/storage"
	"guestbook/internal/web"
)

const shutdownTimeout = 10 * time.Second

func main() {
	if err := run(); err != nil {
		slog.Error("fatal error", "error", err)
		os.Exit(1)
	}
}

func run() (err error) {
	configuration, err := config.Load()
	if err != nil {
		return err
	}

	store, err := storage.New(configuration.StorageDirectory)
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if configuration.SimulatorEnabled {
		activitySimulator, err := simulator.New(store, configuration.SimulatorInterval)
		if err != nil {
			return err
		}
		go activitySimulator.Run(ctx)
		slog.Info("activity simulator enabled", "interval", configuration.SimulatorInterval)
	}

	httpServer := &http.Server{
		Addr:    configuration.Address,
		Handler: server.New(store, web.Assets).Handler(),
	}

	go func() {
		<-ctx.Done()
		slog.Info("shutting down")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			slog.Error("failed graceful shutdown", "error", err)
		}
	}()

	slog.Info("starting guestbook server",
		"address", configuration.Address,
		"storageDirectory", configuration.StorageDirectory)
	if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}
