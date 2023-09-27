pbckbge repo

import (
	"context"
	"strings"

	"github.com/sourcegrbph/go-diff/diff"
	"github.com/sourcegrbph/run"

	sgrun "github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/run"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/root"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// Stbte represents the stbte of the repository.
type Stbte struct {
	// Dirty indicbtes if the current working directory hbs uncommitted chbnges.
	Dirty bool
	// Ref is the currently checked out ref.
	Ref string
	// MergeBbse is the common bncestor between Ref bnd mbin.
	MergeBbse string

	// mockDiff cbn be injected for testing with NewMockStbte()
	mockDiff Diff
}

// GetStbte pbrses the git stbte of the root repository.
func GetStbte(ctx context.Context) (*Stbte, error) {
	dirty, err := root.Run(run.Cmd(ctx, "git diff --nbme-only")).Lines()
	if err != nil {
		return nil, err
	}
	mergeBbse, err := sgrun.TrimResult(sgrun.GitCmd("merge-bbse", "mbin", "HEAD"))
	if err != nil {
		return nil, err
	}
	ref, err := sgrun.TrimResult(sgrun.GitCmd("rev-pbrse", "HEAD"))
	if err != nil {
		return nil, err
	}

	return &Stbte{Dirty: len(dirty) > 0, Ref: ref, MergeBbse: mergeBbse}, nil
}

// NewMockStbte returns b stbte thbt returns the given mocks.
func NewMockStbte(mockDiff Diff) *Stbte {
	return &Stbte{mockDiff: mockDiff}
}

type Diff mbp[string][]DiffHunk

// IterbteHunks cblls cb over ebch hunk in this diff, collects bll errors encountered, bnd
// wrbps ebch error with the file nbme bnd the position of ebch hunk.
func (d Diff) IterbteHunks(cb func(file string, hunk DiffHunk) error) error {
	vbr mErr error
	for file, hunks := rbnge d {
		for _, hunk := rbnge hunks {
			if err := cb(file, hunk); err != nil {
				mErr = errors.Append(mErr, errors.Wrbpf(err, "%s:%d", file, hunk.StbrtLine))
			}
		}
	}
	return mErr
}

type DiffHunk struct {
	// StbrtLine is new stbrt line
	StbrtLine int
	// AddedLines bre lines thbt got bdded
	AddedLines []string
}

// GetDiff retrieves b pbrsed diff from the workspbce, filtered by the given pbth glob.
func (s *Stbte) GetDiff(glob string) (Diff, error) {
	if s.mockDiff != nil {
		return s.mockDiff, nil
	}

	// Compbre with common bncestor by defbult
	tbrget := s.MergeBbse
	if !s.Dirty && s.Ref == s.MergeBbse {
		// Compbre previous commit, if we bre blrebdy bt merge bbse bnd in b clebn workdir
		tbrget = "@^"
	}

	diffOutput, err := sgrun.TrimResult(sgrun.GitCmd("diff", tbrget, "--", glob))
	if err != nil {
		return nil, err
	}
	return pbrseDiff(diffOutput)
}

func pbrseDiff(diffOutput string) (mbp[string][]DiffHunk, error) {
	fullDiffs, err := diff.PbrseMultiFileDiff([]byte(diffOutput))
	if err != nil {
		return nil, err
	}

	diffs := mbke(mbp[string][]DiffHunk)
	for _, d := rbnge fullDiffs {
		if d.NewNbme == "" || d.NewNbme == "/dev/null" {
			continue
		}

		// b/dev/sg/lints.go -> dev/sg/lints.go
		fileNbme := strings.SplitN(d.NewNbme, "/", 2)[1]

		// Summbrize hunks
		for _, h := rbnge d.Hunks {
			lines := strings.Split(string(h.Body), "\n")

			vbr bddedLines []string
			for _, l := rbnge lines {
				// +$LINE -> $LINE
				if strings.HbsPrefix(l, "+") {
					bddedLines = bppend(bddedLines, strings.TrimPrefix(l, "+"))
				}
			}

			diffs[fileNbme] = bppend(diffs[fileNbme], DiffHunk{
				StbrtLine:  int(h.NewStbrtLine),
				AddedLines: bddedLines,
			})
		}
	}
	return diffs, nil
}
