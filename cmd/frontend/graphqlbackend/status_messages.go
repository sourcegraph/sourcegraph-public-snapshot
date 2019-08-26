package graphqlbackend

import (
	"context"
	"errors"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/pkg/repoupdater"
	"github.com/sourcegraph/sourcegraph/pkg/repoupdater/protocol"
)

func (r *schemaResolver) StatusMessages(ctx context.Context) ([]*statusMessageResolver, error) {
	var messages []*statusMessageResolver

	// 🚨 SECURITY: Only site admins can see status messages.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	result, err := repoupdater.DefaultClient.StatusMessages(ctx)
	if err != nil {
		return nil, err
	}

	for _, m := range result.Messages {
		messages = append(messages, &statusMessageResolver{message: m})
	}

	return messages, nil
}

type statusMessageResolver struct {
	message protocol.StatusMessage
}

func (r *statusMessageResolver) ToCloningStatusMessage() (*statusMessageResolver, bool) {
	return r, r.message.Cloning != nil
}

func (r *statusMessageResolver) ToSyncErrorStatusMessage() (*statusMessageResolver, bool) {
	return r, r.message.SyncError != nil
}

func (r *statusMessageResolver) Message() (string, error) {
	if r.message.Cloning != nil {
		return r.message.Cloning.Message, nil
	}
	if r.message.SyncError != nil {
		return r.message.SyncError.Message, nil
	}
	return "", errors.New("status message is of unknown type")
}

func (r *statusMessageResolver) ExternalService(ctx context.Context) (*externalServiceResolver, error) {
	externalService, err := db.ExternalServices.GetByID(ctx, r.message.SyncError.ExternalServiceId)
	if err != nil {
		return nil, err
	}

	return &externalServiceResolver{externalService: externalService}, nil
}
