package docker

import (
	"context"
	"log/slog"
	"sort"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
)

const (
	// TODO make this configurable
	storageThresholdGB = 10
)

type ImageInfo struct {
	ID       string
	LastUsed time.Time
	Size     float64 // Size in GB
}

var (
	imagesLRU = []ImageInfo{} // Array storing all monitored images
	mu        sync.Mutex      // Mutex to avoid race condition when pulling multiple images
	Cli       *client.Client
)

func StartImageMonitoring() {
	slog.Debug("Start monitoring Docker Images according to LRU policy")
	slog.Info("Storage threshold set to %d GB", storageThresholdGB)
	initializingExistingImages()
	monitorDockerEvents()
}

func initializingExistingImages() {
	images, err := Cli.ImageList(context.Background(), image.ListOptions{})
	if err != nil {
		slog.With("error", err).Error("Error listing images")
		return
	}

	// Sort images by creation date
	sort.Slice(images, func(i, j int) bool {
		return images[i].Created < images[j].Created
	})

	mu.Lock()
	defer mu.Unlock()

	for _, img := range images {
		imagesLRU = append(imagesLRU, ImageInfo{
			ID:       img.ID,
			LastUsed: time.Unix(img.Created, 0),                // Approximate LastUsed by image creation time during initialization
			Size:     float64(img.Size) / (1024 * 1024 * 1024), // Convert bytes to GB
		})
	}
	slog.Debug("Found existing images %v", imagesLRU)
}

func monitorDockerEvents() {
	events, errorsCh := Cli.Events(context.Background(), types.EventsOptions{})

	for {
		select {
		case event := <-events:
			switch event.Type {
			case "image":
				if event.Action == "pull" {
					handleImagePull(event.Actor.ID)
				}
			case "container":
				if event.Action == "start" {
					handleContainerStart(event.Actor.ID)
				}
			}
		case err := <-errorsCh:
			slog.With("error", err).Error("Error receiving Docker events")
		}
	}
}
