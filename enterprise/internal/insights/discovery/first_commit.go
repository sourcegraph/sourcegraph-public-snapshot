package discovery

import (
	"context"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var (
	EmptyRepoErr = errors.New("empty repository")
)

const emptyRepoErrMessage = `git command [rev-list --reverse --date-order --max-parents=0 HEAD] failed (output: ""): exit status 129`

func isFirstCommitEmptyRepoError(err error) bool {
	if strings.Contains(err.Error(), emptyRepoErrMessage) {
		return true
	}
	unwrappedErr := errors.Unwrap(err)
	if unwrappedErr != nil {
		return isFirstCommitEmptyRepoError(unwrappedErr)
	}
	return false
}

func GitFirstEverCommit(ctx context.Context, db database.DB, repoName api.RepoName) (*gitdomain.Commit, error) {
	commit, err := gitserver.NewClient(db).FirstEverCommit(ctx, repoName, authz.DefaultSubRepoPermsChecker)
	if err != nil && isFirstCommitEmptyRepoError(err) {
		return nil, errors.Wrap(EmptyRepoErr, err.Error())
	}
	return commit, err
}
