package internal

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/common"
	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/git"
	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/gitserverfs"
	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/vcssyncer"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/fileutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func postRepoFetchActions(
	ctx context.Context,
	logger log.Logger,
	fs gitserverfs.FS,
	db database.DB,
	backend git.GitBackend,
	shardID string,
	repo api.RepoName,
	dir common.GitDir,
	syncer vcssyncer.VCSSyncer,
) (errs error) {
	// Note: We use a multi error in this function to try to make as many of the
	// post repo fetch actions succeed.

	if err := git.RemoveBadRefs(ctx, dir); err != nil {
		errs = errors.Append(errs, errors.Wrapf(err, "failed to remove bad refs for repo %q", repo))
	}

	if err := git.SetRepositoryType(ctx, backend.Config(), syncer.Type()); err != nil {
		errs = errors.Append(errs, errors.Wrapf(err, "failed to set repository type for repo %q", repo))
	}

	if err := git.SetGitAttributes(dir); err != nil {
		errs = errors.Append(errs, errors.Wrap(err, "setting git attributes"))
	}

	if err := gitSetAutoGC(ctx, backend.Config()); err != nil {
		errs = errors.Append(errs, errors.Wrap(err, "setting git gc mode"))
	}

	// Update the last-changed stamp on disk.
	if err := setLastChanged(ctx, logger, dir, backend); err != nil {
		errs = errors.Append(errs, errors.Wrap(err, "failed to update last changed time"))
	}

	// Successfully updated, best-effort calculation of the repo size.
	repoSizeBytes, err := fs.DirSize(dir.Path())
	if err != nil {
		errs = errors.Append(errs, errors.Wrap(err, "failed to calculate repo size"))
	} else if err := db.GitserverRepos().SetRepoSize(ctx, repo, repoSizeBytes, shardID); err != nil {
		errs = errors.Append(errs, errors.Wrap(err, "failed to set repo size"))
	}

	return errs
}

// gitSetAutoGC will set the value of gc.auto. If GC is managed by Sourcegraph
// the value will be 0 (disabled), otherwise if managed by git we will unset
// it to rely on default (on) or global config.
//
// The purpose is to avoid repository corruption which can happen if several
// git-gc operations are running at the same time.
func gitSetAutoGC(ctx context.Context, c git.GitConfigBackend) error {
	switch gitGCMode {
	case gitGCModeGitAutoGC, gitGCModeJanitorAutoGC:
		return c.Unset(ctx, "gc.auto")

	case gitGCModeMaintenance:
		return c.Set(ctx, "gc.auto", "0")

	default:
		// should not happen
		panic(fmt.Sprintf("non exhaustive switch for gitGCMode: %d", gitGCMode))
	}
}

// setLastChanged discerns an approximate last-changed timestamp for a
// repository. This can be approximate; it's used to determine how often we
// should run `git fetch`, but is not relied on strongly. The basic plan
// is as follows: If a repository has never had a timestamp before, we
// guess that the right stamp is *probably* the timestamp of the most
// chronologically-recent commit. If there are no commits, we just use the
// current time because that's probably usually a temporary state.
//
// If a timestamp already exists, we want to update it if and only if
// the set of references (as determined by `git show-ref`) has changed.
//
// To accomplish this, we assert that the file `sg_refhash_v2` in the git
// directory should, if it exists, contain a hash of the output of
// `git show-ref`, and have a timestamp of "the last time this changed",
// except that if we're creating that file for the first time, we set
// it to the timestamp of the top commit. We then compute the hash of
// the show-ref output, and store it in the file if and only if it's
// different from the current contents.
//
// If show-ref fails, we use rev-list to determine whether that's just
// an empty repository (not an error) or some kind of actual error
// that is possibly causing our data to be incorrect, which should
// be reported.
func setLastChanged(ctx context.Context, logger log.Logger, dir common.GitDir, backend git.GitBackend) error {
	hashFile := dir.Path("sg_refhash_v2")

	// Best effort delete the old refhash file.
	_ = os.Remove(dir.Path("sg_refhash"))

	hash, err := backend.RefHash(ctx)
	if err != nil {
		return errors.Wrapf(err, "computing ref hash failed for %s", dir)
	}

	var stamp time.Time
	if _, err := os.Stat(hashFile); os.IsNotExist(err) {
		// This is the first time we are calculating the hash. Give a more
		// approriate timestamp for sg_refhash_v2 than the current time.
		stamp, err = backend.LatestCommitTimestamp(ctx)
		if err != nil {
			logger.Warn("failed to get latest commit timestamp, using current time", log.Error(err))
			stamp = time.Now().UTC()
		}
	}

	_, err = fileutil.UpdateFileIfDifferent(hashFile, hash)
	if err != nil {
		return errors.Wrapf(err, "failed to update %s", hashFile)
	}

	// If stamp is non-zero we have a more approriate mtime.
	if !stamp.IsZero() {
		err = os.Chtimes(hashFile, stamp, stamp)
		if err != nil {
			return errors.Wrapf(err, "failed to set mtime to the lastest commit timestamp for %s", dir)
		}
	}

	return nil
}
