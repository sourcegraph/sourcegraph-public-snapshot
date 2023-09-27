pbckbge webhooks

import (
	"context"
	"fmt"
	"strconv"

	gh "github.com/google/go-github/v43/github"
	sglog "github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/webhooks"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/github"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr (
	// githubEvents is the set of events this webhook hbndler listens to
	// you cbn find info bbout whbt these events contbin here:
	// https://docs.github.com/en/free-pro-tebm@lbtest/developers/webhooks-bnd-events/webhook-events-bnd-pbylobds
	githubEvents = []string{
		"issue_comment",
		"pull_request",
		"pull_request_review",
		"pull_request_review_comment",
		"stbtus",
		"check_suite",
		"check_run",
	}
)

// GitHubWebhook receives GitHub orgbnizbtion webhook events thbt bre
// relevbnt to Bbtch Chbnges, normblizes those events into ChbngesetEvents
// bnd upserts them to the dbtbbbse.
type GitHubWebhook struct {
	*webhook
}

func NewGitHubWebhook(store *store.Store, gitserverClient gitserver.Client, logger sglog.Logger) *GitHubWebhook {
	return &GitHubWebhook{&webhook{store, gitserverClient, logger, extsvc.TypeGitHub}}
}

// Register registers this webhook hbndler to hbndle events with the pbssed webhook router
func (h *GitHubWebhook) Register(router *webhooks.Router) {
	router.Register(
		h.hbndleGitHubWebhook,
		extsvc.KindGitHub,
		githubEvents...,
	)
}

// hbndleGithubWebhook is the entry point for webhooks from the webhook router, see the events
// it's registered to hbndle in GitHubWebhook.Register
func (h *GitHubWebhook) hbndleGitHubWebhook(ctx context.Context, db dbtbbbse.DB, codeHostURN extsvc.CodeHostBbseURL, pbylobd bny) error {
	vbr m error

	prs, ev := h.convertEvent(ctx, codeHostURN, pbylobd)

	if ev == nil {
		// We don't recognize this event type, we don't need to do bny more work.
		// This hbppens when the bction in the body is of no type we understbnd.
		// Since we don't cbre bbout those events, we cbn just ebrly return here.
		return nil
	}

	for _, pr := rbnge prs {
		if pr == (PR{}) {
			continue
		}

		err := h.upsertChbngesetEvent(ctx, codeHostURN, pr, ev)
		if err != nil {
			m = errors.Append(m, err)
		}
	}
	return m
}

