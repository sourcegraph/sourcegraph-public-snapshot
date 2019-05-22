package graphqlbackend

import (
	"context"
	"errors"
	"fmt"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/jsonc"
)

func (schemaResolver) SettingsSubject(ctx context.Context, args *struct{ ID graphql.ID }) (*settingsSubject, error) {
	return settingsSubjectByID(ctx, args.ID)
}

var errUnknownSettingsSubject = errors.New("unknown settings subject")

type settingsSubject struct {
	// Exactly 1 of these fields must be set.
	defaultSettings *defaultSettingsResolver
	site            *siteResolver
	org             *OrgResolver
	user            *UserResolver
}

// settingsSubjectByID fetches the settings subject with the given ID. If the ID refers to a node
// that is not a valid settings subject, an error is returned.
func settingsSubjectByID(ctx context.Context, id graphql.ID) (*settingsSubject, error) {
	resolver, err := NodeByID(ctx, id)
	if err != nil {
		return nil, err
	}

	switch s := resolver.(type) {
	case *siteResolver:
		return &settingsSubject{site: s}, nil

	case *UserResolver:
		// 🚨 SECURITY: Only the user and site admins are allowed to view the user's settings.
		if err := backend.CheckSiteAdminOrSameUser(ctx, s.user.ID); err != nil {
			return nil, err
		}
		return &settingsSubject{user: s}, nil

	case *OrgResolver:
		// 🚨 SECURITY: Check that the current user is a member of the org.
		if err := backend.CheckOrgAccess(ctx, s.org.ID); err != nil {
			return nil, err
		}
		return &settingsSubject{org: s}, nil

	default:
		return nil, errUnknownSettingsSubject
	}
}

func settingsSubjectID(subject api.SettingsSubject) (graphql.ID, error) {
	switch {
	case subject.Site:
		return marshalSiteGQLID(singletonSiteResolver.gqlID), nil
	case subject.User != nil:
		return marshalUserID(*subject.User), nil
	case subject.Org != nil:
		return marshalOrgID(*subject.Org), nil
	default:
		return "", errUnknownSettingsSubject
	}
}

func settingsSubjectsEqual(a, b api.SettingsSubject) bool {
	switch {
	case a.Site || b.Site:
		return a.Site == b.Site
	case a.User != nil && b.User != nil:
		return *a.User == *b.User
	case a.Org != nil && b.Org != nil:
		return *a.Org == *b.Org
	}
	return false
}

func (s *settingsSubject) ToDefaultSettings() (*defaultSettingsResolver, bool) {
	return s.defaultSettings, s.defaultSettings != nil
}

func (s *settingsSubject) ToSite() (*siteResolver, bool) {
	return s.site, s.site != nil
}

func (s *settingsSubject) ToOrg() (*OrgResolver, bool) { return s.org, s.org != nil }

func (s *settingsSubject) ToUser() (*UserResolver, bool) { return s.user, s.user != nil }

func (s *settingsSubject) toSubject() api.SettingsSubject {
	switch {
	case s.site != nil:
		return api.SettingsSubject{Site: true}
	case s.org != nil:
		return api.SettingsSubject{Org: &s.org.org.ID}
	case s.user != nil:
		return api.SettingsSubject{User: &s.user.user.ID}
	default:
		panic("invalid settings subject")
	}
}

func (s *settingsSubject) ID() (graphql.ID, error) {
	switch {
	case s.defaultSettings != nil:
		return s.defaultSettings.ID(), nil
	case s.site != nil:
		return s.site.ID(), nil
	case s.org != nil:
		return s.org.ID(), nil
	case s.user != nil:
		return s.user.ID(), nil
	default:
		return "", errUnknownSettingsSubject
	}
}

func (s *settingsSubject) LatestSettings(ctx context.Context) (*settingsResolver, error) {
	switch {
	case s.defaultSettings != nil:
		return s.defaultSettings.LatestSettings(ctx)
	case s.site != nil:
		return s.site.LatestSettings(ctx)
	case s.org != nil:
		return s.org.LatestSettings(ctx)
	case s.user != nil:
		return s.user.LatestSettings(ctx)
	default:
		return nil, errUnknownSettingsSubject
	}
}

func (s *settingsSubject) SettingsURL() (*string, error) {
	switch {
	case s.defaultSettings != nil:
		return s.defaultSettings.SettingsURL(), nil
	case s.site != nil:
		return s.site.SettingsURL(), nil
	case s.org != nil:
		return s.org.SettingsURL(), nil
	case s.user != nil:
		return s.user.SettingsURL(), nil
	default:
		return nil, errUnknownSettingsSubject
	}
}

func (s *settingsSubject) ViewerCanAdminister(ctx context.Context) (bool, error) {
	switch {
	case s.defaultSettings != nil:
		return s.defaultSettings.ViewerCanAdminister(ctx)
	case s.site != nil:
		return s.site.ViewerCanAdminister(ctx)
	case s.org != nil:
		return s.org.ViewerCanAdminister(ctx)
	case s.user != nil:
		return s.user.ViewerCanAdminister(ctx)
	default:
		return false, errUnknownSettingsSubject
	}
}

func (s *settingsSubject) SettingsCascade() (*settingsCascade, error) {
	switch {
	case s.defaultSettings != nil:
		return s.defaultSettings.SettingsCascade(), nil
	case s.site != nil:
		return s.site.SettingsCascade(), nil
	case s.org != nil:
		return s.org.SettingsCascade(), nil
	case s.user != nil:
		return s.user.SettingsCascade(), nil
	default:
		return nil, errUnknownSettingsSubject
	}
}

func (s *settingsSubject) ConfigurationCascade() (*settingsCascade, error) {
	return s.SettingsCascade()
}

// readSettings unmarshals s's latest settings into v.
func (s *settingsSubject) readSettings(ctx context.Context, v interface{}) error {
	settings, err := s.LatestSettings(ctx)
	if err != nil {
		return err
	}
	if settings == nil {
		return nil
	}
	return jsonc.Unmarshal(settings.Contents(), &v)
}

// checkArgHasSameSubject ensures that the subject encoded in args.ID (or similar resolver
// field) is the same as that passed to the configurationMutationResolver. If they are different,
// it returns an error.
//
// 🚨 SECURITY: It is used when a mutation field inside the configurationMutation also accepts an
// ID field that encodes the settings subject. In that case, it's important to check that the
// subjects are equal to prevent a user from bypassing the permission check to write to the
// settings of the second ID's subject.
func (r *settingsMutation) checkArgHasSameSubject(argSubject api.SettingsSubject) error {
	if !settingsSubjectsEqual(r.subject.toSubject(), argSubject) {
		return fmt.Errorf("settings subject mismatch: %s != %s", r.subject.toSubject(), argSubject)
	}
	return nil
}
