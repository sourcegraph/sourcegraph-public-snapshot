package main

import (
	"context"
	"log"
	"runtime"
	"strings"
	"time"

	log15 "gopkg.in/inconshreveable/log15.v2"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/txemail"
)

func (n *notifier) emailNotify(ctx context.Context) {
	if !conf.CanSendEmail() {
		log15.Warn("cannot send email notification about saved search (SMTP server not in site configuration")
		return
	}

	// Send tx emails asynchronously.
	go func() {
		if r := recover(); r != nil {
			// Same as net/http
			const size = 64 << 10
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]
			log.Printf("email notify: failed due to internal panic: %v\n%s", r, buf)
		}

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

			sendEmail(ctx, userID, "results", newSearchResultsEmailTemplates, struct {
				URL                    string
				Description            string
				Query                  string
				ApproximateResultCount string
				Ownership              string
				MoreThanTwoResults     bool
			}{
				URL:         searchURL(n.newQuery, utmSourceEmail),
				Description: n.query.Description,
				Query:       strings.Join([]string{n.query.ScopeQuery, n.query.Query}, " "),
				ApproximateResultCount: n.results.Data.Search.Results.ApproximateResultCount,
				Ownership:              ownership,
				MoreThanTwoResults:     n.results.Data.Search.Results.ApproximateResultCount != "1",
			})
		}
	}()
}

var newSearchResultsEmailTemplates = txemail.MustParseTemplate(txemail.Templates{
	Subject: `{{.ApproximateResultCount}} new search results found - {{.Description}}`,
	Text: `
{{.ApproximateResultCount}} new search result{{if .MoreThanTwoResults}}s{{end}} have been found for {{.Ownership}} saved search:

  "{{.Description}}"

View the new results on Sourcegraph: {{.URL}}
`,
	HTML: `
<strong>{{.ApproximateResultCount}}</strong> new search results have been found for {{.Ownership}} saved search:

<p style="padding-left: 16px">&quot;{{.Description}}&quot;</p>

<p><a href="{{.URL}}">View the new search results on Sourcegraph</a></p>
`,
})

func emailNotifySubscribeUnsubscribe(ctx context.Context, usersToNotify []int32, query api.SavedQuerySpecAndConfig, template txemail.ParsedTemplates) {
	if !conf.CanSendEmail() {
		log15.Warn("cannot send email notification about saved search (SMTP server not in site configuration")
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
		if r := recover(); r != nil {
			// Same as net/http
			const size = 64 << 10
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]
			log.Printf("email notify: failed due to internal panic: %v\n%s", r, buf)
		}

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

func sendEmail(ctx context.Context, userID int32, eventType string, template txemail.ParsedTemplates, data interface{}) {
	email, err := api.InternalClient.UserEmailsGetEmail(ctx, userID)
	if err != nil {
		log15.Error("email notify: failed to get user email", "user_id", userID, "error", err)
		return
	}
	if email == nil {
		log15.Error("email notify: failed to get user email", "user_id", userID)
		return
	}

	if err := txemail.Send(ctx, txemail.Message{
		To:       []string{*email},
		Template: template,
		Data:     data,
	}); err != nil {
		log15.Error("email notify: failed to send email", "to", *email, "error", err)
		return
	}
	logEvent(*email, "SavedSearchEmailNotificationSent", eventType)
}

var notifySubscribedTemplate = txemail.MustParseTemplate(txemail.Templates{
	Subject: `Subscribed to saved search: {{.Description}}`,
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

var notifyUnsubscribedTemplate = txemail.MustParseTemplate(txemail.Templates{
	Subject: `Unsubscribed from saved search: {{.Description}}`,
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
