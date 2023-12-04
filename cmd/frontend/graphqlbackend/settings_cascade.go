package graphqlbackend

import (
	"context"
	"encoding/json"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/settings"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

// settingsCascade implements the GraphQL type SettingsCascade (and the deprecated type ConfigurationCascade).
//
// It resolves settings from multiple sources.  When there is overlap between values, they will be
// merged in the following cascading order (first is lowest precedence):
//
// - Global site settings
// - Organization settings
// - Current user settings
type settingsCascade struct {
	db      database.DB
	subject *settingsSubjectResolver
}

func (r *settingsCascade) Subjects(ctx context.Context) ([]*settingsSubjectResolver, error) {
	subjects, err := settings.RelevantSubjects(ctx, r.db, r.subject.toSubject())
	if err != nil {
		return nil, err
	}

	return resolversForSubjects(ctx, log.Scoped("settings"), r.db, subjects)
}

func (r *settingsCascade) Final(ctx context.Context) (string, error) {
	settingsTyped, err := settings.Final(ctx, r.db, r.subject.toSubject())
	if err != nil {
		return "", err
	}

	settingsBytes, err := json.Marshal(settingsTyped)
	return string(settingsBytes), err
}

// Deprecated: in the GraphQL API
func (r *settingsCascade) Merged(ctx context.Context) (_ *configurationResolver, err error) {
	tr, ctx := trace.New(ctx, "SettingsCascade.Merged")
	defer tr.EndWithErr(&err)

	var messages []string
	s, err := r.Final(ctx)
	if err != nil {
		messages = append(messages, err.Error())
	}
	return &configurationResolver{contents: s, messages: messages}, nil
}

func (r *schemaResolver) ViewerSettings(ctx context.Context) (*settingsCascade, error) {
	user, err := CurrentUser(ctx, r.db)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return &settingsCascade{db: r.db, subject: &settingsSubjectResolver{site: NewSiteResolver(log.Scoped("settings"), r.db)}}, nil
	}
	return &settingsCascade{db: r.db, subject: &settingsSubjectResolver{user: user}}, nil
}

// Deprecated: in the GraphQL API
func (r *schemaResolver) ViewerConfiguration(ctx context.Context) (*settingsCascade, error) {
	return newSchemaResolver(r.db, r.gitserverClient).ViewerSettings(ctx)
}
