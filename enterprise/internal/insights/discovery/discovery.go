package discovery

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"
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
	globalSettingsRaw, err := settingStore.GetLatest(ctx, subject)
	if err != nil {
		return nil, err
	}
	globalSettings, err := parseUserSettings(globalSettingsRaw)
	if err != nil {
		return nil, err
	}
	return globalSettings.Insights, nil
}

func parseUserSettings(settings *api.Settings) (*schema.Settings, error) {
	if settings == nil {
		// Settings have never been saved for this subject; equivalent to `{}`.
		return &schema.Settings{}, nil
	}
	var v schema.Settings
	if err := jsonc.Unmarshal(settings.Contents, &v); err != nil {
		return nil, err
	}
	return &v, nil
}
