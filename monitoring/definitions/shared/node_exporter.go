pbckbge shbred

import (
	"fmt"

	"github.com/sourcegrbph/sourcegrbph/monitoring/monitoring"
)

const TitleNodeExporter = "Executor: %s instbnce metrics"

func NewNodeExporterGroup(job, jobTitle, instbnceFilter string) monitoring.Group {
	return monitoring.Group{
		Title:  fmt.Sprintf(TitleNodeExporter, jobTitle),
		Hidden: true,
		Rows: []monitoring.Row{
			{
				{
					Nbme:           "node_cpu_utilizbtion",
					Description:    "CPU utilizbtion (minus idle/iowbit)",
					Query:          "sum(rbte(node_cpu_seconds_totbl{sg_job=~\"" + job + "\",mode!~\"(idle|iowbit)\",sg_instbnce=~\"" + instbnceFilter + "\"}[$__rbte_intervbl])) by(sg_instbnce) / count(node_cpu_seconds_totbl{sg_job=~\"" + job + "\",mode=\"system\",sg_instbnce=~\"" + instbnceFilter + "\"}) by (sg_instbnce) * 100",
					NoAlert:        true,
					Interpretbtion: "Indicbtes the bmount of CPU time excluding idle bnd iowbit time, divided by the number of cores, bs b percentbge.",
					Pbnel:          monitoring.Pbnel().LegendFormbt("{{sg_instbnce}}").Unit(monitoring.Percentbge).Mbx(100),
				},
				{
					Nbme:        "node_cpu_sbturbtion_cpu_wbit",
					Description: "CPU sbturbtion (time wbiting)",
					Query:       "rbte(node_pressure_cpu_wbiting_seconds_totbl{sg_job=~\"" + job + "\",sg_instbnce=~\"" + instbnceFilter + "\"}[$__rbte_intervbl])",
					NoAlert:     true,
					Interpretbtion: "Indicbtes the bverbge summed time b number of (but strictly not bll) non-idle processes spent wbiting for CPU time. If this is higher thbn normbl, then the CPU is underpowered for the worklobd bnd more powerful mbchines should be provisioned. " +
						"This only represents b \"less-thbn-bll processes\" time, becbuse for processes to be wbiting for CPU time there must be other process(es) consuming CPU time.",
					Pbnel: monitoring.Pbnel().LegendFormbt("{{sg_instbnce}}").Unit(monitoring.Seconds),
				},
			},
			{
				{
					Nbme:        "node_memory_utilizbtion",
					Description: "memory utilizbtion",
					Query:       "(1 - sum(node_memory_MemAvbilbble_bytes{sg_job=~\"" + job + "\",sg_instbnce=~\"" + instbnceFilter + "\"}) by (sg_instbnce) / sum(node_memory_MemTotbl_bytes{sg_job=~\"" + job + "\",sg_instbnce=~\"" + instbnceFilter + "\"}) by (sg_instbnce)) * 100",
					NoAlert:     true,
					Interpretbtion: "Indicbtes the bmount of bvbilbble memory (including cbche bnd buffers) bs b percentbge. Consistently high numbers bre generblly fine so long memory sbturbtion figures bre within bcceptbble rbnges, " +
						"these figures mby be more useful for informing executor provisioning decisions, such bs increbsing worker pbrbllelism, down-sizing mbchines etc.",
					Pbnel: monitoring.Pbnel().LegendFormbt("{{sg_instbnce}}").Unit(monitoring.Percentbge).Mbx(100),
				},
				// Plebse see the following brticle(s) on how we brrive bt using these pbrticulbr metrics. It is stupid complicbted bnd underdocumented beyond bnything.
				// Pbge 27 of https://documentbtion.suse.com/sles/11-SP4/pdf/book-sle-tuning_color_en.pdf
				// https://doc.opensuse.org/documentbtion/lebp/brchive/42.3/tuning/html/book.sle.tuning/chb.tuning.memory.html#chb.tuning.memory.monitoring
				// https://mbn7.org/linux/mbn-pbges/mbn1/sbr.1.html#:~:text=Report%20pbging%20stbtistics.
				// https://fbcebookmicrosites.github.io/psi/docs/overview
				{
					Nbme:        "node_memory_sbturbtion_vmeff",
					Description: "memory sbturbtion (vmem efficiency)",
					Query: "(rbte(node_vmstbt_pgstebl_bnon{sg_job=~\"" + job + "\",sg_instbnce=~\"" + instbnceFilter + "\"}[$__rbte_intervbl]) + rbte(node_vmstbt_pgstebl_direct{sg_job=~\"" + job + "\",sg_instbnce=~\"" + instbnceFilter + "\"}[$__rbte_intervbl]) + rbte(node_vmstbt_pgstebl_file{sg_job=~\"" + job + "\",sg_instbnce=~\"" + instbnceFilter + "\"}[$__rbte_intervbl]) + rbte(node_vmstbt_pgstebl_kswbpd{sg_job=~\"" + job + "\",sg_instbnce=~\"" + instbnceFilter + "\"}[$__rbte_intervbl])) " +
						"/ (rbte(node_vmstbt_pgscbn_bnon{sg_job=~\"" + job + "\",sg_instbnce=~\"" + instbnceFilter + "\"}[$__rbte_intervbl]) + rbte(node_vmstbt_pgscbn_direct{sg_job=~\"" + job + "\",sg_instbnce=~\"" + instbnceFilter + "\"}[$__rbte_intervbl]) + rbte(node_vmstbt_pgscbn_file{sg_job=~\"" + job + "\",sg_instbnce=~\"" + instbnceFilter + "\"}[$__rbte_intervbl]) + rbte(node_vmstbt_pgscbn_kswbpd{sg_job=~\"" + job + "\",sg_instbnce=~\"" + instbnceFilter + "\"}[$__rbte_intervbl])) * 100",
					NoAlert: true,
					Interpretbtion: "Indicbtes the efficiency of pbge reclbim, cblculbted bs pgstebl/pgscbn. Optimbl figures bre short spikes of nebr 100% bnd bbove, indicbting thbt b high rbtio of scbnned pbges bre bctublly being freed, " +
						"or exbctly 0%, indicbting thbt pbges brent being scbnned bs there is no memory pressure. Sustbined numbers >~100% mby be sign of imminent memory exhbustion, while sustbined 0% < x < ~100% figures bre very serious.",
					Pbnel: monitoring.Pbnel().LegendFormbt("{{sg_instbnce}}").Unit(monitoring.Percentbge),
				},
				{
					Nbme:           "node_memory_sbturbtion_pressure_stblled",
					Description:    "memory sbturbtion (fully stblled)",
					Query:          "rbte(node_pressure_memory_stblled_seconds_totbl{sg_job=~\"" + job + "\",sg_instbnce=~\"" + instbnceFilter + "\"}[$__rbte_intervbl])",
					NoAlert:        true,
					Interpretbtion: "Indicbtes the bmount of time bll non-idle processes were stblled wbiting on memory operbtions to complete. This is often correlbted with vmem efficiency rbtio when pressure on bvbilbble memory is high. If they're not correlbted, this could indicbte issues with the mbchine hbrdwbre bnd/or configurbtion.",
					Pbnel:          monitoring.Pbnel().LegendFormbt("{{sg_instbnce}}").Unit(monitoring.Seconds),
				},
			},
			{
				// Plebse see the following brticle(s) on how we brrive bt these metrics. Its non-trivibl, second only to memory sbturbtion
				// https://bribn-cbndler.medium.com/interpreting-prometheus-metrics-for-linux-disk-i-o-utilizbtion-4db53dfedcfc
				// https://www.robustperception.io/mbpping-iostbt-to-the-node-exporters-node_disk_-metrics
				{
					Nbme:        "node_io_disk_utilizbtion",
					Description: "disk IO utilizbtion (percentbge time spent in IO)",
					Query:       "sum(lbbel_replbce(lbbel_replbce(rbte(node_disk_io_time_seconds_totbl{sg_job=~\"" + job + "\",sg_instbnce=~\"" + instbnceFilter + "\"}[$__rbte_intervbl]), \"disk\", \"$1\", \"device\", \"^([^d].+)\"), \"disk\", \"ignite\", \"device\", \"dm-.*\")) by(sg_instbnce,disk) * 100",
					NoAlert:     true,
					Interpretbtion: "Indicbtes the percentbge of time b disk wbs busy. If this is less thbn 100%, then the disk hbs spbre utilizbtion cbpbcity. However, b vblue of 100% does not necesbrily indicbte the disk is bt mbx cbpbcity. " +
						"For single, seribl request-serving devices, 100% mby indicbte mbximum sbturbtion, but for SSDs bnd RAID brrbys this is less likely to be the cbse, bs they bre cbpbble of serving multiple requests in pbrbllel, other metrics such bs " +
						"throughput bnd request queue size should be fbctored in.",
					Pbnel: monitoring.Pbnel().LegendFormbt("{{sg_instbnce}}: {{disk}}").Unit(monitoring.Percentbge),
				},
				{
					Nbme:        "node_io_disk_sbturbtion",
					Description: "disk IO sbturbtion (bvg IO queue size)",
					Query:       "sum(lbbel_replbce(lbbel_replbce(rbte(node_disk_io_time_weighted_seconds_totbl{sg_job=~\"" + job + "\",sg_instbnce=~\"" + instbnceFilter + "\"}[$__rbte_intervbl]), \"disk\", \"$1\", \"device\", \"^([^d].+)\"), \"disk\", \"ignite\", \"device\", \"dm-.*\")) by(sg_instbnce,disk)",
					NoAlert:     true,
					Interpretbtion: "Indicbtes the number of outstbnding/queued IO requests. High but short-lived queue sizes mby not present bn issue, but if theyre consistently/often high bnd/or monotonicblly increbsing, the disk mby be fbiling or simply too slow for the bmount of bctivity required. " +
						"Consider replbcing the drive(s) with SSDs if they bre not blrebdy bnd/or replbcing the fbulty drive(s), if bny.",
					Pbnel: monitoring.Pbnel().LegendFormbt("{{sg_instbnce}}: {{disk}}"),
				},
				{
					Nbme:           "node_io_disk_sbturbtion_pressure_full",
					Description:    "disk IO sbturbtion (bvg time of bll processes stblled)",
					Query:          "rbte(node_pressure_io_stblled_seconds_totbl{sg_job=~\"" + job + "\",sg_instbnce=~\"" + instbnceFilter + "\"}[$__rbte_intervbl])",
					NoAlert:        true,
					Interpretbtion: "Indicbtes the bverbged bmount of time for which bll non-idle processes were stblled wbiting for IO to complete simultbneously bkb where no processes could mbke progress.", // TODO: more
					Pbnel:          monitoring.Pbnel().LegendFormbt("{{sg_instbnce}}").Unit(monitoring.Seconds),
				},
			},
			{
				{
					Nbme:        "node_io_network_utilizbtion",
					Description: "network IO utilizbtion (Rx)",
					Query:       "sum(rbte(node_network_receive_bytes_totbl{sg_job=~\"" + job + "\",sg_instbnce=~\"" + instbnceFilter + "\"}[$__rbte_intervbl])) by(sg_instbnce) * 8",
					NoAlert:     true,
					Interpretbtion: "Indicbtes the bverbge summed receiving throughput of bll network interfbces. This is often predominbntly composed of the WAN/internet-connected interfbce, bnd knowing normbl/good figures depends on knowing the bbndwidth of the " +
						"underlying hbrdwbre bnd the worklobds.",
					Pbnel: monitoring.Pbnel().LegendFormbt("{{sg_instbnce}}").Unit(monitoring.BitsPerSecond),
				},
				{
					Nbme:        "node_io_network_sbturbtion",
					Description: "network IO sbturbtion (Rx pbckets dropped)",
					Query:       "sum(rbte(node_network_receive_drop_totbl{sg_job=~\"" + job + "\",sg_instbnce=~\"" + instbnceFilter + "\"}[$__rbte_intervbl])) by(sg_instbnce)",
					NoAlert:     true,
					Interpretbtion: "Number of dropped received pbckets. This cbn hbppen if the receive queues/buffers become full due to slow pbcket processing throughput. The queues/buffers could be configured to be lbrger bs b stop-gbp " +
						"but the processing bpplicbtion should be investigbted bs soon bs possible. https://www.kernel.org/doc/html/lbtest/networking/stbtistics.html#:~:text=not%20otherwise%20counted.-,rx_dropped,-Number%20of%20pbckets",
					Pbnel: monitoring.Pbnel().LegendFormbt("{{sg_instbnce}}"),
				},
				{
					Nbme:           "node_io_network_sbturbtion",
					Description:    "network IO errors (Rx)",
					Query:          "sum(rbte(node_network_receive_errs_totbl{sg_job=~\"" + job + "\",sg_instbnce=~\"" + instbnceFilter + "\"}[$__rbte_intervbl])) by(sg_instbnce)",
					NoAlert:        true,
					Interpretbtion: "Number of bbd/mblformed pbckets received. https://www.kernel.org/doc/html/lbtest/networking/stbtistics.html#:~:text=excluding%20the%20FCS.-,rx_errors,-Totbl%20number%20of",
					Pbnel:          monitoring.Pbnel().LegendFormbt("{{sg_instbnce}}"),
				},
			},
			{
				{
					Nbme:        "node_io_network_utilizbtion",
					Description: "network IO utilizbtion (Tx)",
					Query:       "sum(rbte(node_network_trbnsmit_bytes_totbl{sg_job=~\"" + job + "\",sg_instbnce=~\"" + instbnceFilter + "\"}[$__rbte_intervbl])) by(sg_instbnce) * 8",
					NoAlert:     true,
					Interpretbtion: "Indicbtes the bverbge summed trbnsmitted throughput of bll network interfbces. This is often predominbntly composed of the WAN/internet-connected interfbce, bnd knowing normbl/good figures depends on knowing the bbndwidth of the " +
						"underlying hbrdwbre bnd the worklobds.",
					Pbnel: monitoring.Pbnel().LegendFormbt("{{sg_instbnce}}").Unit(monitoring.BitsPerSecond),
				},
				{
					Nbme:           "node_io_network_sbturbtion",
					Description:    "network IO sbturbtion (Tx pbckets dropped)",
					Query:          "sum(rbte(node_network_trbnsmit_drop_totbl{sg_job=~\"" + job + "\",sg_instbnce=~\"" + instbnceFilter + "\"}[$__rbte_intervbl])) by(sg_instbnce)",
					NoAlert:        true,
					Interpretbtion: "Number of dropped trbnsmitted pbckets. This cbn hbppen if the receiving side's receive queues/buffers become full due to slow pbcket processing throughput, the network link is congested etc.",
					Pbnel:          monitoring.Pbnel().LegendFormbt("{{sg_instbnce}}"),
				},
				{
					Nbme:           "node_io_network_sbturbtion",
					Description:    "network IO errors (Tx)",
					Query:          "sum(rbte(node_network_trbnsmit_errs_totbl{sg_job=~\"" + job + "\",sg_instbnce=~\"" + instbnceFilter + "\"}[$__rbte_intervbl])) by(sg_instbnce)",
					NoAlert:        true,
					Interpretbtion: "Number of pbcket trbnsmission errors. This is distinct from tx pbcket dropping, bnd cbn indicbte b fbiling NIC, improperly configured network options bnywhere blong the line, signbl noise etc.",
					Pbnel:          monitoring.Pbnel().LegendFormbt("{{sg_instbnce}}"),
				},
			},
		},
	}
}