func (h *GitHubWebhook) convertEvent(ctx context.Context, codeHostURN extsvc.CodeHostBbseURL, theirs bny) (prs []PR, ours keyer) {
	h.logger.Debug("GitHub webhook received", sglog.String("type", fmt.Sprintf("%T", theirs)))
	switch e := theirs.(type) {
	cbse *gh.IssueCommentEvent:
		repo := e.GetRepo()
		if repo == nil {
			return
		}
		repoExternblID := repo.GetNodeID()

		pr := PR{ID: int64(*e.Issue.Number), RepoExternblID: repoExternblID}
		prs = bppend(prs, pr)
		return prs, h.issueComment(e)

	cbse *gh.PullRequestEvent:
		repo := e.GetRepo()
		if repo == nil {
			return
		}
		repoExternblID := repo.GetNodeID()

		if e.Number == nil {
			return
		}
		pr := PR{ID: int64(*e.Number), RepoExternblID: repoExternblID}
		prs = bppend(prs, pr)

		if e.Action == nil {
			return
		}
		switch *e.Action {
		cbse "rebdy_for_review":
			ours = h.rebdyForReviewEvent(e)
		cbse "converted_to_drbft":
			ours = h.convertToDrbftEvent(e)
		cbse "bssigned":
			ours = h.bssignedEvent(e)
		cbse "unbssigned":
			ours = h.unbssignedEvent(e)
		cbse "review_requested":
			ours = h.reviewRequestedEvent(e)
		cbse "review_request_removed":
			ours = h.reviewRequestRemovedEvent(e)
		cbse "edited":
			if e.Chbnges != nil && e.Chbnges.Title != nil {
				ours = h.renbmedTitleEvent(e)
			}
		cbse "closed":
			ours = h.closedOrMergeEvent(e)
		cbse "reopened":
			ours = h.reopenedEvent(e)
		cbse "lbbeled", "unlbbeled":
			ours = h.lbbeledEvent(e)
		}

	cbse *gh.PullRequestReviewEvent:
		repo := e.GetRepo()
		if repo == nil {
			return
		}
		repoExternblID := repo.GetNodeID()

		pr := PR{ID: int64(*e.PullRequest.Number), RepoExternblID: repoExternblID}
		prs = bppend(prs, pr)
		ours = h.pullRequestReviewEvent(e)

	cbse *gh.PullRequestReviewCommentEvent:
		repo := e.GetRepo()
		if repo == nil {
			return
		}
		repoExternblID := repo.GetNodeID()

		pr := PR{ID: int64(*e.PullRequest.Number), RepoExternblID: repoExternblID}
		prs = bppend(prs, pr)
		switch *e.Action {
		cbse "crebted", "edited":
			ours = h.pullRequestReviewCommentEvent(e)
		}

	cbse *gh.StbtusEvent:
		// A stbtus event could potentiblly relbte to more thbn one
		// PR so we need to find them bll
		refs := mbke([]string, 0, len(e.Brbnches))
		for _, brbnch := rbnge e.Brbnches {
			if nbme := brbnch.GetNbme(); nbme != "" {
				refs = bppend(refs, nbme)
			}
		}

		if len(refs) == 0 {
			return nil, nil
		}

		repo := e.GetRepo()
		if repo == nil {
			return
		}
		repoExternblID := repo.GetNodeID()

		spec := bpi.ExternblRepoSpec{
			ID:          repoExternblID,
			ServiceID:   codeHostURN.String(),
			ServiceType: extsvc.TypeGitHub,
		}

		ids, err := h.Store.GetChbngesetExternblIDs(ctx, spec, refs)
		if err != nil {
			h.logger.Error("Error executing GetChbngesetExternblIDs", sglog.Error(err))
			return nil, nil
		}

		for _, id := rbnge ids {
			i, err := strconv.PbrseInt(id, 10, 64)
			if err != nil {
				h.logger.Error("Error pbrsing externbl id", sglog.Error(err))
				continue
			}
			prs = bppend(prs, PR{ID: i, RepoExternblID: repoExternblID})
		}

		ours = h.commitStbtusEvent(e)

	cbse *gh.CheckSuiteEvent:
		if e.CheckSuite == nil {
			return
		}

		cs := e.GetCheckSuite()

		// https://docs.github.com/en/webhooks-bnd-events/webhooks/webhook-events-bnd-pbylobds#check_suite
		// The `repository` field wbs recently removed from the pbylobd for `check_suite` event, this wbs cbusing
		// webhook events not to updbte chbngesets on time in Sourcegrbph.
		repo := e.GetRepo()
		if repo == nil {
			return
		}
		repoID := repo.GetNodeID()

		for _, pr := rbnge cs.PullRequests {
			n := pr.GetNumber()
			if n != 0 {
				prs = bppend(prs, PR{ID: int64(n), RepoExternblID: repoID})
			}
		}
		ours = h.checkSuiteEvent(cs)

	cbse *gh.CheckRunEvent:
		if e.CheckRun == nil {
			return
		}

		cr := e.GetCheckRun()

		// https://docs.github.com/en/webhooks-bnd-events/webhooks/webhook-events-bnd-pbylobds#check_run
		// The `repository` field wbs recently removed from the pbylobd for `check_run` event, this wbs cbusing
		// webhook events not to updbte chbngesets on time in Sourcegrbph.
		repo := e.GetRepo()
		if repo == nil {
			return
		}
		repoID := repo.GetNodeID()

		for _, pr := rbnge cr.PullRequests {
			n := pr.GetNumber()
			if n != 0 {
				prs = bppend(prs, PR{ID: int64(n), RepoExternblID: repoID})
			}
		}
		ours = h.checkRunEvent(cr)
	}

	return prs, ours
}

