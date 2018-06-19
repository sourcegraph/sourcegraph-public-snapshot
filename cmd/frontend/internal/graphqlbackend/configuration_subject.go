package graphqlbackend

import (
	"context"
	"errors"
	"fmt"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/backend"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
)

type configurationSubject struct {
	site *siteResolver
	org  *orgResolver
	user *userResolver
}

// configurationSubjectByID fetches the configuration subject with the given ID. If the ID
// refers to a node that is not a valid configuration subject, an error is returned.
func configurationSubjectByID(ctx context.Context, id graphql.ID) (*configurationSubject, error) {
	resolver, err := nodeByID(ctx, id)
	if err != nil {
		return nil, err
	}

	switch s := resolver.(type) {
	case *siteResolver:
		return &configurationSubject{site: s}, nil

	case *userResolver:
		// 🚨 SECURITY: Only the user and site admins are allowed to view the user's settings.
		if err := backend.CheckSiteAdminOrSameUser(ctx, s.user.ID); err != nil {
			return nil, err
		}
		return &configurationSubject{user: s}, nil

	case *orgResolver:
		// 🚨 SECURITY: Check that the current user is a member of the org.
		if err := backend.CheckOrgAccess(ctx, s.org.ID); err != nil {
			return nil, err
		}
		return &configurationSubject{org: s}, nil

	default:
		return nil, errors.New("bad configuration subject type")
	}
}

func configurationSubjectID(subject api.ConfigurationSubject) (graphql.ID, error) {
	switch {
	case subject.Site != nil:
		return marshalSiteGQLID(*subject.Site), nil
	case subject.User != nil:
		return marshalUserID(*subject.User), nil
	case subject.Org != nil:
		return marshalOrgID(*subject.Org), nil
	default:
		return "", errors.New("bad configuration subject type")
	}
}

func configurationSubjectsEqual(a, b api.ConfigurationSubject) bool {
	switch {
	case a.Site != nil && b.Site != nil:
		return *a.Site == *b.Site
	case a.User != nil && b.User != nil:
		return *a.User == *b.User
	case a.Org != nil && b.Org != nil:
		return *a.Org == *b.Org
	}
	return false
}

func (s *configurationSubject) ToSite() (*siteResolver, bool) { return s.site, s.site != nil }

func (s *configurationSubject) ToOrg() (*orgResolver, bool) { return s.org, s.org != nil }

func (s *configurationSubject) ToUser() (*userResolver, bool) { return s.user, s.user != nil }

func (s *configurationSubject) toSubject() api.ConfigurationSubject {
	switch {
	case s.site != nil:
		return api.ConfigurationSubject{Site: &s.site.gqlID}
	case s.org != nil:
		return api.ConfigurationSubject{Org: &s.org.org.ID}
	case s.user != nil:
		return api.ConfigurationSubject{User: &s.user.user.ID}
	default:
		return api.ConfigurationSubject{}
	}
}

func (s *configurationSubject) ID() graphql.ID {
	switch {
	case s.site != nil:
		return s.site.ID()
	case s.org != nil:
		return s.org.ID()
	case s.user != nil:
		return s.user.ID()
	}
	return "Global"
}

func (s *configurationSubject) LatestSettings(ctx context.Context) (*settingsResolver, error) {
	switch {
	case s.site != nil:
		return s.site.LatestSettings()
	case s.org != nil:
		return s.org.LatestSettings(ctx)
	case s.user != nil:
		return s.user.LatestSettings(ctx)
	default:
		return currentSiteSettings(ctx)
	}
}

func (s *configurationSubject) SettingsURL() (string, error) {
	switch {
	case s.site != nil:
		return s.site.SettingsURL(), nil
	case s.org != nil:
		return s.org.SettingsURL(), nil
	case s.user != nil:
		return s.user.SettingsURL(), nil
	default:
		return "", errors.New("unknown configuration subject")
	}
}

func (s *configurationSubject) ViewerCanAdminister(ctx context.Context) (bool, error) {
	switch {
	case s.site != nil:
		return s.site.ViewerCanAdminister(ctx)
	case s.org != nil:
		return s.org.ViewerCanAdminister(ctx)
	case s.user != nil:
		return s.user.ViewerCanAdminister(ctx)
	default:
		return false, errors.New("unknown configuration subject")
	}
}

// readConfiguration unmarshals s's latest settings into v.
func (s *configurationSubject) readConfiguration(ctx context.Context, v interface{}) error {
	settings, err := s.LatestSettings(ctx)
	if err != nil {
		return err
	}
	if settings == nil {
		return nil
	}
	return conf.UnmarshalJSON(settings.Contents(), &v)
}

// checkArgHasSameSubject ensures that the subject encoded in args.ID (or similar resolver
// field) is the same as that passed to the configurationMutationResolver. If they are different,
// it returns an error.
//
// 🚨 SECURITY: It is used when a mutation field inside the configurationMutation also accepts an
// ID field that encodes the configuration subject. In that case, it's important to check that the
// subjects are equal to prevent a user from bypassing the permission check to write to the
// configuration of the second ID's subject.
func (r *configurationMutationResolver) checkArgHasSameSubject(argSubject api.ConfigurationSubject) error {
	if !configurationSubjectsEqual(r.subject.toSubject(), argSubject) {
		return fmt.Errorf("configuration subject mismatch: %s != %s", r.subject.toSubject(), argSubject)
	}
	return nil
}
