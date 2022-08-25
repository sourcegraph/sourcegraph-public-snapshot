package repos

import (
	"context"
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var MockStatusMessages func(context.Context, *types.User) ([]StatusMessage, error)

// FetchStatusMessages fetches repo related status messages. When fetching
// external service sync errors we'll fetch any external services owned by the
// user. In addition, if the user is a site admin we'll also fetch site level
// external services.
func FetchStatusMessages(ctx context.Context, db database.DB, u *types.User) ([]StatusMessage, error) {
	if MockStatusMessages != nil {
		return MockStatusMessages(ctx, u)
	}
	if u == nil {
		return nil, errors.New("nil user")
	}
	var messages []StatusMessage

	// We first fetch affiliated sync errors since this will also find all the
	// external services the user cares about.
	externalServiceSyncErrors, err := db.ExternalServices().GetAffiliatedSyncErrors(ctx, u)
	if err != nil {
		return nil, errors.Wrap(err, "fetching sync errors")
	}

	extsvcIDs := make([]int64, 0, len(externalServiceSyncErrors))

	for id, failure := range externalServiceSyncErrors {
		extsvcIDs = append(extsvcIDs, id)

		if failure != "" {
			messages = append(messages, StatusMessage{
				ExternalServiceSyncError: &ExternalServiceSyncError{
					Message:           failure,
					ExternalServiceId: id,
				},
			})
		}
	}

	// Return early since the user doesn't have any affiliated external services
	if len(extsvcIDs) == 0 {
		return messages, nil
	}

	// Look for any repository that is not yet been cloned
	opts := database.ReposListOptions{
		NoCloned:           true,
		ExternalServiceIDs: extsvcIDs,
		LimitOffset: &database.LimitOffset{
			Limit: 1,
		},
	}
	notCloned, err := db.Repos().ListMinimalRepos(ctx, opts)
	if err != nil {
		return nil, errors.Wrap(err, "listing not-cloned repos")
	}
	if len(notCloned) > 0 {
		messages = append(messages, StatusMessage{
			Cloning: &CloningProgress{
				Message: fmt.Sprintf("%d %s cloning...", len(notCloned), pluralize(len(notCloned), "repository", "repositories")),
			},
		})
	}

	// Look for any repository that we could not sync
	opts = database.ReposListOptions{
		FailedFetch:        true,
		ExternalServiceIDs: extsvcIDs,
		LimitOffset: &database.LimitOffset{
			Limit: 1,
		},
	}
	failedSync, err := db.Repos().ListMinimalRepos(ctx, opts)
	if err != nil {
		return nil, errors.Wrap(err, "counting repo sync failures")
	}
	if len(failedSync) > 0 {
		messages = append(messages, StatusMessage{
			SyncError: &SyncError{
				Message: fmt.Sprintf("%d %s could not be synced", len(failedSync), pluralize(len(failedSync), "repository", "repositories")),
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

type IndexingProgress struct {
	Message string
}

type ExternalServiceSyncError struct {
	Message           string
	ExternalServiceId int64
}

type SyncError struct {
	Message string
}

type IndexingError struct {
	Message string
}

type StatusMessage struct {
	Cloning                  *CloningProgress          `json:"cloning"`
	Indexing                 *IndexingProgress         `json:"indexing"`
	ExternalServiceSyncError *ExternalServiceSyncError `json:"external_service_sync_error"`
	SyncError                *SyncError                `json:"sync_error"`
	IndexingError            *IndexingError            `json:"indexing_error"`
}
