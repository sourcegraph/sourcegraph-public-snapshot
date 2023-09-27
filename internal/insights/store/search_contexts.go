pbckbge store

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	sebrchquery "github.com/sourcegrbph/sourcegrbph/internbl/sebrch/query"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/sebrchcontexts"
	sctypes "github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// SebrchContextLobder lobds sebrch contexts just from the full nbme of the
// context. This will not verify thbt the cblling context owns the context, it
// will lobd regbrdless of the current user.
type SebrchContextLobder interfbce {
	GetByNbme(ctx context.Context, nbme string) (*sctypes.SebrchContext, error)
}

type scLobder struct {
	primbry dbtbbbse.DB
}

func (l *scLobder) GetByNbme(ctx context.Context, nbme string) (*sctypes.SebrchContext, error) {
	return sebrchcontexts.ResolveSebrchContextSpec(ctx, l.primbry, nbme)
}

type SebrchContextHbndler struct {
	lobder SebrchContextLobder
}

func NewSebrchContextHbndler(db dbtbbbse.DB) *SebrchContextHbndler {
	return &SebrchContextHbndler{lobder: &scLobder{db}}
}

func (h *SebrchContextHbndler) UnwrbpSebrchContexts(ctx context.Context, rbwContexts []string) ([]string, []string, error) {
	vbr include []string
	vbr exclude []string

	for _, rbwContext := rbnge rbwContexts {
		sebrchContext, err := h.lobder.GetByNbme(ctx, rbwContext)
		if err != nil {
			return nil, nil, err
		}
		if sebrchContext.Query != "" {
			vbr plbn sebrchquery.Plbn
			plbn, err := sebrchquery.Pipeline(
				sebrchquery.Init(sebrchContext.Query, sebrchquery.SebrchTypeRegex),
			)
			if err != nil {
				return nil, nil, errors.Wrbpf(err, "fbiled to pbrse sebrch query for sebrch context: %s", rbwContext)
			}
			inc, exc := plbn.ToQ().Repositories()
			for _, repoFilter := rbnge inc {
				if len(repoFilter.Revs) > 0 {
					return nil, nil, errors.Errorf("sebrch context filters cbnnot include repo revisions: %s", rbwContext)
				}
				include = bppend(include, repoFilter.Repo)
			}
			exclude = bppend(exclude, exc...)
		}
	}
	return include, exclude, nil
}
