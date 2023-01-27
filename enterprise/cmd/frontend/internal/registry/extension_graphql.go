package registry

import (
	"context"

	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	registry "github.com/sourcegraph/sourcegraph/cmd/frontend/registry/api"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/registry/stores"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

// extensionDBResolver implements the GraphQL type RegistryExtension.
type extensionDBResolver struct {
	db database.DB
	v  *stores.Extension

	// Supplied as part of list endpoints, but
	// calculated as part of single-extension endpoints
	r *stores.Release
}

func (r *extensionDBResolver) ID() graphql.ID {
	return registry.MarshalRegistryExtensionID(registry.RegistryExtensionID{LocalID: r.v.ID})
}

func (r *extensionDBResolver) ExtensionID() string { return r.v.NonCanonicalExtensionID }

func (r *extensionDBResolver) Manifest(ctx context.Context) (graphqlbackend.ExtensionManifest, error) {
	release, err := r.release(ctx)
	if err != nil {
		return nil, err
	}
	if release == nil {
		return registry.NewExtensionManifest(nil), nil
	}
	return registry.NewExtensionManifest(&release.Manifest), nil
}

func (r *extensionDBResolver) release(ctx context.Context) (*stores.Release, error) {
	if r.r != nil {
		return r.r, nil
	}

	var err error
	r.r, err = getLatestRelease(ctx, stores.Releases(r.db), r.v.NonCanonicalExtensionID, r.v.ID, "release")
	return r.r, err
}
