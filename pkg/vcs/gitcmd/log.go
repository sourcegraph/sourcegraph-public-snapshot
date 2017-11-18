package gitcmd

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"time"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
)

const (
	logFormatFlag  = "--format=format:%H%x00%aN%x00%aE%x00%at%x00%cN%x00%cE%x00%ct%x00%B%x00%P%x00"
	partsPerCommit = 9 // number of \x00-separated fields per commit
)

// parseCommitFromLog parses the next commit from data and returns the commit and the remaining
// data. The data arg is a byte array that contains NUL-separated log fields as formatted by
// logFormatFlag.
func parseCommitFromLog(forLogFormatFlag string, data []byte) (*vcs.Commit, []byte, error) {
	if forLogFormatFlag != logFormatFlag {
		// Ensure we're parsing our known format; require callers to be explicit.
		return nil, nil, errors.New("invalid log format flag")
	}

	parts := bytes.SplitN(data, []byte{'\x00'}, partsPerCommit+1)
	if len(parts) < partsPerCommit {
		return nil, nil, fmt.Errorf("invalid commit log entry: %q", parts)
	}

	// log outputs are newline separated, so all but the 1st commit ID part
	// has an erroneous leading newline.
	parts[0] = bytes.TrimPrefix(parts[0], []byte{'\n'})

	authorTime, err := strconv.ParseInt(string(parts[3]), 10, 64)
	if err != nil {
		return nil, nil, fmt.Errorf("parsing git commit author time: %s", err)
	}
	committerTime, err := strconv.ParseInt(string(parts[6]), 10, 64)
	if err != nil {
		return nil, nil, fmt.Errorf("parsing git commit committer time: %s", err)
	}

	var parents []vcs.CommitID
	if parentPart := parts[8]; len(parentPart) > 0 {
		parentIDs := bytes.Split(parentPart, []byte{' '})
		parents = make([]vcs.CommitID, len(parentIDs))
		for i, id := range parentIDs {
			parents[i] = vcs.CommitID(id)
		}
	}

	commit := &vcs.Commit{
		ID:        vcs.CommitID(parts[0]),
		Author:    vcs.Signature{Name: string(parts[1]), Email: string(parts[2]), Date: time.Unix(authorTime, 0).UTC()},
		Committer: &vcs.Signature{Name: string(parts[4]), Email: string(parts[5]), Date: time.Unix(committerTime, 0).UTC()},
		Message:   string(bytes.TrimSuffix(parts[7], []byte{'\n'})),
		Parents:   parents,
	}
	if len(parts) == 10 {
		data = parts[9]
	} else {
		data = nil
	}
	return commit, data, nil
}
