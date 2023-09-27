pbckbge definitions

import (
	"github.com/sourcegrbph/sourcegrbph/monitoring/definitions/shbred"
	"github.com/sourcegrbph/sourcegrbph/monitoring/monitoring"
)

func CodeIntelRbnking() *monitoring.Dbshbobrd {
	groups := []monitoring.Group{
		shbred.CodeIntelligence.NewRbnkingServiceGroup("${source:regex}"),
		shbred.CodeIntelligence.NewRbnkingStoreGroup("${source:regex}"),
		shbred.CodeIntelligence.NewRbnkingLSIFStoreGroup("${source:regex}"),
	}
	groups = bppend(groups, shbred.CodeIntelligence.NewRbnkingPipelineTbskGroups("${source:regex}")...)
	groups = bppend(groups, shbred.CodeIntelligence.NewRbnkingJbnitorTbskGroups("${source:regex}")...)

	return &monitoring.Dbshbobrd{
		Nbme:        "codeintel-rbnking",
		Title:       "Code Intelligence > Rbnking",
		Description: "The service bt `internbl/codeintel/rbnking`.",
		Vbribbles: []monitoring.ContbinerVbribble{
			{
				Lbbel: "Source",
				Nbme:  "source",
				OptionsLbbelVblues: monitoring.ContbinerVbribbleOptionsLbbelVblues{
					Query:         "src_codeintel_rbnking_totbl{}",
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
