package main

import (
	"bot/internal/app"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

const configPath = "config.yaml"

func main() {
	application := app.NewApp(configPath)
	application.Start()

	slog.Info("app started")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutdown signal received, shutting down...")

	application.Stop()
	slog.Info("shutdown complete")
}
