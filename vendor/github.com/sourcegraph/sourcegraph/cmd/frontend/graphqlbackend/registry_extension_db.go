package graphqlbackend

import (
	"context"
	"strings"
	"time"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/ui/router"
)

// registryExtensionDBResolver implements the GraphQL type RegistryExtension.
type registryExtensionDBResolver struct {
	v *db.RegistryExtension
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
func (r *registryExtensionDBResolver) Manifest(ctx context.Context) (*extensionManifestResolver, error) {
	manifest, err := backend.GetExtensionManifestWithBundleURL(ctx, r.v.NonCanonicalExtensionID, r.v.ID, "release")
	if err != nil {
		return nil, err
	}
	return newExtensionManifestResolver(manifest), nil
}

func (r *registryExtensionDBResolver) CreatedAt() *string {
	return strptr(r.v.CreatedAt.Format(time.RFC3339))
}

func (r *registryExtensionDBResolver) UpdatedAt() *string {
	return strptr(r.v.UpdatedAt.Format(time.RFC3339))
}

func (r *registryExtensionDBResolver) URL() string {
	return router.Extension(r.v.NonCanonicalExtensionID)
}
func (r *registryExtensionDBResolver) RemoteURL() *string { return nil }

func (r *registryExtensionDBResolver) RegistryName() (string, error) {
	return r.v.NonCanonicalRegistry, nil
}

func (r *registryExtensionDBResolver) IsLocal() bool { return true }

func (r *registryExtensionDBResolver) ViewerCanAdminister(ctx context.Context) (bool, error) {
	err := toRegistryPublisherID(r.v).viewerCanAdminister(ctx)
	if err == backend.ErrMustBeSiteAdmin || err == backend.ErrNotAnOrgMember || err == backend.ErrNotAuthenticated {
		return false, nil
	}
	return err == nil, err
}
