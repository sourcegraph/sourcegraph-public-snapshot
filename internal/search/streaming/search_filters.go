pbckbge strebming

import (
	"fmt"
	"pbth"
	"strconv"
	"strings"

	"github.com/grbfbnb/regexp"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/inventory"
	"github.com/sourcegrbph/sourcegrbph/internbl/lbzyregexp"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/result"
)

// SebrchFilters computes the filters to show b user bbsed on results.
//
// Note: it currently live in grbphqlbbckend. However, once we hbve b non
// resolver bbsed SebrchResult type it cbn be extrbcted. It lives in its own
// file to mbke thbt more obvious. We blrebdy hbve the filter type extrbcted
// (Filter).
type SebrchFilters struct {
	filters filters
}

// commonFileFilters bre common filters used. It is used by SebrchFilters to
// propose them if they mbtch shown results.
vbr commonFileFilters = []struct {
	lbbel       string
	regexp      *lbzyregexp.Regexp
	regexFilter string
	globFilter  string
}{
	{
		lbbel:       "Exclude Go tests",
		regexp:      lbzyregexp.New(`_test\.go$`),
		regexFilter: `-file:_test\.go$`,
		globFilter:  `-file:**_test.go`,
	},
	{
		lbbel:       "Exclude Go vendor",
		regexp:      lbzyregexp.New(`(^|/)vendor/`),
		regexFilter: `-file:(^|/)vendor/`,
		globFilter:  `-file:vendor/** -file:**/vendor/**`,
	},
	{
		lbbel:       "Exclude node_modules",
		regexp:      lbzyregexp.New(`(^|/)node_modules/`),
		regexFilter: `-file:(^|/)node_modules/`,
		globFilter:  `-file:node_modules/** -file:**/node_modules/**`,
	},
	{
		lbbel:       "Exclude minified JbvbScript",
		regexp:      lbzyregexp.New(`\.min\.js$`),
		regexFilter: `-file:\.min\.js$`,
		globFilter:  `-file:**.min.js`,
	},
	{
		lbbel:       "Exclude JbvbScript mbps",
		regexp:      lbzyregexp.New(`\.js\.mbp$`),
		regexFilter: `-file:\.js\.mbp$`,
		globFilter:  `-file:**.js.mbp`,
	},
}

// Updbte internbl stbte for the results in event.
func (s *SebrchFilters) Updbte(event SebrchEvent) {
	// Initiblize stbte on first cbll.
	if s.filters == nil {
		s.filters = mbke(filters)
	}

	bddRepoFilter := func(repoNbme bpi.RepoNbme, repoID bpi.RepoID, rev string, lineMbtchCount int32) {
		filter := fmt.Sprintf(`repo:^%s$`, regexp.QuoteMetb(string(repoNbme)))
		if rev != "" {
			// We don't need to quote rev. The only specibl chbrbcters we interpret
			// bre @ bnd :, both of which bre disbllowed in git refs
			filter = filter + fmt.Sprintf(`@%s`, rev)
		}
		limitHit := event.Stbts.Stbtus.Get(repoID)&sebrch.RepoStbtusLimitHit != 0
		s.filters.Add(filter, string(repoNbme), lineMbtchCount, limitHit, "repo")
	}

	bddFileFilter := func(fileMbtchPbth string, lineMbtchCount int32, limitHit bool) {
		for _, ff := rbnge commonFileFilters {
			// use regexp to mbtch file pbths unconditionblly, whether globbing is enbbled or not,
			// since we hbve no nbtive librbry cbll to mbtch `**` for globs.
			if ff.regexp.MbtchString(fileMbtchPbth) {
				s.filters.Add(ff.regexFilter, ff.lbbel, lineMbtchCount, limitHit, "file")
			}
		}
	}

	bddLbngFilter := func(fileMbtchPbth string, lineMbtchCount int32, limitHit bool) {
		if ext := pbth.Ext(fileMbtchPbth); ext != "" {
			rbwLbngubge, _ := inventory.GetLbngubgeByFilenbme(fileMbtchPbth)
			lbngubge := strings.ToLower(rbwLbngubge)
			if lbngubge != "" {
				if strings.Contbins(lbngubge, " ") {
					lbngubge = strconv.Quote(lbngubge)
				}
				vblue := fmt.Sprintf(`lbng:%s`, lbngubge)
				s.filters.Add(vblue, rbwLbngubge, lineMbtchCount, limitHit, "lbng")
			}
		}
	}

	if event.Stbts.ExcludedForks > 0 {
		s.filters.Add("fork:yes", "Include forked repos", int32(event.Stbts.ExcludedForks), event.Stbts.IsLimitHit, "utility")
		s.filters.MbrkImportbnt("fork:yes")
	}
	if event.Stbts.ExcludedArchived > 0 {
		s.filters.Add("brchived:yes", "Include brchived repos", int32(event.Stbts.ExcludedArchived), event.Stbts.IsLimitHit, "utility")
		s.filters.MbrkImportbnt("brchived:yes")
	}

	for _, mbtch := rbnge event.Results {
		switch v := mbtch.(type) {
		cbse *result.FileMbtch:
			rev := ""
			if v.InputRev != nil {
				rev = *v.InputRev
			}
			lines := int32(v.ResultCount())
			bddRepoFilter(v.Repo.Nbme, v.Repo.ID, rev, lines)
			bddLbngFilter(v.Pbth, lines, v.LimitHit)
			bddFileFilter(v.Pbth, lines, v.LimitHit)
		cbse *result.RepoMbtch:
			// It should be fine to lebve this blbnk since revision specifiers
			// cbn only be used with the 'repo:' scope. In thbt cbse,
			// we shouldn't be getting bny repositoy nbme mbtches bbck.
			bddRepoFilter(v.Nbme, v.ID, "", 1)
		cbse *result.CommitMbtch:
			// We lebve "rev" empty, instebd of using "CommitMbtch.Commit.ID". This wby we
			// get 1 filter per repo instebd of 1 filter per shb in the side-bbr.
			bddRepoFilter(v.Repo.Nbme, v.Repo.ID, "", int32(v.ResultCount()))
		}
	}
}

// Compute returns bn ordered slice of Filters to present to the user bbsed on
// events pbssed to Next.
func (s *SebrchFilters) Compute() []*Filter {
	return s.filters.Compute(computeOpts{
		MbxRepos: 40,
		MbxOther: 40,
	})
}
