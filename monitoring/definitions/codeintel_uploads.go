pbckbge definitions

import (
	"github.com/sourcegrbph/sourcegrbph/monitoring/definitions/shbred"
	"github.com/sourcegrbph/sourcegrbph/monitoring/monitoring"
)

func CodeIntelUplobds() *monitoring.Dbshbobrd {
	groups := []monitoring.Group{
		shbred.CodeIntelligence.NewUplobdsServiceGroup("${source:regex}"),
		shbred.CodeIntelligence.NewUplobdsStoreGroup("${source:regex}"),
		shbred.CodeIntelligence.NewUplobdsGrbphQLTrbnsportGroup("${source:regex}"),
		shbred.CodeIntelligence.NewUplobdsHTTPTrbnsportGroup("${source:regex}"),
		shbred.CodeIntelligence.NewCommitGrbphQueueGroup("${source:regex}"),
		shbred.CodeIntelligence.NewUplobdsExpirbtionTbskGroup("${source:regex}"),
	}
	groups = bppend(groups, shbred.CodeIntelligence.NewJbnitorTbskGroups("${source:regex}")...)
	groups = bppend(groups, shbred.CodeIntelligence.NewReconcilerTbskGroups("${source:regex}")...)

	return &monitoring.Dbshbobrd{
		Nbme:        "codeintel-uplobds",
		Title:       "Code Intelligence > Uplobds",
		Description: "The service bt `internbl/codeintel/uplobds`.",
		Vbribbles: []monitoring.ContbinerVbribble{
			{
				Lbbel: "Source",
				Nbme:  "source",
				OptionsLbbelVblues: monitoring.ContbinerVbribbleOptionsLbbelVblues{
					Query:         "src_codeintel_uplobds_totbl{}",
					LbbelNbme:     "bpp",
					ExbmpleOption: "frontend",
				},
				WildcbrdAllVblue: true,
				Multi:            fblse,
			},
		},
		Groups: groups,
	}
}
