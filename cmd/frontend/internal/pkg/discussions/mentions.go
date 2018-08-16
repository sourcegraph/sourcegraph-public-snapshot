package discussions

import (
	"context"
	"fmt"
	"html/template"
	"net/url"
	"path"
	"regexp"
	"strings"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/globals"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/markdown"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/types"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/pkg/errcode"
	"github.com/sourcegraph/sourcegraph/pkg/txemail"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

// NotifyNewThread should be invoked after a new thread (and its first comment)
// have been created, in order to send relevant notifications.
//
// It returns immediately and does not block.
func NotifyNewThread(newThread *types.DiscussionThread, newComment *types.DiscussionComment) {
	notifyMentions(&notifier{
		typ:               newThreadNotification,
		eventAuthorUserID: newComment.AuthorUserID,
		thread:            newThread,
		comment:           newComment,
		template:          newThreadEmailTemplate,
	})
}

// NotifyNewComment should be invoked after a new comment has been added to a
// discussion thread, in order to send relevant notifications.
//
// It returns immediately and does not block.
func NotifyNewComment(updatedThread *types.DiscussionThread, newComment *types.DiscussionComment) {
	notifyMentions(&notifier{
		typ:               newCommentNotification,
		eventAuthorUserID: newComment.AuthorUserID,
		thread:            updatedThread,
		comment:           newComment,
		template:          newCommentEmailTemplate,
	})
}

func notifyMentions(n *notifier) {
	goroutine.Go(func() {
		ctx := context.Background()
		subscribers, err := n.subscribers(ctx)
		if err != nil {
			log15.Error("discussions: determining subscribers", "error", err)
		}
		for _, username := range subscribers {
			if err := n.notifyUsername(ctx, username); err != nil {
				log15.Error("discussions: notifyUsername", "error", err)
			}
		}
	})
}

type notificationType int

const (
	newThreadNotification  notificationType = iota
	newCommentNotification notificationType = iota
)

type notifier struct {
	typ               notificationType
	eventAuthorUserID int32
	thread            *types.DiscussionThread
	comment           *types.DiscussionComment
	template          txemail.Templates
}

// subscribers returns a list of all usernames who are subscribed to receive
// notifications from the thread. Currently, there is no underlying
// subscription store, so we rely on some simple mechanics to get a good-enough
// result:
//
// 	1. If you were previously mentioned in the thread, you are subscribed.
// 	2. If you previously authored a comment, you are subscribed.
//
func (n *notifier) subscribers(ctx context.Context) ([]string, error) {
	comments, err := db.DiscussionComments.List(ctx, &db.DiscussionCommentsListOptions{
		LimitOffset: &db.LimitOffset{
			Limit: 1000,
		},
		ThreadID: &n.thread.ID,
	})
	if err != nil {
		return nil, err
	}
	var (
		subscribers []string
		set         = make(map[string]struct{})
	)
	for _, comment := range comments {
		commentAuthor, err := db.Users.GetByID(ctx, comment.AuthorUserID)
		if err != nil {
			return nil, errors.Wrap(err, "CommentAuthor: GetByID")
		}
		if _, ok := set[commentAuthor.Username]; !ok {
			set[commentAuthor.Username] = struct{}{}
			subscribers = append(subscribers, commentAuthor.Username)
		}
		for _, mention := range parseMentions(comment.Contents) {
			if _, ok := set[mention]; !ok {
				set[mention] = struct{}{}
				subscribers = append(subscribers, mention)
			}
		}
	}
	return subscribers, nil
}

func (n *notifier) notifyUsername(ctx context.Context, username string) error {
	if !conf.CanSendEmail() {
		// Can't send email, so we have nothing to do.
		return nil
	}

	user, err := db.Users.GetByUsername(ctx, username)
	if err != nil {
		return errors.Wrap(err, "GetByUsername")
	}
	if n.eventAuthorUserID == user.ID {
		// Do not send notifications to the user who created the event.
		return nil
	}

	url, err := n.threadCommentURL(ctx)
	if err != nil {
		return errors.Wrap(err, "threadURL")
	}
	if url == "" {
		return nil // can't generate a link to this thread target type
	}

	email, verified, err := db.UserEmails.GetPrimaryEmail(ctx, user.ID)
	if err != nil && !errcode.IsNotFound(err) {
		return errors.Wrap(err, "GetPrimaryEmail")
	}
	if errcode.IsNotFound(err) || !verified {
		// User has no email or it is not verified, do not send them any emails.
		return nil
	}

	var repoShortName string
	if n.thread.TargetRepo != nil {
		repo, err := db.Repos.Get(ctx, n.thread.TargetRepo.RepoID)
		if err != nil {
			return errors.Wrap(err, "repoShortName: db.Repos.Get")
		}
		split := strings.Split(string(repo.URI), "/")
		if len(split) > 2 {
			split = split[len(split)-2:]
		}
		repoShortName = strings.Join(split, "/")
	}

	commentAuthor, err := db.Users.GetByID(ctx, n.comment.AuthorUserID)
	if err != nil {
		return errors.Wrap(err, "CommentAuthor: GetByID")
	}
	fromName := commentAuthor.DisplayName
	if fromName == "" {
		fromName = commentAuthor.Username
	}

	return txemail.Send(ctx, txemail.Message{
		To:       []string{email},
		FromName: fromName,
		Template: n.template,
		Data: struct {
			ThreadTitle         string
			CommentContents     string
			CommentContentsHTML template.HTML
			URL                 string
			RepoName            string
			UniqueValue         string
		}{
			ThreadTitle:         n.thread.Title,
			CommentContents:     n.comment.Contents,
			CommentContentsHTML: template.HTML(markdown.Render(n.comment.Contents, nil)),
			URL:                 url,
			RepoName:            repoShortName,
			UniqueValue:         fmt.Sprint(n.comment.ID),
		},
	})
}

func (n *notifier) threadCommentURL(ctx context.Context) (string, error) {
	t := n.thread
	c := n.comment
	u := &url.URL{}
	switch {
	case t.TargetRepo != nil:
		repo, err := db.Repos.Get(ctx, t.TargetRepo.RepoID)
		if err != nil {
			return "", errors.Wrap(err, "db.Repos.Get")
		}
		u.Path = string(repo.URI)
		if t.TargetRepo.Path != nil {
			u.Path = path.Join(u.Path, "-", "blob", *t.TargetRepo.Path)
		}
		// TODO(slimsag:discussions): frontend doesn't link to the comment directly
		// unless these are in this exact order. Why?
		//fragment := url.Values{}
		//fragment.Set("tab", "discussions")
		//fragment.Set("threadID", strconv.FormatInt(t.ID, 10))
		//fragment.Set("commentID", strconv.FormatInt(c.ID, 10))
		//u.Fragment = fragment.Encode()
		u.Fragment = fmt.Sprintf("tab=discussions&threadID=%v&commentID=%v", t.ID, c.ID)
	default:
		return "", nil // can't generate a link to this target type
	}
	return globals.AppURL.ResolveReference(u).String(), nil
}

var mentions = regexp.MustCompile(`(^|\s)@(\S*)`)

// parseMentions parses the @mentions from the given markdown comment contents.
func parseMentions(contents string) []string {
	matches := mentions.FindAllStringSubmatch(contents, -1)
	mentions := make([]string, 0, len(matches))
	for _, groups := range matches {
		mentions = append(mentions, groups[2])
	}
	return mentions
}

var (
	sharedCommentTextTemplate = `
{{.CommentContents}}

Reply on Sourcegraph:

  {{.URL}}
`

	sharedCommentHTMLTemplate = `
<html>
<body>
<script type="application/ld+json">
{
	"@context": "http://schema.org",
	"@type": "EmailMessage",
	"potentialAction": {
		"@type": "ViewAction",
		"target": "{{.URL}}",
		"name": "View Discussion"
	},
	"description": "View this discussion on Sourcegraph"
}
</script>
{{.CommentContentsHTML}}
<p><a href="{{.URL}}">Reply on Sourcegraph</a></p>
<!-- this ensures Gmail doesn't trim the email -->
<span style="opacity: 0">{{.UniqueValue}}</span>
</body>
</html>
`

	newThreadEmailTemplate = txemail.MustValidate(txemail.Templates{
		Subject: `[{{.RepoName}}] {{.ThreadTitle}}`,
		Text:    sharedCommentTextTemplate,
		HTML:    sharedCommentHTMLTemplate,
	})

	newCommentEmailTemplate = txemail.MustValidate(txemail.Templates{
		Subject: `[{{.RepoName}}] {{.ThreadTitle}}`,
		Text:    sharedCommentTextTemplate,
		HTML:    sharedCommentHTMLTemplate,
	})
)
