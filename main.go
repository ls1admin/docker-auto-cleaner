package main

import (
	"context"
	"docker-auto-cleaner/docker"
	"time"

	"log/slog"

	"github.com/docker/docker/client"
)

func main() {
	slog.SetLogLoggerLevel(slog.LevelDebug)
	docker.Cli = initDockerClient()
	ctx := context.Background()

	docker.CleanContainersRunningLongerThan(ctx, time.Minute*30)
	docker.RemoveDanglingVolumes(ctx)

	go docker.StartImageMonitoring() // Start this in a background thread

	ticker := time.NewTicker(time.Hour)
	// Make an endless loop
	for range ticker.C {

		// TODO RUN THE SCHEDULED TASK HERE
	}
	ticker.Stop()
}

func initDockerClient() *client.Client {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	slog.Debug("Docker client", "version", cli.ClientVersion())

	if err != nil {
		slog.With("error", err).Error("Failed to create docker client")
	}
	slog.Info("Created docker client")

	return cli
}
