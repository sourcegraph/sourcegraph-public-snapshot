package graphqlbackend

import (
	"context"

	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func (r *schemaResolver) SettingsSubject(ctx context.Context, args *struct{ ID graphql.ID }) (*settingsSubjectResolver, error) {
	n, err := r.nodeByID(ctx, args.ID)
	if err != nil {
		return nil, err
	}

	return settingsSubjectForNodeAndCheckAccess(ctx, n)
}

var errUnknownSettingsSubject = errors.New("unknown settings subject")

type settingsSubjectResolver struct {
	// Exactly 1 of these fields must be set.
	defaultSettings *defaultSettingsResolver
	site            *siteResolver
	org             *OrgResolver
	user            *UserResolver

	// 🚨 SECURITY: Only the settingsSubjectForNodeAndCheckAccess function can set this. It is used
	// to ensure that access checks have been run on this value, so that we don't leak settings to
	// an unauthorized viewer by an accidental bypass of access checks. This struct type is
	// naturally constructed all over the place (because many types of nodes have settings), and it
	// was too easy to bypass the access check accidentally.
	checkedAccess_DO_NOT_SET_THIS_MANUALLY_OR_YOU_WILL_LEAK_SECRETS bool
}

func (r *settingsSubjectResolver) assertCheckedAccess() {
	if !r.checkedAccess_DO_NOT_SET_THIS_MANUALLY_OR_YOU_WILL_LEAK_SECRETS {
		panic("settingsSubjectResolver.assertCheckedAccess: access checks have not been run on this value")
	}
}

func resolverForSubject(ctx context.Context, logger log.Logger, db database.DB, subject api.SettingsSubject) (*settingsSubjectResolver, error) {
	if subject.Default {
		return &settingsSubjectResolver{defaultSettings: newDefaultSettingsResolver(db)}, nil
	}

	var (
		node Node
		err  error
	)
	switch {
	case subject.Site:
		node = NewSiteResolver(logger, db)
	case subject.Org != nil:
		node, err = OrgByIDInt32(ctx, db, *subject.Org)
	case subject.User != nil:
		node, err = UserByIDInt32(ctx, db, *subject.User)
	default:
		panic("subject must have exactly one field set")
	}
	if err != nil {
		return nil, err
	}

	// 🚨 SECURITY: Call settingsSubjectForNode to reuse the security checks implemented there.
	return settingsSubjectForNodeAndCheckAccess(ctx, node)
}

func resolversForSubjects(ctx context.Context, logger log.Logger, db database.DB, subjects []api.SettingsSubject) (_ []*settingsSubjectResolver, err error) {
	res := make([]*settingsSubjectResolver, len(subjects))
	for i, subject := range subjects {
		res[i], err = resolverForSubject(ctx, logger, db, subject)
		if err != nil {
			return nil, err
		}
	}
	return res, nil
}

// settingsSubjectForNodeAndCheckAccess fetches the settings subject for the given Node. If the node
// is not a valid settings subject, an error is returned.
//
// 🚨 SECURITY: This function also ensures that the actor is permitted to view the node's settings.
// It is the ONLY place that the
// (settingsSubjectResolver).checkedAccess_DO_NOT_SET_THIS_MANUALLY_OR_YOU_WILL_LEAK_SECRETS field
// can be set.
func settingsSubjectForNodeAndCheckAccess(ctx context.Context, n Node) (*settingsSubjectResolver, error) {
	var subject settingsSubjectResolver

	switch s := n.(type) {
	case *defaultSettingsResolver:
		subject.defaultSettings = s

	case *siteResolver:
		subject.site = s

	case *UserResolver:
		// 🚨 SECURITY: The user and site admins are allowed to view the user's settings otherwise.
		if err := auth.CheckSiteAdminOrSameUser(ctx, s.db, s.user.ID); err != nil {
			return nil, err
		}
		subject.user = s

	case *OrgResolver:
		// 🚨 SECURITY: Only org members or site admins can view the org settings.
		if err := auth.CheckOrgAccessOrSiteAdmin(ctx, s.db, s.org.ID); err != nil {
			return nil, err
		}
		subject.org = s

	default:
		return nil, errUnknownSettingsSubject
	}

	// 🚨 SECURITY: This is the ONLY place that this field can be set.
	subject.checkedAccess_DO_NOT_SET_THIS_MANUALLY_OR_YOU_WILL_LEAK_SECRETS = true

	return &subject, nil
}

func (s *settingsSubjectResolver) ToDefaultSettings() (*defaultSettingsResolver, bool) {
	return s.defaultSettings, s.defaultSettings != nil
}

func (s *settingsSubjectResolver) ToSite() (*siteResolver, bool) {
	return s.site, s.site != nil
}

func (s *settingsSubjectResolver) ToOrg() (*OrgResolver, bool) { return s.org, s.org != nil }

func (s *settingsSubjectResolver) ToUser() (*UserResolver, bool) { return s.user, s.user != nil }

func (s *settingsSubjectResolver) toSubject() api.SettingsSubject {
	switch {
	case s.defaultSettings != nil:
		return api.SettingsSubject{Default: true}
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

func (s *settingsSubjectResolver) ID() (graphql.ID, error) {
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

func (s *settingsSubjectResolver) LatestSettings(ctx context.Context) (*settingsResolver, error) {
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

func (s *settingsSubjectResolver) SettingsURL() (*string, error) {
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

func (s *settingsSubjectResolver) ViewerCanAdminister(ctx context.Context) (bool, error) {
	switch {
	case s.defaultSettings != nil:
		return s.defaultSettings.ViewerCanAdminister(ctx)
	case s.site != nil:
		return s.site.ViewerCanAdminister(ctx)
	case s.org != nil:
		return s.org.ViewerCanAdminister(ctx)
	case s.user != nil:
		return s.user.viewerCanAdministerSettings()
	default:
		return false, errUnknownSettingsSubject
	}
}

func (s *settingsSubjectResolver) SettingsCascade(ctx context.Context) (*settingsCascade, error) {
	switch {
	case s.defaultSettings != nil:
		return s.defaultSettings.SettingsCascade(ctx)
	case s.site != nil:
		return s.site.SettingsCascade(ctx)
	case s.org != nil:
		return s.org.SettingsCascade(ctx)
	case s.user != nil:
		return s.user.SettingsCascade(ctx)
	default:
		return nil, errUnknownSettingsSubject
	}
}

func (s *settingsSubjectResolver) ConfigurationCascade(ctx context.Context) (*settingsCascade, error) {
	return s.SettingsCascade(ctx)
}
