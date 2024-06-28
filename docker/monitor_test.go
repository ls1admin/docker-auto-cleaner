package docker

import (
	"context"
	"testing"
	"time"

	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/stretchr/testify/suite"
)

type MonitorTestSuite struct {
	suite.Suite
	cli *client.Client
}

func (suite *MonitorTestSuite) SetupSuite() {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	suite.NoError(err)
	suite.cli = cli
}

func (suite *MonitorTestSuite) TearDownSuite() {
	suite.cli.Close()
}

func (suite *MonitorTestSuite) TearDownTest() {
	ctx := context.Background()
	images, err := suite.cli.ImageList(ctx, image.ListOptions{})
	suite.NoError(err)
	for _, img := range images {
		suite.cli.ImageRemove(ctx, img.ID, image.RemoveOptions{Force: true})
	}

	imagesLRU.Clear()
	suite.True(imagesLRU.IsEmpty())
}

func (suite *MonitorTestSuite) TestInitialization() {
	ctx := context.Background()
	dm := NewDockerMonitor(ctx, 1, time.Hour)
	if dm == nil {
		suite.T().Errorf("Failed to create DockerMonitor")
	}

	_, err := suite.cli.ImagePull(ctx, "alpine:3.20", image.PullOptions{})
	suite.NoError(err)
	_, err = suite.cli.ImagePull(ctx, "alpine:3.19", image.PullOptions{})
	suite.NoError(err)

	dm.initializingExistingImages()

	suite.Equal(2, imagesLRU.Len())
	suite.False(imagesLRU.IsEmpty())
}

func TestMonitorSuite(t *testing.T) {
	suite.Run(t, new(MonitorTestSuite))
}
