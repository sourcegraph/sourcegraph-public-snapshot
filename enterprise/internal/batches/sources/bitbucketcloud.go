package sources

import (
	"context"
	"strconv"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketcloud"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

type BitbucketCloudSource struct {
	client *bitbucketcloud.Client
}

var (
	_ ForkableChangesetSource = BitbucketCloudSource{}
)

func NewBitbucketCloudSource(svc *types.ExternalService, cf *httpcli.Factory) (*BitbucketCloudSource, error) {
	var c schema.BitbucketCloudConnection
	if err := jsonc.Unmarshal(svc.Config, &c); err != nil {
		return nil, errors.Wrapf(err, "external service id=%d", svc.ID)
	}
	return newBitbucketCloudSource(&c, cf)
}

func newBitbucketCloudSource(c *schema.BitbucketCloudConnection, cf *httpcli.Factory) (*BitbucketCloudSource, error) {
	if cf == nil {
		cf = httpcli.ExternalClientFactory
	}

	// No options to provide here, since Bitbucket Cloud doesn't support custom
	// certificates, unlike the other
	cli, err := cf.Doer()
	if err != nil {
		return nil, errors.Wrap(err, "creating external client")
	}

	client, err := bitbucketcloud.NewClient(c, cli)
	if err != nil {
		return nil, errors.Wrap(err, "creating Bitbucket Cloud client")
	}

	return &BitbucketCloudSource{client: client}, nil
}

// GitserverPushConfig returns an authenticated push config used for pushing
// commits to the code host.
func (s BitbucketCloudSource) GitserverPushConfig(ctx context.Context, store database.ExternalServiceStore, repo *types.Repo) (*protocol.PushConfig, error) {
	return gitserverPushConfig(ctx, store, repo, s.client.Auth)
}

// WithAuthenticator returns a copy of the original Source configured to use the
// given authenticator, provided that authenticator type is supported by the
// code host.
func (s BitbucketCloudSource) WithAuthenticator(a auth.Authenticator) (ChangesetSource, error) {
	switch a.(type) {
	case *auth.BasicAuth,
		*auth.BasicAuthWithSSH:
		break

	default:
		return nil, newUnsupportedAuthenticatorError("BitbucketCloudSource", a)
	}

	return &BitbucketCloudSource{client: s.client.WithAuthenticator(a)}, nil
}

// ValidateAuthenticator validates the currently set authenticator is usable.
// Returns an error, when validating the Authenticator yielded an error.
func (s BitbucketCloudSource) ValidateAuthenticator(ctx context.Context) error {
	return s.client.Ping(ctx)
}

// LoadChangeset loads the given Changeset from the source and updates it. If
// the Changeset could not be found on the source, a ChangesetNotFoundError is
// returned.
func (s BitbucketCloudSource) LoadChangeset(ctx context.Context, cs *Changeset) error {
	repo := cs.TargetRepo.Metadata.(*bitbucketcloud.Repo)
	number, err := strconv.Atoi(cs.ExternalID)
	if err != nil {
		return errors.Wrapf(err, "converting external ID %q", cs.ExternalID)
	}

	pr, err := s.client.GetPullRequest(ctx, repo, int64(number))
	if err != nil {
		if errcode.IsNotFound(err) {
			return ChangesetNotFoundError{Changeset: cs}
		}
		return errors.Wrap(err, "getting pull request")
	}

	// TODO: an equivalent of the loadPullRequestData call to get comments,
	// build statuses, and approvals.

	if err := cs.SetMetadata(pr); err != nil {
		return errors.Wrap(err, "setting metadata")
	}

	return nil
}

// CreateChangeset will create the Changeset on the source. If it already
// exists, *Changeset will be populated and the return value will be true.
func (s BitbucketCloudSource) CreateChangeset(context.Context, *Changeset) (bool, error) {
	panic("not implemented") // TODO: Implement
}

// CloseChangeset will close the Changeset on the source, where "close"
// means the appropriate final state on the codehost (e.g. "declined" on
// Bitbucket Server).
func (s BitbucketCloudSource) CloseChangeset(_ context.Context, _ *Changeset) error {
	panic("not implemented") // TODO: Implement
}

// UpdateChangeset can update Changesets.
func (s BitbucketCloudSource) UpdateChangeset(_ context.Context, _ *Changeset) error {
	panic("not implemented") // TODO: Implement
}

// ReopenChangeset will reopen the Changeset on the source, if it's closed.
// If not, it's a noop.
func (s BitbucketCloudSource) ReopenChangeset(_ context.Context, _ *Changeset) error {
	panic("not implemented") // TODO: Implement
}

// CreateComment posts a comment on the Changeset.
func (s BitbucketCloudSource) CreateComment(_ context.Context, _ *Changeset, _ string) error {
	panic("not implemented") // TODO: Implement
}

// MergeChangeset merges a Changeset on the code host, if in a mergeable state.
// If squash is true, and the code host supports squash merges, the source
// must attempt a squash merge. Otherwise, it is expected to perform a regular
// merge. If the changeset cannot be merged, because it is in an unmergeable
// state, ChangesetNotMergeableError must be returned.
func (s BitbucketCloudSource) MergeChangeset(ctx context.Context, ch *Changeset, squash bool) error {
	panic("not implemented") // TODO: Implement
}

// GetNamespaceFork returns a repo pointing to a fork of the given repo in
// the given namespace, ensuring that the fork exists and is a fork of the
// target repo.
func (s BitbucketCloudSource) GetNamespaceFork(ctx context.Context, targetRepo *types.Repo, namespace string) (*types.Repo, error) {
	panic("not implemented") // TODO: Implement
}

// GetUserFork returns a repo pointing to a fork of the given repo in the
// currently authenticated user's namespace.
func (s BitbucketCloudSource) GetUserFork(ctx context.Context, targetRepo *types.Repo) (*types.Repo, error) {
	panic("not implemented") // TODO: Implement
}
