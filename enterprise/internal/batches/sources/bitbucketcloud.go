package sources

import (
	"context"
	"strconv"

	"github.com/cockroachdb/errors"
	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketcloud"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
	"github.com/sourcegraph/sourcegraph/schema"
)

type BitbucketCloudSource struct {
	client *bitbucketcloud.Client
	au     auth.Authenticator
}

// NewBitbucketCloudSource returns a new BitbucketCloudSource from the given external service.
func NewBitbucketCloudSource(svc *types.ExternalService, cf *httpcli.Factory) (*BitbucketCloudSource, error) {
	var c schema.BitbucketCloudConnection
	if err := jsonc.Unmarshal(svc.Config, &c); err != nil {
		return nil, errors.Errorf("external service id=%d config error: %s", svc.ID, err)
	}
	return newBitbucketCloudSource(&c, cf, nil)
}

func newBitbucketCloudSource(c *schema.BitbucketCloudConnection, cf *httpcli.Factory, au auth.Authenticator) (*BitbucketCloudSource, error) {
	if cf == nil {
		cf = httpcli.ExternalClientFactory
	}

	cli, err := cf.Doer()
	if err != nil {
		return nil, err
	}

	client, err := bitbucketcloud.NewClient(c, cli)
	if err != nil {
		return nil, err
	}

	if au != nil {
		client = client.WithAuthenticator(au)
	}

	return &BitbucketCloudSource{
		au:     client.Auth,
		client: client,
	}, nil
}

func (s BitbucketCloudSource) GitserverPushConfig(ctx context.Context, store database.ExternalServiceStore, repo *types.Repo) (*protocol.PushConfig, error) {
	return gitserverPushConfig(ctx, store, repo, s.au)
}

func (s BitbucketCloudSource) WithAuthenticator(a auth.Authenticator) (ChangesetSource, error) {
	switch a.(type) {
	case *auth.BasicAuth,
		// TODO: Those don't work, remove them. It currently breaks the create credential modal though.
		*auth.OAuthBearerToken,
		*auth.OAuthBearerTokenWithSSH,
		*auth.BasicAuthWithSSH:
		// *bitbucketserver.SudoableOAuthClient:
		break

	default:
		return nil, newUnsupportedAuthenticatorError("BitbucketCloudSource", a)
	}

	return &BitbucketCloudSource{
		client: s.client.WithAuthenticator(a),
		au:     a,
	}, nil
}

// AuthenticatedUsername uses the underlying bitbucketserver.Client to get the
// username belonging to the credentials associated with the
// BitbucketCloudSource.
func (s BitbucketCloudSource) AuthenticatedUsername(ctx context.Context) (string, error) {
	return "eseliger", nil // s.client.AuthenticatedUsername(ctx)
}

func (s BitbucketCloudSource) ValidateAuthenticator(ctx context.Context) error {
	return nil
	// _, err := s.client.AuthenticatedUsername(ctx)
	// return err
}

