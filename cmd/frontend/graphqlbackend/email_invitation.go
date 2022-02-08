package graphqlbackend

import (
	"context"
	"net/url"
	"strconv"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/lib/errors"

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

var emailTemplateEmailInvitation = txemail.MustValidate(txtypes.Templates{
	Subject: `{{.FromName}} has invited you to Sourcegraph`,
	Text: `
{{.FromName}} has invited you to Sourcegraph

Sourcegraph is browser-based code search and navigation at its core. Features like code insights, batch changes, code monitors, and a corpus of more than 2 million open source repositories are why the top developers at the world’s most innovated companies can’t live without it.

Claim your invitation:

{{.URL}}

Learn more about Sourcegraph:

https://about.sourcegraph.com
`,
	HTML: `
<p><strong>{{.FromName}}</strong> has invited you to Sourcegraph</p>

<p>Sourcegraph is browser-based code search and navigation at its core. Features like code insights, batch changes, code monitors, and a corpus of more than 2 million open source repositories are why the top developers at the world’s most innovated companies can’t live without it.</p>

<p><strong><a href="{{.URL}}">Claim your invitation</a></strong></p>

<p><a href="https://about.sourcegraph.com">Learn more about Sourcegraph</a></p>
`,
})
