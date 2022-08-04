package shared

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
)

const TitleNodeExporter = "Executor: %s instance metrics"

func NewNodeExporterGroup(job, jobTitle, instanceFilter string) monitoring.Group {
	return monitoring.Group{
		Title:  fmt.Sprintf(TitleNodeExporter, jobTitle),
		Hidden: true,
		Rows: []monitoring.Row{
			{
				{
					Name:           "node_cpu_utilization",
					Description:    "CPU utilization (minus idle/iowait)",
					Query:          "sum(rate(node_cpu_seconds_total{sg_job=~\"" + job + "\",mode!~\"(idle|iowait)\",sg_instance=~\"" + instanceFilter + "\"}[$__rate_interval])) by(sg_instance) / count(node_cpu_seconds_total{sg_job=~\"" + job + "\",mode=\"system\",sg_instance=~\"" + instanceFilter + "\"}) by (sg_instance) * 100",
					NoAlert:        true,
					Interpretation: "Indicates the amount of CPU time excluding idle and iowait time, divided by the number of cores, as a percentage.",
					Panel:          monitoring.Panel().LegendFormat("{{sg_instance}}").Unit(monitoring.Percentage).Max(100),
				},
				{
					Name:        "node_cpu_saturation_cpu_wait",
					Description: "CPU saturation (time waiting)",
					Query:       "rate(node_pressure_cpu_waiting_seconds_total{sg_job=~\"" + job + "\",sg_instance=~\"" + instanceFilter + "\"}[$__rate_interval])",
					NoAlert:     true,
					Interpretation: "Indicates the average summed time a number of (but strictly not all) non-idle processes spent waiting for CPU time. If this is higher than normal, then the CPU is underpowered for the workload and more powerful machines should be provisioned. " +
						"This only represents a \"less-than-all processes\" time, because for processes to be waiting for CPU time there must be other process(es) consuming CPU time.",
					Panel: monitoring.Panel().LegendFormat("{{sg_instance}}").Unit(monitoring.Seconds),
				},
			},
			{
				{
					Name:        "node_memory_utilization",
					Description: "memory utilization",
					Query:       "(1 - sum(node_memory_MemAvailable_bytes{sg_job=~\"" + job + "\",sg_instance=~\"" + instanceFilter + "\"}) by (sg_instance) / sum(node_memory_MemTotal_bytes{sg_job=~\"" + job + "\",sg_instance=~\"" + instanceFilter + "\"}) by (sg_instance)) * 100",
					NoAlert:     true,
					Interpretation: "Indicates the amount of available memory (including cache and buffers) as a percentage. Consistently high numbers are generally fine so long memory saturation figures are within acceptable ranges, " +
						"these figures may be more useful for informing executor provisioning decisions, such as increasing worker parallelism, down-sizing machines etc.",
					Panel: monitoring.Panel().LegendFormat("{{sg_instance}}").Unit(monitoring.Percentage).Max(100),
				},
				// Please see the following article(s) on how we arrive at using these particular metrics. It is stupid complicated and underdocumented beyond anything.
				// Page 27 of https://documentation.suse.com/sles/11-SP4/pdf/book-sle-tuning_color_en.pdf
				// https://doc.opensuse.org/documentation/leap/archive/42.3/tuning/html/book.sle.tuning/cha.tuning.memory.html#cha.tuning.memory.monitoring
				// https://man7.org/linux/man-pages/man1/sar.1.html#:~:text=Report%20paging%20statistics.
				// https://facebookmicrosites.github.io/psi/docs/overview
				{
					Name:        "node_memory_saturation_vmeff",
					Description: "memory saturation (vmem efficiency)",
					Query: "(rate(node_vmstat_pgsteal_anon{sg_job=~\"" + job + "\",sg_instance=~\"" + instanceFilter + "\"}[$__rate_interval]) + rate(node_vmstat_pgsteal_direct{sg_job=~\"" + job + "\",sg_instance=~\"" + instanceFilter + "\"}[$__rate_interval]) + rate(node_vmstat_pgsteal_file{sg_job=~\"" + job + "\",sg_instance=~\"" + instanceFilter + "\"}[$__rate_interval]) + rate(node_vmstat_pgsteal_kswapd{sg_job=~\"" + job + "\",sg_instance=~\"" + instanceFilter + "\"}[$__rate_interval])) " +
						"/ (rate(node_vmstat_pgscan_anon{sg_job=~\"" + job + "\",sg_instance=~\"" + instanceFilter + "\"}[$__rate_interval]) + rate(node_vmstat_pgscan_direct{sg_job=~\"" + job + "\",sg_instance=~\"" + instanceFilter + "\"}[$__rate_interval]) + rate(node_vmstat_pgscan_file{sg_job=~\"" + job + "\",sg_instance=~\"" + instanceFilter + "\"}[$__rate_interval]) + rate(node_vmstat_pgscan_kswapd{sg_job=~\"" + job + "\",sg_instance=~\"" + instanceFilter + "\"}[$__rate_interval])) * 100",
					NoAlert: true,
					Interpretation: "Indicates the efficiency of page reclaim, calculated as pgsteal/pgscan. Optimal figures are short spikes of near 100% and above, indicating that a high ratio of scanned pages are actually being freed, " +
						"or exactly 0%, indicating that pages arent being scanned as there is no memory pressure. Sustained numbers >~100% may be sign of imminent memory exhaustion, while sustained 0% < x < ~100% figures are very serious.",
					Panel: monitoring.Panel().LegendFormat("{{sg_instance}}").Unit(monitoring.Percentage),
				},
				{
					Name:           "node_memory_saturation_pressure_stalled",
					Description:    "memory saturation (fully stalled)",
					Query:          "rate(node_pressure_memory_stalled_seconds_total{sg_job=~\"" + job + "\",sg_instance=~\"" + instanceFilter + "\"}[$__rate_interval])",
					NoAlert:        true,
					Interpretation: "Indicates the amount of time all non-idle processes were stalled waiting on memory operations to complete. This is often correlated with vmem efficiency ratio when pressure on available memory is high. If they're not correlated, this could indicate issues with the machine hardware and/or configuration.",
					Panel:          monitoring.Panel().LegendFormat("{{sg_instance}}").Unit(monitoring.Seconds),
				},
			},
			{
				// Please see the following article(s) on how we arrive at these metrics. Its non-trivial, second only to memory saturation
				// https://brian-candler.medium.com/interpreting-prometheus-metrics-for-linux-disk-i-o-utilization-4db53dfedcfc
				// https://www.robustperception.io/mapping-iostat-to-the-node-exporters-node_disk_-metrics
				{
					Name:        "node_io_disk_utilization",
					Description: "disk IO utilization (percentage time spent in IO)",
					Query:       "sum(label_replace(label_replace(rate(node_disk_io_time_seconds_total{sg_job=~\"" + job + "\",sg_instance=~\"" + instanceFilter + "\"}[$__rate_interval]), \"disk\", \"$1\", \"device\", \"^([^d].+)\"), \"disk\", \"ignite\", \"device\", \"dm-.*\")) by(sg_instance,disk) * 100",
					NoAlert:     true,
					Interpretation: "Indicates the percentage of time a disk was busy. If this is less than 100%, then the disk has spare utilization capacity. However, a value of 100% does not necesarily indicate the disk is at max capacity. " +
						"For single, serial request-serving devices, 100% may indicate maximum saturation, but for SSDs and RAID arrays this is less likely to be the case, as they are capable of serving multiple requests in parallel, other metrics such as " +
						"throughput and request queue size should be factored in.",
					Panel: monitoring.Panel().LegendFormat("{{sg_instance}}: {{disk}}").Unit(monitoring.Percentage),
				},
				{
					Name:        "node_io_disk_saturation",
					Description: "disk IO saturation (avg IO queue size)",
					Query:       "sum(label_replace(label_replace(rate(node_disk_io_time_weighted_seconds_total{sg_job=~\"" + job + "\",sg_instance=~\"" + instanceFilter + "\"}[$__rate_interval]), \"disk\", \"$1\", \"device\", \"^([^d].+)\"), \"disk\", \"ignite\", \"device\", \"dm-.*\")) by(sg_instance,disk)",
					NoAlert:     true,
					Interpretation: "Indicates the number of outstanding/queued IO requests. High but short-lived queue sizes may not present an issue, but if theyre consistently/often high and/or monotonically increasing, the disk may be failing or simply too slow for the amount of activity required. " +
						"Consider replacing the drive(s) with SSDs if they are not already and/or replacing the faulty drive(s), if any.",
					Panel: monitoring.Panel().LegendFormat("{{sg_instance}}: {{disk}}"),
				},
				{
					Name:           "node_io_disk_saturation_pressure_full",
					Description:    "disk IO saturation (avg time of all processes stalled)",
					Query:          "rate(node_pressure_io_stalled_seconds_total{sg_job=~\"" + job + "\",sg_instance=~\"" + instanceFilter + "\"}[$__rate_interval])",
					NoAlert:        true,
					Interpretation: "Indicates the averaged amount of time for which all non-idle processes were stalled waiting for IO to complete simultaneously aka where no processes could make progress.", // TODO: more
					Panel:          monitoring.Panel().LegendFormat("{{sg_instance}}").Unit(monitoring.Seconds),
				},
			},
			{
				{
					Name:        "node_io_network_utilization",
					Description: "network IO utilization (Rx)",
					Query:       "sum(rate(node_network_receive_bytes_total{sg_job=~\"" + job + "\",sg_instance=~\"" + instanceFilter + "\"}[$__rate_interval])) by(sg_instance) * 8",
					NoAlert:     true,
					Interpretation: "Indicates the average summed receiving throughput of all network interfaces. This is often predominantly composed of the WAN/internet-connected interface, and knowing normal/good figures depends on knowing the bandwidth of the " +
						"underlying hardware and the workloads.",
					Panel: monitoring.Panel().LegendFormat("{{sg_instance}}").Unit(monitoring.BitsPerSecond),
				},
				{
					Name:        "node_io_network_saturation",
					Description: "network IO saturation (Rx packets dropped)",
					Query:       "sum(rate(node_network_receive_drop_total{sg_job=~\"" + job + "\",sg_instance=~\"" + instanceFilter + "\"}[$__rate_interval])) by(sg_instance)",
					NoAlert:     true,
					Interpretation: "Number of dropped received packets. This can happen if the receive queues/buffers become full due to slow packet processing throughput. The queues/buffers could be configured to be larger as a stop-gap " +
						"but the processing application should be investigated as soon as possible. https://www.kernel.org/doc/html/latest/networking/statistics.html#:~:text=not%20otherwise%20counted.-,rx_dropped,-Number%20of%20packets",
					Panel: monitoring.Panel().LegendFormat("{{sg_instance}}"),
				},
				{
					Name:           "node_io_network_saturation",
					Description:    "network IO errors (Rx)",
					Query:          "sum(rate(node_network_receive_errs_total{sg_job=~\"" + job + "\",sg_instance=~\"" + instanceFilter + "\"}[$__rate_interval])) by(sg_instance)",
					NoAlert:        true,
					Interpretation: "Number of bad/malformed packets received. https://www.kernel.org/doc/html/latest/networking/statistics.html#:~:text=excluding%20the%20FCS.-,rx_errors,-Total%20number%20of",
					Panel:          monitoring.Panel().LegendFormat("{{sg_instance}}"),
				},
			},
			{
				{
					Name:        "node_io_network_utilization",
					Description: "network IO utilization (Tx)",
					Query:       "sum(rate(node_network_transmit_bytes_total{sg_job=~\"" + job + "\",sg_instance=~\"" + instanceFilter + "\"}[$__rate_interval])) by(sg_instance) * 8",
					NoAlert:     true,
					Interpretation: "Indicates the average summed transmitted throughput of all network interfaces. This is often predominantly composed of the WAN/internet-connected interface, and knowing normal/good figures depends on knowing the bandwidth of the " +
						"underlying hardware and the workloads.",
					Panel: monitoring.Panel().LegendFormat("{{sg_instance}}").Unit(monitoring.BitsPerSecond),
				},
				{
					Name:           "node_io_network_saturation",
					Description:    "network IO saturation (Tx packets dropped)",
					Query:          "sum(rate(node_network_transmit_drop_total{sg_job=~\"" + job + "\",sg_instance=~\"" + instanceFilter + "\"}[$__rate_interval])) by(sg_instance)",
					NoAlert:        true,
					Interpretation: "Number of dropped transmitted packets. This can happen if the receiving side's receive queues/buffers become full due to slow packet processing throughput, the network link is congested etc.",
					Panel:          monitoring.Panel().LegendFormat("{{sg_instance}}"),
				},
				{
					Name:           "node_io_network_saturation",
					Description:    "network IO errors (Tx)",
					Query:          "sum(rate(node_network_transmit_errs_total{sg_job=~\"" + job + "\",sg_instance=~\"" + instanceFilter + "\"}[$__rate_interval])) by(sg_instance)",
					NoAlert:        true,
					Interpretation: "Number of packet transmission errors. This is distinct from tx packet dropping, and can indicate a failing NIC, improperly configured network options anywhere along the line, signal noise etc.",
					Panel:          monitoring.Panel().LegendFormat("{{sg_instance}}"),
				},
			},
		},
	}
}

