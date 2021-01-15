package definitions

import (
	"time"

	"github.com/sourcegraph/sourcegraph/monitoring/definitions/shared"
	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
)

func GitServer() *monitoring.Container {
	return &monitoring.Container{
		Name:        "gitserver",
		Title:       "Git Server",
		Description: "Stores, manages, and operates Git repositories.",
		Groups: []monitoring.Group{
			{
				Title: "General",
				Rows: []monitoring.Row{
					{
						{
							Name:        "disk_space_remaining",
							Description: "disk space remaining by instance",
							Query:       `(src_gitserver_disk_space_available / src_gitserver_disk_space_total) * 100`,
							Warning:     monitoring.Alert().LessOrEqual(25),
							Critical:    monitoring.Alert().LessOrEqual(15),
							Panel:       monitoring.Panel().LegendFormat("{{instance}}").Unit(monitoring.Percentage),
							Owner:       monitoring.ObservableOwnerCloud,
							PossibleSolutions: `
								- **Provision more disk space:** Sourcegraph will begin deleting least-used repository clones at 10% disk space remaining which may result in decreased performance, users having to wait for repositories to clone, etc.
							`,
						},
						{
							Name:        "running_git_commands",
							Description: "running git commands",
							Query:       "max(src_gitserver_exec_running)",
							Warning:     monitoring.Alert().GreaterOrEqual(50).For(2 * time.Minute),
							Critical:    monitoring.Alert().GreaterOrEqual(100).For(5 * time.Minute),
							Panel:       monitoring.Panel().LegendFormat("running commands"),
							Owner:       monitoring.ObservableOwnerCloud,
							Interpretation: `
								A high value signals load.
							`,
							PossibleSolutions: `
								- **Check if the problem may be an intermittent and temporary peak** using the "Container monitoring" section at the bottom of the Git Server dashboard.
								- **Single container deployments:** Consider upgrading to a [Docker Compose deployment](../install/docker-compose/migrate.md) which offers better scalability and resource isolation.
								- **Kubernetes and Docker Compose:** Check that you are running a similar number of git server replicas and that their CPU/memory limits are allocated according to what is shown in the [Sourcegraph resource estimator](../install/resource_estimator.md).
							`,
						},
					}, {
						{
							Name:        "repository_clone_queue_size",
							Description: "repository clone queue size",
							Query:       "sum(src_gitserver_clone_queue)",
							Warning:     monitoring.Alert().GreaterOrEqual(25),
							Panel:       monitoring.Panel().LegendFormat("queue size"),
							Owner:       monitoring.ObservableOwnerCloud,
							PossibleSolutions: `
								- **If you just added several repositories**, the warning may be expected.
								- **Check which repositories need cloning**, by visiting e.g. https://sourcegraph.example.com/site-admin/repositories?filter=not-cloned
							`,
						},
						{
							Name:        "repository_existence_check_queue_size",
							Description: "repository existence check queue size",
							Query:       "sum(src_gitserver_lsremote_queue)",
							Warning:     monitoring.Alert().GreaterOrEqual(25),
							Panel:       monitoring.Panel().LegendFormat("queue size"),
							Owner:       monitoring.ObservableOwnerCloud,
							PossibleSolutions: `
								- **Check the code host status indicator for errors:** on the Sourcegraph app homepage, when signed in as an admin click the cloud icon in the top right corner of the page.
								- **Check if the issue continues to happen after 30 minutes**, it may be temporary.
								- **Check the gitserver logs for more information.**
							`,
						},
					}, {
						{
							Name:        "echo_command_duration_test",
							Description: "echo test command duration",
							Query:       "max(src_gitserver_echo_duration_seconds)",
							NoAlert:     true,
							Panel:       monitoring.Panel().LegendFormat("running commands").Unit(monitoring.Seconds),
							Owner:       monitoring.ObservableOwnerCloud,
							Interpretation: `
								A high value here likely indicates a problem, especially if consistently high.
								You can query for individual commands using 'sum by (cmd)(src_gitserver_exec_running)' in Grafana ('/-/debug/grafana') to see if a specific Git Server command might be spiking in frequency.

								If this value is consistently high, consider the following:

								- **Single container deployments:** Upgrade to a [Docker Compose deployment](../install/docker-compose/migrate.md) which offers better scalability and resource isolation.
								- **Kubernetes and Docker Compose:** Check that you are running a similar number of git server replicas and that their CPU/memory limits are allocated according to what is shown in the [Sourcegraph resource estimator](../install/resource_estimator.md).
							`,
						},
						shared.FrontendInternalAPIErrorResponses("gitserver", monitoring.ObservableOwnerCloud).Observable(),
					},
				},
			},
			{
				Title:  shared.TitleContainerMonitoring,
				Hidden: true,
				Rows: []monitoring.Row{
					{
						shared.ContainerCPUUsage("gitserver", monitoring.ObservableOwnerCloud).Observable(),
						shared.ContainerMemoryUsage("gitserver", monitoring.ObservableOwnerCloud).Observable(),
					},
					{
						// git server does not have 0-downtime deploy, so deploys can
						// cause extended container restarts. still seta warning alert for
						// extended periods of container restarts, since this might still
						// indicate a problem.
						shared.ContainerRestarts("gitserver", monitoring.ObservableOwnerCloud).
							WithWarning(monitoring.Alert().Greater(1).For(10 * time.Minute)).
							WithCritical(nil).
							Observable(),
						shared.ContainerIOUsage("gitserver", monitoring.ObservableOwnerCloud).Observable(),
					},
				},
			},
			{
				Title:  shared.TitleProvisioningIndicators,
				Hidden: true,
				Rows: []monitoring.Row{
					{
						shared.ProvisioningCPUUsageLongTerm("gitserver", monitoring.ObservableOwnerCloud).Observable(),
						shared.ProvisioningMemoryUsageLongTerm("gitserver", monitoring.ObservableOwnerCloud).
							WithNoAlerts(`Git Server is expected to use up all the memory it is provided.`).
							Observable(),
					},
					{
						shared.ProvisioningCPUUsageShortTerm("gitserver", monitoring.ObservableOwnerCloud).Observable(),
						shared.ProvisioningMemoryUsageShortTerm("gitserver", monitoring.ObservableOwnerCloud).
							WithNoAlerts(`Git Server is expected to use up all the memory it is provided.`).
							Observable(),
					},
				},
			},
			{
				Title:  shared.TitleGolangMonitoring,
				Hidden: true,
				Rows: []monitoring.Row{
					{
						shared.GoGoroutines("gitserver", monitoring.ObservableOwnerCloud).Observable(),
						shared.GoGcDuration("gitserver", monitoring.ObservableOwnerCloud).Observable(),
					},
				},
			},
			{
				Title:  "Kubernetes monitoring (ignore if using Docker Compose or server)",
				Hidden: true,
				Rows: []monitoring.Row{
					{
						shared.KubernetesPodsAvailable("gitserver", monitoring.ObservableOwnerCloud).Observable(),
					},
				},
			},
		},
	}
}
