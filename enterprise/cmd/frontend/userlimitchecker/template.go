package userlimitchecker

import (
	"github.com/sourcegraph/sourcegraph/internal/txemail"
	"github.com/sourcegraph/sourcegraph/internal/txemail/txtypes"
)

var approachingUserLimitEmailTemplate = txemail.MustValidate(txtypes.Templates{
	Subject: `Your user count is approaching your license's limit`,
	Text: `
Hi there! You're approaching the user limit allowed by your Sourcegraph License. You currently have {{.RemainingUsers}} left.

Reach out to your rep at Sourcegraph if you'd like to increase the limit.
`,
	HTML: `
<p>
Hi there! You're approaching the user limit allowed by your Sourcegraph License. You currently have {{.RemainingUsers}} left.
</p>
<p>Reach out to your rep at Sourcegraph if you'd like to increase the limit.</p>
`,
})