// Below are additional panels that are excluded from the dashboard but kept for reference

// CPU load panels
/* {
	{
		Name:           "node_cpu_saturation_load1",
		Description:    "host CPU saturation (1min average)",
		Query:          "sum(node_load1{job=~\""+job+"\",sg_instance=~\"$instance\"}) by (sg_instance) / count(node_cpu_seconds_total{job=~\""+job+"\",mode=\"system\",sg_instance=~\"$instance\"}) by (sg_instance) * 100",
		NoAlert:        true,
		Interpretation: "banana",
		Panel:          monitoring.Panel().LegendFormat("{{instance}}").Unit(monitoring.Percentage),
	},
	{
		Name:           "node_cpu_saturation_load5",
		Description:    "host CPU saturation (5min average)",
		Query:          "sum(node_load5{job=~\""+job+"\",sg_instance=~\"$instance\"}) by (sg_instance) / count(node_cpu_seconds_total{job=~\""+job+"\",mode=\"system\",sg_instance=~\"$instance\"}) by (sg_instance) * 100",
		NoAlert:        true,
		Interpretation: "banana",
		Panel:          monitoring.Panel().LegendFormat("{{instance}}").Unit(monitoring.Percentage),
	},
} */

// Memory page fault panel
/* {
	Name:           "node_memory_saturation",
	Description:    "host memory saturation (major page fault rate)",
	Query:          "sum(rate(node_vmstat_pgmajfault{job=~\""+job+"\",sg_instance=~\"$instance\"}[$__rate_interval])) by (sg_instance)",
	NoAlert:        true,
	Interpretation: "banana",
	Panel:          monitoring.Panel().LegendFormat("{{instance}}"),
} */

// Disk saturation panel
/* {
	Name:           "node_io_disk_saturation_pressure_some",
	Description:    "disk IO saturation (some-processes time waiting)",
	Query:          "rate(node_pressure_io_waiting_seconds_total{job=~\""+job+"\",sg_instance=~\"$instance\"}[$__rate_interval])-rate(node_pressure_io_stalled_seconds_total{job=~\""+job+"\",sg_instance=~\"$instance\"}[$__rate_interval])",
	NoAlert:        true,
	Interpretation: "banana",
	Panel:          monitoring.Panel().LegendFormat("{{instance}}").Unit(monitoring.Seconds),
} */
