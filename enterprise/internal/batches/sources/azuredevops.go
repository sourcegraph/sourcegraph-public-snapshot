package sources

import (
	"context"
	"fmt"
	azuredevops2 "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/sources/azuredevops"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/azuredevops"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
	"strings"
)

type AzureDevOpsSource struct {
	client azuredevops.Client
}

var (
	_ ForkableChangesetSource = AzureDevOpsSource{}
)

func NewAzureDevOpsSource(ctx context.Context, svc *types.ExternalService, cf *httpcli.Factory) (*AzureDevOpsSource, error) {
	rawConfig, err := svc.Config.Decrypt(ctx)
	if err != nil {
		return nil, errors.Errorf("external service id=%d config error: %s", svc.ID, err)
	}
	var c schema.AzureDevOpsConnection
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

	client, err := azuredevops.NewClient(svc.URN(), &c, cli)
	if err != nil {
		return nil, errors.Wrap(err, "creating Azure DevOps client")
	}

	return &AzureDevOpsSource{client: client}, nil
}

// GitserverPushConfig returns an authenticated push config used for pushing
// commits to the code host.
func (s AzureDevOpsSource) GitserverPushConfig(repo *types.Repo) (*protocol.PushConfig, error) {
	return GitserverPushConfig(repo, s.client.Authenticator())
}

// WithAuthenticator returns a copy of the original Source configured to use the
// given authenticator, provided that authenticator type is supported by the
// code host.
func (s AzureDevOpsSource) WithAuthenticator(a auth.Authenticator) (ChangesetSource, error) {
	client, err := s.client.WithAuthenticator(a)
	if err != nil {
		return nil, err
	}

	return &AzureDevOpsSource{client: client}, nil
}

// ValidateAuthenticator validates the currently set authenticator is usable.
// Returns an error, when validating the Authenticator yielded an error.
func (s AzureDevOpsSource) ValidateAuthenticator(ctx context.Context) error {
	_, err := s.client.AzureServicesProfile(ctx)
	return err
}

// LoadChangeset loads the given Changeset from the source and updates it. If
// the Changeset could not be found on the source, a ChangesetNotFoundError is
// returned.
func (s AzureDevOpsSource) LoadChangeset(ctx context.Context, cs *Changeset) error {
	repo := cs.TargetRepo.Metadata.(*azuredevops.Repository)
	args, err := s.createCommonPullRequestArgs(*repo, *cs)
	if err != nil {
		return err
	}

	pr, err := s.client.GetPullRequest(ctx, args)
	if err != nil {
		if errcode.IsNotFound(err) {
			return ChangesetNotFoundError{Changeset: cs}
		}
		return errors.Wrap(err, "getting pull request")
	}

	return s.setChangesetMetadata(ctx, repo, &pr, cs)
}

// CreateChangeset will create the Changeset on the source. If it already
// exists, *Changeset will be populated and the return value will be true.
func (s AzureDevOpsSource) CreateChangeset(ctx context.Context, cs *Changeset) (bool, error) {
	input := s.changesetToPullRequestInput(cs)
	repo := cs.TargetRepo.Metadata.(*azuredevops.Repository)
	org, err := repo.GetOrganization()
	if err != nil {
		return false, errors.Wrap(err, "getting Azure DevOps organization from project")
	}
	args := azuredevops.OrgProjectRepoArgs{
		Org:          org,
		Project:      repo.Project.Name,
		RepoNameOrID: repo.Name,
	}
	pr, err := s.client.CreatePullRequest(ctx, args, input)
	if err != nil {
		return false, errors.Wrap(err, "creating pull request")
	}

	if err := s.setChangesetMetadata(ctx, repo, &pr, cs); err != nil {
		return false, err
	}

	return true, nil
}

