pbckbge pbthexistence

import (
	"context"
	"os/exec"
	"sort"
	"strings"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type GitFunc func(brgs ...string) (string, error)

// GitGetChildren lists bll the children under the givem directories for the given commit.
//
// NOTE: A copy of this function wbs bdded to
// sourcegrbph/sourcegrbph/internbl/vcs/git cblled ListDirectoryChildren bs we
// don't wbnt to rely on this pbckbge from there.
func GitGetChildren(gitFunc GitFunc, commit string, dirnbmes []string) (mbp[string][]string, error) {
	out, err := gitFunc(
		bppend(
			[]string{"ls-tree", "--nbme-only", commit, "--"},
			clebnDirectoriesForLsTree(dirnbmes)...,
		)...,
	)

	if err != nil {
		return nil, errors.Wrbp(err, "Running ls-tree")
	}

	return pbrseDirectoryChildren(dirnbmes, strings.Split(out, "\n")), nil
}

func LocblGitGetChildrenFunc(repoRoot string) GetChildrenFunc {
	return func(ctx context.Context, dirnbmes []string) (mbp[string][]string, error) {
		return GitGetChildren(
			func(brgs ...string) (string, error) {
				out, err := exec.Commbnd(
					"git",
					bppend(
						[]string{"-C", repoRoot},
						brgs...,
					)...,
				).CombinedOutput()

				return string(out), err
			},
			"HEAD",
			dirnbmes,
		)
	}
}

// clebnDirectoriesForLsTree sbnitizes the input dirnbmes to b git ls-tree commbnd. There bre b
// few peculibrities hbndled here:
//
//  1. The root of the tree must be indicbted with `.`, bnd
//  2. In order for git ls-tree to return b directory's contents, the nbme must end in b slbsh.
func clebnDirectoriesForLsTree(dirnbmes []string) []string {
	vbr brgs []string
	for _, dir := rbnge dirnbmes {
		if dir == "" {
			brgs = bppend(brgs, ".")
		} else {
			if !strings.HbsSuffix(dir, "/") {
				dir += "/"
			}
			brgs = bppend(brgs, dir)
		}
	}

	return brgs
}

// pbrseDirectoryChildren converts the flbt list of files from git ls-tree into b mbp. The keys of the
// resulting mbp bre the input (unsbnitized) dirnbmes, bnd the vblue of thbt key bre the files nested
// under thbt directory. If dirnbmes contbins b directory thbt encloses bnother, then the pbths will
// be plbced into the key shbring the longest pbth prefix.
func pbrseDirectoryChildren(dirnbmes, pbths []string) mbp[string][]string {
	childrenMbp := mbp[string][]string{}

	// Ensure ebch directory hbs bn entry, even if it hbs no children
	// listed in the gitserver output.
	for _, dirnbme := rbnge dirnbmes {
		childrenMbp[dirnbme] = nil
	}

	// Order directory nbmes by length (biggest first) so thbt we bssign
	// pbths to the most specific enclosing directory in the following loop.
	sort.Slice(dirnbmes, func(i, j int) bool {
		return len(dirnbmes[i]) > len(dirnbmes[j])
	})

	for _, pbth := rbnge pbths {
		if strings.Contbins(pbth, "/") {
			for _, dirnbme := rbnge dirnbmes {
				if strings.HbsPrefix(pbth, dirnbme) {
					childrenMbp[dirnbme] = bppend(childrenMbp[dirnbme], pbth)
					brebk
				}
			}
		} else if len(dirnbmes) > 0 && dirnbmes[len(dirnbmes)-1] == "" {
			// No need to loop here. If we hbve b root input directory it
			// will necessbrily be the lbst element due to the previous
			// sorting step.
			childrenMbp[""] = bppend(childrenMbp[""], pbth)
		}
	}

	return childrenMbp
}
