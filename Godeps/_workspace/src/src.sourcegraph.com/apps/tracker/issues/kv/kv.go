// Package kv implements issues.Service using the Sourcegraph platform storage API.
package kv

import (
	"fmt"
	"html/template"
	"log"
	"net/url"
	"os"
	"path"
	"strconv"
	"time"

	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/vcsstore/vcsclient"
	"sourcegraph.com/sqs/pbtypes"
	notif "src.sourcegraph.com/apps/notifications/notifications" // TODO: Make this better.
	"src.sourcegraph.com/apps/tracker/issues"
	trackerrouter "src.sourcegraph.com/apps/tracker/router"
	sgrouter "src.sourcegraph.com/sourcegraph/app/router"
	"src.sourcegraph.com/sourcegraph/conf"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/platform/notifications"
	"src.sourcegraph.com/sourcegraph/platform/putil"
	"src.sourcegraph.com/sourcegraph/platform/storage"
	"src.sourcegraph.com/sourcegraph/util/htmlutil"
	"src.sourcegraph.com/sourcegraph/util/mdutil"
)

// NewService creates a Sourcegraph platform storage-backed issues.Service,
// using appCtx context and platformStorageAppName as the app name identifier.
func NewService(appCtx context.Context, platformStorageAppName string) issues.Service {
	return service{
		appCtx:  appCtx,
		appName: platformStorageAppName,
	}
}

type service struct {
	// appCtx is the app context with high priveldge. It's used to access the Sourcegraph platform storage
	// (on behalf of users that may not have write access). This service implementation is responsible for doing
	// authorization checks.
	appCtx context.Context

	// appName is the app name used for Sourcegraph platform storage.
	appName string
}

func (s service) List(ctx context.Context, repo issues.RepoSpec, opt issues.IssueListOptions) ([]issues.Issue, error) {
	sg := sourcegraph.NewClientFromContext(ctx)
	sys := storage.Namespace(s.appCtx, s.appName, repo.URI)

	var is []issues.Issue

	threads, err := readIDs(sys, issuesBucket)
	if err != nil {
		return is, err
	}
	for i := len(threads); i > 0; i-- {
		threadID := threads[i-1]
		var issue issue
		if err := storage.GetJSON(sys, issuesBucket, formatUint64(threadID), &issue); err != nil {
			return is, err
		}

		if opt.State != issues.AllStates && issue.State != issues.State(opt.State) {
			continue
		}

		// Count comments.
		comments, err := sys.List(issueCommentsBucket(threadID))
		if err != nil {
			return is, err
		}
		user, err := sg.Users.Get(ctx, &sourcegraph.UserSpec{UID: issue.AuthorUID})
		if err != nil {
			return is, err
		}
		is = append(is, issues.Issue{
			ID:    threadID,
			State: issue.State,
			Title: issue.Title,
			Comment: issues.Comment{
				User:      sgUser(ctx, user),
				CreatedAt: issue.CreatedAt,
			},
			Replies: len(comments) - 1,
		})
	}

	return is, nil
}

func (s service) Count(ctx context.Context, repo issues.RepoSpec, opt issues.IssueListOptions) (uint64, error) {
	sys := storage.Namespace(s.appCtx, s.appName, repo.URI)

	var count uint64

	threads, err := readIDs(sys, issuesBucket)
	if err != nil {
		return 0, err
	}
	for _, threadID := range threads {
		var issue issue
		err := storage.GetJSON(sys, issuesBucket, formatUint64(threadID), &issue)
		if err != nil {
			return 0, err
		}

		if opt.State != issues.AllStates && issue.State != issues.State(opt.State) {
			continue
		}

		count++
	}

	return count, nil
}

