package sources

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
	"github.com/sourcegraph/sourcegraph/schema"
)

type BitbucketServerSource struct {
	client *bitbucketserver.Client
	au     auth.Authenticator
}

// NewBitbucketServerSource returns a new BitbucketServerSource from the given external service.
func NewBitbucketServerSource(svc *types.ExternalService, cf *httpcli.Factory) (*BitbucketServerSource, error) {
	var c schema.BitbucketServerConnection
	if err := jsonc.Unmarshal(svc.Config, &c); err != nil {
		return nil, errors.Errorf("external service id=%d config error: %s", svc.ID, err)
	}
	return newBitbucketServerSource(&c, cf, nil)
}

func newBitbucketServerSource(c *schema.BitbucketServerConnection, cf *httpcli.Factory, au auth.Authenticator) (*BitbucketServerSource, error) {
	if cf == nil {
		cf = httpcli.ExternalClientFactory
	}

	var opts []httpcli.Opt
	if c.Certificate != "" {
		opts = append(opts, httpcli.NewCertPoolOpt(c.Certificate))
	}

	cli, err := cf.Doer(opts...)
	if err != nil {
		return nil, err
	}

	client, err := bitbucketserver.NewClient(c, cli)
	if err != nil {
		return nil, err
	}

	if au != nil {
		client = client.WithAuthenticator(au)
	}

	return &BitbucketServerSource{
		au:     client.Auth,
		client: client,
	}, nil
}

func (s BitbucketServerSource) GitserverPushConfig(ctx context.Context, store *database.ExternalServiceStore, repo *types.Repo) (*protocol.PushConfig, error) {
	return gitserverPushConfig(ctx, store, repo, s.au)
}

func (s BitbucketServerSource) WithAuthenticator(a auth.Authenticator) (ChangesetSource, error) {
	switch a.(type) {
	case *auth.OAuthBearerToken,
		*auth.OAuthBearerTokenWithSSH,
		*auth.BasicAuth,
		*auth.BasicAuthWithSSH,
		*bitbucketserver.SudoableOAuthClient:
		break

	default:
		return nil, newUnsupportedAuthenticatorError("BitbucketServerSource", a)
	}

	return &BitbucketServerSource{
		client: s.client.WithAuthenticator(a),
		au:     a,
	}, nil
}

// AuthenticatedUsername uses the underlying bitbucketserver.Client to get the
// username belonging to the credentials associated with the
// BitbucketServerSource.
func (s BitbucketServerSource) AuthenticatedUsername(ctx context.Context) (string, error) {
	return s.client.AuthenticatedUsername(ctx)
}

func (s BitbucketServerSource) ValidateAuthenticator(ctx context.Context) error {
	_, err := s.client.AuthenticatedUsername(ctx)
	return err
}

// CreateChangeset creates the given *Changeset in the code host.
func (s BitbucketServerSource) CreateChangeset(ctx context.Context, c *Changeset) (bool, error) {
	var exists bool

	remoteRepo := c.RemoteRepo.Metadata.(*bitbucketserver.Repo)
	targetRepo := c.TargetRepo.Metadata.(*bitbucketserver.Repo)

	pr := &bitbucketserver.PullRequest{Title: c.Title, Description: c.Body}

	pr.ToRef.Repository.Slug = targetRepo.Slug
	pr.ToRef.Repository.ID = targetRepo.ID
	pr.ToRef.Repository.Project.Key = targetRepo.Project.Key
	pr.ToRef.ID = git.EnsureRefPrefix(c.BaseRef)

	pr.FromRef.Repository.Slug = remoteRepo.Slug
	pr.FromRef.Repository.ID = remoteRepo.ID
	pr.FromRef.Repository.Project.Key = remoteRepo.Project.Key
	pr.FromRef.ID = git.EnsureRefPrefix(c.HeadRef)

	err := s.client.CreatePullRequest(ctx, pr)
	if err != nil {
		var e *bitbucketserver.ErrAlreadyExists
		if errors.As(err, &e) {
			if e.Existing == nil {
				return exists, errors.Errorf("existing PR is nil")
			}
			log15.Info("Existing PR extracted", "ID", e.Existing.ID)
			pr = e.Existing
			exists = true
		} else {
			return exists, err
		}
	}

	if err := s.loadPullRequestData(ctx, pr); err != nil {
		return false, errors.Wrap(err, "loading extra metadata")
	}
	if err = c.SetMetadata(pr); err != nil {
		return false, errors.Wrap(err, "setting changeset metadata")
	}

	return exists, nil
}

