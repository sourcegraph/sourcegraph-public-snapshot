package campaigns

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/httpcli"

	gh "github.com/google/go-github/v28/github"
	"github.com/hashicorp/go-multierror"
	"github.com/inconshreveable/log15"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	bbs "github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketserver"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/schema"
)

type Webhook struct {
	Store *Store
	Repos repos.Store
	Now   func() time.Time

	// ServiceType corresponds to api.ExternalRepoSpec.ServiceType
	// Example values: bitbucketserver.ServiceType, github.ServiceType
	ServiceType string
}

type PR struct {
	ID             int64
	RepoExternalID string
}

func (h Webhook) getRepoForPR(
	ctx context.Context,
	tx *Store,
	pr PR,
	externalServiceID string,
) (*repos.Repo, error) {
	reposTx := repos.NewDBStore(tx.DB(), sql.TxOptions{})
	rs, err := reposTx.ListRepos(ctx, repos.StoreListReposArgs{
		ExternalRepos: []api.ExternalRepoSpec{
			{
				ID:          pr.RepoExternalID,
				ServiceType: h.ServiceType,
				ServiceID:   externalServiceID,
			},
		},
	})
	if err != nil {
		return nil, errors.Wrap(err, "Failed to load repository")
	}

	if len(rs) != 1 {
		return nil, fmt.Errorf("Fetched repositories have wrong length: %d", len(rs))
	}

	return rs[0], nil
}

func extractExternalServiceID(extSvc *repos.ExternalService) (string, error) {
	c, err := extSvc.Configuration()
	if err != nil {
		return "", errors.Wrap(err, "Failed to get external service config")
	}

	var serviceID string
	switch c := c.(type) {
	case *schema.GitHubConnection:
		serviceID = c.Url
	case *schema.BitbucketServerConnection:
		serviceID = c.Url
	}
	if serviceID == "" {
		return "", errors.New("could not determine service id")
	}

	u, err := url.Parse(serviceID)
	if err != nil {
		return "", errors.Wrap(err, "Failed to parse service ID")
	}

	return extsvc.NormalizeBaseURL(u).String(), nil
}

func (h Webhook) upsertChangesetEvent(
	ctx context.Context,
	externalServiceID string,
	pr PR,
	ev interface{ Key() string },
) (err error) {
	var tx *Store
	if tx, err = h.Store.Transact(ctx); err != nil {
		return err
	}
	defer tx.Done(&err)

	r, err := h.getRepoForPR(ctx, tx, pr, externalServiceID)
	if err != nil {
		log15.Debug("Webhook event could not be matched to repo", "err", err)
		return nil
	}

	cs, err := tx.GetChangeset(ctx, GetChangesetOpts{
		RepoID:              r.ID,
		ExternalID:          strconv.FormatInt(pr.ID, 10),
		ExternalServiceType: h.ServiceType,
	})
	if err != nil {
		if err == ErrNoResults {
			err = nil // Nothing to do
		}
		return err
	}

	now := h.Now()
	event := &campaigns.ChangesetEvent{
		ChangesetID: cs.ID,
		Kind:        campaigns.ChangesetEventKindFor(ev),
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
		// Update is used to create or update the record in the database,
		// but we're actually "patching" the record with specific merge semantics
		// encoded in Update. This is because some webhooks payloads don't contain
		// all the information that we can get from the API, so we only update the
		// bits that we know are more up to date and leave the others as they were.
		existing.Update(event)
		event = existing
	}

	// Add new event
	if err := tx.UpsertChangesetEvents(ctx, event); err != nil {
		return err
	}

	// The webhook may have caused the external state of the changeset to change
	// so we need to update it. We need all events as we may have received more than just the
	// event we are currently handling
	events, _, err := tx.ListChangesetEvents(ctx, ListChangesetEventsOpts{
		ChangesetIDs: []int64{cs.ID},
		Limit:        -1,
	})
	SetDerivedState(cs, events)
	if err := tx.UpdateChangesets(ctx, cs); err != nil {
		return err
	}

	return nil
}

// GitHubWebhook receives GitHub organization webhook events that are
// relevant to campaigns, normalizes those events into ChangesetEvents
// and upserts them to the database.
type GitHubWebhook struct {
	*Webhook
}

type BitbucketServerWebhook struct {
	*Webhook
	Name string

	// externalServiceID -> secret
	// It keeps track of secrets we know have been stored
	// in the remote Bitbucket webhook config
	secrets map[int64]string

	// Optional httpClient
	httpClient httpcli.Doer
}

