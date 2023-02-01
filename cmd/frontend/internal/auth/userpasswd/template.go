package userpasswd

import (
	"github.com/sourcegraph/sourcegraph/internal/txemail"
	"github.com/sourcegraph/sourcegraph/internal/txemail/txtypes"
)

type SetPasswordEmailTemplateData struct {
	Username string
	URL      string
	Host     string
}

var defaultSetPasswordEmailTemplate = txemail.MustValidate(txtypes.Templates{
	Subject: `Set your Sourcegraph password ({{.Host}})`,
	Text: `
Your administrator created an account for you on Sourcegraph ({{.Host}}).

To set the password for {{.Username}} on Sourcegraph, follow this link:

  {{.URL}}
`,
	HTML: `
<p>
  Your administrator created an account for you on Sourcegraph ({{.Host}}).
</p>

<p><strong><a href="{{.URL}}">Set password for {{.Username}}</a></strong></p>
`,
})

var defaultResetPasswordEmailTemplates = txemail.MustValidate(txtypes.Templates{
	Subject: `Reset your Sourcegraph password ({{.Host}})`,
	Text: `
Somebody (likely you) requested a password reset for the user {{.Username}} on Sourcegraph ({{.Host}}).

To reset the password for {{.Username}} on Sourcegraph, follow this link:

  {{.URL}}
`,
	HTML: `
<p>
  Somebody (likely you) requested a password reset for <strong>{{.Username}}</strong>
  on Sourcegraph ({{.Host}}).
</p>

<p><strong><a href="{{.URL}}">Reset password for {{.Username}}</a></strong></p>
`,
})
