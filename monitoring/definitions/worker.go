package definitions

import (
	"fmt"
	"time"

	"github.com/sourcegraph/sourcegraph/monitoring/definitions/shared"
	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
)

func Worker() *monitoring.Container {
	return &monitoring.Container{
		Name:        "worker",
		Title:       "Worker",
		Description: "Manages background processes.",
		Groups: []monitoring.Group{
			{
				Title: "Active jobs",
				Rows: append([]monitoring.Row{
					{
						{
							Name:           "worker_job_count",
							Description:    "number of worker instances running each job",
							Query:          `sum by (job_name) (src_worker_jobs{job="worker"})`,
							Panel:          monitoring.Panel().LegendFormat("instances running {{job_name}}"),
							NoAlert:        true,
							Interpretation: "Number of worker instances running each job type",
						},
					},
				}, createWorkerActiveJobRows()...),
			},
			{
				Title:  "Internal service requests",
				Hidden: true,
				Rows: []monitoring.Row{
					{
						shared.FrontendInternalAPIErrorResponses("worker", monitoring.ObservableOwnerCodeIntel).Observable(),
					},
				},
			},
			{
				Title:  shared.TitleContainerMonitoring,
				Hidden: true,
				Rows: []monitoring.Row{
					{
						shared.ContainerCPUUsage("worker", monitoring.ObservableOwnerCodeIntel).Observable(),
						shared.ContainerMemoryUsage("worker", monitoring.ObservableOwnerCodeIntel).Observable(),
					},
					{
						shared.ContainerMissing("worker", monitoring.ObservableOwnerCodeIntel).Observable(),
					},
				},
			},
			{
				Title:  shared.TitleProvisioningIndicators,
				Hidden: true,
				Rows: []monitoring.Row{
					{
						shared.ProvisioningCPUUsageLongTerm("worker", monitoring.ObservableOwnerCodeIntel).Observable(),
						shared.ProvisioningMemoryUsageLongTerm("worker", monitoring.ObservableOwnerCodeIntel).Observable(),
					},
					{
						shared.ProvisioningCPUUsageShortTerm("worker", monitoring.ObservableOwnerCodeIntel).Observable(),
						shared.ProvisioningMemoryUsageShortTerm("worker", monitoring.ObservableOwnerCodeIntel).Observable(),
					},
				},
			},
			{
				Title:  shared.TitleGolangMonitoring,
				Hidden: true,
				Rows: []monitoring.Row{
					{
						shared.GoGoroutines("worker", monitoring.ObservableOwnerCodeIntel).Observable(),
						shared.GoGcDuration("worker", monitoring.ObservableOwnerCodeIntel).Observable(),
					},
				},
			},
			{
				Title:  shared.TitleKubernetesMonitoring,
				Hidden: true,
				Rows: []monitoring.Row{
					{
						shared.KubernetesPodsAvailable("worker", monitoring.ObservableOwnerCodeIntel).Observable(),
					},
				},
			},
		},
	}
}

var workerJobs = []struct {
	Name  string
	Owner monitoring.ObservableOwner
}{
	{Name: "codeintel-janitor", Owner: monitoring.ObservableOwnerCodeIntel},
	{Name: "codeintel-commitgraph", Owner: monitoring.ObservableOwnerCodeIntel},
}

func createWorkerActiveJobRows() []monitoring.Row {
	var activeJobObservables []monitoring.Observable
	for _, job := range workerJobs {
		activeJobObservables = append(activeJobObservables, monitoring.Observable{
			Name:          fmt.Sprintf("worker_job_%s_count", job.Name),
			Description:   fmt.Sprintf("number of worker instances running the %s job", job.Name),
			Query:         fmt.Sprintf(`sum (src_worker_jobs{job="worker", job_name="%s"})`, job.Name),
			Panel:         monitoring.Panel().LegendFormat(fmt.Sprintf("instances running %s", job.Name)),
			DataMustExist: true,
			Warning:       monitoring.Alert().Less(1, nil).For(1 * time.Minute),
			Critical:      monitoring.Alert().Less(1, nil).For(5 * time.Minute),
			Owner:         job.Owner,
			PossibleSolutions: fmt.Sprintf(`
				- Ensure your instance defines a worker container such that:
					- `+"`"+`WORKER_JOB_ALLOWLIST`+"`"+` contains "%[1]s" (or "all"), and
					- `+"`"+`WORKER_JOB_BLOCKLIST`+"`"+` does not contain "%[1]s"
				- Ensure that such a container is not failing to start or stay active
			`, job.Name),
		})
	}

	panelsPerRow := 4
	if rem := len(activeJobObservables) % panelsPerRow; rem == 1 || rem == 2 {
		// If we'd leave one or two panels on the only/last row, then reduce
		// the number of panels in previous rows so that we have less of a width
		// difference at the end
		panelsPerRow = 3
	}

	var activeJobRows []monitoring.Row
	for _, observable := range activeJobObservables {
		if n := len(activeJobRows); n == 0 || len(activeJobRows[n-1]) >= panelsPerRow {
			activeJobRows = append(activeJobRows, nil)
		}

		n := len(activeJobRows)
		activeJobRows[n-1] = append(activeJobRows[n-1], observable)
	}

	return activeJobRows
}
