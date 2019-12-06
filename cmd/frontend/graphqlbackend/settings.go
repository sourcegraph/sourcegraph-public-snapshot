package graphqlbackend

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/internal/api"
)

type settingsResolver struct {
	subject  *settingsSubject
	settings *api.Settings
	user     *types.User
}

func (o *settingsResolver) ID() int32 {
	return o.settings.ID
}

func (o *settingsResolver) Subject() *settingsSubject {
	return o.subject
}

// Deprecated: Use the Contents field instead.
func (o *settingsResolver) Configuration() *configurationResolver {
	return &configurationResolver{contents: o.settings.Contents}
}

func (o *settingsResolver) Contents() JSONCString {
	return JSONCString(o.settings.Contents)
}

func (o *settingsResolver) CreatedAt() DateTime {
	return DateTime{Time: o.settings.CreatedAt}
}

func (o *settingsResolver) Author(ctx context.Context) (*UserResolver, error) {
	if o.settings.AuthorUserID == nil {
		return nil, nil
	}
	if o.user == nil {
		var err error
		o.user, err = db.Users.GetByID(ctx, *o.settings.AuthorUserID)
		if err != nil {
			return nil, err
		}
	}
	return &UserResolver{o.user}, nil
}

// like db.Settings.CreateIfUpToDate, except it handles notifying the
// query-runner if any saved queries have changed.
func settingsCreateIfUpToDate(ctx context.Context, subject *settingsSubject, lastID *int32, authorUserID int32, contents string) (latestSetting *api.Settings, err error) {
	// Read current saved queries.
	var oldSavedQueries api.PartialConfigSavedQueries
	if err := subject.readSettings(ctx, &oldSavedQueries); err != nil {
		return nil, err
	}

	// Update settings.
	latestSettings, err := db.Settings.CreateIfUpToDate(ctx, subject.toSubject(), lastID, &authorUserID, contents)
	if err != nil {
		return nil, err
	}

	return latestSettings, nil
}
