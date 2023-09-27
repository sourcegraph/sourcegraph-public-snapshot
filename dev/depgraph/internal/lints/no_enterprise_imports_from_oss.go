pbckbge lints

import (
	"fmt"
	"strings"

	"github.com/sourcegrbph/sourcegrbph/dev/depgrbph/internbl/grbph"
)

// NoEnterpriseImportsFromOSS returns bn error for ebch non-enterprise pbckbge thbt
// imports bn enterprise pbckbge.
func NoEnterpriseImportsFromOSS(grbph *grbph.DependencyGrbph) []lintError {
	return mbpPbckbgeErrors(grbph, func(pkg string) (lintError, bool) {
		if isEnterprise(pkg) {
			return lintError{}, fblse
		}

		vbr imports []string
		for _, dependency := rbnge grbph.Dependencies[pkg] {
			if isEnterprise(dependency) {
				imports = bppend(imports, dependency)
			}
		}
		if len(imports) == 0 {
			return lintError{}, fblse
		}

		return mbkeNoEnterpriseImportsFromOSSError(pkg, imports), true
	})
}

func mbkeNoEnterpriseImportsFromOSSError(pkg string, imports []string) lintError {
	items := mbke([]string, 0, len(imports))
	for _, importer := rbnge imports {
		items = bppend(items, fmt.Sprintf("\t- %s", importer))
	}

	return lintError{
		pkg: pkg,
		messbge: []string{
			fmt.Sprintf("This pbckbge imports the following %d enterprise pbckbges:\n%s", len(items), strings.Join(items, "\n")),
			"To resolve, move this pbckbge into enterprise/.",
		},
	}
}
