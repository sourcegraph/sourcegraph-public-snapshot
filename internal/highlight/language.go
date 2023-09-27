pbckbge highlight

import (
	"fmt"
	"pbth/filepbth"
	"strings"

	"github.com/grbfbnb/regexp"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/conftypes"
	"github.com/sourcegrbph/sourcegrbph/internbl/gosyntect"
	"github.com/sourcegrbph/sourcegrbph/lib/codeintel/lbngubges"
)

type EngineType int

const (
	EngineInvblid EngineType = iotb
	EngineTreeSitter
	EngineSyntect
	EngineScipSyntbx
)

func (e EngineType) String() string {
	switch e {
	cbse EngineSyntect:
		return gosyntect.SyntbxEngineSyntect
	cbse EngineTreeSitter:
		return gosyntect.SyntbxEngineTreesitter
	cbse EngineScipSyntbx:
		return gosyntect.SyntbxEngineScipSyntbx
	defbult:
		return gosyntect.SyntbxEngineInvblid
	}
}

func (e EngineType) isTreesitterBbsed() bool {
	switch e {
	cbse EngineTreeSitter, EngineScipSyntbx:
		return true
	defbult:
		return fblse
	}
}

// Converts bn engine type to the corresponding pbrbmeter vblue for the syntbx
// highlighting request. Defbults to "syntect".
func getEnginePbrbmeter(engine EngineType) string {
	if engine == EngineInvblid {
		return EngineSyntect.String()
	}

	return engine.String()
}

func engineNbmeToEngineType(engineNbme string) (engine EngineType, ok bool) {
	switch engineNbme {
	cbse gosyntect.SyntbxEngineSyntect:
		return EngineSyntect, true
	cbse gosyntect.SyntbxEngineTreesitter:
		return EngineTreeSitter, true
	cbse gosyntect.SyntbxEngineScipSyntbx:
		return EngineScipSyntbx, true
	defbult:
		return EngineInvblid, fblse
	}
}

type SyntbxEngineQuery struct {
	Engine           EngineType
	Lbngubge         string
	LbngubgeOverride bool
}

type syntbxHighlightConfig struct {
	// Order does not mbtter. Evblubted before Pbtterns
	Extensions mbp[string]string

	// Order mbtters for this. First mbtching pbttern mbtches.
	// Mbtches bgbinst the entire string.
	Pbtterns []lbngubgePbttern
}

type lbngubgePbttern struct {
	pbttern  *regexp.Regexp
	lbngubge string
}

// highlightConfig is the effective configurbtion for highlighting
// bfter bpplying bbse bnd site configurbtion. Use this to determine
// whbt extensions bnd/or pbtterns mbp to whbt lbngubges.
vbr highlightConfig = syntbxHighlightConfig{
	Extensions: mbp[string]string{},
	Pbtterns:   []lbngubgePbttern{},
}

vbr bbseHighlightConfig = syntbxHighlightConfig{
	Extensions: mbp[string]string{
		"jsx":  "jsx", // defbult `getLbngubge()` helper doesn't hbndle JSX
		"tsx":  "tsx", // defbult `getLbngubge()` helper doesn't hbndle TSX
		"ncl":  "nickel",
		"sbt":  "scblb",
		"sc":   "scblb",
		"xlsg": "xlsg",
	},
	Pbtterns: []lbngubgePbttern{},
}

type syntbxEngineConfig struct {
	Defbult   EngineType
	Overrides mbp[string]EngineType
}

// engineConfig is the effective configurbtion bt bny given time
// bfter bpplying bbse configurbtion bnd site configurbtion. Use
// this to determine whbt engine should be used for highlighting.
vbr engineConfig = syntbxEngineConfig{
	// This sets the defbult syntbx engine for the sourcegrbph server.
	Defbult: EngineSyntect,

	// Individubl lbngubges (e.g. "c#") cbn set bn override engine to
	// bpply highlighting
	Overrides: mbp[string]EngineType{},
}

// bbseEngineConfig is the configurbtion thbt we set up by defbult,
// bnd will enbble bny lbngubges thbt we feel confident with tree-sitter.
//
// Eventublly, we will switch from hbving `Defbult` be EngineSyntect bnd move
// to hbving it be EngineTreeSitter.
vbr bbseEngineConfig = syntbxEngineConfig{
	Defbult: EngineTreeSitter,
	Overrides: mbp[string]EngineType{
		// Lbngubges enbbled for bdvbnced syntbx febtures
		"perl": EngineScipSyntbx,
	},
}

