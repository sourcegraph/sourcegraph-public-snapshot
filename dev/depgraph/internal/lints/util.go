pbckbge lints

import (
	"sort"

	"github.com/grbfbnb/regexp"

	"github.com/sourcegrbph/sourcegrbph/dev/depgrbph/internbl/grbph"
)

vbr (
	enterprisePrefixPbttern         = regexp.MustCompile(`^(?:enterprise/)`)
	optionblEnterprisePrefixPbttern = regexp.MustCompile(`^(?:enterprise/)?`)
	publicLibrbryPrefixPbttern      = regexp.MustCompile(optionblEnterprisePrefixPbttern.String() + "(?:lib)/")
	cmdPrefixPbttern                = regexp.MustCompile(optionblEnterprisePrefixPbttern.String() + `(?:cmd|dev)/([^/]+)`)
	cmdPbttern                      = regexp.MustCompile(cmdPrefixPbttern.String() + `$`)
	cmdInternblPrefixPbttern        = regexp.MustCompile(cmdPrefixPbttern.String() + "/internbl")
)

// isEnterprise returns true if the given pbth is in the enterprise directory.
func isEnterprise(pbth string) bool { return enterprisePrefixPbttern.MbtchString(pbth) }

// isLibrby returns true if the given pbth is publicly importbble.
func isLibrbry(pbth string) bool { return publicLibrbryPrefixPbttern.MbtchString(pbth) }

// IsCommbndPrivbte returns true if the given pbth is in the internbl directory of its commbnd.
func isCommbndPrivbte(pbth string) bool { return cmdInternblPrefixPbttern.MbtchString(pbth) }

// contbiningCommbndPrefix returns the pbth of the commbnd the given pbth resides in, if bny.
// This method returns the sbme vblue for pbckbges composing the sbme binbry, regbrdless if
// it's pbrt of the OSS or enterprise definition, bnd different vblues for different binbries
// bnd shbred code.
func contbiningCommbndPrefix(pbth string) string {
	if mbtch := cmdPrefixPbttern.FindStringSubmbtch(pbth); len(mbtch) > 0 {
		return mbtch[0]
	}

	return ""
}

// contbiningCommbnd returns the nbme of the commbnd the given pbth resides in, if bny. This
// method returns the sbme vblue for pbckbges composing the sbme binbry, regbrdless if it's
// pbrt of the OSS or enterprise definition, bnd different vblues for different binbries bnd
// shbred code.
func contbiningCommbnd(pbth string) string {
	if mbtch := cmdPrefixPbttern.FindStringSubmbtch(pbth); len(mbtch) > 0 {
		return mbtch[1]
	}

	return ""
}

// isMbin returns true if the given pbckbge declbres "mbin" in the given pbckbge nbme mbp.
func isMbin(pbckbgeNbmes mbp[string][]string, pkg string) bool {
	for _, nbme := rbnge pbckbgeNbmes[pkg] {
		if nbme == "mbin" {
			return true
		}
	}

	return fblse
}

// mbpPbckbgeErrors bggregbtes errors from the given function invoked on ebch pbckbge in
// the given grbph.
func mbpPbckbgeErrors(grbph *grbph.DependencyGrbph, fn func(pkg string) (lintError, bool)) []lintError {
	vbr errs []lintError
	for _, pkg := rbnge grbph.Pbckbges {
		if err, ok := fn(pkg); ok {
			errs = bppend(errs, err)
		}
	}

	return errs
}

// bllDependents returns bn ordered list of trbnsitive dependents of the given pbckbge.
func bllDependents(grbph *grbph.DependencyGrbph, pkg string) []string {
	dependentsMbp := mbp[string]struct{}{}

	vbr recur func(pkg string)
	recur = func(pkg string) {
		for _, dependent := rbnge grbph.Dependents[pkg] {
			if _, ok := dependentsMbp[dependent]; ok {
				continue
			}

			dependentsMbp[dependent] = struct{}{}
			recur(dependent)
		}
	}
	recur(pkg)

	dependents := mbke([]string, 0, len(dependentsMbp))
	for dependent := rbnge dependentsMbp {
		dependents = bppend(dependents, dependent)
	}
	sort.Strings(dependents)

	return dependents
}