// CloseChangeset will close the Changeset on the source, where "close"
// means the appropriate final state on the codehost (e.g. "abandoned" on
// AzureDevOps).
func (s AzureDevOpsSource) CloseChangeset(ctx context.Context, cs *Changeset) error {
	repo := cs.TargetRepo.Metadata.(*azuredevops.Repository)
	args, err := s.createCommonPullRequestArgs(*repo, *cs)
	if err != nil {
		return err
	}

	updated, err := s.client.AbandonPullRequest(ctx, args)
	if err != nil {
		return errors.Wrap(err, "declining pull request")
	}

	return s.setChangesetMetadata(ctx, repo, &updated, cs)
}

// UpdateChangeset can update Changesets.
func (s AzureDevOpsSource) UpdateChangeset(ctx context.Context, cs *Changeset) error {
	input := s.changesetToUpdatePullRequestInput(cs)
	repo := cs.TargetRepo.Metadata.(*azuredevops.Repository)
	args, err := s.createCommonPullRequestArgs(*repo, *cs)
	if err != nil {
		return err
	}

	updated, err := s.client.UpdatePullRequest(ctx, args, input)
	if err != nil {
		return errors.Wrap(err, "updating pull request")
	}

	return s.setChangesetMetadata(ctx, repo, &updated, cs)
}

// TODO: @varsanojidan Look at this
// ReopenChangeset will reopen the Changeset on the source, if it's closed.
// If not, it's a noop.
func (s AzureDevOpsSource) ReopenChangeset(ctx context.Context, cs *Changeset) error {
	// Azure DevOps is a bit special, and can't reopen a declined PR under
	// any circumstances. (See https://jira.atlassian.com/browse/BCLOUD-4954 for
	// more details.)
	//
	// It will, however, allow a pull request to be recreated. So we're going to
	// do something a bit different to the other external services, and just
	// recreate the changeset wholesale.
	//
	// If the PR hasn't been declined, this will also work fine: Azure DevOps will
	// return the same PR in that case when we try to create it, so this is
	// still (effectively) a no-op, as required by the interface.
	_, err := s.CreateChangeset(ctx, cs)
	return err
}

// CreateComment posts a comment on the Changeset.
func (s AzureDevOpsSource) CreateComment(ctx context.Context, cs *Changeset, comment string) error {
	repo := cs.TargetRepo.Metadata.(*azuredevops.Repository)
	args, err := s.createCommonPullRequestArgs(*repo, *cs)
	if err != nil {
		return err
	}

	_, err = s.client.CreatePullRequestCommentThread(ctx, args, azuredevops.PullRequestCommentInput{
		Comments: []azuredevops.PullRequestCommentForInput{
			{
				ParentCommitID: 0,
				Content:        comment,
				CommentType:    1,
			},
		},
	})
	return err
}

// MergeChangeset merges a Changeset on the code host, if in a mergeable state.
// If squash is true, and the code host supports squash merges, the source
// must attempt a squash merge. Otherwise, it is expected to perform a regular
// merge. If the changeset cannot be merged, because it is in an unmergeable
// state, ChangesetNotMergeableError must be returned.
func (s AzureDevOpsSource) MergeChangeset(ctx context.Context, cs *Changeset, squash bool) error {
	repo := cs.TargetRepo.Metadata.(*azuredevops.Repository)
	args, err := s.createCommonPullRequestArgs(*repo, *cs)
	if err != nil {
		return err
	}

	var mergeStrategy *azuredevops.PullRequestMergeStrategy
	if squash {
		ms := azuredevops.PullRequestMergeStrategySquash
		mergeStrategy = &ms
	}

	updated, err := s.client.CompletePullRequest(ctx, args, azuredevops.PullRequestCompleteInput{
		CommitID:      cs.HeadRef,
		MergeStrategy: mergeStrategy,
	})
	if err != nil {
		if errcode.IsNotFound(err) {
			return errors.Wrap(err, "merging pull request")
		}
		return ChangesetNotMergeableError{ErrorMsg: err.Error()}
	}

	return s.setChangesetMetadata(ctx, repo, &updated, cs)
}

