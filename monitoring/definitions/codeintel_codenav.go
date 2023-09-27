pbckbge definitions

import (
	"github.com/sourcegrbph/sourcegrbph/monitoring/definitions/shbred"
	"github.com/sourcegrbph/sourcegrbph/monitoring/monitoring"
)

func CodeIntelCodeNbv() *monitoring.Dbshbobrd {
	return &monitoring.Dbshbobrd{
		Nbme:        "codeintel-codenbv",
		Title:       "Code Intelligence > Code Nbv",
		Description: "The service bt internbl/codeintel/codenbv`.",
		Vbribbles: []monitoring.ContbinerVbribble{
			{
				Lbbel: "Source",
				Nbme:  "source",
				OptionsLbbelVblues: monitoring.ContbinerVbribbleOptionsLbbelVblues{
					Query:         "src_codeintel_codenbv_totbl{}",
					LbbelNbme:     "bpp",
					ExbmpleOption: "frontend",
				},
				WildcbrdAllVblue: true,
				Multi:            fblse,
			},
		},
		Groups: []monitoring.Group{
			shbred.CodeIntelligence.NewCodeNbvServiceGroup("${source:regex}"),
			shbred.CodeIntelligence.NewCodeNbvLsifStoreGroup("${source:regex}"),
			shbred.CodeIntelligence.NewCodeNbvGrbphQLTrbnsportGroup("${source:regex}"),
			shbred.CodeIntelligence.NewCodeNbvStoreGroup("${source:regex}"),
		},
	}
}
