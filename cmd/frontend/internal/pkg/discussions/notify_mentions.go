package discussions

import (
	"context"
	"fmt"
	"html/template"
	"path"
	"strings"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/discussions/mentions"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/markdown"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/pkg/errcode"
	"github.com/sourcegraph/sourcegraph/pkg/txemail"
	"github.com/sourcegraph/sourcegraph/pkg/txemail/txtypes"
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
	template          txtypes.Templates
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
	for _, mention := range mentions.Parse(n.thread.Title) {
		if _, ok := set[mention]; !ok {
			set[mention] = struct{}{}
			subscribers = append(subscribers, mention)
		}
	}
	for _, comment := range comments {
		commentAuthor, err := db.Users.GetByID(ctx, comment.AuthorUserID)
		if err != nil {
			return nil, errors.Wrap(err, "CommentAuthor: GetByID")
		}
		if _, ok := set[commentAuthor.Username]; !ok {
			set[commentAuthor.Username] = struct{}{}
			subscribers = append(subscribers, commentAuthor.Username)
		}
		for _, mention := range mentions.Parse(comment.Contents) {
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

	var (
		replyTo    *string
		messageID  *string
		references []string
	)
	if conf.CanReadEmail() {
		// Generate a secure token that will allow the notified user to reply
		// via email securely.
		//
		// 🚨 SECURITY: It is crucial that the user ID and thread ID passed here
		// are correct, as the token effectively grants anonymous posting in the
		// specified thread on the specified user's behalf.
		secureToken, err := db.DiscussionMailReplyTokens.Generate(ctx, user.ID, n.thread.ID)
		if err != nil {
			return errors.Wrap(err, "DiscussionMailReplyTokens.Generate")
		}

		conf := conf.Get()
		emailParts := strings.Split(conf.EmailImap.Username, "@")
		secureReplyTo := fmt.Sprintf("%s+%s@%s", emailParts[0], secureToken, emailParts[1])
		replyTo = &secureReplyTo

		// Generate a unique message ID. This is used by e.g. Gmail to uniquely
		// identify this email message and so that we can reference it in later
		// messages and have them all properly show up in the same email thread.
		msgID := func(commentID int64) string {
			return fmt.Sprintf("%s+%d.%d@%s", emailParts[0], n.thread.ID, commentID, emailParts[1])
		}
		id := msgID(n.comment.ID)
		messageID = &id

		// Get a list of prior comments in the thread and generate the
		// references list. This makes e.g. Gmail understand that this email is
		// part of the thread.
		comments, err := db.DiscussionComments.List(ctx, &db.DiscussionCommentsListOptions{
			LimitOffset: &db.LimitOffset{
				Limit: 100,
			},
			ThreadID: &n.thread.ID,
		})
		if err != nil {
			return errors.Wrap(err, "DiscussionComments.List")
		}
		for _, comment := range comments {
			if comment.ID == n.comment.ID {
				continue
			}
			references = append(references, msgID(comment.ID))
		}
	}

	url, err := URLToInlineComment(ctx, n.thread, n.comment)
	if err != nil {
		return errors.Wrap(err, "URLToInlineComment")
	}
	if url == nil {
		return nil // can't generate a link to this thread target type
	}
	q := url.Query()
	q.Set("utm_source", "email")
	url.RawQuery = q.Encode()

	email, verified, err := db.UserEmails.GetPrimaryEmail(ctx, user.ID)
	if err != nil && !errcode.IsNotFound(err) {
		return errors.Wrap(err, "GetPrimaryEmail")
	}
	if errcode.IsNotFound(err) || !verified {
		// User has no email or it is not verified, do not send them any emails.
		return nil
	}

	// TODO(sqs): This only takes the 1st target. Support multiple targets.
	var (
		repoShortName   string
		fileName        string
		codeContextText string
		codeContextHTML template.HTML
	)
	targets, err := db.DiscussionThreads.ListTargets(ctx, db.DiscussionThreadsListTargetsOptions{ThreadID: n.thread.ID})
	if err != nil {
		return errors.Wrap(err, "DiscussionThreads.ListTargets")
	}
	if len(targets) > 0 {
		target := targets[0]
		repo, err := db.Repos.Get(ctx, target.RepoID)
		if err != nil {
			return errors.Wrap(err, "repoShortName: db.Repos.Get")
		}
		split := strings.Split(string(repo.Name), "/")
		if len(split) > 2 {
			split = split[len(split)-2:]
		}
		repoShortName = strings.Join(split, "/")
		if target.Path != nil {
			fileName = path.Base(*target.Path)
		}

		codeContextText = formatTargetRepoLinesText(target)
		codeContextHTML, err = formatTargetRepoLinesHTML(ctx, target)
		if err != nil {
			return errors.Wrap(err, "formatTargetRepoLinesHTML")
		}
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
		To:         []string{email},
		FromName:   fromName,
		ReplyTo:    replyTo,
		MessageID:  messageID,
		References: references,
		Template:   n.template,
		Data: struct {
			ThreadTitle           string
			CommentAuthorUsername string
			CommentContents       string
			CommentContentsHTML   template.HTML
			URL                   string
			UniqueValue           string
			CanReply              bool

			// These fields may be empty strings depending on the type of comment..
			RepoName        string
			FileName        string
			CodeContextText string
			CodeContextHTML template.HTML
		}{
			ThreadTitle:           n.thread.Title,
			CommentAuthorUsername: commentAuthor.Username,
			CommentContents:       n.comment.Contents,
			CommentContentsHTML:   template.HTML(markdown.Render(n.comment.Contents, nil)),
			URL:                   url.String(),
			UniqueValue:           fmt.Sprint(n.comment.ID),
			CanReply:              conf.CanReadEmail(),

			RepoName:        repoShortName,
			FileName:        fileName,
			CodeContextText: codeContextText,
			CodeContextHTML: codeContextHTML,
		},
	})
}

var (
	sharedCommentSubjectTemplate = `
{{- with .RepoName -}}
	{{- "[" -}}{{- . -}}{{- "] " -}}
{{- end -}}
{{- .ThreadTitle -}}
`

	sharedCommentTextTemplate = `
{{- "@" -}}{{- .CommentAuthorUsername -}}{{- " commented" -}}
	{{- with .FileName -}}{{- " on " -}}{{- . -}}{{- end -}}
	{{- ":\n" -}}
{{- .CommentContents -}}
{{- with .CodeContextText -}}
	{{- "\n" -}}
	{{- "--------------------------------------------------------------------------------\n" -}}
	{{- . -}}
{{- end -}}
{{- "\n" -}}
{{- "—\n" -}}
{{- if .CanReply -}}
	{{- "Reply to this email directly, or view it on Sourcegraph:\n" -}}
{{- else -}}
	{{- "View and reply on Sourcegraph:\n" -}}
{{- end -}}
{{- "\n" -}}
{{- "  " -}}{{- .URL -}}
{{- "\n" -}}
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
<p><strong>@{{.CommentAuthorUsername}}</strong> commented{{with .FileName}} on <strong>{{.}}</strong>{{end}}:</p>
{{.CommentContentsHTML}}
{{with .CodeContextHTML}}
	{{.}}
{{end}}
{{if .CanReply}}
	<p style="font-size: small; color: #666;">—<br/>Reply to this email directly or <a href="{{.URL}}">view it on Sourcegraph</a>.</p>
{{else}}
	<p style="font-size: small; color: #666;">—<br/><a href="{{.URL}}">View and reply on Sourcegraph</a></p>
{{end}}
<!-- this ensures Gmail doesn't trim the email -->
<span style="opacity: 0">{{.UniqueValue}}</span>
</body>
</html>
`
	newThreadEmailTemplate = txemail.MustValidate(txtypes.Templates{
		Subject: sharedCommentSubjectTemplate,
		Text:    sharedCommentTextTemplate,
		HTML:    sharedCommentHTMLTemplate,
	})

	newCommentEmailTemplate = txemail.MustValidate(txtypes.Templates{
		Subject: sharedCommentSubjectTemplate,
		Text:    sharedCommentTextTemplate,
		HTML:    sharedCommentHTMLTemplate,
	})
)
