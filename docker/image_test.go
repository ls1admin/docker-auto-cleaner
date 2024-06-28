package docker

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/stretchr/testify/assert"
)

func TestImageDelete(t *testing.T) {
	ctx := context.Background()
	// Define a threshold for the storage
	dm := NewDockerMonitor(ctx, 0, time.Hour)
	if dm == nil {
		t.Errorf("Failed to create DockerMonitor")
	}

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	assert.NoError(t, err)
	response, err := cli.ImagePull(ctx, "alpine:3.20", image.PullOptions{})
	assert.NoError(t, err)
	defer response.Close()
	io.Copy(io.Discard, response)

	images, err := cli.ImageList(ctx, image.ListOptions{All: true})
	assert.NoError(t, err)
	assert.Equal(t, 1, len(images))

	// Handle the image pull which should be removed because of threshold 0
	dm.deleteImage(images[0].ID)

	images, err = cli.ImageList(ctx, image.ListOptions{All: true})
	assert.NoError(t, err)
	assert.Equal(t, 0, len(images))
}

func TestImageSize(t *testing.T) {
	ctx := context.Background()
	// Define a threshold for the storage
	dm := NewDockerMonitor(ctx, 0, time.Hour)
	if dm == nil {
		t.Errorf("Failed to create DockerMonitor")
	}

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	assert.NoError(t, err)
	response, err := cli.ImagePull(ctx, "alpine:3.20", image.PullOptions{})
	assert.NoError(t, err)
	defer response.Close()
	io.Copy(io.Discard, response)

	images, err := cli.ImageList(ctx, image.ListOptions{All: true})
	assert.NoError(t, err)
	assert.Equal(t, 1, len(images))

	// Handle the image pull which should be removed because of threshold 0
	actual_size := dm.getImageSize(images[0].ID)
	assert.Equal(t, images[0].Size, actual_size)

	_, err = cli.ImageRemove(ctx, images[0].ID, image.RemoveOptions{Force: true})
	assert.NoError(t, err)
}

func TestImagePullHandle(t *testing.T) {
	ctx := context.Background()

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	assert.NoError(t, err)
	response, err := cli.ImagePull(ctx, "alpine:3.20", image.PullOptions{})
	assert.NoError(t, err)
	defer response.Close()
	io.Copy(io.Discard, response)

	response2, err := cli.ImagePull(ctx, "alpine:3.19", image.PullOptions{})
	assert.NoError(t, err)
	defer response2.Close()
	io.Copy(io.Discard, response2)

	images, err := cli.ImageList(ctx, image.ListOptions{All: true})
	assert.NoError(t, err)
	assert.Equal(t, 2, len(images))

	// Handle the image pull which should be removed because of threshold 0
	// Define a threshold for the storage
	dm := NewDockerMonitor(context.Background(), images[1].Size, time.Hour)
	if dm == nil {
		t.Errorf("Failed to create DockerMonitor")
	}
	dm.handleImagePull(images[0].ID)
	dm.handleImagePull(images[1].ID)

	remaining_images, err := cli.ImageList(ctx, image.ListOptions{All: true})
	assert.NoError(t, err)
	assert.Equal(t, 1, len(remaining_images))
	assert.Equal(t, images[1].ID, remaining_images[0].ID)

	_, err = cli.ImageRemove(ctx, remaining_images[0].ID, image.RemoveOptions{Force: true})
	assert.NoError(t, err)
}
