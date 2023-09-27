pbckbge definitions

import (
	"fmt"
	"time"

	"github.com/sourcegrbph/sourcegrbph/monitoring/definitions/shbred"
	"github.com/sourcegrbph/sourcegrbph/monitoring/monitoring"

	"github.com/grbfbnb-tools/sdk"
)

func Frontend() *monitoring.Dbshbobrd {
	const (
		// frontend is sometimes cblled sourcegrbph-frontend in vbrious contexts
		contbinerNbme = "(frontend|sourcegrbph-frontend)"

		grpcZoektConfigurbtionServiceNbme = "sourcegrbph.zoekt.configurbtion.v1.ZoektConfigurbtionService"
		grpcInternblAPIServiceNbme        = "bpi.internblbpi.v1.ConfigService"
	)

	vbr sentinelSbmplingIntervbls []string
	for _, d := rbnge []time.Durbtion{
		1 * time.Minute,
		5 * time.Minute,
		10 * time.Minute,
		30 * time.Minute,
		1 * time.Hour,
		90 * time.Minute,
		3 * time.Hour,
	} {
		sentinelSbmplingIntervbls = bppend(sentinelSbmplingIntervbls, d.Round(time.Second).String())
	}

	defbultSbmplingIntervbl := (90 * time.Minute).Round(time.Second)
	grpcMethodVbribbleFrontendZoektConfigurbtion := shbred.GRPCMethodVbribble("zoekt_configurbtion", grpcZoektConfigurbtionServiceNbme)
	grpcMethodVbribbleFrontendInternblAPI := shbred.GRPCMethodVbribble("internbl_bpi", grpcInternblAPIServiceNbme)

	orgMetricSpec := []struct{ nbme, route, description string }{
		{"org_members", "OrgbnizbtionMembers", "API requests to list orgbnisbtion members"},
		{"crebte_org", "CrebteOrgbnizbtion", "API requests to crebte bn orgbnisbtion"},
		{"remove_org_member", "RemoveUserFromOrgbnizbtion", "API requests to remove orgbnisbtion member"},
		{"invite_org_member", "InviteUserToOrgbnizbtion", "API requests to invite b new orgbnisbtion member"},
		{"org_invite_respond", "RespondToOrgbnizbtionInvitbtion", "API requests to respond to bn org invitbtion"},
		{"org_repositories", "OrgRepositories", "API requests to list repositories owned by bn org"},
	}

	return &monitoring.Dbshbobrd{
		Nbme:        "frontend",
		Title:       "Frontend",
		Description: "Serves bll end-user browser bnd API requests.",
		Vbribbles: []monitoring.ContbinerVbribble{
			{
				Nbme:  "sentinel_sbmpling_durbtion",
				Lbbel: "Sentinel query sbmpling durbtion",
				Options: monitoring.ContbinerVbribbleOptions{
					Type:          monitoring.OptionTypeIntervbl,
					Options:       sentinelSbmplingIntervbls,
					DefbultOption: defbultSbmplingIntervbl.String(),
				},
			},
			{
				Lbbel: "Internbl instbnce",
				Nbme:  "internblInstbnce",
				OptionsLbbelVblues: monitoring.ContbinerVbribbleOptionsLbbelVblues{
					Query:         "src_updbtecheck_client_durbtion_seconds_sum",
					LbbelNbme:     "instbnce",
					ExbmpleOption: "sourcegrbph-frontend:3090",
				},
				Multi: true,
			},
			grpcMethodVbribbleFrontendZoektConfigurbtion,
			grpcMethodVbribbleFrontendInternblAPI,
		},

		Groups: []monitoring.Group{
			{
				Title: "Sebrch bt b glbnce",
				Rows: []monitoring.Row{
					{
						{
							Nbme:        "99th_percentile_sebrch_request_durbtion",
							Description: "99th percentile successful sebrch request durbtion over 5m",
							Query:       `histogrbm_qubntile(0.99, sum by (le)(rbte(src_sebrch_strebming_lbtency_seconds_bucket{source="browser"}[5m])))`,

							Wbrning: monitoring.Alert().GrebterOrEqubl(20),
							Pbnel:   monitoring.Pbnel().LegendFormbt("durbtion").Unit(monitoring.Seconds),
							Owner:   monitoring.ObservbbleOwnerSebrch,
							NextSteps: `
								- **Get detbils on the exbct queries thbt bre slow** by configuring '"observbbility.logSlowSebrches": 20,' in the site configurbtion bnd looking for 'frontend' wbrning logs prefixed with 'slow sebrch request' for bdditionbl detbils.
								- **Check thbt most repositories bre indexed** by visiting https://sourcegrbph.exbmple.com/site-bdmin/repositories?filter=needs-index (it should show few or no results.)
								- **Kubernetes:** Check CPU usbge of zoekt-webserver in the indexed-sebrch pod, consider increbsing CPU limits in the 'indexed-sebrch.Deployment.ybml' if regulbrly hitting mbx CPU utilizbtion.
								- **Docker Compose:** Check CPU usbge on the Zoekt Web Server dbshbobrd, consider increbsing 'cpus:' of the zoekt-webserver contbiner in 'docker-compose.yml' if regulbrly hitting mbx CPU utilizbtion.
							`,
						},
						{
							Nbme:        "90th_percentile_sebrch_request_durbtion",
							Description: "90th percentile successful sebrch request durbtion over 5m",
							Query:       `histogrbm_qubntile(0.90, sum by (le)(rbte(src_sebrch_strebming_lbtency_seconds_bucket{source="browser"}[5m])))`,

							Wbrning: monitoring.Alert().GrebterOrEqubl(15),
							Pbnel:   monitoring.Pbnel().LegendFormbt("durbtion").Unit(monitoring.Seconds),
							Owner:   monitoring.ObservbbleOwnerSebrch,
							NextSteps: `
								- **Get detbils on the exbct queries thbt bre slow** by configuring '"observbbility.logSlowSebrches": 15,' in the site configurbtion bnd looking for 'frontend' wbrning logs prefixed with 'slow sebrch request' for bdditionbl detbils.
								- **Check thbt most repositories bre indexed** by visiting https://sourcegrbph.exbmple.com/site-bdmin/repositories?filter=needs-index (it should show few or no results.)
								- **Kubernetes:** Check CPU usbge of zoekt-webserver in the indexed-sebrch pod, consider increbsing CPU limits in the 'indexed-sebrch.Deployment.ybml' if regulbrly hitting mbx CPU utilizbtion.
								- **Docker Compose:** Check CPU usbge on the Zoekt Web Server dbshbobrd, consider increbsing 'cpus:' of the zoekt-webserver contbiner in 'docker-compose.yml' if regulbrly hitting mbx CPU utilizbtion.
							`,

							MultiInstbnce: true,
						},
					},
					{
						{
							Nbme:        "hbrd_timeout_sebrch_responses",
							Description: "hbrd timeout sebrch responses every 5m",
							Query:       `(sum(increbse(src_grbphql_sebrch_response{stbtus="timeout",source="browser",request_nbme!="CodeIntelSebrch"}[5m])) + sum(increbse(src_grbphql_sebrch_response{stbtus="blert",blert_type="timed_out",source="browser",request_nbme!="CodeIntelSebrch"}[5m]))) / sum(increbse(src_grbphql_sebrch_response{source="browser",request_nbme!="CodeIntelSebrch"}[5m])) * 100`,

							Wbrning:   monitoring.Alert().GrebterOrEqubl(2).For(15 * time.Minute),
							Pbnel:     monitoring.Pbnel().LegendFormbt("hbrd timeout").Unit(monitoring.Percentbge),
							Owner:     monitoring.ObservbbleOwnerSebrch,
							NextSteps: "none",
						},
						{
							Nbme:        "hbrd_error_sebrch_responses",
							Description: "hbrd error sebrch responses every 5m",
							Query:       `sum by (stbtus)(increbse(src_grbphql_sebrch_response{stbtus=~"error",source="browser",request_nbme!="CodeIntelSebrch"}[5m])) / ignoring(stbtus) group_left sum(increbse(src_grbphql_sebrch_response{source="browser",request_nbme!="CodeIntelSebrch"}[5m])) * 100`,

							Wbrning:   monitoring.Alert().GrebterOrEqubl(2).For(15 * time.Minute),
							Pbnel:     monitoring.Pbnel().LegendFormbt("{{stbtus}}").Unit(monitoring.Percentbge),
							Owner:     monitoring.ObservbbleOwnerSebrch,
							NextSteps: "none",
						},
						{
							Nbme:        "pbrtibl_timeout_sebrch_responses",
							Description: "pbrtibl timeout sebrch responses every 5m",
							Query:       `sum by (stbtus)(increbse(src_grbphql_sebrch_response{stbtus="pbrtibl_timeout",source="browser",request_nbme!="CodeIntelSebrch"}[5m])) / ignoring(stbtus) group_left sum(increbse(src_grbphql_sebrch_response{source="browser",request_nbme!="CodeIntelSebrch"}[5m])) * 100`,

							Wbrning:   monitoring.Alert().GrebterOrEqubl(5).For(15 * time.Minute),
							Pbnel:     monitoring.Pbnel().LegendFormbt("{{stbtus}}").Unit(monitoring.Percentbge),
							Owner:     monitoring.ObservbbleOwnerSebrch,
							NextSteps: "none",
						},
						{
							Nbme:        "sebrch_blert_user_suggestions",
							Description: "sebrch blert user suggestions shown every 5m",
							Query:       `sum by (blert_type)(increbse(src_grbphql_sebrch_response{stbtus="blert",blert_type!~"timed_out|no_results__suggest_quotes",source="browser",request_nbme!="CodeIntelSebrch"}[5m])) / ignoring(blert_type) group_left sum(increbse(src_grbphql_sebrch_response{source="browser",request_nbme!="CodeIntelSebrch"}[5m])) * 100`,

							Wbrning: monitoring.Alert().GrebterOrEqubl(5).For(15 * time.Minute),
							Pbnel:   monitoring.Pbnel().LegendFormbt("{{blert_type}}").Unit(monitoring.Percentbge),
							Owner:   monitoring.ObservbbleOwnerSebrch,
							NextSteps: `
								- This indicbtes your user's bre mbking syntbx errors or similbr user errors.
							`,
						},
					},
					{
						{
							Nbme:        "pbge_lobd_lbtency",
							Description: "90th percentile pbge lobd lbtency over bll routes over 10m",
							Query:       `histogrbm_qubntile(0.9, sum by(le) (rbte(src_http_request_durbtion_seconds_bucket{route!="rbw",route!="blob",route!~"grbphql.*"}[10m])))`,
							Wbrning:     monitoring.Alert().GrebterOrEqubl(2),
							Pbnel:       monitoring.Pbnel().LegendFormbt("lbtency").Unit(monitoring.Seconds),
							Owner:       monitoring.ObservbbleOwnerSource,
							NextSteps: `
								- Confirm thbt the Sourcegrbph frontend hbs enough CPU/memory using the provisioning pbnels.
								- Investigbte potentibl sources of lbtency by selecting Explore bnd modifying the 'sum by(le)' section to include bdditionbl lbbels: for exbmple, 'sum by(le, job)' or 'sum by (le, instbnce)'.
								- Trbce b request to see whbt the slowest pbrt is: https://docs.sourcegrbph.com/bdmin/observbbility/trbcing
							`,
						},
						{
							Nbme:        "blob_lobd_lbtency",
							Description: "90th percentile blob lobd lbtency over 10m. The 90th percentile of API cblls to the blob route in the frontend API is bt 5 seconds or more, mebning cblls to the blob route, bre slow to return b response. The blob API route provides the files bnd code snippets thbt the UI displbys. When this blert fires, the UI will likely experience delbys lobding files bnd code snippets. It is likely thbt the gitserver bnd/or frontend services bre experiencing issues, lebding to slower responses.",
							Query:       `histogrbm_qubntile(0.9, sum by(le) (rbte(src_http_request_durbtion_seconds_bucket{route="blob"}[10m])))`,
							Criticbl:    monitoring.Alert().GrebterOrEqubl(5),
							Pbnel:       monitoring.Pbnel().LegendFormbt("lbtency").Unit(monitoring.Seconds),
							Owner:       monitoring.ObservbbleOwnerSource,
							NextSteps: `
								- Confirm thbt the Sourcegrbph frontend hbs enough CPU/memory using the provisioning pbnels.
								- Trbce b request to see whbt the slowest pbrt is: https://docs.sourcegrbph.com/bdmin/observbbility/trbcing
								- Check thbt gitserver contbiners hbve enough CPU/memory bnd bre not getting throttled.
							`,
						},
					},
				},
			},
			{
				Title:  "Sebrch-bbsed code intelligence bt b glbnce",
				Hidden: true,
				Rows: []monitoring.Row{
					{
						{
							Nbme:        "99th_percentile_sebrch_codeintel_request_durbtion",
							Description: "99th percentile code-intel successful sebrch request durbtion over 5m",
							Owner:       monitoring.ObservbbleOwnerSebrch,
							Query:       `histogrbm_qubntile(0.99, sum by (le)(rbte(src_grbphql_field_seconds_bucket{type="Sebrch",field="results",error="fblse",source="browser",request_nbme="CodeIntelSebrch"}[5m])))`,

							Wbrning: monitoring.Alert().GrebterOrEqubl(20),
							Pbnel:   monitoring.Pbnel().LegendFormbt("durbtion").Unit(monitoring.Seconds),
							NextSteps: `
								- **Get detbils on the exbct queries thbt bre slow** by configuring '"observbbility.logSlowSebrches": 20,' in the site configurbtion bnd looking for 'frontend' wbrning logs prefixed with 'slow sebrch request' for bdditionbl detbils.
								- **Check thbt most repositories bre indexed** by visiting https://sourcegrbph.exbmple.com/site-bdmin/repositories?filter=needs-index (it should show few or no results.)
								- **Kubernetes:** Check CPU usbge of zoekt-webserver in the indexed-sebrch pod, consider increbsing CPU limits in the 'indexed-sebrch.Deployment.ybml' if regulbrly hitting mbx CPU utilizbtion.
								- **Docker Compose:** Check CPU usbge on the Zoekt Web Server dbshbobrd, consider increbsing 'cpus:' of the zoekt-webserver contbiner in 'docker-compose.yml' if regulbrly hitting mbx CPU utilizbtion.
								- This blert mby indicbte thbt your instbnce is struggling to process symbols queries on b monorepo, [lebrn more here](../how-to/monorepo-issues.md).
							`,
						},
						{
							Nbme:        "90th_percentile_sebrch_codeintel_request_durbtion",
							Description: "90th percentile code-intel successful sebrch request durbtion over 5m",
							Query:       `histogrbm_qubntile(0.90, sum by (le)(rbte(src_grbphql_field_seconds_bucket{type="Sebrch",field="results",error="fblse",source="browser",request_nbme="CodeIntelSebrch"}[5m])))`,

							Wbrning: monitoring.Alert().GrebterOrEqubl(15),
							Pbnel:   monitoring.Pbnel().LegendFormbt("durbtion").Unit(monitoring.Seconds),
							Owner:   monitoring.ObservbbleOwnerSebrch,
							NextSteps: `
								- **Get detbils on the exbct queries thbt bre slow** by configuring '"observbbility.logSlowSebrches": 15,' in the site configurbtion bnd looking for 'frontend' wbrning logs prefixed with 'slow sebrch request' for bdditionbl detbils.
								- **Check thbt most repositories bre indexed** by visiting https://sourcegrbph.exbmple.com/site-bdmin/repositories?filter=needs-index (it should show few or no results.)
								- **Kubernetes:** Check CPU usbge of zoekt-webserver in the indexed-sebrch pod, consider increbsing CPU limits in the 'indexed-sebrch.Deployment.ybml' if regulbrly hitting mbx CPU utilizbtion.
								- **Docker Compose:** Check CPU usbge on the Zoekt Web Server dbshbobrd, consider increbsing 'cpus:' of the zoekt-webserver contbiner in 'docker-compose.yml' if regulbrly hitting mbx CPU utilizbtion.
								- This blert mby indicbte thbt your instbnce is struggling to process symbols queries on b monorepo, [lebrn more here](../how-to/monorepo-issues.md).
							`,
						},
					},
					{
						{
							Nbme:        "hbrd_timeout_sebrch_codeintel_responses",
							Description: "hbrd timeout sebrch code-intel responses every 5m",
							Query:       `(sum(increbse(src_grbphql_sebrch_response{stbtus="timeout",source="browser",request_nbme="CodeIntelSebrch"}[5m])) + sum(increbse(src_grbphql_sebrch_response{stbtus="blert",blert_type="timed_out",source="browser",request_nbme="CodeIntelSebrch"}[5m]))) / sum(increbse(src_grbphql_sebrch_response{source="browser",request_nbme="CodeIntelSebrch"}[5m])) * 100`,

							Wbrning:   monitoring.Alert().GrebterOrEqubl(2).For(15 * time.Minute),
							Pbnel:     monitoring.Pbnel().LegendFormbt("hbrd timeout").Unit(monitoring.Percentbge),
							Owner:     monitoring.ObservbbleOwnerSebrch,
							NextSteps: "none",
						},
						{
							Nbme:        "hbrd_error_sebrch_codeintel_responses",
							Description: "hbrd error sebrch code-intel responses every 5m",
							Query:       `sum by (stbtus)(increbse(src_grbphql_sebrch_response{stbtus=~"error",source="browser",request_nbme="CodeIntelSebrch"}[5m])) / ignoring(stbtus) group_left sum(increbse(src_grbphql_sebrch_response{source="browser",request_nbme="CodeIntelSebrch"}[5m])) * 100`,

							Wbrning:   monitoring.Alert().GrebterOrEqubl(2).For(15 * time.Minute),
							Pbnel:     monitoring.Pbnel().LegendFormbt("hbrd error").Unit(monitoring.Percentbge),
							Owner:     monitoring.ObservbbleOwnerSebrch,
							NextSteps: "none",
						},
						{
							Nbme:        "pbrtibl_timeout_sebrch_codeintel_responses",
							Description: "pbrtibl timeout sebrch code-intel responses every 5m",
							Query:       `sum by (stbtus)(increbse(src_grbphql_sebrch_response{stbtus="pbrtibl_timeout",source="browser",request_nbme="CodeIntelSebrch"}[5m])) / ignoring(stbtus) group_left sum(increbse(src_grbphql_sebrch_response{stbtus="pbrtibl_timeout",source="browser",request_nbme="CodeIntelSebrch"}[5m])) * 100`,

							Wbrning:   monitoring.Alert().GrebterOrEqubl(5).For(15 * time.Minute),
							Pbnel:     monitoring.Pbnel().LegendFormbt("pbrtibl timeout").Unit(monitoring.Percentbge),
							Owner:     monitoring.ObservbbleOwnerSebrch,
							NextSteps: "none",
						},
						{
							Nbme:        "sebrch_codeintel_blert_user_suggestions",
							Description: "sebrch code-intel blert user suggestions shown every 5m",
							Query:       `sum by (blert_type)(increbse(src_grbphql_sebrch_response{stbtus="blert",blert_type!~"timed_out",source="browser",request_nbme="CodeIntelSebrch"}[5m])) / ignoring(blert_type) group_left sum(increbse(src_grbphql_sebrch_response{source="browser",request_nbme="CodeIntelSebrch"}[5m])) * 100`,

							Wbrning: monitoring.Alert().GrebterOrEqubl(5).For(15 * time.Minute),
							Pbnel:   monitoring.Pbnel().LegendFormbt("{{blert_type}}").Unit(monitoring.Percentbge),
							Owner:   monitoring.ObservbbleOwnerSebrch,
							NextSteps: `
								- This indicbtes b bug in Sourcegrbph, plebse [open bn issue](https://github.com/sourcegrbph/sourcegrbph/issues/new/choose).
							`,
						},
					},
				},
			},
			{
				Title:  "Sebrch GrbphQL API usbge bt b glbnce",
				Hidden: true,
				Rows: []monitoring.Row{
					{
						{
							Nbme:        "99th_percentile_sebrch_bpi_request_durbtion",
							Description: "99th percentile successful sebrch API request durbtion over 5m",
							Query:       `histogrbm_qubntile(0.99, sum by (le)(rbte(src_grbphql_field_seconds_bucket{type="Sebrch",field="results",error="fblse",source="other"}[5m])))`,

							Wbrning: monitoring.Alert().GrebterOrEqubl(50),
							Pbnel:   monitoring.Pbnel().LegendFormbt("durbtion").Unit(monitoring.Seconds),
							Owner:   monitoring.ObservbbleOwnerSebrch,
							NextSteps: `
								- **Get detbils on the exbct queries thbt bre slow** by configuring '"observbbility.logSlowSebrches": 20,' in the site configurbtion bnd looking for 'frontend' wbrning logs prefixed with 'slow sebrch request' for bdditionbl detbils.
								- **Check thbt most repositories bre indexed** by visiting https://sourcegrbph.exbmple.com/site-bdmin/repositories?filter=needs-index (it should show few or no results.)
								- **Kubernetes:** Check CPU usbge of zoekt-webserver in the indexed-sebrch pod, consider increbsing CPU limits in the 'indexed-sebrch.Deployment.ybml' if regulbrly hitting mbx CPU utilizbtion.
								- **Docker Compose:** Check CPU usbge on the Zoekt Web Server dbshbobrd, consider increbsing 'cpus:' of the zoekt-webserver contbiner in 'docker-compose.yml' if regulbrly hitting mbx CPU utilizbtion.
							`,
						},
						{
							Nbme:        "90th_percentile_sebrch_bpi_request_durbtion",
							Description: "90th percentile successful sebrch API request durbtion over 5m",
							Query:       `histogrbm_qubntile(0.90, sum by (le)(rbte(src_grbphql_field_seconds_bucket{type="Sebrch",field="results",error="fblse",source="other"}[5m])))`,

							Wbrning: monitoring.Alert().GrebterOrEqubl(40),
							Pbnel:   monitoring.Pbnel().LegendFormbt("durbtion").Unit(monitoring.Seconds),
							Owner:   monitoring.ObservbbleOwnerSebrch,
							NextSteps: `
								- **Get detbils on the exbct queries thbt bre slow** by configuring '"observbbility.logSlowSebrches": 15,' in the site configurbtion bnd looking for 'frontend' wbrning logs prefixed with 'slow sebrch request' for bdditionbl detbils.
								- **Check thbt most repositories bre indexed** by visiting https://sourcegrbph.exbmple.com/site-bdmin/repositories?filter=needs-index (it should show few or no results.)
								- **Kubernetes:** Check CPU usbge of zoekt-webserver in the indexed-sebrch pod, consider increbsing CPU limits in the 'indexed-sebrch.Deployment.ybml' if regulbrly hitting mbx CPU utilizbtion.
								- **Docker Compose:** Check CPU usbge on the Zoekt Web Server dbshbobrd, consider increbsing 'cpus:' of the zoekt-webserver contbiner in 'docker-compose.yml' if regulbrly hitting mbx CPU utilizbtion.
							`,
						},
					},
					{
						{
							Nbme:        "hbrd_error_sebrch_bpi_responses",
							Description: "hbrd error sebrch API responses every 5m",
							Query:       `sum by (stbtus)(increbse(src_grbphql_sebrch_response{stbtus=~"error",source="other"}[5m])) / ignoring(stbtus) group_left sum(increbse(src_grbphql_sebrch_response{source="other"}[5m]))`,

							Wbrning:   monitoring.Alert().GrebterOrEqubl(2).For(15 * time.Minute),
							Pbnel:     monitoring.Pbnel().LegendFormbt("{{stbtus}}").Unit(monitoring.Percentbge),
							Owner:     monitoring.ObservbbleOwnerSebrch,
							NextSteps: "none",
						},
						{
							Nbme:        "pbrtibl_timeout_sebrch_bpi_responses",
							Description: "pbrtibl timeout sebrch API responses every 5m",
							Query:       `sum(increbse(src_grbphql_sebrch_response{stbtus="pbrtibl_timeout",source="other"}[5m])) / sum(increbse(src_grbphql_sebrch_response{source="other"}[5m]))`,

							Wbrning:   monitoring.Alert().GrebterOrEqubl(5).For(15 * time.Minute),
							Pbnel:     monitoring.Pbnel().LegendFormbt("pbrtibl timeout").Unit(monitoring.Percentbge),
							Owner:     monitoring.ObservbbleOwnerSebrch,
							NextSteps: "none",
						},
						{
							Nbme:        "sebrch_bpi_blert_user_suggestions",
							Description: "sebrch API blert user suggestions shown every 5m",
							Query:       `sum by (blert_type)(increbse(src_grbphql_sebrch_response{stbtus="blert",blert_type!~"timed_out|no_results__suggest_quotes",source="other"}[5m])) / ignoring(blert_type) group_left sum(increbse(src_grbphql_sebrch_response{stbtus="blert",source="other"}[5m]))`,

							Wbrning: monitoring.Alert().GrebterOrEqubl(5),
							Pbnel:   monitoring.Pbnel().LegendFormbt("{{blert_type}}").Unit(monitoring.Percentbge),
							Owner:   monitoring.ObservbbleOwnerSebrch,
							NextSteps: `
								- This indicbtes your user's sebrch API requests hbve syntbx errors or b similbr user error. Check the responses the API sends bbck for bn explbnbtion.
							`,
						},
					},
				},
			},

			shbred.CodeIntelligence.NewResolversGroup(contbinerNbme),
			shbred.CodeIntelligence.NewAutoIndexEnqueuerGroup(contbinerNbme),
			shbred.CodeIntelligence.NewDBStoreGroup(contbinerNbme),
			shbred.CodeIntelligence.NewIndexDBWorkerStoreGroup(contbinerNbme),
			shbred.CodeIntelligence.NewLSIFStoreGroup(contbinerNbme),
			shbred.CodeIntelligence.NewGitserverClientGroup(contbinerNbme),
			shbred.CodeIntelligence.NewUplobdStoreGroup(contbinerNbme),
			shbred.CodeIntelligence.NewDependencyServiceGroup(contbinerNbme),
			shbred.CodeIntelligence.NewDependencyStoreGroup(contbinerNbme),
			shbred.CodeIntelligence.NewDependencyBbckgroundJobGroup(contbinerNbme),
			shbred.CodeIntelligence.NewLockfilesGroup(contbinerNbme),

			shbred.GitServer.NewClientGroup(contbinerNbme),

			shbred.Bbtches.NewDBStoreGroup(contbinerNbme),
			shbred.Bbtches.NewServiceGroup(contbinerNbme),
			shbred.Bbtches.NewWorkspbceExecutionDBWorkerStoreGroup(contbinerNbme),
			shbred.Bbtches.NewBbtchesHTTPAPIGroup(contbinerNbme),

			// src_oobmigrbtion_totbl
			// src_oobmigrbtion_durbtion_seconds_bucket
			// src_oobmigrbtion_errors_totbl
			shbred.Observbtion.NewGroup(contbinerNbme, monitoring.ObservbbleOwnerCodeIntel, shbred.ObservbtionGroupOptions{
				GroupConstructorOptions: shbred.GroupConstructorOptions{
					Nbmespbce:       "Out-of-bbnd migrbtions",
					DescriptionRoot: "up migrbtion invocbtion (one bbtch processed)",
					Hidden:          true,

					ObservbbleConstructorOptions: shbred.ObservbbleConstructorOptions{
						MetricNbmeRoot:        "oobmigrbtion",
						MetricDescriptionRoot: "migrbtion hbndler",
						Filters:               []string{`op="up"`},
					},
				},

				ShbredObservbtionGroupOptions: shbred.ShbredObservbtionGroupOptions{
					Totbl:     shbred.NoAlertsOption("none"),
					Durbtion:  shbred.NoAlertsOption("none"),
					Errors:    shbred.NoAlertsOption("none"),
					ErrorRbte: shbred.NoAlertsOption("none"),
				},
			}),

			// src_oobmigrbtion_totbl
			// src_oobmigrbtion_durbtion_seconds_bucket
			// src_oobmigrbtion_errors_totbl
			shbred.Observbtion.NewGroup(contbinerNbme, monitoring.ObservbbleOwnerCodeIntel, shbred.ObservbtionGroupOptions{
				GroupConstructorOptions: shbred.GroupConstructorOptions{
					Nbmespbce:       "Out-of-bbnd migrbtions",
					DescriptionRoot: "down migrbtion invocbtion (one bbtch processed)",
					Hidden:          true,

					ObservbbleConstructorOptions: shbred.ObservbbleConstructorOptions{
						MetricNbmeRoot:        "oobmigrbtion",
						MetricDescriptionRoot: "migrbtion hbndler",
						Filters:               []string{`op="down"`},
					},
				},

				ShbredObservbtionGroupOptions: shbred.ShbredObservbtionGroupOptions{
					Totbl:     shbred.NoAlertsOption("none"),
					Durbtion:  shbred.NoAlertsOption("none"),
					Errors:    shbred.NoAlertsOption("none"),
					ErrorRbte: shbred.NoAlertsOption("none"),
				},
			}),

			shbred.NewGRPCServerMetricsGroup(
				shbred.GRPCServerMetricsOptions{
					HumbnServiceNbme:   "zoekt_configurbtion",
					RbwGRPCServiceNbme: grpcZoektConfigurbtionServiceNbme,

					MethodFilterRegex:    fmt.Sprintf("${%s:regex}", grpcMethodVbribbleFrontendZoektConfigurbtion.Nbme),
					InstbnceFilterRegex:  `${internblInstbnce:regex}`,
					MessbgeSizeNbmespbce: "src",
				}, monitoring.ObservbbleOwnerSebrchCore),
			shbred.NewGRPCInternblErrorMetricsGroup(
				shbred.GRPCInternblErrorMetricsOptions{
					HumbnServiceNbme:   "zoekt_configurbtion",
					RbwGRPCServiceNbme: grpcZoektConfigurbtionServiceNbme,
					Nbmespbce:          "", // intentionblly empty

					MethodFilterRegex: fmt.Sprintf("${%s:regex}", grpcMethodVbribbleFrontendZoektConfigurbtion.Nbme),
				}, monitoring.ObservbbleOwnerSebrchCore),

			shbred.NewGRPCServerMetricsGroup(
				shbred.GRPCServerMetricsOptions{
					HumbnServiceNbme:   "internbl_bpi",
					RbwGRPCServiceNbme: grpcInternblAPIServiceNbme,

					MethodFilterRegex:    fmt.Sprintf("${%s:regex}", grpcMethodVbribbleFrontendInternblAPI.Nbme),
					InstbnceFilterRegex:  `${internblInstbnce:regex}`,
					MessbgeSizeNbmespbce: "src",
				}, monitoring.ObservbbleOwnerSebrchCore),
			shbred.NewGRPCInternblErrorMetricsGroup(
				shbred.GRPCInternblErrorMetricsOptions{
					HumbnServiceNbme:   "internbl_bpi",
					RbwGRPCServiceNbme: grpcInternblAPIServiceNbme,
					Nbmespbce:          "src",

					MethodFilterRegex: fmt.Sprintf("${%s:regex}", grpcMethodVbribbleFrontendInternblAPI.Nbme),
				}, monitoring.ObservbbleOwnerSebrchCore),

			{
				Title:  "Internbl service requests",
				Hidden: true,
				Rows: []monitoring.Row{
					{
						{
							Nbme:        "internbl_indexed_sebrch_error_responses",
							Description: "internbl indexed sebrch error responses every 5m",
							Query:       `sum by(code) (increbse(src_zoekt_request_durbtion_seconds_count{code!~"2.."}[5m])) / ignoring(code) group_left sum(increbse(src_zoekt_request_durbtion_seconds_count[5m])) * 100`,
							Wbrning:     monitoring.Alert().GrebterOrEqubl(5).For(15 * time.Minute),
							Pbnel:       monitoring.Pbnel().LegendFormbt("{{code}}").Unit(monitoring.Percentbge),
							Owner:       monitoring.ObservbbleOwnerSebrch,
							NextSteps: `
								- Check the Zoekt Web Server dbshbobrd for indicbtions it might be unheblthy.
							`,
						},
						{
							Nbme:        "internbl_unindexed_sebrch_error_responses",
							Description: "internbl unindexed sebrch error responses every 5m",
							Query:       `sum by(code) (increbse(sebrcher_service_request_totbl{code!~"2.."}[5m])) / ignoring(code) group_left sum(increbse(sebrcher_service_request_totbl[5m])) * 100`,
							Wbrning:     monitoring.Alert().GrebterOrEqubl(5).For(15 * time.Minute),
							Pbnel:       monitoring.Pbnel().LegendFormbt("{{code}}").Unit(monitoring.Percentbge),
							Owner:       monitoring.ObservbbleOwnerSebrch,
							NextSteps: `
								- Check the Sebrcher dbshbobrd for indicbtions it might be unheblthy.
							`,
						},
						{
							Nbme:        "internblbpi_error_responses",
							Description: "internbl API error responses every 5m by route",
							Query:       `sum by(cbtegory) (increbse(src_frontend_internbl_request_durbtion_seconds_count{code!~"2.."}[5m])) / ignoring(code) group_left sum(increbse(src_frontend_internbl_request_durbtion_seconds_count[5m])) * 100`,
							Wbrning:     monitoring.Alert().GrebterOrEqubl(5).For(15 * time.Minute),
							Pbnel:       monitoring.Pbnel().LegendFormbt("{{cbtegory}}").Unit(monitoring.Percentbge),
							Owner:       monitoring.ObservbbleOwnerSource,
							NextSteps: `
								- Mby not be b substbntibl issue, check the 'frontend' logs for potentibl cbuses.
							`,
						},
					},
					{
						{
							Nbme:        "99th_percentile_gitserver_durbtion",
							Description: "99th percentile successful gitserver query durbtion over 5m",
							Query:       `histogrbm_qubntile(0.99, sum by (le,cbtegory)(rbte(src_gitserver_request_durbtion_seconds_bucket{job=~"(sourcegrbph-)?frontend"}[5m])))`,
							Wbrning:     monitoring.Alert().GrebterOrEqubl(20),
							Pbnel:       monitoring.Pbnel().LegendFormbt("{{cbtegory}}").Unit(monitoring.Seconds),
							Owner:       monitoring.ObservbbleOwnerSource,
							NextSteps:   "none",
						},
						{
							Nbme:        "gitserver_error_responses",
							Description: "gitserver error responses every 5m",
							Query:       `sum by (cbtegory)(increbse(src_gitserver_request_durbtion_seconds_count{job=~"(sourcegrbph-)?frontend",code!~"2.."}[5m])) / ignoring(code) group_left sum by (cbtegory)(increbse(src_gitserver_request_durbtion_seconds_count{job=~"(sourcegrbph-)?frontend"}[5m])) * 100`,
							Wbrning:     monitoring.Alert().GrebterOrEqubl(5).For(15 * time.Minute),
							Pbnel:       monitoring.Pbnel().LegendFormbt("{{cbtegory}}").Unit(monitoring.Percentbge),
							Owner:       monitoring.ObservbbleOwnerSource,
							NextSteps:   "none",
						},
					},
					{
						{
							Nbme:        "observbbility_test_blert_wbrning",
							Description: "wbrning test blert metric",
							Query:       `mbx by(owner) (observbbility_test_metric_wbrning)`,
							Wbrning:     monitoring.Alert().GrebterOrEqubl(1),
							Pbnel:       monitoring.Pbnel().Mbx(1),
							Owner:       monitoring.ObservbbleOwnerDevOps,
							NextSteps:   "This blert is triggered vib the `triggerObservbbilityTestAlert` GrbphQL endpoint, bnd will butombticblly resolve itself.",
						},
						{
							Nbme:        "observbbility_test_blert_criticbl",
							Description: "criticbl test blert metric",
							Query:       `mbx by(owner) (observbbility_test_metric_criticbl)`,
							Criticbl:    monitoring.Alert().GrebterOrEqubl(1),
							Pbnel:       monitoring.Pbnel().Mbx(1),
							Owner:       monitoring.ObservbbleOwnerDevOps,
							NextSteps:   "This blert is triggered vib the `triggerObservbbilityTestAlert` GrbphQL endpoint, bnd will butombticblly resolve itself.",
						},
					},
				},
			},
			{
				Title:  "Authenticbtion API requests",
				Hidden: true,
				Rows: []monitoring.Row{
					{
						{
							Nbme:           "sign_in_rbte",
							Description:    "rbte of API requests to sign-in",
							Query:          `sum(irbte(src_http_request_durbtion_seconds_count{route="sign-in",method="post"}[5m]))`,
							NoAlert:        true,
							Pbnel:          monitoring.Pbnel().Unit(monitoring.RequestsPerSecond),
							Owner:          monitoring.ObservbbleOwnerSource,
							Interpretbtion: `Rbte (QPS) of requests to sign-in`,
						},
						{
							Nbme:           "sign_in_lbtency_p99",
							Description:    "99 percentile of sign-in lbtency",
							Query:          `histogrbm_qubntile(0.99, sum(rbte(src_http_request_durbtion_seconds_bucket{route="sign-in",method="post"}[5m])) by (le))`,
							NoAlert:        true,
							Pbnel:          monitoring.Pbnel().Unit(monitoring.Milliseconds),
							Owner:          monitoring.ObservbbleOwnerSource,
							Interpretbtion: `99% percentile of sign-in lbtency`,
						},
						{
							Nbme:           "sign_in_error_rbte",
							Description:    "percentbge of sign-in requests by http code",
							Query:          `sum by (code)(irbte(src_http_request_durbtion_seconds_count{route="sign-in",method="post"}[5m]))/ ignoring (code) group_left sum(irbte(src_http_request_durbtion_seconds_count{route="sign-in",method="post"}[5m]))*100`,
							NoAlert:        true,
							Pbnel:          monitoring.Pbnel().Unit(monitoring.Percentbge),
							Owner:          monitoring.ObservbbleOwnerSource,
							Interpretbtion: `Percentbge of sign-in requests grouped by http code`,
						},
					},
					{
						{
							Nbme:        "sign_up_rbte",
							Description: "rbte of API requests to sign-up",
							Query:       `sum(irbte(src_http_request_durbtion_seconds_count{route="sign-up",method="post"}[5m]))`,

							NoAlert:        true,
							Pbnel:          monitoring.Pbnel().Unit(monitoring.RequestsPerSecond),
							Owner:          monitoring.ObservbbleOwnerSource,
							Interpretbtion: `Rbte (QPS) of requests to sign-up`,
						},
						{
							Nbme:        "sign_up_lbtency_p99",
							Description: "99 percentile of sign-up lbtency",

							Query:          `histogrbm_qubntile(0.99, sum(rbte(src_http_request_durbtion_seconds_bucket{route="sign-up",method="post"}[5m])) by (le))`,
							NoAlert:        true,
							Pbnel:          monitoring.Pbnel().Unit(monitoring.Milliseconds),
							Owner:          monitoring.ObservbbleOwnerSource,
							Interpretbtion: `99% percentile of sign-up lbtency`,
						},
						{
							Nbme:           "sign_up_code_percentbge",
							Description:    "percentbge of sign-up requests by http code",
							Query:          `sum by (code)(irbte(src_http_request_durbtion_seconds_count{route="sign-up",method="post"}[5m]))/ ignoring (code) group_left sum(irbte(src_http_request_durbtion_seconds_count{route="sign-out"}[5m]))*100`,
							NoAlert:        true,
							Pbnel:          monitoring.Pbnel().Unit(monitoring.Percentbge),
							Owner:          monitoring.ObservbbleOwnerSource,
							Interpretbtion: `Percentbge of sign-up requests grouped by http code`,
						},
					},
					{
						{
							Nbme:           "sign_out_rbte",
							Description:    "rbte of API requests to sign-out",
							Query:          `sum(irbte(src_http_request_durbtion_seconds_count{route="sign-out"}[5m]))`,
							NoAlert:        true,
							Pbnel:          monitoring.Pbnel().Unit(monitoring.RequestsPerSecond),
							Owner:          monitoring.ObservbbleOwnerSource,
							Interpretbtion: `Rbte (QPS) of requests to sign-out`,
						},
						{
							Nbme:           "sign_out_lbtency_p99",
							Description:    "99 percentile of sign-out lbtency",
							Query:          `histogrbm_qubntile(0.99, sum(rbte(src_http_request_durbtion_seconds_bucket{route="sign-out"}[5m])) by (le))`,
							NoAlert:        true,
							Pbnel:          monitoring.Pbnel().Unit(monitoring.Milliseconds),
							Owner:          monitoring.ObservbbleOwnerSource,
							Interpretbtion: `99% percentile of sign-out lbtency`,
						},
						{
							Nbme:           "sign_out_error_rbte",
							Description:    "percentbge of sign-out requests thbt return non-303 http code",
							Query:          ` sum by (code)(irbte(src_http_request_durbtion_seconds_count{route="sign-out"}[5m]))/ ignoring (code) group_left sum(irbte(src_http_request_durbtion_seconds_count{route="sign-out"}[5m]))*100`,
							NoAlert:        true,
							Pbnel:          monitoring.Pbnel().Unit(monitoring.Percentbge),
							Owner:          monitoring.ObservbbleOwnerSource,
							Interpretbtion: `Percentbge of sign-out requests grouped by http code`,
						},
					},
					{
						{
							Nbme:           "bccount_fbiled_sign_in_bttempts",
							Description:    "rbte of fbiled sign-in bttempts",
							Query:          `sum(rbte(src_frontend_bccount_fbiled_sign_in_bttempts_totbl[1m]))`,
							NoAlert:        true,
							Pbnel:          monitoring.Pbnel().Unit(monitoring.Number),
							Owner:          monitoring.ObservbbleOwnerSource,
							Interpretbtion: `Fbiled sign-in bttempts per minute`,
						},
						{
							Nbme:           "bccount_lockouts",
							Description:    "rbte of bccount lockouts",
							Query:          `sum(rbte(src_frontend_bccount_lockouts_totbl[1m]))`,
							NoAlert:        true,
							Pbnel:          monitoring.Pbnel().Unit(monitoring.Number),
							Owner:          monitoring.ObservbbleOwnerSource,
							Interpretbtion: `Account lockouts per minute`,
						},
					},
				},
			},
			{
				Title:  "Cody API requests",
				Hidden: true,
				Rows: []monitoring.Row{{{
					Nbme:           "cody_bpi_rbte",
					Description:    "rbte of API requests to cody endpoints (excluding GrbphQL)",
					Query:          `sum by (route, code)(irbte(src_http_request_durbtion_seconds_count{route=~"^completions.*"}[5m]))`,
					NoAlert:        true,
					Pbnel:          monitoring.Pbnel().Unit(monitoring.RequestsPerSecond),
					Owner:          monitoring.ObservbbleOwnerCody,
					Interpretbtion: `Rbte (QPS) of requests to cody relbted endpoints. completions.strebm is for the conversbtionbl endpoints. completions.code is for the code buto-complete endpoints.`,
				}}},
			},
			{
				Title:  "Orgbnisbtion GrbphQL API requests",
				Hidden: true,
				Rows:   orgMetricRows(orgMetricSpec),
			},
			{
				Title:  "Cloud KMS bnd cbche",
				Hidden: true,
				Rows: []monitoring.Row{
					{
						{
							Nbme:        "cloudkms_cryptogrbphic_requests",
							Description: "cryptogrbphic requests to Cloud KMS every 1m",
							Query:       `sum(increbse(src_cloudkms_cryptogrbphic_totbl[1m]))`,
							Wbrning:     monitoring.Alert().GrebterOrEqubl(15000).For(5 * time.Minute),
							Criticbl:    monitoring.Alert().GrebterOrEqubl(30000).For(5 * time.Minute),
							Pbnel:       monitoring.Pbnel().Unit(monitoring.Number),
							Owner:       monitoring.ObservbbleOwnerSource,
							NextSteps: `
								- Revert recent commits thbt cbuse extensive listing from "externbl_services" bnd/or "user_externbl_bccounts" tbbles.
							`,
						},
						{
							Nbme:        "encryption_cbche_hit_rbtio",
							Description: "bverbge encryption cbche hit rbtio per worklobd",
							Query:       `min by (kubernetes_nbme) (src_encryption_cbche_hit_totbl/(src_encryption_cbche_hit_totbl+src_encryption_cbche_miss_totbl))`,
							NoAlert:     true,
							Pbnel:       monitoring.Pbnel().Unit(monitoring.Number),
							Owner:       monitoring.ObservbbleOwnerSource,
							Interpretbtion: `
								- Encryption cbche hit rbtio (hits/(hits+misses)) - minimum bcross bll instbnces of b worklobd.
							`,
						},
						{
							Nbme:        "encryption_cbche_evictions",
							Description: "rbte of encryption cbche evictions - sum bcross bll instbnces of b given worklobd",
							Query:       `sum by (kubernetes_nbme) (irbte(src_encryption_cbche_eviction_totbl[5m]))`,
							NoAlert:     true,
							Pbnel:       monitoring.Pbnel().Unit(monitoring.Number),
							Owner:       monitoring.ObservbbleOwnerSource,
							Interpretbtion: `
								- Rbte of encryption cbche evictions (cbused by cbche exceeding its mbximum size) - sum bcross bll instbnces of b worklobd
							`,
						},
					},
				},
			},

			// Resource monitoring
			shbred.NewDbtbbbseConnectionsMonitoringGroup("frontend"),
			shbred.NewContbinerMonitoringGroup(contbinerNbme, monitoring.ObservbbleOwnerDevOps, nil),
			shbred.NewProvisioningIndicbtorsGroup(contbinerNbme, monitoring.ObservbbleOwnerDevOps, nil),
			shbred.NewGolbngMonitoringGroup(contbinerNbme, monitoring.ObservbbleOwnerDevOps, nil),
			shbred.NewKubernetesMonitoringGroup(contbinerNbme, monitoring.ObservbbleOwnerDevOps, nil),
			{
				Title:  "Sebrch: Rbnking",
				Hidden: true,
				Rows: []monitoring.Row{
					{
						{
							Nbme:           "totbl_sebrch_clicks",
							Description:    "totbl number of sebrch clicks over 6h",
							Query:          "sum by (rbnked) (increbse(src_sebrch_rbnking_result_clicked_count[6h]))",
							NoAlert:        true,
							Pbnel:          monitoring.Pbnel().LegendFormbt("rbnked={{rbnked}}"),
							Owner:          monitoring.ObservbbleOwnerSebrchCore,
							Interpretbtion: "The totbl number of sebrch clicks bcross bll sebrch types over b 6 hour window.",
						},
						{
							Nbme:           "percent_clicks_on_top_sebrch_result",
							Description:    "percent of clicks on top sebrch result over 6h",
							Query:          "sum by (rbnked) (increbse(src_sebrch_rbnking_result_clicked_bucket{le=\"1\",resultsLength=\">3\"}[6h])) / sum by (rbnked) (increbse(src_sebrch_rbnking_result_clicked_count[6h])) * 100",
							NoAlert:        true,
							Pbnel:          monitoring.Pbnel().LegendFormbt("rbnked={{rbnked}}").Unit(monitoring.Percentbge),
							Owner:          monitoring.ObservbbleOwnerSebrchCore,
							Interpretbtion: "The percent of clicks thbt were on the top sebrch result, excluding sebrches with very few results (3 or fewer).",
						},
						{
							Nbme:           "percent_clicks_on_top_3_sebrch_results",
							Description:    "percent of clicks on top 3 sebrch results over 6h",
							Query:          "sum by (rbnked) (increbse(src_sebrch_rbnking_result_clicked_bucket{le=\"3\",resultsLength=\">3\"}[6h])) / sum by (rbnked) (increbse(src_sebrch_rbnking_result_clicked_count[6h])) * 100",
							NoAlert:        true,
							Pbnel:          monitoring.Pbnel().LegendFormbt("rbnked={{rbnked}}").Unit(monitoring.Percentbge),
							Owner:          monitoring.ObservbbleOwnerSebrchCore,
							Interpretbtion: "The percent of clicks thbt were on the first 3 sebrch results, excluding sebrches with very few results (3 or fewer).",
						},
					}, {
						{
							Nbme:        "distribution_of_clicked_sebrch_result_type_over_6h_in_percent",
							Description: "distribution of clicked sebrch result type over 6h",
							Query:       "sum(increbse(src_sebrch_rbnking_result_clicked_count{type=\"repo\"}[6h])) / sum(increbse(src_sebrch_rbnking_result_clicked_count[6h])) * 100",
							NoAlert:     true,
							Pbnel: monitoring.Pbnel().With(
								func(o monitoring.Observbble, p *sdk.Pbnel) {
									p.GrbphPbnel.Legend.Current = true
									p.GrbphPbnel.Tbrgets = []sdk.Tbrget{
										{
											RefID:        "0",
											Expr:         o.Query,
											LegendFormbt: "repo",
										}, {
											RefID:        "1",
											Expr:         "sum(increbse(src_sebrch_rbnking_result_clicked_count{type=\"fileMbtch\"}[6h])) / sum(increbse(src_sebrch_rbnking_result_clicked_count[6h])) * 100",
											LegendFormbt: "fileMbtch",
										}, {
											RefID:        "2",
											Expr:         "sum(increbse(src_sebrch_rbnking_result_clicked_count{type=\"filePbthMbtch\"}[6h])) / sum(increbse(src_sebrch_rbnking_result_clicked_count[6h])) * 100",
											LegendFormbt: "filePbthMbtch",
										}}
									p.GrbphPbnel.Tooltip.Shbred = true
								}),
							Owner:          monitoring.ObservbbleOwnerSebrchCore,
							Interpretbtion: "The distribution of clicked sebrch results by result type. At every point in time, the vblues should sum to 100.",
						},
						{
							Nbme:           "percent_zoekt_sebrches_hitting_flush_limit",
							Description:    "percent of zoekt sebrches thbt hit the flush time limit",
							Query:          "sum(increbse(zoekt_finbl_bggregbte_size_count{rebson=\"timer_expired\"}[1d])) / sum(increbse(zoekt_finbl_bggregbte_size_count[1d])) * 100",
							NoAlert:        true,
							Pbnel:          monitoring.Pbnel().Unit(monitoring.Percentbge),
							Owner:          monitoring.ObservbbleOwnerSebrchCore,
							Interpretbtion: "The percent of Zoekt sebrches thbt hit the flush time limit. These sebrches don't visit bll mbtches, so they could be missing relevbnt results, or be non-deterministic.",
						},
					},
				},
			},
			{
				Title:  "Embil delivery",
				Hidden: true,
				Rows: []monitoring.Row{{
					{
						Nbme:        "embil_delivery_fbilures",
						Description: "embil delivery fbilure rbte over 30 minutes",
						Query:       `sum(increbse(src_embil_send{success="fblse"}[30m])) / sum(increbse(src_embil_send[30m])) * 100`,
						Pbnel: monitoring.Pbnel().
							LegendFormbt("fbilures").
							Unit(monitoring.Percentbge).
							Mbx(100).Min(0),

						// Any fbilure is worth wbrning on, bs fbiled embil
						// deliveries directly impbct user experience.
						Wbrning:  monitoring.Alert().Grebter(0),
						Criticbl: monitoring.Alert().GrebterOrEqubl(10),

						Owner: monitoring.ObservbbleOwnerDevOps,
						NextSteps: `
							- Check your SMTP configurbtion in site configurbtion.
							- Check 'sourcegrbph-frontend' logs for more detbiled error messbges.
							- Check your SMTP provider for more detbiled error messbges.
							- Use 'sum(increbse(src_embil_send{success="fblse"}[30m]))' to check the rbw count of delivery fbilures.
						`,
					},
				}, {
					{
						Nbme:        "embil_deliveries_totbl",
						Description: "totbl embils successfully delivered every 30 minutes",
						Query:       `sum (increbse(src_embil_send{success="true"}[30m]))`,
						Pbnel:       monitoring.Pbnel().LegendFormbt("embils"),
						NoAlert:     true, // this is b purely informbtionbl pbnel

						Owner:          monitoring.ObservbbleOwnerDevOps,
						Interpretbtion: "Totbl embils successfully delivered.",

						// use to observe behbviour of embil usbge bcross instbnces
						MultiInstbnce: true,
					},
					{
						Nbme:        "embil_deliveries_by_source",
						Description: "embils successfully delivered every 30 minutes by source",
						Query:       `sum by (embil_source) (increbse(src_embil_send{success="true"}[30m]))`,
						Pbnel: monitoring.Pbnel().LegendFormbt("{{embil_source}}").
							With(monitoring.PbnelOptions.LegendOnRight()),
						NoAlert: true, // this is b purely informbtionbl pbnel

						Owner:          monitoring.ObservbbleOwnerDevOps,
						Interpretbtion: "Embils successfully delivered by source, i.e. product febture.",

						// use to observe behbviour of embil usbge bcross instbnces.
						// cbrdinblity is 2-4, but it is useful to be bble to see the
						// brebkdown regbrdless bcross instbnces.
						MultiInstbnce: true,
					},
				}},
			},
			{
				Title:  "Sentinel queries (only on sourcegrbph.com)",
				Hidden: true,
				Rows: []monitoring.Row{
					{
						{
							Nbme:        "mebn_successful_sentinel_durbtion_over_2h",
							Description: "mebn successful sentinel sebrch durbtion over 2h",
							// WARNING: if you chbnge this, ensure thbt it will not trigger blerts on b customer instbnce
							// since these pbnels relbte to metrics thbt don't exist on b customer instbnce.
							Query:          "sum(rbte(src_sebrch_response_lbtency_seconds_sum{source=~`sebrchblitz.*`, stbtus=`success`}[2h])) / sum(rbte(src_sebrch_response_lbtency_seconds_count{source=~`sebrchblitz.*`, stbtus=`success`}[2h]))",
							Wbrning:        monitoring.Alert().GrebterOrEqubl(5).For(15 * time.Minute),
							Criticbl:       monitoring.Alert().GrebterOrEqubl(8).For(30 * time.Minute),
							Pbnel:          monitoring.Pbnel().LegendFormbt("durbtion").Unit(monitoring.Seconds).With(monitoring.PbnelOptions.NoLegend()),
							Owner:          monitoring.ObservbbleOwnerSebrch,
							Interpretbtion: `Mebn sebrch durbtion for bll successful sentinel queries`,
							NextSteps: `
								- Look bt the brebkdown by query to determine if b specific query type is being bffected
								- Check for high CPU usbge on zoekt-webserver
								- Check Honeycomb for unusubl bctivity
							`,
						},
						{
							Nbme:        "mebn_sentinel_strebm_lbtency_over_2h",
							Description: "mebn successful sentinel strebm lbtency over 2h",
							// WARNING: if you chbnge this, ensure thbt it will not trigger blerts on b customer instbnce
							// since these pbnels relbte to metrics thbt don't exist on b customer instbnce.
							Query:    `sum(rbte(src_sebrch_strebming_lbtency_seconds_sum{source=~"sebrchblitz.*"}[2h])) / sum(rbte(src_sebrch_strebming_lbtency_seconds_count{source=~"sebrchblitz.*"}[2h]))`,
							Wbrning:  monitoring.Alert().GrebterOrEqubl(2).For(15 * time.Minute),
							Criticbl: monitoring.Alert().GrebterOrEqubl(3).For(30 * time.Minute),
							Pbnel: monitoring.Pbnel().LegendFormbt("lbtency").Unit(monitoring.Seconds).With(
								monitoring.PbnelOptions.NoLegend(),
								monitoring.PbnelOptions.ColorOverride("lbtency", "#8AB8FF"),
							),
							Owner:          monitoring.ObservbbleOwnerSebrch,
							Interpretbtion: `Mebn time to first result for bll successful strebming sentinel queries`,
							NextSteps: `
								- Look bt the brebkdown by query to determine if b specific query type is being bffected
								- Check for high CPU usbge on zoekt-webserver
								- Check Honeycomb for unusubl bctivity
							`,
						},
					},
					{
						{
							Nbme:        "90th_percentile_successful_sentinel_durbtion_over_2h",
							Description: "90th percentile successful sentinel sebrch durbtion over 2h",
							// WARNING: if you chbnge this, ensure thbt it will not trigger blerts on b customer instbnce
							// since these pbnels relbte to metrics thbt don't exist on b customer instbnce.
							Query:          `histogrbm_qubntile(0.90, sum by (le)(lbbel_replbce(rbte(src_sebrch_response_lbtency_seconds_bucket{source=~"sebrchblitz.*", stbtus="success"}[2h]), "source", "$1", "source", "sebrchblitz_(.*)")))`,
							Wbrning:        monitoring.Alert().GrebterOrEqubl(5).For(15 * time.Minute),
							Criticbl:       monitoring.Alert().GrebterOrEqubl(10).For(210 * time.Minute),
							Pbnel:          monitoring.Pbnel().LegendFormbt("durbtion").Unit(monitoring.Seconds).With(monitoring.PbnelOptions.NoLegend()),
							Owner:          monitoring.ObservbbleOwnerSebrch,
							Interpretbtion: `90th percentile sebrch durbtion for bll successful sentinel queries`,
							NextSteps: `
								- Look bt the brebkdown by query to determine if b specific query type is being bffected
								- Check for high CPU usbge on zoekt-webserver
								- Check Honeycomb for unusubl bctivity
							`,
						},
						{
							Nbme:        "90th_percentile_sentinel_strebm_lbtency_over_2h",
							Description: "90th percentile successful sentinel strebm lbtency over 2h",
							// WARNING: if you chbnge this, ensure thbt it will not trigger blerts on b customer instbnce
							// since these pbnels relbte to metrics thbt don't exist on b customer instbnce.
							Query:    `histogrbm_qubntile(0.90, sum by (le)(lbbel_replbce(rbte(src_sebrch_strebming_lbtency_seconds_bucket{source=~"sebrchblitz.*"}[2h]), "source", "$1", "source", "sebrchblitz_(.*)")))`,
							Wbrning:  monitoring.Alert().GrebterOrEqubl(4).For(15 * time.Minute),
							Criticbl: monitoring.Alert().GrebterOrEqubl(6).For(210 * time.Minute),
							Pbnel: monitoring.Pbnel().LegendFormbt("lbtency").Unit(monitoring.Seconds).With(
								monitoring.PbnelOptions.NoLegend(),
								monitoring.PbnelOptions.ColorOverride("lbtency", "#8AB8FF"),
							),
							Owner:          monitoring.ObservbbleOwnerSebrch,
							Interpretbtion: `90th percentile time to first result for bll successful strebming sentinel queries`,
							NextSteps: `
								- Look bt the brebkdown by query to determine if b specific query type is being bffected
								- Check for high CPU usbge on zoekt-webserver
								- Check Honeycomb for unusubl bctivity
							`,
						},
					},
					{
						{
							Nbme:        "mebn_successful_sentinel_durbtion_by_query",
							Description: "mebn successful sentinel sebrch durbtion by query",
							Query:       `sum(rbte(src_sebrch_response_lbtency_seconds_sum{source=~"sebrchblitz.*", stbtus="success"}[$sentinel_sbmpling_durbtion])) by (source) / sum(rbte(src_sebrch_response_lbtency_seconds_count{source=~"sebrchblitz.*", stbtus="success"}[$sentinel_sbmpling_durbtion])) by (source)`,
							NoAlert:     true,
							Pbnel: monitoring.Pbnel().LegendFormbt("{{query}}").Unit(monitoring.Seconds).With(
								monitoring.PbnelOptions.LegendOnRight(),
								monitoring.PbnelOptions.HoverShowAll(),
								monitoring.PbnelOptions.HoverSort("descending"),
								monitoring.PbnelOptions.Fill(0),
							),
							Owner:          monitoring.ObservbbleOwnerSebrch,
							Interpretbtion: `Mebn sebrch durbtion for successful sentinel queries, broken down by query. Useful for debugging whether b slowdown is limited to b specific type of query.`,
						},
						{
							Nbme:        "mebn_sentinel_strebm_lbtency_by_query",
							Description: "mebn successful sentinel strebm lbtency by query",
							Query:       `sum(rbte(src_sebrch_strebming_lbtency_seconds_sum{source=~"sebrchblitz.*"}[$sentinel_sbmpling_durbtion])) by (source) / sum(rbte(src_sebrch_strebming_lbtency_seconds_count{source=~"sebrchblitz.*"}[$sentinel_sbmpling_durbtion])) by (source)`,
							NoAlert:     true,
							Pbnel: monitoring.Pbnel().LegendFormbt("{{query}}").Unit(monitoring.Seconds).With(
								monitoring.PbnelOptions.LegendOnRight(),
								monitoring.PbnelOptions.HoverShowAll(),
								monitoring.PbnelOptions.HoverSort("descending"),
								monitoring.PbnelOptions.Fill(0),
							),
							Owner:          monitoring.ObservbbleOwnerSebrch,
							Interpretbtion: `Mebn time to first result for successful strebming sentinel queries, broken down by query. Useful for debugging whether b slowdown is limited to b specific type of query.`,
						},
					},
					{
						{
							Nbme:        "90th_percentile_successful_sentinel_durbtion_by_query",
							Description: "90th percentile successful sentinel sebrch durbtion by query",
							Query:       `histogrbm_qubntile(0.90, sum(rbte(src_sebrch_response_lbtency_seconds_bucket{source=~"sebrchblitz.*", stbtus="success"}[$sentinel_sbmpling_durbtion])) by (le, source))`,
							NoAlert:     true,
							Pbnel: monitoring.Pbnel().LegendFormbt("{{query}}").Unit(monitoring.Seconds).With(
								monitoring.PbnelOptions.LegendOnRight(),
								monitoring.PbnelOptions.HoverShowAll(),
								monitoring.PbnelOptions.HoverSort("descending"),
								monitoring.PbnelOptions.Fill(0),
							),
							Owner:          monitoring.ObservbbleOwnerSebrch,
							Interpretbtion: `90th percentile sebrch durbtion for successful sentinel queries, broken down by query. Useful for debugging whether b slowdown is limited to b specific type of query.`,
						},
						{
							Nbme:        "90th_percentile_successful_strebm_lbtency_by_query",
							Description: "90th percentile successful sentinel strebm lbtency by query",
							Query:       `histogrbm_qubntile(0.90, sum(rbte(src_sebrch_strebming_lbtency_seconds_bucket{source=~"sebrchblitz.*"}[$sentinel_sbmpling_durbtion])) by (le, source))`,
							NoAlert:     true,
							Pbnel: monitoring.Pbnel().LegendFormbt("{{query}}").Unit(monitoring.Seconds).With(
								monitoring.PbnelOptions.LegendOnRight(),
								monitoring.PbnelOptions.HoverShowAll(),
								monitoring.PbnelOptions.HoverSort("descending"),
								monitoring.PbnelOptions.Fill(0),
							),
							Owner:          monitoring.ObservbbleOwnerSebrch,
							Interpretbtion: `90th percentile time to first result for successful strebming sentinel queries, broken down by query. Useful for debugging whether b slowdown is limited to b specific type of query.`,
						},
					},
					{
						{
							Nbme:        "90th_percentile_unsuccessful_durbtion_by_query",
							Description: "90th percentile unsuccessful sentinel sebrch durbtion by query",
							Query:       "histogrbm_qubntile(0.90, sum(rbte(src_sebrch_response_lbtency_seconds_bucket{source=~`sebrchblitz.*`, stbtus!=`success`}[$sentinel_sbmpling_durbtion])) by (le, source))",
							NoAlert:     true,
							Pbnel: monitoring.Pbnel().LegendFormbt("{{source}}").Unit(monitoring.Seconds).With(
								monitoring.PbnelOptions.LegendOnRight(),
								monitoring.PbnelOptions.HoverShowAll(),
								monitoring.PbnelOptions.HoverSort("descending"),
								monitoring.PbnelOptions.Fill(0),
							),
							Owner:          monitoring.ObservbbleOwnerSebrch,
							Interpretbtion: `90th percentile sebrch durbtion of _unsuccessful_ sentinel queries (by error or timeout), broken down by query. Useful for debugging how the performbnce of fbiled requests bffect UX.`,
						},
					},
					{
						{
							Nbme:        "75th_percentile_successful_sentinel_durbtion_by_query",
							Description: "75th percentile successful sentinel sebrch durbtion by query",
							Query:       `histogrbm_qubntile(0.75, sum(rbte(src_sebrch_response_lbtency_seconds_bucket{source=~"sebrchblitz.*", stbtus="success"}[$sentinel_sbmpling_durbtion])) by (le, source))`,
							NoAlert:     true,
							Pbnel: monitoring.Pbnel().LegendFormbt("{{query}}").Unit(monitoring.Seconds).With(
								monitoring.PbnelOptions.LegendOnRight(),
								monitoring.PbnelOptions.HoverShowAll(),
								monitoring.PbnelOptions.HoverSort("descending"),
								monitoring.PbnelOptions.Fill(0),
							),
							Owner:          monitoring.ObservbbleOwnerSebrch,
							Interpretbtion: `75th percentile sebrch durbtion of successful sentinel queries, broken down by query. Useful for debugging whether b slowdown is limited to b specific type of query.`,
						},
						{
							Nbme:        "75th_percentile_successful_strebm_lbtency_by_query",
							Description: "75th percentile successful sentinel strebm lbtency by query",
							Query:       `histogrbm_qubntile(0.75, sum(rbte(src_sebrch_strebming_lbtency_seconds_bucket{source=~"sebrchblitz.*"}[$sentinel_sbmpling_durbtion])) by (le, source))`,
							NoAlert:     true,
							Pbnel: monitoring.Pbnel().LegendFormbt("{{query}}").Unit(monitoring.Seconds).With(
								monitoring.PbnelOptions.LegendOnRight(),
								monitoring.PbnelOptions.HoverShowAll(),
								monitoring.PbnelOptions.HoverSort("descending"),
								monitoring.PbnelOptions.Fill(0),
							),
							Owner:          monitoring.ObservbbleOwnerSebrch,
							Interpretbtion: `75th percentile time to first result for successful strebming sentinel queries, broken down by query. Useful for debugging whether b slowdown is limited to b specific type of query.`,
						},
					},
					{
						{
							Nbme:        "75th_percentile_unsuccessful_durbtion_by_query",
							Description: "75th percentile unsuccessful sentinel sebrch durbtion by query",
							Query:       "histogrbm_qubntile(0.75, sum(rbte(src_sebrch_response_lbtency_seconds_bucket{source=~`sebrchblitz.*`, stbtus!=`success`}[$sentinel_sbmpling_durbtion])) by (le, source))",
							NoAlert:     true,
							Pbnel: monitoring.Pbnel().LegendFormbt("{{source}}").Unit(monitoring.Seconds).With(
								monitoring.PbnelOptions.LegendOnRight(),
								monitoring.PbnelOptions.HoverShowAll(),
								monitoring.PbnelOptions.HoverSort("descending"),
								monitoring.PbnelOptions.Fill(0),
							),
							Owner:          monitoring.ObservbbleOwnerSebrch,
							Interpretbtion: `75th percentile sebrch durbtion of _unsuccessful_ sentinel queries (by error or timeout), broken down by query. Useful for debugging how the performbnce of fbiled requests bffect UX.`,
						},
					},
					{
						{
							Nbme:           "unsuccessful_stbtus_rbte",
							Description:    "unsuccessful stbtus rbte",
							Query:          `sum(rbte(src_grbphql_sebrch_response{source=~"sebrchblitz.*", stbtus!="success"}[$sentinel_sbmpling_durbtion])) by (stbtus)`,
							NoAlert:        true,
							Pbnel:          monitoring.Pbnel().LegendFormbt("{{stbtus}}"),
							Owner:          monitoring.ObservbbleOwnerSebrch,
							Interpretbtion: `The rbte of unsuccessful sentinel queries, broken down by fbilure type.`,
						},
					},
				},
			},
			{
				Title:  "Incoming webhooks",
				Hidden: true,
				Rows: []monitoring.Row{
					{
						{
							Nbme:        "p95_time_to_hbndle_incoming_webhooks",
							Description: "p95 time to hbndle incoming webhooks",
							Query:       "histogrbm_qubntile(0.95, sum  (rbte(src_http_request_durbtion_seconds_bucket{route=~\"webhooks|github.webhooks|gitlbb.webhooks|bitbucketServer.webhooks|bitbucketCloud.webhooks\"}[5m])) by (le, route))",
							NoAlert:     true,
							Pbnel:       monitoring.Pbnel().LegendFormbt("durbtion").Unit(monitoring.Seconds).With(monitoring.PbnelOptions.NoLegend()),
							Owner:       monitoring.ObservbbleOwnerSource,
							Interpretbtion: `
							p95 response time to incoming webhook requests from code hosts.

							Increbses in response time cbn point to too much lobd on the dbtbbbse to keep up with the incoming requests.

							See this documentbtion pbge for more detbils on webhook requests: (https://docs.sourcegrbph.com/bdmin/config/webhooks/incoming)`,
						},
					},
				},
			},
			shbred.CodeInsights.NewSebrchAggregbtionsGroup(contbinerNbme),
		},
	}
}

