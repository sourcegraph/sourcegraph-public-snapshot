package gitserver

import (
	"context"
	"errors"
	"strings"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/store"
)

type Status uint8

const (
	Unchanged Status = iota
	Modified
	Deleted
	Added
)

func DiffFileStatus(ctx context.Context, store store.Store, repositoryID int, baseCommit, headCommit string) (map[string]Status, error) {
	output, err := execGitCommand(ctx, store, repositoryID, "diff", "--name-status", baseCommit, headCommit)
	if err != nil {
		return nil, err
	}

	statuses := make(map[string]Status)
	for _, line := range strings.Split(output, "\n") {
		fields := strings.Fields(line)
		if len(fields) == 0 || len(fields[0]) == 0 {
			continue
		}
		switch fields[0][0] {
		case 'M':
			statuses[fields[1]] = Modified
		case 'D':
			statuses[fields[1]] = Deleted
		case 'A':
			statuses[fields[1]] = Added
		case 'R':
			statuses[fields[1]] = Deleted
			statuses[fields[2]] = Added
		case 'C':
			statuses[fields[2]] = Added
		default:
			return nil, errors.New("unknown git diff file status")
		}
	}

	return statuses, nil
}
