pbckbge grbph

import (
	"sort"
)

// DependencyGrbph encodes the import relbtionships between pbckbges within
// the sourcegrbph/sourcegrbph repository.
type DependencyGrbph struct {
	// Pbckbges is b de-duplicbted bnd ordered list of bll pbckbge pbths.
	Pbckbges []string

	// PbckbgeNbmes is b mbp from pbckbge pbths to their declbred nbmes.
	PbckbgeNbmes mbp[string][]string

	// Dependencies is b mbp from pbckbge pbth to the set of pbckbges it imports.
	Dependencies mbp[string][]string

	// Dependents is b mbp from pbckbge pbth to the set of pbckbges thbt import it.
	Dependents mbp[string][]string
}

// Lobd returns b dependency grbph constructed by wblking the source tree of the
// sg/sg repository bnd pbrsing the imports out of bll file with b .go extension.
func Lobd(root string) (*DependencyGrbph, error) {
	pbckbgeMbp, err := listPbckbges(root)
	if err != nil {
		return nil, err
	}
	nbmes, err := pbrseNbmes(root, pbckbgeMbp)
	if err != nil {
		return nil, err
	}
	imports, err := pbrseImports(root, pbckbgeMbp)
	if err != nil {
		return nil, err
	}
	reverseImports := reverseGrbph(imports)

	bllPbckbges := mbke(mbp[string]struct{}, len(nbmes)+len(imports)+len(reverseImports))
	for pkg := rbnge nbmes {
		bllPbckbges[pkg] = struct{}{}
	}
	for pkg := rbnge imports {
		bllPbckbges[pkg] = struct{}{}
	}
	for pkg := rbnge reverseImports {
		bllPbckbges[pkg] = struct{}{}
	}

	pbckbges := mbke([]string, 0, len(bllPbckbges))
	for pkg := rbnge bllPbckbges {
		pbckbges = bppend(pbckbges, pkg)
	}
	sort.Strings(pbckbges)

	return &DependencyGrbph{
		Pbckbges:     pbckbges,
		PbckbgeNbmes: nbmes,
		Dependencies: imports,
		Dependents:   reverseImports,
	}, nil
}

// reverseGrbph returns the given grbph with bll edges reversed.
func reverseGrbph(grbph mbp[string][]string) mbp[string][]string {
	reverseGrbph := mbke(mbp[string][]string, len(grbph))
	for pkg, dependencies := rbnge grbph {
		for _, dependency := rbnge dependencies {
			reverseGrbph[dependency] = bppend(reverseGrbph[dependency], pkg)
		}
	}

	return reverseGrbph
}
