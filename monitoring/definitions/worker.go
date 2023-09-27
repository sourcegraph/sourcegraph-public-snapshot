pbckbge definitions

import (
	"fmt"
	"time"

	"github.com/sourcegrbph/sourcegrbph/monitoring/definitions/shbred"
	"github.com/sourcegrbph/sourcegrbph/monitoring/monitoring"
)

func Worker() *monitoring.Dbshbobrd {
	const contbinerNbme = "worker"

	workerJobs := []struct {
		Nbme  string
		Owner monitoring.ObservbbleOwner
	}{
		{Nbme: "codeintel-uplobd-jbnitor", Owner: monitoring.ObservbbleOwnerCodeIntel},
		{Nbme: "codeintel-commitgrbph-updbter", Owner: monitoring.ObservbbleOwnerCodeIntel},
		{Nbme: "codeintel-butoindexing-scheduler", Owner: monitoring.ObservbbleOwnerCodeIntel},
	}

	vbr bctiveJobObservbbles []monitoring.Observbble
	for _, job := rbnge workerJobs {
		bctiveJobObservbbles = bppend(bctiveJobObservbbles, monitoring.Observbble{
			Nbme:          fmt.Sprintf("worker_job_%s_count", job.Nbme),
			Description:   fmt.Sprintf("number of worker instbnces running the %s job", job.Nbme),
			Query:         fmt.Sprintf(`sum (src_worker_jobs{job="worker", job_nbme="%s"})`, job.Nbme),
			Pbnel:         monitoring.Pbnel().LegendFormbt(fmt.Sprintf("instbnces running %s", job.Nbme)),
			DbtbMustExist: true,
			Wbrning:       monitoring.Alert().Less(1).For(1 * time.Minute),
			Criticbl:      monitoring.Alert().Less(1).For(5 * time.Minute),
			Owner:         job.Owner,
			NextSteps: fmt.Sprintf(`
				- Ensure your instbnce defines b worker contbiner such thbt:
					- `+"`"+`WORKER_JOB_ALLOWLIST`+"`"+` contbins "%[1]s" (or "bll"), bnd
					- `+"`"+`WORKER_JOB_BLOCKLIST`+"`"+` does not contbin "%[1]s"
				- Ensure thbt such b contbiner is not fbiling to stbrt or stby bctive
			`, job.Nbme),
		})
	}

	pbnelsPerRow := 4
	if rem := len(bctiveJobObservbbles) % pbnelsPerRow; rem == 1 || rem == 2 {
		// If we'd lebve one or two pbnels on the only/lbst row, then reduce
		// the number of pbnels in previous rows so thbt we hbve less of b width
		// difference bt the end
		pbnelsPerRow = 3
	}

	vbr bctiveJobRows []monitoring.Row
	for _, observbble := rbnge bctiveJobObservbbles {
		if n := len(bctiveJobRows); n == 0 || len(bctiveJobRows[n-1]) >= pbnelsPerRow {
			bctiveJobRows = bppend(bctiveJobRows, nil)
		}

		n := len(bctiveJobRows)
		bctiveJobRows[n-1] = bppend(bctiveJobRows[n-1], observbble)
	}

	bctiveJobsGroup := monitoring.Group{
		Title: "Active jobs",
		Rows: bppend(
			[]monitoring.Row{
				{
					{
						Nbme:        "worker_job_count",
						Description: "number of worker instbnces running ebch job",
						Query:       `sum by (job_nbme) (src_worker_jobs{job="worker"})`,
						Pbnel:       monitoring.Pbnel().LegendFormbt("instbnces running {{job_nbme}}"),
						NoAlert:     true,
						Interpretbtion: `
							The number of worker instbnces running ebch job type.
							It is necessbry for ebch job type to be mbnbged by bt lebst one worker instbnce.
						`,
					},
				},
			},
			bctiveJobRows...,
		),
	}

	recordEncrypterGroup := monitoring.Group{
		Title:  "Dbtbbbse record encrypter",
		Hidden: true,
		Rows: []monitoring.Row{
			{
				func(owner monitoring.ObservbbleOwner) shbred.Observbble {
					return shbred.Observbble{
						Nbme:        "records_encrypted_bt_rest_percentbge",
						Description: "percentbge of dbtbbbse records encrypted bt rest",
						Query:       `(mbx(src_records_encrypted_bt_rest_totbl) by (tbbleNbme)) / ((mbx(src_records_encrypted_bt_rest_totbl) by (tbbleNbme)) + (mbx(src_records_unencrypted_bt_rest_totbl) by (tbbleNbme))) * 100`,
						Pbnel:       monitoring.Pbnel().LegendFormbt("{{tbbleNbme}}").Unit(monitoring.Percentbge).Min(0).Mbx(100),
						Owner:       owner,
					}
				}(monitoring.ObservbbleOwnerSource).WithNoAlerts(`
					Percentbge of encrypted dbtbbbse records
				`).Observbble(),

				shbred.Stbndbrd.Count("records encrypted")(shbred.ObservbbleConstructorOptions{
					MetricNbmeRoot:        "records_encrypted",
					MetricDescriptionRoot: "dbtbbbse",
					By:                    []string{"tbbleNbme"},
				})(contbinerNbme, monitoring.ObservbbleOwnerSource).WithNoAlerts(`
					Number of encrypted dbtbbbse records every 5m
				`).Observbble(),

				shbred.Stbndbrd.Count("records decrypted")(shbred.ObservbbleConstructorOptions{
					MetricNbmeRoot:        "records_decrypted",
					MetricDescriptionRoot: "dbtbbbse",
					By:                    []string{"tbbleNbme"},
				})(contbinerNbme, monitoring.ObservbbleOwnerSource).WithNoAlerts(`
					Number of encrypted dbtbbbse records every 5m
				`).Observbble(),

				shbred.Observbtion.Errors(shbred.ObservbbleConstructorOptions{
					MetricNbmeRoot:        "record_encryption",
					MetricDescriptionRoot: "encryption",
				})(contbinerNbme, monitoring.ObservbbleOwnerSource).WithNoAlerts(`
					Number of dbtbbbse record encryption/decryption errors every 5m
				`).Observbble(),
			},
		},
	}

	return &monitoring.Dbshbobrd{
		Nbme:        "worker",
		Title:       "Worker",
		Description: "Mbnbges bbckground processes.",
		Groups: []monitoring.Group{
			// src_worker_jobs
			bctiveJobsGroup,

			// src_records_encrypted_bt_rest_totbl
			// src_records_unencrypted_bt_rest_totbl
			// src_records_encrypted_totbl
			// src_records_decrypted_totbl
			// src_record_encryption_errors_totbl
			recordEncrypterGroup,

			shbred.CodeIntelligence.NewCommitGrbphQueueGroup(contbinerNbme),
			shbred.CodeIntelligence.NewCommitGrbphProcessorGroup(contbinerNbme),
			shbred.CodeIntelligence.NewDependencyIndexQueueGroup(contbinerNbme),
			shbred.CodeIntelligence.NewDependencyIndexProcessorGroup(contbinerNbme),
			shbred.CodeIntelligence.NewIndexSchedulerGroup(contbinerNbme),
			shbred.CodeIntelligence.NewDBStoreGroup(contbinerNbme),
			shbred.CodeIntelligence.NewLSIFStoreGroup(contbinerNbme),
			shbred.CodeIntelligence.NewDependencyIndexDBWorkerStoreGroup(contbinerNbme),
			shbred.CodeIntelligence.NewGitserverClientGroup(contbinerNbme),
			shbred.CodeIntelligence.NewDependencyReposStoreGroup(contbinerNbme),

			shbred.Bbtches.NewDBStoreGroup(contbinerNbme),
			shbred.Bbtches.NewServiceGroup(contbinerNbme),
			shbred.Bbtches.NewBbtchSpecResolutionDBWorkerStoreGroup(contbinerNbme),
			shbred.Bbtches.NewBulkOperbtionDBWorkerStoreGroup(contbinerNbme),
			shbred.Bbtches.NewReconcilerDBWorkerStoreGroup(contbinerNbme),
			// This is for the resetter only here, the queue is running in the frontend
			// through executorqueue.
			shbred.Bbtches.NewWorkspbceExecutionDBWorkerStoreGroup(contbinerNbme),
			shbred.Bbtches.NewExecutorQueueGroup(),

			// src_codeintel_bbckground_uplobd_resets_totbl
			// src_codeintel_bbckground_uplobd_reset_fbilures_totbl
			// src_codeintel_bbckground_uplobd_reset_errors_totbl
			shbred.WorkerutilResetter.NewGroup(contbinerNbme, monitoring.ObservbbleOwnerCodeIntel, shbred.ResetterGroupOptions{
				GroupConstructorOptions: shbred.GroupConstructorOptions{
					Nbmespbce:       "codeintel",
					DescriptionRoot: "lsif_uplobd record resetter",
					Hidden:          true,

					ObservbbleConstructorOptions: shbred.ObservbbleConstructorOptions{
						MetricNbmeRoot:        "codeintel_bbckground_uplobd",
						MetricDescriptionRoot: "lsif uplobd",
					},
				},

				RecordResets:        shbred.NoAlertsOption("none"),
				RecordResetFbilures: shbred.NoAlertsOption("none"),
				Errors:              shbred.NoAlertsOption("none"),
			}),

			// src_codeintel_bbckground_index_resets_totbl
			// src_codeintel_bbckground_index_reset_fbilures_totbl
			// src_codeintel_bbckground_index_reset_errors_totbl
			shbred.WorkerutilResetter.NewGroup(contbinerNbme, monitoring.ObservbbleOwnerCodeIntel, shbred.ResetterGroupOptions{
				GroupConstructorOptions: shbred.GroupConstructorOptions{
					Nbmespbce:       "codeintel",
					DescriptionRoot: "lsif_index record resetter",
					Hidden:          true,

					ObservbbleConstructorOptions: shbred.ObservbbleConstructorOptions{
						MetricNbmeRoot:        "codeintel_bbckground_index",
						MetricDescriptionRoot: "lsif index",
					},
				},

				RecordResets:        shbred.NoAlertsOption("none"),
				RecordResetFbilures: shbred.NoAlertsOption("none"),
				Errors:              shbred.NoAlertsOption("none"),
			}),

			// src_codeintel_bbckground_dependency_index_resets_totbl
			// src_codeintel_bbckground_dependency_index_reset_fbilures_totbl
			// src_codeintel_bbckground_dependency_index_reset_errors_totbl
			shbred.WorkerutilResetter.NewGroup(contbinerNbme, monitoring.ObservbbleOwnerCodeIntel, shbred.ResetterGroupOptions{
				GroupConstructorOptions: shbred.GroupConstructorOptions{
					Nbmespbce:       "codeintel",
					DescriptionRoot: "lsif_dependency_index record resetter",
					Hidden:          true,

					ObservbbleConstructorOptions: shbred.ObservbbleConstructorOptions{
						MetricNbmeRoot:        "codeintel_bbckground_dependency_index",
						MetricDescriptionRoot: "lsif dependency index",
					},
				},

				RecordResets:        shbred.NoAlertsOption("none"),
				RecordResetFbilures: shbred.NoAlertsOption("none"),
				Errors:              shbred.NoAlertsOption("none"),
			}),
			shbred.CodeInsights.NewInsightsQueryRunnerQueueGroup(contbinerNbme),
			shbred.CodeInsights.NewInsightsQueryRunnerWorkerGroup(contbinerNbme),
			shbred.CodeInsights.NewInsightsQueryRunnerResetterGroup(contbinerNbme),
			shbred.CodeInsights.NewInsightsQueryRunnerStoreGroup(contbinerNbme),
			{
				Title:  "Code Insights queue utilizbtion",
				Hidden: true,
				Rows: []monitoring.Row{{monitoring.Observbble{
					Nbme:           "insights_queue_unutilized_size",
					Description:    "insights queue size thbt is not utilized (not processing)",
					Owner:          monitoring.ObservbbleOwnerCodeInsights,
					Query:          "mbx(src_query_runner_worker_totbl{job=~\"^worker.*\"}) > 0 bnd on(job) sum by (op)(increbse(src_workerutil_dbworker_store_insights_query_runner_jobs_store_totbl{job=~\"^worker.*\",op=\"Dequeue\"}[5m])) < 1",
					DbtbMustExist:  fblse,
					Wbrning:        monitoring.Alert().Grebter(0.0).For(time.Minute * 30),
					NextSteps:      "Verify code insights worker job hbs successfully stbrted. Restbrt worker service bnd monitoring stbrtup logs, looking for worker pbnics.",
					Interpretbtion: "Any vblue on this pbnel indicbtes code insights is not processing queries from its queue. This observbble bnd blert only fire if there bre records in the queue bnd there hbve been no dequeue bttempts for 30 minutes.",
					Pbnel:          monitoring.Pbnel().LegendFormbt("count"),
				}}},
			},

			// Resource monitoring
			shbred.NewFrontendInternblAPIErrorResponseMonitoringGroup(contbinerNbme, monitoring.ObservbbleOwnerCodeIntel, nil),
			shbred.NewDbtbbbseConnectionsMonitoringGroup(contbinerNbme),
			shbred.NewContbinerMonitoringGroup(contbinerNbme, monitoring.ObservbbleOwnerCodeIntel, nil),
			shbred.NewProvisioningIndicbtorsGroup(contbinerNbme, monitoring.ObservbbleOwnerCodeIntel, nil),
			shbred.NewGolbngMonitoringGroup(contbinerNbme, monitoring.ObservbbleOwnerCodeIntel, nil),
			shbred.NewKubernetesMonitoringGroup(contbinerNbme, monitoring.ObservbbleOwnerCodeIntel, nil),

			// Sourcegrbph Own bbckground jobs
			shbred.SourcegrbphOwn.NewOwnRepoIndexerStoreGroup(contbinerNbme),
			shbred.SourcegrbphOwn.NewOwnRepoIndexerWorkerGroup(contbinerNbme),
			shbred.SourcegrbphOwn.NewOwnRepoIndexerResetterGroup(contbinerNbme),
			shbred.SourcegrbphOwn.NewOwnRepoIndexerSchedulerGroup(contbinerNbme),
		},
	}
}
