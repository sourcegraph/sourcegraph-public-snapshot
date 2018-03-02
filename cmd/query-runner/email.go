package main

import (
	"context"
	"log"
	"runtime"
	"time"

	log15 "gopkg.in/inconshreveable/log15.v2"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/txemail"
)

func (n *notifier) emailNotify(ctx context.Context) {
	canSendEmail, err := api.InternalClient.CanSendEmail(ctx)
	if err != nil {
		log15.Warn("cannot send email notification about saved search (failed to retrieve email configuration)", "error", err)
		return
	}
	if !canSendEmail {
		log15.Warn("cannot send email notification about saved search (SMTP server not in site configuration)")
		return
	}

	// Send tx emails asynchronously.
	go func() {
		defer func() {
			if r := recover(); r != nil {
				// Same as net/http
				const size = 64 << 10
				buf := make([]byte, size)
				buf = buf[:runtime.Stack(buf, false)]
				log.Printf("email notify: failed due to internal panic: %v\n%s", r, buf)
			}
		}()

		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()

		for _, userID := range n.usersToNotify {
			ownership := "the" // example: "new search results have been found for {{.Ownership}} saved search"
			if n.spec.Subject.User != nil && *n.spec.Subject.User == userID {
				ownership = "your"
			}
			if n.spec.Subject.Org != nil {
				ownership = "your organization's"
			}

			plural := ""
			if n.results.Data.Search.Results.ApproximateResultCount != "1" {
				plural = "s"
			}
			sendEmail(ctx, userID, "results", newSearchResultsEmailTemplates, struct {
				URL                    string
				Description            string
				Query                  string
				ApproximateResultCount string
				Ownership              string
				PluralResults          string
			}{
				URL:         searchURL(n.newQuery, utmSourceEmail),
				Description: n.query.Description,
				Query:       n.query.Query,
				ApproximateResultCount: n.results.Data.Search.Results.ApproximateResultCount,
				Ownership:              ownership,
				PluralResults:          plural,
			})
		}
	}()
}

var newSearchResultsEmailTemplates = txemail.MustValidate(txemail.Templates{
	Subject: `[{{.ApproximateResultCount}} new result{{.PluralResults}}] {{.Description}}`,
	Text: `
{{.ApproximateResultCount}} new search result{{.PluralResults}} found for {{.Ownership}} saved search:

  "{{.Description}}"

View the new result{{.PluralResults}} on Sourcegraph: {{.URL}}
`,
	HTML: `
<strong>{{.ApproximateResultCount}}</strong> new search result{{.PluralResults}} found for {{.Ownership}} saved search:

<p style="padding-left: 16px">&quot;{{.Description}}&quot;</p>

<p><a href="{{.URL}}">View the new result{{.PluralResults}} on Sourcegraph</a></p>
`,
})

func emailNotifySubscribeUnsubscribe(ctx context.Context, usersToNotify []int32, query api.SavedQuerySpecAndConfig, template txemail.Templates) {
	canSendEmail, err := api.InternalClient.CanSendEmail(ctx)
	if err != nil {
		log15.Warn("cannot send email notification about saved search (failed to retrieve email configuration)", "error", err)
		return
	}
	if !canSendEmail {
		log15.Warn("cannot send email notification about saved search (SMTP server not in site configuration)")
		return
	}

	var eventType string
	switch {
	case template == notifySubscribedTemplate:
		eventType = "subscribed"
	case template == notifyUnsubscribedTemplate:
		eventType = "unsubscribed"
	default:
		eventType = "unknown"
	}

	// Send tx emails asynchronously.
	go func() {
		defer func() {
			if r := recover(); r != nil {
				// Same as net/http
				const size = 64 << 10
				buf := make([]byte, size)
				buf = buf[:runtime.Stack(buf, false)]
				log.Printf("email notify: failed due to internal panic: %v\n%s", r, buf)
			}
		}()

		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()

		for _, userID := range usersToNotify {
			ownership := "the" // example: "new search results have been found for {{.Ownership}} saved search"
			if query.Spec.Subject.User != nil && *query.Spec.Subject.User == userID {
				ownership = "your"
			}
			if query.Spec.Subject.Org != nil {
				ownership = "your organization's"
			}

			sendEmail(ctx, userID, eventType, template, struct {
				Ownership   string
				Description string
			}{
				Ownership:   ownership,
				Description: query.Config.Description,
			})
		}
	}()
	return
}

func sendEmail(ctx context.Context, userID int32, eventType string, template txemail.Templates, data interface{}) {
	email, err := api.InternalClient.UserEmailsGetEmail(ctx, userID)
	if err != nil {
		log15.Error("email notify: failed to get user email", "user_id", userID, "error", err)
		return
	}
	if email == nil {
		log15.Error("email notify: failed to get user email", "user_id", userID)
		return
	}

	if err := api.InternalClient.SendEmail(ctx, txemail.Message{
		To:       []string{*email},
		Template: template,
		Data:     data,
	}); err != nil {
		log15.Error("email notify: failed to send email", "to", *email, "error", err)
		return
	}
	logEvent(*email, "SavedSearchEmailNotificationSent", eventType)
}

var notifySubscribedTemplate = txemail.MustValidate(txemail.Templates{
	Subject: `[Subscribed] {{.Description}}`,
	Text: `
You are now receiving notifications for {{.Ownership}} saved search:

  "{{.Description}}"

When new search results become available, we will notify you.
`,
	HTML: `
<p>You are now receiving notifications for {{.Ownership}} saved search:</p>

<p style="padding-left: 16px">&quot;{{.Description}}&quot;</p>

<p>When new search results become available, we will notify you.</p>
`,
})

var notifyUnsubscribedTemplate = txemail.MustValidate(txemail.Templates{
	Subject: `[Unsubscribed] {{.Description}}`,
	Text: `
You will no longer receive notifications for {{.Ownership}} saved search:

  "{{.Description}}"

(either you were removed as a person to notify, or the saved search was deleted)
`,
	HTML: `
<p>You will no longer receive notifications for {{.Ownership}} saved search:</p>

<p style="padding-left: 16px">&quot;{{.Description}}&quot;</p>

<p>(either you were removed as a person to notify, or the saved search was deleted)</p>
`,
})