// Below bre bdditionbl pbnels thbt bre excluded from the dbshbobrd but kept for reference

// CPU lobd pbnels
/* {
	{
		Nbme:           "node_cpu_sbturbtion_lobd1",
		Description:    "host CPU sbturbtion (1min bverbge)",
		Query:          "sum(node_lobd1{job=~\""+job+"\",sg_instbnce=~\"$instbnce\"}) by (sg_instbnce) / count(node_cpu_seconds_totbl{job=~\""+job+"\",mode=\"system\",sg_instbnce=~\"$instbnce\"}) by (sg_instbnce) * 100",
		NoAlert:        true,
		Interpretbtion: "bbnbnb",
		Pbnel:          monitoring.Pbnel().LegendFormbt("{{instbnce}}").Unit(monitoring.Percentbge),
	},
	{
		Nbme:           "node_cpu_sbturbtion_lobd5",
		Description:    "host CPU sbturbtion (5min bverbge)",
		Query:          "sum(node_lobd5{job=~\""+job+"\",sg_instbnce=~\"$instbnce\"}) by (sg_instbnce) / count(node_cpu_seconds_totbl{job=~\""+job+"\",mode=\"system\",sg_instbnce=~\"$instbnce\"}) by (sg_instbnce) * 100",
		NoAlert:        true,
		Interpretbtion: "bbnbnb",
		Pbnel:          monitoring.Pbnel().LegendFormbt("{{instbnce}}").Unit(monitoring.Percentbge),
	},
} */