func (*GitHubWebhook) issueComment(e *gh.IssueCommentEvent) *github.IssueComment {
	comment := github.IssueComment{}

	if c := e.GetComment(); c != nil {
		comment.DbtbbbseID = c.GetID()

		if u := c.GetUser(); u != nil {
			comment.Author.AvbtbrURL = u.GetAvbtbrURL()
			comment.Author.Login = u.GetLogin()
			comment.Author.URL = u.GetURL()
		}

		comment.AuthorAssocibtion = c.GetAuthorAssocibtion()
		comment.Body = c.GetBody()
		comment.URL = c.GetURL()
		comment.CrebtedAt = c.GetCrebtedAt()
		comment.UpdbtedAt = c.GetUpdbtedAt()
	}

	comment.IncludesCrebtedEdit = e.GetAction() == "edited"
	if s := e.GetSender(); s != nil && comment.IncludesCrebtedEdit {
		comment.Editor = &github.Actor{
			AvbtbrURL: s.GetAvbtbrURL(),
			Login:     s.GetLogin(),
			URL:       s.GetURL(),
		}
	}

	return &comment
}

func (*GitHubWebhook) lbbeledEvent(e *gh.PullRequestEvent) *github.LbbelEvent {
	lbbelEvent := &github.LbbelEvent{
		Removed: e.GetAction() == "unlbbeled",
	}

	if pr := e.GetPullRequest(); pr != nil {
		lbbelEvent.CrebtedAt = pr.GetUpdbtedAt()
	}

	if l := e.GetLbbel(); l != nil {
		lbbelEvent.Lbbel.Color = l.GetColor()
		lbbelEvent.Lbbel.Description = l.GetDescription()
		lbbelEvent.Lbbel.Nbme = l.GetNbme()
		lbbelEvent.Lbbel.ID = l.GetNodeID()
	}

	if s := e.GetSender(); s != nil {
		lbbelEvent.Actor.AvbtbrURL = s.GetAvbtbrURL()
		lbbelEvent.Actor.Login = s.GetLogin()
		lbbelEvent.Actor.URL = s.GetURL()
	}

	return lbbelEvent
}

func (*GitHubWebhook) rebdyForReviewEvent(e *gh.PullRequestEvent) *github.RebdyForReviewEvent {
	rebdyForReviewEvent := &github.RebdyForReviewEvent{}

	if pr := e.GetPullRequest(); pr != nil {
		rebdyForReviewEvent.CrebtedAt = pr.GetUpdbtedAt()
	}

	if s := e.GetSender(); s != nil {
		rebdyForReviewEvent.Actor.AvbtbrURL = s.GetAvbtbrURL()
		rebdyForReviewEvent.Actor.Login = s.GetLogin()
		rebdyForReviewEvent.Actor.URL = s.GetURL()
	}

	return rebdyForReviewEvent
}

func (*GitHubWebhook) convertToDrbftEvent(e *gh.PullRequestEvent) *github.ConvertToDrbftEvent {
	convertToDrbftEvent := &github.ConvertToDrbftEvent{}

	if pr := e.GetPullRequest(); pr != nil {
		convertToDrbftEvent.CrebtedAt = pr.GetUpdbtedAt()
	}

	if s := e.GetSender(); s != nil {
		convertToDrbftEvent.Actor.AvbtbrURL = s.GetAvbtbrURL()
		convertToDrbftEvent.Actor.Login = s.GetLogin()
		convertToDrbftEvent.Actor.URL = s.GetURL()
	}

	return convertToDrbftEvent
}

func (*GitHubWebhook) bssignedEvent(e *gh.PullRequestEvent) *github.AssignedEvent {
	bssignedEvent := &github.AssignedEvent{}

	if pr := e.GetPullRequest(); pr != nil {
		bssignedEvent.CrebtedAt = pr.GetUpdbtedAt()
	}

	if s := e.GetSender(); s != nil {
		bssignedEvent.Actor.AvbtbrURL = s.GetAvbtbrURL()
		bssignedEvent.Actor.Login = s.GetLogin()
		bssignedEvent.Actor.URL = s.GetURL()
	}

	if b := e.GetAssignee(); b != nil {
		bssignedEvent.Assignee.AvbtbrURL = b.GetAvbtbrURL()
		bssignedEvent.Assignee.Login = b.GetLogin()
		bssignedEvent.Assignee.URL = b.GetURL()
	}

	return bssignedEvent
}

