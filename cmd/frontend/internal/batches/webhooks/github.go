package webhooks

import (
	"context"
	"fmt"
	"strconv"

	gh "github.com/google/go-github/v55/github"
	sglog "github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/webhooks"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/batches/store"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var (
	// githubEvents is the set of events this webhook handler listens to
	// you can find info about what these events contain here:
	// https://docs.github.com/en/free-pro-team@latest/developers/webhooks-and-events/webhook-events-and-payloads
	githubEvents = []string{
		"issue_comment",
		"pull_request",
		"pull_request_review",
		"pull_request_review_comment",
		"status",
		"check_suite",
		"check_run",
	}
)

// GitHubWebhook receives GitHub organization webhook events that are
// relevant to Batch Changes, normalizes those events into ChangesetEvents
// and upserts them to the database.
type GitHubWebhook struct {
	*webhook
}

func NewGitHubWebhook(store *store.Store, gitserverClient gitserver.Client, logger sglog.Logger) *GitHubWebhook {
	return &GitHubWebhook{&webhook{store, gitserverClient, logger, extsvc.TypeGitHub}}
}

// Register registers this webhook handler to handle events with the passed webhook router
func (h *GitHubWebhook) Register(router *webhooks.Router) {
	router.Register(
		h.handleGitHubWebhook,
		extsvc.KindGitHub,
		githubEvents...,
	)
}

// handleGithubWebhook is the entry point for webhooks from the webhook router, see the events
// it's registered to handle in GitHubWebhook.Register
func (h *GitHubWebhook) handleGitHubWebhook(ctx context.Context, db database.DB, codeHostURN extsvc.CodeHostBaseURL, payload any) error {
	var m error

	prs, ev := h.convertEvent(ctx, codeHostURN, payload)

	if ev == nil {
		// We don't recognize this event type, we don't need to do any more work.
		// This happens when the action in the body is of no type we understand.
		// Since we don't care about those events, we can just early return here.
		return nil
	}

	for _, pr := range prs {
		if pr == (PR{}) {
			continue
		}

		err := h.upsertChangesetEvent(ctx, codeHostURN, pr, ev)
		if err != nil {
			m = errors.Append(m, err)
		}
	}
	return m
}

