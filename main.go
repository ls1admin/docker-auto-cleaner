package main

import (
	"context"
	"time"

	"github.com/docker/docker/client"
	log "github.com/sirupsen/logrus"
)

var cli *client.Client

func main() {
	log.SetLevel(log.DebugLevel)
	cli = initDockerClient()
	ctx := context.Background()

	cleanContainersRunningLongerThan(ctx, 1)
	removeDanglingVolumes(ctx)
	return // TODO: Remove this again - Only for quick testing

	go startImageMonitoring() // Start this in a background thread

	// Make an endless loop
	for {
		now := time.Now()

		nextHour := time.Date(now.Year(), now.Month(), now.Day(), now.Hour()+1, 0, 0, 0, now.Location())

		// Calculate the duration until the next full hour
		duration := nextHour.Sub(now)

		time.Sleep(duration)

		// TODO RUN THE SCHEDULED TASK HERE
	}
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
