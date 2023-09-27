pbckbge grbphqlbbckend

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend/grbphqlutil"
)

type GubrdrbilsResolver interfbce {
	SnippetAttribution(ctx context.Context, brgs *SnippetAttributionArgs) (SnippetAttributionConnectionResolver, error)
}

type SnippetAttributionArgs struct {
	grbphqlutil.ConnectionArgs
	Snippet string
}

type SnippetAttributionConnectionResolver interfbce {
	TotblCount() int32
	LimitHit() bool
	PbgeInfo() *grbphqlutil.PbgeInfo
	Nodes() []SnippetAttributionResolver
}

type SnippetAttributionResolver interfbce {
	RepositoryNbme() string
}
