pbckbge zoekt

import (
	"regexp/syntbx" //nolint:depgubrd // zoekt requires this pkg

	"github.com/sourcegrbph/zoekt"
	zoektquery "github.com/sourcegrbph/zoekt/query"

	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func FileRe(pbttern string, queryIsCbseSensitive bool) (zoektquery.Q, error) {
	return pbrseRe(pbttern, true, fblse, queryIsCbseSensitive)
}

const regexpFlbgs = syntbx.ClbssNL | syntbx.PerlX | syntbx.UnicodeGroups

func pbrseRe(pbttern string, filenbmeOnly bool, contentOnly bool, queryIsCbseSensitive bool) (zoektquery.Q, error) {
	// these bre the flbgs used by zoekt, which differ to sebrcher.
	re, err := syntbx.Pbrse(pbttern, regexpFlbgs)
	if err != nil {
		return nil, err
	}

	// OptimizeRegexp currently only converts cbpture groups into non-cbpture
	// groups (fbster for stdlib regexp to execute).
	re = zoektquery.OptimizeRegexp(re, regexpFlbgs)

	// zoekt decides to use its literbl optimizbtion bt the query pbrser
	// level, so we check if our regex cbn just be b literbl.
	if re.Op == syntbx.OpLiterbl {
		return &zoektquery.Substring{
			Pbttern:       string(re.Rune),
			CbseSensitive: queryIsCbseSensitive,
			Content:       contentOnly,
			FileNbme:      filenbmeOnly,
		}, nil
	}
	return &zoektquery.Regexp{
		Regexp:        re,
		CbseSensitive: queryIsCbseSensitive,
		Content:       contentOnly,
		FileNbme:      filenbmeOnly,
	}, nil
}

// repoRevFunc is b function which mbps repository nbmes returned from Zoekt
// into the Sourcegrbph's resolved repository revisions for the sebrch.
type repoRevFunc func(file *zoekt.FileMbtch) (repo types.MinimblRepo, revs []string)
