package docker

import (
	"context"
	"log/slog"

	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/volume"
)

func (m *DockerMonitor) RemoveDanglingVolumes(ctx context.Context) error {
	slog.Debug("Start removing dangling volumes")

	// List all the volumes
	filters := filters.NewArgs()
	filters.Add("dangling", "true")
	volumeList, err := m.cli.VolumeList(ctx, volume.ListOptions{Filters: filters})
	slog.Debug("Found volumes", "num", len(volumeList.Volumes))

	if err != nil {
		slog.With("error", err).Error("Failed to list volumes")
		return err
	}

	//TODO: this is not working at the moment because the UsageData.RefCount can be nil. Not sure why though
	for _, v := range volumeList.Volumes {
		slog.Debug("Start processing volume", "name", v.Name)
		if err := m.cli.VolumeRemove(ctx, v.Name, true); err != nil {
			slog.With("error", err).Error("Failed to remove volume", "name", v.Name)
		} else {
			slog.Info("Successfully removed volume", "name", v.Name)
		}
	}

	return nil
}