// CloseChangeset closes the given *Changeset on the code host and updates the
// Metadata column in the *batches.Changeset to the newly closed pull request.
func (s BitbucketServerSource) CloseChangeset(ctx context.Context, c *Changeset) error {
	pr, ok := c.Changeset.Metadata.(*bitbucketserver.PullRequest)
	if !ok {
		return errors.New("Changeset is not a Bitbucket Server pull request")
	}

	err := s.client.DeclinePullRequest(ctx, pr)
	if err != nil {
		return err
	}

	return c.Changeset.SetMetadata(pr)
}

// LoadChangeset loads the latest state of the given Changeset from the codehost.
func (s BitbucketServerSource) LoadChangeset(ctx context.Context, cs *Changeset) error {
	repo := cs.TargetRepo.Metadata.(*bitbucketserver.Repo)
	number, err := strconv.Atoi(cs.ExternalID)
	if err != nil {
		return err
	}

	pr := &bitbucketserver.PullRequest{ID: number}
	pr.ToRef.Repository.Slug = repo.Slug
	pr.ToRef.Repository.Project.Key = repo.Project.Key

	err = s.client.LoadPullRequest(ctx, pr)
	if err != nil {
		if err == bitbucketserver.ErrPullRequestNotFound {
			return ChangesetNotFoundError{Changeset: cs}
		}

		return err
	}

	err = s.loadPullRequestData(ctx, pr)
	if err != nil {
		return errors.Wrap(err, "loading pull request data")
	}
	if err = cs.SetMetadata(pr); err != nil {
		return errors.Wrap(err, "setting changeset metadata")
	}

	return nil
}

func (s BitbucketServerSource) loadPullRequestData(ctx context.Context, pr *bitbucketserver.PullRequest) error {
	if err := s.client.LoadPullRequestActivities(ctx, pr); err != nil {
		return errors.Wrap(err, "loading pr activities")
	}

	if err := s.client.LoadPullRequestCommits(ctx, pr); err != nil {
		return errors.Wrap(err, "loading pr commits")
	}

	if err := s.client.LoadPullRequestBuildStatuses(ctx, pr); err != nil {
		return errors.Wrap(err, "loading pr build status")
	}

	return nil
}

func (s BitbucketServerSource) UpdateChangeset(ctx context.Context, c *Changeset) error {
	pr, ok := c.Changeset.Metadata.(*bitbucketserver.PullRequest)
	if !ok {
		return errors.New("Changeset is not a Bitbucket Server pull request")
	}

	update := &bitbucketserver.UpdatePullRequestInput{
		PullRequestID: strconv.Itoa(pr.ID),
		Title:         c.Title,
		Description:   c.Body,
		Version:       pr.Version,
	}
	update.ToRef.ID = c.BaseRef
	update.ToRef.Repository.Slug = pr.ToRef.Repository.Slug
	update.ToRef.Repository.Project.Key = pr.ToRef.Repository.Project.Key

	updated, err := s.client.UpdatePullRequest(ctx, update)
	if err != nil {
		return err
	}

	return c.Changeset.SetMetadata(updated)
}