// GetNamespaceFork returns a repo pointing to a fork of the given repo in
// the given namespace, ensuring that the fork exists and is a fork of the
// target repo.
func (s AzureDevOpsSource) GetNamespaceFork(ctx context.Context, targetRepo *types.Repo, namespace string) (*types.Repo, error) {
	targetMeta := targetRepo.Metadata.(*azuredevops.Repository)

	org, err := targetMeta.GetOrganization()
	if err != nil {
		return nil, errors.Wrap(err, "getting Azure DevOps organization from project")
	}

	// Our namespace is structured like org/project
	forkOrg, forkProject, found := strings.Cut(namespace, "/")
	if !found {
		return nil, errors.Errorf("invalid namespace, does not match pattern (org/project): %s", namespace)
	}
	forkName := fmt.Sprintf("%s-%s-%s", org, targetMeta.Project.Name, targetMeta.Name)

	// Figure out if we already have the repo.
	if fork, err := s.client.GetRepo(ctx, azuredevops.OrgProjectRepoArgs{
		Project:      forkProject,
		Org:          forkOrg,
		RepoNameOrID: forkName,
	}); err == nil {
		return s.copyRepoAsFork(targetRepo, &fork)
	} else if !errcode.IsNotFound(err) {
		return nil, errors.Wrap(err, "checking for fork existence")
	}

	fork, err := s.client.ForkRepository(ctx, org, azuredevops.ForkRepositoryInput{
		Name: forkName,
		Project: azuredevops.ForkRepositoryInputProject{
			ID: targetMeta.Project.ID,
		},
		ParentRepository: azuredevops.ForkRepositoryInputParentRepository{
			ID: targetMeta.ID,
			Project: azuredevops.ForkRepositoryInputProject{
				ID: targetMeta.Project.ID,
			},
		},
	})
	if err != nil {
		return nil, errors.Wrap(err, "forking repository")
	}

	return s.copyRepoAsFork(targetRepo, &fork)
}

// TODO: ADO does not have user namespaces what do we do here?
// GetUserFork returns a repo pointing to a fork of the given repo in the
// currently authenticated user's namespace.
func (s AzureDevOpsSource) GetUserFork(ctx context.Context, targetRepo *types.Repo) (*types.Repo, error) {
	//user, err := s.client.CurrentUser(ctx)
	//if err != nil {
	//	return nil, errors.Wrap(err, "getting the current user")
	//}
	//
	//return s.GetNamespaceFork(ctx, targetRepo, user.Username)

	return nil, nil
}

func (s AzureDevOpsSource) copyRepoAsFork(targetRepo *types.Repo, fork *azuredevops.Repository) (*types.Repo, error) {
	targetMeta := targetRepo.Metadata.(*azuredevops.Repository)

	originalNamespace, err := targetMeta.GetNamespace()
	if err != nil {
		return nil, err
	}

	forkNamespace, err := fork.GetNamespace()
	if err != nil {
		return nil, err
	}

	// Now we make a copy of the target repo, but with its sources and metadata updated to
	// point to the fork
	forkRepo, err := copyAzureDevOpsRepoAsFork(targetRepo, fork, originalNamespace, forkNamespace, fork.Name)
	if err != nil {
		return nil, errors.Wrap(err, "updating target repo sources")
	}

	return forkRepo, nil
}

func (s AzureDevOpsSource) annotatePullRequest(ctx context.Context, repo *azuredevops.Repository, pr *azuredevops.PullRequest) (*azuredevops2.AzureDevOpsAnnotatedPullRequest, error) {
	org, err := repo.GetOrganization()
	if err != nil {
		return nil, err
	}
	srs, err := s.client.GetPullRequestStatuses(ctx, azuredevops.PullRequestCommonArgs{
		Org:          org,
		Project:      repo.Project.Name,
		RepoNameOrID: repo.Name,
	})
	if err != nil {
		return nil, errors.Wrap(err, "getting pull request statuses")
	}

	var statuses []*azuredevops.PullRequestBuildStatus
	for _, status := range srs {
		statuses = append(statuses, &status)
	}

	return &azuredevops2.AzureDevOpsAnnotatedPullRequest{
		PullRequest: pr,
		Statuses:    statuses,
	}, nil
}

