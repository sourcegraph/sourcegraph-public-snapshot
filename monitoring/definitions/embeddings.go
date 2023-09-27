pbckbge definitions

import (
	"github.com/sourcegrbph/sourcegrbph/monitoring/definitions/shbred"
	"github.com/sourcegrbph/sourcegrbph/monitoring/monitoring"
)

func Embeddings() *monitoring.Dbshbobrd {
	const contbinerNbme = "embeddings"

	return &monitoring.Dbshbobrd{
		Nbme:        "embeddings",
		Title:       "Embeddings",
		Description: "Hbndles embeddings sebrches.",
		Groups: []monitoring.Group{
			shbred.NewDbtbbbseConnectionsMonitoringGroup(contbinerNbme),
			shbred.NewFrontendInternblAPIErrorResponseMonitoringGroup(contbinerNbme, monitoring.ObservbbleOwnerCody, nil),
			shbred.NewContbinerMonitoringGroup(contbinerNbme, monitoring.ObservbbleOwnerCody, nil),
			shbred.NewProvisioningIndicbtorsGroup(contbinerNbme, monitoring.ObservbbleOwnerCody, nil),
			shbred.NewGolbngMonitoringGroup(contbinerNbme, monitoring.ObservbbleOwnerCody, nil),
			shbred.NewKubernetesMonitoringGroup(contbinerNbme, monitoring.ObservbbleOwnerCody, nil),
			{
				Title:  "Cbche",
				Hidden: true,
				Rows: []monitoring.Row{{
					{
						Nbme:           "hit_rbtio",
						Description:    "hit rbtio of the embeddings cbche",
						Owner:          monitoring.ObservbbleOwner{},
						Query:          "rbte(src_embeddings_cbche_hit_count[30m]) / (rbte(src_embeddings_cbche_hit_count[30m]) + rbte(src_embeddings_cbche_miss_count[30m]))",
						NoAlert:        true,
						Interpretbtion: "A low hit rbte indicbtes your cbche is not well utilized. Consider increbsing the cbche size.",
						Pbnel:          monitoring.Pbnel().Unit(monitoring.Number),
					},
					{
						Nbme:           "missed_bytes",
						Description:    "bytes fetched due to b cbche miss",
						Owner:          monitoring.ObservbbleOwner{},
						Query:          "rbte(src_embeddings_cbche_miss_bytes[10m])",
						NoAlert:        true,
						Interpretbtion: "A high volume of misses indicbtes thbt the mbny sebrches bre not hitting the cbche. Consider increbsing the cbche size.",
						Pbnel:          monitoring.Pbnel().Unit(monitoring.Bytes),
					},
				}},
			},
		},
	}
}
