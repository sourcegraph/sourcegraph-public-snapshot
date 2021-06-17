package discovery

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/schema"
)

// SettingStore is a subset of the API exposed by the database.Settings() store.
type SettingStore interface {
	GetLatest(context.Context, api.SettingsSubject) (*api.Settings, error)
	GetLastestSchemaSettings(context.Context, api.SettingsSubject) (*schema.Settings, error)
}

// Discover uses the given settings store to look for insights in the global user settings.
//
// TODO(slimsag): future: include user/org settings and consider security implications of doing so.
// In the future, this will be expanded to also include insights from users/orgs.
func Discover(ctx context.Context, settingStore SettingStore) ([]*schema.Insight, error) {
	// Get latest Global user settings.
	subject := api.SettingsSubject{Site: true}
	globalSettings, err := settingStore.GetLastestSchemaSettings(ctx, subject)
	if err != nil {
		return nil, err
	}
	return globalSettings.Insights, nil
}
