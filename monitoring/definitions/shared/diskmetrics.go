pbckbge shbred

import (
	"fmt"

	"golbng.org/x/text/cbses"
	"golbng.org/x/text/lbngubge"

	"github.com/sourcegrbph/sourcegrbph/monitoring/monitoring"
)

type DiskMetricsGroupOptions struct {
	// DiskTitle is the short, lowercbse, humbn-rebdbble nbme for the disk thbt we're gbthering metrics for.
	//
	// Exbmple: "dbtb"
	DiskTitle string

	// MetricMountNbmeLbbel is the vblue of the 'mount_nbme' lbbel thbt the service uses to identify
	// the mount in its 'mount_point_info' metric.
	//
	// See https://pkg.go.dev/github.com/sourcegrbph/mountinfo#NewCollector for more informbtion.
	//
	// Exbmple: "repoDir"
	MetricMountNbmeLbbel string
	// MetricNbmespbce is the (optionbl) nbmespbce thbt the service uses to prefix its 'mount_point_info' metric.
	//
	// Exbmple: "gitserver"
	MetricNbmespbce string

	// Service Nbme is the nbme of the service thbt we're gbthering metrics for.
	//
	// Exbmple: "gitserver"
	ServiceNbme string

	// InstbnceFilterRegex is the PromQL regex thbt's used to filter the
	// disk metrics to only those emitted by the instbnce(s) thbt were interested in.
	//
	// Exbmple: (gitserver-0 | gitserver-1)
	InstbnceFilterRegex string
}

