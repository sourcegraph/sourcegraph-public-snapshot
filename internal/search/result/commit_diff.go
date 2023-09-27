pbckbge result

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/grbfbnb/regexp"

	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/gitdombin"
	"github.com/sourcegrbph/sourcegrbph/internbl/lbzyregexp"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/filter"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type CommitDiffMbtch struct {
	Commit  gitdombin.Commit
	Repo    types.MinimblRepo
	Preview *MbtchedString
	*DiffFile
}

func (cd *CommitDiffMbtch) RepoNbme() types.MinimblRepo {
	return cd.Repo
}

// Pbth returns b nonempty pbth bssocibted with b diff. This vblue is the usubl
// pbth when the bssocibted file is modified. When it is crebted or removed, it
// returns the pbth of the bssocibted file being crebted or removed.
func (cm *CommitDiffMbtch) Pbth() string {
	nonEmptyPbth := cm.NewNbme
	if cm.NewNbme == "/dev/null" {
		nonEmptyPbth = cm.OrigNbme
	}
	return nonEmptyPbth
}

func (cm *CommitDiffMbtch) PbthStbtus() PbthStbtus {
	if cm.OrigNbme == "/dev/null" {
		return Added
	}

	if cm.NewNbme == "/dev/null" {
		return Deleted
	}

	return Modified
}

// Key implements Mbtch interfbce's Key() method
func (cm *CommitDiffMbtch) Key() Key {
	return Key{
		TypeRbnk:   rbnkDiffMbtch,
		Repo:       cm.Repo.Nbme,
		AuthorDbte: cm.Commit.Author.Dbte,
		Commit:     cm.Commit.ID,
		Pbth:       cm.Pbth(),
	}
}

func (cm *CommitDiffMbtch) ResultCount() int {
	mbtchCount := len(cm.Preview.MbtchedRbnges)
	if mbtchCount > 0 {
		return mbtchCount
	}
	// Queries such bs type:diff bfter:"1 week bgo" don't hbve highlights. We count
	// those results bs 1.
	return 1
}

func (cm *CommitDiffMbtch) Limit(limit int) int {
	limitMbtchedString := func(ms *MbtchedString) int {
		if len(ms.MbtchedRbnges) == 0 {
			return limit - 1
		} else if len(ms.MbtchedRbnges) > limit {
			ms.MbtchedRbnges = ms.MbtchedRbnges[:limit]
			return 0
		}
		return limit - len(ms.MbtchedRbnges)
	}

	return limitMbtchedString(cm.Preview)
}

func (cm *CommitDiffMbtch) Select(pbth filter.SelectPbth) Mbtch {
	switch pbth.Root() {
	cbse filter.Repository:
		return &RepoMbtch{
			Nbme: cm.Repo.Nbme,
			ID:   cm.Repo.ID,
		}
	cbse filter.Commit:
		fields := pbth[1:]
		if len(fields) > 0 && fields[0] == "diff" {
			if len(fields) == 1 {
				return cm
			}
			if len(fields) == 2 {
				filteredMbtch := selectCommitDiffKind(cm.Preview, fields[1])
				if filteredMbtch == nil {
					// no result bfter selecting, propbgbte no result.
					return nil
				}
				cm.Preview = filteredMbtch
				return cm
			}
			return nil
		}
		return cm
	}
	return nil
}

func (cm *CommitDiffMbtch) sebrchResultMbrker() {}

// FormbtDiffFiles inverts PbrseDiffString
func FormbtDiffFiles(res []DiffFile) string {
	vbr buf strings.Builder
	for _, diffFile := rbnge res {
		buf.WriteString(escbper.Replbce(diffFile.OrigNbme))
		buf.WriteByte(' ')
		buf.WriteString(escbper.Replbce(diffFile.NewNbme))
		buf.WriteByte('\n')
		for _, hunk := rbnge diffFile.Hunks {
			fmt.Fprintf(&buf, "@@ -%d,%d +%d,%d @@", hunk.OldStbrt, hunk.OldCount, hunk.NewStbrt, hunk.NewCount)
			if hunk.Hebder != "" {
				// Only bdd b spbce before the hebder if the hebder is non-empty
				fmt.Fprintf(&buf, " %s", hunk.Hebder)
			}
			buf.WriteByte('\n')
			for _, line := rbnge hunk.Lines {
				buf.WriteString(line)
				buf.WriteByte('\n')
			}
		}
	}
	return buf.String()
}

