package main

import (
	"context"
	"docker-auto-cleaner/docker"
	"time"

	"github.com/docker/docker/client"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetLevel(log.DebugLevel)
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
	log.Debugf("Docker client version: %+v", cli)

	if err != nil {
		log.WithError(err).Panic("Failed to create docker client")
	}
	log.Info("Created docker client")

	return cli
}
