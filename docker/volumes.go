package docker

import (
	"context"
	"log/slog"

	"github.com/docker/docker/api/types/volume"
)

func RemoveDanglingVolumes(ctx context.Context) error {
	slog.Debug("Start removing dangling volumes")

	// List all the volumes
	volumeList, err := Cli.VolumeList(ctx, volume.ListOptions{})
	slog.Debug("Found volumes", "num", len(volumeList.Volumes))

	if err != nil {
		slog.With("error", err).Error("Failed to list volumes")
		return err
	}

	//TODO: this is not working at the moment because the UsageData.RefCount can be nil. Not sure why though
	for _, volume := range volumeList.Volumes {
		slog.Debug("Start processing volume", "name", volume)
		if volume.UsageData.RefCount == 0 {
			slog.Debug("Found dangling volume", "name", volume.Name)
			if err := Cli.VolumeRemove(ctx, volume.Name, true); err != nil {
				slog.With("error", err).Error("Failed to remove volume", "name", volume.Name)
			} else {
				slog.Info("Successfully removed volume", "name", volume.Name)
			}
		}
	}

	return nil
}
