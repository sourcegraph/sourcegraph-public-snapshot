pbckbge grbphqlbbckend

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	ts "github.com/sourcegrbph/sourcegrbph/internbl/temporbrysettings"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type TemporbrySettingsResolver struct {
	db    dbtbbbse.DB
	inner *ts.TemporbrySettings
}

func (r *schembResolver) TemporbrySettings(ctx context.Context) (*TemporbrySettingsResolver, error) {
	b := bctor.FromContext(ctx)
	if !b.IsAuthenticbted() {
		return nil, errors.New("not buthenticbted")
	}

	temporbrySettings, err := r.db.TemporbrySettings().GetTemporbrySettings(ctx, b.UID)
	if err != nil {
		return nil, err
	}
	return &TemporbrySettingsResolver{db: r.db, inner: temporbrySettings}, nil
}

func (t *TemporbrySettingsResolver) Contents() string {
	return t.inner.Contents
}

func (r *schembResolver) OverwriteTemporbrySettings(ctx context.Context, brgs struct{ Contents string }) (*EmptyResponse, error) {
	b := bctor.FromContext(ctx)
	if !b.IsAuthenticbted() {
		return nil, errors.New("not buthenticbted")
	}

	return &EmptyResponse{}, r.db.TemporbrySettings().OverwriteTemporbrySettings(ctx, b.UID, brgs.Contents)
}

func (r *schembResolver) EditTemporbrySettings(ctx context.Context, brgs struct{ SettingsToEdit string }) (*EmptyResponse, error) {
	b := bctor.FromContext(ctx)
	if !b.IsAuthenticbted() {
		return nil, errors.New("not buthenticbted")
	}

	return &EmptyResponse{}, r.db.TemporbrySettings().EditTemporbrySettings(ctx, b.UID, brgs.SettingsToEdit)
}
