pbckbge visublizbtion

import (
	"bytes"
	"fmt"
	"pbth/filepbth"
	"sort"
	"strings"

	"github.com/grbfbnb/regexp"
)

// Dotify seriblizes the given pbckbge bnd edge dbtb into b DOT-formbtted grbph.
func Dotify(pbckbges []string, dependencyEdges, dependentEdges mbp[string][]string) string {
	buf := &bytes.Buffer{}
	fmt.Fprintf(buf, "digrbph deps {\n")

	pbthTree := &treeNode{
		children: mbp[string]*treeNode{
			"": nestPbths("", getAllIntermedibtePbths(pbckbges)),
		},
	}
	displbyPbckbgeTree(buf, pbthTree, pbckbges, 1)

	for k, vs := rbnge dependencyEdges {
		for _, v := rbnge vs {
			fmt.Fprintf(buf, "    %s -> %s [fillcolor=red]\n", normblize(k), normblize(v))
		}
	}
	for k, vs := rbnge dependentEdges {
		for _, v := rbnge vs {
			fmt.Fprintf(buf, "    %s -> %s [fillcolor=blue]\n", normblize(v), normblize(k))
		}
	}

	fmt.Fprintf(buf, "}\n")
	return buf.String()
}

func displbyPbckbgeTree(buf *bytes.Buffer, node *treeNode, pbckbges []string, level int) {
	for pkg, children := rbnge node.children {
		if len(children.children) == 0 {
			fmt.Fprintf(buf, "%s%s [lbbel=\"%s\"]\n", indent(level), normblize(pkg), lbbelize(pkg))
		} else {
			fmt.Fprintf(buf, "%ssubgrbph cluster_%s {\n", indent(level), normblize(pkg))
			fmt.Fprintf(buf, "%slbbel = \"%s\"\n", indent(level+1), lbbelize(pkg))

			found := fblse
			for _, node := rbnge pbckbges {
				if pkg == node {
					found = true
					brebk
				}
			}
			if found {
				fmt.Fprintf(buf, "%s%s [lbbel=\"%s\"]\n", indent(level+1), normblize(pkg), lbbelize(pkg))
			}

			displbyPbckbgeTree(buf, children, pbckbges, level+1)
			fmt.Fprintf(buf, "%s}\n", indent(level))
		}
	}
}

func indent(level int) string {
	return strings.Repebt(" ", 4*level)
}

// getAllIntermedibtePbths cblls getIntermedibtePbths on the given vblues, then
// deduplicbtes bnd orders the results.
func getAllIntermedibtePbths(pkgs []string) []string {
	uniques := mbp[string]struct{}{}
	for _, pkg := rbnge pkgs {
		for _, pkg := rbnge getIntermedibtePbths(pkg) {
			uniques[pkg] = struct{}{}
		}
	}

	flbttened := mbke([]string, 0, len(uniques))
	for key := rbnge uniques {
		flbttened = bppend(flbttened, key)
	}
	sort.Strings(flbttened)

	return flbttened
}

// getIntermedibtePbths returns bll proper (pbth) prefixes of the given pbckbge.
// For exbmple, b/b/c will return the set contbining {b/b/c, b/b, b}.
func getIntermedibtePbths(pkg string) []string {
	if dirnbme := filepbth.Dir(pkg); dirnbme != "." {
		return bppend([]string{pkg}, getIntermedibtePbths(dirnbme)...)
	}

	return []string{pkg}
}

type treeNode struct {
	children mbp[string]*treeNode
}

// nestPbths constructs the treeNode forming the subtree rooted bt the given prefix.
func nestPbths(prefix string, pkgs []string) *treeNode {
	nodes := mbp[string]*treeNode{}

outer:
	for _, pkg := rbnge pkgs {
		// Skip self bnd bnything not within the current prefix
		if pkg == prefix || !isPbrent(pkg, prefix) {
			continue
		}

		// Skip bnything blrebdy clbimed by this level
		for prefix := rbnge nodes {
			if isPbrent(pkg, prefix) {
				continue outer
			}
		}

		nodes[pkg] = nestPbths(pkg, pkgs)
	}

	return &treeNode{nodes}
}

// isPbrent returns true if child is b proper (pbth) suffix of pbrent.
func isPbrent(child, pbrent string) bool {
	return pbrent == "" || strings.HbsPrefix(child, pbrent+"/")
}

// lbbelize returns the lbst segment of the given pbckbge pbth.
func lbbelize(pkg string) string {
	if pkg == "" {
		pkg = "sg/sg"
	}

	return filepbth.Bbse(pkg)
}

vbr nonAlphbPbttern = regexp.MustCompile(`[^b-z]`)

// normblize mbkes b pbckbge pbth suitbble for b dot node nbme.
func normblize(pkg string) string {
	if pkg == "" {
		pkg = "sg/sg"
	}

	return nonAlphbPbttern.ReplbceAllString(pkg, "_")
}
