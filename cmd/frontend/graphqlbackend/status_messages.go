package graphqlbackend

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/pkg/repoupdater"
)

func (r *schemaResolver) StatusMessages(ctx context.Context) ([]statusMessageResolver, error) {
	var messages []statusMessageResolver

	// 🚨 SECURITY: Only site admins can see status messages.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	result, err := repoupdater.DefaultClient.StatusMessages(ctx)
	if err != nil {
		return nil, err
	}

	for _, m := range result.Messages {
		if m.Cloning != nil {
			messages = append(messages, &cloningStatusMessageResolver{
				message: m.Cloning.Message,
			})
		}

		if m.SyncError != nil {
			messages = append(messages, &syncErrorStatusMessageResolver{
				message:           m.SyncError.Message,
				externalServiceId: m.SyncError.ExternalServiceId,
			})
		}
	}

	return messages, nil
}

type statusMessageResolver interface {
	ToCloningStatusMessage() (*cloningStatusMessageResolver, bool)
	ToSyncErrorStatusMessage() (*syncErrorStatusMessageResolver, bool)
}

type cloningStatusMessageResolver struct {
	message string
}

func (n *cloningStatusMessageResolver) Message() string { return n.message }
func (n *cloningStatusMessageResolver) ToCloningStatusMessage() (*cloningStatusMessageResolver, bool) {
	return n, true
}
func (n *cloningStatusMessageResolver) ToSyncErrorStatusMessage() (*syncErrorStatusMessageResolver, bool) {
	return nil, false
}

type syncErrorStatusMessageResolver struct {
	message           string
	externalServiceId int64
}

func (n *syncErrorStatusMessageResolver) Message() string { return n.message }
func (n *syncErrorStatusMessageResolver) ToCloningStatusMessage() (*cloningStatusMessageResolver, bool) {
	return nil, false
}
func (n *syncErrorStatusMessageResolver) ToSyncErrorStatusMessage() (*syncErrorStatusMessageResolver, bool) {
	return n, true
}
func (n *syncErrorStatusMessageResolver) ExternalService(ctx context.Context) (*externalServiceResolver, error) {
	externalService, err := db.ExternalServices.GetByID(ctx, n.externalServiceId)
	if err != nil {
		return nil, err
	}

	return &externalServiceResolver{externalService: externalService}, nil
}
