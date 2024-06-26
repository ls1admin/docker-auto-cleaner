package docker

import (
	"context"
	"time"

	"github.com/docker/docker/api/types"
	log "github.com/sirupsen/logrus"
)

func handleContainerStart(containerID string) {
	container, err := Cli.ContainerInspect(context.Background(), containerID)
	if err != nil {
		log.WithError(err).Warning("Error inspecting container")
		return
	}

	mu.Lock()
	defer mu.Unlock()

	// Update the usage timestamp for the image associated with this container
	for idx, imgInfo := range imagesLRU {
		if imgInfo.ID == container.Image {
			imgInfo.LastUsed = time.Now() // Update usage timestamp
			imagesLRU[idx] = imgInfo
			break
		}
	}
}

// Function that removes contaienr running for longer tha `minutes` minutes
// TODO: Handle stopped containers
func CleanContainersRunningLongerThan(ctx context.Context, duration time.Duration) error {
	log.Debug("Start cleaning containers")
	containers, err := Cli.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		log.WithError(err).Error("Failed to list containers")
		return err
	}
	log.Debugf("Found %d running containers in total", len(containers))

	threshold := time.Now().Add(-duration)
	log.Debugf("Cleanup threshold is set to: %v", threshold)

	for _, container := range containers {
		startTime := time.Unix(container.Created, 0)
		log.Debugf("Container %s started at %v", container.ID, startTime)
		if err != nil {
			log.Debugf("Failed to parse start time for container %s: %v", container.ID, err)
			continue
		}
		if startTime.Before(threshold) {
			if err := Cli.ContainerRemove(context.Background(), container.ID, types.ContainerRemoveOptions{Force: true}); err != nil {
				log.Infof("Failed to remove container %s: %v", container.ID, err)
			} else {
				log.Infof("Removed container %s, started at %s with ID %s", container.Names[0], startTime, container.ID)
			}
		}
	}

	return nil
}
