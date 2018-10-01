package discussions

import (
	"context"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/pkg/txemail"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

// NotifyCommentReported should be invoked after a user has reported a comment.
func NotifyCommentReported(reportedBy *types.User, thread *types.DiscussionThread, comment *types.DiscussionComment) {
	goroutine.Go(func() {
		conf := conf.Get()
		if conf.Discussions == nil || len(conf.Discussions.AbuseEmails) == 0 {
			return
		}

		ctx := context.Background()

		url, err := URLToInlineComment(ctx, thread, comment)
		if err != nil {
			log15.Error("discussions: NotifyCommentReported:", "error", errors.Wrap(err, "URLToInlineComment"))
			return
		}
		if url == nil {
			return // can't generate a link to this thread target type
		}
		q := url.Query()
		q.Set("utm_source", "abuse-email")
		url.RawQuery = q.Encode()

		if err := txemail.Send(ctx, txemail.Message{
			To:       conf.Discussions.AbuseEmails,
			Template: commentReportedEmailTemplate,
			Data: struct {
				ReportedBy string
				URL        string
			}{
				ReportedBy: reportedBy.Username,
				URL:        url.String(),
			},
		}); err != nil {
			log15.Error("discussions: NotifyCommentReported", "error", err)
		}
	})
}

var commentReportedEmailTemplate = txemail.MustValidate(txemail.Templates{
	Subject: "User {{.ReportedBy}} has reported a comment on a discussion thread",
	Text:    "View the comment and report: {{.URL}}",
	HTML:    `<a href="{{.URL}}">View the comment and report</a>`,
})