// CreateChangeset creates the given *Changeset in the code host.
func (s BitbucketCloudSource) CreateChangeset(ctx context.Context, c *Changeset) (bool, error) {
	var exists bool

	repo := c.Repo.Metadata.(*bitbucketcloud.Repo)

	pr := &bitbucketcloud.PullRequest{Title: c.Title, Description: c.Body}

	pr.ToRef.Repository.Slug = repo.Slug
	pr.ToRef.Repository.ID = repo.ID
	pr.ToRef.Repository.Project.Key = repo.Project.Key
	pr.ToRef.ID = git.EnsureRefPrefix(c.BaseRef)

	pr.FromRef.Repository.Slug = repo.Slug
	pr.FromRef.Repository.ID = repo.ID
	pr.FromRef.Repository.Project.Key = repo.Project.Key
	pr.FromRef.ID = git.EnsureRefPrefix(c.HeadRef)

	err := s.client.CreatePullRequest(ctx, pr)
	if err != nil {
		var e *bitbucketcloud.ErrAlreadyExists
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
func (s BitbucketCloudSource) CloseChangeset(ctx context.Context, c *Changeset) error {
	return nil
	// pr, ok := c.Changeset.Metadata.(*bitbucketserver.PullRequest)
	// if !ok {
	// 	return errors.New("Changeset is not a Bitbucket Server pull request")
	// }

	// err := s.client.DeclinePullRequest(ctx, pr)
	// if err != nil {
	// 	return err
	// }

	// return c.Changeset.SetMetadata(pr)
}

// LoadChangeset loads the latest state of the given Changeset from the codehost.
func (s BitbucketCloudSource) LoadChangeset(ctx context.Context, cs *Changeset) error {
	repo := cs.Repo.Metadata.(*bitbucketcloud.Repo)
	number, err := strconv.Atoi(cs.ExternalID)
	if err != nil {
		return err
	}

	pr := &bitbucketcloud.PullRequest{ID: number}
	pr.ToRef.Repository.Slug = repo.Slug
	pr.ToRef.Repository.Project.Key = repo.Project.Key

	err = s.client.LoadPullRequest(ctx, pr)
	if err != nil {
		if err == bitbucketcloud.ErrPullRequestNotFound {
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

func (s BitbucketCloudSource) loadPullRequestData(ctx context.Context, pr *bitbucketcloud.PullRequest) error {
	// if err := s.client.LoadPullRequestActivities(ctx, pr); err != nil {
	// 	return errors.Wrap(err, "loading pr activities")
	// }

	// if err := s.client.LoadPullRequestCommits(ctx, pr); err != nil {
	// 	return errors.Wrap(err, "loading pr commits")
	// }

	// if err := s.client.LoadPullRequestBuildStatuses(ctx, pr); err != nil {
	// 	return errors.Wrap(err, "loading pr build status")
	// }

	return nil
}

func (s BitbucketCloudSource) UpdateChangeset(ctx context.Context, c *Changeset) error {
	return nil
	// pr, ok := c.Changeset.Metadata.(*bitbucketserver.PullRequest)
	// if !ok {
	// 	return errors.New("Changeset is not a Bitbucket Server pull request")
	// }

	// update := &bitbucketserver.UpdatePullRequestInput{
	// 	PullRequestID: strconv.Itoa(pr.ID),
	// 	Title:         c.Title,
	// 	Description:   c.Body,
	// 	Version:       pr.Version,
	// }
	// update.ToRef.ID = c.BaseRef
	// update.ToRef.Repository.Slug = pr.ToRef.Repository.Slug
	// update.ToRef.Repository.Project.Key = pr.ToRef.Repository.Project.Key

	// updated, err := s.client.UpdatePullRequest(ctx, update)
	// if err != nil {
	// 	return err
	// }

	// return c.Changeset.SetMetadata(updated)
}

// ReopenChangeset reopens the *Changeset on the code host and updates the
// Metadata column in the *batches.Changeset.
func (s BitbucketCloudSource) ReopenChangeset(ctx context.Context, c *Changeset) error {
	return nil
	// pr, ok := c.Changeset.Metadata.(*bitbucketserver.PullRequest)
	// if !ok {
	// 	return errors.New("Changeset is not a Bitbucket Server pull request")
	// }

	// if err := s.client.ReopenPullRequest(ctx, pr); err != nil {
	// 	return err
	// }

	// return c.Changeset.SetMetadata(pr)
}

// CreateComment posts a comment on the Changeset.
func (s BitbucketCloudSource) CreateComment(ctx context.Context, c *Changeset, text string) error {
	return nil
	// pr, ok := c.Changeset.Metadata.(*bitbucketserver.PullRequest)
	// if !ok {
	// 	return errors.New("Changeset is not a Bitbucket Server pull request")
	// }

	// return s.client.CreatePullRequestComment(ctx, pr, text)
}

// MergeChangeset merges a Changeset on the code host, if in a mergeable state.
// The squash parameter is ignored, as Bitbucket Server does not support
// squash merges.
func (s BitbucketCloudSource) MergeChangeset(ctx context.Context, c *Changeset, squash bool) error {
	return nil
	// pr, ok := c.Changeset.Metadata.(*bitbucketserver.PullRequest)
	// if !ok {
	// 	return errors.New("Changeset is not a Bitbucket Server pull request")
	// }

	// if err := s.client.MergePullRequest(ctx, pr); err != nil {
	// 	if errors.Is(err, bitbucketserver.ErrNotMergeable) {
	// 		return &ChangesetNotMergeableError{ErrorMsg: err.Error()}
	// 	}
	// 	return err
	// }

	// return c.Changeset.SetMetadata(pr)
}