func (s AzureDevOpsSource) setChangesetMetadata(ctx context.Context, repo *azuredevops.Repository, pr *azuredevops.PullRequest, cs *Changeset) error {
	apr, err := s.annotatePullRequest(ctx, repo, pr)
	if err != nil {
		return errors.Wrap(err, "annotating pull request")
	}

	if err := cs.SetMetadata(apr); err != nil {
		return errors.Wrap(err, "setting changeset metadata")
	}

	return nil
}

func (s AzureDevOpsSource) changesetToPullRequestInput(cs *Changeset) azuredevops.CreatePullRequestInput {
	destBranch := gitdomain.AbbreviateRef(cs.BaseRef)
	input := azuredevops.CreatePullRequestInput{
		Title:         cs.Title,
		Description:   cs.Body,
		SourceRefName: gitdomain.AbbreviateRef(cs.HeadRef),
		TargetRefName: destBranch,
	}

	// If we're forking, then we need to set the source repository as well.
	if cs.RemoteRepo != cs.TargetRepo {
		input.ForkSource = &azuredevops.ForkRef{
			Repository: *cs.RemoteRepo.Metadata.(*azuredevops.Repository),
		}
	}

	return input
}

func (s AzureDevOpsSource) changesetToUpdatePullRequestInput(cs *Changeset) azuredevops.PullRequestUpdateInput {
	destBranch := gitdomain.AbbreviateRef(cs.BaseRef)
	input := azuredevops.PullRequestUpdateInput{
		Title:       &cs.Title,
		Description: &cs.Body,
		// TODO: does this matter?
		// SourceRefName: gitdomain.AbbreviateRef(cs.HeadRef),
		TargetRefName: &destBranch,
	}
	fmt.Printf("CHANGESET: %+v\n", cs)
	fmt.Printf("INPUT: %+v\n", input)

	// TODO: does this matter?
	// If we're forking, then we need to set the source repository as well.
	//if cs.RemoteRepo != cs.TargetRepo {
	//	input.ForkSource.Repository = *cs.RemoteRepo.Metadata.(*azuredevops.Repository)
	//}

	return input
}

func (s AzureDevOpsSource) createCommonPullRequestArgs(repo azuredevops.Repository, cs Changeset) (azuredevops.PullRequestCommonArgs, error) {
	org, err := repo.GetOrganization()
	if err != nil {
		return azuredevops.PullRequestCommonArgs{}, errors.Wrap(err, "getting Azure DevOps organization from project")
	}
	fmt.Printf("repo: %+v\n", repo)
	return azuredevops.PullRequestCommonArgs{
		PullRequestID: cs.ExternalID,
		Org:           org,
		Project:       repo.Project.Name,
		RepoNameOrID:  repo.Name,
	}, nil
}

func copyAzureDevOpsRepoAsFork(repo *types.Repo, fork *azuredevops.Repository, originalNamespace, forkNamespace, forkName string) (*types.Repo, error) {
	forkRepo := *repo

	if repo.Sources == nil || len(repo.Sources) == 0 {
		return nil, errors.New("repo has no sources")
	}

	forkSources := map[string]*types.SourceInfo{}

	for urn, src := range repo.Sources {
		if src != nil || src.CloneURL != "" {
			forkURL := strings.Replace(
				strings.ToLower(src.CloneURL),
				strings.ToLower(originalNamespace),
				strings.ToLower(forkNamespace),
				1,
			)
			lastSlash := strings.LastIndex(forkURL, "/")
			if lastSlash <= 0 {
				return nil, errors.New("repo has malformed clone url")
			}

			forkURL = forkURL[:lastSlash+1] + forkName

			forkSources[urn] = &types.SourceInfo{
				ID:       src.ID,
				CloneURL: forkURL,
			}
		}
	}

	forkRepo.Sources = forkSources
	forkRepo.Metadata = fork

	return &forkRepo, nil
}
