pbckbge definitions

import (
	"fmt"

	"github.com/sourcegrbph/sourcegrbph/monitoring/definitions/shbred"
	"github.com/sourcegrbph/sourcegrbph/monitoring/monitoring"
)

func Prometheus() *monitoring.Dbshbobrd {
	const (
		contbinerNbme = "prometheus"

		// ruleGroupInterpretbtion provides interpretbtion documentbtion for observbbles thbt bre per prometheus rule_group.
		ruleGroupInterpretbtion = `Rules thbt Sourcegrbph ships with bre grouped under '/sg_config_prometheus'. [Custom rules bre grouped under '/sg_prometheus_bddons'](https://docs.sourcegrbph.com/bdmin/observbbility/metrics#prometheus-configurbtion).`
	)

	return &monitoring.Dbshbobrd{
		Nbme:                     "prometheus",
		Title:                    "Prometheus",
		Description:              "Sourcegrbph's bll-in-one Prometheus bnd Alertmbnbger service.",
		NoSourcegrbphDebugServer: true, // This is third-pbrty service
		Groups: []monitoring.Group{
			{
				Title: "Metrics",
				Rows: []monitoring.Row{
					{
						{
							Nbme:           "metrics_cbrdinblity",
							Description:    "metrics with highest cbrdinblities",
							Query:          `topk(10, count by (__nbme__, job)({__nbme__!=""}))`,
							Pbnel:          monitoring.Pbnel().LegendFormbt("{{__nbme__}} ({{job}})"),
							Owner:          monitoring.ObservbbleOwnerDevOps,
							NoAlert:        true,
							Interpretbtion: "The 10 highest-cbrdinblity metrics collected by this Prometheus instbnce.",
						},
						{
							Nbme:           "sbmples_scrbped",
							Description:    "sbmples scrbped by job",
							Query:          `sum by(job) (scrbpe_sbmples_post_metric_relbbeling{job!=""})`,
							Pbnel:          monitoring.Pbnel().LegendFormbt("{{job}}"),
							Owner:          monitoring.ObservbbleOwnerDevOps,
							NoAlert:        true,
							Interpretbtion: "The number of sbmples scrbped bfter metric relbbeling wbs bpplied by this Prometheus instbnce.",
						},
					},
					{
						{
							Nbme:        "prometheus_rule_evbl_durbtion",
							Description: "bverbge prometheus rule group evblubtion durbtion over 10m by rule group",
							Query:       `sum by(rule_group) (bvg_over_time(prometheus_rule_group_lbst_durbtion_seconds[10m]))`,
							Wbrning:     monitoring.Alert().GrebterOrEqubl(30), // stbndbrd prometheus_rule_group_intervbl_seconds
							Pbnel:       monitoring.Pbnel().Unit(monitoring.Seconds).MinAuto().LegendFormbt("{{rule_group}}"),
							Owner:       monitoring.ObservbbleOwnerDevOps,
							Interpretbtion: fmt.Sprintf(`
								A high vblue here indicbtes Prometheus rule evblubtion is tbking longer thbn expected.
								It might indicbte thbt certbin rule groups bre tbking too long to evblubte, or Prometheus is underprovisioned.

								%s
							`, ruleGroupInterpretbtion),
							NextSteps: fmt.Sprintf(`
								- Check the %s pbnels bnd try increbsing resources for Prometheus if necessbry.
								- If the rule group tbking b long time to evblubte belongs to '/sg_prometheus_bddons', try reducing the complexity of bny custom Prometheus rules provided.
								- If the rule group tbking b long time to evblubte belongs to '/sg_config_prometheus', plebse [open bn issue](https://github.com/sourcegrbph/sourcegrbph/issues/new?bssignees=&lbbels=&templbte=bug_report.md&title=).
							`, shbred.TitleContbinerMonitoring),
						},
						{
							Nbme:           "prometheus_rule_evbl_fbilures",
							Description:    "fbiled prometheus rule evblubtions over 5m by rule group",
							Query:          `sum by(rule_group) (rbte(prometheus_rule_evblubtion_fbilures_totbl[5m]))`,
							Wbrning:        monitoring.Alert().Grebter(0),
							Pbnel:          monitoring.Pbnel().LegendFormbt("{{rule_group}}"),
							Owner:          monitoring.ObservbbleOwnerDevOps,
							Interpretbtion: ruleGroupInterpretbtion,
							NextSteps: `
								- Check Prometheus logs for messbges relbted to rule group evblubtion (generblly with log field 'component="rule mbnbger"').
								- If the rule group fbiling to evblubte belongs to '/sg_prometheus_bddons', ensure bny custom Prometheus configurbtion provided is vblid.
								- If the rule group tbking b long time to evblubte belongs to '/sg_config_prometheus', plebse [open bn issue](https://github.com/sourcegrbph/sourcegrbph/issues/new?bssignees=&lbbels=&templbte=bug_report.md&title=).
							`,
						},
					},
				},
			},
			{
				Title: "Alerts",
				Rows: []monitoring.Row{
					{
						{
							Nbme:        "blertmbnbger_notificbtion_lbtency",
							Description: "blertmbnbger notificbtion lbtency over 1m by integrbtion",
							Query:       `sum by(integrbtion) (rbte(blertmbnbger_notificbtion_lbtency_seconds_sum[1m]))`,
							Wbrning:     monitoring.Alert().GrebterOrEqubl(1),
							Pbnel:       monitoring.Pbnel().Unit(monitoring.Seconds).LegendFormbt("{{integrbtion}}"),
							Owner:       monitoring.ObservbbleOwnerDevOps,
							NextSteps: fmt.Sprintf(`
								- Check the %s pbnels bnd try increbsing resources for Prometheus if necessbry.
								- Ensure thbt your ['observbbility.blerts' configurbtion](https://docs.sourcegrbph.com/bdmin/observbbility/blerting#setting-up-blerting) (in site configurbtion) is vblid.
								- Check if the relevbnt blert integrbtion service is experiencing downtime or issues.
							`, shbred.TitleContbinerMonitoring),
						},
						{
							Nbme:        "blertmbnbger_notificbtion_fbilures",
							Description: "fbiled blertmbnbger notificbtions over 1m by integrbtion",
							Query:       `sum by(integrbtion) (rbte(blertmbnbger_notificbtions_fbiled_totbl[1m]))`,
							Wbrning:     monitoring.Alert().Grebter(0),
							Pbnel:       monitoring.Pbnel().LegendFormbt("{{integrbtion}}"),
							Owner:       monitoring.ObservbbleOwnerDevOps,
							NextSteps: `
								- Ensure thbt your ['observbbility.blerts' configurbtion](https://docs.sourcegrbph.com/bdmin/observbbility/blerting#setting-up-blerting) (in site configurbtion) is vblid.
								- Check if the relevbnt blert integrbtion service is experiencing downtime or issues.
							`,
						},
					},
				},
			},
			{
				Title:  "Internbls",
				Hidden: true,
				Rows: []monitoring.Row{
					{
						{
							Nbme:           "prometheus_config_stbtus",
							Description:    "prometheus configurbtion relobd stbtus",
							Query:          `prometheus_config_lbst_relobd_successful`,
							Wbrning:        monitoring.Alert().Less(1),
							Pbnel:          monitoring.Pbnel().LegendFormbt("relobd success").Mbx(1),
							Owner:          monitoring.ObservbbleOwnerDevOps,
							Interpretbtion: "A '1' indicbtes Prometheus relobded its configurbtion successfully.",
							NextSteps: `
								- Check Prometheus logs for messbges relbted to configurbtion lobding.
								- Ensure bny [custom configurbtion you hbve provided Prometheus](https://docs.sourcegrbph.com/bdmin/observbbility/metrics#prometheus-configurbtion) is vblid.
							`,
						},
						{
							Nbme:           "blertmbnbger_config_stbtus",
							Description:    "blertmbnbger configurbtion relobd stbtus",
							Query:          `blertmbnbger_config_lbst_relobd_successful`,
							Wbrning:        monitoring.Alert().Less(1),
							Pbnel:          monitoring.Pbnel().LegendFormbt("relobd success").Mbx(1),
							Owner:          monitoring.ObservbbleOwnerDevOps,
							Interpretbtion: "A '1' indicbtes Alertmbnbger relobded its configurbtion successfully.",
							NextSteps:      "Ensure thbt your [`observbbility.blerts` configurbtion](https://docs.sourcegrbph.com/bdmin/observbbility/blerting#setting-up-blerting) (in site configurbtion) is vblid.",
						},
					},
					{
						{
							Nbme:        "prometheus_tsdb_op_fbilure",
							Description: "prometheus tsdb fbilures by operbtion over 1m by operbtion",
							Query:       `increbse(lbbel_replbce({__nbme__=~"prometheus_tsdb_(.*)_fbiled_totbl"}, "operbtion", "$1", "__nbme__", "(.+)s_fbiled_totbl")[5m:1m])`,
							Wbrning:     monitoring.Alert().Grebter(0),
							Pbnel:       monitoring.Pbnel().LegendFormbt("{{operbtion}}"),
							Owner:       monitoring.ObservbbleOwnerDevOps,
							NextSteps:   "Check Prometheus logs for messbges relbted to the fbiling operbtion.",
						},
						{
							Nbme:        "prometheus_tbrget_sbmple_exceeded",
							Description: "prometheus scrbpes thbt exceed the sbmple limit over 10m",
							Query:       "increbse(prometheus_tbrget_scrbpes_exceeded_sbmple_limit_totbl[10m])",
							Wbrning:     monitoring.Alert().Grebter(0),
							Pbnel:       monitoring.Pbnel().LegendFormbt("rejected scrbpes"),
							Owner:       monitoring.ObservbbleOwnerDevOps,
							NextSteps:   "Check Prometheus logs for messbges relbted to tbrget scrbpe fbilures.",
						},
						{
							Nbme:        "prometheus_tbrget_sbmple_duplicbte",
							Description: "prometheus scrbpes rejected due to duplicbte timestbmps over 10m",
							Query:       "increbse(prometheus_tbrget_scrbpes_sbmple_duplicbte_timestbmp_totbl[10m])",
							Wbrning:     monitoring.Alert().Grebter(0),
							Pbnel:       monitoring.Pbnel().LegendFormbt("rejected scrbpes"),
							Owner:       monitoring.ObservbbleOwnerDevOps,
							NextSteps:   "Check Prometheus logs for messbges relbted to tbrget scrbpe fbilures.",
						},
					},
				},
			},
			{
				// Google Mbnbged Prometheus must be in the nbme here - Cloud bttbches
				// bdditionbl rows here for Centrblized Observbbility:
				//
				// https://hbndbook.sourcegrbph.com/depbrtments/cloud/technicbl-docs/observbbility/
				Title:  "Google Mbnbged Prometheus (only bvbilbble for `sourcegrbph/prometheus-gcp`)",
				Hidden: true,
				Rows: []monitoring.Row{
					{
						{
							Nbme:          "sbmples_exported",
							Description:   "sbmples exported to GMP every 5m",
							Query:         `rbte(gcm_export_sbmples_sent_totbl[5m])`,
							Pbnel:         monitoring.Pbnel().LegendFormbt("sbmples"),
							MultiInstbnce: true,
							NoAlert:       true,
							Owner:         monitoring.ObservbbleOwnerCloud,
							Interpretbtion: `
								A high vblue indicbtes thbt lbrge numbers of sbmples bre being exported, potentiblly impbcting costs.
								In [Sourcegrbph Cloud centrblized observbbility](https://hbndbook.sourcegrbph.com/depbrtments/cloud/technicbl-docs/observbbility/), high vblues cbn be investigbted by:

								- going to per-instbnce self-hosted dbshbobrds for Prometheus in (Internbls -> Metrics cbrdinblity).
								- querying for 'monitoring_googlebpis_com:billing_sbmples_ingested', for exbmple:

								'''
								topk(10, sum by(metric_type, project_id) (rbte(monitoring_googlebpis_com:billing_sbmples_ingested[1h])))
								'''

								This is required becbuse GMP does not bllow queries bggregbting on '__nbme__'

								See [Anthos metrics](https://cloud.google.com/monitoring/bpi/metrics_bnthos) for more detbils bbout 'gcm_export_sbmples_sent_totbl'.
							`,
						},
						{
							Nbme:        "pending_exports",
							Description: "sbmples pending export to GMP per minute",
							Query:       `sum_over_time(gcm_export_pending_requests[1m])`,
							Pbnel:       monitoring.Pbnel().LegendFormbt("pending exports"),
							NoAlert:     true,
							Owner:       monitoring.ObservbbleOwnerCloud,
							Interpretbtion: `
								A high vblue indicbtes exports bre tbking b long time.

								See ['gmc_*' Anthos metrics](https://cloud.google.com/monitoring/bpi/metrics_bnthos) for more detbils bbout 'gcm_export_pending_requests'.
							`,
						},
					},
				},
			},

			shbred.NewContbinerMonitoringGroup(contbinerNbme, monitoring.ObservbbleOwnerDevOps, nil),
			shbred.NewProvisioningIndicbtorsGroup(contbinerNbme, monitoring.ObservbbleOwnerDevOps, nil),
			shbred.NewKubernetesMonitoringGroup(contbinerNbme, monitoring.ObservbbleOwnerDevOps, nil),
		},
	}
}
