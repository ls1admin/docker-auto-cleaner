package docker

import (
	"context"
	"log/slog"
	"sort"
	"time"

	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
)

type DockerMonitor struct {
	ctx                context.Context
	cli                *client.Client
	queue              *ImageQueue
	storageThresholdGB int64
	interval           time.Duration
}

func NewDockerMonitor(ctx context.Context, threshold int64, interval time.Duration) *DockerMonitor {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	slog.Debug("Docker client", "version", cli.ClientVersion())

	if err != nil {
		slog.With("error", err).Error("Failed to create docker client")
	}
	slog.Info("Created docker client")

	return &DockerMonitor{cli: cli, ctx: ctx, queue: NewImageQueue(), storageThresholdGB: threshold, interval: interval}
}

func (m *DockerMonitor) Start() {
	slog.Debug("Start monitoring Docker Images according to LRU policy")
	slog.Info("Storage threshold set to (GB)", "storage_threshold", m.storageThresholdGB)
	m.initializingExistingImages()
	m.monitorDockerEvents()

	ticker := time.NewTicker(m.interval)
	// Make an endless loop
	for range ticker.C {
		// Run the scheduled tasks here
		m.RemoveDanglingVolumes(m.ctx)
		m.RemoveDanglingNetworks(m.ctx)
	}
	ticker.Stop()
}

func (m *DockerMonitor) initializingExistingImages() {
	images, err := m.cli.ImageList(context.Background(), image.ListOptions{All: true})
	if err != nil {
		slog.With("error", err).Error("Error listing images")
		return
	}

	// Sort images by creation date
	sort.Slice(images, func(i, j int) bool {
		return images[i].Created < images[j].Created
	})

	for _, img := range images {
		m.queue.Enqueue(ImageInfo{
			ID:       img.ID,
			LastUsed: time.Unix(img.Created, 0), // Approximate LastUsed by image creation time during initialization
			Size:     img.Size,                  // Convert bytes to GB
		})
	}
	slog.Info("Found existing images", "len", len(images))
}

func (m *DockerMonitor) monitorDockerEvents() {
	events, errorsCh := m.cli.Events(context.Background(), events.ListOptions{})
	for {
		select {
		case event := <-events:
			switch event.Type {
			case "image":
				if event.Action == "pull" {
					go m.handleImagePull(event.Actor.ID)
				}
			case "container":
				if event.Action == "start" {
					go m.handleContainerStart(event.Actor.ID)
				}
			}
		case err := <-errorsCh:
			slog.With("error", err).Error("Error receiving Docker events")
		}
	}
}
