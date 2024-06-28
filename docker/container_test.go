package docker

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/stretchr/testify/assert"
)

func TestContainerStartHandle(t *testing.T) {
	first_timestamp := time.Now()
	imagesLRU.Enqueue(ImageInfo{ID: "1234", LastUsed: first_timestamp, Size: 0})

	time.Sleep(2 * time.Second)
	imagesLRU.UpdateLastUsed("1234")

	assert.Equal(t, 1, imagesLRU.Len())
	assert.Greater(t, imagesLRU.items[0].LastUsed, first_timestamp)
}

func TestContainerDelete(t *testing.T) {
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

	resp, err := cli.ContainerCreate(ctx, &container.Config{Image: "alpine:3.20", Cmd: []string{"tail", "-f", "/dev/null"}}, nil, nil, nil, "test_container")
	assert.NoError(t, err)

	err = cli.ContainerStart(ctx, resp.ID, container.StartOptions{})
	assert.NoError(t, err)

	time.Sleep(2 * time.Second)
	ctrs, err := cli.ContainerList(ctx, container.ListOptions{All: true})
	assert.NoError(t, err)
	assert.Equal(t, 1, len(ctrs))

	err = dm.CleanContainersRunningLongerThan(1 * time.Second)
	assert.NoError(t, err)

	ctrs, err = cli.ContainerList(ctx, container.ListOptions{All: true})
	assert.NoError(t, err)
	assert.Equal(t, 0, len(ctrs))
}
