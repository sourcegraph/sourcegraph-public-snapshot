package kv

import (
	"fmt"
	"html/template"
	"log"
	"net/url"
	"path"
	"time"

	"golang.org/x/net/context"
	notif "src.sourcegraph.com/apps/notifications/notifications" // TODO: Make this better.
	"src.sourcegraph.com/apps/tracker/issues"
	trackerrouter "src.sourcegraph.com/apps/tracker/router"
	sgrouter "src.sourcegraph.com/sourcegraph/app/router"
	"src.sourcegraph.com/sourcegraph/conf"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/platform/notifications"
	"src.sourcegraph.com/sourcegraph/platform/putil"
	"src.sourcegraph.com/sourcegraph/platform/storage"
	"src.sourcegraph.com/sourcegraph/util/mdutil"
)

// subscribe subscribes the author and anyone mentioned in body to the issue.
func (s service) subscribe(ctx context.Context, repo issues.RepoSpec, issueID uint64, author *sourcegraph.UserSpec, body string) error {
	if notifications.Service == nil {
		return nil
	}

	subscribers := []issues.UserSpec{ // Author.
		{
			ID:     uint64(author.UID),
			Domain: author.Domain, // TODO: If blank, set it to "sourcegraph.com"?
		}, // TODO: Dedup?
	}
	mentions, err := mdutil.Mentions(ctx, []byte(body))
	if err != nil {
		return err
	}
	for _, mention := range mentions { // Mentions.
		subscribers = append(subscribers, issues.UserSpec{
			ID:     uint64(mention.UID),
			Domain: mention.Domain,
		})
	}

	return notifications.Service.Subscribe(ctx, s.appName, repo, issueID, subscribers)
}

// markRead marks the specified issueID as read for current user.
func (s service) markRead(ctx context.Context, repo issues.RepoSpec, issueID uint64) error {
	if notifications.Service == nil {
		return nil
	}

	return notifications.Service.MarkRead(ctx, s.appName, repo, issueID)
}

// notify notifies all subscribed users of an update that shows up in their Notification Center.
func (s service) notify(ctx context.Context, repo issues.RepoSpec, issueID uint64, fragment string, sys storage.System, createdAt time.Time, action string) error {
	if notifications.Service == nil {
		return nil
	}

	// TODO: Pass this through events system for asynchronous (rather than blocking) notifications.
	/*{
		events.Publish(events.TrackerCreateCommentEvent, events.TrackerPayload{
			Repo:      sourcegraph.RepoSpec{URI: repo.URI},
			Title:     "Title is TODO in events.Publish track",
			HTMLURL:   "http://www.example.com/TODO",
			UpdatedAt: createdAt,
			State:     "open",
		})
	}*/

	// TODO, THINK: Is this the best place/time?
	// Get issue from storage for to populate notification fields.
	var issue issue
	err := storage.GetJSON(sys, issuesBucket, formatUint64(issueID), &issue)
	if err != nil {
		return err
	}

	// Use Sourcegraph app router for repo app path and Tracker app router for the rest.
	trackerURL, err := sgrouter.Rel.Get(sgrouter.RepoAppFrame).URLPath(
		"Repo", repo.URI,
		"App", s.appName,
		"AppPath", "",
	)
	if err != nil {
		return fmt.Errorf("failed to produce relative URL for tracker app: %v", err)
	}
	issueURL, err := trackerrouter.Router.Get(trackerrouter.Issue).URLPath("id", formatUint64(issueID))
	if err != nil {
		return fmt.Errorf("failed to produce relative URL for issue: %v", err)
	}
	u := &url.URL{
		Path:     path.Join(trackerURL.Path, issueURL.Path),
		Fragment: fragment,
	}
	htmlURL := template.URL(conf.AppURL(s.appCtx).ResolveReference(u).String())

	n := notif.Notification{
		Title:     issue.Title,
		Icon:      notificationIcon(issue.State),
		UpdatedAt: createdAt,
		HTMLURL:   htmlURL,
	}

	if cl, err := sourcegraph.NewClientFromContext(ctx); err == nil {
		// TODO(keegancsmith) we should unify notification center and
		// the notify service
		e := &sourcegraph.NotifyGenericEvent{
			Actor:       putil.UserFromContext(ctx),
			ActionType:  action,
			ObjectID:    int64(issueID),
			ObjectRepo:  repo.URI,
			ObjectType:  s.appName,
			ObjectTitle: n.Title,
			ObjectURL:   string(n.HTMLURL),
		}
		if _, err = cl.Notify.GenericEvent(ctx, e); err != nil {
			log.Println("tracker: failed to Notify.GenericEvent:", err)
		}
	}

	return notifications.Service.Notify(ctx, s.appName, repo, issueID, n)
}

// TODO: This is display/presentation logic; try to factor it out of the backend service implementation.
//       (Have it be provided to the service, maybe? Or another way.)
func notificationIcon(state issues.State) notif.OcticonID {
	switch state {
	case issues.OpenState:
		return "issue-opened"
	case issues.ClosedState:
		return "issue-closed"
	default:
		return ""
	}
}
