package graphqlbackend

import (
	"context"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/pkg/repoupdater"
	"github.com/sourcegraph/sourcegraph/pkg/repoupdater/protocol"
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
		if m.Type == protocol.Cloning && m.Cloning != nil {
			messages = append(messages, &cloningStatusMessageResolver{
				message: m.Cloning.Message,
			})
		}

		if m.Type == protocol.SyncError && m.SyncError != nil {
			messages = append(messages, &syncErrorStatusMessageResolver{
				message:                    m.SyncError.Message,
				externalServiceId:          marshalExternalServiceID(m.SyncError.ExternalServiceId),
				externalServiceDisplayName: m.SyncError.ExternalServiceDisplayName,
				externalServiceKind:        m.SyncError.ExternalServiceKind,
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
	message                    string
	externalServiceId          graphql.ID
	externalServiceDisplayName string
	externalServiceKind        string
}

func (n *syncErrorStatusMessageResolver) Message() string { return n.message }
func (n *syncErrorStatusMessageResolver) ToCloningStatusMessage() (*cloningStatusMessageResolver, bool) {
	return nil, false
}
func (n *syncErrorStatusMessageResolver) ToSyncErrorStatusMessage() (*syncErrorStatusMessageResolver, bool) {
	return n, true
}
func (n *syncErrorStatusMessageResolver) ExternalServiceId() string {
	return string(n.externalServiceId)
}
func (n *syncErrorStatusMessageResolver) ExternalServiceDisplayName() string {
	return n.externalServiceDisplayName
}
func (n *syncErrorStatusMessageResolver) ExternalServiceKind() string {
	return n.externalServiceKind
}
