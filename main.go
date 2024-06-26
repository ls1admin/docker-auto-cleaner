package main

import (
	"context"
	"docker-auto-cleaner/docker"
	"os"

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

	ctx := context.Background()
	monitor := docker.NewDockerMonitor(ctx, threshold)
	monitor.Start()
}
