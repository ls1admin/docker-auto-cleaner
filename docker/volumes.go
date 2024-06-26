package docker

import (
	"context"

	"github.com/docker/docker/api/types/volume"
	log "github.com/sirupsen/logrus"
)

func RemoveDanglingVolumes(ctx context.Context) error {
	log.Debug("Start removing dangling volumes")

	// List all the volumes
	volumeList, err := Cli.VolumeList(ctx, volume.ListOptions{})
	log.Debugf("Found %d volumes", len(volumeList.Volumes))

	if err != nil {
		log.WithError(err).Error("Failed to list volumes")
		return err
	}

	//TODO: this is not working at the moment because the UsageData.RefCount can be nil. Not sure why though
	for _, volume := range volumeList.Volumes {
		log.Debugf("Start processing volume: %+v", volume)
		if volume.UsageData.RefCount == 0 {
			log.Debugf("Found dangling volume: %s", volume.Name)
			if err := Cli.VolumeRemove(ctx, volume.Name, true); err != nil {
				log.Printf("Failed to remove volume %s: %v", volume.Name, err)
			} else {
				log.Printf("Successfully removed volume: %s", volume.Name)
			}
		}
	}

	return nil
}
