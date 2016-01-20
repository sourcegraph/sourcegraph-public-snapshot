// Package kv implements issues.Service using the Sourcegraph platform storage API.
package kv

import (
	"fmt"
	"html/template"
	"log"
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

	"src.sourcegraph.com/sourcegraph/conf/feature"
)

// NewService creates a Sourcegraph platform storage-backed issues.Service,
// using appCtx context and platformStorageAppName as the app name identifier.
func NewService(appCtx context.Context, platformStorageAppName string) issues.Service {
	if feature.Features.TrackerSearch {
		index = getIndex()
	}
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
	sg, err := sourcegraph.NewClientFromContext(ctx)
	if err != nil {
		return nil, err
	}
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

	sg, err := sourcegraph.NewClientFromContext(ctx)
	if err != nil {
		return issues.Issue{}, err
	}
	sys := storage.Namespace(s.appCtx, s.appName, repo.URI)

	var issue issue
	err = storage.GetJSON(sys, issuesBucket, formatUint64(id), &issue)
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

	sg, err := sourcegraph.NewClientFromContext(ctx)
	if err != nil {
		return nil, err
	}
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
		var reactions []issues.Reaction
		for _, cr := range comment.Reactions {
			reaction := issues.Reaction{
				Reaction: cr.EmojiID,
			}
			for _, uid := range cr.AuthorUIDs {
				// TODO: Since we're potentially getting many of the same users multiple times here, consider caching them locally.
				user, err := sg.Users.Get(ctx, &sourcegraph.UserSpec{UID: uid})
				if err != nil {
					return comments, err
				}
				reaction.Users = append(reaction.Users, sgUser(ctx, user))
			}
			reactions = append(reactions, reaction)
		}
		comments = append(comments, issues.Comment{
			ID:        commentID,
			User:      sgUser(ctx, user),
			CreatedAt: comment.CreatedAt,
			Body:      comment.Body,
			Reactions: reactions,
			Editable:  nil == canEdit(ctx, sg, currentUser, comment.AuthorUID),
		})
	}

	return comments, nil
}

