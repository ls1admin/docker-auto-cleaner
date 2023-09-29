package main

import (
	"context"
	"time"

	"github.com/docker/docker/api/types"
	log "github.com/sirupsen/logrus"
)

// Function that removes contaienr running for longer tha `minutes` minutes
// TODO: Handle stopped containers
func cleanContainersRunningLongerThan(ctx context.Context, minutes uint) error {
	log.Debug("Start cleaning containers")
	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		log.WithError(err).Error("Failed to list containers")
		return err
	}
	log.Debugf("Found %d running containers in total", len(containers))

	threshold := time.Now().Add(-time.Duration(minutes) * time.Minute)
	log.Debugf("Cleanup threshold is set to: %v", threshold)

	for _, container := range containers {
		startTime := time.Unix(container.Created, 0)
		log.Debugf("Container %s started at %v", container.ID, startTime)
		if err != nil {
			log.Debugf("Failed to parse start time for container %s: %v", container.ID, err)
			continue
		}
		if startTime.Before(threshold) {
			if err := cli.ContainerRemove(context.Background(), container.ID, types.ContainerRemoveOptions{Force: true}); err != nil {
				log.Infof("Failed to remove container %s: %v", container.ID, err)
			} else {
				log.Infof("Removed container %s, started at %s with ID %s", container.Names[0], startTime, container.ID)
			}
		}
	}

	return nil

}