func NewGitHubWebhook(store *Store, repos repos.Store, now func() time.Time) *GitHubWebhook {
	return &GitHubWebhook{&Webhook{store, repos, now, github.ServiceType}}
}

func NewBitbucketServerWebhook(store *Store, repos repos.Store, now func() time.Time, name string) *BitbucketServerWebhook {
	return &BitbucketServerWebhook{
		Webhook: &Webhook{store, repos, now, bbs.ServiceType},
		Name:    name,
		secrets: make(map[int64]string),
	}
}

// ServeHTTP implements the http.Handler interface.
func (h *GitHubWebhook) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	e, extSvc, httpErr := h.parseEvent(r)
	if httpErr != nil {
		respond(w, httpErr.code, httpErr)
		return
	}

	externalServiceID, err := extractExternalServiceID(extSvc)
	if err != nil {
		respond(w, http.StatusInternalServerError, err)
		return
	}

	prs, ev := h.convertEvent(r.Context(), externalServiceID, e)
	if len(prs) == 0 || ev == nil {
		respond(w, http.StatusOK, nil) // Nothing to do
		return
	}

	m := new(multierror.Error)
	for _, pr := range prs {
		if pr == (PR{}) {
			continue
		}

		err := h.upsertChangesetEvent(r.Context(), externalServiceID, pr, ev)
		if err != nil {
			m = multierror.Append(m, err)
		}
	}
	if m.ErrorOrNil() != nil {
		respond(w, http.StatusInternalServerError, m)
	}
}

func (h *GitHubWebhook) parseEvent(r *http.Request) (interface{}, *repos.ExternalService, *httpError) {
	payload, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, nil, &httpError{http.StatusInternalServerError, err}
	}

	// 🚨 SECURITY: Try to authenticate the request with any of the stored secrets
	// in GitHub external services config. Since n is usually small here,
	// it's ok for this to be have linear complexity.
	// If there are no secrets or no secret managed to authenticate the request,
	// we return a 401 to the client.
	args := repos.StoreListExternalServicesArgs{Kinds: []string{"GITHUB"}}
	es, err := h.Repos.ListExternalServices(r.Context(), args)
	if err != nil {
		return nil, nil, &httpError{http.StatusInternalServerError, err}
	}

	sig := r.Header.Get("X-Hub-Signature")

	var extSvc *repos.ExternalService
	for _, e := range es {
		c, _ := e.Configuration()
		for _, hook := range c.(*schema.GitHubConnection).Webhooks {
			if hook.Secret == "" {
				continue
			}

			if err = gh.ValidateSignature(sig, payload, []byte(hook.Secret)); err == nil {
				extSvc = e
				break
			}
		}
	}

	if extSvc == nil || err != nil {
		return nil, nil, &httpError{http.StatusUnauthorized, err}
	}

	e, err := gh.ParseWebHook(gh.WebHookType(r), payload)
	if err != nil {
		return nil, nil, &httpError{http.StatusBadRequest, err}
	}

	return e, extSvc, nil
}

