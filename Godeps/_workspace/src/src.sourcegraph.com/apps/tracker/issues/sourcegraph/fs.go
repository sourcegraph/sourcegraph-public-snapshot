// Package sourcegraph implements issues.Service using a the Sourcegraph platform storage API.
package sourcegraph

import (
	"html/template"
	"os"
	"strconv"
	"time"

	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/vcsstore/vcsclient"
	"sourcegraph.com/sqs/pbtypes"
	"src.sourcegraph.com/apps/tracker/issues"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/platform/putil"
	"src.sourcegraph.com/sourcegraph/platform/storage"
	"src.sourcegraph.com/sourcegraph/util/htmlutil"
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

const (
	// threadsBucket is the bucket used for storing issues by thread ID.
	threadsBucket = "threads"

	// commentsBucket is the bucket name prefix used for storing comments. Actual
	// comments for a thread are stored in "comments-<thread ID>".
	commentsBucket = "comments"

	// eventsBucket is the bucket name prefix used for storing events. Actual
	// events for a thread are stored in "events-<thread ID>".
	eventsBucket = "events"
)

func threadCommentsBucket(threadID uint64) string {
	return commentsBucket + "-" + formatUint64(threadID)
}

func threadEventsBucket(threadID uint64) string {
	return eventsBucket + "-" + formatUint64(threadID)
}

func (s service) List(ctx context.Context, repo issues.RepoSpec, opt issues.IssueListOptions) ([]issues.Issue, error) {
	sg := sourcegraph.NewClientFromContext(ctx)
	sys := storage.Namespace(s.appCtx, s.appName, repo.URI)

	var is []issues.Issue

	threads, err := readIDs(sys, threadsBucket)
	if err != nil {
		return is, err
	}
	for i := len(threads); i > 0; i-- {
		threadID := threads[i-1]
		var issue issue
		if err := storage.GetJSON(sys, threadsBucket, formatUint64(threadID), &issue); err != nil {
			return is, err
		}

		if issue.State != opt.State {
			continue
		}

		// Count comments.
		comments, err := sys.List(threadCommentsBucket(threadID))
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

	threads, err := readIDs(sys, threadsBucket)
	if err != nil {
		return 0, err
	}
	for _, threadID := range threads {
		var issue issue
		err := storage.GetJSON(sys, threadsBucket, formatUint64(threadID), &issue)
		if err != nil {
			return 0, err
		}

		if issue.State != opt.State {
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
	err := storage.GetJSON(sys, threadsBucket, formatUint64(id), &issue)
	if err != nil {
		return issues.Issue{}, err
	}

	user, err := sg.Users.Get(ctx, &sourcegraph.UserSpec{UID: issue.AuthorUID})
	if err != nil {
		return issues.Issue{}, err
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

	commentIDs, err := readIDs(sys, threadCommentsBucket(id))
	if err != nil {
		return comments, err
	}
	for _, commentID := range commentIDs {
		var comment comment
		err := storage.GetJSON(sys, threadCommentsBucket(id), formatUint64(commentID), &comment)
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

	eventIDs, err := readIDs(sys, threadEventsBucket(id))
	if err != nil {
		return events, err
	}
	for _, eventID := range eventIDs {
		var event event
		err := storage.GetJSON(sys, threadEventsBucket(id), formatUint64(eventID), &event)
		if err != nil {
			return events, err
		}

		user, err := sg.Users.Get(ctx, &sourcegraph.UserSpec{UID: event.ActorUID})
		if err != nil {
			return events, err
		}
		events = append(events, issues.Event{
			Actor:     sgUser(ctx, user),
			CreatedAt: event.CreatedAt,
			Type:      event.Type,
			Rename:    event.Rename,
		})
	}

	return events, nil
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

	comment := comment{
		AuthorUID: currentUser.UID,
		CreatedAt: time.Now(),
		Body:      c.Body,
	}

	user, err := sg.Users.Get(ctx, &sourcegraph.UserSpec{UID: comment.AuthorUID})
	if err != nil {
		return issues.Comment{}, err
	}

	// Put in storage.
	commentID, err := nextID(sys, threadCommentsBucket(id))
	if err != nil {
		return issues.Comment{}, err
	}
	err = storage.PutJSON(sys, threadCommentsBucket(id), formatUint64(commentID), comment)
	if err != nil {
		return issues.Comment{}, err
	}

	return issues.Comment{
		ID:        commentID,
		User:      sgUser(ctx, user),
		CreatedAt: comment.CreatedAt,
		Body:      comment.Body,
		Editable:  true, // You can always edit comments you've created.
	}, nil
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

	createdAt := time.Now()
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
	threadID, err := nextID(sys, threadsBucket)
	if err != nil {
		return issues.Issue{}, err
	}
	if err := storage.PutJSON(sys, threadsBucket, formatUint64(threadID), issue); err != nil {
		return issues.Issue{}, err
	}

	// Put first comment in storage.
	if err := storage.PutJSON(sys, threadCommentsBucket(threadID), "0", comment); err != nil {
		return issues.Issue{}, err
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
	perm, err := sg.Auth.GetPermissions(ctx, &pbtypes.Void{})
	if err != nil {
		return err
	}
	switch perm.Write {
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
	err := storage.GetJSON(sys, threadsBucket, formatUint64(id), &issue)
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
	err = storage.PutJSON(sys, threadsBucket, formatUint64(id), issue)
	if err != nil {
		return issues.Issue{}, err
	}

	// THINK: Is this the best place to do this? Should it be returned from this func? How would GH backend do it?
	// Create event and commit to storage.
	eventID, err := nextID(sys, threadEventsBucket(id))
	if err != nil {
		return issues.Issue{}, err
	}
	event := event{
		ActorUID:  currentUser.UID,
		CreatedAt: time.Now(),
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
	err = storage.PutJSON(sys, threadEventsBucket(id), formatUint64(eventID), event)
	if err != nil {
		return issues.Issue{}, err
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
	err := storage.GetJSON(sys, threadCommentsBucket(id), formatUint64(c.ID), &comment)
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
	err = storage.PutJSON(sys, threadCommentsBucket(id), formatUint64(c.ID), comment)
	if err != nil {
		return issues.Comment{}, err
	}

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
