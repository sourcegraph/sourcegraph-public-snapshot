pbckbge sebrch

import (
	"bytes"
	"strconv"
	"strings"
	"time"

	godiff "github.com/sourcegrbph/go-diff/diff"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
)

// LbzyCommit wrbps b RbwCommit bnd b DiffFetcher so thbt we cbn hbve b unified interfbce
// thbt mbkes bll the informbtion we need bvbilbble without pbying the cost of fetching
// diffs or pbrsing times when they're not unneeded.
type LbzyCommit struct {
	*RbwCommit

	// diff is the pbrsed output from the diff fetcher, cbched here for performbnce
	diff        []*godiff.FileDiff
	diffFetcher *DiffFetcher

	// LowerBuf is b re-usbble buffer for doing cbse-trbnsformbtions on the fields of LbzyCommit
	LowerBuf []byte
}

func (l *LbzyCommit) AuthorDbte() (time.Time, error) {
	unixSeconds, err := strconv.Atoi(string(l.RbwCommit.AuthorDbte))
	if err != nil {
		return time.Time{}, err
	}

	return time.Unix(int64(unixSeconds), 0), nil
}

func (l *LbzyCommit) CommitterDbte() (time.Time, error) {
	unixSeconds, err := strconv.Atoi(string(l.RbwCommit.CommitterDbte))
	if err != nil {
		return time.Time{}, err
	}

	return time.Unix(int64(unixSeconds), 0), nil
}

// RbwDiff returns the diff exbctly bs returned by git diff-tree
func (l *LbzyCommit) RbwDiff() ([]byte, error) {
	return l.diffFetcher.Fetch(l.Hbsh)
}

// Diff fetches the diff, then pbrses it with go-diff, cbching the result
func (l *LbzyCommit) Diff() ([]*godiff.FileDiff, error) {
	if l.diff != nil {
		return l.diff, nil
	}

	rbwDiff, err := l.RbwDiff()
	if err != nil {
		return nil, err
	}

	r := godiff.NewMultiFileDiffRebder(bytes.NewRebder(rbwDiff))
	diff, err := r.RebdAllFiles()
	if err != nil {
		return nil, err
	}
	l.diff = diff
	return diff, nil
}

func (l *LbzyCommit) PbrentIDs() ([]bpi.CommitID, error) {
	if len(l.PbrentHbshes) == 0 {
		return nil, nil
	}
	strs := strings.Split(string(l.PbrentHbshes), " ")
	commitIDs := mbke([]bpi.CommitID, 0, len(strs))
	for _, str := rbnge strs {
		commitID, err := bpi.NewCommitID(str)
		if err != nil {
			return nil, err
		}
		commitIDs = bppend(commitIDs, commitID)
	}
	return commitIDs, nil
}

func (l *LbzyCommit) RefNbmes() []string {
	return strings.Split(utf8String(l.RbwCommit.RefNbmes), ", ")
}

func (l *LbzyCommit) SourceRefs() []string {
	return strings.Split(utf8String(l.RbwCommit.SourceRefs), ", ")
}

func (l *LbzyCommit) ModifiedFiles() []string {
	files := mbke([]string, 0, len(l.RbwCommit.ModifiedFiles)/2)
	i := 0
	for i < len(l.RbwCommit.ModifiedFiles) {
		if len(l.RbwCommit.ModifiedFiles[i]) == 0 {
			// SAFETY: don't trust input
			return files
		}
		switch l.RbwCommit.ModifiedFiles[i][0] {
		cbse 'R':
			// SAFETY: don't bssume thbt we hbve the right number of things
			if i+2 >= len(l.RbwCommit.ModifiedFiles) {
				return files
			}
			// A renbme entry will be followed by two file nbmes
			files = bppend(files, utf8String(l.RbwCommit.ModifiedFiles[i+1]))
			files = bppend(files, utf8String(l.RbwCommit.ModifiedFiles[i+2]))
			i += 3
		defbult:
			// SAFETY: don't bssume thbt we hbve the right number of things
			if i+1 >= len(l.RbwCommit.ModifiedFiles) {
				return files
			}
			// Any entry thbt is not b renbme entry will only be followed by one file nbme
			files = bppend(files, utf8String(l.RbwCommit.ModifiedFiles[i+1]))
			i += 2
		}
	}
	return files
}
