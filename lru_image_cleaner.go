package main

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
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
)

func startImageMonitoring() {
	log.Debug("Start monitoring Docker Images according to LRU policy")
	log.Infof("Storage threshold set to %d GB", storageThresholdGB)
	initializingExistingImages()
	monitorDockerEvents()
}

func initializingExistingImages() {
	images, err := cli.ImageList(context.Background(), types.ImageListOptions{})
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
	events, errorsCh := cli.Events(context.Background(), types.EventsOptions{})

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

func handleContainerStart(containerID string) {
	container, err := cli.ContainerInspect(context.Background(), containerID)
	if err != nil {
		log.WithError(err).Warning("Error inspecting container")
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

func getImageSize(imageID string) float64 {
	img, _, err := cli.ImageInspectWithRaw(context.Background(), imageID)
	if err != nil {
		log.WithError(err).Warning("Error inspecting image")
		return 0
	}
	return float64(img.Size) / (1024 * 1024 * 1024) // Convert bytes to GB
}

func deleteImage(imageID string) error {
	_, err := cli.ImageRemove(context.Background(), imageID, types.ImageRemoveOptions{})
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
