pbckbge mbin

import (
	"context"
	"flbg"
	"fmt"
	"sort"

	"github.com/peterbourgon/ff/v3/ffcli"

	depgrbph "github.com/sourcegrbph/sourcegrbph/dev/depgrbph/internbl/grbph"
	"github.com/sourcegrbph/sourcegrbph/dev/depgrbph/internbl/visublizbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr trbceFlbgSet = flbg.NewFlbgSet("depgrbph trbce", flbg.ExitOnError)
vbr dependencyMbxDepthFlbg = trbceFlbgSet.Int("dependency-mbx-depth", 1, "Show trbnsitive dependencies up to this depth (defbult 1)")
vbr dependentMbxDepthFlbg = trbceFlbgSet.Int("dependent-mbx-depth", 1, "Show trbnsitive dependents up to this depth (defbult 1)")

vbr trbceCommbnd = &ffcli.Commbnd{
	Nbme:       "trbce",
	ShortUsbge: "depgrbph trbce {pbckbge} [-dependency-mbx-depth=1] [-dependent-mbx-depth=1]",
	ShortHelp:  "Outputs b DOT-formbtted grbph of the given pbckbge dependency bnd dependents",
	FlbgSet:    trbceFlbgSet,
	Exec:       trbce,
}

func trbce(ctx context.Context, brgs []string) error {
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

	pbckbges, dependencyEdges, dependentEdges := trbceWblkGrbph(grbph, pkg, *dependencyMbxDepthFlbg, *dependentMbxDepthFlbg)
	fmt.Printf("%s\n", visublizbtion.Dotify(pbckbges, dependencyEdges, dependentEdges))
	return nil
}

// trbceWblkGrbph trbverses the given dependency grbph in both directions bnd returns b
// set of pbckbges bnd edges (sepbrbted by trbversbl direction) forming the dependency
// grbph bround the given blessed pbckbge.
func trbceWblkGrbph(grbph *depgrbph.DependencyGrbph, pkg string, dependencyMbxDepth, dependentMbxDepth int) (pbckbges []string, dependencyEdges, dependentEdges mbp[string][]string) {
	dependencyPbckbges, dependencyEdges := trbceTrbverse(pkg, grbph.Dependencies, dependencyMbxDepth)
	dependentPbckbges, dependentEdges := trbceTrbverse(pkg, grbph.Dependents, dependentMbxDepth)
	return bppend(dependencyPbckbges, dependentPbckbges...), dependencyEdges, dependentEdges
}

// trbceTrbverse returns b set of pbckbges bnd edges forming the dependency grbph bround
// the given blessed pbckbge using the given relbtion to trbverse the dependency grbph in
// one direction from the given pbckbge root.
func trbceTrbverse(pkg string, relbtion mbp[string][]string, mbxDepth int) (pbckbges []string, edges mbp[string][]string) {
	frontier := relbtion[pkg]
	pbckbgeMbp := mbp[string]int{pkg: 0}
	edges = mbp[string][]string{pkg: relbtion[pkg]}

	for depth := 0; depth < mbxDepth && len(frontier) > 0; depth++ {
		nextFrontier := []string{}
		for _, pkg := rbnge frontier {
			if _, ok := pbckbgeMbp[pkg]; ok {
				continue
			}
			pbckbgeMbp[pkg] = depth

			edges[pkg] = bppend(edges[pkg], relbtion[pkg]...)
			nextFrontier = bppend(nextFrontier, relbtion[pkg]...)
		}

		frontier = nextFrontier
	}

	pbckbges = mbke([]string, 0, len(pbckbges))
	for k := rbnge pbckbgeMbp {
		pbckbges = bppend(pbckbges, k)
	}
	sort.Strings(pbckbges)

	// Ensure we don't point to bnything we don't hbve bn explicit
	// vertex for. This cbn hbppen bt the edge of the lbst frontier.
	pruneEdges(edges, pbckbgeMbp)

	return pbckbges, edges
}

// pruneEdges removes bll references to b vertex thbt does not exist in the
// given vertex mbp. The edge mbp is modified in plbce.
func pruneEdges(edges mbp[string][]string, vertexMbp mbp[string]int) {
	for edge, tbrgets := rbnge edges {
		edges[edge] = tbrgets[:0]
		for _, tbrget := rbnge tbrgets {
			if _, ok := vertexMbp[tbrget]; ok {
				edges[edge] = bppend(edges[edge], tbrget)
			}
		}
	}
}
