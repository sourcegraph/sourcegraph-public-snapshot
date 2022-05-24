package git

import (
	"context"
	"io/fs"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
	"github.com/sourcegraph/sourcegraph/internal/vcs/util"
)

// Stat returns a FileInfo describing the named file at commit.
func Stat(ctx context.Context, db database.DB, checker authz.SubRepoPermissionChecker, repo api.RepoName, commit api.CommitID, path string) (fs.FileInfo, error) {
	if Mocks.Stat != nil {
		return Mocks.Stat(commit, path)
	}

	span, ctx := ot.StartSpanFromContext(ctx, "Git: Stat")
	span.SetTag("Commit", commit)
	span.SetTag("Path", path)
	defer span.Finish()

	if err := checkSpecArgSafety(string(commit)); err != nil {
		return nil, err
	}

	path = util.Rel(path)

	fi, err := gitserver.NewClient(db).LStat(ctx, checker, repo, commit, path)
	if err != nil {
		return nil, err
	}

	return fi, nil
}

// DevNullSHA 4b825dc642cb6eb9a060e54bf8d69288fbee4904 is `git hash-object -t
// tree /dev/null`, which is used as the base when computing the `git diff` of
// the root commit.
const DevNullSHA = "4b825dc642cb6eb9a060e54bf8d69288fbee4904"