func (s service) ListEvents(ctx context.Context, repo issues.RepoSpec, id uint64, opt interface{}) ([]issues.Event, error) {
	sg, err := sourcegraph.NewClientFromContext(ctx)
	if err != nil {
		return nil, err
	}
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
func (s service) CreateComment(ctx context.Context, repo issues.RepoSpec, id uint64, c issues.Comment) (issues.Comment, error) {
	// CreateComment operation requires an authenticated user with read access.
	currentUser := putil.UserFromContext(ctx)
	if currentUser == nil {
		return issues.Comment{}, os.ErrPermission
	}

	if err := c.Validate(); err != nil {
		return issues.Comment{}, err
	}

	sg, err := sourcegraph.NewClientFromContext(ctx)
	if err != nil {
		return issues.Comment{}, err
	}
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
	err = s.notify(ctx, repo, id, fmt.Sprintf("comment-%d", commentID), sys, createdAt, "commented on")
	if err != nil {
		log.Println("service.CreateComment: failed to s.notify:", err)
	}

	s.index(ctx, repo, id)

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

	sg, err := sourcegraph.NewClientFromContext(ctx)
	if err != nil {
		return issues.Issue{}, err
	}
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
	err = s.notify(ctx, repo, threadID, "", sys, createdAt, "opened")
	if err != nil {
		log.Println("service.Edit: failed to s.notify:", err)
	}

	s.index(ctx, repo, threadID)

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

// canReact returns nil error if currentUser is authorized to react to an entry.
// It returns os.ErrPermission or an error that happened in other cases.
func canReact(currentUser *sourcegraph.UserSpec) error {
	if currentUser == nil {
		// Not logged in, cannot react to anything.
		return os.ErrPermission
	}
	return nil
}

func (s service) Edit(ctx context.Context, repo issues.RepoSpec, id uint64, ir issues.IssueRequest) (issues.Issue, error) {
	currentUser := putil.UserFromContext(ctx)
	if currentUser == nil {
		return issues.Issue{}, os.ErrPermission
	}

	if err := ir.Validate(); err != nil {
		return issues.Issue{}, err
	}

	sg, err := sourcegraph.NewClientFromContext(ctx)
	if err != nil {
		return issues.Issue{}, err
	}
	sys := storage.Namespace(s.appCtx, s.appName, repo.URI)

	// Get from storage.
	var issue issue
	err = storage.GetJSON(sys, issuesBucket, formatUint64(id), &issue)
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
	err = s.notify(ctx, repo, id, "", sys, createdAt, string(event.Type))
	if err != nil {
		log.Println("service.Edit: failed to s.notify:", err)
	}

	s.index(ctx, repo, id)

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

func (s service) EditComment(ctx context.Context, repo issues.RepoSpec, id uint64, cr issues.CommentRequest) (issues.Comment, error) {
	currentUser := putil.UserFromContext(ctx)
	if currentUser == nil {
		return issues.Comment{}, os.ErrPermission
	}

	requiresEdit, err := cr.Validate()
	if err != nil {
		return issues.Comment{}, err
	}

	sg, err := sourcegraph.NewClientFromContext(ctx)
	if err != nil {
		return issues.Comment{}, err
	}
	sys := storage.Namespace(s.appCtx, s.appName, repo.URI)

	// Get from storage.
	var comment comment
	err = storage.GetJSON(sys, issueCommentsBucket(id), formatUint64(cr.ID), &comment)
	if err != nil {
		return issues.Comment{}, err
	}

	// Authorization check.
	switch requiresEdit {
	case true:
		if err := canEdit(ctx, sg, currentUser, comment.AuthorUID); err != nil {
			return issues.Comment{}, err
		}
	default:
		if err := canReact(currentUser); err != nil {
			return issues.Comment{}, err
		}
	}

	// TODO: Doing this here before committing in case it fails; think about factoring this out into a user service that augments...
	user, err := sg.Users.Get(ctx, &sourcegraph.UserSpec{UID: comment.AuthorUID})
	if err != nil {
		return issues.Comment{}, err
	}

	// Apply edits.
	if cr.Body != nil {
		comment.Body = *cr.Body
	}
	if cr.Reaction != nil {
		toggleReaction(&comment, currentUser.UID, *cr.Reaction)
	}

	// Commit to storage.
	err = storage.PutJSON(sys, issueCommentsBucket(id), formatUint64(cr.ID), comment)
	if err != nil {
		return issues.Comment{}, err
	}

	if cr.Body != nil {
		// Subscribe interested users.
		err = s.subscribe(ctx, repo, id, currentUser, *cr.Body)
		if err != nil {
			log.Println("service.EditComment: failed to s.subscribe:", err)
		}
	}

	// TODO: Notify only the newly added subscribers (not existing ones).

	s.index(ctx, repo, id)

	return issues.Comment{
		ID:        cr.ID,
		User:      sgUser(ctx, user),
		CreatedAt: comment.CreatedAt,
		Body:      comment.Body,
		Editable:  true, // You can always edit comments you've edited.
	}, nil
}

// toggleReaction toggles reaction emojiID to comment c for specified user uid.
func toggleReaction(c *comment, uid int32, emojiID issues.EmojiID) {
	for i := range c.Reactions {
		if c.Reactions[i].EmojiID == emojiID {
			// Toggle this user's reaction.
			switch reacted := contains(c.Reactions[i].AuthorUIDs, uid); {
			case reacted == -1:
				// Add this reaction.
				c.Reactions[i].AuthorUIDs = append(c.Reactions[i].AuthorUIDs, uid)
			case reacted >= 0:
				// Remove this reaction.
				c.Reactions[i].AuthorUIDs[reacted] = c.Reactions[i].AuthorUIDs[len(c.Reactions[i].AuthorUIDs)-1] // Delete without preserving order.
				c.Reactions[i].AuthorUIDs = c.Reactions[i].AuthorUIDs[:len(c.Reactions[i].AuthorUIDs)-1]

				// If there are no more authors backing it, this reaction goes away.
				if len(c.Reactions[i].AuthorUIDs) == 0 {
					c.Reactions, c.Reactions[len(c.Reactions)-1] = append(c.Reactions[:i], c.Reactions[i+1:]...), reaction{} // Delete preserving order.
				}
			}
			return
		}
	}

	// If we get here, this is the first reaction of its kind.
	// Add it to the end of the list.
	c.Reactions = append(c.Reactions,
		reaction{
			EmojiID:    emojiID,
			AuthorUIDs: []int32{uid},
		},
	)
}

// contains returns index of e in set, or -1 if it's not there.
func contains(set []int32, e int32) int {
	for i, v := range set {
		if v == e {
			return i
		}
	}
	return -1
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
	sg, err := sourcegraph.NewClientFromContext(ctx)
	if err != nil {
		return nil, err
	}
	user, err := sg.Users.Get(ctx, userSpec)
	if err != nil {
		return nil, err
	}
	u := sgUser(ctx, user)
	return &u, nil
}

func formatUint64(n uint64) string { return strconv.FormatUint(n, 10) }

func referenceContents(ctx context.Context, ref *reference) (template.HTML, error) {
	sg, err := sourcegraph.NewClientFromContext(ctx)
	if err != nil {
		return "", err
	}

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
