package repos

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

var MockStatusMessages func(context.Context, *types.User) ([]StatusMessage, error)

// FetchStatusMessages fetches repo related status messages. When fetching
// external service sync errors we'll fetch any external services owned by the
// user. In addition, if the user is a site admin we'll also fetch site level
// external services.
func FetchStatusMessages(ctx context.Context, db dbutil.DB, u *types.User, cloud bool) ([]StatusMessage, error) {
	if MockStatusMessages != nil {
		return MockStatusMessages(ctx, u)
	}
	if u == nil {
		return nil, errors.New("nil user")
	}
	var messages []StatusMessage

	// We first fetch affiliated sync errors since this will also find all the
	// external services the user cares about.
	externalServiceSyncErrors, err := database.ExternalServices(db).GetAffiliatedSyncErrors(ctx, u)
	if err != nil {
		return nil, errors.Wrap(err, "fetching sync errors")
	}

	for id, failure := range externalServiceSyncErrors {
		if failure == "" {
			continue
		}
		messages = append(messages, StatusMessage{
			ExternalServiceSyncError: &ExternalServiceSyncError{
				Message:           failure,
				ExternalServiceId: id,
			},
		})
	}

	extsvcIDs := make([]int64, 0, len(externalServiceSyncErrors))
	for id := range externalServiceSyncErrors {
		extsvcIDs = append(extsvcIDs, id)
	}

	// Return early since the user doesn't have any affiliated external services
	if len(extsvcIDs) == 0 {
		return messages, nil
	}

	// Now, for all the affiliated external services, look for any repos they own
	// that have not yet been cloned
	opts := database.ReposListOptions{
		NoCloned:           true,
		ExternalServiceIDs: extsvcIDs,
	}
	notCloned, err := database.Repos(db).Count(ctx, opts)
	if err != nil {
		return nil, errors.Wrap(err, "counting uncloned repos")
	}
	if notCloned > 0 {
		noun := "repositories"
		if notCloned == 1 {
			noun = "repository"
		}
		messages = append(messages, StatusMessage{
			Cloning: &CloningProgress{
				Message: fmt.Sprintf("%d %s not yet cloned", notCloned, noun),
			},
		})
	}

	return messages, nil
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

type StatusMessage struct {
	Cloning                  *CloningProgress          `json:"cloning"`
	Indexing                 *IndexingProgress         `json:"indexing"`
	ExternalServiceSyncError *ExternalServiceSyncError `json:"external_service_sync_error"`
	SyncError                *SyncError                `json:"sync_error"`
}
