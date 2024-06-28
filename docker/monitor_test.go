package docker

import (
	"context"
	"testing"

	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/stretchr/testify/assert"
)

func TestInitialization(t *testing.T) {
	ctx := context.Background()
	dm := NewDockerMonitor(ctx, 1)
	if dm == nil {
		t.Errorf("Failed to create DockerMonitor")
	}

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	assert.NoError(t, err)
	_, err = cli.ImagePull(ctx, "alpine:3.20", image.PullOptions{})
	assert.NoError(t, err)
	_, err = cli.ImagePull(ctx, "alpine:3.19", image.PullOptions{})
	assert.NoError(t, err)

	t.Log("Starting Docker Monitoring")
	dm.initializingExistingImages()

	assert.Equal(t, 2, imagesLRU.Len())
	assert.False(t, imagesLRU.IsEmpty())

	images, err := cli.ImageList(ctx, image.ListOptions{})
	assert.NoError(t, err)
	for _, img := range images {
		cli.ImageRemove(ctx, img.ID, image.RemoveOptions{Force: true})
	}

	imagesLRU.Clear()
	assert.True(t, imagesLRU.IsEmpty())
}
