package graphqlbackend

import (
	"context"
	"fmt"
	"strconv"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
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

	for _, rn := range result.Messages {
		fmt.Printf("rn=%+v\n", rn)
		metadata := make([]types.StatusMessageMetadata, len(rn.Metadata))

		for i, m := range rn.Metadata {
			metadata[i] = types.StatusMessageMetadata{Name: m.Name, Value: m.Value}
		}

		messages = append(messages, &statusMessageResolver{&types.StatusMessage{
			Message:  rn.Message,
			Type:     string(rn.Type),
			Metadata: metadata,
		}})
	}

	return messages, nil
}

type statusMessageResolver struct {
	message *types.StatusMessage
}

func (n *statusMessageResolver) Type() string    { return n.message.Type }
func (n *statusMessageResolver) Message() string { return n.message.Message }
func (n *statusMessageResolver) Metadata() []*statusMessageMetadataResolver {
	var resolvers []*statusMessageMetadataResolver

	for _, m := range n.message.Metadata {
		resolvers = append(resolvers, &statusMessageMetadataResolver{
			messageType: protocol.StatusMessageType(n.message.Type),
			metadata:    m,
		})
	}

	return resolvers
}

type statusMessageMetadataResolver struct {
	messageType protocol.StatusMessageType
	metadata    types.StatusMessageMetadata
}

func (n *statusMessageMetadataResolver) Name() string { return n.metadata.Name }
func (n *statusMessageMetadataResolver) Value() (string, error) {
	if n.messageType == protocol.SyncingErrorMessage && n.Name() == "ext_svc_id" {
		id, err := strconv.ParseInt(n.metadata.Value, 10, 64)
		if err != nil {
			return "", err
		}

		graphqlID := marshalExternalServiceID(id)
		return string(graphqlID), nil
	}

	return n.metadata.Value, nil
}
