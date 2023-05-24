package sources

import (
	"context"
	"net/url"

	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gerrit"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

type GerritSource struct {
	client gerrit.Client
}

var _ ForkableChangesetSource = GerritSource{}

func NewGerritSource(ctx context.Context, svc *types.ExternalService, cf *httpcli.Factory) (*GerritSource, error) {
	rawConfig, err := svc.Config.Decrypt(ctx)
	if err != nil {
		return nil, errors.Errorf("external service id=%d config error: %s", svc.ID, err)
	}
	var c schema.GerritConnection
	if err := jsonc.Unmarshal(rawConfig, &c); err != nil {
		return nil, errors.Wrapf(err, "external service id=%d", svc.ID)
	}

	if cf == nil {
		cf = httpcli.ExternalClientFactory
	}

	cli, err := cf.Doer()
	if err != nil {
		return nil, errors.Wrap(err, "creating external client")
	}

	gerritURL, err := url.Parse(c.Url)
	if err != nil {
		return nil, errors.Wrap(err, "parsing Gerrit URL")
	}

	client, err := gerrit.NewClient(svc.URN(), gerritURL, &gerrit.AccountCredentials{Username: c.Username, Password: c.Password}, cli)
	if err != nil {
		return nil, errors.Wrap(err, "creating Gerrit client")
	}

	return &GerritSource{client: client}, nil
}

// GitserverPushConfig returns an authenticated push config used for pushing
// commits to the code host.
func (s GerritSource) GitserverPushConfig(repo *types.Repo) (*protocol.PushConfig, error) {
	return GitserverPushConfig(repo, s.client.Authenticator())
}

// WithAuthenticator returns a copy of the original Source configured to use the
// given authenticator, provided that authenticator type is supported by the
// code host.
func (s GerritSource) WithAuthenticator(a auth.Authenticator) (ChangesetSource, error) {
	client, err := s.client.WithAuthenticator(a)
	if err != nil {
		return nil, err
	}

	return &GerritSource{client: client}, nil
}

// ValidateAuthenticator validates the currently set authenticator is usable.
// Returns an error, when validating the Authenticator yielded an error.
func (s GerritSource) ValidateAuthenticator(ctx context.Context) error {
	_, err := s.client.GetAuthenticatedUserAccount(ctx)
	return err
}

// LoadChangeset loads the given Changeset from the source and updates it. If
// the Changeset could not be found on the source, a ChangesetNotFoundError is
// returned.
func (s GerritSource) LoadChangeset(ctx context.Context, cs *Changeset) error {
	pr, err := s.client.GetChange(ctx, cs.ExternalID)
	if err != nil {
		if errcode.IsNotFound(err) {
			return ChangesetNotFoundError{Changeset: cs}
		}
		return errors.Wrap(err, "getting pull request")
	}

	return errors.Wrap(s.setChangesetMetadata(ctx, pr, cs), "setting Gerrit changeset metadata")
}

// CreateChangeset will create the Changeset on the source. If it already
// exists, *Changeset will be populated and the return value will be true.
// noop, Gerrit creates changes through commits directly
func (s GerritSource) CreateChangeset(_ context.Context, _ *Changeset) (bool, error) {
	return true, nil
}

// CreateDraftChangeset creates the given changeset on the code host in draft mode.
// no-op, Gerrit creates changes through commits directly
func (s GerritSource) CreateDraftChangeset(context.Context, *Changeset) (bool, error) {
	return true, nil
}

// UndraftChangeset will update the Changeset on the source to be not in draft mode anymore.
// Noop, Gerrit creates changes through commits directly
func (s GerritSource) UndraftChangeset(context.Context, *Changeset) error {
	return nil
}

// CloseChangeset will close the Changeset on the source, where "close"
// means the appropriate final state on the codehost (e.g. "abandoned" on
// Gerrit).
func (s GerritSource) CloseChangeset(ctx context.Context, cs *Changeset) error {
	updated, err := s.client.AbandonChange(ctx, cs.ExternalID)
	if err != nil {
		return errors.Wrap(err, "abandoning pull request")
	}

	return errors.Wrap(s.setChangesetMetadata(ctx, updated, cs), "setting Gerrit changeset metadata")
}

// UpdateChangeset can update Changesets.
// Noop, Gerrit updates changes through git push directly
func (s GerritSource) UpdateChangeset(context.Context, *Changeset) error {
	return nil
}

// ReopenChangeset will reopen the Changeset on the source, if it's closed.
// If not, it's a noop.
func (s GerritSource) ReopenChangeset(ctx context.Context, cs *Changeset) error {
	updated, err := s.client.RestoreChange(ctx, cs.ExternalID)
	if err != nil {
		return errors.Wrap(err, "updating pull request")
	}

	return errors.Wrap(s.setChangesetMetadata(ctx, updated, cs), "setting Gerrit changeset metadata")
}

// CreateComment posts a comment on the Changeset.
func (s GerritSource) CreateComment(ctx context.Context, cs *Changeset, comment string) error {
	return s.client.WriteReviewComment(ctx, cs.ExternalID, gerrit.ChangeReviewComment{
		Message: comment,
	})
}

// MergeChangeset merges a Changeset on the code host, if in a mergeable state.
// If squash is true, and the code host supports squash merges, the source
// must attempt a squash merge. Otherwise, it is expected to perform a regular
// merge. If the changeset cannot be merged, because it is in an unmergeable
// state, ChangesetNotMergeableError must be returned.
// Gerrit changes are always single commit, so squash does not matter.
func (s GerritSource) MergeChangeset(ctx context.Context, cs *Changeset, squash bool) error {
	updated, err := s.client.SubmitChange(ctx, cs.ExternalID)
	if err != nil {
		if errcode.IsNotFound(err) {
			return errors.Wrap(err, "merging pull request")
		}
		return ChangesetNotMergeableError{ErrorMsg: err.Error()}
	}

	return errors.Wrap(s.setChangesetMetadata(ctx, updated, cs), "setting Gerrit changeset metadata")
}

// GetFork returns a repo pointing to a fork of the target repo, ensuring that the fork
// exists and creating it if it doesn't. If namespace is not provided, the original namespace is used.
// If name is not provided, the fork will be named with the default Sourcegraph convention:
// "${original-namespace}-${original-name}"
// Noop, Gerrit does not support Changes from forks.
func (s GerritSource) GetFork(_ context.Context, _ *types.Repo, _, _ *string) (*types.Repo, error) {
	return nil, nil
}

func (s GerritSource) setChangesetMetadata(_ context.Context, pr *gerrit.Change, cs *Changeset) error {
	if err := cs.SetMetadata(pr); err != nil {
		return errors.Wrap(err, "setting changeset metadata")
	}

	return nil
}
