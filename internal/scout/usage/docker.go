package usage

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
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

		if cfg.container != "" {
			if containerInfo.Name == cfg.container {
				row := makeDockerUsageRow(ctx, cfg, containerInfo)
				rows = append(rows, row)
				break
			} else {
				continue
			}
		}

		row := makeDockerUsageRow(ctx, cfg, containerInfo)
		rows = append(rows, row)
	}

	if len(rows) == 0 {
		msg := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFA500"))
		if cfg.container == "" {
			fmt.Println(msg.Render(`No docker containers are running.`))
			os.Exit(1)
		}
		fmt.Println(msg.Render(
			fmt.Sprintf(`No container with name '%s' running.`, cfg.container),
		))
		os.Exit(1)
	}

	style.UsageTable(columns, rows)
	return nil
}

// makeDockerUsageRow generates a table row displaying CPU and memory usage for a Docker container.
func makeDockerUsageRow(ctx context.Context, cfg *Config, container types.ContainerJSON) table.Row {
	stats, err := cfg.dockerClient.ContainerStats(ctx, container.ID, false)
	if err != nil {
		errors.Wrap(err, "could not get container stats")
		os.Exit(1)
	}
	defer func() { _ = stats.Body.Close() }()

	var usage types.StatsJSON
	if err := json.NewDecoder(stats.Body).Decode(&usage); err != nil {
		errors.Wrap(err, "could not get container stats")
		os.Exit(1)
	}

	cpuCores := float64(container.HostConfig.NanoCPUs)
	memory := float64(container.HostConfig.Memory)
	cpuUsage := float64(usage.CPUStats.CPUUsage.TotalUsage)
	memoryUsage := float64(usage.MemoryStats.Usage)

	return table.Row{
		container.Name,
		fmt.Sprintf("%.2f", cpuCores/1_000_000_000),
		fmt.Sprintf("%.2f%%", getPercentage(cpuUsage, cpuCores)),
		fmt.Sprintf("%.2fG", memory/1_000_000_000),
		fmt.Sprintf("%.2f%%", getPercentage(memoryUsage, memory)),
	}
}