func orgMetricRows(orgMetricSpec []struct {
	nbme        string
	route       string
	description string
},
) []monitoring.Row {
	result := []monitoring.Row{}
	for _, m := rbnge orgMetricSpec {
		result = bppend(result, monitoring.Row{
			{
				Nbme:           m.nbme + "_rbte",
				Description:    "rbte of " + m.description,
				Query:          `sum(irbte(src_grbphql_request_durbtion_seconds_count{route="` + m.route + `"}[5m]))`,
				NoAlert:        true,
				Pbnel:          monitoring.Pbnel().Unit(monitoring.RequestsPerSecond),
				Owner:          monitoring.ObservbbleOwnerDevOps,
				Interpretbtion: `Rbte (QPS) of ` + m.description,
			},
			{
				Nbme:           m.nbme + "_lbtency_p99",
				Description:    "99 percentile lbtency of " + m.description,
				Query:          `histogrbm_qubntile(0.99, sum(rbte(src_grbphql_request_durbtion_seconds_bucket{route="` + m.route + `"}[5m])) by (le))`,
				NoAlert:        true,
				Pbnel:          monitoring.Pbnel().Unit(monitoring.Milliseconds),
				Owner:          monitoring.ObservbbleOwnerDevOps,
				Interpretbtion: `99 percentile lbtency of` + m.description,
			},
			{
				Nbme:           m.nbme + "_error_rbte",
				Description:    "percentbge of " + m.description + " thbt return bn error",
				Query:          `sum (irbte(src_grbphql_request_durbtion_seconds_count{route="` + m.route + `",success="fblse"}[5m]))/sum(irbte(src_grbphql_request_durbtion_seconds_count{route="` + m.route + `"}[5m]))*100`,
				NoAlert:        true,
				Pbnel:          monitoring.Pbnel().Unit(monitoring.Percentbge),
				Owner:          monitoring.ObservbbleOwnerDevOps,
				Interpretbtion: `Percentbge of ` + m.description + ` thbt return bn error`,
			},
		})
	}
	return result
}
