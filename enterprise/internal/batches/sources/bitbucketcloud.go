package sources

import (
	"context"
	"strconv"

	bbcs "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/sources/bitbucketcloud"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketcloud"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

type BitbucketCloudSource struct {
	client bitbucketcloud.Client
}

var (
	_ ForkableChangesetSource = BitbucketCloudSource{}
)

func NewBitbucketCloudSource(svc *types.ExternalService, cf *httpcli.Factory) (*BitbucketCloudSource, error) {
	var c schema.BitbucketCloudConnection
	if err := jsonc.Unmarshal(svc.Config, &c); err != nil {
		return nil, errors.Wrapf(err, "external service id=%d", svc.ID)
	}

	if cf == nil {
		cf = httpcli.ExternalClientFactory
	}

	// No options to provide here, since Bitbucket Cloud doesn't support custom
	// certificates, unlike the other
	cli, err := cf.Doer()
	if err != nil {
		return nil, errors.Wrap(err, "creating external client")
	}

	client, err := bitbucketcloud.NewClient(svc.URN(), &c, cli)
	if err != nil {
		return nil, errors.Wrap(err, "creating Bitbucket Cloud client")
	}

	return &BitbucketCloudSource{client: client}, nil
}

// GitserverPushConfig returns an authenticated push config used for pushing
// commits to the code host.
func (s BitbucketCloudSource) GitserverPushConfig(ctx context.Context, store database.ExternalServiceStore, repo *types.Repo) (*protocol.PushConfig, error) {
	return gitserverPushConfig(ctx, store, repo, s.client.Authenticator())
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

	return s.setChangesetMetadata(ctx, repo, pr, cs)
}

// CreateChangeset will create the Changeset on the source. If it already
// exists, *Changeset will be populated and the return value will be true.
func (s BitbucketCloudSource) CreateChangeset(ctx context.Context, cs *Changeset) (bool, error) {
	opts := s.changesetToPullRequestInput(cs)
	targetRepo := cs.TargetRepo.Metadata.(*bitbucketcloud.Repo)

	pr, err := s.client.CreatePullRequest(ctx, targetRepo, opts)
	if err != nil {
		return false, errors.Wrap(err, "creating pull request")
	}

	if err := s.setChangesetMetadata(ctx, targetRepo, pr, cs); err != nil {
		return false, err
	}

	// Fun fact: Bitbucket Cloud will silently update an existing pull request
	// if one already exists, rather than returning some sort of error. We don't
	// really have a way to tell if the PR existed or not, so we'll simply say
	// it did, and we can go through the IsOutdated check after regardless.
	return true, nil
}

// CloseChangeset will close the Changeset on the source, where "close"
// means the appropriate final state on the codehost (e.g. "declined" on
// Bitbucket Server).
func (s BitbucketCloudSource) CloseChangeset(ctx context.Context, cs *Changeset) error {
	repo := cs.TargetRepo.Metadata.(*bitbucketcloud.Repo)
	pr := cs.Metadata.(*bbcs.AnnotatedPullRequest)
	updated, err := s.client.DeclinePullRequest(ctx, repo, pr.ID)
	if err != nil {
		return errors.Wrap(err, "declining pull request")
	}

	return s.setChangesetMetadata(ctx, repo, updated, cs)
}

// UpdateChangeset can update Changesets.
func (s BitbucketCloudSource) UpdateChangeset(ctx context.Context, cs *Changeset) error {
	opts := s.changesetToPullRequestInput(cs)
	targetRepo := cs.TargetRepo.Metadata.(*bitbucketcloud.Repo)
	pr := cs.Metadata.(*bbcs.AnnotatedPullRequest)

	updated, err := s.client.UpdatePullRequest(ctx, targetRepo, pr.ID, opts)
	if err != nil {
		return errors.Wrap(err, "updating pull request")
	}

	return s.setChangesetMetadata(ctx, targetRepo, updated, cs)
}

// ReopenChangeset will reopen the Changeset on the source, if it's closed.
// If not, it's a noop.
func (s BitbucketCloudSource) ReopenChangeset(ctx context.Context, cs *Changeset) error {
	// Bitbucket Cloud is a bit special, and can't reopen a declined PR under
	// any circumstances. (See https://jira.atlassian.com/browse/BCLOUD-4954 for
	// more details.)
	//
	// It will, however, allow a pull request to be recreated. So we're going to
	// do something a bit different to the other external services, and just
	// recreate the changeset wholesale.
	//
	// If the PR hasn't been declined, this will also work fine: Bitbucket will
	// return the same PR in that case when we try to create it, so this is
	// still (effectively) a no-op, as required by the interface.
	_, err := s.CreateChangeset(ctx, cs)
	return err
}

// CreateComment posts a comment on the Changeset.
func (s BitbucketCloudSource) CreateComment(ctx context.Context, cs *Changeset, comment string) error {
	repo := cs.TargetRepo.Metadata.(*bitbucketcloud.Repo)
	pr := cs.Metadata.(*bbcs.AnnotatedPullRequest)

	_, err := s.client.CreatePullRequestComment(ctx, repo, pr.ID, bitbucketcloud.CommentInput{
		Content: comment,
	})
	return err
}

// MergeChangeset merges a Changeset on the code host, if in a mergeable state.
// If squash is true, and the code host supports squash merges, the source
// must attempt a squash merge. Otherwise, it is expected to perform a regular
// merge. If the changeset cannot be merged, because it is in an unmergeable
// state, ChangesetNotMergeableError must be returned.
func (s BitbucketCloudSource) MergeChangeset(ctx context.Context, cs *Changeset, squash bool) error {
	repo := cs.TargetRepo.Metadata.(*bitbucketcloud.Repo)
	pr := cs.Metadata.(*bbcs.AnnotatedPullRequest)

	var mergeStrategy *bitbucketcloud.MergeStrategy
	if squash {
		ms := bitbucketcloud.MergeStrategySquash
		mergeStrategy = &ms
	}

	updated, err := s.client.MergePullRequest(ctx, repo, pr.ID, bitbucketcloud.MergePullRequestOpts{
		MergeStrategy: mergeStrategy,
	})
	if err != nil {
		if errcode.IsNotFound(err) {
			return errors.Wrap(err, "merging pull request")
		}
		return ChangesetNotMergeableError{ErrorMsg: err.Error()}
	}

	return s.setChangesetMetadata(ctx, repo, updated, cs)
}

// GetNamespaceFork returns a repo pointing to a fork of the given repo in
// the given namespace, ensuring that the fork exists and is a fork of the
// target repo.
func (s BitbucketCloudSource) GetNamespaceFork(ctx context.Context, targetRepo *types.Repo, namespace string) (*types.Repo, error) {
	upstreamRepo := targetRepo.Metadata.(*bitbucketcloud.Repo)

	// Figure out if we already have the repo.
	if fork, err := s.client.Repo(ctx, namespace, upstreamRepo.Slug); err == nil {
		return s.createRemoteRepo(targetRepo, fork), nil
	} else if !errcode.IsNotFound(err) {
		return nil, errors.Wrap(err, "checking for fork existence")
	}

	fork, err := s.client.ForkRepository(ctx, upstreamRepo, bitbucketcloud.ForkInput{
		Workspace: bitbucketcloud.ForkInputWorkspace(namespace),
	})
	if err != nil {
		return nil, errors.Wrap(err, "forking repository")
	}

	return s.createRemoteRepo(targetRepo, fork), nil
}

// GetUserFork returns a repo pointing to a fork of the given repo in the
// currently authenticated user's namespace.
func (s BitbucketCloudSource) GetUserFork(ctx context.Context, targetRepo *types.Repo) (*types.Repo, error) {
	user, err := s.client.CurrentUser(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "getting the current user")
	}

	return s.GetNamespaceFork(ctx, targetRepo, user.Username)
}

