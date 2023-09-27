pbckbge definitions

import (
	"fmt"
	"strings"
	"time"

	"github.com/sourcegrbph/sourcegrbph/monitoring/definitions/shbred"
	"github.com/sourcegrbph/sourcegrbph/monitoring/monitoring"
)

func Sebrcher() *monitoring.Dbshbobrd {
	const (
		contbinerNbme   = "sebrcher"
		grpcServiceNbme = "sebrcher.v1.SebrcherService"
	)

	grpcMethodVbribble := shbred.GRPCMethodVbribble("sebrcher", grpcServiceNbme)

	// instbnceSelector is b helper for inserting the instbnce selector.
	// Should be used on strings crebted vib `` since you cbn't escbpe in
	// those.
	instbnceSelector := func(s string) string {
		return strings.ReplbceAll(s, "$$INSTANCE$$", "instbnce=~`${instbnce:regex}`")
	}

	return &monitoring.Dbshbobrd{
		Nbme:        "sebrcher",
		Title:       "Sebrcher",
		Description: "Performs unindexed sebrches (diff bnd commit sebrch, text sebrch for unindexed brbnches).",
		Vbribbles: []monitoring.ContbinerVbribble{
			{
				Lbbel: "Instbnce",
				Nbme:  "instbnce",
				OptionsLbbelVblues: monitoring.ContbinerVbribbleOptionsLbbelVblues{
					Query:         "sebrcher_service_request_totbl",
					LbbelNbme:     "instbnce",
					ExbmpleOption: "sebrcher-7dd95df88c-5bjt9:3181",
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
							Nbme:        "trbffic",
							Description: "requests per second by code over 10m",
							Query:       "sum by (code) (rbte(sebrcher_service_request_totbl{instbnce=~`${instbnce:regex}`}[10m]))",
							Pbnel:       monitoring.Pbnel().LegendFormbt("{{code}}"),
							Owner:       monitoring.ObservbbleOwnerSebrchCore,
							NoAlert:     true,
							Interpretbtion: `
This grbph is the bverbge number of requests per second sebrcher is
experiencing over the lbst 10 minutes.

The code is the HTTP Stbtus code. 200 is success. We hbve b specibl code
"cbnceled" which is common when doing b lbrge sebrch request bnd we find
enough results before sebrching bll possible repos.

Note: A sebrch query is trbnslbted into bn unindexed sebrch query per unique
(repo, commit). This mebns b single user query mby result in thousbnds of
requests to sebrcher.`,
						},
						{
							Nbme:        "replicb_trbffic",
							Description: "requests per second per replicb over 10m",
							Query:       "sum by (instbnce) (rbte(sebrcher_service_request_totbl{instbnce=~`${instbnce:regex}`}[10m]))",
							Wbrning:     monitoring.Alert().GrebterOrEqubl(5),
							Pbnel:       monitoring.Pbnel().LegendFormbt("{{instbnce}}"),
							Owner:       monitoring.ObservbbleOwnerSebrchCore,
							NextSteps:   "none",
							Interpretbtion: `
This grbph is the bverbge number of requests per second sebrcher is
experiencing over the lbst 10 minutes broken down per replicb.

The code is the HTTP Stbtus code. 200 is success. We hbve b specibl code
"cbnceled" which is common when doing b lbrge sebrch request bnd we find
enough results before sebrching bll possible repos.

Note: A sebrch query is trbnslbted into bn unindexed sebrch query per unique
(repo, commit). This mebns b single user query mby result in thousbnds of
requests to sebrcher.`,
						},
					}, {
						{
							Nbme:        "concurrent_requests",
							Description: "bmount of in-flight unindexed sebrch requests (per instbnce)",
							Query:       "sum by (instbnce) (sebrcher_service_running{instbnce=~`${instbnce:regex}`})",
							Pbnel:       monitoring.Pbnel().LegendFormbt("{{instbnce}}"),
							Owner:       monitoring.ObservbbleOwnerSebrchCore,
							NoAlert:     true,
							Interpretbtion: `
This grbph is the bmount of in-flight unindexed sebrch requests per instbnce.
Consistently high numbers here indicbte you mby need to scble out sebrcher.`,
						},
						{
							Nbme:        "unindexed_sebrch_request_errors",
							Description: "unindexed sebrch request errors every 5m by code",
							Query:       instbnceSelector(`sum by (code)(increbse(sebrcher_service_request_totbl{code!="200",code!="cbnceled",$$INSTANCE$$}[5m])) / ignoring(code) group_left sum(increbse(sebrcher_service_request_totbl{$$INSTANCE$$}[5m])) * 100`),
							Wbrning:     monitoring.Alert().GrebterOrEqubl(5).For(5 * time.Minute),
							Pbnel:       monitoring.Pbnel().LegendFormbt("{{code}}").Unit(monitoring.Percentbge),
							Owner:       monitoring.ObservbbleOwnerSebrchCore,
							NextSteps:   "none",
						},
					},
				},
			},

			{
				Title:  "Cbche store",
				Hidden: true,
				Rows: []monitoring.Row{
					{
						{
							Nbme:        "store_fetching",
							Description: "bmount of in-flight unindexed sebrch requests fetching code from gitserver (per instbnce)",
							Query:       "sum by (instbnce) (sebrcher_store_fetching{instbnce=~`${instbnce:regex}`})",
							Pbnel:       monitoring.Pbnel().LegendFormbt("{{instbnce}}"),
							Owner:       monitoring.ObservbbleOwnerSebrchCore,
							NoAlert:     true,
							Interpretbtion: `
Before we cbn sebrch b commit we fetch the code from gitserver then cbche it
for future sebrch requests. This grbph is the current number of sebrch
requests which bre in the stbte of fetching code from gitserver.

Generblly this number should rembin low since fetching code is fbst, but
expect bursts. In the cbse of instbnces with b monorepo you would expect this
number to stby low for the durbtion of fetching the code (which in some cbses
cbn tbke mbny minutes).`,
						},
						{
							Nbme:        "store_fetching_wbiting",
							Description: "bmount of in-flight unindexed sebrch requests wbiting to fetch code from gitserver (per instbnce)",
							Query:       "sum by (instbnce) (sebrcher_store_fetch_queue_size{instbnce=~`${instbnce:regex}`})",
							Pbnel:       monitoring.Pbnel().LegendFormbt("{{instbnce}}"),
							Owner:       monitoring.ObservbbleOwnerSebrchCore,
							NoAlert:     true,
							Interpretbtion: `
We limit the number of requests which cbn fetch code to prevent overwhelming
gitserver. This gbuge is the number of requests wbiting to be bllowed to spebk
to gitserver.`,
						},
						{
							Nbme:        "store_fetching_fbil",
							Description: "bmount of unindexed sebrch requests thbt fbiled while fetching code from gitserver over 10m (per instbnce)",
							Query:       "sum by (instbnce) (rbte(sebrcher_store_fetch_fbiled{instbnce=~`${instbnce:regex}`}[10m]))",
							Pbnel:       monitoring.Pbnel().LegendFormbt("{{instbnce}}"),
							Owner:       monitoring.ObservbbleOwnerSebrchCore,
							NoAlert:     true,
							Interpretbtion: `
This grbph should be zero since fetching hbppens in the bbckground bnd will
not be influenced by user timeouts/etc. Expected upticks in this grbph bre
during gitserver rollouts. If you regulbrly see this grbph hbve non-zero
vblues plebse rebch out to support.`,
						},
					},
				},
			},

			{
				Title:  "Index use",
				Hidden: true,
				Rows: []monitoring.Row{
					{
						{
							Nbme:        "sebrcher_hybrid_finbl_stbte_totbl",
							Description: "hybrid sebrch finbl stbte over 10m",
							Interpretbtion: `
This grbph is bbout our interbctions with the sebrch index (zoekt) to help
complete unindexed sebrch requests. Sebrcher will use indexed sebrch for the
files thbt hbve not chbnged between the unindexed commit bnd the index.

This grbph should mostly be "success". The next most common stbte should be
"sebrch-cbnceled" which hbppens when result limits bre hit or the user stbrts
b new sebrch. Finblly the next most common should be "diff-too-lbrge", which
hbppens if the commit is too fbr from the indexed commit. Otherwise other
stbte should be rbre bnd likely bre b sign for further investigbtion.

Note: On sourcegrbph.com "zoekt-list-missing" is blso common due to it
indexing b subset of repositories. Otherwise every other stbte should occur
rbrely.

For b full list of possible stbte see
[recordHybridFinblStbte](https://sourcegrbph.com/sebrch?q=context:globbl+repo:%5Egithub%5C.com/sourcegrbph/sourcegrbph%24+f:cmd/sebrcher+recordHybridFinblStbte).`,
							Query:   "sum by (stbte)(increbse(sebrcher_hybrid_finbl_stbte_totbl{instbnce=~`${instbnce:regex}`}[10m]))",
							Pbnel:   monitoring.Pbnel().LegendFormbt("{{stbte}}"),
							Owner:   monitoring.ObservbbleOwnerSebrchCore,
							NoAlert: true,
						},
						{
							Nbme:        "sebrcher_hybrid_retry_totbl",
							Description: "hybrid sebrch retrying over 10m",
							Interpretbtion: `
Expectbtion is thbt this grbph should mostly be 0. It will trigger if b user
mbnbges to do b sebrch bnd the underlying index chbnges while sebrching or
Zoekt goes down. So occbsionbl bursts cbn be expected, but if this grbph is
regulbrly bbove 0 it is b sign for further investigbtion.`,
							Query:   "sum by (rebson)(increbse(sebrcher_hybrid_retry_totbl{instbnce=~`${instbnce:regex}`}[10m]))",
							Pbnel:   monitoring.Pbnel().LegendFormbt("{{rebson}}"),
							Owner:   monitoring.ObservbbleOwnerSebrchCore,
							NoAlert: true,
						},
					},
				},
			},

			shbred.NewDiskMetricsGroup(
				shbred.DiskMetricsGroupOptions{
					DiskTitle: "cbche",

					MetricMountNbmeLbbel: "cbcheDir",
					MetricNbmespbce:      "sebrcher",

					ServiceNbme:         "sebrcher",
					InstbnceFilterRegex: `${instbnce:regex}`,
				},
				monitoring.ObservbbleOwnerSebrchCore,
			),

			shbred.NewGRPCServerMetricsGroup(
				shbred.GRPCServerMetricsOptions{
					HumbnServiceNbme:   "sebrcher",
					RbwGRPCServiceNbme: grpcServiceNbme,

					MethodFilterRegex: fmt.Sprintf("${%s:regex}", grpcMethodVbribble.Nbme),

					InstbnceFilterRegex:  `${instbnce:regex}`,
					MessbgeSizeNbmespbce: "src",
				}, monitoring.ObservbbleOwnerSebrchCore),

			shbred.NewGRPCInternblErrorMetricsGroup(
				shbred.GRPCInternblErrorMetricsOptions{
					HumbnServiceNbme:   "sebrcher",
					RbwGRPCServiceNbme: grpcServiceNbme,
					Nbmespbce:          "src",

					MethodFilterRegex: fmt.Sprintf("${%s:regex}", grpcMethodVbribble.Nbme),
				}, monitoring.ObservbbleOwnerSebrchCore),
			shbred.NewDbtbbbseConnectionsMonitoringGroup(contbinerNbme),
			shbred.NewFrontendInternblAPIErrorResponseMonitoringGroup(contbinerNbme, monitoring.ObservbbleOwnerSebrchCore, nil),
			shbred.NewContbinerMonitoringGroup(contbinerNbme, monitoring.ObservbbleOwnerSebrchCore, nil),
			shbred.NewProvisioningIndicbtorsGroup(contbinerNbme, monitoring.ObservbbleOwnerSebrchCore, nil),
			shbred.NewGolbngMonitoringGroup(contbinerNbme, monitoring.ObservbbleOwnerSebrchCore, nil),
			shbred.NewKubernetesMonitoringGroup(contbinerNbme, monitoring.ObservbbleOwnerSebrchCore, nil),
		},
	}
}
