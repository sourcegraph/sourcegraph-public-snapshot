// Package github implements issues.Service using GitHub API client.
package github

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strings"

	"github.com/google/go-github/github"
	"golang.org/x/net/context"
	"src.sourcegraph.com/apps/tracker/issues"
)

// NewService creates a GitHub-backed issues.Service using given GitHub client.
// At this time it infers the current user from the client (its authentication info), and cannot be used to serve multiple users.
func NewService(client *github.Client) issues.Service {
	if client == nil {
		client = github.NewClient(nil)
	}

	s := service{
		cl: client,
	}

	if user, _, err := client.Users.Get(""); err == nil {
		u := ghUser(user)
		s.currentUser = &u
	} else if ghErr, ok := err.(*github.ErrorResponse); ok && ghErr.Response.StatusCode == http.StatusUnauthorized {
	} else {
		s.currentUserErr = err
	}

	return s
}

type service struct {
	cl *github.Client

	currentUser    *issues.User
	currentUserErr error
}

// We use 0 as a special ID for the comment that is the issue description. This comment is edited differently.
const issueDescriptionCommentID = uint64(0)

func (s service) List(_ context.Context, rs issues.RepoSpec, opt issues.IssueListOptions) ([]issues.Issue, error) {
	repo := ghRepoSpec(rs)
	ghOpt := github.IssueListByRepoOptions{}
	switch opt.State {
	case issues.OpenState:
		// Do nothing, this is the GitHub default.
	case issues.ClosedState:
		ghOpt.State = "closed"
	}
	ghIssuesAndPRs, _, err := s.cl.Issues.ListByRepo(repo.Owner, repo.Repo, &ghOpt)
	if err != nil {
		return nil, err
	}

	var is []issues.Issue
	for _, issue := range ghIssuesAndPRs {
		// Filter out PRs.
		if issue.PullRequestLinks != nil {
			continue
		}

		is = append(is, issues.Issue{
			ID:    uint64(*issue.Number),
			State: issues.State(*issue.State),
			Title: *issue.Title,
			Comment: issues.Comment{
				User:      ghUser(issue.User),
				CreatedAt: *issue.CreatedAt,
			},
			Replies: *issue.Comments,
		})
	}

	return is, nil
}

func (s service) Count(_ context.Context, rs issues.RepoSpec, opt issues.IssueListOptions) (uint64, error) {
	repo := ghRepoSpec(rs)
	var count uint64

	// Count Issues and PRs (since there appears to be no way to count just issues in GitHub API).
	{
		ghOpt := github.IssueListByRepoOptions{ListOptions: github.ListOptions{PerPage: 1}}
		switch opt.State {
		case issues.OpenState:
			// Do nothing, this is the GitHub default.
		case issues.ClosedState:
			ghOpt.State = "closed"
		}
		ghIssuesAndPRs, ghIssuesAndPRsResp, err := s.cl.Issues.ListByRepo(repo.Owner, repo.Repo, &ghOpt)
		if err != nil {
			return 0, err
		}
		if ghIssuesAndPRsResp.LastPage != 0 {
			count = uint64(ghIssuesAndPRsResp.LastPage)
		} else {
			count = uint64(len(ghIssuesAndPRs))
		}
	}

	// Subtract PRs.
	{
		ghOpt := github.PullRequestListOptions{ListOptions: github.ListOptions{PerPage: 1}}
		switch opt.State {
		case issues.OpenState:
			// Do nothing, this is the GitHub default.
		case issues.ClosedState:
			ghOpt.State = "closed"
		}
		ghPRs, ghPRsResp, err := s.cl.PullRequests.List(repo.Owner, repo.Repo, &ghOpt)
		if err != nil {
			return 0, err
		}
		if ghPRsResp.LastPage != 0 {
			count -= uint64(ghPRsResp.LastPage)
		} else {
			count -= uint64(len(ghPRs))
		}
	}

	return count, nil
}

