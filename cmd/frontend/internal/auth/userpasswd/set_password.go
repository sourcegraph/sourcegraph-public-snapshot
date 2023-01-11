package userpasswd

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/globals"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/txemail"
	"github.com/sourcegraph/sourcegraph/internal/txemail/txtypes"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

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

// HandleSetPasswordEmail sends the password reset email directly to the user for users created by site admins.
func HandleSetPasswordEmail(ctx context.Context, db database.DB, id int32) (string, error) {
	e, _, err := db.UserEmails().GetPrimaryEmail(ctx, id)
	if err != nil {
		return "", errors.Wrap(err, "get user primary email")
	}

	usr, err := db.Users().GetByID(ctx, id)
	if err != nil {
		return "", errors.Wrap(err, "get user by ID")
	}

	ru, err := backend.MakePasswordResetURL(ctx, db, id)
	if err == database.ErrPasswordResetRateLimit {
		return "", err
	} else if err != nil {
		return "", errors.Wrap(err, "make password reset URL")
	}

	// Configure the template
	emailTemplate := defaultSetPasswordEmailTemplate
	if customTemplates := conf.SiteConfig().EmailTemplates; customTemplates != nil {
		emailTemplate = txemail.FromSiteConfigTemplateWithDefault(customTemplates.SetPassword, emailTemplate)
	}

	rus := globals.ExternalURL().ResolveReference(ru).String()
	if err := txemail.Send(ctx, "password_set", txemail.Message{
		To:       []string{e},
		Template: emailTemplate,
		Data: struct {
			Username string
			URL      string
			Host     string
		}{
			Username: usr.Username,
			URL:      rus,
			Host:     globals.ExternalURL().Host,
		},
	}); err != nil {
		return "", err
	}

	return rus, nil
}
