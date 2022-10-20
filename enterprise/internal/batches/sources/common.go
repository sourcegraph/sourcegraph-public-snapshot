package sources

import (
	"context"
	"fmt"

	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

// ChangesetNotFoundError is returned by LoadChangeset if the changeset
// could not be found on the codehost. This is only returned, if the
// changeset is actually not found. Other not found errors, such as
// repo not found should NOT raise this error, since it will cause
// the changeset to be marked as deleted.
type ChangesetNotFoundError struct {
	Changeset *Changeset
}

func (e ChangesetNotFoundError) Error() string {
	return fmt.Sprintf("Changeset with external ID %s not found", e.Changeset.Changeset.ExternalID)
}

func (e ChangesetNotFoundError) NonRetryable() bool { return true }

// ArchivableChangesetSource represents a changeset source that has a
// concept of archived repositories.
type ArchivableChangesetSource interface {
	ChangesetSource

	// IsArchivedPushError parses the given error output from `git push` to
	// detect whether the error was caused by the repository being archived.
	IsArchivedPushError(output string) bool
}

// A DraftChangesetSource can create draft changesets and undraft them.
type DraftChangesetSource interface {
	ChangesetSource

	// CreateDraftChangeset will create the Changeset on the source. If it already
	// exists, *Changeset will be populated and the return value will be
	// true.
	CreateDraftChangeset(context.Context, *Changeset) (bool, error)
	// UndraftChangeset will update the Changeset on the source to be not in draft mode anymore.
	UndraftChangeset(context.Context, *Changeset) error
}

type ForkableChangesetSource interface {
	ChangesetSource

	// GetNamespaceFork returns a repo pointing to a fork of the given repo in
	// the given namespace, ensuring that the fork exists and is a fork of the
	// target repo.
	GetNamespaceFork(ctx context.Context, targetRepo *types.Repo, namespace string) (*types.Repo, error)

	// GetUserFork returns a repo pointing to a fork of the given repo in the
	// currently authenticated user's namespace.
	GetUserFork(ctx context.Context, targetRepo *types.Repo) (*types.Repo, error)
}

// A ChangesetSource can load the latest state of a list of Changesets.
type ChangesetSource interface {
	// GitserverPushConfig returns an authenticated push config used for pushing
	// commits to the code host.
	GitserverPushConfig(*types.Repo) (*protocol.PushConfig, error)
	// WithAuthenticator returns a copy of the original Source configured to use
	// the given authenticator, provided that authenticator type is supported by
	// the code host.
	WithAuthenticator(auth.Authenticator) (ChangesetSource, error)
	// ValidateAuthenticator validates the currently set authenticator is usable.
	// Returns an error, when validating the Authenticator yielded an error.
	ValidateAuthenticator(ctx context.Context) error

	// LoadChangeset loads the given Changeset from the source and updates it.
	// If the Changeset could not be found on the source, a ChangesetNotFoundError is returned.
	LoadChangeset(context.Context, *Changeset) error
	// CreateChangeset will create the Changeset on the source. If it already
	// exists, *Changeset will be populated and the return value will be
	// true.
	CreateChangeset(context.Context, *Changeset) (bool, error)
	// CloseChangeset will close the Changeset on the source, where "close"
	// means the appropriate final state on the codehost (e.g. "declined" on
	// Bitbucket Server).
	CloseChangeset(context.Context, *Changeset) error
	// UpdateChangeset can update Changesets.
	UpdateChangeset(context.Context, *Changeset) error
	// ReopenChangeset will reopen the Changeset on the source, if it's closed.
	// If not, it's a noop.
	ReopenChangeset(context.Context, *Changeset) error
	// CreateComment posts a comment on the Changeset.
	CreateComment(context.Context, *Changeset, string) error
	// MergeChangeset merges a Changeset on the code host, if in a mergeable state.
	// If squash is true, and the code host supports squash merges, the source
	// must attempt a squash merge. Otherwise, it is expected to perform a regular
	// merge. If the changeset cannot be merged, because it is in an unmergeable
	// state, ChangesetNotMergeableError must be returned.
	MergeChangeset(ctx context.Context, ch *Changeset, squash bool) error
}

// ChangesetNotMergeableError is returned by MergeChangeset if the changeset
// could not be merged on the codehost, because some precondition is not met. This
// is only returned, if the changeset is not mergeable. Other errors, such as
// network or auth errors should NOT raise this error.
type ChangesetNotMergeableError struct {
	ErrorMsg string
}

func (e ChangesetNotMergeableError) Error() string {
	return fmt.Sprintf("changeset cannot be merged:\n%s", e.ErrorMsg)
}

func (e ChangesetNotMergeableError) NonRetryable() bool { return true }

// A Changeset of an existing Repo.
type Changeset struct {
	Title   string
	Body    string
	HeadRef string
	BaseRef string

	// RemoteRepo is the repository the branch will be pushed to. This must be
	// the same as TargetRepo if forking is not in use.
	RemoteRepo *types.Repo
	// TargetRepo is the repository in which the pull or merge request will be
	// opened.
	TargetRepo *types.Repo

	*btypes.Changeset
}

// IsOutdated returns true when the attributes of the nested
// batches.Changeset do not match the attributes (title, body, ...) set on
// the Changeset.
func (c *Changeset) IsOutdated() (bool, error) {
	currentTitle, err := c.Changeset.Title()
	if err != nil {
		return false, err
	}

	if currentTitle != c.Title {
		return true, nil
	}

	currentBody, err := c.Changeset.Body()
	if err != nil {
		return false, err
	}

	if currentBody != c.Body {
		return true, nil
	}

	currentBaseRef, err := c.Changeset.BaseRef()
	if err != nil {
		return false, err
	}

	if gitdomain.EnsureRefPrefix(currentBaseRef) != gitdomain.EnsureRefPrefix(c.BaseRef) {
		return true, nil
	}

	return false, nil
}
