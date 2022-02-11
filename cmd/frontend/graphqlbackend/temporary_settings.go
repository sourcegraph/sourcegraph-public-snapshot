package graphqlbackend

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	ts "github.com/sourcegraph/sourcegraph/internal/temporarysettings"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type TemporarySettingsResolver struct {
	db    database.DB
	inner *ts.TemporarySettings
}

func (r *schemaResolver) TemporarySettings(ctx context.Context) (*TemporarySettingsResolver, error) {
	a := actor.FromContext(ctx)
	if !a.IsAuthenticated() {
		return nil, errors.New("not authenticated")
	}

	temporarySettings, err := r.db.TemporarySettings().GetTemporarySettings(ctx, a.UID)
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

	return &EmptyResponse{}, r.db.TemporarySettings().OverwriteTemporarySettings(ctx, a.UID, args.Contents)
}

func (r *schemaResolver) EditTemporarySettings(ctx context.Context, args struct{ SettingsToEdit string }) (*EmptyResponse, error) {
	a := actor.FromContext(ctx)
	if !a.IsAuthenticated() {
		return nil, errors.New("not authenticated")
	}

	return &EmptyResponse{}, r.db.TemporarySettings().EditTemporarySettings(ctx, a.UID, args.SettingsToEdit)
}
