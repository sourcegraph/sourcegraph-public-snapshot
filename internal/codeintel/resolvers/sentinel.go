package resolvers

import (
	"context"

	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
)

type SentinelServiceResolver interface {
	// Fetch vulnerabilities
	Vulnerabilities(ctx context.Context, args GetVulnerabilitiesArgs) (VulnerabilityConnectionResolver, error)
	VulnerabilityByID(ctx context.Context, id graphql.ID) (_ VulnerabilityResolver, err error)

	// Fetch matches
	VulnerabilityMatches(ctx context.Context, args GetVulnerabilityMatchesArgs) (VulnerabilityMatchConnectionResolver, error)
	VulnerabilityMatchByID(ctx context.Context, id graphql.ID) (_ VulnerabilityMatchResolver, err error)
	VulnerabilityMatchesGroupByRepository(ctx context.Context, args GetVulnerabilityMatchGroupByRepositoryArgs) (VulnerabilityMatchGroupByRepositoryConnectionResolver, error)
}

type (
	GetVulnerabilitiesArgs               = PagedConnectionArgs
	GetVulnerabilityMatchesArgs          = PagedConnectionArgs
	VulnerabilityConnectionResolver      = PagedConnectionWithTotalCountResolver[VulnerabilityResolver]
	VulnerabilityMatchConnectionResolver = PagedConnectionWithTotalCountResolver[VulnerabilityMatchResolver]
)

type VulnerabilityResolver interface {
	ID() graphql.ID
	SourceID() string
	Summary() string
	Details() string
	CPEs() []string
	CWEs() []string
	Aliases() []string
	Related() []string
	DataSource() string
	URLs() []string
	Severity() string
	CVSSVector() string
	CVSSScore() string
	Published() gqlutil.DateTime
	Modified() *gqlutil.DateTime
	Withdrawn() *gqlutil.DateTime
	AffectedPackages() []VulnerabilityAffectedPackageResolver
}

type VulnerabilityAffectedPackageResolver interface {
	PackageName() string
	Language() string
	Namespace() string
	VersionConstraint() []string
	Fixed() bool
	FixedIn() *string
	AffectedSymbols() []VulnerabilityAffectedSymbolResolver
}

type VulnerabilityAffectedSymbolResolver interface {
	Path() string
	Symbols() []string
}

type VulnerabilityMatchResolver interface {
	ID() graphql.ID
	Vulnerability(ctx context.Context) (VulnerabilityResolver, error)
	AffectedPackage(ctx context.Context) (VulnerabilityAffectedPackageResolver, error)
	PreciseIndex(ctx context.Context) (PreciseIndexResolver, error)
}

type GetVulnerabilityMatchGroupByRepositoryArgs struct {
	PagedConnectionArgs
	RepositoryName *string
}

type VulnerabilityMatchGroupByRepositoryResolver interface {
	ID() graphql.ID
	RepositoryName() string
	MatchCount() int32
}

type VulnerabilityMatchGroupByRepositoryConnectionResolver interface {
	Nodes() []VulnerabilityMatchGroupByRepositoryResolver
	TotalCount() *int32
	PageInfo() PageInfo
}