func (s service) Get(ctx context.Context, repo issues.RepoSpec, id uint64) (issues.Issue, error) {
	currentUser := putil.UserFromContext(ctx)

	sg := sourcegraph.NewClientFromContext(ctx)
	sys := storage.Namespace(s.appCtx, s.appName, repo.URI)

	var issue issue
	err := storage.GetJSON(sys, issuesBucket, formatUint64(id), &issue)
	if err != nil {
		return issues.Issue{}, err
	}

	user, err := sg.Users.Get(ctx, &sourcegraph.UserSpec{UID: issue.AuthorUID})
	if err != nil {
		return issues.Issue{}, err
	}

	if currentUser != nil {
		// Mark as read.
		err = s.markRead(ctx, repo, id)
		if err != nil {
			log.Println("service.Get: failed to s.markRead:", err)
		}
	}

	var reference *issues.Reference
	if issue.Reference != nil {
		contents, err := referenceContents(ctx, issue.Reference)
		if err != nil {
			return issues.Issue{}, err
		}
		reference = &issues.Reference{
			Repo:      issue.Reference.Repo,
			Path:      issue.Reference.Path,
			CommitID:  issue.Reference.CommitID,
			StartLine: issue.Reference.StartLine,
			EndLine:   issue.Reference.EndLine,
			Contents:  contents,
		}
	}
	return issues.Issue{
		ID:    id,
		State: issue.State,
		Title: issue.Title,
		Comment: issues.Comment{
			User:      sgUser(ctx, user),
			CreatedAt: issue.CreatedAt,
			Editable:  nil == canEdit(ctx, sg, currentUser, issue.AuthorUID),
		},
		Reference: reference,
	}, nil
}

func (s service) ListComments(ctx context.Context, repo issues.RepoSpec, id uint64, opt interface{}) ([]issues.Comment, error) {
	currentUser := putil.UserFromContext(ctx)

	sg := sourcegraph.NewClientFromContext(ctx)
	sys := storage.Namespace(s.appCtx, s.appName, repo.URI)

	var comments []issues.Comment

	commentIDs, err := readIDs(sys, issueCommentsBucket(id))
	if err != nil {
		return comments, err
	}
	for _, commentID := range commentIDs {
		var comment comment
		err := storage.GetJSON(sys, issueCommentsBucket(id), formatUint64(commentID), &comment)
		if err != nil {
			return comments, err
		}

		user, err := sg.Users.Get(ctx, &sourcegraph.UserSpec{UID: comment.AuthorUID})
		if err != nil {
			return comments, err
		}
		comments = append(comments, issues.Comment{
			ID:        commentID,
			User:      sgUser(ctx, user),
			CreatedAt: comment.CreatedAt,
			Body:      comment.Body,
			Editable:  nil == canEdit(ctx, sg, currentUser, comment.AuthorUID),
		})
	}

	return comments, nil
}

func (s service) ListEvents(ctx context.Context, repo issues.RepoSpec, id uint64, opt interface{}) ([]issues.Event, error) {
	sg := sourcegraph.NewClientFromContext(ctx)
	sys := storage.Namespace(s.appCtx, s.appName, repo.URI)

	var events []issues.Event

	eventIDs, err := readIDs(sys, issueEventsBucket(id))
	if err != nil {
		return events, err
	}
	for _, eventID := range eventIDs {
		var event event
		err := storage.GetJSON(sys, issueEventsBucket(id), formatUint64(eventID), &event)
		if err != nil {
			return events, err
		}

		user, err := sg.Users.Get(ctx, &sourcegraph.UserSpec{UID: event.ActorUID})
		if err != nil {
			return events, err
		}
		events = append(events, issues.Event{
			ID:        eventID,
			Actor:     sgUser(ctx, user),
			CreatedAt: event.CreatedAt,
			Type:      event.Type,
			Rename:    event.Rename,
		})
	}

	return events, nil
}

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
func (s service) notify(ctx context.Context, repo issues.RepoSpec, issueID uint64, fragment string, sys storage.System, createdAt time.Time) error {
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

	return notifications.Service.Notify(ctx, s.appName, repo, issueID, notif.Notification{
		Title:     issue.Title,
		Icon:      notificationIcon(issue.State),
		UpdatedAt: createdAt,
		HTMLURL:   htmlURL,
	})
}

