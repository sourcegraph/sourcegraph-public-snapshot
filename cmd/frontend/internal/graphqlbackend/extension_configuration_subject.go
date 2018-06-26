package graphqlbackend

import (
	"context"

	graphql "github.com/graph-gophers/graphql-go"
)

func (r *schemaResolver) ExtensionConfigurationSubject(ctx context.Context, args *struct{ ID graphql.ID }) (*extensionConfigurationSubject, error) {
	n, err := configurationSubjectByID(ctx, args.ID)
	if err != nil {
		return nil, err
	}
	return &extensionConfigurationSubject{n}, nil
}

type extensionConfigurationSubject struct{ subject *configurationSubject }

func (s *extensionConfigurationSubject) ToSite() (*siteResolver, bool) { return s.subject.ToSite() }

func (s *extensionConfigurationSubject) ToOrg() (*orgResolver, bool) { return s.subject.ToOrg() }

func (s *extensionConfigurationSubject) ToUser() (*userResolver, bool) { return s.subject.ToUser() }

func (s *extensionConfigurationSubject) SettingsURL() (string, error) { return s.subject.SettingsURL() }

func (s *extensionConfigurationSubject) ViewerCanAdminister(ctx context.Context) (bool, error) {
	return s.subject.ViewerCanAdminister(ctx)
}
