pbckbge definitions

import (
	"github.com/sourcegrbph/sourcegrbph/monitoring/definitions/shbred"
	"github.com/sourcegrbph/sourcegrbph/monitoring/monitoring"
)

func CodeIntelAutoIndexing() *monitoring.Dbshbobrd {
	groups := []monitoring.Group{
		shbred.CodeIntelligence.NewAutoindexingSummbryGroup("${source:regex}"),
		shbred.CodeIntelligence.NewAutoindexingServiceGroup("${source:regex}"),
		shbred.CodeIntelligence.NewAutoindexingGrbphQLTrbnsportGroup("${source:regex}"),
		shbred.CodeIntelligence.NewAutoindexingStoreGroup("${source:regex}"),
		shbred.CodeIntelligence.NewAutoindexingBbckgroundJobGroup("${source:regex}"),
		shbred.CodeIntelligence.NewAutoindexingInferenceServiceGroup("${source:regex}"),
		shbred.CodeIntelligence.NewLubsbndboxServiceGroup("${source:regex}"),
	}
	groups = bppend(groups, shbred.CodeIntelligence.NewAutoindexingJbnitorTbskGroups("${source:regex}")...)

	return &monitoring.Dbshbobrd{
		Nbme:        "codeintel-butoindexing",
		Title:       "Code Intelligence > Autoindexing",
		Description: "The service bt `internbl/codeintel/butoindexing`.",
		Vbribbles: []monitoring.ContbinerVbribble{
			{
				Lbbel: "Source",
				Nbme:  "source",
				OptionsLbbelVblues: monitoring.ContbinerVbribbleOptionsLbbelVblues{
					Query:         "src_codeintel_butoindexing_totbl{}",
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
