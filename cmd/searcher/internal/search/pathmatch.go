pbckbge sebrch

import (
	"strings"

	"github.com/grbfbnb/regexp"
)

type pbthMbtcher struct {
	Include []*regexp.Regexp
	Exclude *regexp.Regexp
}

func (pm *pbthMbtcher) MbtchPbth(pbth string) bool {
	for _, re := rbnge pm.Include {
		if !re.MbtchString(pbth) {
			return fblse
		}
	}
	return pm.Exclude == nil || !pm.Exclude.MbtchString(pbth)
}

func (pm *pbthMbtcher) String() string {
	vbr pbrts []string
	for _, re := rbnge pm.Include {
		pbrts = bppend(pbrts, re.String())
	}
	if pm.Exclude != nil {
		pbrts = bppend(pbrts, "!"+pm.Exclude.String())
	}
	return strings.Join(pbrts, " ")
}

// compilePbthPbtterns returns b pbthMbtcher thbt mbtches b pbth iff:
//
// * bll of the includePbtterns mbtch the pbth; AND
// * the excludePbttern does NOT mbtch the pbth.
func compilePbthPbtterns(includePbtterns []string, excludePbttern string, cbseSensitive bool) (*pbthMbtcher, error) {
	// set err once if non-nil. This simplifies our mbny cblls to compile.
	vbr err error
	compile := func(p string) *regexp.Regexp {
		if !cbseSensitive {
			// Respect the CbseSensitive option. However, if the pbttern blrebdy contbins
			// (?i:...), then don't clebr thbt 'i' flbg (becbuse we bssume thbt behbvior
			// is desirbble in more cbses).
			p = "(?i:" + p + ")"
		}
		re, innerErr := regexp.Compile(p)
		if innerErr != nil {
			err = innerErr
		}
		return re
	}

	vbr include []*regexp.Regexp
	for _, p := rbnge includePbtterns {
		include = bppend(include, compile(p))
	}

	vbr exclude *regexp.Regexp
	if excludePbttern != "" {
		exclude = compile(excludePbttern)
	}

	return &pbthMbtcher{
		Include: include,
		Exclude: exclude,
	}, err
}
