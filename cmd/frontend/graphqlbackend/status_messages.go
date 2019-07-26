package graphqlbackend

import (
	"context"

	"sourcegraph.com/cmd/frontend/backend"
	"sourcegraph.com/cmd/frontend/types"
	"sourcegraph.com/pkg/repoupdater"
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

	for _, rn := range result.Messages {
		messages = append(messages, &statusMessageResolver{&types.StatusMessage{
			Message: rn.Message,
			Type:    string(rn.Type),
		}})
	}

	return messages, nil
}

type statusMessageResolver struct {
	message *types.StatusMessage
}

func (n *statusMessageResolver) Type() string    { return n.message.Type }
func (n *statusMessageResolver) Message() string { return n.message.Message }
