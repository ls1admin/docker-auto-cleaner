package docker

import (
	"context"
	"testing"
	"time"

	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	"github.com/stretchr/testify/suite"
)

type VolumeTestSuite struct {
	suite.Suite
	cli *client.Client
}

func (suite *VolumeTestSuite) SetupSuite() {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	suite.NoError(err)
	suite.cli = cli
}

func (suite *VolumeTestSuite) TearDownSuite() {
	suite.cli.Close()
	imagesLRU.Clear()
}

func (suite *VolumeTestSuite) TearDownTest() {
	ctx := context.Background()
	volumes, err := suite.cli.VolumeList(ctx, volume.ListOptions{})
	suite.NoError(err)

	for _, v := range volumes.Volumes {
		err = suite.cli.VolumeRemove(ctx, v.Name, true)
		suite.NoError(err)
	}
}

func (suite *VolumeTestSuite) TestVolumeRemove() {
	ctx := context.Background()
	// Define a threshold for the storage
	dm := NewDockerMonitor(ctx, 0, time.Hour)
	if dm == nil {
		suite.T().Errorf("Failed to create DockerMonitor")
	}

	args := filters.NewArgs()
	args.Add("type", "custom")

	// Get the list of networks in advance
	default_volumes, err := suite.cli.VolumeList(ctx, volume.ListOptions{})
	suite.NoError(err)

	_, err = suite.cli.VolumeCreate(ctx, volume.CreateOptions{})
	suite.NoError(err)

	// Get the list of networks after creating a new dangling network
	volumes, err := suite.cli.VolumeList(ctx, volume.ListOptions{})
	suite.NoError(err)
	suite.Equal(len(default_volumes.Volumes)+1, len(volumes.Volumes))

	err = dm.RemoveDanglingVolumes(ctx)
	suite.NoError(err)

	// Verify that the dangling network has been removed
	post_remove_volumes, err := suite.cli.VolumeList(ctx, volume.ListOptions{})
	suite.NoError(err)
	suite.Equal(len(default_volumes.Volumes), len(post_remove_volumes.Volumes))
}

func TestVolumeTestSuite(t *testing.T) {
	suite.Run(t, new(VolumeTestSuite))
}
