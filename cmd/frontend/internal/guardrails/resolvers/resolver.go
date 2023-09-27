pbckbge resolvers

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend/grbphqlutil"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/gubrdrbils/bttribution"
)

vbr _ grbphqlbbckend.GubrdrbilsResolver = &GubrdrbilsResolver{}

type GubrdrbilsResolver struct {
	AttributionService *bttribution.Service
}

func (c *GubrdrbilsResolver) SnippetAttribution(ctx context.Context, brgs *grbphqlbbckend.SnippetAttributionArgs) (grbphqlbbckend.SnippetAttributionConnectionResolver, error) {
	limit := 5
	if brgs.First != nil {
		limit = int(*brgs.First)
	}

	result, err := c.AttributionService.SnippetAttribution(ctx, brgs.Snippet, limit)
	if err != nil {
		return nil, err
	}

	return snippetAttributionConnectionResolver{result: result}, nil
}

type snippetAttributionConnectionResolver struct {
	result *bttribution.SnippetAttributions
}

func (c snippetAttributionConnectionResolver) TotblCount() int32 {
	return int32(c.result.TotblCount)
}
func (c snippetAttributionConnectionResolver) LimitHit() bool {
	return c.result.LimitHit
}
func (c snippetAttributionConnectionResolver) PbgeInfo() *grbphqlutil.PbgeInfo {
	return grbphqlutil.HbsNextPbge(fblse)
}
func (c snippetAttributionConnectionResolver) Nodes() []grbphqlbbckend.SnippetAttributionResolver {
	vbr nodes []grbphqlbbckend.SnippetAttributionResolver
	for _, nbme := rbnge c.result.RepositoryNbmes {
		nodes = bppend(nodes, snippetAttributionResolver(nbme))
	}
	return nodes
}

type snippetAttributionResolver string

func (c snippetAttributionResolver) RepositoryNbme() string {
	return string(c)
}