func (h *GitHubWebhook) convertEvent(ctx context.Context, codeHostURN extsvc.CodeHostBaseURL, theirs any) (prs []PR, ours keyer) {
	h.logger.Debug("GitHub webhook received", sglog.String("type", fmt.Sprintf("%T", theirs)))
	switch e := theirs.(type) {
	case *gh.IssueCommentEvent:
		repo := e.GetRepo()
		if repo == nil {
			return
		}
		repoExternalID := repo.GetNodeID()

		pr := PR{ID: int64(*e.Issue.Number), RepoExternalID: repoExternalID}
		prs = append(prs, pr)
		return prs, h.issueComment(e)

	case *gh.PullRequestEvent:
		repo := e.GetRepo()
		if repo == nil {
			return
		}
		repoExternalID := repo.GetNodeID()

		if e.Number == nil {
			return
		}
		pr := PR{ID: int64(*e.Number), RepoExternalID: repoExternalID}
		prs = append(prs, pr)

		if e.Action == nil {
			return
		}
		switch *e.Action {
		case "ready_for_review":
			ours = h.readyForReviewEvent(e)
		case "converted_to_draft":
			ours = h.convertToDraftEvent(e)
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
			ours = h.closedOrMergeEvent(e)
		case "reopened":
			ours = h.reopenedEvent(e)
		case "labeled", "unlabeled":
			ours = h.labeledEvent(e)
		}

	case *gh.PullRequestReviewEvent:
		repo := e.GetRepo()
		if repo == nil {
			return
		}
		repoExternalID := repo.GetNodeID()

		pr := PR{ID: int64(*e.PullRequest.Number), RepoExternalID: repoExternalID}
		prs = append(prs, pr)
		ours = h.pullRequestReviewEvent(e)

	case *gh.PullRequestReviewCommentEvent:
		repo := e.GetRepo()
		if repo == nil {
			return
		}
		repoExternalID := repo.GetNodeID()

		pr := PR{ID: int64(*e.PullRequest.Number), RepoExternalID: repoExternalID}
		prs = append(prs, pr)
		switch *e.Action {
		case "created", "edited":
			ours = h.pullRequestReviewCommentEvent(e)
		}

	case *gh.StatusEvent:
		// A status event could potentially relate to more than one
		// PR so we need to find them all
		refs := make([]string, 0, len(e.Branches))
		for _, branch := range e.Branches {
			if name := branch.GetName(); name != "" {
				refs = append(refs, name)
			}
		}

		if len(refs) == 0 {
			return nil, nil
		}

		repo := e.GetRepo()
		if repo == nil {
			return
		}
		repoExternalID := repo.GetNodeID()

		spec := api.ExternalRepoSpec{
			ID:          repoExternalID,
			ServiceID:   codeHostURN.String(),
			ServiceType: extsvc.TypeGitHub,
		}

		ids, err := h.Store.GetChangesetExternalIDs(ctx, spec, refs)
		if err != nil {
			h.logger.Error("Error executing GetChangesetExternalIDs", sglog.Error(err))
			return nil, nil
		}

		for _, id := range ids {
			i, err := strconv.ParseInt(id, 10, 64)
			if err != nil {
				h.logger.Error("Error parsing external id", sglog.Error(err))
				continue
			}
			prs = append(prs, PR{ID: i, RepoExternalID: repoExternalID})
		}

		ours = h.commitStatusEvent(e)

	case *gh.CheckSuiteEvent:
		if e.CheckSuite == nil {
			return
		}

		cs := e.GetCheckSuite()

		// https://docs.github.com/en/webhooks-and-events/webhooks/webhook-events-and-payloads#check_suite
		// The `repository` field was recently removed from the payload for `check_suite` event, this was causing
		// webhook events not to update changesets on time in Sourcegraph.
		repo := e.GetRepo()
		if repo == nil {
			return
		}
		repoID := repo.GetNodeID()

		for _, pr := range cs.PullRequests {
			n := pr.GetNumber()
			if n != 0 {
				prs = append(prs, PR{ID: int64(n), RepoExternalID: repoID})
			}
		}
		ours = h.checkSuiteEvent(cs)

	case *gh.CheckRunEvent:
		if e.CheckRun == nil {
			return
		}

		cr := e.GetCheckRun()

		// https://docs.github.com/en/webhooks-and-events/webhooks/webhook-events-and-payloads#check_run
		// The `repository` field was recently removed from the payload for `check_run` event, this was causing
		// webhook events not to update changesets on time in Sourcegraph.
		repo := e.GetRepo()
		if repo == nil {
			return
		}
		repoID := repo.GetNodeID()

		for _, pr := range cr.PullRequests {
			n := pr.GetNumber()
			if n != 0 {
				prs = append(prs, PR{ID: int64(n), RepoExternalID: repoID})
			}
		}
		ours = h.checkRunEvent(cr)
	}

	return prs, ours
}

func (*GitHubWebhook) issueComment(e *gh.IssueCommentEvent) *github.IssueComment {
	comment := github.IssueComment{}

	if c := e.GetComment(); c != nil {
		comment.DatabaseID = c.GetID()

		if u := c.GetUser(); u != nil {
			comment.Author.AvatarURL = u.GetAvatarURL()
			comment.Author.Login = u.GetLogin()
			comment.Author.URL = u.GetURL()
		}

		comment.AuthorAssociation = c.GetAuthorAssociation()
		comment.Body = c.GetBody()
		comment.URL = c.GetURL()
		comment.CreatedAt = c.GetCreatedAt().Time
		comment.UpdatedAt = c.GetUpdatedAt().Time
	}

	comment.IncludesCreatedEdit = e.GetAction() == "edited"
	if s := e.GetSender(); s != nil && comment.IncludesCreatedEdit {
		comment.Editor = &github.Actor{
			AvatarURL: s.GetAvatarURL(),
			Login:     s.GetLogin(),
			URL:       s.GetURL(),
		}
	}

	return &comment
}

