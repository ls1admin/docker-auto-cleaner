package docker

import (
	"context"
	"sort"
	"time"

	"github.com/docker/docker/api/types"
	log "github.com/sirupsen/logrus"
)

func handleImagePull(imageID string) {
	size := getImageSize(imageID)
	now := time.Now()

	mu.Lock()
	defer mu.Unlock()

	// Mark image as used now
	imagesLRU = append(imagesLRU, ImageInfo{
		ID:       imageID,
		LastUsed: now,
		Size:     size,
	})

	checkAndFreeSpace()
}

func getImageSize(imageID string) float64 {
	img, _, err := Cli.ImageInspectWithRaw(context.Background(), imageID)
	if err != nil {
		log.WithError(err).Warning("Error inspecting image")
		return 0
	}
	return float64(img.Size) / (1024 * 1024 * 1024) // Convert bytes to GB
}

func deleteImage(imageID string) error {
	_, err := Cli.ImageRemove(context.Background(), imageID, types.ImageRemoveOptions{})
	if err != nil {
		log.WithError(err).Warning("Error removing image")
		return err
	}
	log.WithField("imageID", imageID).WithField("size", getImageSize(imageID)).Debug("Image removed")
	return nil
}

func checkAndFreeSpace() {
	totalSize := float64(0)
	for _, img := range imagesLRU {
		totalSize += img.Size
	}

	// Make sure no new image is inserted during cleanup
	mu.Lock()
	defer mu.Unlock()

	for totalSize > storageThresholdGB && len(imagesLRU) > 0 {
		// Sorting by LastUsed to get the least recently used image
		sort.SliceStable(imagesLRU, func(i, j int) bool {
			return imagesLRU[i].LastUsed.Before(imagesLRU[j].LastUsed)
		})
		oldestImage := imagesLRU[0]
		err := deleteImage(oldestImage.ID)
		if err == nil {
			totalSize -= oldestImage.Size
			imagesLRU = imagesLRU[1:] // Remove the oldest image from the list
		}
	}
}
