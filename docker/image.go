package docker

import (
	"context"
	"log/slog"
	"time"

	"github.com/docker/docker/api/types/image"
)

type ImageInfo struct {
	ID       string
	LastUsed time.Time
	Size     float64 // Size in GB
}

func (m *DockerMonitor) handleImagePull(imageID string) {
	img_info := ImageInfo{
		ID:       imageID,
		LastUsed: time.Now(),
	}
	img_info.Size = m.getImageSize(img_info.ID)
	imagesLRU.Enqueue(img_info)

	totalSize := imagesLRU.TotalSize()
	for totalSize > m.storageThresholdGB && !imagesLRU.IsEmpty() {
		oldestImage := imagesLRU.Dequeue()
		err := m.deleteImage(oldestImage.ID)
		if err != nil {
			slog.With("error", err).Error("Failed to remove image")
		}
	}
}

func (m *DockerMonitor) getImageSize(ImageID string) float64 {
	img, _, err := m.cli.ImageInspectWithRaw(context.Background(), ImageID)
	if err != nil {
		slog.With("error", err).Error("Error inspecting image")
		return 0
	}
	return float64(img.Size) / (1024 * 1024 * 1024) // Convert bytes to GB
}

func (m *DockerMonitor) deleteImage(ImageID string) error {
	_, err := m.cli.ImageRemove(context.Background(), ImageID, image.RemoveOptions{})
	if err != nil {
		slog.With("error", err).Error("Error removing image")
		return err
	}
	slog.With("imageID", ImageID).With("size", m.getImageSize(ImageID)).Debug("Image removed")
	return nil
}
