package background

import (
	"context"
	_ "embed"
	"fmt"
	"net/url"
	"sync"

	"github.com/graph-gophers/graphql-go/relay"

	edb "github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/api/internalapi"
	"github.com/sourcegraph/sourcegraph/internal/txemail"
	"github.com/sourcegraph/sourcegraph/internal/txemail/txtypes"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// To avoid a circular dependency with the codemonitors/resolvers package
// we have to redeclare the MonitorKind.
const MonitorKind = "CodeMonitor"
const utmSourceEmail = "code-monitoring-email"
const priorityCritical = "CRITICAL"

var MockSendEmailForNewSearchResult func(ctx context.Context, userID int32, data *TemplateDataNewSearchResults) error
var MockExternalURL func() *url.URL

func SendEmailForNewSearchResult(ctx context.Context, userID int32, data *TemplateDataNewSearchResults) error {
	if MockSendEmailForNewSearchResult != nil {
		return MockSendEmailForNewSearchResult(ctx, userID, data)
	}
	return sendEmail(ctx, userID, newSearchResultsEmailTemplates, data)
}

var (
	//go:embed email_template.html.tmpl
	htmlTemplate string

	//go:embed email_template.txt.tmpl
	textTemplate string
)

var newSearchResultsEmailTemplates = txemail.MustValidate(txtypes.Templates{
	Subject: `{{ if .IsTest }}Test: {{ end }}[{{.Priority}} event] {{.Description}}`,
	Text:    textTemplate,
	HTML:    htmlTemplate,
})

type TemplateDataNewSearchResults struct {
	Priority                  string
	CodeMonitorURL            string
	SearchURL                 string
	Description               string
	NumberOfResultsWithDetail string
	IsTest                    bool
}

func NewTemplateDataForNewSearchResults(args actionArgs, email *edb.EmailAction) (d *TemplateDataNewSearchResults, err error) {
	var (
		priority                  string
		numberOfResultsWithDetail string
	)

	searchURL := getSearchURL(args.ExternalURL, args.Query, utmSourceEmail)
	codeMonitorURL := getCodeMonitorURL(args.ExternalURL, email.Monitor, utmSourceEmail)

	if email.Priority == priorityCritical {
		priority = "Critical"
	} else {
		priority = "New"
	}

	if len(args.Results) == 1 {
		numberOfResultsWithDetail = fmt.Sprintf("There was %d new search result for your query", len(args.Results))
	} else {
		numberOfResultsWithDetail = fmt.Sprintf("There were %d new search results for your query", len(args.Results))
	}

	return &TemplateDataNewSearchResults{
		Priority:                  priority,
		CodeMonitorURL:            codeMonitorURL,
		SearchURL:                 searchURL,
		Description:               args.MonitorDescription,
		NumberOfResultsWithDetail: numberOfResultsWithDetail,
	}, nil
}

func NewTestTemplateDataForNewSearchResults(ctx context.Context, monitorDescription string) *TemplateDataNewSearchResults {
	return &TemplateDataNewSearchResults{
		Priority:                  "New",
		Description:               monitorDescription,
		NumberOfResultsWithDetail: "There was 1 new search result for your query",
		IsTest:                    true,
	}
}

func sendEmail(ctx context.Context, userID int32, template txtypes.Templates, data interface{}) error {
	email, err := internalapi.Client.UserEmailsGetEmail(ctx, userID)
	if err != nil {
		return errors.Errorf("internalapi.Client.UserEmailsGetEmail for userID=%d: %w", userID, err)
	}
	if email == nil {
		return errors.Errorf("unable to send email to user ID %d with unknown email address", userID)
	}
	if err := internalapi.Client.SendEmail(ctx, txtypes.Message{
		To:       []string{*email},
		Template: template,
		Data:     data,
	}); err != nil {
		return errors.Errorf("internalapi.Client.SendEmail to email=%q userID=%d: %w", *email, userID, err)
	}
	return nil
}

func getSearchURL(externalURL *url.URL, query, utmSource string) string {
	return sourcegraphURL(externalURL, "search", query, utmSource)
}

func getCodeMonitorURL(externalURL *url.URL, monitorID int64, utmSource string) string {
	return sourcegraphURL(externalURL, fmt.Sprintf("code-monitoring/%s", relay.MarshalID(MonitorKind, monitorID)), "", utmSource)
}

func getCommitURL(externalURL *url.URL, repoName, oid, utmSource string) string {
	return sourcegraphURL(externalURL, fmt.Sprintf("%s/-/commit/%s", repoName, oid), "", utmSource)
}

var (
	externalURLOnce  sync.Once
	externalURLValue *url.URL
	externalURLError error
)

func getExternalURL(ctx context.Context) (*url.URL, error) {
	if MockExternalURL != nil {
		return MockExternalURL(), nil
	}

	externalURLOnce.Do(func() {
		externalURLStr, err := internalapi.Client.ExternalURL(ctx)
		if err != nil {
			externalURLError = err
			return
		}
		externalURLValue, externalURLError = url.Parse(externalURLStr)
	})
	return externalURLValue, externalURLError
}

func sourcegraphURL(externalURL *url.URL, path, query, utmSource string) string {
	// Construct URL to the search query.
	u := externalURL.ResolveReference(&url.URL{Path: path})
	q := u.Query()
	if query != "" {
		q.Set("q", query)
	}
	q.Set("utm_source", utmSource)
	u.RawQuery = q.Encode()
	return u.String()
}
