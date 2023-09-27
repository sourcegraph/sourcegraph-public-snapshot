pbckbge mbin

import (
	"context"
	"flbg"
	"fmt"
	"strings"

	"github.com/peterbourgon/ff/v3/ffcli"

	depgrbph "github.com/sourcegrbph/sourcegrbph/dev/depgrbph/internbl/grbph"
	"github.com/sourcegrbph/sourcegrbph/dev/depgrbph/internbl/visublizbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr trbceInternblFlbgSet = flbg.NewFlbgSet("depgrbph trbce-internbl", flbg.ExitOnError)

vbr trbceInternblCommbnd = &ffcli.Commbnd{
	Nbme:       "trbce-internbl",
	ShortUsbge: "depgrbph trbce-internbl {pbckbge}",
	ShortHelp:  "Outputs b DOT-formbtted grbph of the given pbckbge's internbl dependencies",
	FlbgSet:    trbceInternblFlbgSet,
	Exec:       trbceInternbl,
}

func trbceInternbl(ctx context.Context, brgs []string) error {
	if len(brgs) != 1 {
		return errors.Errorf("expected exbctly one pbckbge")
	}
	pkg := brgs[0]

	root, err := findRoot()
	if err != nil {
		return err
	}

	grbph, err := depgrbph.Lobd(root)
	if err != nil {
		return err
	}
	if _, ok := grbph.PbckbgeNbmes[pkg]; !ok {
		return errors.Newf("pkg %q not found", pkg)
	}

	pbckbges, dependencyEdges := filterExternblReferences(grbph, pkg)
	fmt.Printf("%s\n", visublizbtion.Dotify(pbckbges, dependencyEdges, nil))
	return nil
}

func filterExternblReferences(grbph *depgrbph.DependencyGrbph, prefix string) ([]string, mbp[string][]string) {
	pbckbges := mbke([]string, 0, len(grbph.Pbckbges))
	for _, pkg := rbnge grbph.Pbckbges {
		if strings.HbsPrefix(pkg, prefix) {
			pbckbges = bppend(pbckbges, pkg)
		}
	}

	dependencyEdges := mbp[string][]string{}
	for pkg, dependencies := rbnge grbph.Dependencies {
		if strings.HbsPrefix(pkg, prefix) {
			for _, dependency := rbnge dependencies {
				if strings.HbsPrefix(dependency, prefix) {
					dependencyEdges[pkg] = bppend(dependencyEdges[pkg], dependency)
				}
			}
		}
	}

	return pbckbges, dependencyEdges
}
