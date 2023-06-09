package sources

import (
	"context"
	"fmt"

	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

type PerforceSource struct {
	server          schema.PerforceConnection
	gitServerClient gitserver.Client
}

func NewPerforceSource(ctx context.Context, svc *types.ExternalService, _ *httpcli.Factory) (*PerforceSource, error) {
	rawConfig, err := svc.Config.Decrypt(ctx)
	if err != nil {
		return nil, errors.Errorf("external service id=%d config error: %s", svc.ID, err)
	}
	var c schema.PerforceConnection
	if err := jsonc.Unmarshal(rawConfig, &c); err != nil {
		return nil, errors.Wrapf(err, "external service id=%d", svc.ID)
	}

	return &PerforceSource{server: c, gitServerClient: gitserver.NewClient()}, nil
}

// GitserverPushConfig returns an authenticated push config used for pushing commits to the code host.
func (s PerforceSource) GitserverPushConfig(_ *types.Repo) (*protocol.PushConfig, error) {
	// Return a PushConfig with a crafted URL that includes the Perforce scheme and the credentials
	// The perforce scheme will tell `createCommitFromPatch` that this repo is a Perforce repo
	// so it can handle it differently from Git repos.
	// TODO: @peterguy include the depot in the path component. Not sure where to get that from yet. It's not api.Repo.
	return &protocol.PushConfig{
		RemoteURL: fmt.Sprintf("perforce://%s:%s@%s", s.server.P4User, s.server.P4Passwd, s.server.P4Port),
	}, nil
}

// WithAuthenticator returns a copy of the original Source configured to use the
// given authenticator, provided that authenticator type is supported by the
// code host.
func (s PerforceSource) WithAuthenticator(_ auth.Authenticator) (ChangesetSource, error) {
	return s, nil
}

// ValidateAuthenticator validates the currently set authenticator is usable.
// Returns an error, when validating the Authenticator yielded an error.
func (s PerforceSource) ValidateAuthenticator(_ context.Context) error {
	return nil
}

// LoadChangeset loads the given Changeset from the source and updates it. If
// the Changeset could not be found on the source, a ChangesetNotFoundError is
// returned.
func (s PerforceSource) LoadChangeset(ctx context.Context, cs *Changeset) error {
	cl, err := s.gitServerClient.P4GetChangelist(ctx, cs.ExternalID)
	if err != nil {
		return errors.Wrap(err, "getting changelist")
	}
	return errors.Wrap(s.setChangesetMetadata(&cl, cs), "setting perforce changeset metadata")
}

// CreateChangeset will create the Changeset on the source. If it already
// exists, *Changeset will be populated and the return value will be true.
func (s PerforceSource) CreateChangeset(ctx context.Context, cs *Changeset) (bool, error) {
	return false, s.LoadChangeset(ctx, cs)
}

// CreateDraftChangeset creates the given changeset on the code host in draft mode.
// Perforce does not support draft changelists
func (s PerforceSource) CreateDraftChangeset(_ context.Context, _ *Changeset) (bool, error) {
	return false, errors.New("not implemented")
}

func (s PerforceSource) setChangesetMetadata(cl *protocol.PerforceChangelist, cs *Changeset) error {
	if err := cs.SetMetadata(cl); err != nil {
		return errors.Wrap(err, "setting changeset metadata")
	}

	return nil
}

// UndraftChangeset will update the Changeset on the source to be not in draft mode anymore.
func (s PerforceSource) UndraftChangeset(_ context.Context, _ *Changeset) error {
	// TODO: @peterguy implement this function?
	// not sure what it means in Perforce - submit the changelist?
	return errors.New("not implemented")
}

// CloseChangeset will close the Changeset on the source, where "close"
// means the appropriate final state on the codehost.
// deleted on Perforce, maybe?
func (s PerforceSource) CloseChangeset(_ context.Context, _ *Changeset) error {
	// TODO: @peterguy implement this function
	// delete changelist?
	return errors.New("not implemented")
}

// UpdateChangeset can update Changesets.
func (s PerforceSource) UpdateChangeset(_ context.Context, _ *Changeset) error {
	// TODO: @peterguy implement this function
	// not sure what this means for Perforce
	return errors.New("not implemented")
}

// ReopenChangeset will reopen the Changeset on the source, if it's closed.
// If not, it's a noop.
func (s PerforceSource) ReopenChangeset(_ context.Context, _ *Changeset) error {
	// TODO: @peterguy implement function
	// noop for Perforce?
	return errors.New("not implemented")
}

// CreateComment posts a comment on the Changeset.
func (s PerforceSource) CreateComment(_ context.Context, _ *Changeset, _ string) error {
	// TODO: @peterguy implement function
	// comment on changelist?
	return errors.New("not implemented")
}

// MergeChangeset merges a Changeset on the code host, if in a mergeable state.
// If squash is true, and the code host supports squash merges, the source
// must attempt a squash merge. Otherwise, it is expected to perform a regular
// merge. If the changeset cannot be merged, because it is in an unmergeable
// state, ChangesetNotMergeableError must be returned.
func (s PerforceSource) MergeChangeset(_ context.Context, _ *Changeset, _ bool) error {
	// TODO: @peterguy implement function
	// submit CL? Or no-op because we want to keep CLs pending and let the Perforce users manage them in other tools?
	return errors.New("not implemented")
}

func (s PerforceSource) BuildCommitOpts(repo *types.Repo, _ *btypes.Changeset, spec *btypes.ChangesetSpec, pushOpts *protocol.PushConfig) protocol.CreateCommitFromPatchRequest {
	return BuildCommitOptsCommon(repo, spec, pushOpts)
}
