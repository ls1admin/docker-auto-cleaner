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
	Size     int64 // Size in GB
}

func (m *DockerMonitor) handleImagePull(imageID string) {
	img_info := ImageInfo{
		ID:       imageID,
		LastUsed: time.Now(),
	}
	img_info.Size = m.getImageSize(img_info.ID)
	m.queue.Enqueue(img_info)

	totalSize := m.queue.TotalSize()
	for totalSize > m.storageThresholdGB && !m.queue.IsEmpty() {
		oldestImage := m.queue.Dequeue()
		err := m.deleteImage(oldestImage.ID)
		if err != nil {
			slog.With("error", err).Error("Failed to remove image")
		} else {
			slog.With("imageID", oldestImage.ID).With("size", oldestImage.Size).Info("Image removed")
		}
		totalSize -= oldestImage.Size
	}
}

func (m *DockerMonitor) getImageSize(ImageID string) int64 {
	img, _, err := m.cli.ImageInspectWithRaw(context.Background(), ImageID)
	if err != nil {
		slog.With("error", err).Error("Error inspecting image")
		return 0
	}
	return img.Size
}

func (m *DockerMonitor) deleteImage(ImageID string) error {
	_, err := m.cli.ImageRemove(context.Background(), ImageID, image.RemoveOptions{Force: true})
	if err != nil {
		slog.With("error", err).Error("Error removing image")
		return err
	}
	return nil
}
