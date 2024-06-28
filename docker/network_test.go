package docker

import (
	"context"
	"testing"
	"time"

	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/stretchr/testify/assert"
)

func TestNetworkRemove(t *testing.T) {
	ctx := context.Background()
	// Define a threshold for the storage
	dm := NewDockerMonitor(ctx, 0, time.Hour)
	if dm == nil {
		t.Errorf("Failed to create DockerMonitor")
	}

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	assert.NoError(t, err)

	_, err = cli.NetworksPrune(ctx, filters.Args{})
	assert.NoError(t, err)

	args := filters.NewArgs()
	args.Add("type", "custom")

	// Get the list of networks in advance
	default_networks, err := cli.NetworkList(ctx, network.ListOptions{Filters: args})
	assert.NoError(t, err)

	_, err = cli.NetworkCreate(ctx, "test-network", network.CreateOptions{})
	assert.NoError(t, err)

	// Get the list of networks after creating a new dangling network
	networks, err := cli.NetworkList(ctx, network.ListOptions{Filters: args})
	assert.NoError(t, err)
	assert.Equal(t, len(default_networks)+1, len(networks))

	err = dm.RemoveDanglingNetworks(ctx)
	assert.NoError(t, err)

	// Verify that the dangling network has been removed
	post_remove_networks, err := cli.NetworkList(ctx, network.ListOptions{Filters: args})
	assert.NoError(t, err)
	assert.Equal(t, len(default_networks), len(post_remove_networks))
}