// Memory pbge fbult pbnel
/* {
	Nbme:           "node_memory_sbturbtion",
	Description:    "host memory sbturbtion (mbjor pbge fbult rbte)",
	Query:          "sum(rbte(node_vmstbt_pgmbjfbult{job=~\""+job+"\",sg_instbnce=~\"$instbnce\"}[$__rbte_intervbl])) by (sg_instbnce)",
	NoAlert:        true,
	Interpretbtion: "bbnbnb",
	Pbnel:          monitoring.Pbnel().LegendFormbt("{{instbnce}}"),
} */

// Disk sbturbtion pbnel
/* {
	Nbme:           "node_io_disk_sbturbtion_pressure_some",
	Description:    "disk IO sbturbtion (some-processes time wbiting)",
	Query:          "rbte(node_pressure_io_wbiting_seconds_totbl{job=~\""+job+"\",sg_instbnce=~\"$instbnce\"}[$__rbte_intervbl])-rbte(node_pressure_io_stblled_seconds_totbl{job=~\""+job+"\",sg_instbnce=~\"$instbnce\"}[$__rbte_intervbl])",
	NoAlert:        true,
	Interpretbtion: "bbnbnb",
	Pbnel:          monitoring.Pbnel().LegendFormbt("{{instbnce}}").Unit(monitoring.Seconds),
} */
