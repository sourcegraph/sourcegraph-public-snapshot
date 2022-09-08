package graphqlbackend

import (
	"context"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/repos"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func (r *schemaResolver) StatusMessages(ctx context.Context) ([]*statusMessageResolver, error) {
	// 🚨 SECURITY: Only site admins can fetch status messages.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	messages, err := repos.FetchStatusMessages(ctx, r.db)
	if err != nil {
		return nil, err
	}

	var messageResolvers []*statusMessageResolver
	for _, m := range messages {
		messageResolvers = append(messageResolvers, &statusMessageResolver{db: r.db, message: m})
	}

	return messageResolvers, nil
}

type statusMessageResolver struct {
	message repos.StatusMessage
	db      database.DB
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
	externalService, err := r.db.ExternalServices().GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &externalServiceResolver{logger: log.Scoped("externalServiceResolver", ""), db: r.db, externalService: externalService}, nil
}