// ReopenChangeset reopens the *Changeset on the code host and updates the
// Metadata column in the *batches.Changeset.
func (s BitbucketServerSource) ReopenChangeset(ctx context.Context, c *Changeset) error {
	pr, ok := c.Changeset.Metadata.(*bitbucketserver.PullRequest)
	if !ok {
		return errors.New("Changeset is not a Bitbucket Server pull request")
	}

	if err := s.client.ReopenPullRequest(ctx, pr); err != nil {
		return err
	}

	return c.Changeset.SetMetadata(pr)
}

// CreateComment posts a comment on the Changeset.
func (s BitbucketServerSource) CreateComment(ctx context.Context, c *Changeset, text string) error {
	pr, ok := c.Changeset.Metadata.(*bitbucketserver.PullRequest)
	if !ok {
		return errors.New("Changeset is not a Bitbucket Server pull request")
	}

	return s.client.CreatePullRequestComment(ctx, pr, text)
}

// MergeChangeset merges a Changeset on the code host, if in a mergeable state.
// The squash parameter is ignored, as Bitbucket Server does not support
// squash merges.
func (s BitbucketServerSource) MergeChangeset(ctx context.Context, c *Changeset, squash bool) error {
	pr, ok := c.Changeset.Metadata.(*bitbucketserver.PullRequest)
	if !ok {
		return errors.New("Changeset is not a Bitbucket Server pull request")
	}

	if err := s.client.MergePullRequest(ctx, pr); err != nil {
		if errors.Is(err, bitbucketserver.ErrNotMergeable) {
			return &ChangesetNotMergeableError{ErrorMsg: err.Error()}
		}
		return err
	}

	return c.Changeset.SetMetadata(pr)
}

func (s BitbucketServerSource) GetChangesetForkRepo(ctx context.Context, targetRepo *types.Repo) (*types.Repo, error) {
	parent := targetRepo.Metadata.(*bitbucketserver.Repo)

	// Ascertain the user name for the token we're using.
	user, err := s.AuthenticatedUsername(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "getting username")
	}

	// See if we already have a fork.
	fork, err := s.getUserFork(ctx, parent, user)
	log15.Info("getUserFork", "fork", fork, "err", err)
	if err != nil {
		return nil, errors.Wrapf(err, "getting user fork for %q", user)
	}

	// If not, then we need to create a fork.
	if fork == nil {
		fork, err = s.client.CreateFork(ctx, parent.Project.Key, parent.Slug, bitbucketserver.CreateForkInput{})
		if err != nil {
			return nil, errors.Wrapf(err, "creating user fork for %q", user)
		}
	}

	// We have a fork! Now we have to make a *types.Repo look legitimate.
	// bitbucketServerCloneURL() ultimately only looks at the
	// bitbucketserver.Repo in the Metadata field, so we'll replace that with
	// the fork's metadata.
	remoteRepo := *targetRepo
	remoteRepo.Metadata = fork

	return &remoteRepo, nil
}

func (s BitbucketServerSource) getUserFork(ctx context.Context, parent *bitbucketserver.Repo, user string) (*bitbucketserver.Repo, error) {
	var pageToken *bitbucketserver.PageToken
	for pageToken.HasMore() {
		var forks []*bitbucketserver.Repo
		var err error

		forks, pageToken, err = s.client.Forks(ctx, parent.Project.Key, parent.Slug, pageToken)
		if err != nil {
			return nil, errors.Wrap(err, "retrieving forks")
		}

		for _, fork := range forks {
			js, err := json.Marshal(fork)
			if err != nil {
				panic(err)
			}
			log15.Info("comparing fork", "fork", string(js), "fork.Owner", fork.Owner, "user", user)
			// This looks insane, because the underlying API is insane: there's
			// an Owner field that is _sometimes_ populated on the fork, but not
			// always, and without it the only reference to the username is the
			// self link back to the user profile on the project.
			if fork.Project.Type == "PERSONAL" {
				for _, link := range fork.Project.Links.Self {
					log15.Info("comparing self link", "href", link.Href, "user", user)
					if strings.HasSuffix(link.Href, "/"+user) {
						return fork, nil
					}
				}
			}
		}
	}

	return nil, nil
}
