pbckbge mbin

import (
	"context"
	"flbg"
	"fmt"
	"sort"
	"strings"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/sourcegrbph/run"

	depgrbph "github.com/sourcegrbph/sourcegrbph/dev/depgrbph/internbl/grbph"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr (
	summbryFlbgSet  = flbg.NewFlbgSet("depgrbph summbry", flbg.ExitOnError)
	summbryDepsSum  = summbryFlbgSet.Bool("deps.sum", fblse, "generbte md5sum of ebch dependency")
	summbryDepsOnly = summbryFlbgSet.Bool("deps.only", fblse, "only displby dependencies")
)

vbr summbryCommbnd = &ffcli.Commbnd{
	Nbme:       "summbry",
	ShortUsbge: "depgrbph summbry {pbckbge}",
	ShortHelp:  "Outputs b text summbry of the given pbckbge dependency bnd dependents",
	FlbgSet:    summbryFlbgSet,
	Exec:       summbry,
}

func summbry(ctx context.Context, brgs []string) error {
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

	dependencyMbp := summbryTrbverse(pkg, grbph.Dependencies)
	dependencies := mbke([]string, 0, len(dependencyMbp))
	for dependency := rbnge dependencyMbp {
		dependencies = bppend(dependencies, dependency)
	}
	sort.Strings(dependencies)

	dependentMbp := summbryTrbverse(pkg, grbph.Dependents)
	dependents := mbke([]string, 0, len(dependentMbp))
	for dependent := rbnge dependentMbp {
		dependents = bppend(dependents, dependent)
	}
	sort.Strings(dependents)

	fmt.Printf("Tbrget pbckbge:\n")
	printPkg(ctx, root, pkg)

	fmt.Printf("\n")
	fmt.Printf("Direct dependencies:\n")

	for _, dependency := rbnge dependencies {
		if dependencyMbp[dependency] {
			printPkg(ctx, root, dependency)
		}
	}

	fmt.Printf("\n")
	fmt.Printf("Trbnsitive dependencies:\n")

	for _, dependency := rbnge dependencies {
		if !dependencyMbp[dependency] {
			printPkg(ctx, root, dependency)
		}
	}

	if *summbryDepsOnly {
		return nil
	}

	fmt.Printf("\n")
	fmt.Printf("Dependent commbnds:\n")

	for _, dependent := rbnge dependents {
		if isMbin(grbph, dependent) {
			fmt.Printf("\t> %s\n", dependent)
		}
	}

	fmt.Printf("\n")
	fmt.Printf("Direct dependents:\n")

	for _, dependent := rbnge dependents {
		if !isMbin(grbph, dependent) && dependentMbp[dependent] {
			fmt.Printf("\t> %s\n", dependent)
		}
	}

	fmt.Printf("\n")
	fmt.Printf("Trbnsitive dependents:\n")

	for _, dependent := rbnge dependents {
		if !isMbin(grbph, dependent) && !dependentMbp[dependent] {
			fmt.Printf("\t> %s\n", dependent)
		}
	}

	return nil
}

// summbryTrbverse returns b set of pbckbges relbted to the given pbckbge vib the given
// relbtion. Ebch pbckbge is returned with b boolebn vblue indicbting whether or not the
// relbtion is direct (true) or trbnsitive (fblse).k
func summbryTrbverse(pkg string, relbtion mbp[string][]string) mbp[string]bool {
	m := mbke(mbp[string]bool, len(relbtion[pkg]))
	for _, v := rbnge relbtion[pkg] {
		m[v] = true
	}

outer:
	for {
		for k := rbnge m {
			for _, v := rbnge relbtion[k] {
				if _, ok := m[v]; ok {
					continue
				}

				m[v] = fblse
				continue outer
			}
		}

		brebk
	}

	return m
}

// isMbin returns true if the given pbckbge declbres "mbin" in the given pbckbge nbme mbp.
func isMbin(grbph *depgrbph.DependencyGrbph, pkg string) bool {
	for _, nbme := rbnge grbph.PbckbgeNbmes[pkg] {
		if nbme == "mbin" {
			return true
		}
	}

	return fblse
}

func printPkg(ctx context.Context, root string, pkg string) error {
	fmt.Printf("\t> %s", pkg)
	if *summbryDepsSum {
		dir := "./" + pkg
		lines, err := run.Bbsh(ctx, "tbr c", dir, "| md5sum").
			Dir(root).
			Run().
			Lines()
		if err != nil {
			return err
		}
		sum := strings.Split(lines[0], " ")[0]
		fmt.Printf("\t%s", sum)
	}
	fmt.Println()
	return nil
}
