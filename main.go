package main

import (
	"context"
	"docker-auto-cleaner/docker"
	"os"

	"log/slog"
)

func main() {
	if os.Getenv("DEBUG") == "true" {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}
	ctx := context.Background()
	monitor := docker.NewDockerMonitor(ctx)
	monitor.Start()
}
