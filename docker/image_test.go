package docker

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/stretchr/testify/suite"
)

type ImageTestSuite struct {
	suite.Suite
	cli *client.Client
}

func (suite *ImageTestSuite) SetupSuite() {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	suite.NoError(err)
	suite.cli = cli
}

func (suite *ImageTestSuite) TearDownSuite() {
	suite.cli.Close()
}

func (suite *ImageTestSuite) TearDownTest() {
	ctx := context.Background()
	images, err := suite.cli.ImageList(ctx, image.ListOptions{})
	suite.NoError(err)

	for _, img := range images {
		_, err = suite.cli.ImageRemove(ctx, img.ID, image.RemoveOptions{Force: true})
		suite.NoError(err)
	}
}

func (suite *ImageTestSuite) TestImageDelete() {
	ctx := context.Background()
	// Define a threshold for the storage
	dm := NewDockerMonitor(ctx, 0, time.Hour)
	if dm == nil {
		suite.T().Errorf("Failed to create DockerMonitor")
	}

	default_images, err := suite.cli.ImageList(ctx, image.ListOptions{All: true})
	suite.NoError(err)

	response, err := suite.cli.ImagePull(ctx, "alpine:3.20", image.PullOptions{})
	suite.NoError(err)
	defer response.Close()
	io.Copy(io.Discard, response)

	images, err := suite.cli.ImageList(ctx, image.ListOptions{All: true})
	suite.NoError(err)
	suite.Equal(len(default_images)+1, len(images))

	// Handle the image pull which should be removed because of threshold 0
	dm.deleteImage(images[0].ID)

	post_images, err := suite.cli.ImageList(ctx, image.ListOptions{All: true})
	suite.NoError(err)
	suite.Equal(len(images)-1, len(post_images))
}

func (suite *ImageTestSuite) TestImageSize() {
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

	images, err := suite.cli.ImageList(ctx, image.ListOptions{All: true})
	suite.NoError(err)
	suite.Equal(1, len(images))

	// Handle the image pull which should be removed because of threshold 0
	actual_size := dm.getImageSize(images[0].ID)
	suite.Equal(images[0].Size, actual_size)
}

func (suite *ImageTestSuite) TestImagePullHandle() {
	ctx := context.Background()

	response, err := suite.cli.ImagePull(ctx, "alpine:3.20", image.PullOptions{})
	suite.NoError(err)
	defer response.Close()
	io.Copy(io.Discard, response)

	response2, err := suite.cli.ImagePull(ctx, "alpine:3.19", image.PullOptions{})
	suite.NoError(err)
	defer response2.Close()
	io.Copy(io.Discard, response2)

	images, err := suite.cli.ImageList(ctx, image.ListOptions{All: true})
	suite.NoError(err)
	suite.Equal(2, len(images))

	// Handle the image pull which should be removed because of threshold 0
	// Define a threshold for the storage
	dm := NewDockerMonitor(context.Background(), images[1].Size, time.Hour)
	if dm == nil {
		suite.T().Errorf("Failed to create DockerMonitor")
	}
	dm.handleImagePull(images[0].ID)
	dm.handleImagePull(images[1].ID)

	remaining_images, err := suite.cli.ImageList(ctx, image.ListOptions{All: true})
	suite.NoError(err)
	suite.Equal(1, len(remaining_images))
	suite.Equal(images[1].ID, remaining_images[0].ID)
}

func TestImageTestSuite(t *testing.T) {
	suite.Run(t, new(ImageTestSuite))
}