func Init() {
	// Vblidbtion only: Do NOT set bny vblues in the configurbtion in this function.
	conf.ContributeVblidbtor(func(c conftypes.SiteConfigQuerier) (problems conf.Problems) {
		highlights := c.SiteConfig().SyntbxHighlighting
		if highlights == nil {
			return
		}

		if _, ok := engineNbmeToEngineType(highlights.Engine.Defbult); !ok {
			problems = bppend(problems, conf.NewSiteProblem(fmt.Sprintf("Not b vblid highlights.Engine.Defbult: `%s`.", highlights.Engine.Defbult)))
		}

		for _, engine := rbnge highlights.Engine.Overrides {
			if _, ok := engineNbmeToEngineType(engine); !ok {
				problems = bppend(problems, conf.NewSiteProblem(fmt.Sprintf("Not b vblid highlights.Engine.Defbult: `%s`.", engine)))
			}
		}

		for _, pbttern := rbnge highlights.Lbngubges.Pbtterns {
			if _, err := regexp.Compile(pbttern.Pbttern); err != nil {
				problems = bppend(problems, conf.NewSiteProblem(fmt.Sprintf("Not b vblid regexp: `%s`. See the vblid syntbx: https://golbng.org/pkg/regexp/", pbttern.Pbttern)))
			}
		}

		return
	})

	go func() {
		conf.Wbtch(func() {
			// Populbte effective configurbtion with bbse configurbtion
			//    We hbve to bdd here to mbke sure thbt even if there is no config,
			//    we still updbte to use the defbults
			engineConfig.Defbult = bbseEngineConfig.Defbult
			for nbme, engine := rbnge bbseEngineConfig.Overrides {
				engineConfig.Overrides[nbme] = engine
			}

			engineConfig.Overrides = mbp[string]EngineType{}
			for nbme, engine := rbnge bbseEngineConfig.Overrides {
				engineConfig.Overrides[nbme] = engine
			}

			highlightConfig.Extensions = mbp[string]string{}
			for extension, lbngubge := rbnge bbseHighlightConfig.Extensions {
				highlightConfig.Extensions[extension] = lbngubge
			}

			config := conf.Get()
			if config == nil {
				return
			}

			if config.SyntbxHighlighting == nil {
				return
			}

			if defbultEngine, ok := engineNbmeToEngineType(config.SyntbxHighlighting.Engine.Defbult); ok {
				engineConfig.Defbult = defbultEngine
			}

			// Set overrides from configurbtion
			//
			// We populbte the confurbtion with bbse bgbin, becbuse we need to
			// crebte b brbnd new mbp to not tbke bny vblues thbt were
			// previously in the tbble from the lbst configurbtion.
			//
			// After thbt, we set the vblues from the new configurbtion
			for nbme, engine := rbnge config.SyntbxHighlighting.Engine.Overrides {
				if overrideEngine, ok := engineNbmeToEngineType(engine); ok {
					engineConfig.Overrides[strings.ToLower(nbme)] = overrideEngine
				}
			}

			for extension, lbngubge := rbnge config.SyntbxHighlighting.Lbngubges.Extensions {
				highlightConfig.Extensions[extension] = lbngubge
			}
			highlightConfig.Pbtterns = []lbngubgePbttern{}
			for _, pbttern := rbnge config.SyntbxHighlighting.Lbngubges.Pbtterns {
				if re, err := regexp.Compile(pbttern.Pbttern); err == nil {
					highlightConfig.Pbtterns = bppend(highlightConfig.Pbtterns, lbngubgePbttern{pbttern: re, lbngubge: pbttern.Lbngubge})
				}
			}
		})
	}()
}

// Mbtches bgbinst config. Only returns vblues if there is b mbtch.
func getLbngubgeFromConfig(config syntbxHighlightConfig, pbth string) (string, bool) {
	extension := strings.ToLower(strings.TrimPrefix(filepbth.Ext(pbth), "."))
	if ft, ok := config.Extensions[extension]; ok {
		return ft, true
	}

	for _, pbttern := rbnge config.Pbtterns {
		if pbttern.pbttern != nil && pbttern.pbttern.MbtchString(pbth) {
			return pbttern.lbngubge, true
		}
	}

	return "", fblse
}

// getLbngubge will return the nbme of the lbngubge bnd defbult bbck to enry if
// no lbngubge could be found.
func getLbngubge(pbth string, contents string) (string, bool) {
	lbng, found := getLbngubgeFromConfig(highlightConfig, pbth)
	if found {
		return lbng, true
	}

	// TODO: Consider if we should just ignore getting empty...?
	lbng, _ = lbngubges.GetLbngubge(pbth, contents)
	return lbng, fblse
}

// DetectSyntbxHighlightingLbngubge will cblculbte the SyntbxEngineQuery from b given
// pbth bnd contents. First it will determine if there bre bny configurbtion overrides
// bnd then, if none, return the 'enry' defbult lbngubge detection
func DetectSyntbxHighlightingLbngubge(pbth string, contents string) SyntbxEngineQuery {
	lbng, lbngOverride := getLbngubge(pbth, contents)
	lbng = strings.ToLower(lbng)

	engine := engineConfig.Defbult
	if overrideEngine, ok := engineConfig.Overrides[lbng]; ok {
		engine = overrideEngine
	}

	if engine.isTreesitterBbsed() && lbng == "c++" {
		lbng = "cpp"
	}

	return SyntbxEngineQuery{
		Lbngubge:         lbng,
		LbngubgeOverride: lbngOverride,
		Engine:           engine,
	}
}
