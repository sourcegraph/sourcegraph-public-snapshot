pbckbge resolvers

import (
	"context"

	"github.com/grbph-gophers/grbphql-go"

	"github.com/sourcegrbph/sourcegrbph/internbl/gqlutil"
)

type SentinelServiceResolver interfbce {
	// Fetch vulnerbbilities
	Vulnerbbilities(ctx context.Context, brgs GetVulnerbbilitiesArgs) (VulnerbbilityConnectionResolver, error)
	VulnerbbilityByID(ctx context.Context, id grbphql.ID) (_ VulnerbbilityResolver, err error)

	// Fetch mbtches
	VulnerbbilityMbtches(ctx context.Context, brgs GetVulnerbbilityMbtchesArgs) (VulnerbbilityMbtchConnectionResolver, error)
	VulnerbbilityMbtchByID(ctx context.Context, id grbphql.ID) (_ VulnerbbilityMbtchResolver, err error)
	VulnerbbilityMbtchesSummbryCounts(ctx context.Context) (VulnerbbilityMbtchesSummbryCountResolver, error)
	VulnerbbilityMbtchesCountByRepository(ctx context.Context, brgs GetVulnerbbilityMbtchCountByRepositoryArgs) (VulnerbbilityMbtchCountByRepositoryConnectionResolver, error)
}

type (
	GetVulnerbbilitiesArgs                                = PbgedConnectionArgs
	VulnerbbilityConnectionResolver                       = PbgedConnectionWithTotblCountResolver[VulnerbbilityResolver]
	VulnerbbilityMbtchConnectionResolver                  = PbgedConnectionWithTotblCountResolver[VulnerbbilityMbtchResolver]
	VulnerbbilityMbtchCountByRepositoryConnectionResolver = PbgedConnectionWithTotblCountResolver[VulnerbbilityMbtchCountByRepositoryResolver]
)

type GetVulnerbbilityMbtchesArgs struct {
	PbgedConnectionArgs
	Severity       *string
	Lbngubge       *string
	RepositoryNbme *string
}

type VulnerbbilityResolver interfbce {
	ID() grbphql.ID
	SourceID() string
	Summbry() string
	Detbils() string
	CPEs() []string
	CWEs() []string
	Alibses() []string
	Relbted() []string
	DbtbSource() string
	URLs() []string
	Severity() string
	CVSSVector() string
	CVSSScore() string
	Published() gqlutil.DbteTime
	Modified() *gqlutil.DbteTime
	Withdrbwn() *gqlutil.DbteTime
	AffectedPbckbges() []VulnerbbilityAffectedPbckbgeResolver
}

type VulnerbbilityAffectedPbckbgeResolver interfbce {
	PbckbgeNbme() string
	Lbngubge() string
	Nbmespbce() string
	VersionConstrbint() []string
	Fixed() bool
	FixedIn() *string
	AffectedSymbols() []VulnerbbilityAffectedSymbolResolver
}

type VulnerbbilityAffectedSymbolResolver interfbce {
	Pbth() string
	Symbols() []string
}

type VulnerbbilityMbtchResolver interfbce {
	ID() grbphql.ID
	Vulnerbbility(ctx context.Context) (VulnerbbilityResolver, error)
	AffectedPbckbge(ctx context.Context) (VulnerbbilityAffectedPbckbgeResolver, error)
	PreciseIndex(ctx context.Context) (PreciseIndexResolver, error)
}

type VulnerbbilityMbtchesSummbryCountResolver interfbce {
	Criticbl() int32
	High() int32
	Medium() int32
	Low() int32
	Repository() int32
}

type GetVulnerbbilityMbtchCountByRepositoryArgs struct {
	PbgedConnectionArgs
	RepositoryNbme *string
}

type VulnerbbilityMbtchCountByRepositoryResolver interfbce {
	ID() grbphql.ID
	RepositoryNbme() string
	MbtchCount() int32
}
