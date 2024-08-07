package docker

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	"github.com/stretchr/testify/suite"
)

type ContainerTestSuite struct {
	suite.Suite
	cli *client.Client
	dm  *DockerMonitor
}

func (suite *ContainerTestSuite) SetupSuite() {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	suite.NoError(err)
	suite.cli = cli
	suite.dm = NewDockerMonitor(context.Background(), 0, time.Hour)
	suite.NotNil(suite.dm)
}

func (suite *ContainerTestSuite) TearDownSuite() {
	suite.cli.Close()
}

func (suite *ContainerTestSuite) TearDownTest() {
	ctx := context.Background()
	volumes, err := suite.cli.VolumeList(ctx, volume.ListOptions{})
	suite.NoError(err)

	for _, v := range volumes.Volumes {
		err = suite.cli.VolumeRemove(ctx, v.Name, true)
		suite.NoError(err)
	}

	images, err := suite.cli.ImageList(ctx, image.ListOptions{})
	suite.NoError(err)

	for _, img := range images {
		_, err = suite.cli.ImageRemove(ctx, img.ID, image.RemoveOptions{Force: true})
		suite.NoError(err)
	}

	suite.dm.queue.Clear()
}

func (suite *ContainerTestSuite) TestContainerStartHandle() {
	first_timestamp := time.Now()

	suite.dm.queue.Enqueue(ImageInfo{ID: "1234", LastUsed: first_timestamp, Size: 0})

	time.Sleep(2 * time.Second)
	suite.dm.queue.UpdateLastUsed("1234")

	suite.Equal(1, suite.dm.queue.Len())
	suite.Greater(suite.dm.queue.items[0].LastUsed, first_timestamp)
}

func (suite *ContainerTestSuite) TestContainerDelete() {
	ctx := context.Background()
	// Define a threshold for the storage
	dm := NewDockerMonitor(ctx, 0, time.Hour)
	if dm == nil {
		suite.T().Errorf("Failed to create DockerMonitor")
	}

	response, err := suite.cli.ImagePull(ctx, "alpine:3.20", image.PullOptions{})
	suite.NoError(err)
	defer response.Close()
	io.Copy(io.Discard, response)

	// Create a container that will run forever
	resp, err := suite.cli.ContainerCreate(ctx, &container.Config{Image: "alpine:3.20", Cmd: []string{"tail", "-f", "/dev/null"}}, nil, nil, nil, "test_container")
	suite.NoError(err)

	err = suite.cli.ContainerStart(ctx, resp.ID, container.StartOptions{})
	suite.NoError(err)

	time.Sleep(2 * time.Second)
	ctrs, err := suite.cli.ContainerList(ctx, container.ListOptions{All: true})
	suite.NoError(err)
	suite.Equal(1, len(ctrs))

	err = dm.CleanContainersRunningLongerThan(1 * time.Second)
	suite.NoError(err)

	ctrs, err = suite.cli.ContainerList(ctx, container.ListOptions{All: true})
	suite.NoError(err)
	suite.Equal(0, len(ctrs))
}

func TestContainerTestSuite(t *testing.T) {
	suite.Run(t, new(ContainerTestSuite))
}
