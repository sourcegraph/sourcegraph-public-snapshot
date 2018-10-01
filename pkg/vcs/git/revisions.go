package git

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

// IsAbsoluteRevision checks if the revision is a git OID SHA string.
//
// Note: This doesn't mean the SHA exists in a repository, nor does it mean it
// isn't a ref. Git allows 40-char hexadecimal strings to be references.
func IsAbsoluteRevision(s string) bool {
	if len(s) != 40 {
		return false
	}
	for _, r := range s {
		if !(('0' <= r && r <= '9') ||
			('a' <= r && r <= 'f') ||
			('A' <= r && r <= 'F')) {
			return false
		}
	}
	return true
}

func ensureAbsCommit(commitID api.CommitID) {
	// We don't want to even be running commands on non-absolute
	// commit IDs if we can avoid it, because we can't cache the
	// expensive part of those computations.
	if !IsAbsoluteRevision(string(commitID)) {
		panic(fmt.Errorf("non-absolute commit ID: %q", commitID))
	}
}

type ResolveRevisionOptions struct {
	NoEnsureRevision bool // do not try to fetch from remote if revision doesn't exist locally
}

// ResolveRevision will return the absolute commit for a commit-ish spec. If spec is empty, HEAD is
// used.
//
// Error cases:
// * Repo does not exist: vcs.RepoNotExistError
// * Commit does not exist: RevisionNotFoundError
// * Empty repository: RevisionNotFoundError
// * Other unexpected errors.
//
// The remoteURLFunc is called to get the Git remote URL if it's not set in r and if it is
// needed. The Git remote URL is only required if the gitserver doesn't already contain a clone of
// the repository or if the revision must be fetched from the remote. This only happens when calling
// ResolveRevision.
func ResolveRevision(ctx context.Context, repo gitserver.Repo, remoteURLFunc func() (string, error), spec string, opt *ResolveRevisionOptions) (api.CommitID, error) {
	if Mocks.ResolveRevision != nil {
		return Mocks.ResolveRevision(spec, opt)
	}

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

	cmd := gitserver.DefaultClient.Command("git", "rev-parse", spec)
	cmd.Repo = repo
	commit, err := runRevParse(ctx, cmd, spec)
	if err == nil {
		return commit, nil
	}

	tryAgain := func(err error) bool {
		// We need to try again with remote URL set so we can clone.
		if vcs.IsRepoNotExist(err) {
			return true
		}

		// We need to try again with the remote URL set so we can fetch.
		if IsRevisionNotFound(err) {
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
	doEnsureRevision := IsRevisionNotFound(err)

	var remoteURL string
	if remoteURLFunc != nil {
		remoteURL, err = remoteURLFunc()
		if err != nil {
			return "", err
		}
	}

	cmd = gitserver.DefaultClient.Command("git", "rev-parse", spec)
	cmd.Repo = gitserver.Repo{Name: repo.Name, URL: remoteURL}
	if doEnsureRevision {
		cmd.EnsureRevision = spec
	}
	return runRevParse(ctx, cmd, spec)
}

// runRevParse sends the git rev-parse command to gitserver. It interprets
// missing revision responses and converts them into RevisionNotFoundError.
func runRevParse(ctx context.Context, cmd *gitserver.Cmd, spec string) (api.CommitID, error) {
	stdout, stderr, err := cmd.DividedOutput(ctx)
	if err != nil {
		if vcs.IsRepoNotExist(err) {
			return "", err
		}
		if bytes.Contains(stderr, []byte("unknown revision")) {
			return "", &RevisionNotFoundError{Repo: cmd.Name, Spec: spec}
		}
		return "", errors.WithMessage(err, fmt.Sprintf("git command %v failed (stderr: %q)", cmd.Args, stderr))
	}
	commit := api.CommitID(bytes.TrimSpace(stdout))
	if !IsAbsoluteRevision(string(commit)) {
		if commit == "HEAD" {
			// We don't verify the existence of HEAD (see above comments), but
			// if HEAD doesn't point to anything git just returns `HEAD` as the
			// output of rev-parse. An example where this occurs is an empty
			// repository.
			return "", &RevisionNotFoundError{Repo: cmd.Name, Spec: spec}
		}
		return "", fmt.Errorf("ResolveRevision: got bad commit %q for repo %q at revision %q", commit, cmd.Name, spec)
	}
	return commit, nil
}