func (h *GitHubWebhook) convertEvent(ctx context.Context, externalServiceID string, theirs interface{}) (prs []PR, ours interface{ Key() string }) {
	log15.Debug("GitHub webhook received", "type", fmt.Sprintf("%T", theirs))
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
		pr := PR{ID: int64(*e.Number), RepoExternalID: repoExternalID}
		prs = append(prs, pr)

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
			ServiceID:   externalServiceID,
			ServiceType: "github",
		}

		ids, err := h.Store.GetChangesetExternalIDs(ctx, spec, refs)
		if err != nil {
			log15.Error("Error executing GetChangesetExternalIDs", "err", err)
			return nil, nil
		}

		for _, id := range ids {
			i, err := strconv.ParseInt(id, 10, 64)
			if err != nil {
				log15.Error("Error parsing external id", "err", err)
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

		repo := cs.GetRepository()
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

		cs := cr.GetCheckSuite()
		if cs == nil {
			return
		}

		repo := cs.GetRepository()
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

	return
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
		comment.CreatedAt = c.GetCreatedAt()
		comment.UpdatedAt = c.GetUpdatedAt()
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
		labelEvent.CreatedAt = pr.GetUpdatedAt()
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

func (*GitHubWebhook) assignedEvent(e *gh.PullRequestEvent) *github.AssignedEvent {
	assignedEvent := &github.AssignedEvent{}

	if pr := e.GetPullRequest(); pr != nil {
		assignedEvent.CreatedAt = pr.GetUpdatedAt()
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
		unassignedEvent.CreatedAt = pr.GetUpdatedAt()
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
		event.CreatedAt = pr.GetUpdatedAt()
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
		event.CreatedAt = pr.GetUpdatedAt()
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
		event.CreatedAt = pr.GetUpdatedAt()
	}

	if ch := e.GetChanges(); ch != nil && ch.Title != nil && ch.Title.From != nil {
		event.PreviousTitle = *ch.Title.From
	}

	return event
}

func (*GitHubWebhook) closedEvent(e *gh.PullRequestEvent) *github.ClosedEvent {
	event := &github.ClosedEvent{}

	if s := e.GetSender(); s != nil {
		event.Actor.AvatarURL = s.GetAvatarURL()
		event.Actor.Login = s.GetLogin()
		event.Actor.URL = s.GetURL()
	}

	if pr := e.GetPullRequest(); pr != nil {
		event.CreatedAt = pr.GetUpdatedAt()

		// This is different from the URL returned by GraphQL because the precise
		// event URL isn't available in this webhook payload. This means if we expose
		// this URL in the UI, and users click it, they'll just go to the PR page, rather
		// than the precise location of the "close" event, until the background syncing
		// runs and updates this URL to the exact one.
		event.URL = pr.GetURL()
	}

	return event
}

func (*GitHubWebhook) reopenedEvent(e *gh.PullRequestEvent) *github.ReopenedEvent {
	event := &github.ReopenedEvent{}

	if s := e.GetSender(); s != nil {
		event.Actor.AvatarURL = s.GetAvatarURL()
		event.Actor.Login = s.GetLogin()
		event.Actor.URL = s.GetURL()
	}

	if pr := e.GetPullRequest(); pr != nil {
		event.CreatedAt = pr.GetUpdatedAt()
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
		review.CreatedAt = e.Review.GetSubmittedAt()
		review.UpdatedAt = e.Review.GetSubmittedAt()

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
		comment.CreatedAt = c.GetCreatedAt()
		comment.UpdatedAt = c.GetUpdatedAt()

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

func (*GitHubWebhook) commitStatusEvent(e *gh.StatusEvent) *github.CommitStatus {
	return &github.CommitStatus{
		SHA:        e.GetSHA(),
		State:      e.GetState(),
		Context:    e.GetContext(),
		ReceivedAt: time.Now(),
	}
}

func (*GitHubWebhook) checkSuiteEvent(cs *gh.CheckSuite) *github.CheckSuite {
	return &github.CheckSuite{
		ID:         cs.GetNodeID(),
		Status:     cs.GetStatus(),
		Conclusion: cs.GetConclusion(),
		ReceivedAt: time.Now(),
	}
}

func (*GitHubWebhook) checkRunEvent(cr *gh.CheckRun) *github.CheckRun {
	return &github.CheckRun{
		ID:         cr.GetNodeID(),
		Status:     cr.GetStatus(),
		Conclusion: cr.GetConclusion(),
		ReceivedAt: time.Now(),
	}
}

// SyncWebhooks ensures the creation / deletion of the BitbucketServer campaigns webhook.
// This happens periodically at the specified interval.
func (h *BitbucketServerWebhook) SyncWebhooks(every time.Duration) {
	externalURL := func() string {
		return conf.Cached(func() interface{} {
			return conf.Get().ExternalURL
		})().(string)
	}

	for {
		args := repos.StoreListExternalServicesArgs{Kinds: []string{"BITBUCKETSERVER"}}
		es, err := h.Repos.ListExternalServices(context.Background(), args)
		if err != nil {
			log15.Error("Upserting BBS Webhook failed [Listing BBS extsvc]", "err", err)
			continue
		}

		for _, e := range es {
			c, _ := e.Configuration()
			con, ok := c.(*schema.BitbucketServerConnection)
			if !ok {
				continue
			}
			err = h.syncWebhook(e.ID, con, externalURL())
			if err != nil {
				log15.Error("Upserting BBS Webhook failed", "err", err)
			}
		}

		time.Sleep(every)
	}
}

const externalServiceIDParam = "externalServiceID"

// syncWebook ensures that the webhook has been configured correctly on Bitbucket. If no secret has been set, we delete
// the exising webhook config.
func (h *BitbucketServerWebhook) syncWebhook(externalServiceID int64, con *schema.BitbucketServerConnection, externalURL string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	secret := con.WebhookSecret()
	oldSecret, ok := h.secrets[externalServiceID]

	if ok && oldSecret == secret {
		// Nothing has changed since our last check
		return nil
	}

	client, err := bbs.NewClient(con, h.httpClient)
	if err != nil {
		return errors.Wrap(err, "creating client")
	}

	if secret == "" {
		// Secret now blank, delete hook.
		// If this is the first iteration we don't know if the server
		// has a hook configured or not. If not, the delete will be a noop
		err = client.DeleteWebhook(ctx, h.Name)
		if err != nil {
			return errors.Wrap(err, "deleting webhook")
		}
		h.secrets[externalServiceID] = secret
		return nil
	}

	// Secret has changed to a non blank value, upsert
	endpoint := fmt.Sprintf("%s/.api/bitbucket-server-webhooks?%s=%d", externalURL, externalServiceIDParam, externalServiceID)
	wh := bbs.Webhook{
		Name:     h.Name,
		Scope:    "global",
		Events:   []string{"pr", "repo"},
		Endpoint: endpoint,
		Secret:   secret,
	}

	err = client.UpsertWebhook(ctx, wh)
	if err != nil {
		return errors.Wrap(err, "upserting webhook")
	}
	h.secrets[externalServiceID] = secret
	return nil
}

func (h *BitbucketServerWebhook) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	e, extSvc, hErr := h.parseEvent(r)
	if hErr != nil {
		respond(w, hErr.code, hErr)
		return
	}

	externalServiceID, err := extractExternalServiceID(extSvc)
	if err != nil {
		respond(w, http.StatusInternalServerError, err)
		return
	}

	prs, ev := h.convertEvent(e)

	m := new(multierror.Error)
	for _, pr := range prs {
		if pr == (PR{}) {
			log15.Warn("Dropping Bitbucket Server webhook event", "type", fmt.Sprintf("%T", e))
			continue
		}

		err := h.upsertChangesetEvent(r.Context(), externalServiceID, pr, ev)
		if err != nil {
			m = multierror.Append(m, err)
		}
	}
	if m.ErrorOrNil() != nil {
		respond(w, http.StatusInternalServerError, m)
	}
}

func (h *BitbucketServerWebhook) parseEvent(r *http.Request) (interface{}, *repos.ExternalService, *httpError) {
	payload, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, nil, &httpError{http.StatusInternalServerError, err}
	}

	sig := r.Header.Get("X-Hub-Signature")

	rawID := r.FormValue(externalServiceIDParam)
	var externalServiceID int64
	// id could be blank temporarily if we haven't updated the hook url to include the param yet
	if rawID != "" {
		externalServiceID, err = strconv.ParseInt(rawID, 10, 64)
		if err != nil {
			return nil, nil, &httpError{http.StatusBadRequest, errors.Wrap(err, "invalid external service id")}
		}
	}

	args := repos.StoreListExternalServicesArgs{Kinds: []string{"BITBUCKETSERVER"}}
	if externalServiceID != 0 {
		args.IDs = append(args.IDs, externalServiceID)
	}
	es, err := h.Repos.ListExternalServices(r.Context(), args)
	if err != nil {
		return nil, nil, &httpError{http.StatusInternalServerError, err}
	}

	var extSvc *repos.ExternalService
	for _, e := range es {
		if externalServiceID != 0 && e.ID != externalServiceID {
			continue
		}

		c, _ := e.Configuration()
		con, ok := c.(*schema.BitbucketServerConnection)
		if !ok {
			continue
		}

		if secret := con.WebhookSecret(); secret != "" {
			if err = gh.ValidateSignature(sig, payload, []byte(secret)); err == nil {
				extSvc = e
				break
			}
		}
	}

	if extSvc == nil || err != nil {
		return nil, nil, &httpError{http.StatusUnauthorized, err}
	}

	e, err := bbs.ParseWebhookEvent(bbs.WebhookEventType(r), payload)
	if err != nil {
		return nil, nil, &httpError{http.StatusBadRequest, err}
	}
	return e, extSvc, nil
}

func (h *BitbucketServerWebhook) convertEvent(theirs interface{}) (prs []PR, ours interface{ Key() string }) {
	log15.Debug("Bitbucket Server webhook received", "type", fmt.Sprintf("%T", theirs))

	switch e := theirs.(type) {
	case *bbs.PullRequestEvent:
		repoID := strconv.Itoa(e.PullRequest.FromRef.Repository.ID)
		pr := PR{ID: int64(e.PullRequest.ID), RepoExternalID: repoID}
		prs = append(prs, pr)
		return prs, e.Activity
	case *bbs.BuildStatusEvent:
		for _, p := range e.PullRequests {
			repoID := strconv.Itoa(p.FromRef.Repository.ID)
			pr := PR{ID: int64(p.ID), RepoExternalID: repoID}
			prs = append(prs, pr)
		}
		return prs, &bbs.CommitStatus{
			Commit: e.Commit,
			Status: e.Status,
		}
	}

	return
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
