package docker

import (
	"context"
	"testing"
	"time"

	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	"github.com/stretchr/testify/assert"
)

func TestVolumeRemove(t *testing.T) {
	ctx := context.Background()
	// Define a threshold for the storage
	dm := NewDockerMonitor(ctx, 0, time.Hour)
	if dm == nil {
		t.Errorf("Failed to create DockerMonitor")
	}

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	assert.NoError(t, err)

	args := filters.NewArgs()
	args.Add("type", "custom")

	// Get the list of networks in advance
	default_volumes, err := cli.VolumeList(ctx, volume.ListOptions{})
	assert.NoError(t, err)

	_, err = cli.VolumeCreate(ctx, volume.CreateOptions{})
	assert.NoError(t, err)

	// Get the list of networks after creating a new dangling network
	volumes, err := cli.VolumeList(ctx, volume.ListOptions{})
	assert.NoError(t, err)
	assert.Equal(t, len(default_volumes.Volumes)+1, len(volumes.Volumes))

	err = dm.RemoveDanglingVolumes(ctx)
	assert.NoError(t, err)

	// Verify that the dangling network has been removed
	post_remove_volumes, err := cli.VolumeList(ctx, volume.ListOptions{})
	assert.NoError(t, err)
	assert.Equal(t, len(default_volumes.Volumes), len(post_remove_volumes.Volumes))
}
