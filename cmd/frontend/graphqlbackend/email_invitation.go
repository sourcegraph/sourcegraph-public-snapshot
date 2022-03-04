package graphqlbackend

import (
	"context"
	"net/url"
	"strconv"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/globals"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/rcache"
	"github.com/sourcegraph/sourcegraph/internal/txemail"
	"github.com/sourcegraph/sourcegraph/internal/txemail/txtypes"
)

var (
	disableEmailInvites, _                    = strconv.ParseBool(env.Get("DISABLE_EMAIL_INVITES", "false", "Disable email invitations entirely."))
	debugEmailInvitesMock, _                  = strconv.ParseBool(env.Get("DEBUG_EMAIL_INVITES_MOCK", "false", "Do not actually send email invitations, instead just print that we did."))
	debugEmailInvitesDisableSpamProtection, _ = strconv.ParseBool(env.Get("DEBUG_EMAIL_INVITES_DISABLE_SPAM_PROTECTION", "false", "Disables spam protection"))

	invitedEmailsLimiter = rcache.NewWithTTL("invited_emails", 60*60*24) // 24h
)

func (r *schemaResolver) InviteEmailToSourcegraph(ctx context.Context, args *struct {
	Email string
}) (*EmptyResponse, error) {
	if disableEmailInvites {
		return nil, errors.New("email invites disabled.")
	}
	// You must be authenticated to send email invites (we need to know who it is from.)
	a := actor.FromContext(ctx)
	if !a.IsAuthenticated() {
		return nil, errors.New("no current user")
	}
	invitedBy, err := a.User(ctx, r.db.Users())
	if err != nil {
		return nil, err
	}

	// Email sending always happens in the background so the GraphQL request is fast.
	goroutine.Go(func() {
		// SECURITY: We only allow inviting the same email address once every 24 hours to reduce the potential
		// for spam.
		debug := debugEmailInvitesDisableSpamProtection || debugEmailInvitesMock
		if !debug {
			_, alreadyInvitedRecently := invitedEmailsLimiter.Get(args.Email)
			if alreadyInvitedRecently {
				log15.Warn("email invites: refusing to send email invite (spam prevention): already tried inviting in last 24h", "email", args.Email)
				return
			}
		}
		invitedEmailsLimiter.Set(args.Email, []byte{})

		if debugEmailInvitesMock {
			log15.Info("email invites: mock invited to Sourcegraph", "invited_by", invitedBy.Username, "invited", args.Email)
			return
		}

		emailTemplateEmailInvitation := emailTemplateEmailInvitationServer
		if envvar.SourcegraphDotComMode() {
			emailTemplateEmailInvitation = emailTemplateEmailInvitationCloud
		}

		urlSignUp, _ := url.Parse("/sign-up?invitedBy=" + invitedBy.Username)
		if err := txemail.Send(ctx, txemail.Message{
			To:       []string{args.Email},
			Template: emailTemplateEmailInvitation,
			Data: struct {
				FromName string
				URL      string
			}{
				FromName: invitedBy.Username,
				URL:      globals.ExternalURL().ResolveReference(urlSignUp).String(),
			},
		}); err != nil {
			log15.Warn("email invites: failed to send email", "error", err)
			invitedEmailsLimiter.Delete(args.Email) // allow attempting to invite this email again without waiting 24h
			return
		}
		log15.Info("email invites: invitation sent", "from", invitedBy.Username, "to", args.Email)
	})
	return &EmptyResponse{}, nil
}

var emailTemplateEmailInvitationCloud = txemail.MustValidate(txtypes.Templates{
	Subject: `{{.FromName}} has invited you to Sourcegraph`,
	Text: `
Sourcegraph enables you to quickly understand and fix your code.

You can use Sourcegraph to:
  - Search and navigate multiple repositories with cross-repository dependency navigation
  - Share links directly to lines of code to work more collaboratively with your team
  - Search more than 2 million open source repositories, all in one place
  - Create code monitors to alert you about changes in code

Join {{ .FromName }} on Sourcegraph to experience the power of great code search.

Claim your invitation:

{{.URL}}

Learn more about Sourcegraph:

https://about.sourcegraph.com
`,
	HTML: `
<p>Sourcegraph enables you to quickly understand and fix your code.</p>

<p>
	You can use Sourcegraph to:<br/>
	<ul>
		<li>Search and navigate multiple repositories with cross-repository dependency navigation</li>
		<li>Share links directly to lines of code to work more collaboratively with your team</li>
		<li>Search more than 2 million open source repositories, all in one place</li>
		<li>Create code monitors to alert you about changes in code</li>
	</ul>
</p>

<p>Join <strong>{{.FromName}}</strong> on Sourcegraph to experience the power of great code search.</p>

<p><strong><a href="{{.URL}}">Claim your invitation</a></strong></p>

<p><a href="https://about.sourcegraph.com">Learn more about Sourcegraph</a></p>
`,
})

var emailTemplateEmailInvitationServer = txemail.MustValidate(txtypes.Templates{
	Subject: `{{.FromName}} has invited you to Sourcegraph`,
	Text: `
Sourcegraph enables your team to quickly understand, fix, and automate changes to your code.

You can use Sourcegraph to:
  - Search and navigate multiple repositories with cross-repository dependency navigation
  - Share links directly to lines of code to work more collaboratively together
  - Automate large-scale code changes with Batch Changes
  - Create code monitors to alert you about changes in code

Join {{ .FromName }} on Sourcegraph to experience the power of great code search.

Claim your invitation:

{{.URL}}

Learn more about Sourcegraph:

https://about.sourcegraph.com
`,
	HTML: `
<p>Sourcegraph enables your team to quickly understand, fix, and automate changes to your code.</p>

<p>
	You can use Sourcegraph to:<br/>
	<ul>
		<li>Search and navigate multiple repositories with cross-repository dependency navigation</li>
		<li>Share links directly to lines of code to work more collaboratively together</li>
		<li>Automate large-scale code changes with Batch Changes</li>
		<li>Create code monitors to alert you about changes in code</li>
	</ul>
</p>

<p>Join <strong>{{.FromName}}</strong> on Sourcegraph to experience the power of great code search.</p>

<p><strong><a href="{{.URL}}">Claim your invitation</a></strong></p>

<p><a href="https://about.sourcegraph.com">Learn more about Sourcegraph</a></p>
`,
})
