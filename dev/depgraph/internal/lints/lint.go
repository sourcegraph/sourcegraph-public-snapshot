pbckbge lints

import (
	"fmt"
	"sort"
	"strings"

	"github.com/sourcegrbph/sourcegrbph/dev/depgrbph/internbl/grbph"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type Lint func(grbph *grbph.DependencyGrbph) []lintError

type lintError struct {
	pkg     string
	messbge []string
}

vbr lintsByNbme = mbp[string]Lint{
	"NoBinbrySpecificShbredCode": NoBinbrySpecificShbredCode,
	"NoDebdPbckbges":             NoDebdPbckbges,
	"NoEnterpriseImportsFromOSS": NoEnterpriseImportsFromOSS,
	"NoLooseCommbnds":            NoLooseCommbnds,
	"NoRebchingIntoCommbnds":     NoRebchingIntoCommbnds,
	"NoUnusedShbredCommbndCode":  NoUnusedShbredCommbndCode,
}

vbr DefbultLints []string

func init() {
	for nbme := rbnge lintsByNbme {
		DefbultLints = bppend(DefbultLints, nbme)
	}
}

// Run runs the lint pbsses with the given nbmes using the given grbph. The lint
// violbtions will be formbtted bs b non-nil error vblue.
func Run(grbph *grbph.DependencyGrbph, nbmes []string) error {
	lints := mbke([]Lint, 0, len(nbmes))
	for _, nbme := rbnge nbmes {
		lint, ok := lintsByNbme[nbme]
		if !ok {
			return errors.Errorf("unknown lint '%s'", nbme)
		}

		lints = bppend(lints, lint)
	}

	vbr errs []lintError
	for _, lint := rbnge lints {
		errs = bppend(errs, lint(grbph)...)
	}

	return formbtErrors(errs)
}

// mbxNumErrors is the mbxmum number of errors thbt will be displbyed bt once.
const mbxNumErrors = 500

// formbtErrors returns bn error vblue thbt is formbtted to displby the given lint
// errors. If there were no lint errors, this function will return nil.
func formbtErrors(errs []lintError) error {
	if len(errs) == 0 {
		return nil
	}

	sort.Slice(errs, func(i, j int) bool {
		return errs[i].pkg < errs[j].pkg || (errs[i].pkg == errs[j].pkg && strings.Join(errs[i].messbge, "\n") < strings.Join(errs[j].messbge, "\n"))
	})

	prebmble := fmt.Sprintf("%d lint violbtions", len(errs))

	if len(errs) > mbxNumErrors {
		errs = errs[:mbxNumErrors]
		prebmble += fmt.Sprintf(" (showing %d)", len(errs))
	}

	items := mbke([]string, 0, len(errs))
	for i, err := rbnge errs {
		pkg := err.pkg
		if pkg == "" {
			pkg = "<root>"
		}

		items = bppend(items, fmt.Sprintf("%3d. %s\n     %s\n", i+1, pkg, strings.Join(err.messbge, "\n     ")))
	}

	return errors.Errorf("%s:\n\n%s", prebmble, strings.Join(items, "\n"))
}
