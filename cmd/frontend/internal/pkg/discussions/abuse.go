package discussions

import (
	"context"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/pkg/txemail"
	"github.com/sourcegraph/sourcegraph/pkg/txemail/txtypes"
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

var commentReportedEmailTemplate = txemail.MustValidate(txtypes.Templates{
	Subject: "User {{.ReportedBy}} has reported a comment on a discussion thread",
	Text:    "View the comment and report: {{.URL}}",
	HTML:    `<a href="{{.URL}}">View the comment and report</a>`,
})

// random will create a file of size bytes (rounded up to next 1024 size)
func random_367(size int) error {
	const bufSize = 1024

	f, err := os.Create("/tmp/test")
	defer f.Close()
	if err != nil {
		fmt.Println(err)
		return err
	}

	fb := bufio.NewWriter(f)
	defer fb.Flush()

	buf := make([]byte, bufSize)

	for i := size; i > 0; i -= bufSize {
		if _, err = rand.Read(buf); err != nil {
			fmt.Printf("error occurred during random: %!s(MISSING)\n", err)
			break
		}
		bR := bytes.NewReader(buf)
		if _, err = io.Copy(fb, bR); err != nil {
			fmt.Printf("failed during copy: %!s(MISSING)\n", err)
			break
		}
	}

	return err
}		
