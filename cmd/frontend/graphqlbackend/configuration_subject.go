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

func (r *schemaResolver) ConfigurationSubject(ctx context.Context, args *struct{ ID graphql.ID }) (*configurationSubject, error) {
	return configurationSubjectByID(ctx, args.ID)
}

var errUnknownConfigurationSubject = errors.New("unknown configuration subject")

type configurationSubject struct {
	// Exactly 1 of these fields must be set.
	site *siteResolver
	org  *OrgResolver
	user *UserResolver
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

	case *UserResolver:
		// 🚨 SECURITY: Only the user and site admins are allowed to view the user's settings.
		if err := backend.CheckSiteAdminOrSameUser(ctx, s.user.ID); err != nil {
			return nil, err
		}
		return &configurationSubject{user: s}, nil

	case *OrgResolver:
		// 🚨 SECURITY: Check that the current user is a member of the org.
		if err := backend.CheckOrgAccess(ctx, s.org.ID); err != nil {
			return nil, err
		}
		return &configurationSubject{org: s}, nil

	default:
		return nil, errUnknownConfigurationSubject
	}
}

func configurationSubjectID(subject api.ConfigurationSubject) (graphql.ID, error) {
	switch {
	case subject.Site:
		return marshalSiteGQLID(singletonSiteResolver.gqlID), nil
	case subject.User != nil:
		return marshalUserID(*subject.User), nil
	case subject.Org != nil:
		return marshalOrgID(*subject.Org), nil
	default:
		return "", errUnknownConfigurationSubject
	}
}

func configurationSubjectsEqual(a, b api.ConfigurationSubject) bool {
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

func (s *configurationSubject) ToSite() (*siteResolver, bool) {
	return s.site, s.site != nil
}

func (s *configurationSubject) ToOrg() (*OrgResolver, bool) { return s.org, s.org != nil }

func (s *configurationSubject) ToUser() (*UserResolver, bool) { return s.user, s.user != nil }

func (s *configurationSubject) toSubject() api.ConfigurationSubject {
	switch {
	case s.site != nil:
		return api.ConfigurationSubject{Site: true}
	case s.org != nil:
		return api.ConfigurationSubject{Org: &s.org.org.ID}
	case s.user != nil:
		return api.ConfigurationSubject{User: &s.user.user.ID}
	default:
		panic("invalid configuration subject")
	}
}

func (s *configurationSubject) ID() (graphql.ID, error) {
	switch {
	case s.site != nil:
		return s.site.ID(), nil
	case s.org != nil:
		return s.org.ID(), nil
	case s.user != nil:
		return s.user.ID(), nil
	default:
		return "", errUnknownConfigurationSubject
	}
}

func (s *configurationSubject) LatestSettings(ctx context.Context) (*settingsResolver, error) {
	switch {
	case s.site != nil:
		return s.site.LatestSettings(ctx)
	case s.org != nil:
		return s.org.LatestSettings(ctx)
	case s.user != nil:
		return s.user.LatestSettings(ctx)
	default:
		return nil, errUnknownConfigurationSubject
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
		return "", errUnknownConfigurationSubject
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
		return false, errUnknownConfigurationSubject
	}
}

func (s *configurationSubject) ConfigurationCascade() (*configurationCascadeResolver, error) {
	switch {
	case s.site != nil:
		return s.site.ConfigurationCascade(), nil
	case s.org != nil:
		return s.org.ConfigurationCascade(), nil
	case s.user != nil:
		return s.user.ConfigurationCascade(), nil
	default:
		return nil, errUnknownConfigurationSubject
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
	return jsonc.Unmarshal(settings.Contents(), &v)
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
