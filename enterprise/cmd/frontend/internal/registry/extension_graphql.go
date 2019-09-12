package registry

import (
	"context"
	"strings"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/registry"
)

// extensionDBResolver implements the GraphQL type RegistryExtension.
type extensionDBResolver struct {
	v *dbExtension
}

func (r *extensionDBResolver) ID() graphql.ID {
	return registry.MarshalRegistryExtensionID(registry.RegistryExtensionID{LocalID: r.v.ID})
}

func (r *extensionDBResolver) UUID() string        { return r.v.UUID }
func (r *extensionDBResolver) ExtensionID() string { return r.v.NonCanonicalExtensionID }
func (r *extensionDBResolver) ExtensionIDWithoutRegistry() string {
	if r.v.NonCanonicalRegistry != "" {
		return strings.TrimPrefix(r.v.NonCanonicalExtensionID, r.v.NonCanonicalRegistry+"/")
	}
	return r.v.NonCanonicalExtensionID
}

func (r *extensionDBResolver) Publisher(ctx context.Context) (graphqlbackend.RegistryPublisher, error) {
	return getRegistryPublisher(ctx, r.v.Publisher)
}

func (r *extensionDBResolver) Name() string { return r.v.Name }
func (r *extensionDBResolver) Manifest(ctx context.Context) (graphqlbackend.ExtensionManifest, error) {
	manifest, _, err := getExtensionManifestWithBundleURL(ctx, r.v.NonCanonicalExtensionID, r.v.ID, "release")
	if err != nil {
		return nil, err
	}
	return registry.NewExtensionManifest(manifest), nil
}

func (r *extensionDBResolver) CreatedAt() *graphqlbackend.DateTime {
	return &graphqlbackend.DateTime{Time: r.v.CreatedAt}
}

func (r *extensionDBResolver) UpdatedAt() *graphqlbackend.DateTime {
	return &graphqlbackend.DateTime{Time: r.v.UpdatedAt}
}

func (r *extensionDBResolver) PublishedAt(ctx context.Context) (*graphqlbackend.DateTime, error) {
	_, publishedAt, err := getExtensionManifestWithBundleURL(ctx, r.v.NonCanonicalExtensionID, r.v.ID, "release")
	if err != nil {
		return nil, err
	}
	return &graphqlbackend.DateTime{Time: publishedAt}, nil
}

func (r *extensionDBResolver) URL() string {
	return registry.ExtensionURL(r.v.NonCanonicalExtensionID)
}
func (r *extensionDBResolver) RemoteURL() *string { return nil }

func (r *extensionDBResolver) RegistryName() (string, error) {
	return r.v.NonCanonicalRegistry, nil
}

func (r *extensionDBResolver) IsLocal() bool { return true }

func (r *extensionDBResolver) IsWorkInProgress() bool {
	return r.v.NonCanonicalIsWorkInProgress
}

func (r *extensionDBResolver) ViewerCanAdminister(ctx context.Context) (bool, error) {
	err := toRegistryPublisherID(r.v).viewerCanAdminister(ctx)
	if err == backend.ErrMustBeSiteAdmin || err == backend.ErrNotAnOrgMember || err == backend.ErrNotAuthenticated {
		return false, nil
	}
	if _, ok := err.(*backend.InsufficientAuthorizationError); ok {
		return false, nil
	}
	return err == nil, err
}

func strptr(s string) *string { return &s }
