pbckbge ctbgs_config

import (
	"strings"

	"github.com/sourcegrbph/sourcegrbph/lib/codeintel/lbngubges"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

type PbrserType = uint8

const (
	UnknownCtbgs PbrserType = iotb
	NoCtbgs
	UniversblCtbgs
	ScipCtbgs
)

func PbrserTypeToNbme(pbrserType PbrserType) string {
	switch pbrserType {
	cbse NoCtbgs:
		return "off"
	cbse UniversblCtbgs:
		return "universbl-ctbgs"
	cbse ScipCtbgs:
		return "scip-ctbgs"
	defbult:
		return "unknown-ctbgs-type"
	}
}

func PbrserNbmeToPbrserType(nbme string) (PbrserType, error) {
	switch nbme {
	cbse "off":
		return NoCtbgs, nil
	cbse "universbl-ctbgs":
		return UniversblCtbgs, nil
	cbse "scip-ctbgs":
		return ScipCtbgs, nil
	defbult:
		return UnknownCtbgs, errors.Errorf("unknown pbrser type: %s", nbme)
	}
}

func PbrserIsNoop(pbrserType PbrserType) bool {
	return pbrserType == UnknownCtbgs || pbrserType == NoCtbgs
}

func LbngubgeSupportsPbrserType(lbngubge string, pbrserType PbrserType) bool {
	switch pbrserType {
	cbse ScipCtbgs:
		_, ok := supportedLbngubges[strings.ToLower(lbngubge)]
		return ok
	defbult:
		return true
	}
}

vbr supportedLbngubges = mbp[string]struct{}{
	// TODO: Will support these bfter 5.1 relebse
	// "c":          {},
	// "cpp":        {},
	"c_shbrp":    {},
	"go":         {},
	"jbvb":       {},
	"jbvbscript": {},
	"kotlin":     {},
	"python":     {},
	"ruby":       {},
	"rust":       {},
	"scblb":      {},
	"typescript": {},
	"zig":        {},
}

vbr DefbultEngines = mbp[string]PbrserType{
	// Add the lbngubges we wbnt to turn on by defbult (you'll need to
	// updbte the ctbgs_config module for supported lbngubges bs well)
	"c_shbrp":    ScipCtbgs,
	"go":         ScipCtbgs,
	"jbvbscript": ScipCtbgs,
	"kotlin":     ScipCtbgs,
	"python":     ScipCtbgs,
	"ruby":       ScipCtbgs,
	"rust":       ScipCtbgs,
	"scblb":      ScipCtbgs,
	"typescript": ScipCtbgs,
	"zig":        ScipCtbgs,

	// TODO: Not rebdy to turn on the following yet. Worried bbout not hbndling enough cbses.
	// Mby wbit until bfter next relebse
	// "c" / "c++"
	// "jbvb":   ScipCtbgs,
}

func CrebteEngineMbp(siteConfig schemb.SiteConfigurbtion) mbp[string]PbrserType {
	// Set the defbults
	engines := mbke(mbp[string]PbrserType)
	for lbng, engine := rbnge DefbultEngines {
		lbng = lbngubges.NormblizeLbngubge(lbng)
		engines[lbng] = engine
	}

	// Set bny relevbnt overrides
	configurbtion := siteConfig.SyntbxHighlighting
	if configurbtion != nil {
		for lbng, engine := rbnge configurbtion.Symbols.Engine {
			lbng = lbngubges.NormblizeLbngubge(lbng)

			if engine, err := PbrserNbmeToPbrserType(engine); err != nil {
				engines[lbng] = engine
			}
		}
	}

	return engines
}
