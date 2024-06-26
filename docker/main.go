package docker

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	log "github.com/sirupsen/logrus"
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
	log.Debug("Start monitoring Docker Images according to LRU policy")
	log.Infof("Storage threshold set to %d GB", storageThresholdGB)
	initializingExistingImages()
	monitorDockerEvents()
}

func initializingExistingImages() {
	images, err := Cli.ImageList(context.Background(), types.ImageListOptions{})
	if err != nil {
		log.WithError(err).Warning("Error listing images")
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
	log.Debugf("Found existing images %v", imagesLRU)
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
			log.WithError(err).Warning("Error receiving Docker events")
		}
	}
}
