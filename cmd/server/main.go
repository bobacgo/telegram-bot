package main

import (
	botapp "bot"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

const configPath = "config.yaml"

func main() {
	app := botapp.NewApp(configPath)
	app.Start()

	slog.Info("app started")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutdown signal received, shutting down...")

	app.Stop()
	slog.Info("shutdown complete")
}
