package graphqlbackend

import (
	"context"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/repos"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func (r *schemaResolver) StatusMessages(ctx context.Context) ([]*statusMessageResolver, error) {
	// 🚨 SECURITY: Only site admins can fetch status messages.
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	messages, err := repos.FetchStatusMessages(ctx, r.db, r.gitserverClient)
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

func (r *statusMessageResolver) ToGitUpdatesDisabled() (*statusMessageResolver, bool) {
	return r, r.message.GitUpdatesDisabled != nil
}

func (r *statusMessageResolver) ToNoRepositoriesDetected() (*statusMessageResolver, bool) {
	return r, r.message.NoRepositoriesDetected != nil
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

func (r *statusMessageResolver) ToIndexingProgress() (*indexingProgressMessageResolver, bool) {
	if r.message.Indexing != nil {
		return &indexingProgressMessageResolver{message: r.message.Indexing}, true
	}
	return nil, false
}

func (r *statusMessageResolver) ToGitserverDiskThresholdReached() (*statusMessageResolver, bool) {
	return r, r.message.GitserverDiskThresholdReached != nil
}

func (r *statusMessageResolver) Message() (string, error) {
	if r.message.GitUpdatesDisabled != nil {
		return r.message.GitUpdatesDisabled.Message, nil
	}
	if r.message.NoRepositoriesDetected != nil {
		return r.message.NoRepositoriesDetected.Message, nil
	}
	if r.message.Cloning != nil {
		return r.message.Cloning.Message, nil
	}
	if r.message.ExternalServiceSyncError != nil {
		return r.message.ExternalServiceSyncError.Message, nil
	}
	if r.message.SyncError != nil {
		return r.message.SyncError.Message, nil
	}
	if r.message.GitserverDiskThresholdReached != nil {
		return r.message.GitserverDiskThresholdReached.Message, nil
	}
	return "", errors.New("status message is of unknown type")
}

func (r *statusMessageResolver) ExternalService(ctx context.Context) (*externalServiceResolver, error) {
	id := r.message.ExternalServiceSyncError.ExternalServiceId
	externalService, err := r.db.ExternalServices().GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &externalServiceResolver{logger: log.Scoped("externalServiceResolver"), db: r.db, externalService: externalService}, nil
}

type indexingProgressMessageResolver struct {
	message *repos.IndexingProgress
}

func (r *indexingProgressMessageResolver) NotIndexed() int32 { return int32(r.message.NotIndexed) }
func (r *indexingProgressMessageResolver) Indexed() int32    { return int32(r.message.Indexed) }
