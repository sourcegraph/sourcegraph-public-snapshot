package a8n

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	gh "github.com/google/go-github/v28/github"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	"github.com/sourcegraph/sourcegraph/internal/a8n"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/schema"
	"gopkg.in/inconshreveable/log15.v2"
)

// GitHubWebhook receives GitHub organization webhook events that are
// relevant to a8n, normalizes those events into ChangesetEvents
// and upserts them to the database.
type GitHubWebhook struct {
	Store *Store
	Repos repos.Store
	Now   func() time.Time
}

// ServeHTTP implements the http.Handler interface.
func (h *GitHubWebhook) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	e, err := h.parseEvent(r)
	if err != nil {
		respond(w, err.code, err)
		return
	}

	pr, ev := h.convertEvent(e)
	if pr == 0 || ev == nil {
		respond(w, http.StatusOK, nil) // Nothing to do
		return
	}

	if err := h.upsertChangesetEvent(r.Context(), pr, ev); err != nil {
		respond(w, http.StatusInternalServerError, err)
	}
}

func (h *GitHubWebhook) parseEvent(r *http.Request) (interface{}, *httpError) {
	args := repos.StoreListExternalServicesArgs{Kinds: []string{"GITHUB"}}
	es, err := h.Repos.ListExternalServices(r.Context(), args)
	if err != nil {
		return nil, &httpError{http.StatusInternalServerError, err}
	}

	payload, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, &httpError{http.StatusInternalServerError, err}
	}

	// 🚨 SECURITY: Try to authenticate the request with any of the stored secrets
	// in GitHub external services config. Since n is usually small here,
	// it's ok for this to be have linear complexity.
	// If there are no secrets or no secret managed to authenticate the request,
	// we return a 401 to the client.

	var secrets [][]byte
	for _, e := range es {
		c, _ := e.Configuration()
		for _, hook := range c.(*schema.GitHubConnection).Webhooks {
			secrets = append(secrets, []byte(hook.Secret))
		}
	}

	sig := r.Header.Get("X-Hub-Signature")
	for _, secret := range secrets {
		if err = gh.ValidateSignature(sig, payload, secret); err == nil {
			break
		}
	}

	if len(secrets) == 0 || err != nil {
		return nil, &httpError{http.StatusUnauthorized, err}
	}

	e, err := gh.ParseWebHook(gh.WebHookType(r), payload)
	if err != nil {
		return nil, &httpError{http.StatusBadRequest, err}
	}

	return e, nil
}

func (h *GitHubWebhook) convertEvent(theirs interface{}) (pr int64, ours interface{ Key() string }) {
	switch e := theirs.(type) {
	case *gh.IssueCommentEvent:
		pr = int64(*e.Issue.Number)
		return pr, h.issueComment(e)

	case *gh.PullRequestEvent:
		pr = int64(*e.Number)

		switch *e.Action {
		case "assigned":
			ours = h.assignedEvent(e)
		case "unassigned":
			ours = h.unassignedEvent(e)
		case "review_requested":
			ours = h.reviewRequestedEvent(e)
		case "review_request_removed":
			ours = h.reviewRequestRemovedEvent(e)
		case "edited":
			if e.Changes != nil && e.Changes.Title != nil {
				ours = h.renamedTitleEvent(e)
			}
		case "closed":
			ours = h.closedEvent(e)
		case "reopened":
			ours = h.reopenedEvent(e)
		}

	case *gh.PullRequestReviewEvent:
		pr = int64(*e.PullRequest.Number)
		ours = h.pullRequestReviewEvent(e)

	case *gh.PullRequestReviewCommentEvent:
		pr = int64(*e.PullRequest.Number)
		switch *e.Action {
		case "created", "edited":
			ours = h.pullRequestReviewCommentEvent(e)
		}
	}

	return
}