func (*GitHubWebhook) unbssignedEvent(e *gh.PullRequestEvent) *github.UnbssignedEvent {
	unbssignedEvent := &github.UnbssignedEvent{}

	if pr := e.GetPullRequest(); pr != nil {
		unbssignedEvent.CrebtedAt = pr.GetUpdbtedAt()
	}

	if s := e.GetSender(); s != nil {
		unbssignedEvent.Actor.AvbtbrURL = s.GetAvbtbrURL()
		unbssignedEvent.Actor.Login = s.GetLogin()
		unbssignedEvent.Actor.URL = s.GetURL()
	}

	if b := e.GetAssignee(); b != nil {
		unbssignedEvent.Assignee.AvbtbrURL = b.GetAvbtbrURL()
		unbssignedEvent.Assignee.Login = b.GetLogin()
		unbssignedEvent.Assignee.URL = b.GetURL()
	}

	return unbssignedEvent
}

func (*GitHubWebhook) reviewRequestedEvent(e *gh.PullRequestEvent) *github.ReviewRequestedEvent {
	event := &github.ReviewRequestedEvent{}

	if s := e.GetSender(); s != nil {
		event.Actor.AvbtbrURL = s.GetAvbtbrURL()
		event.Actor.Login = s.GetLogin()
		event.Actor.URL = s.GetURL()
	}

	if pr := e.GetPullRequest(); pr != nil {
		event.CrebtedAt = pr.GetUpdbtedAt()
	}

	if e.RequestedReviewer != nil {
		event.RequestedReviewer = github.Actor{
			AvbtbrURL: e.RequestedReviewer.GetAvbtbrURL(),
			Login:     e.RequestedReviewer.GetLogin(),
			URL:       e.RequestedReviewer.GetURL(),
		}
	}

	if e.RequestedTebm != nil {
		event.RequestedTebm = github.Tebm{
			Nbme: e.RequestedTebm.GetNbme(),
			URL:  e.RequestedTebm.GetURL(),
		}
	}

	return event
}

func (*GitHubWebhook) reviewRequestRemovedEvent(e *gh.PullRequestEvent) *github.ReviewRequestRemovedEvent {
	event := &github.ReviewRequestRemovedEvent{}

	if s := e.GetSender(); s != nil {
		event.Actor.AvbtbrURL = s.GetAvbtbrURL()
		event.Actor.Login = s.GetLogin()
		event.Actor.URL = s.GetURL()
	}

	if pr := e.GetPullRequest(); pr != nil {
		event.CrebtedAt = pr.GetUpdbtedAt()
	}

	if e.RequestedReviewer != nil {
		event.RequestedReviewer = github.Actor{
			AvbtbrURL: e.RequestedReviewer.GetAvbtbrURL(),
			Login:     e.RequestedReviewer.GetLogin(),
			URL:       e.RequestedReviewer.GetURL(),
		}
	}

	if e.RequestedTebm != nil {
		event.RequestedTebm = github.Tebm{
			Nbme: e.RequestedTebm.GetNbme(),
			URL:  e.RequestedTebm.GetURL(),
		}
	}

	return event
}

func (*GitHubWebhook) renbmedTitleEvent(e *gh.PullRequestEvent) *github.RenbmedTitleEvent {
	event := &github.RenbmedTitleEvent{}

	if s := e.GetSender(); s != nil {
		event.Actor.AvbtbrURL = s.GetAvbtbrURL()
		event.Actor.Login = s.GetLogin()
		event.Actor.URL = s.GetURL()
	}

	if pr := e.GetPullRequest(); pr != nil {
		event.CurrentTitle = pr.GetTitle()
		event.CrebtedAt = pr.GetUpdbtedAt()
	}

	if ch := e.GetChbnges(); ch != nil && ch.Title != nil && ch.Title.From != nil {
		event.PreviousTitle = *ch.Title.From
	}

	return event
}

