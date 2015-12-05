// Package github implements notifications.Service using GitHub API client.
package github

import (
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/google/go-github/github"
	"golang.org/x/net/context"
	"src.sourcegraph.com/apps/notifications/notifications"
	"src.sourcegraph.com/apps/tracker/issues"
)

// NewService creates a GitHub-backed notifications.Service using given GitHub client.
// At this time it infers the current user from the client (its authentication info), and cannot be used to serve multiple users.
func NewService(client *github.Client) notifications.Service {
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

func (s service) CurrentUser(_ context.Context) (*issues.User, error) {
	return s.currentUser, s.currentUserErr
}

type repoSpec struct {
	Owner string
	Repo  string
}

func (s service) List(ctx context.Context, opt interface{}) ([]notifications.Notification, error) {
	var ns []notifications.Notification

	ghNotifications, _, err := s.cl.Activity.ListNotifications(nil)
	if err != nil {
		return nil, err
	}
	for _, n := range ghNotifications {
		notification := notifications.Notification{
			Type:      *n.Subject.Type,
			RepoSpec:  issues.RepoSpec{URI: *n.Repository.FullName},
			Title:     *n.Subject.Title,
			UpdatedAt: *n.UpdatedAt,
		}

		switch *n.Subject.Type {
		case "Issue":
			var err error
			notification.HTMLURL, err = s.getIssueURL(*n.Subject)
			if err != nil {
				return ns, err
			}

			notification.State, err = s.getIssueState(*n.Subject.URL)
			if err != nil {
				return ns, err
			}
		}

		ns = append(ns, notification)
	}

	return ns, nil
}

func (s service) getIssueState(issueAPIURL string) (string, error) {
	req, err := s.cl.NewRequest("GET", issueAPIURL, nil)
	if err != nil {
		return "", err
	}
	issue := new(github.Issue)
	_, err = s.cl.Do(req, issue)
	if err != nil {
		return "", err
	}
	if issue.State == nil {
		return "", fmt.Errorf("for some reason issue.State is nil for %q: %v", issueAPIURL, issue)
	}
	return *issue.State, nil
}

func (s service) getIssueURL(n github.NotificationSubject) (template.URL, error) {
	rs, issueID, err := parseIssueSpec(*n.URL)
	if err != nil {
		return "", err
	}
	var fragment string
	if _, commentID, err := parseIssueCommentSpec(*n.LatestCommentURL); err == nil {
		fragment = fmt.Sprintf("#comment-%d", commentID)
	}
	return template.URL(fmt.Sprintf("/github.com/%s/issues/%d%s", rs.URI, issueID, fragment)), nil
}

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

func parseIssueCommentSpec(issueAPIURL string) (issues.RepoSpec, int, error) {
	u, err := url.Parse(issueAPIURL)
	if err != nil {
		return issues.RepoSpec{}, 0, err
	}
	e := strings.Split(u.Path, "/")
	if len(e) < 6 {
		return issues.RepoSpec{}, 0, fmt.Errorf("unexpected path (too few elements): %q", u.Path)
	}
	if got, want := e[len(e)-2], "comments"; got != want {
		return issues.RepoSpec{}, 0, fmt.Errorf(`unexpected path element %q, expecting %q`, got, want)
	}
	if got, want := e[len(e)-3], "issues"; got != want {
		return issues.RepoSpec{}, 0, fmt.Errorf(`unexpected path element %q, expecting %q`, got, want)
	}
	id, err := strconv.Atoi(e[len(e)-1])
	if err != nil {
		return issues.RepoSpec{}, 0, err
	}
	return issues.RepoSpec{URI: e[len(e)-5] + "/" + e[len(e)-4]}, id, nil
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
