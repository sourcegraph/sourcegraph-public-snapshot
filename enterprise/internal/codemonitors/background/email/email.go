package email

import (
	"context"
	"fmt"
	"net/url"

	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codemonitors"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/txemail/txtypes"
)

var externalURL *url.URL

const utmSourceEmail = "code-monitoring-email"
const priorityCritical = "CRITICAL"

func SendEmailForNewSearchResult(ctx context.Context, userID int32, data *TemplateDataNewSearchResults) error {
	return sendEmail(ctx, userID, newSearchResultsEmailTemplates, data)
}

func canSendEmail(ctx context.Context) error {
	canSendEmail, err := api.InternalClient.CanSendEmail(ctx)
	if err != nil {
		return fmt.Errorf("InternalClient.CanSendEmail: %w", err)
	}
	if !canSendEmail {
		return fmt.Errorf("SMTP server not set in site configuration")
	}
	return nil
}

type TemplateDataNewSearchResults struct {
	Priority                  string
	CodeMonitorURL            string
	SearchURL                 string
	Description               string
	NumberOfResultsWithDetail string
}

func NewTemplateDataForNewSearchResults(monitorDescription string, queryString string, email *codemonitors.MonitorEmail, numResults int) (d *TemplateDataNewSearchResults, err error) {
	var (
		searchURL                 string
		codeMonitorURL            string
		priority                  string
		numberOfResultsWithDetail string
	)
	searchURL, err = getSearchURL(queryString, utmSourceEmail)
	if err != nil {
		return nil, err
	}

	codeMonitorURL, err = getCodeMonitorURL(email.Monitor, utmSourceEmail)
	if err != nil {
		return nil, err
	}

	if email.Priority == priorityCritical {
		priority = "Critical"
	} else {
		priority = "New"
	}

	if numResults == 1 {
		numberOfResultsWithDetail = fmt.Sprintf("There was %d new search result for your query", numResults)
	} else {
		numberOfResultsWithDetail = fmt.Sprintf("There were %d new search results for your query", numResults)
	}

	return &TemplateDataNewSearchResults{
		Priority:                  priority,
		CodeMonitorURL:            codeMonitorURL,
		SearchURL:                 searchURL,
		Description:               monitorDescription,
		NumberOfResultsWithDetail: numberOfResultsWithDetail,
	}, nil

}

func sendEmail(ctx context.Context, userID int32, template txtypes.Templates, data interface{}) error {
	err := canSendEmail(ctx)
	if err != nil {
		return err
	}

	email, err := api.InternalClient.UserEmailsGetEmail(ctx, userID)
	if err != nil {
		return fmt.Errorf("InternalClient.UserEmailsGetEmail for userID=%d: %w", userID, err)
	}
	if email == nil {
		return fmt.Errorf("unable to send email to user ID %d with unknown email address", userID)
	}

	if err := api.InternalClient.SendEmail(ctx, txtypes.Message{
		To:       []string{*email},
		Template: template,
		Data:     data,
	}); err != nil {
		return fmt.Errorf("InternalClient.SendEmail to email=%q userID=%d: %w", *email, userID, err)
	}
	return nil
}

func getSearchURL(query, utmSource string) (string, error) {
	return sourcegraphURL("search", query, utmSource)
}

func getCodeMonitorURL(monitorID int64, utmSource string) (string, error) {
	return sourcegraphURL(fmt.Sprintf("code-monitoring/%s", relay.MarshalID("FIXME", monitorID)), "", utmSource)
}

func sourcegraphURL(path, query, utmSource string) (string, error) {
	if externalURL == nil {
		// Determine the external URL.
		externalURLStr, err := api.InternalClient.ExternalURL(context.Background())
		if err != nil {
			return "", fmt.Errorf("failed to get ExternalURL: %w", err)
		}
		externalURL, err = url.Parse(externalURLStr)
		if err != nil {

			return "", fmt.Errorf("failed to get ExternalURL: %w", err)
		}
	}

	// Construct URL to the search query.
	u := externalURL.ResolveReference(&url.URL{Path: path})
	q := u.Query()
	if query != "" {
		q.Set("q", query)
	}
	q.Set("utm_source", utmSource)
	u.RawQuery = q.Encode()
	return u.String(), nil
}