func (h *GitHubWebhook) upsertChangesetEvent(
	ctx context.Context,
	pr int64,
	ev interface{ Key() string },
) (err error) {
	var tx *Store
	if tx, err = h.Store.Transact(ctx); err != nil {
		return err
	}

	defer tx.Done(&err)

	cs, err := tx.GetChangeset(ctx, GetChangesetOpts{
		ExternalID:          strconv.FormatInt(pr, 10),
		ExternalServiceType: github.ServiceType,
	})
	if err != nil {
		if err == ErrNoResults {
			err = nil // Nothing to do
		}
		return err
	}

	now := h.Now()
	event := &a8n.ChangesetEvent{
		ChangesetID: cs.ID,
		Kind:        a8n.ChangesetEventKindFor(ev),
		Key:         ev.Key(),
		CreatedAt:   now,
		UpdatedAt:   now,
		Metadata:    ev,
	}

	existing, err := tx.GetChangesetEvent(ctx, GetChangesetEventOpts{
		ChangesetID: cs.ID,
		Kind:        event.Kind,
		Key:         event.Key,
	})

	if err != nil && err != ErrNoResults {
		return err
	}

	if existing != nil {
		existing.Update(event)
		event = existing
	}

	return h.Store.UpsertChangesetEvents(ctx, event)
}

func (*GitHubWebhook) issueComment(e *gh.IssueCommentEvent) *github.IssueComment {
	comment := github.IssueComment{
		DatabaseID: *e.Comment.ID,
		Author: github.Actor{
			AvatarURL: *e.Comment.User.AvatarURL,
			Login:     *e.Comment.User.Login,
			URL:       *e.Comment.User.URL,
		},
		AuthorAssociation:   *e.Comment.AuthorAssociation,
		Body:                *e.Comment.Body,
		URL:                 *e.Comment.URL,
		CreatedAt:           *e.Comment.CreatedAt,
		UpdatedAt:           *e.Comment.UpdatedAt,
		IncludesCreatedEdit: *e.Action == "edited",
	}

	if comment.IncludesCreatedEdit {
		comment.Editor = &github.Actor{
			AvatarURL: *e.Sender.AvatarURL,
			Login:     *e.Sender.Login,
			URL:       *e.Sender.URL,
		}
	}

	return &comment
}

func (*GitHubWebhook) assignedEvent(e *gh.PullRequestEvent) *github.AssignedEvent {
	return &github.AssignedEvent{
		Actor: github.Actor{
			AvatarURL: *e.Sender.AvatarURL,
			Login:     *e.Sender.Login,
			URL:       *e.Sender.URL,
		},
		Assignee: github.Actor{
			AvatarURL: *e.Assignee.AvatarURL,
			Login:     *e.Assignee.Login,
			URL:       *e.Assignee.URL,
		},
		CreatedAt: *e.PullRequest.UpdatedAt,
	}
}

func (*GitHubWebhook) unassignedEvent(e *gh.PullRequestEvent) *github.UnassignedEvent {
	return &github.UnassignedEvent{
		Actor: github.Actor{
			AvatarURL: *e.Sender.AvatarURL,
			Login:     *e.Sender.Login,
			URL:       *e.Sender.URL,
		},
		Assignee: github.Actor{
			AvatarURL: *e.Assignee.AvatarURL,
			Login:     *e.Assignee.Login,
			URL:       *e.Assignee.URL,
		},
		CreatedAt: *e.PullRequest.UpdatedAt,
	}
}

func (*GitHubWebhook) reviewRequestedEvent(e *gh.PullRequestEvent) *github.ReviewRequestedEvent {
	return &github.ReviewRequestedEvent{
		Actor: github.Actor{
			AvatarURL: *e.Sender.AvatarURL,
			Login:     *e.Sender.Login,
			URL:       *e.Sender.URL,
		},
		RequestedReviewer: github.Actor{
			AvatarURL: *e.RequestedReviewer.AvatarURL,
			Login:     *e.RequestedReviewer.Login,
			URL:       *e.RequestedReviewer.URL,
		},
		CreatedAt: *e.PullRequest.UpdatedAt,
	}
}

func (*GitHubWebhook) reviewRequestRemovedEvent(e *gh.PullRequestEvent) *github.ReviewRequestRemovedEvent {
	return &github.ReviewRequestRemovedEvent{
		Actor: github.Actor{
			AvatarURL: *e.Sender.AvatarURL,
			Login:     *e.Sender.Login,
			URL:       *e.Sender.URL,
		},
		RequestedReviewer: github.Actor{
			AvatarURL: *e.RequestedReviewer.AvatarURL,
			Login:     *e.RequestedReviewer.Login,
			URL:       *e.RequestedReviewer.URL,
		},
		CreatedAt: *e.PullRequest.UpdatedAt,
	}
}