// canEdit returns nil error if currentUser is authorized to edit an entry created by authorUID.
// It returns os.ErrPermission or an error that happened in other cases.
func (s service) canEdit(repo repoSpec, authorLogin string) error {
	if s.currentUser == nil {
		// Not logged in, cannot edit anything.
		return os.ErrPermission
	}
	if s.currentUser.Login == authorLogin {
		// If you're the author, you can always edit it.
		return nil
	}
	isCollaborator, _, err := s.cl.Repositories.IsCollaborator(repo.Owner, repo.Repo, s.currentUser.Login)
	if err != nil {
		return err
	}
	switch isCollaborator {
	case true:
		// If you have write access (or greater), you can edit.
		return nil
	default:
		return os.ErrPermission
	}
}

func (s service) Get(_ context.Context, rs issues.RepoSpec, id uint64) (issues.Issue, error) {
	repo := ghRepoSpec(rs)
	issue, _, err := s.cl.Issues.Get(repo.Owner, repo.Repo, int(id))
	if err != nil {
		return issues.Issue{}, err
	}

	return issues.Issue{
		ID:    uint64(*issue.Number),
		State: issues.State(*issue.State),
		Title: *issue.Title,
		Comment: issues.Comment{
			User:      ghUser(issue.User),
			CreatedAt: *issue.CreatedAt,
			Editable:  nil == s.canEdit(repo, *issue.User.Login),
		},
	}, nil
}

func (s service) ListComments(_ context.Context, rs issues.RepoSpec, id uint64, opt interface{}) ([]issues.Comment, error) {
	repo := ghRepoSpec(rs)
	var comments []issues.Comment

	issue, _, err := s.cl.Issues.Get(repo.Owner, repo.Repo, int(id))
	if err != nil {
		return comments, err
	}
	comments = append(comments, issues.Comment{
		ID:        issueDescriptionCommentID,
		User:      ghUser(issue.User),
		CreatedAt: *issue.CreatedAt,
		Body:      *issue.Body,
		Editable:  nil == s.canEdit(repo, *issue.User.Login),
	})

	ghComments, _, err := s.cl.Issues.ListComments(repo.Owner, repo.Repo, int(id), nil) // TODO: Pagination.
	if err != nil {
		return comments, err
	}
	for _, comment := range ghComments {
		comments = append(comments, issues.Comment{
			ID:        uint64(*comment.ID),
			User:      ghUser(comment.User),
			CreatedAt: *comment.CreatedAt,
			Body:      *comment.Body,
			Editable:  nil == s.canEdit(repo, *comment.User.Login),
		})
	}

	return comments, nil
}

func (s service) ListEvents(_ context.Context, rs issues.RepoSpec, id uint64, opt interface{}) ([]issues.Event, error) {
	repo := ghRepoSpec(rs)
	var events []issues.Event

	ghEvents, _, err := s.cl.Issues.ListIssueEvents(repo.Owner, repo.Repo, int(id), nil) // TODO: Pagination.
	if err != nil {
		return events, err
	}
	for _, event := range ghEvents {
		et := issues.EventType(*event.Event)
		if !et.Valid() {
			continue
		}
		e := issues.Event{
			Actor:     ghUser(event.Actor),
			CreatedAt: *event.CreatedAt,
			Type:      et,
		}
		switch et {
		case issues.Renamed:
			e.Rename = &issues.Rename{
				From: *event.Rename.From,
				To:   *event.Rename.To,
			}
		}
		events = append(events, e)
	}

	return events, nil
}

func (s service) CreateComment(_ context.Context, rs issues.RepoSpec, id uint64, c issues.Comment) (issues.Comment, error) {
	repo := ghRepoSpec(rs)
	comment, _, err := s.cl.Issues.CreateComment(repo.Owner, repo.Repo, int(id), &github.IssueComment{
		Body: &c.Body,
	})
	if err != nil {
		return issues.Comment{}, err
	}

	return issues.Comment{
		ID:        uint64(*comment.ID),
		User:      ghUser(comment.User),
		CreatedAt: *comment.CreatedAt,
		Body:      *comment.Body,
	}, nil
}

