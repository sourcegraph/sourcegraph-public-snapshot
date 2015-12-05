// Package github implements issues.Service using GitHub API client.
package github

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
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
		s.currentUserErr = nil
	} else if ghErr, ok := err.(*github.ErrorResponse); ok && ghErr.Response.StatusCode == http.StatusUnauthorized {
		// There's no authenticated user.
		s.currentUser = nil
		s.currentUserErr = nil
	} else {
		s.currentUser = nil
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
	case issues.StateFilter(issues.OpenState):
		// Do nothing, this is the GitHub default.
	case issues.StateFilter(issues.ClosedState):
		ghOpt.State = "closed"
	case issues.AllStates:
		ghOpt.State = "all"
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
	var ghState string
	switch opt.State {
	case issues.StateFilter(issues.OpenState):
		// Do nothing, this is the GitHub default.
	case issues.StateFilter(issues.ClosedState):
		ghState = "closed"
	case issues.AllStates:
		ghState = "all"
	}

	var count uint64

	// Count Issues and PRs (since there appears to be no way to count just issues in GitHub API).
	{
		ghOpt := github.IssueListByRepoOptions{
			State:       ghState,
			ListOptions: github.ListOptions{PerPage: 1},
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
		ghOpt := github.PullRequestListOptions{
			State:       ghState,
			ListOptions: github.ListOptions{PerPage: 1},
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

// TODO: Dedup.
func parseIssueSpec(issueAPIURL string) (issues.RepoSpec, int, error) {
	u, err := url.Parse(issueAPIURL)
	if err != nil {
		return issues.RepoSpec{}, 0, err
	}
	e := strings.Split(u.Path, "/")
	if len(e) < 5 {
		return issues.RepoSpec{}, 0, fmt.Errorf("unexpected path (too few elements): %q", u.Path)
	}
	if got, want := e[len(e)-2], "issues"; got != want {
		return issues.RepoSpec{}, 0, fmt.Errorf(`unexpected path element %q, expecting %q`, got, want)
	}
	id, err := strconv.Atoi(e[len(e)-1])
	if err != nil {
		return issues.RepoSpec{}, 0, err
	}
	return issues.RepoSpec{URI: e[len(e)-4] + "/" + e[len(e)-3]}, id, nil
}

// markRead marks the specified issue as read.
func (s service) markRead(repo repoSpec, id uint64) error {
	ns, _, err := s.cl.Activity.ListRepositoryNotifications(repo.Owner, repo.Repo, nil)
	if err != nil {
		return fmt.Errorf("failed to ListRepositoryNotifications: %v", err)
	}

	for _, n := range ns {
		if *n.Subject.Type != "Issue" {
			continue
		}
		_, issueID, err := parseIssueSpec(*n.Subject.URL)
		if err != nil {
			return fmt.Errorf("failed to parseIssueSpec: %v", err)
		}
		if uint64(issueID) != id {
			continue
		}

		_, err = s.cl.Activity.MarkThreadRead(*n.ID)
		if err != nil {
			return fmt.Errorf("failed to MarkThreadRead: %v", err)
		}
		break
	}

	return nil
}

func (s service) Get(_ context.Context, rs issues.RepoSpec, id uint64) (issues.Issue, error) {
	repo := ghRepoSpec(rs)
	issue, _, err := s.cl.Issues.Get(repo.Owner, repo.Repo, int(id))
	if err != nil {
		return issues.Issue{}, err
	}

	if s.currentUser != nil {
		// Mark as read.
		err = s.markRead(repo, id)
		if err != nil {
			log.Println("service.Get: failed to markRead:", err)
		}
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

	ghOpt := &github.IssueListCommentsOptions{}
	for {
		ghComments, resp, err := s.cl.Issues.ListComments(repo.Owner, repo.Repo, int(id), ghOpt)
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
		if resp.NextPage == 0 {
			break
		}
		ghOpt.ListOptions.Page = resp.NextPage
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
			ID:        uint64(*event.ID),
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
		Editable:  true, // You can always edit comments you've created.
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
			Editable:  true, // You can always edit issues you've created.
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
			Editable:  true, // You can always edit issues you've edited.
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
			Editable:  true, // You can always edit comments you've edited.
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
		Editable:  true, // You can always edit comments you've edited.
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
		UserSpec: issues.UserSpec{
			ID:     uint64(*user.ID),
			Domain: "github.com",
		},
		Login:     *user.Login,
		AvatarURL: template.URL(*user.AvatarURL),
		HTMLURL:   template.URL(*user.HTMLURL),
	}
}
