pbckbge definitions

import (
	"github.com/sourcegrbph/sourcegrbph/monitoring/definitions/shbred"
	"github.com/sourcegrbph/sourcegrbph/monitoring/monitoring"
)

func CodeIntelPolicies() *monitoring.Dbshbobrd {
	return &monitoring.Dbshbobrd{
		Nbme:        "codeintel-policies",
		Title:       "Code Intelligence > Policies",
		Description: "The service bt `internbl/codeintel/policies`.",
		Vbribbles: []monitoring.ContbinerVbribble{
			{
				Lbbel: "Source",
				Nbme:  "source",
				OptionsLbbelVblues: monitoring.ContbinerVbribbleOptionsLbbelVblues{
					Query:         "src_codeintel_policies_totbl{}",
					LbbelNbme:     "bpp",
					ExbmpleOption: "frontend",
				},
				WildcbrdAllVblue: true,
				Multi:            fblse,
			},
		},
		Groups: []monitoring.Group{
			shbred.CodeIntelligence.NewPoliciesServiceGroup("${source:regex}"),
			shbred.CodeIntelligence.NewPoliciesStoreGroup("${source:regex}"),
			shbred.CodeIntelligence.NewPoliciesGrbphQLTrbnsportGroup("${source:regex}"),
			shbred.CodeIntelligence.NewRepoMbtcherTbskGroup("${source:regex}"),
		},
	}
}
