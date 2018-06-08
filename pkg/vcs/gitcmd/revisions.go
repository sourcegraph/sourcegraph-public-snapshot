package gitcmd

import (
	"bytes"
	"context"
	"fmt"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/gitserver"
	"github.com/sourcegraph/sourcegraph/pkg/vcs"
)

func (r *Repository) ensureAbsCommit(commitID api.CommitID) {
	// We don't want to even be running commands on non-absolute
	// commit IDs if we can avoid it, because we can't cache the
	// expensive part of those computations.
	if !vcs.IsAbsoluteRevision(string(commitID)) {
		panic(fmt.Errorf("non-absolute commit ID: %q on %s", commitID, r.String()))
	}
}

// ResolveRevision will return the absolute commit for a commit-ish spec. If spec is empty, HEAD is
// used.
//
// Error cases:
// * Repo does not exist: vcs.RepoNotExistError
// * Commit does not exist: RevisionNotFoundError
// * Empty repository: RevisionNotFoundError
// * Other unexpected errors.
func (r *Repository) ResolveRevision(ctx context.Context, spec string, opt *vcs.ResolveRevisionOptions) (api.CommitID, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "Git: ResolveRevision")
	span.SetTag("Spec", spec)
	span.SetTag("Opt", fmt.Sprintf("%+v", opt))
	defer span.Finish()

	if err := checkSpecArgSafety(spec); err != nil {
		return "", err
	}
	if spec == "" {
		spec = "HEAD"
	}
	if spec != "HEAD" {
		// "git rev-parse HEAD^0" is slower than "git rev-parse HEAD"
		// since it checks that the resolved git object exists. We can
		// assume it exists for HEAD, but for other commits we should
		// check.
		spec = spec + "^0"
	}

	cmd := r.command("git", "rev-parse", spec)
	commit, err := r.runRevParse(ctx, cmd, spec)
	if err == nil {
		return commit, nil
	}

	tryAgain := func(err error) bool {
		// We need to try again with remote URL set so we can clone.
		if vcs.IsRepoNotExist(err) {
			return true
		}

		// We need to try again with the remote URL set so we can fetch.
		if vcs.IsRevisionNotFound(err) {
			// If we didn't find HEAD, then the repo is empty.
			if spec == "HEAD" {
				return false
			}

			// We can also disable enqueuing an update.
			if opt != nil && opt.NoEnsureRevision {
				return false
			}

			return true
		}

		return false
	}

	if !tryAgain(err) {
		return "", err
	}
	doEnsureRevision := vcs.IsRevisionNotFound(err)

	r.once.Do(func() {
		r.remoteURL, r.remoteURLErr = r.remoteURLFunc()
	})
	if r.remoteURLErr != nil {
		return "", r.remoteURLErr
	}

	cmd = r.command("git", "rev-parse", spec)
	cmd.Repo = gitserver.Repo{Name: r.repoURI, URL: r.remoteURL}
	if doEnsureRevision {
		cmd.EnsureRevision = spec
	}
	return r.runRevParse(ctx, cmd, spec)
}

// runRevParse sends the git rev-parse command to gitserver. It interprets
// missing revision responses and converts them into RevisionNotFoundError.
func (r *Repository) runRevParse(ctx context.Context, cmd *gitserver.Cmd, spec string) (api.CommitID, error) {
	stdout, stderr, err := cmd.DividedOutput(ctx)
	if err != nil {
		if vcs.IsRepoNotExist(err) {
			return "", err
		}
		if bytes.Contains(stderr, []byte("unknown revision")) {
			return "", &vcs.RevisionNotFoundError{Repo: r.repoURI, Spec: spec}
		}
		return "", errors.WithMessage(err, fmt.Sprintf("exec `git rev-parse` failed with stderr: %s", stderr))
	}
	commit := api.CommitID(bytes.TrimSpace(stdout))
	if !vcs.IsAbsoluteRevision(string(commit)) {
		if commit == "HEAD" {
			// We don't verify the existence of HEAD (see above comments), but
			// if HEAD doesn't point to anything git just returns `HEAD` as the
			// output of rev-parse. An example where this occurs is an empty
			// repository.
			return "", &vcs.RevisionNotFoundError{Repo: r.repoURI, Spec: spec}
		}
		return "", fmt.Errorf("ResolveRevision: got bad commit %q for repo %q at revision %q", commit, r.repoURI, spec)
	}
	return commit, nil
}
