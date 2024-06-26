package docker

import (
	"context"
	"log/slog"
	"time"

	"github.com/docker/docker/api/types/container"
)

func (m *DockerMonitor) handleContainerStart(containerID string) {
	container, err := cli.ContainerInspect(context.Background(), containerID)
	if err != nil {
		slog.With("error", err).Error("Error inspecting container")
		return
	}

	imagesLRU.UpdateLastUsed(container.Image)
}

// Function that removes contaienr running for longer tha `minutes` minutes
func (m *DockerMonitor) CleanContainersRunningLongerThan(duration time.Duration) error {
	slog.Debug("Start cleaning containers")
	containers, err := m.cli.ContainerList(m.ctx, container.ListOptions{All: true})
	if err != nil {
		slog.With("error", err).Error("Failed to list containers")
		return err
	}
	slog.Debug("Found running containers in total", "number", len(containers))

	threshold := time.Now().Add(-duration)
	slog.Debug("Cleanup threshold is set to", "time_threshold", threshold)

	for _, cont := range containers {
		// TODO handle case when the application itself is running inside a container
		startTime := time.Unix(cont.Created, 0)
		slog.Debug("Container started ", "id", cont.ID, "time", startTime)
		if startTime.Before(threshold) {
			if err := m.cli.ContainerRemove(m.ctx, cont.ID, container.RemoveOptions{}); err != nil {
				slog.With("error", err).Error("Failed to remove container", "ID", cont.ID)
			} else {
				slog.Info("Removed container", "name", cont.Names[0], "startTime", startTime, "ID", cont.ID)
			}
		}
	}

	return nil
}
