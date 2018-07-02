package graphqlbackend

import (
	"context"
	"strings"
	"time"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/ui/router"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/db"
)

// registryExtensionDBResolver implements the GraphQL type RegistryExtension.
type registryExtensionDBResolver struct {
	v *db.RegistryExtension

	cache *extensionRegistryCache
}

func (r *registryExtensionDBResolver) ID() graphql.ID {
	return marshalRegistryExtensionID(registryExtensionID{LocalID: r.v.ID})
}

func (r *registryExtensionDBResolver) UUID() string        { return r.v.UUID }
func (r *registryExtensionDBResolver) ExtensionID() string { return r.v.NonCanonicalExtensionID }
func (r *registryExtensionDBResolver) ExtensionIDWithoutRegistry() string {
	if r.v.NonCanonicalRegistry != "" {
		return strings.TrimPrefix(r.v.NonCanonicalExtensionID, r.v.NonCanonicalRegistry+"/")
	}
	return r.v.NonCanonicalExtensionID
}

func (r *registryExtensionDBResolver) Publisher(ctx context.Context) (*registryPublisher, error) {
	return getRegistryPublisher(ctx, r.v.Publisher)
}

func (r *registryExtensionDBResolver) Name() string { return r.v.Name }
func (r *registryExtensionDBResolver) Manifest() *extensionManifestResolver {
	return newExtensionManifestResolver(r.v.Manifest)
}

func (r *registryExtensionDBResolver) CreatedAt() *string {
	return strptr(r.v.CreatedAt.Format(time.RFC3339))
}

func (r *registryExtensionDBResolver) UpdatedAt() *string {
	return strptr(r.v.UpdatedAt.Format(time.RFC3339))
}

func (r *registryExtensionDBResolver) URL() string {
	return router.RegistryExtension(r.v.NonCanonicalExtensionID)
}
func (r *registryExtensionDBResolver) RemoteURL() *string { return nil }

func (r *registryExtensionDBResolver) RegistryName() (string, error) {
	return r.v.NonCanonicalRegistry, nil
}

func (r *registryExtensionDBResolver) IsLocal() bool { return true }

func (r *registryExtensionDBResolver) ExtensionConfigurationSubjects(ctx context.Context, args *registryExtensionExtensionConfigurationSubjectsConnectionArgs) (*extensionConfigurationSubjectConnection, error) {
	return listExtensionConfigurationSubjects(ctx, r.cache, &registryExtensionMultiResolver{local: r}, args)
}

func (r *registryExtensionDBResolver) Users(ctx context.Context, args *connectionArgs) (*userConnectionResolver, error) {
	return listRegistryExtensionUsers(ctx, r.cache, r.v.NonCanonicalExtensionID, args)
}

func (r *registryExtensionDBResolver) ViewerHasEnabled(ctx context.Context) (bool, error) {
	return viewerHasEnabledRegistryExtension(ctx, r.v.NonCanonicalExtensionID)
}

func (r *registryExtensionDBResolver) ViewerCanConfigure(ctx context.Context) (bool, error) {
	return viewerCanConfigureRegistryExtension(ctx)
}

func (r *registryExtensionDBResolver) ViewerCanAdminister(ctx context.Context) (bool, error) {
	err := toRegistryPublisherID(r.v).viewerCanAdminister(ctx)
	if err == backend.ErrMustBeSiteAdmin || err == backend.ErrNotAnOrgMember || err == backend.ErrNotAuthenticated {
		return false, nil
	}
	return err == nil, err
}

type configuredExtensionFromRegistryExtensionArgs struct {
	Subject *graphql.ID
}

func (r *registryExtensionDBResolver) ConfiguredExtension(ctx context.Context, args *configuredExtensionFromRegistryExtensionArgs) (*configuredExtensionResolver, error) {
	return configuredExtensionFromRegistryExtension(ctx, r.v.NonCanonicalExtensionID, *args)
}
