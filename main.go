package main

import (
	"context"
	"docker-auto-cleaner/docker"
	"os"
	"time"

	"log/slog"

	"github.com/docker/go-units"
)

func main() {
	if os.Getenv("DEBUG") == "true" {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}

	threshold, err := units.FromHumanSize(os.Getenv("THRESHOLD"))
	if err != nil {
		slog.With("error", err).Error("Failed to parse threshold")
		threshold = 10 * units.GiB
		slog.Warn("Setting threshold to default value", "threshold", threshold)
	}

	interval, err := time.ParseDuration(os.Getenv("INTERVAL"))
	if err != nil {
		slog.With("error", err).Error("Failed to parse interval")
		interval = 1 * time.Hour
		slog.Warn("Setting interval to default value", "interval", interval)
	}

	ctx := context.Background()
	monitor := docker.NewDockerMonitor(ctx, threshold, interval)
	monitor.Start()
}
