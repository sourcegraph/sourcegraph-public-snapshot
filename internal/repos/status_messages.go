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
	opts := database.ReposListOptions{
		NoCloned: true,
	}
	if !u.SiteAdmin {
		opts.UserID = u.ID
	}

	if !cloud {
		// The number of uncloned repos on cloud is misleading due to the fact that we do
		// on demand syncing and also remove stale repos.
		notCloned, err := database.Repos(db).Count(ctx, opts)
		if err != nil {
			return nil, errors.Wrap(err, "counting uncloned repos")
		}

		if notCloned != 0 {
			messages = append(messages, StatusMessage{
				Cloning: &CloningProgress{
					Message: fmt.Sprintf("%d repositories enqueued for cloning...", notCloned),
				},
			})
		}
	}

	syncErrors, err := database.ExternalServices(db).GetAffiliatedSyncErrors(ctx, u)
	if err != nil {
		return nil, errors.Wrap(err, "fetching sync errors")
	}

	for id, failure := range syncErrors {
		messages = append(messages, StatusMessage{
			ExternalServiceSyncError: &ExternalServiceSyncError{
				Message:           failure,
				ExternalServiceId: id,
			},
		})
	}

	return messages, nil
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
