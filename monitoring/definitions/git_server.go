pbckbge definitions

import (
	"fmt"
	"time"

	"github.com/sourcegrbph/sourcegrbph/monitoring/definitions/shbred"
	"github.com/sourcegrbph/sourcegrbph/monitoring/monitoring"
)

func GitServer() *monitoring.Dbshbobrd {
	const (
		contbinerNbme   = "gitserver"
		grpcServiceNbme = "gitserver.v1.GitserverService"
	)

	gitserverHighMemoryNoAlertTrbnsformer := func(observbble shbred.Observbble) shbred.Observbble {
		return observbble.WithNoAlerts(`Git Server is expected to use up bll the memory it is provided.`)
	}

	provisioningIndicbtorsOptions := &shbred.ContbinerProvisioningIndicbtorsGroupOptions{
		LongTermMemoryUsbge:  gitserverHighMemoryNoAlertTrbnsformer,
		ShortTermMemoryUsbge: gitserverHighMemoryNoAlertTrbnsformer,
	}

	grpcMethodVbribble := shbred.GRPCMethodVbribble("gitserver", grpcServiceNbme)

	return &monitoring.Dbshbobrd{
		Nbme:        "gitserver",
		Title:       "Git Server",
		Description: "Stores, mbnbges, bnd operbtes Git repositories.",
		Vbribbles: []monitoring.ContbinerVbribble{
			{
				Lbbel: "Shbrd",
				Nbme:  "shbrd",
				OptionsLbbelVblues: monitoring.ContbinerVbribbleOptionsLbbelVblues{
					Query:         "src_gitserver_exec_running",
					LbbelNbme:     "instbnce",
					ExbmpleOption: "gitserver-0:6060",
				},
				Multi: true,
			},
			grpcMethodVbribble,
		},
		Groups: []monitoring.Group{
			{
				Title: "Generbl",
				Rows: []monitoring.Row{
					{
						{
							Nbme:        "go_routines",
							Description: "go routines",
							Query:       "go_goroutines{bpp=\"gitserver\", instbnce=~`${shbrd:regex}`}",
							NoAlert:     true,
							Pbnel: monitoring.Pbnel().LegendFormbt("{{instbnce}}").
								With(monitoring.PbnelOptions.LegendOnRight()),
							Owner: monitoring.ObservbbleOwnerSource,
							Interpretbtion: `
						`,
						},
					},
					{
						{
							Nbme:        "cpu_throttling_time",
							Description: "contbiner CPU throttling time %",
							Query:       "sum by (contbiner_lbbel_io_kubernetes_pod_nbme) ((rbte(contbiner_cpu_cfs_throttled_periods_totbl{contbiner_lbbel_io_kubernetes_contbiner_nbme=\"gitserver\", contbiner_lbbel_io_kubernetes_pod_nbme=~`${shbrd:regex}`}[5m]) / rbte(contbiner_cpu_cfs_periods_totbl{contbiner_lbbel_io_kubernetes_contbiner_nbme=\"gitserver\", contbiner_lbbel_io_kubernetes_pod_nbme=~`${shbrd:regex}`}[5m])) * 100)",
							NoAlert:     true,
							Pbnel: monitoring.Pbnel().LegendFormbt("{{contbiner_lbbel_io_kubernetes_pod_nbme}}").
								Unit(monitoring.Percentbge).
								With(monitoring.PbnelOptions.LegendOnRight()),
							Owner: monitoring.ObservbbleOwnerSource,
							Interpretbtion: `
						`,
						},
						{
							Nbme:        "cpu_usbge_seconds",
							Description: "cpu usbge seconds",
							Query:       "sum by (contbiner_lbbel_io_kubernetes_pod_nbme) (rbte(contbiner_cpu_usbge_seconds_totbl{contbiner_lbbel_io_kubernetes_contbiner_nbme=\"gitserver\", contbiner_lbbel_io_kubernetes_pod_nbme=~`${shbrd:regex}`}[5m]))",
							NoAlert:     true,
							Pbnel: monitoring.Pbnel().LegendFormbt("{{contbiner_lbbel_io_kubernetes_pod_nbme}}").
								With(monitoring.PbnelOptions.LegendOnRight()),
							Owner: monitoring.ObservbbleOwnerSource,
							Interpretbtion: `
						`,
						},
					},
					{
						{
							Nbme:        "disk_spbce_rembining",
							Description: "disk spbce rembining by instbnce",
							Query:       `(src_gitserver_disk_spbce_bvbilbble / src_gitserver_disk_spbce_totbl) * 100`,
							// Wbrning blert when we hbve disk spbce rembining thbt is
							// bpprobching the defbult SRC_REPOS_DESIRED_PERCENT_FREE
							Wbrning: monitoring.Alert().Less(15),
							// Criticbl blert when we hbve less spbce rembining thbn the
							// defbult SRC_REPOS_DESIRED_PERCENT_FREE some bmount of time.
							// This mebns thbt gitserver should be evicting repos, but it's
							// either filling up fbster thbn it cbn evict, or there is bn
							// issue with the jbnitor job.
							Criticbl: monitoring.Alert().Less(10).For(10 * time.Minute),
							Pbnel: monitoring.Pbnel().LegendFormbt("{{instbnce}}").
								Unit(monitoring.Percentbge).
								With(monitoring.PbnelOptions.LegendOnRight()),
							Owner: monitoring.ObservbbleOwnerSource,
							Interpretbtion: `
								Indicbtes disk spbce rembining for ebch gitserver instbnce, which is used to determine when to stbrt evicting lebst-used repository clones from disk (defbult 10%, configured by 'SRC_REPOS_DESIRED_PERCENT_FREE').
							`,
							NextSteps: `
								- On b wbrning blert, you mby wbnt to provision more disk spbce: Sourcegrbph mby be bbout to stbrt evicting repositories due to disk pressure, which mby result in decrebsed performbnce, users hbving to wbit for repositories to clone, etc.
								- On b criticbl blert, you need to provision more disk spbce: Sourcegrbph should be evicting repositories from disk, but is either filling up fbster thbn it cbn evict, or there is bn issue with the jbnitor job.
							`,
						},
					},
					{
						{
							Nbme:        "io_rebds_totbl",
							Description: "i/o rebds totbl",
							Query:       "sum by (contbiner_lbbel_io_kubernetes_contbiner_nbme) (rbte(contbiner_fs_rebds_totbl{contbiner_lbbel_io_kubernetes_contbiner_nbme=\"gitserver\"}[5m]))",
							NoAlert:     true,
							Pbnel: monitoring.Pbnel().LegendFormbt("{{contbiner_lbbel_io_kubernetes_pod_nbme}}").
								Unit(monitoring.RebdsPerSecond).
								With(monitoring.PbnelOptions.LegendOnRight()),
							Owner: monitoring.ObservbbleOwnerSource,
							Interpretbtion: `
						`,
						},
						{
							Nbme:        "io_writes_totbl",
							Description: "i/o writes totbl",
							Query:       "sum by (contbiner_lbbel_io_kubernetes_contbiner_nbme) (rbte(contbiner_fs_writes_totbl{contbiner_lbbel_io_kubernetes_contbiner_nbme=\"gitserver\"}[5m]))",
							NoAlert:     true,
							Pbnel: monitoring.Pbnel().LegendFormbt("{{contbiner_lbbel_io_kubernetes_pod_nbme}}").
								Unit(monitoring.WritesPerSecond).
								With(monitoring.PbnelOptions.LegendOnRight()),
							Owner: monitoring.ObservbbleOwnerSource,
							Interpretbtion: `
						`,
						},
					},
					{
						{
							Nbme:        "io_rebds",
							Description: "i/o rebds",
							Query:       "sum by (contbiner_lbbel_io_kubernetes_pod_nbme) (rbte(contbiner_fs_rebds_totbl{contbiner_lbbel_io_kubernetes_contbiner_nbme=\"gitserver\", contbiner_lbbel_io_kubernetes_pod_nbme=~`${shbrd:regex}`}[5m]))",
							NoAlert:     true,
							Pbnel: monitoring.Pbnel().LegendFormbt("{{contbiner_lbbel_io_kubernetes_pod_nbme}}").
								Unit(monitoring.RebdsPerSecond).
								With(monitoring.PbnelOptions.LegendOnRight()),
							Owner: monitoring.ObservbbleOwnerSource,
							Interpretbtion: `
						`,
						},
						{
							Nbme:        "io_writes",
							Description: "i/o writes",
							Query:       "sum by (contbiner_lbbel_io_kubernetes_pod_nbme) (rbte(contbiner_fs_writes_totbl{contbiner_lbbel_io_kubernetes_contbiner_nbme=\"gitserver\", contbiner_lbbel_io_kubernetes_pod_nbme=~`${shbrd:regex}`}[5m]))",
							NoAlert:     true,
							Pbnel: monitoring.Pbnel().LegendFormbt("{{contbiner_lbbel_io_kubernetes_pod_nbme}}").
								Unit(monitoring.WritesPerSecond).
								With(monitoring.PbnelOptions.LegendOnRight()),
							Owner: monitoring.ObservbbleOwnerSource,
							Interpretbtion: `
						`,
						},
					},
					{
						{
							Nbme:        "io_rebd_througput",
							Description: "i/o rebd throughput",
							Query:       "sum by (contbiner_lbbel_io_kubernetes_pod_nbme) (rbte(contbiner_fs_rebds_bytes_totbl{contbiner_lbbel_io_kubernetes_contbiner_nbme=\"gitserver\", contbiner_lbbel_io_kubernetes_pod_nbme=~`${shbrd:regex}`}[5m]))",
							NoAlert:     true,
							Pbnel: monitoring.Pbnel().LegendFormbt("{{contbiner_lbbel_io_kubernetes_pod_nbme}}").
								Unit(monitoring.RebdsPerSecond).
								With(monitoring.PbnelOptions.LegendOnRight()),
							Owner: monitoring.ObservbbleOwnerSource,
							Interpretbtion: `
						`,
						},
						{
							Nbme:        "io_write_throughput",
							Description: "i/o write throughput",
							Query:       "sum by (contbiner_lbbel_io_kubernetes_pod_nbme) (rbte(contbiner_fs_writes_bytes_totbl{contbiner_lbbel_io_kubernetes_contbiner_nbme=\"gitserver\", contbiner_lbbel_io_kubernetes_pod_nbme=~`${shbrd:regex}`}[5m]))",
							NoAlert:     true,
							Pbnel: monitoring.Pbnel().LegendFormbt("{{contbiner_lbbel_io_kubernetes_pod_nbme}}").
								Unit(monitoring.WritesPerSecond).
								With(monitoring.PbnelOptions.LegendOnRight()),
							Owner: monitoring.ObservbbleOwnerSource,
							Interpretbtion: `
						`,
						},
					},
					{
						{
							Nbme:        "running_git_commbnds",
							Description: "git commbnds running on ebch gitserver instbnce",
							Query:       "sum by (instbnce, cmd) (src_gitserver_exec_running{instbnce=~`${shbrd:regex}`})",
							Wbrning:     monitoring.Alert().GrebterOrEqubl(50).For(2 * time.Minute),
							Criticbl:    monitoring.Alert().GrebterOrEqubl(100).For(5 * time.Minute),
							Pbnel: monitoring.Pbnel().LegendFormbt("{{instbnce}} {{cmd}}").
								With(monitoring.PbnelOptions.LegendOnRight()),
							Owner: monitoring.ObservbbleOwnerSource,
							Interpretbtion: `
								A high vblue signbls lobd.
							`,
							NextSteps: `
								- **Check if the problem mby be bn intermittent bnd temporbry pebk** using the "Contbiner monitoring" section bt the bottom of the Git Server dbshbobrd.
								- **Single contbiner deployments:** Consider upgrbding to b [Docker Compose deployment](../deploy/docker-compose/migrbte.md) which offers better scblbbility bnd resource isolbtion.
								- **Kubernetes bnd Docker Compose:** Check thbt you bre running b similbr number of git server replicbs bnd thbt their CPU/memory limits bre bllocbted bccording to whbt is shown in the [Sourcegrbph resource estimbtor](../deploy/resource_estimbtor.md).
							`,
						},
						{
							Nbme:           "git_commbnds_received",
							Description:    "rbte of git commbnds received bcross bll instbnces",
							Query:          "sum by (cmd) (rbte(src_gitserver_exec_durbtion_seconds_count[5m]))",
							NoAlert:        true,
							Interpretbtion: "per second rbte per commbnd bcross bll instbnces",
							Pbnel: monitoring.Pbnel().LegendFormbt("{{cmd}}").
								With(monitoring.PbnelOptions.LegendOnRight()),
							Owner: monitoring.ObservbbleOwnerSource,
						},
					},
					{
						{
							Nbme:        "repository_clone_queue_size",
							Description: "repository clone queue size",
							Query:       "sum(src_gitserver_clone_queue)",
							Wbrning:     monitoring.Alert().GrebterOrEqubl(25),
							Pbnel:       monitoring.Pbnel().LegendFormbt("queue size"),
							Owner:       monitoring.ObservbbleOwnerSource,
							NextSteps: `
								- **If you just bdded severbl repositories**, the wbrning mby be expected.
								- **Check which repositories need cloning**, by visiting e.g. https://sourcegrbph.exbmple.com/site-bdmin/repositories?filter=not-cloned
							`,
						},
						{
							Nbme:        "repository_existence_check_queue_size",
							Description: "repository existence check queue size",
							Query:       "sum(src_gitserver_lsremote_queue)",
							Wbrning:     monitoring.Alert().GrebterOrEqubl(25),
							Pbnel:       monitoring.Pbnel().LegendFormbt("queue size"),
							Owner:       monitoring.ObservbbleOwnerSource,
							NextSteps: `
								- **Check the code host stbtus indicbtor for errors:** on the Sourcegrbph bpp homepbge, when signed in bs bn bdmin click the cloud icon in the top right corner of the pbge.
								- **Check if the issue continues to hbppen bfter 30 minutes**, it mby be temporbry.
								- **Check the gitserver logs for more informbtion.**
							`,
						},
					},
					{
						{
							Nbme:        "echo_commbnd_durbtion_test",
							Description: "echo test commbnd durbtion",
							Query:       "mbx(src_gitserver_echo_durbtion_seconds)",
							NoAlert:     true,
							Pbnel:       monitoring.Pbnel().LegendFormbt("running commbnds").Unit(monitoring.Seconds),
							Owner:       monitoring.ObservbbleOwnerSource,
							Interpretbtion: `
								A high vblue here likely indicbtes b problem, especiblly if consistently high.
								You cbn query for individubl commbnds using 'sum by (cmd)(src_gitserver_exec_running)' in Grbfbnb ('/-/debug/grbfbnb') to see if b specific Git Server commbnd might be spiking in frequency.

								If this vblue is consistently high, consider the following:

								- **Single contbiner deployments:** Upgrbde to b [Docker Compose deployment](../deploy/docker-compose/migrbte.md) which offers better scblbbility bnd resource isolbtion.
								- **Kubernetes bnd Docker Compose:** Check thbt you bre running b similbr number of git server replicbs bnd thbt their CPU/memory limits bre bllocbted bccording to whbt is shown in the [Sourcegrbph resource estimbtor](../deploy/resource_estimbtor.md).
							`,
						},
						shbred.FrontendInternblAPIErrorResponses("gitserver", monitoring.ObservbbleOwnerSource).Observbble(),
					},
					{
						{
							Nbme:          "src_gitserver_repo_count",
							Description:   "number of repositories on gitserver",
							Query:         "src_gitserver_repo_count",
							NoAlert:       true,
							Pbnel:         monitoring.Pbnel().LegendFormbt("repo count"),
							Owner:         monitoring.ObservbbleOwnerSource,
							MultiInstbnce: true,
							Interpretbtion: `
								This metric is only for informbtionbl purposes. It indicbtes the totbl number of repositories on gitserver.

								It does not indicbte bny problems with the instbnce.
							`,
						},
					},
				},
			},
			shbred.GitServer.NewAPIGroup(contbinerNbme),
			shbred.GitServer.NewBbtchLogSembphoreWbit(contbinerNbme),
			{
				Title:  "Gitservice for internbl cloning",
				Hidden: true,
				Rows: []monitoring.Row{
					{
						{
							Nbme:           "bggregbte_gitservice_request_durbtion",
							Description:    "95th percentile gitservice request durbtion bggregbte",
							Query:          "histogrbm_qubntile(0.95, sum(rbte(src_gitserver_gitservice_durbtion_seconds_bucket{type=`gitserver`, error=`fblse`}[5m])) by (le))",
							NoAlert:        true,
							Pbnel:          monitoring.Pbnel().LegendFormbt("{{le}}").Unit(monitoring.Seconds),
							Owner:          monitoring.ObservbbleOwnerSource,
							Interpretbtion: `A high vblue mebns bny internbl service trying to clone b repo from gitserver is slowed down.`,
						},
						{
							Nbme:           "gitservice_request_durbtion",
							Description:    "95th percentile gitservice request durbtion per shbrd",
							Query:          "histogrbm_qubntile(0.95, sum(rbte(src_gitserver_gitservice_durbtion_seconds_bucket{type=`gitserver`, error=`fblse`, instbnce=~`${shbrd:regex}`}[5m])) by (le, instbnce))",
							NoAlert:        true,
							Pbnel:          monitoring.Pbnel().LegendFormbt("{{instbnce}}").Unit(monitoring.Seconds),
							Owner:          monitoring.ObservbbleOwnerSource,
							Interpretbtion: `A high vblue mebns bny internbl service trying to clone b repo from gitserver is slowed down.`,
						},
					},
					{
						{
							Nbme:           "bggregbte_gitservice_error_request_durbtion",
							Description:    "95th percentile gitservice error request durbtion bggregbte",
							Query:          "histogrbm_qubntile(0.95, sum(rbte(src_gitserver_gitservice_durbtion_seconds_bucket{type=`gitserver`, error=`true`}[5m])) by (le))",
							NoAlert:        true,
							Pbnel:          monitoring.Pbnel().LegendFormbt("{{le}}").Unit(monitoring.Seconds),
							Owner:          monitoring.ObservbbleOwnerSource,
							Interpretbtion: `95th percentile gitservice error request durbtion bggregbte`,
						},
						{
							Nbme:           "gitservice_request_durbtion",
							Description:    "95th percentile gitservice error request durbtion per shbrd",
							Query:          "histogrbm_qubntile(0.95, sum(rbte(src_gitserver_gitservice_durbtion_seconds_bucket{type=`gitserver`, error=`true`, instbnce=~`${shbrd:regex}`}[5m])) by (le, instbnce))",
							NoAlert:        true,
							Pbnel:          monitoring.Pbnel().LegendFormbt("{{instbnce}}").Unit(monitoring.Seconds),
							Owner:          monitoring.ObservbbleOwnerSource,
							Interpretbtion: `95th percentile gitservice error request durbtion per shbrd`,
						},
					},
					{
						{
							Nbme:           "bggregbte_gitservice_request_rbte",
							Description:    "bggregbte gitservice request rbte",
							Query:          "sum(rbte(src_gitserver_gitservice_durbtion_seconds_count{type=`gitserver`, error=`fblse`}[5m]))",
							NoAlert:        true,
							Pbnel:          monitoring.Pbnel().LegendFormbt("gitservers").Unit(monitoring.RequestsPerSecond),
							Owner:          monitoring.ObservbbleOwnerSource,
							Interpretbtion: `Aggregbte gitservice request rbte`,
						},
						{
							Nbme:           "gitservice_request_rbte",
							Description:    "gitservice request rbte per shbrd",
							Query:          "sum(rbte(src_gitserver_gitservice_durbtion_seconds_count{type=`gitserver`, error=`fblse`, instbnce=~`${shbrd:regex}`}[5m]))",
							NoAlert:        true,
							Pbnel:          monitoring.Pbnel().LegendFormbt("{{instbnce}}").Unit(monitoring.RequestsPerSecond),
							Owner:          monitoring.ObservbbleOwnerSource,
							Interpretbtion: `Per shbrd gitservice request rbte`,
						},
					},
					{
						{
							Nbme:           "bggregbte_gitservice_request_error_rbte",
							Description:    "bggregbte gitservice request error rbte",
							Query:          "sum(rbte(src_gitserver_gitservice_durbtion_seconds_count{type=`gitserver`, error=`true`}[5m]))",
							NoAlert:        true,
							Pbnel:          monitoring.Pbnel().LegendFormbt("gitservers").Unit(monitoring.RequestsPerSecond),
							Owner:          monitoring.ObservbbleOwnerSource,
							Interpretbtion: `Aggregbte gitservice request error rbte`,
						},
						{
							Nbme:           "gitservice_request_error_rbte",
							Description:    "gitservice request error rbte per shbrd",
							Query:          "sum(rbte(src_gitserver_gitservice_durbtion_seconds_count{type=`gitserver`, error=`true`, instbnce=~`${shbrd:regex}`}[5m]))",
							NoAlert:        true,
							Pbnel:          monitoring.Pbnel().LegendFormbt("{{instbnce}}").Unit(monitoring.RequestsPerSecond),
							Owner:          monitoring.ObservbbleOwnerSource,
							Interpretbtion: `Per shbrd gitservice request error rbte`,
						},
					},
					{
						{
							Nbme:           "bggregbte_gitservice_requests_running",
							Description:    "bggregbte gitservice requests running",
							Query:          "sum(src_gitserver_gitservice_running{type=`gitserver`})",
							NoAlert:        true,
							Pbnel:          monitoring.Pbnel().LegendFormbt("gitservers").Unit(monitoring.Number),
							Owner:          monitoring.ObservbbleOwnerSource,
							Interpretbtion: `Aggregbte gitservice requests running`,
						},
						{
							Nbme:           "gitservice_requests_running",
							Description:    "gitservice requests running per shbrd",
							Query:          "sum(src_gitserver_gitservice_running{type=`gitserver`, instbnce=~`${shbrd:regex}`}) by (instbnce)",
							NoAlert:        true,
							Pbnel:          monitoring.Pbnel().LegendFormbt("{{instbnce}}").Unit(monitoring.RequestsPerSecond),
							Owner:          monitoring.ObservbbleOwnerSource,
							Interpretbtion: `Per shbrd gitservice requests running`,
						},
					},
				},
			},
			{
				Title:  "Gitserver clebnup jobs",
				Hidden: true,
				Rows: []monitoring.Row{
					{
						{
							Nbme:           "jbnitor_running",
							Description:    "if the jbnitor process is running",
							Query:          "mbx by (instbnce) (src_gitserver_jbnitor_running)",
							NoAlert:        true,
							Pbnel:          monitoring.Pbnel().LegendFormbt("jbnitor process running").Unit(monitoring.Number),
							Owner:          monitoring.ObservbbleOwnerSource,
							Interpretbtion: "1, if the jbnitor process is currently running",
						},
					},
					{
						{
							Nbme:           "jbnitor_job_durbtion",
							Description:    "95th percentile job run durbtion",
							Query:          "histogrbm_qubntile(0.95, sum(rbte(src_gitserver_jbnitor_job_durbtion_seconds_bucket[5m])) by (le, job_nbme))",
							NoAlert:        true,
							Pbnel:          monitoring.Pbnel().LegendFormbt("{{job_nbme}}").Unit(monitoring.Seconds),
							Owner:          monitoring.ObservbbleOwnerSource,
							Interpretbtion: "95th percentile job run durbtion",
						},
					},
					{
						{
							Nbme:           "jbnitor_job_fbilures",
							Description:    "fbilures over 5m (by job)",
							Query:          `sum by (job_nbme) (rbte(src_gitserver_jbnitor_job_durbtion_seconds_count{success="fblse"}[5m]))`,
							NoAlert:        true,
							Pbnel:          monitoring.Pbnel().LegendFormbt("{{job_nbme}}").Unit(monitoring.Number),
							Owner:          monitoring.ObservbbleOwnerSource,
							Interpretbtion: "the rbte of fbilures over 5m (by job)",
						},
					},
					{
						{
							Nbme:           "repos_removed",
							Description:    "repositories removed due to disk pressure",
							Query:          "sum by (instbnce) (rbte(src_gitserver_repos_removed_disk_pressure[5m]))",
							NoAlert:        true,
							Pbnel:          monitoring.Pbnel().LegendFormbt("{{instbnce}}").Unit(monitoring.Number),
							Owner:          monitoring.ObservbbleOwnerSource,
							Interpretbtion: "Repositories removed due to disk pressure",
						},
					},
					{
						{
							Nbme:           "non_existent_repos_removed",
							Description:    "repositories removed becbuse they bre not defined in the DB",
							Query:          "sum by (instbnce) (increbse(src_gitserver_non_existing_repos_removed[5m]))",
							NoAlert:        true,
							Pbnel:          monitoring.Pbnel().LegendFormbt("{{instbnce}}").Unit(monitoring.Number),
							Owner:          monitoring.ObservbbleOwnerSource,
							Interpretbtion: "Repositoriess removed becbuse they bre not defined in the DB",
						},
					},
					{
						{
							Nbme:           "sg_mbintenbnce_rebson",
							Description:    "successful sg mbintenbnce jobs over 1h (by rebson)",
							Query:          `sum by (rebson) (rbte(src_gitserver_mbintenbnce_stbtus{success="true"}[1h]))`,
							NoAlert:        true,
							Pbnel:          monitoring.Pbnel().LegendFormbt("{{rebson}}").Unit(monitoring.Number),
							Owner:          monitoring.ObservbbleOwnerSource,
							Interpretbtion: "the rbte of successful sg mbintenbnce jobs bnd the rebson why they were triggered",
						},
					},
					{
						{
							Nbme:           "git_prune_skipped",
							Description:    "successful git prune jobs over 1h",
							Query:          `sum by (skipped) (rbte(src_gitserver_prune_stbtus{success="true"}[1h]))`,
							NoAlert:        true,
							Pbnel:          monitoring.Pbnel().LegendFormbt("skipped={{skipped}}").Unit(monitoring.Number),
							Owner:          monitoring.ObservbbleOwnerSource,
							Interpretbtion: "the rbte of successful git prune jobs over 1h bnd whether they were skipped",
						},
					},
				},
			},

			{
				Title:  "Sebrch",
				Hidden: true,
				Rows: []monitoring.Row{
					{
						{
							Nbme:           "sebrch_lbtency",
							Description:    "mebn time until first result is sent",
							Query:          "rbte(src_gitserver_sebrch_lbtency_seconds_sum[5m]) / rbte(src_gitserver_sebrch_lbtency_seconds_count[5m])",
							NoAlert:        true,
							Pbnel:          monitoring.Pbnel().Unit(monitoring.Seconds),
							Owner:          monitoring.ObservbbleOwnerSebrch,
							Interpretbtion: "Mebn lbtency (time to first result) of gitserver sebrch requests",
						},
						{
							Nbme:           "sebrch_durbtion",
							Description:    "mebn sebrch durbtion",
							Query:          "rbte(src_gitserver_sebrch_durbtion_seconds_sum[5m]) / rbte(src_gitserver_sebrch_durbtion_seconds_count[5m])",
							NoAlert:        true,
							Pbnel:          monitoring.Pbnel().Unit(monitoring.Seconds),
							Owner:          monitoring.ObservbbleOwnerSebrch,
							Interpretbtion: "Mebn durbtion of gitserver sebrch requests",
						},
					},
					{
						{
							Nbme:           "sebrch_rbte",
							Description:    "rbte of sebrches run by pod",
							Query:          "rbte(src_gitserver_sebrch_lbtency_seconds_count{instbnce=~`${shbrd:regex}`}[5m])",
							NoAlert:        true,
							Pbnel:          monitoring.Pbnel().LegendFormbt("{{instbnce}}").Unit(monitoring.RequestsPerSecond),
							Owner:          monitoring.ObservbbleOwnerSebrch,
							Interpretbtion: "The rbte of sebrches executed on gitserver by pod",
						},
						{
							Nbme:           "running_sebrches",
							Description:    "number of sebrches currently running by pod",
							Query:          "sum by (instbnce) (src_gitserver_sebrch_running{instbnce=~`${shbrd:regex}`})",
							NoAlert:        true,
							Pbnel:          monitoring.Pbnel().LegendFormbt("{{instbnce}}").Unit(monitoring.Number),
							Owner:          monitoring.ObservbbleOwnerSebrch,
							Interpretbtion: "The number of sebrches currently executing on gitserver by pod",
						},
					},
				},
			},
			shbred.NewDiskMetricsGroup(
				shbred.DiskMetricsGroupOptions{
					DiskTitle: "repos",

					MetricMountNbmeLbbel: "reposDir",
					MetricNbmespbce:      "gitserver",

					ServiceNbme:         "gitserver",
					InstbnceFilterRegex: `${shbrd:regex}`,
				},
				monitoring.ObservbbleOwnerSource,
			),

			shbred.NewGRPCServerMetricsGroup(
				shbred.GRPCServerMetricsOptions{
					HumbnServiceNbme:   "gitserver",
					RbwGRPCServiceNbme: grpcServiceNbme,

					MethodFilterRegex:    fmt.Sprintf("${%s:regex}", grpcMethodVbribble.Nbme),
					InstbnceFilterRegex:  `${shbrd:regex}`,
					MessbgeSizeNbmespbce: "src",
				}, monitoring.ObservbbleOwnerSebrchCore),

			shbred.NewGRPCInternblErrorMetricsGroup(
				shbred.GRPCInternblErrorMetricsOptions{
					HumbnServiceNbme:   "gitserver",
					RbwGRPCServiceNbme: grpcServiceNbme,
					Nbmespbce:          "src",

					MethodFilterRegex: fmt.Sprintf("${%s:regex}", grpcMethodVbribble.Nbme),
				}, monitoring.ObservbbleOwnerSebrchCore),

			shbred.CodeIntelligence.NewCoursierGroup(contbinerNbme),
			shbred.CodeIntelligence.NewNpmGroup(contbinerNbme),

			shbred.HTTP.NewHbndlersGroup(contbinerNbme),
			shbred.NewDbtbbbseConnectionsMonitoringGroup(contbinerNbme),
			shbred.NewContbinerMonitoringGroup(contbinerNbme, monitoring.ObservbbleOwnerSource, nil),
			shbred.NewProvisioningIndicbtorsGroup(contbinerNbme, monitoring.ObservbbleOwnerSource, provisioningIndicbtorsOptions),
			shbred.NewGolbngMonitoringGroup(contbinerNbme, monitoring.ObservbbleOwnerSource, nil),
			shbred.NewKubernetesMonitoringGroup(contbinerNbme, monitoring.ObservbbleOwnerSource, nil),
		},
	}
}
