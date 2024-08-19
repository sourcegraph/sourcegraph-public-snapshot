package shared

import (
	"fmt"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
)

type DiskMetricsGroupOptions struct {
	// DiskTitle is the short, lowercase, human-readable name for the disk that we're gathering metrics for.
	//
	// Example: "data"
	DiskTitle string

	// MetricMountNameLabel is the value of the 'mount_name' label that the service uses to identify
	// the mount in its 'mount_point_info' metric.
	//
	// See https://pkg.go.dev/github.com/sourcegraph/mountinfo#NewCollector for more information.
	//
	// Example: "repoDir"
	MetricMountNameLabel string
	// MetricNamespace is the (optional) namespace that the service uses to prefix its 'mount_point_info' metric.
	//
	// Example: "gitserver"
	MetricNamespace string

	// Service Name is the name of the service that we're gathering metrics for.
	//
	// Example: "gitserver"
	ServiceName string

	// InstanceFilterRegex is the PromQL regex that's used to filter the
	// disk metrics to only those emitted by the instance(s) that were interested in.
	//
	// Example: (gitserver-0 | gitserver-1)
	InstanceFilterRegex string
}

// NewDiskMetricsGroup creates a group containing statistics (r/w rate, throughtput, etc.) for the disk
// specified in the given opts.
func NewDiskMetricsGroup(opts DiskMetricsGroupOptions, owner monitoring.ObservableOwner) monitoring.Group {
	mountMetric := "mount_point_info"
	if opts.MetricNamespace != "" {
		mountMetric = opts.MetricNamespace + "_mount_point_info"
	}

	diskStatsQuery := func(nodeExporterMetric string) string {
		return fmt.Sprintf("(max by (instance) (%s * on (device, nodename) group_left() (%s)))",
			fmt.Sprintf("%s{mount_name=%q,instance=~`%s`}", mountMetric, opts.MetricMountNameLabel, opts.InstanceFilterRegex),
			fmt.Sprintf("max by (device, nodename) (rate(%s{instance=~`node-exporter.*`}[1m]))", nodeExporterMetric),
		)
	}

	sharedInterpretationNote := fmt.Sprintf(
		"Note: Disk statistics are per _device_, not per _service_. "+
			"In certain environments (such as common docker-compose setups), %s could be one of _many services_ using this disk. "+
			"These statistics are best interpreted as the load experienced by the device %s is using, not the load %s is solely responsible for causing.",
		opts.ServiceName, opts.ServiceName, opts.ServiceName)

	return monitoring.Group{
		Title:  fmt.Sprintf("%s disk I/O metrics", cases.Title(language.English).String(opts.DiskTitle)),
		Hidden: true,
		Rows: []monitoring.Row{
			{
				{
					Name:        fmt.Sprintf("%s_disk_reads_sec", opts.DiskTitle),
					Description: "read request rate over 1m (per instance)",
					Query:       diskStatsQuery("node_disk_reads_completed_total"),
					NoAlert:     true,
					Panel: monitoring.Panel().LegendFormat("{{instance}}").
						Unit(monitoring.ReadsPerSecond).
						With(monitoring.PanelOptions.LegendOnRight()),
					Owner:          owner,
					Interpretation: fmt.Sprintf("The number of read requests that were issued to the device per second.\n\n%s", sharedInterpretationNote),
				},
				{
					Name:        fmt.Sprintf("%s_disk_writes_sec", opts.DiskTitle),
					Description: "write request rate over 1m (per instance)",
					Query:       diskStatsQuery("node_disk_writes_completed_total"),
					NoAlert:     true,
					Panel: monitoring.Panel().LegendFormat("{{instance}}").
						Unit(monitoring.WritesPerSecond).
						With(monitoring.PanelOptions.LegendOnRight()),
					Owner:          owner,
					Interpretation: fmt.Sprintf("The number of write requests that were issued to the device per second.\n\n%s", sharedInterpretationNote),
				},
			},
			{
				{
					Name:        fmt.Sprintf("%s_disk_read_throughput", opts.DiskTitle),
					Description: "read throughput over 1m (per instance)",
					Query:       diskStatsQuery("node_disk_read_bytes_total"),
					NoAlert:     true,
					Panel: monitoring.Panel().LegendFormat("{{instance}}").
						Unit(monitoring.BytesPerSecond).
						With(monitoring.PanelOptions.LegendOnRight()),
					Owner:          owner,
					Interpretation: fmt.Sprintf("The amount of data that was read from the device per second.\n\n%s", sharedInterpretationNote),
				},
				{
					Name:        fmt.Sprintf("%s_disk_write_throughput", opts.DiskTitle),
					Description: "write throughput over 1m (per instance)",
					Query:       diskStatsQuery("node_disk_written_bytes_total"),
					NoAlert:     true,
					Panel: monitoring.Panel().LegendFormat("{{instance}}").
						Unit(monitoring.BytesPerSecond).
						With(monitoring.PanelOptions.LegendOnRight()),
					Owner:          owner,
					Interpretation: fmt.Sprintf("The amount of data that was written to the device per second.\n\n%s", sharedInterpretationNote),
				},
			},
			{
				{
					Name:        fmt.Sprintf("%s_disk_read_duration", opts.DiskTitle),
					Description: "average read duration over 1m (per instance)",

					Query: fmt.Sprintf("((%s) / (%s))",
						diskStatsQuery("node_disk_read_time_seconds_total"),
						diskStatsQuery("node_disk_reads_completed_total"),
					),
					NoAlert: true,
					Panel: monitoring.Panel().LegendFormat("{{instance}}").
						Unit(monitoring.Seconds).
						With(monitoring.PanelOptions.LegendOnRight()),
					Owner: owner,
					Interpretation: fmt.Sprintf(
						"The average time for read requests issued to the device to be served. This includes the time spent by the requests in queue and the time spent servicing them.\n\n%s",
						sharedInterpretationNote),
				},
				{
					Name:        fmt.Sprintf("%s_disk_write_duration", opts.DiskTitle),
					Description: "average write duration over 1m (per instance)",

					Query: fmt.Sprintf("((%s) / (%s))",
						diskStatsQuery("node_disk_write_time_seconds_total"),
						diskStatsQuery("node_disk_writes_completed_total"),
					),
					NoAlert: true,
					Panel: monitoring.Panel().LegendFormat("{{instance}}").
						Unit(monitoring.Seconds).
						With(monitoring.PanelOptions.LegendOnRight()),
					Owner: owner,
					Interpretation: fmt.Sprintf(
						"The average time for write requests issued to the device to be served. This includes the time spent by the requests in queue and the time spent servicing them.\n\n%s",
						sharedInterpretationNote),
				},
			},
			{
				{
					Name:        fmt.Sprintf("%s_disk_read_request_size", opts.DiskTitle),
					Description: "average read request size over 1m (per instance)",
					Query: fmt.Sprintf("((%s) / (%s))",
						diskStatsQuery("node_disk_read_bytes_total"),
						diskStatsQuery("node_disk_reads_completed_total"),
					),
					NoAlert: true,
					Panel: monitoring.Panel().LegendFormat("{{instance}}").
						Unit(monitoring.Bytes).
						With(monitoring.PanelOptions.LegendOnRight()),
					Owner:          owner,
					Interpretation: fmt.Sprintf("The average size of read requests that were issued to the device.\n\n%s", sharedInterpretationNote),
				},
				{
					Name:        fmt.Sprintf("%s_disk_write_request_size)", opts.DiskTitle),
					Description: "average write request size over 1m (per instance)",
					Query: fmt.Sprintf("((%s) / (%s))",
						diskStatsQuery("node_disk_written_bytes_total"),
						diskStatsQuery("node_disk_writes_completed_total"),
					),
					NoAlert: true,
					Panel: monitoring.Panel().LegendFormat("{{instance}}").
						Unit(monitoring.Bytes).
						With(monitoring.PanelOptions.LegendOnRight()),
					Owner:          owner,
					Interpretation: fmt.Sprintf("The average size of write requests that were issued to the device.\n\n%s", sharedInterpretationNote),
				},
			},
			{
				{
					Name:        fmt.Sprintf("%s_disk_reads_merged_sec", opts.DiskTitle),
					Description: "merged read request rate over 1m (per instance)",
					Query:       diskStatsQuery("node_disk_reads_merged_total"),
					NoAlert:     true,
					Panel: monitoring.Panel().LegendFormat("{{instance}}").
						Unit(monitoring.RequestsPerSecond).
						With(monitoring.PanelOptions.LegendOnRight()),
					Owner:          owner,
					Interpretation: fmt.Sprintf("The number of read requests merged per second that were queued to the device.\n\n%s", sharedInterpretationNote),
				},
				{
					Name:        fmt.Sprintf("%s_disk_writes_merged_sec", opts.DiskTitle),
					Description: "merged writes request rate over 1m (per instance)",
					Query:       diskStatsQuery("node_disk_writes_merged_total"),
					NoAlert:     true,
					Panel: monitoring.Panel().LegendFormat("{{instance}}").
						Unit(monitoring.RequestsPerSecond).
						With(monitoring.PanelOptions.LegendOnRight()),
					Owner:          owner,
					Interpretation: fmt.Sprintf("The number of write requests merged per second that were queued to the device.\n\n%s", sharedInterpretationNote),
				},
			},
			{
				{

					Name:        fmt.Sprintf("%s_disk_average_queue_size", opts.DiskTitle),
					Description: "average queue size over 1m (per instance)",
					Query:       diskStatsQuery("node_disk_io_time_weighted_seconds_total"),
					NoAlert:     true,
					Panel: monitoring.Panel().LegendFormat("{{instance}}").
						Unit("req").
						With(monitoring.PanelOptions.LegendOnRight()),
					Owner: owner,
					Interpretation: fmt.Sprintf(
						"The number of I/O operations that were being queued or being serviced. See https://blog.actorsfit.com/a?ID=00200-428fa2ac-e338-4540-848c-af9a3eb1ebd2 for background (avgqu-sz).\n\n%s",
						sharedInterpretationNote,
					),
				},
			},
		},
	}
}