func (s service) CreateComment(ctx context.Context, repo issues.RepoSpec, id uint64, c issues.Comment) (issues.Comment, error) {
	// CreateComment operation requires an authenticated user with read access.
	currentUser := putil.UserFromContext(ctx)
	if currentUser == nil {
		return issues.Comment{}, os.ErrPermission
	}

	if err := c.Validate(); err != nil {
		return issues.Comment{}, err
	}

	sg := sourcegraph.NewClientFromContext(ctx)
	sys := storage.Namespace(s.appCtx, s.appName, repo.URI)

	createdAt := time.Now().UTC()
	comment := comment{
		AuthorUID: currentUser.UID,
		CreatedAt: createdAt,
		Body:      c.Body,
	}

	user, err := sg.Users.Get(ctx, &sourcegraph.UserSpec{UID: comment.AuthorUID})
	if err != nil {
		return issues.Comment{}, err
	}

	// Put in storage.
	commentID, err := nextID(sys, issueCommentsBucket(id))
	if err != nil {
		return issues.Comment{}, err
	}
	err = storage.PutJSON(sys, issueCommentsBucket(id), formatUint64(commentID), comment)
	if err != nil {
		return issues.Comment{}, err
	}

	// Subscribe interested users.
	err = s.subscribe(ctx, repo, id, currentUser, c.Body)
	if err != nil {
		log.Println("service.CreateComment: failed to s.subscribe:", err)
	}

	// Notify subscribed users.
	// TODO: Use Tracker app router to compute fragment; that logic shouldn't be duplicated here.
	err = s.notify(ctx, repo, id, fmt.Sprintf("comment-%d", commentID), sys, createdAt)
	if err != nil {
		log.Println("service.CreateComment: failed to s.notify:", err)
	}

	return issues.Comment{
		ID:        commentID,
		User:      sgUser(ctx, user),
		CreatedAt: comment.CreatedAt,
		Body:      comment.Body,
		Editable:  true, // You can always edit comments you've created.
	}, nil
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

func (s service) Create(ctx context.Context, repo issues.RepoSpec, i issues.Issue) (issues.Issue, error) {
	// Create operation requires an authenticated user with read access.
	currentUser := putil.UserFromContext(ctx)
	if currentUser == nil {
		return issues.Issue{}, os.ErrPermission
	}

	if err := i.Validate(); err != nil {
		return issues.Issue{}, err
	}

	sg := sourcegraph.NewClientFromContext(ctx)
	sys := storage.Namespace(s.appCtx, s.appName, repo.URI)

	createdAt := time.Now().UTC()
	comment := comment{
		AuthorUID: currentUser.UID,
		CreatedAt: createdAt,
		Body:      i.Body,
	}
	issue := issue{
		State:     issues.OpenState,
		Title:     i.Title,
		AuthorUID: currentUser.UID,
		CreatedAt: createdAt,
	}
	if ref := i.Reference; ref != nil {
		issue.Reference = &reference{
			Repo:      ref.Repo,
			Path:      ref.Path,
			CommitID:  ref.CommitID,
			StartLine: ref.StartLine,
			EndLine:   ref.EndLine,
		}
	}

	user, err := sg.Users.Get(ctx, &sourcegraph.UserSpec{UID: issue.AuthorUID})
	if err != nil {
		return issues.Issue{}, err
	}

	// Put in storage.
	threadID, err := nextID(sys, issuesBucket)
	if err != nil {
		return issues.Issue{}, err
	}
	if err := storage.PutJSON(sys, issuesBucket, formatUint64(threadID), issue); err != nil {
		return issues.Issue{}, err
	}

	// Put first comment in storage.
	if err := storage.PutJSON(sys, issueCommentsBucket(threadID), "0", comment); err != nil {
		return issues.Issue{}, err
	}

	// Subscribe interested users.
	err = s.subscribe(ctx, repo, threadID, currentUser, i.Body)
	if err != nil {
		log.Println("service.Create: failed to s.subscribe:", err)
	}

	// Notify subscribed users.
	err = s.notify(ctx, repo, threadID, "", sys, createdAt)
	if err != nil {
		log.Println("service.Edit: failed to s.notify:", err)
	}

	return issues.Issue{
		ID:    threadID,
		State: issue.State,
		Title: issue.Title,
		Comment: issues.Comment{
			ID:        0,
			User:      sgUser(ctx, user),
			CreatedAt: issue.CreatedAt,
			Body:      comment.Body,
			Editable:  true, // You can always edit issues you've created.
		},
	}, nil
}

// canEdit returns nil error if currentUser is authorized to edit an entry created by authorUID.
// It returns os.ErrPermission or an error that happened in other cases.
func canEdit(ctx context.Context, sg *sourcegraph.Client, currentUser *sourcegraph.UserSpec, authorUID int32) error {
	if currentUser == nil {
		// Not logged in, cannot edit anything.
		return os.ErrPermission
	}
	if currentUser.UID == authorUID {
		// If you're the author, you can always edit it.
		return nil
	}
	authInfo, err := sg.Auth.Identify(ctx, &pbtypes.Void{})
	if err != nil {
		return err
	}
	switch authInfo.Write {
	case true:
		// If you have write access (or greater), you can edit.
		return nil
	default:
		return os.ErrPermission
	}
}

func (s service) Edit(ctx context.Context, repo issues.RepoSpec, id uint64, ir issues.IssueRequest) (issues.Issue, error) {
	currentUser := putil.UserFromContext(ctx)
	if currentUser == nil {
		return issues.Issue{}, os.ErrPermission
	}

	if err := ir.Validate(); err != nil {
		return issues.Issue{}, err
	}

	sg := sourcegraph.NewClientFromContext(ctx)
	sys := storage.Namespace(s.appCtx, s.appName, repo.URI)

	// Get from storage.
	var issue issue
	err := storage.GetJSON(sys, issuesBucket, formatUint64(id), &issue)
	if err != nil {
		return issues.Issue{}, err
	}

	// Authorization check.
	if err := canEdit(ctx, sg, currentUser, issue.AuthorUID); err != nil {
		return issues.Issue{}, err
	}

	// TODO: Doing this here before committing in case it fails; think about factoring this out into a user service that augments...
	user, err := sg.Users.Get(ctx, &sourcegraph.UserSpec{UID: issue.AuthorUID})
	if err != nil {
		return issues.Issue{}, err
	}

	// Apply edits.
	if ir.State != nil {
		issue.State = *ir.State
	}
	origTitle := issue.Title
	if ir.Title != nil {
		issue.Title = *ir.Title
	}

	// Put in storage.
	err = storage.PutJSON(sys, issuesBucket, formatUint64(id), issue)
	if err != nil {
		return issues.Issue{}, err
	}

	// THINK: Is this the best place to do this? Should it be returned from this func? How would GH backend do it?
	// Create event and commit to storage.
	createdAt := time.Now().UTC()
	event := event{
		ActorUID:  currentUser.UID,
		CreatedAt: createdAt,
	}
	switch {
	case ir.State != nil && *ir.State == issues.OpenState:
		event.Type = issues.Reopened
	case ir.State != nil && *ir.State == issues.ClosedState:
		event.Type = issues.Closed
	case ir.Title != nil:
		event.Type = issues.Renamed
		event.Rename = &issues.Rename{
			From: origTitle,
			To:   *ir.Title,
		}
	}
	eventID, err := nextID(sys, issueEventsBucket(id))
	if err != nil {
		return issues.Issue{}, err
	}
	err = storage.PutJSON(sys, issueEventsBucket(id), formatUint64(eventID), event)
	if err != nil {
		return issues.Issue{}, err
	}

	// Subscribe interested users.
	err = s.subscribe(ctx, repo, id, currentUser, "")
	if err != nil {
		log.Println("service.Edit: failed to s.subscribe:", err)
	}

	// Notify subscribed users.
	// TODO: Maybe set fragment to fmt.Sprintf("event-%d", eventID), etc.
	err = s.notify(ctx, repo, id, "", sys, createdAt)
	if err != nil {
		log.Println("service.Edit: failed to s.notify:", err)
	}

	return issues.Issue{
		ID:    id,
		State: issue.State,
		Title: issue.Title,
		Comment: issues.Comment{
			ID:        0,
			User:      sgUser(ctx, user),
			CreatedAt: issue.CreatedAt,
			Editable:  true, // You can always edit issues you've edited.
		},
	}, nil
}

func (s service) EditComment(ctx context.Context, repo issues.RepoSpec, id uint64, c issues.Comment) (issues.Comment, error) {
	currentUser := putil.UserFromContext(ctx)
	if currentUser == nil {
		return issues.Comment{}, os.ErrPermission
	}

	if err := c.Validate(); err != nil {
		return issues.Comment{}, err
	}

	sg := sourcegraph.NewClientFromContext(ctx)
	sys := storage.Namespace(s.appCtx, s.appName, repo.URI)

	// Get from storage.
	var comment comment
	err := storage.GetJSON(sys, issueCommentsBucket(id), formatUint64(c.ID), &comment)
	if err != nil {
		return issues.Comment{}, err
	}

	// Authorization check.
	if err := canEdit(ctx, sg, currentUser, comment.AuthorUID); err != nil {
		return issues.Comment{}, err
	}

	// TODO: Doing this here before committing in case it fails; think about factoring this out into a user service that augments...
	user, err := sg.Users.Get(ctx, &sourcegraph.UserSpec{UID: comment.AuthorUID})
	if err != nil {
		return issues.Comment{}, err
	}

	// Apply edits.
	comment.Body = c.Body

	// Commit to storage.
	err = storage.PutJSON(sys, issueCommentsBucket(id), formatUint64(c.ID), comment)
	if err != nil {
		return issues.Comment{}, err
	}

	// Subscribe interested users.
	err = s.subscribe(ctx, repo, id, currentUser, c.Body)
	if err != nil {
		log.Println("service.EditComment: failed to s.subscribe:", err)
	}

	// TODO: Notify only the newly added subscribers (not existing ones).

	return issues.Comment{
		ID:        c.ID,
		User:      sgUser(ctx, user),
		CreatedAt: comment.CreatedAt,
		Body:      comment.Body,
		Editable:  true, // You can always edit comments you've edited.
	}, nil
}

// nextID returns the next id for the given bucket. If there are no previous
// keys in the bucket, it begins with id 1.
func nextID(sys storage.System, bucket string) (uint64, error) {
	ids, err := readIDs(sys, bucket)
	if err != nil {
		return 0, err
	}
	if len(ids) == 0 {
		return 1, nil
	}
	return ids[len(ids)-1] + 1, nil
}

// TODO.
func (service) CurrentUser(ctx context.Context) (*issues.User, error) {
	userSpec := putil.UserFromContext(ctx)
	if userSpec == nil {
		// Not authenticated, no current user.
		return nil, nil
	}
	sg := sourcegraph.NewClientFromContext(ctx)
	user, err := sg.Users.Get(ctx, userSpec)
	if err != nil {
		return nil, err
	}
	u := sgUser(ctx, user)
	return &u, nil
}

func formatUint64(n uint64) string { return strconv.FormatUint(n, 10) }

func referenceContents(ctx context.Context, ref *reference) (template.HTML, error) {
	sg := sourcegraph.NewClientFromContext(ctx)

	te, err := sg.RepoTree.Get(ctx, &sourcegraph.RepoTreeGetOp{
		Entry: sourcegraph.TreeEntrySpec{
			Path: ref.Path,
			RepoRev: sourcegraph.RepoRevSpec{
				RepoSpec: sourcegraph.RepoSpec{URI: ref.Repo.URI},
				CommitID: ref.CommitID,
			},
		},
		Opt: &sourcegraph.RepoTreeGetOptions{
			Formatted: true,
			GetFileOptions: vcsclient.GetFileOptions{
				FileRange: vcsclient.FileRange{
					StartLine: int64(ref.StartLine),
					EndLine:   int64(ref.EndLine),
				},
			},
		},
	})
	if err != nil {
		return "", err
	}

	sanitizedContents := htmlutil.SanitizeForPB(string(te.Contents)).HTML
	return template.HTML(sanitizedContents), nil
}
