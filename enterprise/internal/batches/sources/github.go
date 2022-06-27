package sources

import (
	"context"
	"net/url"
	"strconv"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

type GithubSource struct {
	client *github.V4Client
	au     auth.Authenticator
}

var _ ForkableChangesetSource = GithubSource{}

func NewGithubSource(svc *types.ExternalService, cf *httpcli.Factory) (*GithubSource, error) {
	var c schema.GitHubConnection
	if err := jsonc.Unmarshal(svc.Config, &c); err != nil {
		return nil, errors.Errorf("external service id=%d config error: %s", svc.ID, err)
	}
	return newGithubSource(svc.URN(), &c, cf, nil)
}

func newGithubSource(urn string, c *schema.GitHubConnection, cf *httpcli.Factory, au auth.Authenticator) (*GithubSource, error) {
	baseURL, err := url.Parse(c.Url)
	if err != nil {
		return nil, err
	}
	baseURL = extsvc.NormalizeBaseURL(baseURL)

	apiURL, _ := github.APIRoot(baseURL)

	if cf == nil {
		cf = httpcli.ExternalClientFactory
	}

	opts := httpClientCertificateOptions([]httpcli.Opt{
		// Use a 30s timeout to avoid running into EOF errors, because GitHub
		// closes idle connections after 60s
		httpcli.NewIdleConnTimeoutOpt(30 * time.Second),
	}, c.Certificate)

	cli, err := cf.Doer(opts...)
	if err != nil {
		return nil, err
	}

	var authr = au
	if au == nil {
		authr = &auth.OAuthBearerToken{Token: c.Token}
	}

	return &GithubSource{
		au:     authr,
		client: github.NewV4Client(urn, apiURL, authr, cli),
	}, nil
}

func (s GithubSource) GitserverPushConfig(ctx context.Context, store database.ExternalServiceStore, repo *types.Repo) (*protocol.PushConfig, error) {
	return gitserverPushConfig(ctx, store, repo, s.au)
}

func (s GithubSource) WithAuthenticator(a auth.Authenticator) (ChangesetSource, error) {
	switch a.(type) {
	case *auth.OAuthBearerToken,
		*auth.OAuthBearerTokenWithSSH:
		break

	default:
		return nil, newUnsupportedAuthenticatorError("GithubSource", a)
	}

	sc := s
	sc.au = a
	sc.client = sc.client.WithAuthenticator(a)

	return &sc, nil
}

func (s GithubSource) ValidateAuthenticator(ctx context.Context) error {
	_, err := s.client.GetAuthenticatedUser(ctx)
	return err
}

// CreateChangeset creates the given changeset on the code host.
func (s GithubSource) CreateChangeset(ctx context.Context, c *Changeset) (bool, error) {
	input, err := buildCreatePullRequestInput(c)
	if err != nil {
		return false, err
	}

	return s.createChangeset(ctx, c, input)
}

// CreateDraftChangeset creates the given changeset on the code host in draft mode.
func (s GithubSource) CreateDraftChangeset(ctx context.Context, c *Changeset) (bool, error) {
	input, err := buildCreatePullRequestInput(c)
	if err != nil {
		return false, err
	}

	input.Draft = true
	return s.createChangeset(ctx, c, input)
}

func buildCreatePullRequestInput(c *Changeset) (*github.CreatePullRequestInput, error) {
	headRef := gitdomain.AbbreviateRef(c.HeadRef)
	if c.RemoteRepo != c.TargetRepo {
		owner, err := c.RemoteRepo.Metadata.(*github.Repository).Owner()
		if err != nil {
			return nil, err
		}

		headRef = owner + ":" + headRef
	}

	return &github.CreatePullRequestInput{
		RepositoryID: c.TargetRepo.Metadata.(*github.Repository).ID,
		Title:        c.Title,
		Body:         c.Body,
		HeadRefName:  headRef,
		BaseRefName:  gitdomain.AbbreviateRef(c.BaseRef),
	}, nil
}

func (s GithubSource) createChangeset(ctx context.Context, c *Changeset, prInput *github.CreatePullRequestInput) (bool, error) {
	var exists bool
	pr, err := s.client.CreatePullRequest(ctx, prInput)
	if err != nil {
		if err != github.ErrPullRequestAlreadyExists {
			return exists, err
		}
		repo := c.TargetRepo.Metadata.(*github.Repository)
		owner, name, err := github.SplitRepositoryNameWithOwner(repo.NameWithOwner)
		if err != nil {
			return exists, errors.Wrap(err, "getting repo owner and name")
		}
		pr, err = s.client.GetOpenPullRequestByRefs(ctx, owner, name, c.BaseRef, c.HeadRef)
		if err != nil {
			return exists, errors.Wrap(err, "fetching existing PR")
		}
		exists = true
	}

	if err := c.SetMetadata(pr); err != nil {
		return false, errors.Wrap(err, "setting changeset metadata")
	}

	return exists, nil
}

// CloseChangeset closes the given *Changeset on the code host and updates the
// Metadata column in the *batches.Changeset to the newly closed pull request.
func (s GithubSource) CloseChangeset(ctx context.Context, c *Changeset) error {
	pr, ok := c.Changeset.Metadata.(*github.PullRequest)
	if !ok {
		return errors.New("Changeset is not a GitHub pull request")
	}

	err := s.client.ClosePullRequest(ctx, pr)
	if err != nil {
		return err
	}

	return c.Changeset.SetMetadata(pr)
}

// UndraftChangeset will update the Changeset on the source to be not in draft mode anymore.
func (s GithubSource) UndraftChangeset(ctx context.Context, c *Changeset) error {
	pr, ok := c.Changeset.Metadata.(*github.PullRequest)
	if !ok {
		return errors.New("Changeset is not a GitHub pull request")
	}

	err := s.client.MarkPullRequestReadyForReview(ctx, pr)
	if err != nil {
		return err
	}

	return c.Changeset.SetMetadata(pr)
}

// LoadChangeset loads the latest state of the given Changeset from the codehost.
func (s GithubSource) LoadChangeset(ctx context.Context, cs *Changeset) error {
	repo := cs.TargetRepo.Metadata.(*github.Repository)
	number, err := strconv.ParseInt(cs.ExternalID, 10, 64)
	if err != nil {
		return errors.Wrap(err, "parsing changeset external id")
	}

	pr := &github.PullRequest{
		RepoWithOwner: repo.NameWithOwner,
		Number:        number,
	}

	if err := s.client.LoadPullRequest(ctx, pr); err != nil {
		if github.IsNotFound(err) {
			return ChangesetNotFoundError{Changeset: cs}
		}
		return err
	}

	if err := cs.SetMetadata(pr); err != nil {
		return errors.Wrap(err, "setting changeset metadata")
	}

	return nil
}

// UpdateChangeset updates the given *Changeset in the code host.
func (s GithubSource) UpdateChangeset(ctx context.Context, c *Changeset) error {
	pr, ok := c.Changeset.Metadata.(*github.PullRequest)
	if !ok {
		return errors.New("Changeset is not a GitHub pull request")
	}

	updated, err := s.client.UpdatePullRequest(ctx, &github.UpdatePullRequestInput{
		PullRequestID: pr.ID,
		Title:         c.Title,
		Body:          c.Body,
		BaseRefName:   gitdomain.AbbreviateRef(c.BaseRef),
	})

	if err != nil {
		return err
	}

	return c.Changeset.SetMetadata(updated)
}

// ReopenChangeset reopens the given *Changeset on the code host.
func (s GithubSource) ReopenChangeset(ctx context.Context, c *Changeset) error {
	pr, ok := c.Changeset.Metadata.(*github.PullRequest)
	if !ok {
		return errors.New("Changeset is not a GitHub pull request")
	}

	err := s.client.ReopenPullRequest(ctx, pr)
	if err != nil {
		return err
	}

	return c.Changeset.SetMetadata(pr)
}

// CreateComment posts a comment on the Changeset.
func (s GithubSource) CreateComment(ctx context.Context, c *Changeset, text string) error {
	pr, ok := c.Changeset.Metadata.(*github.PullRequest)
	if !ok {
		return errors.New("Changeset is not a GitHub pull request")
	}

	return s.client.CreatePullRequestComment(ctx, pr, text)
}

// MergeChangeset merges a Changeset on the code host, if in a mergeable state.
// If squash is true, a squash-then-merge merge will be performed.
func (s GithubSource) MergeChangeset(ctx context.Context, c *Changeset, squash bool) error {
	pr, ok := c.Changeset.Metadata.(*github.PullRequest)
	if !ok {
		return errors.New("Changeset is not a GitHub pull request")
	}

	if err := s.client.MergePullRequest(ctx, pr, squash); err != nil {
		if github.IsNotMergeable(err) {
			return ChangesetNotMergeableError{ErrorMsg: err.Error()}
		}
		return err
	}

	return c.Changeset.SetMetadata(pr)
}

// GetNamespaceFork returns a repo pointing to a fork of the given repo in
// the given namespace, ensuring that the fork exists and is a fork of the
// target repo.
func (s GithubSource) GetNamespaceFork(ctx context.Context, targetRepo *types.Repo, namespace string) (*types.Repo, error) {
	return githubGetUserFork(ctx, targetRepo, s.client, &namespace)
}

// GetUserFork returns a repo pointing to a fork of the given repo in the
// currently authenticated user's namespace.
func (s GithubSource) GetUserFork(ctx context.Context, targetRepo *types.Repo) (*types.Repo, error) {
	// The implementation is separated here so we can mock the GitHub client.
	return githubGetUserFork(ctx, targetRepo, s.client, nil)
}

type githubClientFork interface {
	Fork(context.Context, string, string, *string) (*github.Repository, error)
}

func githubGetUserFork(ctx context.Context, targetRepo *types.Repo, client githubClientFork, namespace *string) (*types.Repo, error) {
	meta, ok := targetRepo.Metadata.(*github.Repository)
	if !ok || meta == nil {
		return nil, errors.New("target repo is not a GitHub repo")
	}

	owner, name, err := github.SplitRepositoryNameWithOwner(meta.NameWithOwner)
	if err != nil {
		return nil, errors.New("parsing repo name")
	}

	fork, err := client.Fork(ctx, owner, name, namespace)
	if err != nil {
		return nil, errors.Wrap(err, "forking repository")
	}

	remoteRepo := *targetRepo
	remoteRepo.Metadata = fork

	return &remoteRepo, nil
}
