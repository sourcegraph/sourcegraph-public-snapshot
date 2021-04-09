package sources

import (
	"context"
	"strconv"

	"github.com/pkg/errors"

	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

type GithubSource struct {
	v4Client github.V4Client
}

// CreateChangeset creates the given changeset on the code host.
func (s GithubSource) CreateChangeset(ctx context.Context, c *Changeset) (bool, error) {
	input := buildCreatePullRequestInput(c)
	return s.createChangeset(ctx, c, input)
}

// CreateDraftChangeset creates the given changeset on the code host in draft mode.
func (s GithubSource) CreateDraftChangeset(ctx context.Context, c *Changeset) (bool, error) {
	input := buildCreatePullRequestInput(c)
	input.Draft = true
	return s.createChangeset(ctx, c, input)
}

func buildCreatePullRequestInput(c *Changeset) *github.CreatePullRequestInput {
	return &github.CreatePullRequestInput{
		RepositoryID: c.Repo.Metadata.(*github.Repository).ID,
		Title:        c.Title,
		Body:         c.Body,
		HeadRefName:  git.AbbreviateRef(c.HeadRef),
		BaseRefName:  git.AbbreviateRef(c.BaseRef),
	}
}

func (s GithubSource) createChangeset(ctx context.Context, c *Changeset, prInput *github.CreatePullRequestInput) (bool, error) {
	var exists bool
	pr, err := s.v4Client.CreatePullRequest(ctx, prInput)
	if err != nil {
		if err != github.ErrPullRequestAlreadyExists {
			return exists, err
		}
		repo := c.Repo.Metadata.(*github.Repository)
		owner, name, err := github.SplitRepositoryNameWithOwner(repo.NameWithOwner)
		if err != nil {
			return exists, errors.Wrap(err, "getting repo owner and name")
		}
		pr, err = s.v4Client.GetOpenPullRequestByRefs(ctx, owner, name, c.BaseRef, c.HeadRef)
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

	err := s.v4Client.ClosePullRequest(ctx, pr)
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

	err := s.v4Client.MarkPullRequestReadyForReview(ctx, pr)
	if err != nil {
		return err
	}

	return c.Changeset.SetMetadata(pr)
}

// LoadChangeset loads the latest state of the given Changeset from the codehost.
func (s GithubSource) LoadChangeset(ctx context.Context, cs *Changeset) error {
	repo := cs.Repo.Metadata.(*github.Repository)
	number, err := strconv.ParseInt(cs.ExternalID, 10, 64)
	if err != nil {
		return errors.Wrap(err, "parsing changeset external id")
	}

	pr := &github.PullRequest{
		RepoWithOwner: repo.NameWithOwner,
		Number:        number,
	}

	if err := s.v4Client.LoadPullRequest(ctx, pr); err != nil {
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

	updated, err := s.v4Client.UpdatePullRequest(ctx, &github.UpdatePullRequestInput{
		PullRequestID: pr.ID,
		Title:         c.Title,
		Body:          c.Body,
		BaseRefName:   git.AbbreviateRef(c.BaseRef),
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

	err := s.v4Client.ReopenPullRequest(ctx, pr)
	if err != nil {
		return err
	}

	return c.Changeset.SetMetadata(pr)
}