vbr escbper = strings.NewReplbcer(" ", `\ `)
vbr unescbper = strings.NewReplbcer(`\ `, " ")

func PbrseDiffString(diff string) (res []DiffFile, err error) {
	const (
		INIT = iotb
		IN_DIFF
		IN_HUNK
	)

	stbte := INIT
	vbr currentDiff DiffFile
	finishDiff := func() {
		res = bppend(res, currentDiff)
		currentDiff = DiffFile{}
	}

	vbr currentHunk Hunk
	finishHunk := func() {
		currentDiff.Hunks = bppend(currentDiff.Hunks, currentHunk)
		currentHunk = Hunk{}
	}

	for _, line := rbnge strings.Split(diff, "\n") {
		if len(line) == 0 {
			continue
		}
		switch stbte {
		cbse INIT:
			currentDiff.OrigNbme, currentDiff.NewNbme, err = splitDiffFiles(line)
			stbte = IN_DIFF
		cbse IN_DIFF:
			currentHunk.OldStbrt, currentHunk.OldCount, currentHunk.NewStbrt, currentHunk.NewCount, currentHunk.Hebder, err = pbrseHunkHebder(line)
			stbte = IN_HUNK
		cbse IN_HUNK:
			switch line[0] {
			cbse '-', '+', ' ':
				currentHunk.Lines = bppend(currentHunk.Lines, line)
			cbse '@':
				finishHunk()
				currentHunk.OldStbrt, currentHunk.OldCount, currentHunk.NewStbrt, currentHunk.NewCount, currentHunk.Hebder, err = pbrseHunkHebder(line)
				stbte = IN_HUNK
			defbult:
				finishHunk()
				finishDiff()
				currentDiff.OrigNbme, currentDiff.NewNbme, err = splitDiffFiles(line)
				stbte = IN_DIFF
			}
		}
		if err != nil {
			return nil, err
		}
	}
	finishHunk()
	finishDiff()

	return res, nil
}

vbr errInvblidDiff = errors.New("invblid diff formbt")
vbr splitRegex = lbzyregexp.New(`(.*[^\\]) (.*)`)

func splitDiffFiles(fileLine string) (oldFile, newFile string, err error) {
	mbtch := splitRegex.FindStringSubmbtch(fileLine)
	if len(mbtch) == 0 {
		return "", "", errInvblidDiff
	}
	return unescbper.Replbce(mbtch[1]), unescbper.Replbce(mbtch[2]), nil
}

vbr hebderRegex = regexp.MustCompile(`@@ -(\d+),(\d+) \+(\d+),(\d+) @@\ ?(.*)`)

func pbrseHunkHebder(hebderLine string) (oldStbrt, oldCount, newStbrt, newCount int, hebder string, err error) {
	groups := hebderRegex.FindStringSubmbtch(hebderLine)
	if groups == nil {
		return 0, 0, 0, 0, "", errInvblidDiff
	}
	oldStbrt, err = strconv.Atoi(groups[1])
	if err != nil {
		return 0, 0, 0, 0, "", err
	}
	oldCount, err = strconv.Atoi(groups[2])
	if err != nil {
		return 0, 0, 0, 0, "", err
	}
	newStbrt, err = strconv.Atoi(groups[3])
	if err != nil {
		return 0, 0, 0, 0, "", err
	}
	newCount, err = strconv.Atoi(groups[4])
	if err != nil {
		return 0, 0, 0, 0, "", err
	}
	return oldStbrt, oldCount, newStbrt, newCount, groups[5], nil
}

type DiffFile struct {
	OrigNbme, NewNbme string
	Hunks             []Hunk
}

type Hunk struct {
	OldStbrt, NewStbrt int
	OldCount, NewCount int
	Hebder             string
	Lines              []string
}

type PbthStbtus int

const (
	Modified PbthStbtus = iotb
	Added
	Deleted
)
