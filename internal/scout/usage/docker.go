package usage

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/src-cli/internal/scout/style"
)

func Docker(ctx context.Context, client client.Client, opts ...Option) error {
	cfg := &Config{
		namespace:    "default",
		docker:       true,
		pod:          "",
		container:    "",
		spy:          false,
		dockerClient: &client,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	containers, err := cfg.dockerClient.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		return errors.Wrap(err, "could not get list of containers")
	}

	return renderDockerUsageTable(ctx, cfg, containers)
}

// renderDockerUsageTable renders a table displaying CPU and memory usage for Docker containers.
func renderDockerUsageTable(ctx context.Context, cfg *Config, containers []types.Container) error {
	columns := []table.Column{
		{Title: "Container", Width: 20},
		{Title: "Cores", Width: 10},
		{Title: "Usage", Width: 10},
		{Title: "Memory", Width: 10},
		{Title: "Usage", Width: 10},
	}
	rows := []table.Row{}

	for _, container := range containers {
		containerInfo, err := cfg.dockerClient.ContainerInspect(ctx, container.ID)
		if err != nil {
			return errors.Wrap(err, "failed to get container info")
		}

		stats, err := cfg.dockerClient.ContainerStats(ctx, container.ID, false)
		if err != nil {
			return errors.Wrap(err, "could not get container stats")
		}
		defer stats.Body.Close()

		var usage types.StatsJSON
		if err := json.NewDecoder(stats.Body).Decode(&usage); err != nil {
			return errors.Wrap(err, "could not decode container stats")
		}

		row := makeDockerUsageRow(usage, containerInfo)
		rows = append(rows, row)
	}

	style.ResourceTable(columns, rows)
	return nil
}

// makeDockerUsageRow generates a table row displaying CPU and memory usage for a Docker container.
func makeDockerUsageRow(containerUsage types.StatsJSON, containerInfo types.ContainerJSON) table.Row {
	cpuCores := float64(containerInfo.HostConfig.NanoCPUs)
	memory := float64(containerInfo.HostConfig.Memory)
	cpuUsage := float64(containerUsage.CPUStats.CPUUsage.TotalUsage)
	memoryUsage := float64(containerUsage.MemoryStats.Usage)

	return table.Row{
		containerInfo.Name,
		fmt.Sprintf("%.2f", cpuCores/1_000_000_000),
		fmt.Sprintf("%.2f%%", getPercentage(cpuUsage, cpuCores)),
		fmt.Sprintf("%.2fG", memory/1_000_000_000),
		fmt.Sprintf("%.2f%%", getPercentage(memoryUsage, memory)),
	}
}
