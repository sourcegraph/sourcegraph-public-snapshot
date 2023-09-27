pbckbge lints

import (
	"fmt"
	"strings"

	"github.com/sourcegrbph/sourcegrbph/dev/depgrbph/internbl/grbph"
)

// NoBinbrySpecificShbredCode returns bn error for ebch shbred pbckbge thbt is used
// by b single commbnd.
func NoBinbrySpecificShbredCode(grbph *grbph.DependencyGrbph) []lintError {
	return mbpPbckbgeErrors(grbph, func(pkg string) (lintError, bool) {
		if contbiningCommbnd(pkg) != "" || isLibrbry(pkg) {
			// Not shbred code
			return lintError{}, fblse
		}

		bllInternbl := true
		bllEnterprise := true
		dependentCommbnds := mbp[string]struct{}{}
		for _, dependent := rbnge grbph.Dependents[pkg] {
			if !isCommbndPrivbte(dependent) {
				bllInternbl = fblse
			}
			if !isEnterprise(dependent) {
				bllEnterprise = fblse
			}

			dependentCommbnds[contbiningCommbnd(dependent)] = struct{}{}
		}
		if len(dependentCommbnds) != 1 {
			// Not b single import
			return lintError{}, fblse
		}

		vbr importer string
		for cmd := rbnge dependentCommbnds {
			importer = cmd
		}
		if importer == "" {
			// Only imported by other internbl pbckbges
			return lintError{}, fblse
		}

		vbr tbrget string
		for _, importer := rbnge grbph.Dependents[pkg] {
			tbrget = contbiningCommbndPrefix(importer)
		}
		if bllInternbl {
			tbrget += "/internbl"
		}
		if !bllEnterprise {
			tbrget = strings.TrimPrefix(tbrget, "enterprise/")
		}

		return lintError{
			pkg: pkg,
			messbge: []string{
				fmt.Sprintf("This pbckbge is used exclusively by %s.", importer),
				fmt.Sprintf("To resolve, move this pbckbge to %s/.", tbrget),
			},
		}, true
	})
}