func (*GitHubWebhook) labeledEvent(e *gh.PullRequestEvent) *github.LabelEvent {
	labelEvent := &github.LabelEvent{
		Removed: e.GetAction() == "unlabeled",
	}

	if pr := e.GetPullRequest(); pr != nil {
		labelEvent.CreatedAt = pr.GetUpdatedAt().Time
	}

	if l := e.GetLabel(); l != nil {
		labelEvent.Label.Color = l.GetColor()
		labelEvent.Label.Description = l.GetDescription()
		labelEvent.Label.Name = l.GetName()
		labelEvent.Label.ID = l.GetNodeID()
	}

	if s := e.GetSender(); s != nil {
		labelEvent.Actor.AvatarURL = s.GetAvatarURL()
		labelEvent.Actor.Login = s.GetLogin()
		labelEvent.Actor.URL = s.GetURL()
	}

	return labelEvent
}

func (*GitHubWebhook) readyForReviewEvent(e *gh.PullRequestEvent) *github.ReadyForReviewEvent {
	readyForReviewEvent := &github.ReadyForReviewEvent{}

	if pr := e.GetPullRequest(); pr != nil {
		readyForReviewEvent.CreatedAt = pr.GetUpdatedAt().Time
	}

	if s := e.GetSender(); s != nil {
		readyForReviewEvent.Actor.AvatarURL = s.GetAvatarURL()
		readyForReviewEvent.Actor.Login = s.GetLogin()
		readyForReviewEvent.Actor.URL = s.GetURL()
	}

	return readyForReviewEvent
}

func (*GitHubWebhook) convertToDraftEvent(e *gh.PullRequestEvent) *github.ConvertToDraftEvent {
	convertToDraftEvent := &github.ConvertToDraftEvent{}

	if pr := e.GetPullRequest(); pr != nil {
		convertToDraftEvent.CreatedAt = pr.GetUpdatedAt().Time
	}

	if s := e.GetSender(); s != nil {
		convertToDraftEvent.Actor.AvatarURL = s.GetAvatarURL()
		convertToDraftEvent.Actor.Login = s.GetLogin()
		convertToDraftEvent.Actor.URL = s.GetURL()
	}

	return convertToDraftEvent
}

func (*GitHubWebhook) assignedEvent(e *gh.PullRequestEvent) *github.AssignedEvent {
	assignedEvent := &github.AssignedEvent{}

	if pr := e.GetPullRequest(); pr != nil {
		assignedEvent.CreatedAt = pr.GetUpdatedAt().Time
	}

	if s := e.GetSender(); s != nil {
		assignedEvent.Actor.AvatarURL = s.GetAvatarURL()
		assignedEvent.Actor.Login = s.GetLogin()
		assignedEvent.Actor.URL = s.GetURL()
	}

	if a := e.GetAssignee(); a != nil {
		assignedEvent.Assignee.AvatarURL = a.GetAvatarURL()
		assignedEvent.Assignee.Login = a.GetLogin()
		assignedEvent.Assignee.URL = a.GetURL()
	}

	return assignedEvent
}

func (*GitHubWebhook) unassignedEvent(e *gh.PullRequestEvent) *github.UnassignedEvent {
	unassignedEvent := &github.UnassignedEvent{}

	if pr := e.GetPullRequest(); pr != nil {
		unassignedEvent.CreatedAt = pr.GetUpdatedAt().Time
	}

	if s := e.GetSender(); s != nil {
		unassignedEvent.Actor.AvatarURL = s.GetAvatarURL()
		unassignedEvent.Actor.Login = s.GetLogin()
		unassignedEvent.Actor.URL = s.GetURL()
	}

	if a := e.GetAssignee(); a != nil {
		unassignedEvent.Assignee.AvatarURL = a.GetAvatarURL()
		unassignedEvent.Assignee.Login = a.GetLogin()
		unassignedEvent.Assignee.URL = a.GetURL()
	}

	return unassignedEvent
}

