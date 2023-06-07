package sources

import (
	"context"
	"net/url"
	"strings"

	p4batches "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/sources/perforce"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/perforce"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

type PerforceSource struct {
	server schema.PerforceConnection
}

func NewPerforceSource(ctx context.Context, svc *types.ExternalService, cf *httpcli.Factory) (*PerforceSource, error) {
	rawConfig, err := svc.Config.Decrypt(ctx)
	if err != nil {
		return nil, errors.Errorf("external service id=%d config error: %s", svc.ID, err)
	}
	var c schema.PerforceConnection
	if err := jsonc.Unmarshal(rawConfig, &c); err != nil {
		return nil, errors.Wrapf(err, "external service id=%d", svc.ID)
	}

	return &PerforceSource{server: c}, nil
}

// GitserverPushConfig returns an authenticated push config used for pushing commits to the code host.
func (s PerforceSource) GitserverPushConfig(repo *types.Repo) (*protocol.PushConfig, error) {
	return &protocol.PushConfig{
		RemoteURL: s.server.P4Port,
		P4Credentials: &protocol.P4Credentials{
			P4User:   s.server.P4User,
			P4Passwd: s.server.P4Passwd,
		},
	}, nil
}

// WithAuthenticator returns a copy of the original Source configured to use the
// given authenticator, provided that authenticator type is supported by the
// code host.
func (s PerforceSource) WithAuthenticator(a auth.Authenticator) (ChangesetSource, error) {
	return s, nil
}

// ValidateAuthenticator validates the currently set authenticator is usable.
// Returns an error, when validating the Authenticator yielded an error.
func (s PerforceSource) ValidateAuthenticator(ctx context.Context) error {
	return nil
}

// LoadChangeset loads the given Changeset from the source and updates it. If
// the Changeset could not be found on the source, a ChangesetNotFoundError is
// returned.
func (s PerforceSource) LoadChangeset(_ context.Context, _ *Changeset) error {
	// TODO: implement this method
	// probably will load a pending changelist
	return nil
}

// CreateChangeset will create the Changeset on the source. If it already
// exists, *Changeset will be populated and the return value will be true.
func (s PerforceSource) CreateChangeset(_ context.Context, cs *Changeset) (bool, error) {
	return s.createChangeset(cs)
}

// CreateDraftChangeset creates the given changeset on the code host in draft mode.
// Perforce does not support draft PRs so it creates a sstandard one
func (s PerforceSource) CreateDraftChangeset(_ context.Context, cs *Changeset) (bool, error) {
	return s.createChangeset(cs)
}

func (s PerforceSource) createChangeset(cs *Changeset) (bool, error) {
	// TODO: implement this function
	// create a pending changelist?
	cl := perforce.Changelist{}

	if err := s.setChangesetMetadata(&cl, cs); err != nil {
		return false, errors.Wrap(err, "setting Perforce changeset metadata")
	}

	return true, nil
}

func (s PerforceSource) setChangesetMetadata(cl *perforce.Changelist, cs *Changeset) error {
	acl := new(p4batches.AnnotatedChangelist)
	acl.Changelist = cl

	if err := cs.SetMetadata(acl); err != nil {
		return errors.Wrap(err, "setting changeset metadata")
	}

	return nil
}

// UndraftChangeset will update the Changeset on the source to be not in draft mode anymore.
func (s PerforceSource) UndraftChangeset(ctx context.Context, cs *Changeset) error {
	// TODO: implement this function?
	// not sure what it means in Perforce - submit the changelist?
	return nil
}

// CloseChangeset will close the Changeset on the source, where "close"
// means the appropriate final state on the codehost.
// deleted on Perforce, maybe?
func (s PerforceSource) CloseChangeset(ctx context.Context, cs *Changeset) error {
	// TODO: implement this function
	// delete changelist?
	return nil
}

// UpdateChangeset can update Changesets.
func (s PerforceSource) UpdateChangeset(ctx context.Context, cs *Changeset) error {
	// TODO: implement this function
	// not sure what this means for Perforce
	return nil
}

// ReopenChangeset will reopen the Changeset on the source, if it's closed.
// If not, it's a noop.
func (s PerforceSource) ReopenChangeset(ctx context.Context, cs *Changeset) error {
	// TODO: implement function
	// noop for Perforce?
	return nil
}

// CreateComment posts a comment on the Changeset.
func (s PerforceSource) CreateComment(ctx context.Context, cs *Changeset, comment string) error {
	// TODO: implement function
	// comment on changelist?
	return nil
}

// MergeChangeset merges a Changeset on the code host, if in a mergeable state.
// If squash is true, and the code host supports squash merges, the source
// must attempt a squash merge. Otherwise, it is expected to perform a regular
// merge. If the changeset cannot be merged, because it is in an unmergeable
// state, ChangesetNotMergeableError must be returned.
func (s PerforceSource) MergeChangeset(ctx context.Context, cs *Changeset, squash bool) error {
	// TODO: implement function
	// submit CL? Or no-op because we want to keep CLs pending and let the Perforce users manage them in other tools?
	return nil
}

// GetFork returns a repo pointing to a fork of the target repo, ensuring that the fork
// exists and creating it if it doesn't. If namespace is not provided, the original namespace is used.
// If name is not provided, the fork will be named with the default Sourcegraph convention:
// "${original-namespace}-${original-name}"
func (s PerforceSource) GetFork(ctx context.Context, targetRepo *types.Repo, ns, n *string) (*types.Repo, error) {
	// TODO: implement function
	// no-op for Perforce?
	return nil, nil
}

func (s PerforceSource) BuildCommitOpts(repo *types.Repo, _ *btypes.Changeset, spec *btypes.ChangesetSpec, pushOpts *protocol.PushConfig) protocol.CreateCommitFromPatchRequest {
	return BuildCommitOptsCommon(repo, spec, pushOpts)
}

func copyPerforceRepoAsFork(repo *types.Repo, fork *perforce.Repository, forkNamespace, forkName string) (*types.Repo, error) {
	if repo.Sources == nil || len(repo.Sources) == 0 {
		return nil, errors.New("repo has no sources")
	}

	forkRepo := *repo
	forkSources := map[string]*types.SourceInfo{}

	for urn, src := range repo.Sources {
		if src == nil || src.CloneURL == "" {
			continue
		}
		forkURL, err := url.Parse(src.CloneURL)
		if err != nil {
			return nil, err
		}

		// Will look like: /org/project/_git/repo, project is our namespace.
		forkURLPathSplit := strings.SplitN(forkURL.Path, "/", 5)
		if len(forkURLPathSplit) < 5 {
			return nil, errors.Errorf("repo has malformed clone url: %s", src.CloneURL)
		}
		forkURLPathSplit[2] = forkNamespace
		forkURLPathSplit[4] = forkName

		forkPath := strings.Join(forkURLPathSplit, "/")
		forkURL.Path = forkPath

		forkSources[urn] = &types.SourceInfo{
			ID:       src.ID,
			CloneURL: forkURL.String(),
		}
	}

	forkRepo.Sources = forkSources
	forkRepo.Metadata = fork

	return &forkRepo, nil
}
