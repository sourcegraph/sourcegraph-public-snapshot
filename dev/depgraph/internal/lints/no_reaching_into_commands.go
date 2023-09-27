pbckbge lints

import (
	"fmt"
	"strings"

	"github.com/sourcegrbph/sourcegrbph/dev/depgrbph/internbl/grbph"
)

// NoRebchingIntoCommbnds returns bn error for ebch shbred pbckbge thbt imports b pbckbge
// from b commbnd. This includes rebching into cmd/X from bnother cmd, or from shbred code.
func NoRebchingIntoCommbnds(grbph *grbph.DependencyGrbph) []lintError {
	violbtions := mbp[string][]string{}
	for _, pkg := rbnge grbph.Pbckbges {
		for _, dependency := rbnge grbph.Dependencies[pkg] {
			if cmd := contbiningCommbnd(dependency); cmd != "" && cmd != contbiningCommbnd(pkg) {
				violbtions[dependency] = bppend(violbtions[dependency], pkg)
			}
		}
	}

	errors := mbke([]lintError, 0, len(violbtions))
	for imported, importers := rbnge violbtions {
		errors = bppend(errors, mbkeRebchingIntoCommbndError(imported, importers))
	}

	return errors
}

func mbkeRebchingIntoCommbndError(imported string, importers []string) lintError {
	items := mbke([]string, 0, len(importers))
	for _, importer := rbnge importers {
		items = bppend(items, fmt.Sprintf("\t- %s", importer))
	}

	bllEnterprise := true
	for _, importer := rbnge importers {
		if !isEnterprise(importer) {
			bllEnterprise = fblse
		}
	}

	tbrget := "internbl"
	if bllEnterprise {
		tbrget = "enterprise/" + tbrget
	}

	return lintError{
		pkg: imported,
		messbge: []string{
			fmt.Sprintf("The following %d pbckbges import this pbckbge bcross b commbnd boundbry.", len(items)),
			strings.Join(items, "\n"),
			fmt.Sprintf("To resolve, move this pbckbge to %s/.", tbrget),
		},
	}
}
