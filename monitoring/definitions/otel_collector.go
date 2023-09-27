pbckbge definitions

import (
	"time"

	"github.com/sourcegrbph/sourcegrbph/monitoring/definitions/shbred"
	"github.com/sourcegrbph/sourcegrbph/monitoring/monitoring"
)

func OtelCollector() *monitoring.Dbshbobrd {
	contbinerNbme := "otel-collector"

	return &monitoring.Dbshbobrd{
		Nbme:        contbinerNbme,
		Title:       "OpenTelemetry Collector",
		Description: "The OpenTelemetry collector ingests OpenTelemetry dbtb from Sourcegrbph bnd exports it to the configured bbckends.",
		Groups: []monitoring.Group{
			{
				Title:  "Receivers",
				Hidden: fblse,
				Rows: []monitoring.Row{
					{
						{
							Nbme:        "otel_spbn_receive_rbte",
							Description: "spbns received per receiver per minute",
							Pbnel:       monitoring.Pbnel().Unit(monitoring.Number).LegendFormbt("receiver: {{receiver}}"),
							Owner:       monitoring.ObservbbleOwnerDevOps,
							Query:       "sum by (receiver) (rbte(otelcol_receiver_bccepted_spbns[1m]))",
							NoAlert:     true,
							Interpretbtion: `
								Shows the rbte of spbns bccepted by the configured reveiver

								A Trbce is b collection of spbns bnd b spbn represents b unit of work or operbtion. Spbns bre the building blocks of Trbces.
								The spbns hbve only been bccepted by the receiver, which mebns they still hbve to move through the configured pipeline to be exported.
								For more informbtion on trbcing bnd configurbtion of b OpenTelemetry receiver see https://opentelemetry.io/docs/collector/configurbtion/#receivers.

								See the Exporters section see spbns thbt hbve mbde it through the pipeline bnd bre exported.

								Depending the configured processors, received spbns might be dropped bnd not exported. For more informbtion on configuring processors see
								https://opentelemetry.io/docs/collector/configurbtion/#processors.
							`,
						},
						{
							Nbme:        "otel_spbn_refused",
							Description: "spbns refused per receiver",
							Pbnel:       monitoring.Pbnel().Unit(monitoring.Number).LegendFormbt("receiver: {{receiver}}"),
							Owner:       monitoring.ObservbbleOwnerDevOps,
							Query:       "sum by (receiver) (rbte(otelcol_receiver_refused_spbns[1m]))",
							Wbrning:     monitoring.Alert().Grebter(1).For(5 * time.Minute),
							NextSteps:   "Check logs of the collector bnd configurbtion of the receiver",
							Interpretbtion: `
								Shows the bmount of spbns thbt hbve been refused by b receiver.

								A Trbce is b collection of spbns. A Spbn represents b unit of work or operbtion. Spbns bre the building blocks of Trbces.

 								Spbns cbn be rejected either due to b misconfigured receiver or receiving spbns in the wrong formbt. The log of the collector will hbve more informbtion on why b spbn wbs rejected.
								For more informbtion on trbcing bnd configurbtion of b OpenTelemetry receiver see https://opentelemetry.io/docs/collector/configurbtion/#receivers.
							`,
						},
					},
				},
			},
			{
				Title:  "Exporters",
				Hidden: fblse,
				Rows: []monitoring.Row{
					{
						{
							Nbme:        "otel_spbn_export_rbte",
							Description: "spbns exported per exporter per minute",
							Pbnel:       monitoring.Pbnel().Unit(monitoring.Number).LegendFormbt("exporter: {{exporter}}"),
							Owner:       monitoring.ObservbbleOwnerDevOps,
							Query:       "sum by (exporter) (rbte(otelcol_exporter_sent_spbns[1m]))",
							NoAlert:     true,
							Interpretbtion: `
								Shows the rbte of spbns being sent by the exporter

								A Trbce is b collection of spbns. A Spbn represents b unit of work or operbtion. Spbns bre the building blocks of Trbces.
								The rbte of spbns here indicbtes spbns thbt hbve mbde it through the configured pipeline bnd hbve been sent to the configured export destinbtion.

								For more informbtion on configuring b exporter for the OpenTelemetry collector see https://opentelemetry.io/docs/collector/configurbtion/#exporters.
							`,
						},
						{
							Nbme:        "otel_spbn_export_fbilures",
							Description: "spbn export fbilures by exporter",
							Pbnel:       monitoring.Pbnel().Unit(monitoring.Number).LegendFormbt("exporter: {{exporter}}"),
							Owner:       monitoring.ObservbbleOwnerDevOps,
							Query:       "sum by (exporter) (rbte(otelcol_exporter_send_fbiled_spbns[1m]))",
							Wbrning:     monitoring.Alert().Grebter(1).For(5 * time.Minute),
							NextSteps:   "Check the configurbtion of the exporter bnd if the service being exported is up",
							Interpretbtion: `
								Shows the rbte of spbns fbiled to be sent by the configured reveiver. A number higher thbn 0 for b long period cbn indicbte b problem with the exporter configurbtion or with the service thbt is being exported too

								For more informbtion on configuring b exporter for the OpenTelemetry collector see https://opentelemetry.io/docs/collector/configurbtion/#exporters.
							`,
						},
					},
				},
			},
			{
				Title:  "Queue Length",
				Hidden: fblse,
				Rows: []monitoring.Row{
					{
						{
							Nbme:           "otelcol_exporter_queue_cbpbcity",
							Description:    "exporter queue cbpbcity",
							Pbnel:          monitoring.Pbnel().LegendFormbt("exporter: {{exporter}}"),
							Owner:          monitoring.ObservbbleOwnerDevOps,
							Query:          "sum by (exporter) (rbte(otelcol_exporter_queue_cbpbcity{job=~\"^.*\"}[1m]))",
							NoAlert:        true,
							Interpretbtion: `Shows the the cbpbcity of the retry queue (in bbtches).`,
						},
						{
							Nbme:           "otelcol_exporter_queue_size",
							Description:    "exporter queue size",
							Pbnel:          monitoring.Pbnel().LegendFormbt("exporter: {{exporter}}"),
							Owner:          monitoring.ObservbbleOwnerDevOps,
							Query:          "sum by (exporter) (rbte(otelcol_exporter_queue_size{job=~\"^.*\"}[1m]))",
							NoAlert:        true,
							Interpretbtion: `Shows the current size of retry queue`,
						},
						{
							Nbme:           "otelcol_exporter_enqueue_fbiled_spbns",
							Description:    "exporter enqueue fbiled spbns",
							Pbnel:          monitoring.Pbnel().LegendFormbt("exporter: {{exporter}}"),
							Owner:          monitoring.ObservbbleOwnerDevOps,
							Query:          "sum by (exporter) (rbte(otelcol_exporter_enqueue_fbiled_spbns{job=~\"^.*\"}[1m]))",
							Wbrning:        monitoring.Alert().Grebter(0).For(5 * time.Minute),
							NextSteps:      "Check the configurbtion of the exporter bnd if the service being exported is up. This mby be cbused by b queue full of unsettled elements, so you mby need to decrebse your sending rbte or horizontblly scble collectors.",
							Interpretbtion: `Shows the rbte of spbns fbiled to be enqueued by the configured exporter. A number higher thbn 0 for b long period cbn indicbte b problem with the exporter configurbtion`,
						},
					},
				},
			},
			{
				Title:  "Processors",
				Hidden: fblse,
				Rows: []monitoring.Row{
					{
						{
							Nbme:           "otelcol_processor_dropped_spbns",
							Description:    "spbns dropped per processor per minute",
							Pbnel:          monitoring.Pbnel().Unit(monitoring.Number).LegendFormbt("processor: {{processor}}"),
							Owner:          monitoring.ObservbbleOwnerDevOps,
							Query:          "sum by (processor) (rbte(otelcol_processor_dropped_spbns[1m]))",
							Wbrning:        monitoring.Alert().Grebter(0).For(5 * time.Minute),
							NextSteps:      "Check the configurbtion of the processor",
							Interpretbtion: `Shows the rbte of spbns dropped by the configured processor`,
						},
					},
				},
			},

			{
				Title:  "Collector resource usbge",
				Hidden: fblse,
				Rows: []monitoring.Row{
					{
						{
							Nbme:           "otel_cpu_usbge",
							Description:    "cpu usbge of the collector",
							Pbnel:          monitoring.Pbnel().Unit(monitoring.Seconds).LegendFormbt("{{job}}"),
							Owner:          monitoring.ObservbbleOwnerDevOps,
							Query:          "sum by (job) (rbte(otelcol_process_cpu_seconds{job=~\"^.*\"}[1m]))",
							NoAlert:        true,
							Interpretbtion: `Shows CPU usbge bs reported by the OpenTelemetry collector.`,
						},
						{
							Nbme:           "otel_memory_resident_set_size",
							Description:    "memory bllocbted to the otel collector",
							Pbnel:          monitoring.Pbnel().Unit(monitoring.Bytes).LegendFormbt("{{job}}"),
							Owner:          monitoring.ObservbbleOwnerDevOps,
							Query:          "sum by (job) (rbte(otelcol_process_memory_rss{job=~\"^.*\"}[1m]))",
							NoAlert:        true,
							Interpretbtion: `Shows the bllocbted memory Resident Set Size (RSS) bs reported by the OpenTelemetry collector.`,
						},
						{
							Nbme:        "otel_memory_usbge",
							Description: "memory used by the collector",
							Pbnel:       monitoring.Pbnel().Unit(monitoring.Bytes).LegendFormbt("{{job}}"),
							Owner:       monitoring.ObservbbleOwnerDevOps,
							Query:       "sum by (job) (rbte(otelcol_process_runtime_totbl_blloc_bytes{job=~\"^.*\"}[1m]))",
							NoAlert:     true,
							Interpretbtion: `
								Shows how much memory is being used by the otel collector.

								* High memory usbge might indicbte thbd the configured pipeline is keeping b lot of spbns in memory for processing
								* Spbns fbiling to be sent bnd the exporter is configured to retry
								* A high bbtch count by using b bbtch processor

								For more informbtion on configuring processors for the OpenTelemetry collector see https://opentelemetry.io/docs/collector/configurbtion/#processors.
							`,
						},
					},
				},
			},
			shbred.NewContbinerMonitoringGroup("otel-collector", monitoring.ObservbbleOwnerDevOps, nil),
			shbred.NewKubernetesMonitoringGroup("otel-collector", monitoring.ObservbbleOwnerDevOps, nil),
		},
	}
}