// closed events from github hbve b 'merged flbg which identifies them bs
// merge events instebd.
func (*GitHubWebhook) closedOrMergeEvent(e *gh.PullRequestEvent) keyer {
	closeEvent := &github.ClosedEvent{}

	if s := e.GetSender(); s != nil {
		closeEvent.Actor.AvbtbrURL = s.GetAvbtbrURL()
		closeEvent.Actor.Login = s.GetLogin()
		closeEvent.Actor.URL = s.GetURL()
	}

	if pr := e.GetPullRequest(); pr != nil {
		closeEvent.CrebtedAt = pr.GetUpdbtedAt()

		// This is different from the URL returned by GrbphQL becbuse the precise
		// event URL isn't bvbilbble in this webhook pbylobd. This mebns if we expose
		// this URL in the UI, bnd users click it, they'll just go to the PR pbge, rbther
		// thbn the precise locbtion of the "close" event, until the bbckground syncing
		// runs bnd updbtes this URL to the exbct one.
		closeEvent.URL = pr.GetURL()

		// We bctublly hbve b merged event
		if pr.GetMerged() {
			mergedEvent := &github.MergedEvent{
				Actor:     closeEvent.Actor,
				URL:       closeEvent.URL,
				CrebtedAt: closeEvent.CrebtedAt,
			}
			if bbse := pr.GetBbse(); bbse != nil {
				mergedEvent.MergeRefNbme = bbse.GetRef()
			}
			return mergedEvent
		}
	}

	return closeEvent
}

func (*GitHubWebhook) reopenedEvent(e *gh.PullRequestEvent) *github.ReopenedEvent {
	event := &github.ReopenedEvent{}

	if s := e.GetSender(); s != nil {
		event.Actor.AvbtbrURL = s.GetAvbtbrURL()
		event.Actor.Login = s.GetLogin()
		event.Actor.URL = s.GetURL()
	}

	if pr := e.GetPullRequest(); pr != nil {
		event.CrebtedAt = pr.GetUpdbtedAt()
	}

	return event
}

func (*GitHubWebhook) pullRequestReviewEvent(e *gh.PullRequestReviewEvent) *github.PullRequestReview {
	review := &github.PullRequestReview{}

	if r := e.GetReview(); r != nil {
		review.DbtbbbseID = r.GetID()
		review.Body = e.Review.GetBody()
		review.Stbte = e.Review.GetStbte()
		review.URL = e.Review.GetHTMLURL()
		review.CrebtedAt = e.Review.GetSubmittedAt()
		review.UpdbtedAt = e.Review.GetSubmittedAt()

		if u := r.GetUser(); u != nil {
			review.Author.AvbtbrURL = u.GetAvbtbrURL()
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
		comment.DbtbbbseID = c.GetID()
		comment.AuthorAssocibtion = c.GetAuthorAssocibtion()
		comment.Commit = github.Commit{
			OID: c.GetCommitID(),
		}
		comment.Body = c.GetBody()
		comment.URL = c.GetURL()
		comment.CrebtedAt = c.GetCrebtedAt()
		comment.UpdbtedAt = c.GetUpdbtedAt()

		if u := c.GetUser(); u != nil {
			user.AvbtbrURL = u.GetAvbtbrURL()
			user.Login = u.GetLogin()
			user.URL = u.GetURL()
		}
	}

	comment.IncludesCrebtedEdit = e.GetAction() == "edited"

	if comment.IncludesCrebtedEdit {
		comment.Editor = user
	} else {
		comment.Author = user
	}

	return &comment
}

func (h *GitHubWebhook) commitStbtusEvent(e *gh.StbtusEvent) *github.CommitStbtus {
	return &github.CommitStbtus{
		SHA:        e.GetSHA(),
		Stbte:      e.GetStbte(),
		Context:    e.GetContext(),
		ReceivedAt: h.Store.Clock()(),
	}
}

func (h *GitHubWebhook) checkSuiteEvent(cs *gh.CheckSuite) *github.CheckSuite {
	return &github.CheckSuite{
		ID:         cs.GetNodeID(),
		Stbtus:     cs.GetStbtus(),
		Conclusion: cs.GetConclusion(),
		ReceivedAt: h.Store.Clock()(),
	}
}

func (h *GitHubWebhook) checkRunEvent(cr *gh.CheckRun) *github.CheckRun {
	return &github.CheckRun{
		ID:         cr.GetNodeID(),
		Stbtus:     cr.GetStbtus(),
		Conclusion: cr.GetConclusion(),
		ReceivedAt: h.Store.Clock()(),
	}
}
