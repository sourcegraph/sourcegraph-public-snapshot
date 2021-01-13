package graphql

import (
	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/codeintel/resolvers"
)

type PackageResolver struct {
	pkg resolvers.AdjustedPackage
}

func NewPackageResolver(pkg resolvers.AdjustedPackage) gql.PackageResolver {
	return &PackageResolver{
		pkg: pkg,
	}
}

func (r *PackageResolver) Name() string {
	return r.pkg.Package.Name
}

func (r *PackageResolver) Version() string {
	return r.pkg.Package.Version
}

func (r *PackageResolver) Manager() string {
	return r.pkg.Package.Manager
}