func (*GitHubWebhook) reviewRequestedEvent(e *gh.PullRequestEvent) *github.ReviewRequestedEvent {
	event := &github.ReviewRequestedEvent{}

	if s := e.GetSender(); s != nil {
		event.Actor.AvatarURL = s.GetAvatarURL()
		event.Actor.Login = s.GetLogin()
		event.Actor.URL = s.GetURL()
	}

	if pr := e.GetPullRequest(); pr != nil {
		event.CreatedAt = pr.GetUpdatedAt().Time
	}

	if e.RequestedReviewer != nil {
		event.RequestedReviewer = github.Actor{
			AvatarURL: e.RequestedReviewer.GetAvatarURL(),
			Login:     e.RequestedReviewer.GetLogin(),
			URL:       e.RequestedReviewer.GetURL(),
		}
	}

	if e.RequestedTeam != nil {
		event.RequestedTeam = github.Team{
			Name: e.RequestedTeam.GetName(),
			URL:  e.RequestedTeam.GetURL(),
		}
	}

	return event
}

func (*GitHubWebhook) reviewRequestRemovedEvent(e *gh.PullRequestEvent) *github.ReviewRequestRemovedEvent {
	event := &github.ReviewRequestRemovedEvent{}

	if s := e.GetSender(); s != nil {
		event.Actor.AvatarURL = s.GetAvatarURL()
		event.Actor.Login = s.GetLogin()
		event.Actor.URL = s.GetURL()
	}

	if pr := e.GetPullRequest(); pr != nil {
		event.CreatedAt = pr.GetUpdatedAt().Time
	}

	if e.RequestedReviewer != nil {
		event.RequestedReviewer = github.Actor{
			AvatarURL: e.RequestedReviewer.GetAvatarURL(),
			Login:     e.RequestedReviewer.GetLogin(),
			URL:       e.RequestedReviewer.GetURL(),
		}
	}

	if e.RequestedTeam != nil {
		event.RequestedTeam = github.Team{
			Name: e.RequestedTeam.GetName(),
			URL:  e.RequestedTeam.GetURL(),
		}
	}

	return event
}

func (*GitHubWebhook) renamedTitleEvent(e *gh.PullRequestEvent) *github.RenamedTitleEvent {
	event := &github.RenamedTitleEvent{}

	if s := e.GetSender(); s != nil {
		event.Actor.AvatarURL = s.GetAvatarURL()
		event.Actor.Login = s.GetLogin()
		event.Actor.URL = s.GetURL()
	}

	if pr := e.GetPullRequest(); pr != nil {
		event.CurrentTitle = pr.GetTitle()
		event.CreatedAt = pr.GetUpdatedAt().Time
	}

	if ch := e.GetChanges(); ch != nil && ch.Title != nil && ch.Title.From != nil {
		event.PreviousTitle = *ch.Title.From
	}

	return event
}

// closed events from github have a 'merged flag which identifies them as
// merge events instead.
func (*GitHubWebhook) closedOrMergeEvent(e *gh.PullRequestEvent) keyer {
	closeEvent := &github.ClosedEvent{}

	if s := e.GetSender(); s != nil {
		closeEvent.Actor.AvatarURL = s.GetAvatarURL()
		closeEvent.Actor.Login = s.GetLogin()
		closeEvent.Actor.URL = s.GetURL()
	}

	if pr := e.GetPullRequest(); pr != nil {
		closeEvent.CreatedAt = pr.GetUpdatedAt().Time

		// This is different from the URL returned by GraphQL because the precise
		// event URL isn't available in this webhook payload. This means if we expose
		// this URL in the UI, and users click it, they'll just go to the PR page, rather
		// than the precise location of the "close" event, until the background syncing
		// runs and updates this URL to the exact one.
		closeEvent.URL = pr.GetURL()

		// We actually have a merged event
		if pr.GetMerged() {
			mergedEvent := &github.MergedEvent{
				Actor:     closeEvent.Actor,
				URL:       closeEvent.URL,
				CreatedAt: closeEvent.CreatedAt,
			}
			if base := pr.GetBase(); base != nil {
				mergedEvent.MergeRefName = base.GetRef()
			}
			return mergedEvent
		}
	}

	return closeEvent
}

