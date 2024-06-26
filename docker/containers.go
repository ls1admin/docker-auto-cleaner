package docker

import (
	"context"
	"log/slog"
	"time"

	"github.com/docker/docker/api/types/container"
)

func handleContainerStart(containerID string) {
	container, err := Cli.ContainerInspect(context.Background(), containerID)
	if err != nil {
		slog.With("error", err).Error("Error inspecting container")
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
	slog.Debug("Start cleaning containers")
	containers, err := Cli.ContainerList(context.Background(), container.ListOptions{})
	if err != nil {
		slog.With("error", err).Error("Failed to list containers")
		return err
	}
	slog.Debug("Found running containers in total", "number", len(containers))

	threshold := time.Now().Add(-duration)
	slog.Debug("Cleanup threshold is set to", "threshold", threshold)

	for _, cont := range containers {
		startTime := time.Unix(cont.Created, 0)
		slog.Debug("Container started ", "id", cont.ID, "time", startTime)
		if startTime.Before(threshold) {
			if err := Cli.ContainerRemove(context.Background(), cont.ID, container.RemoveOptions{}); err != nil {
				slog.With("error", err).Error("Failed to remove container", "ID", cont.ID)
			} else {
				slog.Info("Removed container", "name", cont.Names[0], "startTime", startTime, "ID", cont.ID)
			}
		}
	}

	return nil
}