func (*GitHubWebhook) renamedTitleEvent(e *gh.PullRequestEvent) *github.RenamedTitleEvent {
	return &github.RenamedTitleEvent{
		Actor: github.Actor{
			AvatarURL: *e.Sender.AvatarURL,
			Login:     *e.Sender.Login,
			URL:       *e.Sender.URL,
		},
		PreviousTitle: *e.Changes.Title.From,
		CurrentTitle:  *e.PullRequest.Title,
		CreatedAt:     *e.PullRequest.UpdatedAt,
	}
}

func (*GitHubWebhook) closedEvent(e *gh.PullRequestEvent) *github.ClosedEvent {
	return &github.ClosedEvent{
		Actor: github.Actor{
			AvatarURL: *e.Sender.AvatarURL,
			Login:     *e.Sender.Login,
			URL:       *e.Sender.URL,
		},
		CreatedAt: *e.PullRequest.UpdatedAt,
		// This is different from the URL returned by GraphQL because the precise
		// event URL isn't available in this webhook payload. This means if we expose
		// this URL in the UI, and users click it, they'll just go to the PR page, rather
		// than the precise location of the "close" event, until the background syncing
		// runs and updates this URL to the exact one.
		URL: *e.PullRequest.URL,
	}
}

func (*GitHubWebhook) reopenedEvent(e *gh.PullRequestEvent) *github.ReopenedEvent {
	return &github.ReopenedEvent{
		Actor: github.Actor{
			AvatarURL: *e.Sender.AvatarURL,
			Login:     *e.Sender.Login,
			URL:       *e.Sender.URL,
		},
		CreatedAt: *e.PullRequest.UpdatedAt,
	}
}

func (*GitHubWebhook) pullRequestReviewEvent(e *gh.PullRequestReviewEvent) *github.PullRequestReview {
	return &github.PullRequestReview{
		DatabaseID: *e.Review.ID,
		Author: github.Actor{
			AvatarURL: *e.Review.User.AvatarURL,
			Login:     *e.Review.User.Login,
			URL:       *e.Review.User.URL,
		},
		Body:      *e.Review.Body,
		State:     *e.Review.State,
		URL:       *e.Review.HTMLURL,
		CreatedAt: *e.Review.SubmittedAt,
		UpdatedAt: *e.Review.SubmittedAt,
		Commit: github.Commit{
			OID: *e.Review.CommitID,
		},
	}
}

func (*GitHubWebhook) pullRequestReviewCommentEvent(e *gh.PullRequestReviewCommentEvent) *github.PullRequestReviewComment {
	comment := github.PullRequestReviewComment{
		DatabaseID:        *e.Comment.ID,
		AuthorAssociation: *e.Comment.AuthorAssociation,
		Commit: github.Commit{
			OID: *e.Comment.CommitID,
		},
		Body:                *e.Comment.Body,
		URL:                 *e.Comment.URL,
		CreatedAt:           *e.Comment.CreatedAt,
		UpdatedAt:           *e.Comment.UpdatedAt,
		IncludesCreatedEdit: *e.Action == "edited",
	}

	user := github.Actor{
		AvatarURL: *e.Comment.User.AvatarURL,
		Login:     *e.Comment.User.Login,
		URL:       *e.Comment.User.URL,
	}

	if comment.IncludesCreatedEdit {
		comment.Editor = user
	} else {
		comment.Author = user
	}

	return &comment
}

type httpError struct {
	code int
	err  error
}

func (e httpError) Error() string {
	if e.err != nil {
		return fmt.Sprintf("HTTP %d: %v", e.code, e.err)
	}
	return fmt.Sprintf("HTTP %d: %s", e.code, http.StatusText(e.code))
}

func respond(w http.ResponseWriter, code int, v interface{}) {
	switch val := v.(type) {
	case nil:
		w.WriteHeader(code)
	case error:
		if val != nil {
			log15.Error(val.Error())
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(code)
			fmt.Fprintf(w, "%v", val)
		}
	default:
		w.Header().Set("Content-Type", "application/json")
		bs, err := json.Marshal(v)
		if err != nil {
			respond(w, http.StatusInternalServerError, err)
			return
		}

		w.WriteHeader(code)
		if _, err = w.Write(bs); err != nil {
			log15.Error("failed to write response", "error", err)
		}
	}
}
