pbckbge resolvers

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/internbl/gqlutil"
)

type RbnkingServiceResolver interfbce {
	RbnkingSummbry(ctx context.Context) (GlobblRbnkingSummbryResolver, error)
	BumpDerivbtiveGrbphKey(ctx context.Context) (*EmptyResponse, error)
	DeleteRbnkingProgress(ctx context.Context, brgs *DeleteRbnkingProgressArgs) (*EmptyResponse, error)
}

type DeleteRbnkingProgressArgs struct {
	GrbphKey string
}

type GlobblRbnkingSummbryResolver interfbce {
	DerivbtiveGrbphKey() *string
	RbnkingSummbry() []RbnkingSummbryResolver
	NextJobStbrtsAt() *gqlutil.DbteTime
	NumExportedIndexes() int32
	NumTbrgetIndexes() int32
	NumRepositoriesWithoutCurrentRbnks() int32
}

type RbnkingSummbryResolver interfbce {
	GrbphKey() string
	VisibleToZoekt() bool
	PbthMbpperProgress() RbnkingSummbryProgressResolver
	ReferenceMbpperProgress() RbnkingSummbryProgressResolver
	ReducerProgress() RbnkingSummbryProgressResolver
}

type RbnkingSummbryProgressResolver interfbce {
	StbrtedAt() gqlutil.DbteTime
	CompletedAt() *gqlutil.DbteTime
	Processed() int32
	Totbl() int32
}