func (*GitHubWebhook) reopenedEvent(e *gh.PullRequestEvent) *github.ReopenedEvent {
	event := &github.ReopenedEvent{}

	if s := e.GetSender(); s != nil {
		event.Actor.AvatarURL = s.GetAvatarURL()
		event.Actor.Login = s.GetLogin()
		event.Actor.URL = s.GetURL()
	}

	if pr := e.GetPullRequest(); pr != nil {
		event.CreatedAt = pr.GetUpdatedAt().Time
	}

	return event
}

func (*GitHubWebhook) pullRequestReviewEvent(e *gh.PullRequestReviewEvent) *github.PullRequestReview {
	review := &github.PullRequestReview{}

	if r := e.GetReview(); r != nil {
		review.DatabaseID = r.GetID()
		review.Body = e.Review.GetBody()
		review.State = e.Review.GetState()
		review.URL = e.Review.GetHTMLURL()
		review.CreatedAt = e.Review.GetSubmittedAt().Time
		review.UpdatedAt = e.Review.GetSubmittedAt().Time

		if u := r.GetUser(); u != nil {
			review.Author.AvatarURL = u.GetAvatarURL()
			review.Author.Login = u.GetLogin()
			review.Author.URL = u.GetURL()
		}

		review.Commit.OID = r.GetCommitID()
	}

	return review
}

func (*GitHubWebhook) pullRequestReviewCommentEvent(e *gh.PullRequestReviewCommentEvent) *github.PullRequestReviewComment {
	comment := github.PullRequestReviewComment{}

	user := github.Actor{}

	if c := e.GetComment(); c != nil {
		comment.DatabaseID = c.GetID()
		comment.AuthorAssociation = c.GetAuthorAssociation()
		comment.Commit = github.Commit{
			OID: c.GetCommitID(),
		}
		comment.Body = c.GetBody()
		comment.URL = c.GetURL()
		comment.CreatedAt = c.GetCreatedAt().Time
		comment.UpdatedAt = c.GetUpdatedAt().Time

		if u := c.GetUser(); u != nil {
			user.AvatarURL = u.GetAvatarURL()
			user.Login = u.GetLogin()
			user.URL = u.GetURL()
		}
	}

	comment.IncludesCreatedEdit = e.GetAction() == "edited"

	if comment.IncludesCreatedEdit {
		comment.Editor = user
	} else {
		comment.Author = user
	}

	return &comment
}

func (h *GitHubWebhook) commitStatusEvent(e *gh.StatusEvent) *github.CommitStatus {
	return &github.CommitStatus{
		SHA:        e.GetSHA(),
		State:      e.GetState(),
		Context:    e.GetContext(),
		ReceivedAt: h.Store.Clock()(),
	}
}

func (h *GitHubWebhook) checkSuiteEvent(cs *gh.CheckSuite) *github.CheckSuite {
	return &github.CheckSuite{
		ID:         cs.GetNodeID(),
		Status:     cs.GetStatus(),
		Conclusion: cs.GetConclusion(),
		ReceivedAt: h.Store.Clock()(),
	}
}

func (h *GitHubWebhook) checkRunEvent(cr *gh.CheckRun) *github.CheckRun {
	return &github.CheckRun{
		ID:         cr.GetNodeID(),
		Status:     cr.GetStatus(),
		Conclusion: cr.GetConclusion(),
		ReceivedAt: h.Store.Clock()(),
	}
}
