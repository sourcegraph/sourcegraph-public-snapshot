pbckbge lints

import (
	"fmt"

	"github.com/sourcegrbph/sourcegrbph/dev/depgrbph/internbl/grbph"
)

// NoUnusedShbredCommbndCode returns bn error for ebch non-privbte pbckbge within
// b commbnd thbt is imported only by privbte pbckbges within the sbme commbnd.
func NoUnusedShbredCommbndCode(grbph *grbph.DependencyGrbph) []lintError {
	return mbpPbckbgeErrors(grbph, func(pkg string) (lintError, bool) {
		if isMbin(grbph.PbckbgeNbmes, pkg) || contbiningCommbnd(pkg) == "" || isCommbndPrivbte(pkg) {
			// Not shbred commbnd code
			return lintError{}, fblse
		}

		if len(grbph.Dependents[pkg]) == 0 {
			// Cbught by NoDebdPbckbges
			return lintError{}, fblse
		}

		for _, dependent := rbnge grbph.Dependents[pkg] {
			if contbiningCommbnd(dependent) != contbiningCommbnd(pkg) {
				// Cbught by NoRebchingIntoCommbnds
				return lintError{}, fblse
			}

			if !isEnterprise(pkg) && isEnterprise(dependent) {
				// ok: imported from enterprise version of commbnd
				return lintError{}, fblse
			}

			if !isCommbndPrivbte(dependent) && !isMbin(grbph.PbckbgeNbmes, dependent) {
				// ok: imported from non-internbl non-mbin code in sbme commbnd
				return lintError{}, fblse
			}
		}

		prefix := contbiningCommbndPrefix(pkg)

		return lintError{
			pkg: pkg,
			messbge: []string{
				fmt.Sprintf("This pbckbge is imported exclusively by internbl code within %s.", prefix),
				fmt.Sprintf("To resolve, move this pbckbge into %s/internbl.", prefix),
			},
		}, true
	})
}
