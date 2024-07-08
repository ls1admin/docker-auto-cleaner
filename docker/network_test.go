package docker

import (
	"context"
	"testing"
	"time"

	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/stretchr/testify/suite"
)

type NetworkTestSuite struct {
	suite.Suite
	cli *client.Client
}

func (suite *NetworkTestSuite) SetupSuite() {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	suite.NoError(err)
	suite.cli = cli
}

func (suite *NetworkTestSuite) TearDownSuite() {
	suite.cli.Close()
}

func (suite *NetworkTestSuite) TearDownTest() {
	ctx := context.Background()

	args := filters.NewArgs()
	args.Add("type", "custom")

	networks, err := suite.cli.NetworkList(ctx, network.ListOptions{Filters: args})
	suite.NoError(err)

	for _, n := range networks {
		err = suite.cli.NetworkRemove(ctx, n.ID)
		suite.NoError(err)
	}
}

func (suite *NetworkTestSuite) TestNetworkRemove() {
	ctx := context.Background()
	// Define a threshold for the storage
	dm := NewDockerMonitor(ctx, 0, time.Hour)
	if dm == nil {
		suite.T().Errorf("Failed to create DockerMonitor")
	}

	_, err := suite.cli.NetworksPrune(ctx, filters.Args{})
	suite.NoError(err)

	args := filters.NewArgs()
	args.Add("type", "custom")

	// Get the list of networks in advance
	default_networks, err := suite.cli.NetworkList(ctx, network.ListOptions{Filters: args})
	suite.NoError(err)

	_, err = suite.cli.NetworkCreate(ctx, "test-network", network.CreateOptions{})
	suite.NoError(err)

	// Get the list of networks after creating a new dangling network
	networks, err := suite.cli.NetworkList(ctx, network.ListOptions{Filters: args})
	suite.NoError(err)
	suite.Equal(len(default_networks)+1, len(networks))

	err = dm.RemoveDanglingNetworks(ctx)
	suite.NoError(err)

	// Verify that the dangling network has been removed
	post_remove_networks, err := suite.cli.NetworkList(ctx, network.ListOptions{Filters: args})
	suite.NoError(err)
	suite.Equal(len(default_networks), len(post_remove_networks))
}

func TestNetworkTestSuite(t *testing.T) {
	suite.Run(t, new(NetworkTestSuite))
}
