package gitcmd

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/vcs"
)

const (
	partsPerCommit = 10 // number of \x00-separated fields per commit

	// include refs (slow on repos with many refs)
	logFormatWithRefs = "--format=format:%H%x00%D%x00%aN%x00%aE%x00%at%x00%cN%x00%cE%x00%ct%x00%B%x00%P%x00"

	// don't include refs (faster, should be used if refs are not needed)
	logFormatWithoutRefs = "--format=format:%H%x00%x00%aN%x00%aE%x00%at%x00%cN%x00%cE%x00%ct%x00%B%x00%P%x00"
)

// parseCommitFromLog parses the next commit from data and returns the commit and the remaining
// data. The data arg is a byte array that contains NUL-separated log fields as formatted by
// logFormatFlag.
func parseCommitFromLog(data []byte) (commit *vcs.Commit, refs []string, patch []byte, err error) {
	parts := bytes.SplitN(data, []byte{'\x00'}, partsPerCommit+1)
	if len(parts) < partsPerCommit {
		return nil, nil, nil, fmt.Errorf("invalid commit log entry: %q", parts)
	}

	// log outputs are newline separated, so all but the 1st commit ID part
	// has an erroneous leading newline.
	parts[0] = bytes.TrimPrefix(parts[0], []byte{'\n'})

	authorTime, err := strconv.ParseInt(string(parts[4]), 10, 64)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("parsing git commit author time: %s", err)
	}
	committerTime, err := strconv.ParseInt(string(parts[7]), 10, 64)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("parsing git commit committer time: %s", err)
	}

	var parents []api.CommitID
	if parentPart := parts[9]; len(parentPart) > 0 {
		parentIDs := bytes.Split(parentPart, []byte{' '})
		parents = make([]api.CommitID, len(parentIDs))
		for i, id := range parentIDs {
			parents[i] = api.CommitID(id)
		}
	}

	if len(parts[1]) > 0 {
		refs = strings.Split(string(parts[1]), ", ")
	}

	commit = &vcs.Commit{
		ID:        api.CommitID(parts[0]),
		Author:    vcs.Signature{Name: string(parts[2]), Email: string(parts[3]), Date: time.Unix(authorTime, 0).UTC()},
		Committer: &vcs.Signature{Name: string(parts[5]), Email: string(parts[6]), Date: time.Unix(committerTime, 0).UTC()},
		Message:   string(bytes.TrimSuffix(parts[8], []byte{'\n'})),
		Parents:   parents,
	}

	if len(parts) == partsPerCommit+1 {
		patch = parts[10]
	}

	return commit, refs, patch, nil
}

// onelineCommit contains (a subset of the) information about a commit returned
// by `git log --oneline --source`.
type onelineCommit struct {
	sha1      string // sha1 commit ID
	sourceRef string // `git log --source` source ref
}

// parseCommitsFromOnelineLog parses the commits from the output of:
//
//   git log --oneline -z --source --no-patch
func parseCommitsFromOnelineLog(data []byte) (commits []*onelineCommit, err error) {
	entries := bytes.Split(data, []byte{'\x00'})
	for _, e := range entries {
		if len(e) == 0 {
			continue
		}

		// Format: (40-char SHA) \t (source ref)? 'log size '
		if len(e) <= 40 {
			return commits, fmt.Errorf("parsing git oneline commit: short entry: %q", e)
		}
		sha1 := e[:40]
		i := bytes.Index(e, []byte{' '})
		if i == -1 {
			return commits, fmt.Errorf("parsing git oneline commit: no ' ': %q", e)
		}
		sourceRef := e[41:i]
		commits = append(commits, &onelineCommit{
			sha1:      string(sha1),
			sourceRef: string(sourceRef),
		})
	}
	return commits, nil
}
