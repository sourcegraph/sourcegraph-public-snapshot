package graphqlbackend

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/cmd/query-runner/queryrunnerapi"
	"github.com/sourcegraph/sourcegraph/pkg/api"
)

type settingsResolver struct {
	subject  *configurationSubject
	settings *api.Settings
	user     *types.User
}

func (o *settingsResolver) ID() int32 {
	return o.settings.ID
}

func (o *settingsResolver) Subject() *configurationSubject {
	return o.subject
}

func (o *settingsResolver) Configuration() *configurationResolver {
	return &configurationResolver{contents: o.settings.Contents}
}

func (o *settingsResolver) Contents() string { return o.settings.Contents }

func (o *settingsResolver) CreatedAt() string {
	return o.settings.CreatedAt.Format(time.RFC3339) // ISO
}

func (o *settingsResolver) Author(ctx context.Context) (*UserResolver, error) {
	if o.user == nil {
		var err error
		o.user, err = db.Users.GetByID(ctx, o.settings.AuthorUserID)
		if err != nil {
			return nil, err
		}
	}
	return &UserResolver{o.user}, nil
}

// like db.Settings.CreateIfUpToDate, except it handles notifying the
// query-runner if any saved queries have changed.
func settingsCreateIfUpToDate(ctx context.Context, subject *configurationSubject, lastID *int32, authorUserID int32, contents string) (latestSetting *api.Settings, err error) {
	// Read current saved queries.
	var oldSavedQueries api.PartialConfigSavedQueries
	if err := subject.readConfiguration(ctx, &oldSavedQueries); err != nil {
		return nil, err
	}

	// Update settings.
	latestSettings, err := db.Settings.CreateIfUpToDate(ctx, subject.toSubject(), lastID, authorUserID, contents)
	if err != nil {
		return nil, err
	}

	// Read new saved queries.
	var newSavedQueries api.PartialConfigSavedQueries
	if err := subject.readConfiguration(ctx, &newSavedQueries); err != nil {
		return nil, err
	}

	// Notify query-runner of any changes.
	createdOrUpdated := false
	for i, newQuery := range newSavedQueries.SavedQueries {
		if i >= len(oldSavedQueries.SavedQueries) {
			// Created
			createdOrUpdated = true
			break
		}
		if !newQuery.Equals(oldSavedQueries.SavedQueries[i]) {
			// Updated or list was re-ordered.
			createdOrUpdated = true
			break
		}
	}
	if createdOrUpdated {
		go queryrunnerapi.Client.SavedQueryWasCreatedOrUpdated(context.Background(), subject.toSubject(), newSavedQueries, false)
	}
	for i, deletedQuery := range oldSavedQueries.SavedQueries {
		if i <= len(newSavedQueries.SavedQueries) {
			// Not deleted.
			continue
		}
		// Deleted
		spec := api.SavedQueryIDSpec{Subject: subject.toSubject(), Key: deletedQuery.Key}
		go queryrunnerapi.Client.SavedQueryWasDeleted(context.Background(), spec, false)
	}

	return latestSettings, nil
}
