package graphqlbackend

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/repoupdater"
)

func (r *schemaResolver) StatusMessages(ctx context.Context) ([]*StatusMessageResolver, error) {
	var notifications []*StatusMessageResolver

	// 🚨 SECURITY: Only site admins can see status messages.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	result, err := repoupdater.DefaultClient.StatusMessages(context.Background())
	if err != nil {
		return nil, err
	}

	for _, rn := range result.Messages {
		notifications = append(notifications, NewStatusMessage(&types.StatusMessage{
			Message: rn.Message,
			Type:    string(rn.Type),
		}))
	}

	return notifications, nil
}

type StatusMessageResolver struct {
	message *types.StatusMessage
}

func NewStatusMessage(message *types.StatusMessage) *StatusMessageResolver {
	return &StatusMessageResolver{message: message}
}

func (n *StatusMessageResolver) Type() string {
	return n.message.Type
}

func (n *StatusMessageResolver) Message() string {
	return n.message.Message
}
