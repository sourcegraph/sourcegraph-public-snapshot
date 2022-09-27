package repos

import (
	"context"
	"fmt"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var MockStatusMessages func(context.Context) ([]StatusMessage, error)

// FetchStatusMessages fetches repo related status messages.
func FetchStatusMessages(ctx context.Context, db database.DB) ([]StatusMessage, error) {
	if MockStatusMessages != nil {
		return MockStatusMessages(ctx)
	}
	var messages []StatusMessage

	// We first fetch affiliated sync errors since this will also find all the
	// external services the user cares about.
	externalServiceSyncErrors, err := db.ExternalServices().GetLatestSyncErrors(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "fetching sync errors")
	}
	// Return early since we don't have any affiliated external services
	if len(externalServiceSyncErrors) == 0 {
		return messages, nil
	}

	for id, failure := range externalServiceSyncErrors {
		if failure != "" {
			messages = append(messages, StatusMessage{
				ExternalServiceSyncError: &ExternalServiceSyncError{
					Message:           failure,
					ExternalServiceId: id,
				},
			})
		}
	}

	stats, err := db.RepoStatistics().GetRepoStatistics(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "loading repo statistics")
	}

	if stats.FailedFetch > 0 {
		messages = append(messages, StatusMessage{
			SyncError: &SyncError{
				Message: fmt.Sprintf("%d %s failed last attempt to sync content from code host", stats.FailedFetch, pluralize(stats.FailedFetch, "repository", "repositories")),
			},
		})
	}

	if uncloned := stats.NotCloned + stats.Cloning; uncloned > 0 {
		var sentences []string
		if stats.NotCloned > 0 {
			sentences = append(sentences, fmt.Sprintf("%d %s enqueued for cloning.", stats.NotCloned, pluralize(stats.NotCloned, "repository", "repositories")))
		}
		if stats.Cloning > 0 {
			sentences = append(sentences, fmt.Sprintf("%d %s currently cloning...", stats.Cloning, pluralize(stats.Cloning, "repository", "repositories")))
		}
		messages = append(messages, StatusMessage{
			Cloning: &CloningProgress{
				Message: strings.Join(sentences, " "),
			},
		})
	}

	return messages, nil
}

func pluralize(count int, singularNoun, pluralNoun string) string {
	if count == 1 {
		return singularNoun
	}
	return pluralNoun
}

type CloningProgress struct {
	Message string
}

type ExternalServiceSyncError struct {
	Message           string
	ExternalServiceId int64
}

type SyncError struct {
	Message string
}

type StatusMessage struct {
	Cloning                  *CloningProgress          `json:"cloning"`
	ExternalServiceSyncError *ExternalServiceSyncError `json:"external_service_sync_error"`
	SyncError                *SyncError                `json:"sync_error"`
}
