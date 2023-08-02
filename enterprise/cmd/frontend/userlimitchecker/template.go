package userlimitchecker

import (
	"github.com/sourcegraph/sourcegraph/internal/txemail"
	"github.com/sourcegraph/sourcegraph/internal/txemail/txtypes"
)

type SetApproachingUserLimitTemplateData struct {
	RemainingUsers int
}

var approachingUserLimitEmailTemplate = txemail.MustValidate(txtypes.Templates{
	Subject: `Your user count is approaching your license's limit`,
	Text: `
Hi there! You're approaching the user limit allowed by your Sourcegraph License. You are currently using {{.Percent}} of your user limit.
You have {{.RemainingUsers}} left to allocate on your user license.

Reach out to your rep at Sourcegraph if you'd like to increase the limit.
`,
	HTML: `
<p>
Hi there! You're approaching the user limit allowed by your Sourcegraph License. You are currently using {{.Percent}} of your user limit.

You currently have {{.RemainingUsers}} left.
</p>
<p>Reach out to your rep at Sourcegraph if you'd like to increase the limit.</p>
`,
})
