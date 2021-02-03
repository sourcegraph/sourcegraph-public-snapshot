package graphqlbackend

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater/protocol"
)

func (r *schemaResolver) StatusMessages(ctx context.Context) ([]*statusMessageResolver, error) {
	var messages []*statusMessageResolver

	// 🚨 SECURITY: Only site admins can see status messages.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	result, err := fetchStatusMessages(ctx, r.db)
	if err != nil {
		return nil, err
	}

	for _, m := range result.Messages {
		messages = append(messages, &statusMessageResolver{message: m})
	}

	return messages, nil
}

var MockStatusMessages func(context.Context) (*protocol.StatusMessagesResponse, error)

// TODO: Move this somewhere better
// TODO: No need for protocol.StatusMessageResponse
func fetchStatusMessages(ctx context.Context, db dbutil.DB) (*protocol.StatusMessagesResponse, error) {
	if MockStatusMessages != nil {
		return MockStatusMessages(ctx)
	}

	resp := protocol.StatusMessagesResponse{
		Messages: []protocol.StatusMessage{},
	}

	notCloned, err := database.Repos(db).Count(ctx, database.ReposListOptions{NoCloned: true})
	if err != nil {
		return nil, errors.Wrap(err, "counting uncloned repos")
	}

	if notCloned != 0 {
		resp.Messages = append(resp.Messages, protocol.StatusMessage{
			Cloning: &protocol.CloningProgress{
				Message: fmt.Sprintf("%d repositories enqueued for cloning...", notCloned),
			},
		})
	}

	syncErrors, err := database.ExternalServices(db).ListSyncErrors(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "fetching sync errors")
	}

	for id, failure := range syncErrors {
		resp.Messages = append(resp.Messages, protocol.StatusMessage{
			ExternalServiceSyncError: &protocol.ExternalServiceSyncError{
				Message:           failure,
				ExternalServiceId: id,
			},
		})
	}

	return &resp, nil
}

type statusMessageResolver struct {
	message protocol.StatusMessage
}

func (r *statusMessageResolver) ToCloningProgress() (*statusMessageResolver, bool) {
	return r, r.message.Cloning != nil
}

func (r *statusMessageResolver) ToExternalServiceSyncError() (*statusMessageResolver, bool) {
	return r, r.message.ExternalServiceSyncError != nil
}

func (r *statusMessageResolver) ToSyncError() (*statusMessageResolver, bool) {
	return r, r.message.SyncError != nil
}

func (r *statusMessageResolver) Message() (string, error) {
	if r.message.Cloning != nil {
		return r.message.Cloning.Message, nil
	}
	if r.message.ExternalServiceSyncError != nil {
		return r.message.ExternalServiceSyncError.Message, nil
	}
	if r.message.SyncError != nil {
		return r.message.SyncError.Message, nil
	}
	return "", errors.New("status message is of unknown type")
}

func (r *statusMessageResolver) ExternalService(ctx context.Context) (*externalServiceResolver, error) {
	id := r.message.ExternalServiceSyncError.ExternalServiceId
	externalService, err := database.GlobalExternalServices.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &externalServiceResolver{externalService: externalService}, nil
}
