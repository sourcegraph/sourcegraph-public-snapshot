package repos

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

var MockStatusMessages func(context.Context) ([]StatusMessage, error)

func FetchStatusMessages(ctx context.Context, db dbutil.DB) ([]StatusMessage, error) {
	if MockStatusMessages != nil {
		return MockStatusMessages(ctx)
	}

	var messages []StatusMessage

	notCloned, err := database.Repos(db).Count(ctx, database.ReposListOptions{NoCloned: true})
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

	syncErrors, err := database.ExternalServices(db).ListSyncErrors(ctx)
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
