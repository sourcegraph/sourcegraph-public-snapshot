package graphqlbackend

import (
	"context"

	"github.com/cockroachdb/errors"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	ts "github.com/sourcegraph/sourcegraph/internal/temporarysettings"
)

type TemporarySettingsResolver struct {
	db    dbutil.DB
	inner *ts.TemporarySettings
}

func (r *schemaResolver) TemporarySettings(ctx context.Context) (*TemporarySettingsResolver, error) {
	a := actor.FromContext(ctx)
	if !a.IsAuthenticated() {
		return nil, errors.New("not authenticated")
	}

	temporarySettings, err := database.TemporarySettings(r.db).GetTemporarySettings(ctx, a.UID)
	if err != nil {
		return nil, err
	}
	return &TemporarySettingsResolver{db: r.db, inner: temporarySettings}, nil
}

func (t *TemporarySettingsResolver) Contents() string {
	return t.inner.Contents
}

func (r *schemaResolver) OverwriteTemporarySettings(ctx context.Context, args struct{ Contents string }) (*EmptyResponse, error) {
	a := actor.FromContext(ctx)
	if !a.IsAuthenticated() {
		return nil, errors.New("not authenticated")
	}

	return &EmptyResponse{}, database.TemporarySettings(r.db).UpsertTemporarySettings(ctx, a.UID, args.Contents)
}
