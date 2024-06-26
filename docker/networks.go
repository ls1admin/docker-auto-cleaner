package docker

import (
	"context"
	"log/slog"

	"github.com/docker/docker/api/types/network"
)

func (m *DockerMonitor) RemoveDanglingNetworks(ctx context.Context) error {
	networks, err := m.cli.NetworkList(ctx, network.ListOptions{})
	if err != nil {
		return err
	}

	for _, n := range networks {
		if len(n.Containers) == 0 {
			if err := m.cli.NetworkRemove(ctx, n.ID); err != nil {
				slog.With("error", err).Error("Failed to remove network", "name", n.Name)
			} else {
				slog.Info("Successfully removed network", "name", n.Name)
			}
		}
	}

	return nil
}