func (s service) Create(_ context.Context, rs issues.RepoSpec, i issues.Issue) (issues.Issue, error) {
	repo := ghRepoSpec(rs)
	issue, _, err := s.cl.Issues.Create(repo.Owner, repo.Repo, &github.IssueRequest{
		Title: &i.Title,
		Body:  &i.Body,
	})
	if err != nil {
		return issues.Issue{}, err
	}

	return issues.Issue{
		ID:    uint64(*issue.Number),
		State: issues.State(*issue.State),
		Title: *issue.Title,
		Comment: issues.Comment{
			ID:        issueDescriptionCommentID,
			User:      ghUser(issue.User),
			CreatedAt: *issue.CreatedAt,
		},
	}, nil
}

func (s service) Edit(_ context.Context, rs issues.RepoSpec, id uint64, ir issues.IssueRequest) (issues.Issue, error) {
	// TODO: Why Validate here but not Create, etc.? Figure this out. Might only be needed in fs implementation.
	if err := ir.Validate(); err != nil {
		return issues.Issue{}, err
	}
	repo := ghRepoSpec(rs)

	ghIR := github.IssueRequest{
		Title: ir.Title,
	}
	if ir.State != nil {
		state := string(*ir.State)
		ghIR.State = &state
	}

	issue, _, err := s.cl.Issues.Edit(repo.Owner, repo.Repo, int(id), &ghIR)
	if err != nil {
		return issues.Issue{}, err
	}

	return issues.Issue{
		ID:    uint64(*issue.Number),
		State: issues.State(*issue.State),
		Title: *issue.Title,
		Comment: issues.Comment{
			ID:        issueDescriptionCommentID,
			User:      ghUser(issue.User),
			CreatedAt: *issue.CreatedAt,
		},
	}, nil
}

func (s service) EditComment(_ context.Context, rs issues.RepoSpec, id uint64, c issues.Comment) (issues.Comment, error) {
	// TODO: Why Validate here but not CreateComment, etc.? Figure this out. Might only be needed in fs implementation.
	if err := c.Validate(); err != nil {
		return issues.Comment{}, err
	}
	repo := ghRepoSpec(rs)

	if c.ID == issueDescriptionCommentID {
		// Use Issues.Edit() API to edit comment 0 (the issue description).
		issue, _, err := s.cl.Issues.Edit(repo.Owner, repo.Repo, int(id), &github.IssueRequest{
			Body: &c.Body,
		})
		if err != nil {
			return issues.Comment{}, err
		}

		return issues.Comment{
			ID:        issueDescriptionCommentID,
			User:      ghUser(issue.User),
			CreatedAt: *issue.CreatedAt,
			Body:      *issue.Body,
		}, nil
	}

	// GitHub API uses comment ID and doesn't need issue ID. Comment IDs are unique per repo (not per issue).
	comment, _, err := s.cl.Issues.EditComment(repo.Owner, repo.Repo, int(c.ID), &github.IssueComment{
		Body: &c.Body,
	})
	if err != nil {
		return issues.Comment{}, err
	}

	return issues.Comment{
		ID:        uint64(*comment.ID),
		User:      ghUser(comment.User),
		CreatedAt: *comment.CreatedAt,
		Body:      *comment.Body,
	}, nil
}

func (s service) CurrentUser(_ context.Context) (*issues.User, error) {
	return s.currentUser, s.currentUserErr
}

type repoSpec struct {
	Owner string
	Repo  string
}

func ghRepoSpec(repo issues.RepoSpec) repoSpec {
	ownerRepo := strings.Split(repo.URI, "/")
	if len(ownerRepo) != 2 {
		panic(fmt.Errorf(`RepoSpec is not of form "owner/repo": %v`, repo))
	}
	return repoSpec{
		Owner: ownerRepo[0],
		Repo:  ownerRepo[1],
	}
}

func ghUser(user *github.User) issues.User {
	return issues.User{
		Login:     *user.Login,
		AvatarURL: template.URL(*user.AvatarURL),
		HTMLURL:   template.URL(*user.HTMLURL),
	}
}
