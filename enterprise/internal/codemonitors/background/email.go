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
	"github.com/sourcegraph/sourcegraph/internal/search/result"
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
	Subject: `{{ if .IsTest }}Test: {{ end }}{{.Priority}}Sourcegraph code monitor {{.Description}} detected {{.TotalCount}} new {{.ResultPluralized}}`,
	Text:    textTemplate,
	HTML:    htmlTemplate,
})

type TemplateDataNewSearchResults struct {
	Priority                  string
	CodeMonitorURL            string
	SearchURL                 string
	Description               string
	IncludeResults            bool
	TruncatedResults          []*DisplayResult
	TotalCount                int
	TruncatedCount            int
	ResultPluralized          string
	TruncatedResultPluralized string
	DisplayMoreLink           bool
	IsTest                    bool
}

func NewTemplateDataForNewSearchResults(args actionArgs, email *edb.EmailAction) (d *TemplateDataNewSearchResults, err error) {
	var (
		priority string
	)

	searchURL := getSearchURL(args.ExternalURL, args.Query, utmSourceEmail)
	codeMonitorURL := getCodeMonitorURL(args.ExternalURL, email.Monitor, utmSourceEmail)

	if email.Priority == priorityCritical {
		priority = "[Critical] "
	} else {
		priority = ""
	}

	truncatedResults, totalCount, truncatedCount := truncateResults(args.Results, 5)

	displayResults := make([]*DisplayResult, len(truncatedResults))
	for i, result := range truncatedResults {
		displayResults[i] = toDisplayResult(result, args.ExternalURL)
	}

	return &TemplateDataNewSearchResults{
		Priority:                  priority,
		CodeMonitorURL:            codeMonitorURL,
		SearchURL:                 searchURL,
		Description:               args.MonitorDescription,
		IncludeResults:            args.IncludeResults,
		TruncatedResults:          displayResults,
		TotalCount:                totalCount,
		TruncatedCount:            truncatedCount,
		ResultPluralized:          pluralize("result", totalCount),
		TruncatedResultPluralized: pluralize("result", truncatedCount),
		DisplayMoreLink:           args.IncludeResults && truncatedCount > 0,
	}, nil
}

func NewTestTemplateDataForNewSearchResults(monitorDescription string) *TemplateDataNewSearchResults {
	return &TemplateDataNewSearchResults{
		IsTest:                    true,
		Priority:                  "",
		Description:               monitorDescription,
		TotalCount:                1,
		TruncatedCount:            0,
		ResultPluralized:          "result",
		TruncatedResultPluralized: "results",
		IncludeResults:            true,
		TruncatedResults: []*DisplayResult{{
			ResultType: "Test",
			RepoName:   "testorg/testrepo",
			CommitID:   "0000000",
			CommitURL:  "",
			Content:    "This is a test\nfor a code monitoring result.",
		}},
		DisplayMoreLink: false,
	}
}

func sendEmail(ctx context.Context, userID int32, template txtypes.Templates, data any) error {
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

// Only works for simple plurals (eg. result/results)
func pluralize(word string, count int) string {
	if count == 1 {
		return word
	}
	return word + "s"
}

type DisplayResult struct {
	ResultType string
	CommitURL  string
	RepoName   string
	CommitID   string
	Content    string
}

func toDisplayResult(result *result.CommitMatch, externalURL *url.URL) *DisplayResult {
	resultType := "Message"
	if result.DiffPreview != nil {
		resultType = "Diff"
	}

	var content string
	if result.DiffPreview != nil {
		content = truncateString(result.DiffPreview.Content, 10)
	} else {
		content = truncateString(result.MessagePreview.Content, 10)
	}

	return &DisplayResult{
		ResultType: resultType,
		CommitURL:  getCommitURL(externalURL, string(result.Repo.Name), string(result.Commit.ID), utmSourceEmail),
		RepoName:   string(result.Repo.Name),
		CommitID:   result.Commit.ID.Short(),
		Content:    content,
	}
}