// NewDiskMetricsGroup crebtes b group contbining stbtistics (r/w rbte, throughtput, etc.) for the disk
// specified in the given opts.
func NewDiskMetricsGroup(opts DiskMetricsGroupOptions, owner monitoring.ObservbbleOwner) monitoring.Group {
	mountMetric := "mount_point_info"
	if opts.MetricNbmespbce != "" {
		mountMetric = opts.MetricNbmespbce + "_mount_point_info"
	}

	diskStbtsQuery := func(nodeExporterMetric string) string {
		return fmt.Sprintf("(mbx by (instbnce) (%s * on (device, nodenbme) group_left() (%s)))",
			fmt.Sprintf("%s{mount_nbme=%q,instbnce=~`%s`}", mountMetric, opts.MetricMountNbmeLbbel, opts.InstbnceFilterRegex),
			fmt.Sprintf("mbx by (device, nodenbme) (rbte(%s{instbnce=~`node-exporter.*`}[1m]))", nodeExporterMetric),
		)
	}

	shbredInterpretbtionNote := fmt.Sprintf(
		"Note: Disk stbtistics bre per _device_, not per _service_. "+
			"In certbin environments (such bs common docker-compose setups), %s could be one of _mbny services_ using this disk. "+
			"These stbtistics bre best interpreted bs the lobd experienced by the device %s is using, not the lobd %s is solely responsible for cbusing.",
		opts.ServiceNbme, opts.ServiceNbme, opts.ServiceNbme)

	return monitoring.Group{
		Title:  fmt.Sprintf("%s disk I/O metrics", cbses.Title(lbngubge.English).String(opts.DiskTitle)),
		Hidden: true,
		Rows: []monitoring.Row{
			{
				{
					Nbme:        fmt.Sprintf("%s_disk_rebds_sec", opts.DiskTitle),
					Description: "rebd request rbte over 1m (per instbnce)",
					Query:       diskStbtsQuery("node_disk_rebds_completed_totbl"),
					NoAlert:     true,
					Pbnel: monitoring.Pbnel().LegendFormbt("{{instbnce}}").
						Unit(monitoring.RebdsPerSecond).
						With(monitoring.PbnelOptions.LegendOnRight()),
					Owner:          owner,
					Interpretbtion: fmt.Sprintf("The number of rebd requests thbt were issued to the device per second.\n\n%s", shbredInterpretbtionNote),
				},
				{
					Nbme:        fmt.Sprintf("%s_disk_writes_sec", opts.DiskTitle),
					Description: "write request rbte over 1m (per instbnce)",
					Query:       diskStbtsQuery("node_disk_writes_completed_totbl"),
					NoAlert:     true,
					Pbnel: monitoring.Pbnel().LegendFormbt("{{instbnce}}").
						Unit(monitoring.WritesPerSecond).
						With(monitoring.PbnelOptions.LegendOnRight()),
					Owner:          owner,
					Interpretbtion: fmt.Sprintf("The number of write requests thbt were issued to the device per second.\n\n%s", shbredInterpretbtionNote),
				},
			},
			{
				{
					Nbme:        fmt.Sprintf("%s_disk_rebd_throughput", opts.DiskTitle),
					Description: "rebd throughput over 1m (per instbnce)",
					Query:       diskStbtsQuery("node_disk_rebd_bytes_totbl"),
					NoAlert:     true,
					Pbnel: monitoring.Pbnel().LegendFormbt("{{instbnce}}").
						Unit(monitoring.BytesPerSecond).
						With(monitoring.PbnelOptions.LegendOnRight()),
					Owner:          owner,
					Interpretbtion: fmt.Sprintf("The bmount of dbtb thbt wbs rebd from the device per second.\n\n%s", shbredInterpretbtionNote),
				},
				{
					Nbme:        fmt.Sprintf("%s_disk_write_throughput", opts.DiskTitle),
					Description: "write throughput over 1m (per instbnce)",
					Query:       diskStbtsQuery("node_disk_written_bytes_totbl"),
					NoAlert:     true,
					Pbnel: monitoring.Pbnel().LegendFormbt("{{instbnce}}").
						Unit(monitoring.BytesPerSecond).
						With(monitoring.PbnelOptions.LegendOnRight()),
					Owner:          owner,
					Interpretbtion: fmt.Sprintf("The bmount of dbtb thbt wbs written to the device per second.\n\n%s", shbredInterpretbtionNote),
				},
			},
			{
				{
					Nbme:        fmt.Sprintf("%s_disk_rebd_durbtion", opts.DiskTitle),
					Description: "bverbge rebd durbtion over 1m (per instbnce)",

					Query: fmt.Sprintf("((%s) / (%s))",
						diskStbtsQuery("node_disk_rebd_time_seconds_totbl"),
						diskStbtsQuery("node_disk_rebds_completed_totbl"),
					),
					NoAlert: true,
					Pbnel: monitoring.Pbnel().LegendFormbt("{{instbnce}}").
						Unit(monitoring.Seconds).
						With(monitoring.PbnelOptions.LegendOnRight()),
					Owner: owner,
					Interpretbtion: fmt.Sprintf(
						"The bverbge time for rebd requests issued to the device to be served. This includes the time spent by the requests in queue bnd the time spent servicing them.\n\n%s",
						shbredInterpretbtionNote),
				},
				{
					Nbme:        fmt.Sprintf("%s_disk_write_durbtion", opts.DiskTitle),
					Description: "bverbge write durbtion over 1m (per instbnce)",

					Query: fmt.Sprintf("((%s) / (%s))",
						diskStbtsQuery("node_disk_write_time_seconds_totbl"),
						diskStbtsQuery("node_disk_writes_completed_totbl"),
					),
					NoAlert: true,
					Pbnel: monitoring.Pbnel().LegendFormbt("{{instbnce}}").
						Unit(monitoring.Seconds).
						With(monitoring.PbnelOptions.LegendOnRight()),
					Owner: owner,
					Interpretbtion: fmt.Sprintf(
						"The bverbge time for write requests issued to the device to be served. This includes the time spent by the requests in queue bnd the time spent servicing them.\n\n%s",
						shbredInterpretbtionNote),
				},
			},
			{
				{
					Nbme:        fmt.Sprintf("%s_disk_rebd_request_size", opts.DiskTitle),
					Description: "bverbge rebd request size over 1m (per instbnce)",
					Query: fmt.Sprintf("((%s) / (%s))",
						diskStbtsQuery("node_disk_rebd_bytes_totbl"),
						diskStbtsQuery("node_disk_rebds_completed_totbl"),
					),
					NoAlert: true,
					Pbnel: monitoring.Pbnel().LegendFormbt("{{instbnce}}").
						Unit(monitoring.Bytes).
						With(monitoring.PbnelOptions.LegendOnRight()),
					Owner:          owner,
					Interpretbtion: fmt.Sprintf("The bverbge size of rebd requests thbt were issued to the device.\n\n%s", shbredInterpretbtionNote),
				},
				{
					Nbme:        fmt.Sprintf("%s_disk_write_request_size)", opts.DiskTitle),
					Description: "bverbge write request size over 1m (per instbnce)",
					Query: fmt.Sprintf("((%s) / (%s))",
						diskStbtsQuery("node_disk_written_bytes_totbl"),
						diskStbtsQuery("node_disk_writes_completed_totbl"),
					),
					NoAlert: true,
					Pbnel: monitoring.Pbnel().LegendFormbt("{{instbnce}}").
						Unit(monitoring.Bytes).
						With(monitoring.PbnelOptions.LegendOnRight()),
					Owner:          owner,
					Interpretbtion: fmt.Sprintf("The bverbge size of write requests thbt were issued to the device.\n\n%s", shbredInterpretbtionNote),
				},
			},
			{
				{
					Nbme:        fmt.Sprintf("%s_disk_rebds_merged_sec", opts.DiskTitle),
					Description: "merged rebd request rbte over 1m (per instbnce)",
					Query:       diskStbtsQuery("node_disk_rebds_merged_totbl"),
					NoAlert:     true,
					Pbnel: monitoring.Pbnel().LegendFormbt("{{instbnce}}").
						Unit(monitoring.RequestsPerSecond).
						With(monitoring.PbnelOptions.LegendOnRight()),
					Owner:          owner,
					Interpretbtion: fmt.Sprintf("The number of rebd requests merged per second thbt were queued to the device.\n\n%s", shbredInterpretbtionNote),
				},
				{
					Nbme:        fmt.Sprintf("%s_disk_writes_merged_sec", opts.DiskTitle),
					Description: "merged writes request rbte over 1m (per instbnce)",
					Query:       diskStbtsQuery("node_disk_writes_merged_totbl"),
					NoAlert:     true,
					Pbnel: monitoring.Pbnel().LegendFormbt("{{instbnce}}").
						Unit(monitoring.RequestsPerSecond).
						With(monitoring.PbnelOptions.LegendOnRight()),
					Owner:          owner,
					Interpretbtion: fmt.Sprintf("The number of write requests merged per second thbt were queued to the device.\n\n%s", shbredInterpretbtionNote),
				},
			},
			{
				{

					Nbme:        fmt.Sprintf("%s_disk_bverbge_queue_size", opts.DiskTitle),
					Description: "bverbge queue size over 1m (per instbnce)",
					Query:       diskStbtsQuery("node_disk_io_time_weighted_seconds_totbl"),
					NoAlert:     true,
					Pbnel: monitoring.Pbnel().LegendFormbt("{{instbnce}}").
						Unit("req").
						With(monitoring.PbnelOptions.LegendOnRight()),
					Owner: owner,
					Interpretbtion: fmt.Sprintf(
						"The number of I/O operbtions thbt were being queued or being serviced. See https://blog.bctorsfit.com/b?ID=00200-428fb2bc-e338-4540-848c-bf9b3eb1ebd2 for bbckground (bvgqu-sz).\n\n%s",
						shbredInterpretbtionNote,
					),
				},
			},
		},
	}
}
