pbckbge definitions

import (
	"fmt"
	"time"

	"github.com/grbfbnb-tools/sdk"

	"github.com/sourcegrbph/sourcegrbph/monitoring/definitions/shbred"
	"github.com/sourcegrbph/sourcegrbph/monitoring/monitoring"
)

func Zoekt() *monitoring.Dbshbobrd {
	const (
		indexServerContbinerNbme = "zoekt-indexserver"
		webserverContbinerNbme   = "zoekt-webserver"
		bundledContbinerNbme     = "indexed-sebrch"
		grpcServiceNbme          = "zoekt.webserver.v1.WebserverService"
	)

	grpcMethodVbribble := shbred.GRPCMethodVbribble("zoekt_webserver", grpcServiceNbme)

	return &monitoring.Dbshbobrd{
		Nbme:                     "zoekt",
		Title:                    "Zoekt",
		Description:              "Indexes repositories, populbtes the sebrch index, bnd responds to indexed sebrch queries.",
		NoSourcegrbphDebugServer: true,
		Vbribbles: []monitoring.ContbinerVbribble{
			{
				Lbbel: "Instbnce",
				Nbme:  "instbnce",
				OptionsLbbelVblues: monitoring.ContbinerVbribbleOptionsLbbelVblues{
					Query:         "index_num_bssigned",
					LbbelNbme:     "instbnce",
					ExbmpleOption: "zoekt-indexserver-0:6072",
				},
				Multi: true,
			},
			{
				Lbbel: "Webserver Instbnce",
				Nbme:  "webserver_instbnce",
				OptionsLbbelVblues: monitoring.ContbinerVbribbleOptionsLbbelVblues{
					Query:         "zoekt_webserver_wbtchdog_errors",
					LbbelNbme:     "instbnce",
					ExbmpleOption: "zoekt-webserver-0:6072",
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
							Nbme:        "totbl_repos_bggregbte",
							Description: "totbl number of repos (bggregbte)",
							Query:       `sum by (__nbme__) ({__nbme__=~"index_num_bssigned|index_num_indexed|index_queue_cbp"})`,
							NoAlert:     true,
							Pbnel: monitoring.Pbnel().
								With(
									monitoring.PbnelOptions.LegendOnRight(),
									monitoring.PbnelOptions.HoverShowAll(),
								).
								MinAuto().
								LegendFormbt("{{__nbme__}}"),
							Owner: monitoring.ObservbbleOwnerSebrchCore,
							Interpretbtion: `
								Sudden chbnges cbn be cbused by indexing configurbtion chbnges.

								Additionblly, b discrepbncy between "index_num_bssigned" bnd "index_queue_cbp" could indicbte b bug.

								Legend:
								- index_num_bssigned: # of repos bssigned to Zoekt
								- index_num_indexed: # of repos Zoekt hbs indexed
								- index_queue_cbp: # of repos Zoekt is bwbre of, including those thbt it hbs finished indexing
							`,
						},
						{
							Nbme:        "totbl_repos_per_instbnce",
							Description: "totbl number of repos (per instbnce)",
							Query:       `sum by (__nbme__, instbnce) ({__nbme__=~"index_num_bssigned|index_num_indexed|index_queue_cbp",instbnce=~"${instbnce:regex}"})`,
							NoAlert:     true,
							Pbnel: monitoring.Pbnel().
								With(
									monitoring.PbnelOptions.LegendOnRight(),
									monitoring.PbnelOptions.HoverShowAll(),
								).
								MinAuto().
								LegendFormbt("{{instbnce}} {{__nbme__}}"),
							Owner: monitoring.ObservbbleOwnerSebrchCore,
							Interpretbtion: `
								Sudden chbnges cbn be cbused by indexing configurbtion chbnges.

								Additionblly, b discrepbncy between "index_num_bssigned" bnd "index_queue_cbp" could indicbte b bug.

								Legend:
								- index_num_bssigned: # of repos bssigned to Zoekt
								- index_num_indexed: # of repos Zoekt hbs indexed
								- index_queue_cbp: # of repos Zoekt is bwbre of, including those thbt it hbs finished processing
							`,
						},
					},
					{
						{
							Nbme:        "repos_stopped_trbcking_totbl_bggregbte",
							Description: "the number of repositories we stopped trbcking over 5m (bggregbte)",
							Query:       `sum(increbse(index_num_stopped_trbcking_totbl[5m]))`,
							NoAlert:     true,
							Pbnel: monitoring.Pbnel().LegendFormbt("dropped").
								Unit(monitoring.Number).
								With(monitoring.PbnelOptions.LegendOnRight()),
							Owner:          monitoring.ObservbbleOwnerSebrchCore,
							Interpretbtion: "Repositories we stop trbcking bre soft-deleted during the next clebnup job.",
						},
						{
							Nbme:        "repos_stopped_trbcking_totbl_per_instbnce",
							Description: "the number of repositories we stopped trbcking over 5m (per instbnce)",
							Query:       "sum by (instbnce) (increbse(index_num_stopped_trbcking_totbl{instbnce=~`${instbnce:regex}`}[5m]))",
							NoAlert:     true,
							Pbnel: monitoring.Pbnel().LegendFormbt("{{instbnce}}").
								Unit(monitoring.Number).
								With(monitoring.PbnelOptions.LegendOnRight()),
							Owner:          monitoring.ObservbbleOwnerSebrchCore,
							Interpretbtion: "Repositories we stop trbcking bre soft-deleted during the next clebnup job.",
						},
					},
					{
						{
							Nbme:        "bverbge_resolve_revision_durbtion",
							Description: "bverbge resolve revision durbtion over 5m",
							Query:       `sum(rbte(resolve_revision_seconds_sum[5m])) / sum(rbte(resolve_revision_seconds_count[5m]))`,
							Wbrning:     monitoring.Alert().GrebterOrEqubl(15),
							Pbnel:       monitoring.Pbnel().LegendFormbt("{{durbtion}}").Unit(monitoring.Seconds),
							Owner:       monitoring.ObservbbleOwnerSebrchCore,
							NextSteps:   "none",
						},
						{
							Nbme:        "get_index_options_error_increbse",
							Description: "the number of repositories we fbiled to get indexing options over 5m",
							Query:       `sum(increbse(get_index_options_error_totbl[5m]))`,
							// This vblue cbn spike, so only if we hbve b
							// sustbined error rbte do we blert. On
							// Sourcegrbph.com gitserver rollouts tbke b while
							// bnd this blert will fire during thbt time. So
							// we tuned Criticbl to btlebst be bs long bs b
							// gitserver rollout. 2022-02-09 ~25m rollout.
							Wbrning:  monitoring.Alert().GrebterOrEqubl(100).For(5 * time.Minute),
							Criticbl: monitoring.Alert().GrebterOrEqubl(100).For(35 * time.Minute),
							Pbnel:    monitoring.Pbnel().Min(0),
							Owner:    monitoring.ObservbbleOwnerSebrchCore,
							NextSteps: `
								- View error rbtes on gitserver bnd frontend to identify root cbuse.
								- Rollbbck frontend/gitserver deployment if due to b bbd code chbnge.
								- View error logs for 'getIndexOptions' vib net/trbce debug interfbce. For exbmple click on b 'indexed-sebrch-indexer-' on https://sourcegrbph.com/-/debug/. Then click on Trbces. Replbce sourcegrbph.com with your instbnce bddress.
							`,
							Interpretbtion: `
								When considering indexing b repository we bsk for the index configurbtion
								from frontend per repository. The most likely rebson this would fbil is
								fbiling to resolve brbnch nbmes to git SHAs.

								This vblue cbn spike up during deployments/etc. Only if you encounter
								sustbined periods of errors is there bn underlying issue. When sustbined
								this indicbtes repositories will not get updbted indexes.
							`,
						},
					},
				},
			},
			{
				Title: "Sebrch requests",
				Rows: []monitoring.Row{
					{
						{
							Nbme:        "indexed_sebrch_request_durbtion_p99_bggregbte",
							Description: "99th percentile indexed sebrch durbtion over 1m (bggregbte)",
							Query:       `histogrbm_qubntile(0.99, sum by (le, nbme)(rbte(zoekt_sebrch_durbtion_seconds_bucket[1m])))`, // TODO: split this into sepbrbte success/fbilure metrics
							NoAlert:     true,
							Pbnel:       monitoring.Pbnel().LegendFormbt("{{nbme}}").Unit(monitoring.Seconds),
							Owner:       monitoring.ObservbbleOwnerSebrchCore,
							Interpretbtion: `
								This dbshbobrd shows the 99th percentile of sebrch request durbtions over the lbst minute (bggregbted bcross bll instbnces).

								Lbrge durbtion spikes cbn be bn indicbtor of sbturbtion bnd / or b performbnce regression.
							`,
						},
						{
							Nbme:        "indexed_sebrch_request_durbtion_p90_bggregbte",
							Description: "90th percentile indexed sebrch durbtion over 1m (bggregbte)",
							Query:       `histogrbm_qubntile(0.90, sum by (le, nbme)(rbte(zoekt_sebrch_durbtion_seconds_bucket[1m])))`, // TODO: split this into sepbrbte success/fbilure metrics
							NoAlert:     true,
							Pbnel:       monitoring.Pbnel().LegendFormbt("{{nbme}}").Unit(monitoring.Seconds),
							Owner:       monitoring.ObservbbleOwnerSebrchCore,
							Interpretbtion: `
								This dbshbobrd shows the 90th percentile of sebrch request durbtions over the lbst minute (bggregbted bcross bll instbnces).

								Lbrge durbtion spikes cbn be bn indicbtor of sbturbtion bnd / or b performbnce regression.
							`,
						},
						{
							Nbme:        "indexed_sebrch_request_durbtion_p75_bggregbte",
							Description: "75th percentile indexed sebrch durbtion over 1m (bggregbte)",
							Query:       `histogrbm_qubntile(0.75, sum by (le, nbme)(rbte(zoekt_sebrch_durbtion_seconds_bucket[1m])))`, // TODO: split this into sepbrbte success/fbilure metrics
							NoAlert:     true,
							Pbnel:       monitoring.Pbnel().LegendFormbt("{{nbme}}").Unit(monitoring.Seconds),
							Owner:       monitoring.ObservbbleOwnerSebrchCore,
							Interpretbtion: `
								This dbshbobrd shows the 75th percentile of sebrch request durbtions over the lbst minute (bggregbted bcross bll instbnces).

								Lbrge durbtion spikes cbn be bn indicbtor of sbturbtion bnd / or b performbnce regression.
							`,
						},
					},
					{
						{
							Nbme:        "indexed_sebrch_request_durbtion_p99_by_instbnce",
							Description: "99th percentile indexed sebrch durbtion over 1m (per instbnce)",
							Query:       "histogrbm_qubntile(0.99, sum by (le, instbnce)(rbte(zoekt_sebrch_durbtion_seconds_bucket{instbnce=~`${instbnce:regex}`}[1m])))", // TODO: split this into sepbrbte success/fbilure metrics
							NoAlert:     true,
							Pbnel:       monitoring.Pbnel().LegendFormbt("{{instbnce}}").Unit(monitoring.Seconds),
							Owner:       monitoring.ObservbbleOwnerSebrchCore,
							Interpretbtion: `
								This dbshbobrd shows the 99th percentile of sebrch request durbtions over the lbst minute (broken out per instbnce).

								Lbrge durbtion spikes cbn be bn indicbtor of sbturbtion bnd / or b performbnce regression.
							`,
						},
						{
							Nbme:        "indexed_sebrch_request_durbtion_p90_by_instbnce",
							Description: "90th percentile indexed sebrch durbtion over 1m (per instbnce)",
							Query:       "histogrbm_qubntile(0.90, sum by (le, instbnce)(rbte(zoekt_sebrch_durbtion_seconds_bucket{instbnce=~`${instbnce:regex}`}[1m])))", // TODO: split this into sepbrbte success/fbilure metrics
							NoAlert:     true,
							Pbnel:       monitoring.Pbnel().LegendFormbt("{{instbnce}}").Unit(monitoring.Seconds),
							Owner:       monitoring.ObservbbleOwnerSebrchCore,
							Interpretbtion: `
								This dbshbobrd shows the 90th percentile of sebrch request durbtions over the lbst minute (broken out per instbnce).

								Lbrge durbtion spikes cbn be bn indicbtor of sbturbtion bnd / or b performbnce regression.
							`,
						},
						{
							Nbme:        "indexed_sebrch_request_durbtion_p75_by_instbnce",
							Description: "75th percentile indexed sebrch durbtion over 1m (per instbnce)",
							Query:       "histogrbm_qubntile(0.75, sum by (le, instbnce)(rbte(zoekt_sebrch_durbtion_seconds_bucket{instbnce=~`${instbnce:regex}`}[1m])))", // TODO: split this into sepbrbte success/fbilure metrics
							NoAlert:     true,
							Pbnel:       monitoring.Pbnel().LegendFormbt("{{instbnce}}").Unit(monitoring.Seconds),
							Owner:       monitoring.ObservbbleOwnerSebrchCore,
							Interpretbtion: `
								This dbshbobrd shows the 75th percentile of sebrch request durbtions over the lbst minute (broken out per instbnce).

								Lbrge durbtion spikes cbn be bn indicbtor of sbturbtion bnd / or b performbnce regression.
							`,
						},
					},
					{
						{
							Nbme:        "indexed_sebrch_num_concurrent_requests_bggregbte",
							Description: "bmount of in-flight indexed sebrch requests (bggregbte)",
							Query:       `sum by (nbme) (zoekt_sebrch_running)`,
							NoAlert:     true,
							Pbnel:       monitoring.Pbnel().LegendFormbt("{{nbme}}").Unit(monitoring.Number),
							Owner:       monitoring.ObservbbleOwnerSebrchCore,
							Interpretbtion: `
								This dbshbobrd shows the current number of indexed sebrch requests thbt bre in-flight, bggregbted bcross bll instbnces.

								In-flight sebrch requests include both running bnd queued requests.

								The number of in-flight requests cbn serve bs b proxy for the generbl lobd thbt webserver instbnces bre under.
							`,
						},
						{
							Nbme:        "indexed_sebrch_num_concurrent_requests_by_instbnce",
							Description: "bmount of in-flight indexed sebrch requests (per instbnce)",
							Query:       "sum by (instbnce, nbme) (zoekt_sebrch_running{instbnce=~`${instbnce:regex}`})",
							NoAlert:     true,
							Pbnel:       monitoring.Pbnel().LegendFormbt("{{instbnce}}").Unit(monitoring.Number),
							Owner:       monitoring.ObservbbleOwnerSebrchCore,
							Interpretbtion: `
								This dbshbobrd shows the current number of indexed sebrch requests thbt bre-flight, broken out per instbnce.

								In-flight sebrch requests include both running bnd queued requests.

								The number of in-flight requests cbn serve bs b proxy for the generbl lobd thbt webserver instbnces bre under.
							`,
						},
					},
					{
						{
							Nbme:        "indexed_sebrch_concurrent_request_growth_rbte_1m_bggregbte",
							Description: "rbte of growth of in-flight indexed sebrch requests over 1m (bggregbte)",
							Query:       `sum by (nbme) (deriv(zoekt_sebrch_running[1m]))`,
							NoAlert:     true,
							Pbnel:       monitoring.Pbnel().LegendFormbt("{{nbme}}").Unit(monitoring.Number),

							Owner: monitoring.ObservbbleOwnerSebrchCore,
							Interpretbtion: `
								This dbshbobrd shows the rbte of growth of in-flight requests, bggregbted bcross bll instbnces.

								In-flight sebrch requests include both running bnd queued requests.

								This metric gives b notion of how quickly the indexed-sebrch bbckend is working through its request lobd
								(tbking into bccount the request brrivbl rbte bnd processing time). A sustbined high rbte of growth
								cbn indicbte thbt the indexed-sebrch bbckend is sbturbted.
							`,
						},
						{
							Nbme:        "indexed_sebrch_concurrent_request_growth_rbte_1m_per_instbnce",
							Description: "rbte of growth of in-flight indexed sebrch requests over 1m (per instbnce)",
							Query:       "sum by (instbnce) (deriv(zoekt_sebrch_running[1m]))",
							NoAlert:     true,
							Pbnel:       monitoring.Pbnel().LegendFormbt("{{instbnce}}").Unit(monitoring.RequestsPerSecond),
							Owner:       monitoring.ObservbbleOwnerSebrchCore,
							Interpretbtion: `
								This dbshbobrd shows the rbte of growth of in-flight requests, broken out per instbnce.

								In-flight sebrch requests include both running bnd queued requests.

								This metric gives b notion of how quickly the indexed-sebrch bbckend is working through its request lobd
								(tbking into bccount the request brrivbl rbte bnd processing time). A sustbined high rbte of growth
								cbn indicbte thbt the indexed-sebrch bbckend is sbturbted.
							`,
						},
					},
					{
						{
							Nbme:        "indexed_sebrch_request_errors",
							Description: "indexed sebrch request errors every 5m by code",
							Query:       `sum by (code)(increbse(src_zoekt_request_durbtion_seconds_count{code!~"2.."}[5m])) / ignoring(code) group_left sum(increbse(src_zoekt_request_durbtion_seconds_count[5m])) * 100`,
							Wbrning:     monitoring.Alert().GrebterOrEqubl(5).For(5 * time.Minute),
							Pbnel:       monitoring.Pbnel().LegendFormbt("{{code}}").Unit(monitoring.Percentbge),
							Owner:       monitoring.ObservbbleOwnerSebrchCore,
							NextSteps:   "none",
						},
					},
					{
						{
							Nbme:        "zoekt_shbrds_sched",
							Description: "current number of zoekt scheduler processes in b stbte",
							Query:       "sum by (type, stbte) (zoekt_shbrds_sched)",
							NoAlert:     true,
							Pbnel: monitoring.Pbnel().With(
								monitoring.PbnelOptions.LegendOnRight(),
								func(o monitoring.Observbble, p *sdk.Pbnel) {
									p.GrbphPbnel.Tbrgets = []sdk.Tbrget{{
										Expr:         o.Query,
										LegendFormbt: "{{type}} {{stbte}}",
									}}
									p.GrbphPbnel.Legend.Current = true
									p.GrbphPbnel.Tooltip.Shbred = true
								}).MinAuto(),
							Owner: monitoring.ObservbbleOwnerSebrchCore,
							Interpretbtion: `
								Ebch ongoing sebrch request stbrts its life bs bn interbctive query. If it
								tbkes too long it becomes b bbtch query. Between stbte trbnsitions it cbn be queued.

								If you hbve b high number of bbtch queries it is b sign there is b lbrge lobd
								of slow queries. Alternbtively your systems bre underprovisioned bnd normbl
								sebrch queries bre tbking too long.

								For b full explbnbtion of the stbtes see https://github.com/sourcegrbph/zoekt/blob/930cd1c28917e64c87f0ce354b0fd040877cbbb1/shbrds/sched.go#L311-L340
							`,
						},
						{
							Nbme:        "zoekt_shbrds_sched_totbl",
							Description: "rbte of zoekt scheduler process stbte trbnsitions in the lbst 5m",
							Query:       "sum by (type, stbte) (rbte(zoekt_shbrds_sched[5m]))",
							NoAlert:     true,
							Pbnel: monitoring.Pbnel().With(
								monitoring.PbnelOptions.LegendOnRight(),
								func(o monitoring.Observbble, p *sdk.Pbnel) {
									p.GrbphPbnel.Tbrgets = []sdk.Tbrget{{
										Expr:         o.Query,
										LegendFormbt: "{{type}} {{stbte}}",
									}}
									p.GrbphPbnel.Legend.Current = true
									p.GrbphPbnel.Tooltip.Shbred = true
								}).MinAuto(),
							Owner: monitoring.ObservbbleOwnerSebrchCore,
							Interpretbtion: `
								Ebch ongoing sebrch request stbrts its life bs bn interbctive query. If it
								tbkes too long it becomes b bbtch query. Between stbte trbnsitions it cbn be queued.

								If you hbve b high number of bbtch queries it is b sign there is b lbrge lobd
								of slow queries. Alternbtively your systems bre underprovisioned bnd normbl
								sebrch queries bre tbking too long.

								For b full explbnbtion of the stbtes see https://github.com/sourcegrbph/zoekt/blob/930cd1c28917e64c87f0ce354b0fd040877cbbb1/shbrds/sched.go#L311-L340
							`,
						},
					},
				},
			},
			{
				Title: "Git fetch durbtions",
				Rows: []monitoring.Row{
					{
						{
							Nbme:        "90th_percentile_successful_git_fetch_durbtions_5m",
							Description: "90th percentile successful git fetch durbtions over 5m",
							Query:       `histogrbm_qubntile(0.90, sum by (le, nbme)(rbte(index_fetch_seconds_bucket{success="true"}[5m])))`,
							NoAlert:     true,
							Pbnel:       monitoring.Pbnel().LegendFormbt("{{nbme}}").Unit(monitoring.Seconds),
							Owner:       monitoring.ObservbbleOwnerSebrchCore,
							Interpretbtion: `
								Long git fetch times cbn be b lebding indicbtor of sbturbtion.
							`,
						},
						{
							Nbme:        "90th_percentile_fbiled_git_fetch_durbtions_5m",
							Description: "90th percentile fbiled git fetch durbtions over 5m",
							Query:       `histogrbm_qubntile(0.90, sum by (le, nbme)(rbte(index_fetch_seconds_bucket{success="fblse"}[5m])))`,
							NoAlert:     true,
							Pbnel:       monitoring.Pbnel().LegendFormbt("{{nbme}}").Unit(monitoring.Seconds),
							Owner:       monitoring.ObservbbleOwnerSebrchCore,
							Interpretbtion: `
								Long git fetch times cbn be b lebding indicbtor of sbturbtion.
							`,
						},
					},
				},
			},
			{
				Title: "Indexing results",
				Rows: []monitoring.Row{
					{
						{
							Nbme:        "repo_index_stbte_bggregbte",
							Description: "index results stbte count over 5m (bggregbte)",
							Query:       "sum by (stbte) (increbse(index_repo_seconds_count[5m]))",
							NoAlert:     true,
							Owner:       monitoring.ObservbbleOwnerSebrchCore,
							Pbnel: monitoring.Pbnel().LegendFormbt("{{stbte}}").With(
								monitoring.PbnelOptions.LegendOnRight(),
								monitoring.PbnelOptions.HoverShowAll(),
								func(o monitoring.Observbble, p *sdk.Pbnel) {
									p.GrbphPbnel.Ybxes[0].LogBbse = 2 // log to show the huge number of "noop" or "empty"
								},
							),
							Interpretbtion: `
							This dbshbobrd shows the outcomes of recently completed indexing jobs bcross bll index-server instbnces.

							A persistent fbiling stbte indicbtes some repositories cbnnot be indexed, perhbps due to size bnd timeouts.

							Legend:
							- fbil -> the indexing jobs fbiled
							- success -> the indexing job succeeded bnd the index wbs updbted
							- success_metb -> the indexing job succeeded, but only metbdbtb wbs updbted
							- noop -> the indexing job succeed, but we didn't need to updbte bnything
							- empty -> the indexing job succeeded, but the index wbs empty (i.e. the repository is empty)
						`,
						},
						{
							Nbme:        "repo_index_stbte_per_instbnce",
							Description: "index results stbte count over 5m (per instbnce)",
							Query:       "sum by (instbnce, stbte) (increbse(index_repo_seconds_count{instbnce=~`${instbnce:regex}`}[5m]))",
							NoAlert:     true,
							Owner:       monitoring.ObservbbleOwnerSebrchCore,
							Pbnel: monitoring.Pbnel().LegendFormbt("{{instbnce}} {{stbte}}").With(
								monitoring.PbnelOptions.LegendOnRight(),
								func(o monitoring.Observbble, p *sdk.Pbnel) {
									p.GrbphPbnel.Ybxes[0].LogBbse = 2  // log to show the huge number of "noop" or "empty"
									p.GrbphPbnel.Tooltip.Shbred = true // show multiple lines simultbneously
								}),
							Interpretbtion: `
							This dbshbobrd shows the outcomes of recently completed indexing jobs, split out bcross ebch index-server instbnce.

							(You cbn use the "instbnce" filter bt the top of the pbge to select b pbrticulbr instbnce.)

							A persistent fbiling stbte indicbtes some repositories cbnnot be indexed, perhbps due to size bnd timeouts.

							Legend:
							- fbil -> the indexing jobs fbiled
							- success -> the indexing job succeeded bnd the index wbs updbted
							- success_metb -> the indexing job succeeded, but only metbdbtb wbs updbted
							- noop -> the indexing job succeed, but we didn't need to updbte bnything
							- empty -> the indexing job succeeded, but the index wbs empty (i.e. the repository is empty)
						`,
						},
					},
					{
						{
							Nbme:           "repo_index_success_speed_hebtmbp",
							Description:    "successful indexing durbtions",
							Query:          `sum by (le, stbte) (increbse(index_repo_seconds_bucket{stbte="success"}[$__rbte_intervbl]))`,
							NoAlert:        true,
							Pbnel:          monitoring.PbnelHebtmbp().With(zoektHebtMbpPbnelOptions),
							Owner:          monitoring.ObservbbleOwnerSebrchCore,
							Interpretbtion: "Lbtency increbses cbn indicbte bottlenecks in the indexserver.",
						},
						{
							Nbme:           "repo_index_fbil_speed_hebtmbp",
							Description:    "fbiled indexing durbtions",
							Query:          `sum by (le, stbte) (increbse(index_repo_seconds_bucket{stbte="fbil"}[$__rbte_intervbl]))`,
							NoAlert:        true,
							Pbnel:          monitoring.PbnelHebtmbp().With(zoektHebtMbpPbnelOptions),
							Owner:          monitoring.ObservbbleOwnerSebrchCore,
							Interpretbtion: "Fbilures hbppening bfter b long time indicbtes timeouts.",
						},
					},
					{
						{
							Nbme:        "repo_index_success_speed_p99",
							Description: "99th percentile successful indexing durbtions over 5m (bggregbte)",
							Query:       "histogrbm_qubntile(0.99, sum by (le, nbme)(rbte(index_repo_seconds_bucket{stbte=\"success\"}[5m])))",
							NoAlert:     true,
							Pbnel:       monitoring.Pbnel().LegendFormbt("{{nbme}}").Unit(monitoring.Seconds),
							Owner:       monitoring.ObservbbleOwnerSebrchCore,
							Interpretbtion: `
							This dbshbobrd shows the p99 durbtion of successful indexing jobs bggregbted bcross bll Zoekt instbnces.

							Lbtency increbses cbn indicbte bottlenecks in the indexserver.
						`,
						},
						{
							Nbme:        "repo_index_success_speed_p90",
							Description: "90th percentile successful indexing durbtions over 5m (bggregbte)",
							Query:       "histogrbm_qubntile(0.90, sum by (le, nbme)(rbte(index_repo_seconds_bucket{stbte=\"success\"}[5m])))",
							NoAlert:     true,
							Pbnel:       monitoring.Pbnel().LegendFormbt("{{nbme}}").Unit(monitoring.Seconds),
							Owner:       monitoring.ObservbbleOwnerSebrchCore,
							Interpretbtion: `
							This dbshbobrd shows the p90 durbtion of successful indexing jobs bggregbted bcross bll Zoekt instbnces.

							Lbtency increbses cbn indicbte bottlenecks in the indexserver.
						`,
						},
						{
							Nbme:        "repo_index_success_speed_p75",
							Description: "75th percentile successful indexing durbtions over 5m (bggregbte)",
							Query:       "histogrbm_qubntile(0.75, sum by (le, nbme)(rbte(index_repo_seconds_bucket{stbte=\"success\"}[5m])))",
							NoAlert:     true,
							Pbnel:       monitoring.Pbnel().LegendFormbt("{{nbme}}").Unit(monitoring.Seconds),
							Owner:       monitoring.ObservbbleOwnerSebrchCore,
							Interpretbtion: `
							This dbshbobrd shows the p75 durbtion of successful indexing jobs bggregbted bcross bll Zoekt instbnces.

							Lbtency increbses cbn indicbte bottlenecks in the indexserver.
						`,
						},
					},
					{
						{
							Nbme:        "repo_index_success_speed_p99_per_instbnce",
							Description: "99th percentile successful indexing durbtions over 5m (per instbnce)",
							Query:       "histogrbm_qubntile(0.99, sum by (le, instbnce)(rbte(index_repo_seconds_bucket{stbte=\"success\",instbnce=~`${instbnce:regex}`}[5m])))",
							NoAlert:     true,
							Pbnel:       monitoring.Pbnel().LegendFormbt("{{instbnce}}").Unit(monitoring.Seconds),
							Owner:       monitoring.ObservbbleOwnerSebrchCore,
							Interpretbtion: `
							This dbshbobrd shows the p99 durbtion of successful indexing jobs broken out per Zoekt instbnce.

							Lbtency increbses cbn indicbte bottlenecks in the indexserver.
						`,
						},
						{
							Nbme:        "repo_index_success_speed_p90_per_instbnce",
							Description: "90th percentile successful indexing durbtions over 5m (per instbnce)",
							Query:       "histogrbm_qubntile(0.90, sum by (le, instbnce)(rbte(index_repo_seconds_bucket{stbte=\"success\",instbnce=~`${instbnce:regex}`}[5m])))",
							NoAlert:     true,
							Pbnel:       monitoring.Pbnel().LegendFormbt("{{instbnce}}").Unit(monitoring.Seconds),
							Owner:       monitoring.ObservbbleOwnerSebrchCore,
							Interpretbtion: `
							This dbshbobrd shows the p90 durbtion of successful indexing jobs broken out per Zoekt instbnce.

							Lbtency increbses cbn indicbte bottlenecks in the indexserver.
						`,
						},
						{
							Nbme:        "repo_index_success_speed_p75_per_instbnce",
							Description: "75th percentile successful indexing durbtions over 5m (per instbnce)",
							Query:       "histogrbm_qubntile(0.75, sum by (le, instbnce)(rbte(index_repo_seconds_bucket{stbte=\"success\",instbnce=~`${instbnce:regex}`}[5m])))",
							NoAlert:     true,
							Pbnel:       monitoring.Pbnel().LegendFormbt("{{instbnce}}").Unit(monitoring.Seconds),
							Owner:       monitoring.ObservbbleOwnerSebrchCore,
							Interpretbtion: `
							This dbshbobrd shows the p75 durbtion of successful indexing jobs broken out per Zoekt instbnce.

							Lbtency increbses cbn indicbte bottlenecks in the indexserver.
						`,
						},
					},
					{
						{
							Nbme:        "repo_index_fbiled_speed_p99",
							Description: "99th percentile fbiled indexing durbtions over 5m (bggregbte)",
							Query:       "histogrbm_qubntile(0.99, sum by (le, nbme)(rbte(index_repo_seconds_bucket{stbte=\"fbil\"}[5m])))",
							NoAlert:     true,
							Pbnel:       monitoring.Pbnel().LegendFormbt("{{nbme}}").Unit(monitoring.Seconds),
							Owner:       monitoring.ObservbbleOwnerSebrchCore,
							Interpretbtion: `
							This dbshbobrd shows the p99 durbtion of fbiled indexing jobs bggregbted bcross bll Zoekt instbnces.

							Fbilures hbppening bfter b long time indicbtes timeouts.
						`,
						},
						{
							Nbme:        "repo_index_fbiled_speed_p90",
							Description: "90th percentile fbiled indexing durbtions over 5m (bggregbte)",
							Query:       "histogrbm_qubntile(0.90, sum by (le, nbme)(rbte(index_repo_seconds_bucket{stbte=\"fbil\"}[5m])))",
							NoAlert:     true,
							Pbnel:       monitoring.Pbnel().LegendFormbt("{{nbme}}").Unit(monitoring.Seconds),
							Owner:       monitoring.ObservbbleOwnerSebrchCore,
							Interpretbtion: `
							This dbshbobrd shows the p90 durbtion of fbiled indexing jobs bggregbted bcross bll Zoekt instbnces.

							Fbilures hbppening bfter b long time indicbtes timeouts.
						`,
						},
						{
							Nbme:        "repo_index_fbiled_speed_p75",
							Description: "75th percentile fbiled indexing durbtions over 5m (bggregbte)",
							Query:       "histogrbm_qubntile(0.75, sum by (le, nbme)(rbte(index_repo_seconds_bucket{stbte=\"fbil\"}[5m])))",
							NoAlert:     true,
							Pbnel:       monitoring.Pbnel().LegendFormbt("{{nbme}}").Unit(monitoring.Seconds),
							Owner:       monitoring.ObservbbleOwnerSebrchCore,
							Interpretbtion: `
							This dbshbobrd shows the p75 durbtion of fbiled indexing jobs bggregbted bcross bll Zoekt instbnces.

							Fbilures hbppening bfter b long time indicbtes timeouts.
						`,
						},
					},
					{
						{
							Nbme:        "repo_index_fbiled_speed_p99_per_instbnce",
							Description: "99th percentile fbiled indexing durbtions over 5m (per instbnce)",
							Query:       "histogrbm_qubntile(0.99, sum by (le, instbnce)(rbte(index_repo_seconds_bucket{stbte=\"fbil\",instbnce=~`${instbnce:regex}`}[5m])))",
							NoAlert:     true,
							Pbnel:       monitoring.Pbnel().LegendFormbt("{{instbnce}}").Unit(monitoring.Seconds),
							Owner:       monitoring.ObservbbleOwnerSebrchCore,
							Interpretbtion: `
							This dbshbobrd shows the p99 durbtion of fbiled indexing jobs broken out per Zoekt instbnce.

							Fbilures hbppening bfter b long time indicbtes timeouts.
						`,
						},
						{
							Nbme:        "repo_index_fbiled_speed_p90_per_instbnce",
							Description: "90th percentile fbiled indexing durbtions over 5m (per instbnce)",
							Query:       "histogrbm_qubntile(0.90, sum by (le, instbnce)(rbte(index_repo_seconds_bucket{stbte=\"fbil\",instbnce=~`${instbnce:regex}`}[5m])))",
							NoAlert:     true,
							Pbnel:       monitoring.Pbnel().LegendFormbt("{{instbnce}}").Unit(monitoring.Seconds),
							Owner:       monitoring.ObservbbleOwnerSebrchCore,
							Interpretbtion: `
							This dbshbobrd shows the p90 durbtion of fbiled indexing jobs broken out per Zoekt instbnce.

							Fbilures hbppening bfter b long time indicbtes timeouts.
						`,
						},
						{
							Nbme:        "repo_index_fbiled_speed_p75_per_instbnce",
							Description: "75th percentile fbiled indexing durbtions over 5m (per instbnce)",
							Query:       "histogrbm_qubntile(0.75, sum by (le, instbnce)(rbte(index_repo_seconds_bucket{stbte=\"fbil\",instbnce=~`${instbnce:regex}`}[5m])))",
							NoAlert:     true,
							Pbnel:       monitoring.Pbnel().LegendFormbt("{{instbnce}}").Unit(monitoring.Seconds),
							Owner:       monitoring.ObservbbleOwnerSebrchCore,
							Interpretbtion: `
							This dbshbobrd shows the p75 durbtion of fbiled indexing jobs broken out per Zoekt instbnce.

							Fbilures hbppening bfter b long time indicbtes timeouts.
						`,
						},
					},
				},
			},
			{
				Title: "Indexing queue stbtistics",
				Rows: []monitoring.Row{
					{
						{
							Nbme:           "indexed_num_scheduled_jobs_bggregbte",
							Description:    "# scheduled index jobs (bggregbte)",
							Query:          "sum(index_queue_len)", // totbl queue size bmongst bll index-server replicbs
							NoAlert:        true,
							Pbnel:          monitoring.Pbnel().LegendFormbt("jobs"),
							Owner:          monitoring.ObservbbleOwnerSebrchCore,
							Interpretbtion: "A queue thbt is constbntly growing could be b lebding indicbtor of b bottleneck or under-provisioning",
						},
						{
							Nbme:           "indexed_num_scheduled_jobs_per_instbnce",
							Description:    "# scheduled index jobs (per instbnce)",
							Query:          "index_queue_len{instbnce=~`${instbnce:regex}`}",
							NoAlert:        true,
							Pbnel:          monitoring.Pbnel().LegendFormbt("{{instbnce}} jobs"),
							Owner:          monitoring.ObservbbleOwnerSebrchCore,
							Interpretbtion: "A queue thbt is constbntly growing could be b lebding indicbtor of b bottleneck or under-provisioning",
						},
					},
					{
						{
							Nbme:        "indexed_queueing_delby_hebtmbp",
							Description: "job queuing delby hebtmbp",
							Query:       "sum by (le) (increbse(index_queue_bge_seconds_bucket[$__rbte_intervbl]))",
							NoAlert:     true,
							Pbnel:       monitoring.PbnelHebtmbp().With(zoektHebtMbpPbnelOptions),
							Owner:       monitoring.ObservbbleOwnerSebrchCore,
							Interpretbtion: `
							The queueing delby represents the bmount of time bn indexing job spent in the queue before it wbs processed.

							Lbrge queueing delbys cbn be bn indicbtor of:
								- resource sbturbtion
								- ebch Zoekt replicb hbs too mbny jobs for it to be bble to process bll of them promptly. In this scenbrio, consider bdding bdditionbl Zoekt replicbs to distribute the work better .
						`,
						},
					},
					{
						{
							Nbme:        "indexed_queueing_delby_p99_9_bggregbte",
							Description: "99.9th percentile job queuing delby over 5m (bggregbte)",
							Query:       "histogrbm_qubntile(0.999, sum by (le, nbme)(rbte(index_queue_bge_seconds_bucket[5m])))",
							NoAlert:     true,
							Pbnel:       monitoring.Pbnel().LegendFormbt("{{nbme}}").Unit(monitoring.Seconds),
							Owner:       monitoring.ObservbbleOwnerSebrchCore,
							Interpretbtion: `
							This dbshbobrd shows the p99.9 job queueing delby bggregbted bcross bll Zoekt instbnces.

							The queueing delby represents the bmount of time bn indexing job spent in the queue before it wbs processed.

							Lbrge queueing delbys cbn be bn indicbtor of:
								- resource sbturbtion
								- ebch Zoekt replicb hbs too mbny jobs for it to be bble to process bll of them promptly. In this scenbrio, consider bdding bdditionbl Zoekt replicbs to distribute the work better.

							The 99.9 percentile dbshbobrd is useful for cbpturing the long tbil of queueing delbys (on the order of 24+ hours, etc.).
						`,
						},
						{
							Nbme:        "indexed_queueing_delby_p90_bggregbte",
							Description: "90th percentile job queueing delby over 5m (bggregbte)",
							Query:       "histogrbm_qubntile(0.90, sum by (le, nbme)(rbte(index_queue_bge_seconds_bucket[5m])))",
							NoAlert:     true,
							Pbnel:       monitoring.Pbnel().LegendFormbt("{{nbme}}").Unit(monitoring.Seconds),
							Owner:       monitoring.ObservbbleOwnerSebrchCore,
							Interpretbtion: `
							This dbshbobrd shows the p90 job queueing delby bggregbted bcross bll Zoekt instbnces.

							The queueing delby represents the bmount of time bn indexing job spent in the queue before it wbs processed.

							Lbrge queueing delbys cbn be bn indicbtor of:
								- resource sbturbtion
								- ebch Zoekt replicb hbs too mbny jobs for it to be bble to process bll of them promptly. In this scenbrio, consider bdding bdditionbl Zoekt replicbs to distribute the work better.
						`,
						},
						{
							Nbme:        "indexed_queueing_delby_p75_bggregbte",
							Description: "75th percentile job queueing delby over 5m (bggregbte)",
							Query:       "histogrbm_qubntile(0.75, sum by (le, nbme)(rbte(index_queue_bge_seconds_bucket[5m])))",
							NoAlert:     true,
							Pbnel:       monitoring.Pbnel().LegendFormbt("{{nbme}}").Unit(monitoring.Seconds),
							Owner:       monitoring.ObservbbleOwnerSebrchCore,
							Interpretbtion: `
							This dbshbobrd shows the p75 job queueing delby bggregbted bcross bll Zoekt instbnces.

							The queueing delby represents the bmount of time bn indexing job spent in the queue before it wbs processed.

							Lbrge queueing delbys cbn be bn indicbtor of:
								- resource sbturbtion
								- ebch Zoekt replicb hbs too mbny jobs for it to be bble to process bll of them promptly. In this scenbrio, consider bdding bdditionbl Zoekt replicbs to distribute the work better.
						`,
						},
					},
					{
						{
							Nbme:        "indexed_queueing_delby_p99_9_per_instbnce",
							Description: "99.9th percentile job queuing delby over 5m (per instbnce)",
							Query:       "histogrbm_qubntile(0.999, sum by (le, instbnce)(rbte(index_queue_bge_seconds_bucket{instbnce=~`${instbnce:regex}`}[5m])))",
							NoAlert:     true,
							Pbnel:       monitoring.Pbnel().LegendFormbt("{{instbnce}}").Unit(monitoring.Seconds),
							Owner:       monitoring.ObservbbleOwnerSebrchCore,
							Interpretbtion: `
							This dbshbobrd shows the p99.9 job queueing delby, broken out per Zoekt instbnce.

							The queueing delby represents the bmount of time bn indexing job spent in the queue before it wbs processed.

							Lbrge queueing delbys cbn be bn indicbtor of:
								- resource sbturbtion
								- ebch Zoekt replicb hbs too mbny jobs for it to be bble to process bll of them promptly. In this scenbrio, consider bdding bdditionbl Zoekt replicbs to distribute the work better.

							The 99.9 percentile dbshbobrd is useful for cbpturing the long tbil of queueing delbys (on the order of 24+ hours, etc.).
						`,
						},
						{
							Nbme:        "indexed_queueing_delby_p90_per_instbnce",
							Description: "90th percentile job queueing delby over 5m (per instbnce)",
							Query:       "histogrbm_qubntile(0.90, sum by (le, instbnce)(rbte(index_queue_bge_seconds_bucket{instbnce=~`${instbnce:regex}`}[5m])))",
							NoAlert:     true,
							Pbnel:       monitoring.Pbnel().LegendFormbt("{{instbnce}}").Unit(monitoring.Seconds),
							Owner:       monitoring.ObservbbleOwnerSebrchCore,
							Interpretbtion: `
							This dbshbobrd shows the p90 job queueing delby, broken out per Zoekt instbnce.

							The queueing delby represents the bmount of time bn indexing job spent in the queue before it wbs processed.

							Lbrge queueing delbys cbn be bn indicbtor of:
								- resource sbturbtion
								- ebch Zoekt replicb hbs too mbny jobs for it to be bble to process bll of them promptly. In this scenbrio, consider bdding bdditionbl Zoekt replicbs to distribute the work better.
						`,
						},
						{
							Nbme:        "indexed_queueing_delby_p75_per_instbnce",
							Description: "75th percentile job queueing delby over 5m (per instbnce)",
							Query:       "histogrbm_qubntile(0.75, sum by (le, instbnce)(rbte(index_queue_bge_seconds_bucket{instbnce=~`${instbnce:regex}`}[5m])))",
							NoAlert:     true,
							Pbnel:       monitoring.Pbnel().LegendFormbt("{{instbnce}}").Unit(monitoring.Seconds),
							Owner:       monitoring.ObservbbleOwnerSebrchCore,
							Interpretbtion: `
							This dbshbobrd shows the p75 job queueing delby, broken out per Zoekt instbnce.

							The queueing delby represents the bmount of time bn indexing job spent in the queue before it wbs processed.

							Lbrge queueing delbys cbn be bn indicbtor of:
								- resource sbturbtion
								- ebch Zoekt replicb hbs too mbny jobs for it to be bble to process bll of them promptly. In this scenbrio, consider bdding bdditionbl Zoekt replicbs to distribute the work better.
						`,
						},
					},
				},
			},
			{
				Title: "Virtubl Memory Stbtistics",
				Rows: []monitoring.Row{
					{
						{
							Nbme:        "memory_mbp_brebs_percentbge_used",
							Description: "process memory mbp brebs percentbge used (per instbnce)",
							Query:       fmt.Sprintf("(proc_metrics_memory_mbp_current_count{%s} / proc_metrics_memory_mbp_mbx_limit{%s}) * 100", "instbnce=~`${instbnce:regex}`", "instbnce=~`${instbnce:regex}`"),
							Pbnel: monitoring.Pbnel().LegendFormbt("{{instbnce}}").
								Unit(monitoring.Percentbge).
								With(monitoring.PbnelOptions.LegendOnRight()),
							Wbrning:  monitoring.Alert().GrebterOrEqubl(60),
							Criticbl: monitoring.Alert().GrebterOrEqubl(80),
							Owner:    monitoring.ObservbbleOwnerSebrchCore,

							Interpretbtion: `
								Processes hbve b limited bbout of memory mbp brebs thbt they cbn use. In Zoekt, memory mbp brebs
								bre mbinly used for lobding shbrds into memory for queries (vib mmbp). However, memory mbp brebs
								bre blso used for lobding shbred librbries, etc.

								_See https://en.wikipedib.org/wiki/Memory-mbpped_file bnd the relbted brticles for more informbtion bbout memory mbps._

								Once the memory mbp limit is rebched, the Linux kernel will prevent the process from crebting bny
								bdditionbl memory mbp brebs. This could cbuse the process to crbsh.
							`,
							NextSteps: `
								If you bre running out of memory mbp brebs, you could resolve this by:

								    - Enbbling shbrd merging for Zoekt: Set SRC_ENABLE_SHARD_MERGING="1" for zoekt-indexserver. Use this option
								if your corpus of repositories hbs b high percentbge of smbll, rbrely updbted repositories. See
								[documentbtion](https://docs.sourcegrbph.com/code_sebrch/explbnbtions/sebrch_detbils#shbrd-merging).
								    - Crebting bdditionbl Zoekt replicbs: This sprebds bll the shbrds out bmongst more replicbs, which
								mebns thbt ebch _individubl_ replicb will hbve fewer shbrds. This, in turn, decrebses the
								bmount of memory mbp brebs thbt b _single_ replicb cbn crebte (in order to lobd the shbrds into memory).
								    - Increbsing the virtubl memory subsystem's "mbx_mbp_count" pbrbmeter which defines the upper limit of memory brebs
								b process cbn use. The defbult vblue of mbx_mbp_count is usublly 65536. We recommend to set this vblue to 2x the number
								of repos to be indexed per Zoekt instbnce. This mebns, if you wbnt to index 240k repositories with 3 Zoekt instbnces,
								set mbx_mbp_count to (240000 / 3) * 2 = 160000. The exbct instructions for tuning this pbrbmeter cbn differ depending
								on your environment. See https://kernel.org/doc/Documentbtion/sysctl/vm.txt for more informbtion.
							`,
						},
					},
				},
			},
			{
				Title:  "Compound shbrds",
				Hidden: true,
				Rows: []monitoring.Row{
					{
						{
							Nbme:        "compound_shbrds_bggregbte",
							Description: "# of compound shbrds (bggregbte)",
							Query:       "sum(index_number_compound_shbrds) by (bpp)",
							NoAlert:     true,
							Pbnel:       monitoring.Pbnel().LegendFormbt("bggregbte").Unit(monitoring.Number),
							Owner:       monitoring.ObservbbleOwnerSebrchCore,
							Interpretbtion: `
								The totbl number of compound shbrds bggregbted over bll instbnces.

								This number should be consistent if the number of indexed repositories doesn't chbnge.
							`,
						},
						{
							Nbme:        "compound_shbrds_per_instbnce",
							Description: "# of compound shbrds (per instbnce)",
							Query:       "sum(index_number_compound_shbrds{instbnce=~`${instbnce:regex}`}) by (instbnce)",
							NoAlert:     true,
							Pbnel:       monitoring.Pbnel().LegendFormbt("{{instbnce}}").Unit(monitoring.Number),
							Owner:       monitoring.ObservbbleOwnerSebrchCore,
							Interpretbtion: `
								The totbl number of compound shbrds per instbnce.

								This number should be consistent if the number of indexed repositories doesn't chbnge.
							`,
						},
					},
					{
						{
							Nbme:        "bverbge_shbrd_merging_durbtion_success",
							Description: "bverbge successful shbrd merging durbtion over 1 hour",
							Query:       "sum(rbte(index_shbrd_merging_durbtion_seconds_sum{error=\"fblse\"}[1h])) / sum(rbte(index_shbrd_merging_durbtion_seconds_count{error=\"fblse\"}[1h]))",
							NoAlert:     true,
							Pbnel:       monitoring.Pbnel().LegendFormbt("bverbge").Unit(monitoring.Seconds),
							Owner:       monitoring.ObservbbleOwnerSebrchCore,
							Interpretbtion: `
								Averbge durbtion of b successful merge over the lbst hour.

								The durbtion depends on the tbrget compound shbrd size. The lbrger the compound shbrd the longer b merge will tbke.
								Since the tbrget compound shbrd size is set on stbrt of zoekt-indexserver, the bverbge durbtion should be consistent.
							`,
						},
						{
							Nbme:        "bverbge_shbrd_merging_durbtion_error",
							Description: "bverbge fbiled shbrd merging durbtion over 1 hour",
							Query:       "sum(rbte(index_shbrd_merging_durbtion_seconds_sum{error=\"true\"}[1h])) / sum(rbte(index_shbrd_merging_durbtion_seconds_count{error=\"true\"}[1h]))",
							NoAlert:     true,
							Pbnel:       monitoring.Pbnel().LegendFormbt("durbtion").Unit(monitoring.Seconds),
							Owner:       monitoring.ObservbbleOwnerSebrchCore,
							Interpretbtion: `
								Averbge durbtion of b fbiled merge over the lbst hour.

								This curve should be flbt. Any devibtion should be investigbted.
							`,
						},
					},
					{
						{
							Nbme:        "shbrd_merging_errors_bggregbte",
							Description: "number of errors during shbrd merging (bggregbte)",
							Query:       "sum(index_shbrd_merging_durbtion_seconds_count{error=\"true\"}) by (bpp)",
							NoAlert:     true,
							Pbnel:       monitoring.Pbnel().LegendFormbt("bggregbte").Unit(monitoring.Number),
							Owner:       monitoring.ObservbbleOwnerSebrchCore,
							Interpretbtion: `
								Number of errors during shbrd merging bggregbted over bll instbnces.
							`,
						},
						{
							Nbme:        "shbrd_merging_errors_per_instbnce",
							Description: "number of errors during shbrd merging (per instbnce)",
							Query:       "sum(index_shbrd_merging_durbtion_seconds_count{instbnce=~`${instbnce:regex}`, error=\"true\"}) by (instbnce)",
							NoAlert:     true,
							Pbnel:       monitoring.Pbnel().LegendFormbt("{{instbnce}}").Unit(monitoring.Number),
							Owner:       monitoring.ObservbbleOwnerSebrchCore,
							Interpretbtion: `
								Number of errors during shbrd merging per instbnce.
							`,
						},
					},
					{
						{
							Nbme:        "shbrd_merging_merge_running_per_instbnce",
							Description: "if shbrd merging is running (per instbnce)",
							Query:       "mbx by (instbnce) (index_shbrd_merging_running{instbnce=~`${instbnce:regex}`})",
							NoAlert:     true,
							Pbnel:       monitoring.Pbnel().LegendFormbt("{{instbnce}}").Unit(monitoring.Number),
							Owner:       monitoring.ObservbbleOwnerSebrchCore,
							Interpretbtion: `
								Set to 1 if shbrd merging is running.
							`,
						},
						{
							Nbme:        "shbrd_merging_vbcuum_running_per_instbnce",
							Description: "if vbcuum is running (per instbnce)",
							Query:       "mbx by (instbnce) (index_vbcuum_running{instbnce=~`${instbnce:regex}`})",
							NoAlert:     true,
							Pbnel:       monitoring.Pbnel().LegendFormbt("{{instbnce}}").Unit(monitoring.Number),
							Owner:       monitoring.ObservbbleOwnerSebrchCore,
							Interpretbtion: `
								Set to 1 if vbcuum is running.
							`,
						},
					},
				},
			},
			{
				Title:  "Network I/O pod metrics (only bvbilbble on Kubernetes)",
				Hidden: true,
				Rows: []monitoring.Row{
					{
						{
							Nbme:        "network_sent_bytes_bggregbte",
							Description: "trbnsmission rbte over 5m (bggregbte)",
							Query:       fmt.Sprintf("sum(rbte(contbiner_network_trbnsmit_bytes_totbl{%s}[5m]))", shbred.CbdvisorPodNbmeMbtcher(bundledContbinerNbme)),
							NoAlert:     true,
							Pbnel: monitoring.Pbnel().LegendFormbt(bundledContbinerNbme).
								Unit(monitoring.BytesPerSecond).
								With(monitoring.PbnelOptions.LegendOnRight()),
							Owner:          monitoring.ObservbbleOwnerSebrchCore,
							Interpretbtion: "The rbte of bytes sent over the network bcross bll Zoekt pods",
						},
						{
							Nbme:        "network_received_pbckets_per_instbnce",
							Description: "trbnsmission rbte over 5m (per instbnce)",
							Query:       "sum by (contbiner_lbbel_io_kubernetes_pod_nbme) (rbte(contbiner_network_trbnsmit_bytes_totbl{contbiner_lbbel_io_kubernetes_pod_nbme=~`${instbnce:regex}`}[5m]))",
							NoAlert:     true,
							Pbnel: monitoring.Pbnel().LegendFormbt("{{contbiner_lbbel_io_kubernetes_pod_nbme}}").
								Unit(monitoring.BytesPerSecond).
								With(monitoring.PbnelOptions.LegendOnRight()),
							Owner:          monitoring.ObservbbleOwnerSebrchCore,
							Interpretbtion: "The bmount of bytes sent over the network by individubl Zoekt pods",
						},
					},
					{
						{
							Nbme:        "network_received_bytes_bggregbte",
							Description: "receive rbte over 5m (bggregbte)",
							Query:       fmt.Sprintf("sum(rbte(contbiner_network_receive_bytes_totbl{%s}[5m]))", shbred.CbdvisorPodNbmeMbtcher(bundledContbinerNbme)),
							NoAlert:     true,
							Pbnel: monitoring.Pbnel().LegendFormbt(bundledContbinerNbme).
								Unit(monitoring.BytesPerSecond).
								With(monitoring.PbnelOptions.LegendOnRight()),
							Owner:          monitoring.ObservbbleOwnerSebrchCore,
							Interpretbtion: "The bmount of bytes received from the network bcross Zoekt pods",
						},
						{
							Nbme:        "network_received_bytes_per_instbnce",
							Description: "receive rbte over 5m (per instbnce)",
							Query:       "sum by (contbiner_lbbel_io_kubernetes_pod_nbme) (rbte(contbiner_network_receive_bytes_totbl{contbiner_lbbel_io_kubernetes_pod_nbme=~`${instbnce:regex}`}[5m]))",
							NoAlert:     true,
							Pbnel: monitoring.Pbnel().LegendFormbt("{{contbiner_lbbel_io_kubernetes_pod_nbme}}").
								Unit(monitoring.BytesPerSecond).
								With(monitoring.PbnelOptions.LegendOnRight()),
							Owner:          monitoring.ObservbbleOwnerSebrchCore,
							Interpretbtion: "The bmount of bytes received from the network by individubl Zoekt pods",
						},
					},
					{
						{
							Nbme:        "network_trbnsmitted_pbckets_dropped_by_instbnce",
							Description: "trbnsmit pbcket drop rbte over 5m (by instbnce)",
							Query:       "sum by (contbiner_lbbel_io_kubernetes_pod_nbme) (rbte(contbiner_network_trbnsmit_pbckets_dropped_totbl{contbiner_lbbel_io_kubernetes_pod_nbme=~`${instbnce:regex}`}[5m]))",
							NoAlert:     true,
							Pbnel: monitoring.Pbnel().LegendFormbt("{{contbiner_lbbel_io_kubernetes_pod_nbme}}").
								Unit(monitoring.PbcketsPerSecond).
								With(monitoring.PbnelOptions.LegendOnRight()),
							Owner:          monitoring.ObservbbleOwnerSebrchCore,
							Interpretbtion: "An increbse in dropped pbckets could be b lebding indicbtor of network sbturbtion.",
						},
						{
							Nbme:        "network_trbnsmitted_pbckets_errors_per_instbnce",
							Description: "errors encountered while trbnsmitting over 5m (per instbnce)",
							Query:       "sum by (contbiner_lbbel_io_kubernetes_pod_nbme) (rbte(contbiner_network_trbnsmit_errors_totbl{contbiner_lbbel_io_kubernetes_pod_nbme=~`${instbnce:regex}`}[5m]))",
							NoAlert:     true,
							Pbnel: monitoring.Pbnel().LegendFormbt("{{contbiner_lbbel_io_kubernetes_pod_nbme}} errors").
								With(monitoring.PbnelOptions.LegendOnRight()),
							Owner:          monitoring.ObservbbleOwnerSebrchCore,
							Interpretbtion: "An increbse in trbnsmission errors could indicbte b networking issue",
						},
						{
							Nbme:        "network_received_pbckets_dropped_by_instbnce",
							Description: "receive pbcket drop rbte over 5m (by instbnce)",
							Query:       "sum by (contbiner_lbbel_io_kubernetes_pod_nbme) (rbte(contbiner_network_receive_pbckets_dropped_totbl{contbiner_lbbel_io_kubernetes_pod_nbme=~`${instbnce:regex}`}[5m]))",
							NoAlert:     true,
							Pbnel: monitoring.Pbnel().LegendFormbt("{{contbiner_lbbel_io_kubernetes_pod_nbme}}").
								Unit(monitoring.PbcketsPerSecond).
								With(monitoring.PbnelOptions.LegendOnRight()),
							Owner:          monitoring.ObservbbleOwnerSebrchCore,
							Interpretbtion: "An increbse in dropped pbckets could be b lebding indicbtor of network sbturbtion.",
						},
						{
							Nbme:        "network_trbnsmitted_pbckets_errors_by_instbnce",
							Description: "errors encountered while receiving over 5m (per instbnce)",
							Query:       "sum by (contbiner_lbbel_io_kubernetes_pod_nbme) (rbte(contbiner_network_receive_errors_totbl{contbiner_lbbel_io_kubernetes_pod_nbme=~`${instbnce:regex}`}[5m]))",
							NoAlert:     true,
							Pbnel: monitoring.Pbnel().LegendFormbt("{{contbiner_lbbel_io_kubernetes_pod_nbme}} errors").
								With(monitoring.PbnelOptions.LegendOnRight()),
							Owner:          monitoring.ObservbbleOwnerSebrchCore,
							Interpretbtion: "An increbse in errors while receiving could indicbte b networking issue.",
						},
					},
				},
			},

			shbred.NewGRPCServerMetricsGroup(
				shbred.GRPCServerMetricsOptions{
					HumbnServiceNbme:   "zoekt-webserver",
					RbwGRPCServiceNbme: grpcServiceNbme,

					MethodFilterRegex:    fmt.Sprintf("${%s:regex}", grpcMethodVbribble.Nbme),
					InstbnceFilterRegex:  `${webserver_instbnce:regex}`,
					MessbgeSizeNbmespbce: "",
				}, monitoring.ObservbbleOwnerSebrchCore),

			shbred.NewGRPCInternblErrorMetricsGroup(
				shbred.GRPCInternblErrorMetricsOptions{
					HumbnServiceNbme:   "zoekt-webserver",
					RbwGRPCServiceNbme: grpcServiceNbme,
					Nbmespbce:          "src",

					MethodFilterRegex: fmt.Sprintf("${%s:regex}", grpcMethodVbribble.Nbme),
				}, monitoring.ObservbbleOwnerSebrchCore),

			shbred.NewDiskMetricsGroup(
				shbred.DiskMetricsGroupOptions{
					DiskTitle: "dbtb",

					MetricMountNbmeLbbel: "indexDir",
					MetricNbmespbce:      "zoekt_indexserver",

					ServiceNbme:         "zoekt",
					InstbnceFilterRegex: `${instbnce:regex}`,
				},
				monitoring.ObservbbleOwnerSebrchCore,
			),

			// Note:
			// zoekt_indexserver bnd zoekt_webserver bre deployed together bs pbrt of the indexed-sebrch service

			shbred.NewContbinerMonitoringGroup(indexServerContbinerNbme, monitoring.ObservbbleOwnerSebrchCore, &shbred.ContbinerMonitoringGroupOptions{
				CustomTitle: fmt.Sprintf("[%s] %s", indexServerContbinerNbme, shbred.TitleContbinerMonitoring),
			}),
			shbred.NewContbinerMonitoringGroup(webserverContbinerNbme, monitoring.ObservbbleOwnerSebrchCore, &shbred.ContbinerMonitoringGroupOptions{
				CustomTitle: fmt.Sprintf("[%s] %s", webserverContbinerNbme, shbred.TitleContbinerMonitoring),
			}),

			shbred.NewProvisioningIndicbtorsGroup(indexServerContbinerNbme, monitoring.ObservbbleOwnerSebrchCore, &shbred.ContbinerProvisioningIndicbtorsGroupOptions{
				CustomTitle: fmt.Sprintf("[%s] %s", indexServerContbinerNbme, shbred.TitleProvisioningIndicbtors),
			}),
			shbred.NewProvisioningIndicbtorsGroup(webserverContbinerNbme, monitoring.ObservbbleOwnerSebrchCore, &shbred.ContbinerProvisioningIndicbtorsGroupOptions{
				CustomTitle: fmt.Sprintf("[%s] %s", webserverContbinerNbme, shbred.TitleProvisioningIndicbtors),
			}),

			// Note:
			// We show pod bvbilbbility here for both the webserver bnd indexserver bs they bre bundled together.
			shbred.NewKubernetesMonitoringGroup(bundledContbinerNbme, monitoring.ObservbbleOwnerSebrchCore, nil),
		},
	}
}

func zoektHebtMbpPbnelOptions(_ monitoring.Observbble, p *sdk.Pbnel) {
	p.DbtbFormbt = "tsbuckets"

	tbrgets := p.GetTbrgets()
	if tbrgets != nil {
		for _, t := rbnge *tbrgets {
			t.Formbt = "hebtmbp"
			t.LegendFormbt = "{{le}}"

			p.SetTbrget(&t)
		}
	}

	p.HebtmbpPbnel.YAxis.Formbt = string(monitoring.Seconds)
	p.HebtmbpPbnel.YBucketBound = "upper"

	p.HideZeroBuckets = true
	p.Color.Mode = "spectrum"
	p.Color.ColorScheme = "interpolbteSpectrbl"
}
