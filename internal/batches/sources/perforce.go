package sources

import (
	"context"
	"fmt"
	"net/url"

	btypes "github.com/sourcegraph/sourcegraph/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"
	"github.com/sourcegraph/sourcegraph/internal/perforce"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

type PerforceSource struct {
	conn            schema.PerforceConnection
	gitServerClient gitserver.Client
	perforceCreds   *protocol.PerforceConnectionDetails
}

func NewPerforceSource(ctx context.Context, gitserverClient gitserver.Client, svc *types.ExternalService, _ *httpcli.Factory) (*PerforceSource, error) {
	rawConfig, err := svc.Config.Decrypt(ctx)
	if err != nil {
		return nil, errors.Errorf("external service id=%d config error: %s", svc.ID, err)
	}
	var c schema.PerforceConnection
	if err := jsonc.Unmarshal(rawConfig, &c); err != nil {
		return nil, errors.Wrapf(err, "external service id=%d", svc.ID)
	}

	return &PerforceSource{
		conn:            c,
		gitServerClient: gitserverClient,
	}, nil
}

// GitserverPushConfig returns an authenticated push config used for pushing commits to the code host.
func (s PerforceSource) GitserverPushConfig(_ context.Context, repo *types.Repo) (*protocol.PushConfig, error) {
	if s.perforceCreds == nil {
		return nil, errors.New("no credentials set for Perforce Source")
	}

	// Return a PushConfig with a crafted URL that includes the Perforce scheme and the credentials
	// The perforce scheme will tell `createCommitFromPatch` that this repo is a Perforce repo
	// so it can handle it differently from Git repos.
	// TODO: @peterguy: this seems to be the correct way to include the depot; confirm with more examples from code host configurations
	depot := ""
	u, err := url.Parse(repo.URI)
	if err == nil {
		depot = "//" + u.Path + "/"
	}
	remoteURL := fmt.Sprintf("perforce://%s:%s@%s%s", s.perforceCreds.P4User, s.perforceCreds.P4Passwd, s.conn.P4Port, depot)
	return &protocol.PushConfig{
		RemoteURL: remoteURL,
	}, nil
}

// WithAuthenticator returns a copy of the original Source configured to use the
// given authenticator, provided that authenticator type is supported by the
// code host.
func (s PerforceSource) WithAuthenticator(a auth.Authenticator) (ChangesetSource, error) {
	switch av := a.(type) {
	case *auth.BasicAuthWithSSH:
		s.perforceCreds = &protocol.PerforceConnectionDetails{
			P4Port:   s.conn.P4Port,
			P4User:   av.Username,
			P4Passwd: av.Password,
		}
	case *auth.BasicAuth:
		s.perforceCreds = &protocol.PerforceConnectionDetails{
			P4Port:   s.conn.P4Port,
			P4User:   av.Username,
			P4Passwd: av.Password,
		}
	default:
		return s, errors.New("unexpected auther type for Perforce Source")
	}

	return s, nil
}

// ValidateAuthenticator validates the currently set authenticator is usable.
// Returns an error, when validating the Authenticator yielded an error.
func (s PerforceSource) ValidateAuthenticator(ctx context.Context) error {
	if s.perforceCreds == nil {
		return errors.New("no credentials set for Perforce Source")
	}
	return s.gitServerClient.CheckPerforceCredentials(ctx, *s.perforceCreds)
}

// LoadChangeset loads the given Changeset from the source and updates it. If
// the Changeset could not be found on the source, a ChangesetNotFoundError is
// returned.
func (s PerforceSource) LoadChangeset(ctx context.Context, cs *Changeset) error {
	if s.perforceCreds == nil {
		return errors.New("no credentials set for Perforce Source")
	}
	cl, err := s.gitServerClient.PerforceGetChangelist(ctx, *s.perforceCreds, cs.ExternalID)
	if err != nil {
		return errors.Wrap(err, "getting changelist")
	}

	return errors.Wrap(s.setChangesetMetadata(cl, cs), "setting perforce changeset metadata")
}

// CreateChangeset will create the Changeset on the source. If it already
// exists, *Changeset will be populated and the return value will be true.
func (s PerforceSource) CreateChangeset(ctx context.Context, cs *Changeset) (bool, error) {
	return false, s.LoadChangeset(ctx, cs)
}

func (s PerforceSource) AuthenticationStrategy() AuthenticationStrategy {
	return AuthenticationStrategyUserCredential
}

// CreateDraftChangeset creates the given changeset on the code host in draft mode.
// Perforce does not support draft changelists
func (s PerforceSource) CreateDraftChangeset(_ context.Context, _ *Changeset) (bool, error) {
	return false, errors.New("not implemented")
}

func (s PerforceSource) setChangesetMetadata(cl *perforce.Changelist, cs *Changeset) error {
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