func (BitbucketCloudSource) createRemoteRepo(targetRepo *types.Repo, fork *bitbucketcloud.Repo) *types.Repo {
	// This needs to be good enough to get the right values out of
	// bitbucketCloudCloneURL(), which only looks at the metadata, so we can
	// just copy it in over the top of the targetRepo.
	remoteRepo := *targetRepo
	remoteRepo.Metadata = fork

	return &remoteRepo
}

func (s BitbucketCloudSource) annotatePullRequest(ctx context.Context, repo *bitbucketcloud.Repo, pr *bitbucketcloud.PullRequest) (*bbcs.AnnotatedPullRequest, error) {
	srs, err := s.client.GetPullRequestStatuses(repo, pr.ID)
	if err != nil {
		return nil, errors.Wrap(err, "getting pull request statuses")
	}
	all, err := srs.All(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "getting pull request statuses as slice")
	}

	statuses := []*bitbucketcloud.PullRequestStatus{}
	for _, v := range all {
		statuses = append(statuses, v.(*bitbucketcloud.PullRequestStatus))
	}

	return &bbcs.AnnotatedPullRequest{
		PullRequest: pr,
		Statuses:    statuses,
	}, nil
}

func (s BitbucketCloudSource) setChangesetMetadata(ctx context.Context, repo *bitbucketcloud.Repo, pr *bitbucketcloud.PullRequest, cs *Changeset) error {
	apr, err := s.annotatePullRequest(ctx, repo, pr)
	if err != nil {
		return errors.Wrap(err, "annotating pull request")
	}

	if err := cs.SetMetadata(apr); err != nil {
		return errors.Wrap(err, "setting changeset metadata")
	}

	return nil
}

func (s BitbucketCloudSource) changesetToPullRequestInput(cs *Changeset) bitbucketcloud.PullRequestInput {
	destBranch := gitdomain.AbbreviateRef(cs.BaseRef)
	opts := bitbucketcloud.PullRequestInput{
		Title:             cs.Title,
		Description:       cs.Body,
		SourceBranch:      gitdomain.AbbreviateRef(cs.HeadRef),
		DestinationBranch: &destBranch,
	}

	// If we're forking, then we need to set the source repository as well.
	if cs.RemoteRepo != cs.TargetRepo {
		opts.SourceRepo = cs.RemoteRepo.Metadata.(*bitbucketcloud.Repo)
	}

	return opts
}
