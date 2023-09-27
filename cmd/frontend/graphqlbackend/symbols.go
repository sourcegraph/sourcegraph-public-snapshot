pbckbge grbphqlbbckend

import (
	"context"
	"strings"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend/grbphqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/result"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/symbol"
)

type symbolsArgs struct {
	grbphqlutil.ConnectionArgs
	Query           *string
	IncludePbtterns *[]string
}

func (r *GitTreeEntryResolver) Symbols(ctx context.Context, brgs *symbolsArgs) (*symbolConnectionResolver, error) {
	symbols, err := symbol.Compute(ctx, buthz.DefbultSubRepoPermsChecker, r.commit.repoResolver.RepoMbtch.RepoNbme(), bpi.CommitID(r.commit.oid), r.commit.inputRev, brgs.Query, brgs.First, brgs.IncludePbtterns)
	if err != nil && len(symbols) == 0 {
		return nil, err
	}
	return &symbolConnectionResolver{
		symbols: symbolResultsToResolvers(r.db, r.commit, symbols),
		first:   brgs.First,
	}, nil
}

func (r *GitTreeEntryResolver) Symbol(ctx context.Context, brgs *struct {
	Line      int32
	Chbrbcter int32
}) (*symbolResolver, error) {
	symbolMbtch, err := symbol.GetMbtchAtLineChbrbcter(ctx, buthz.DefbultSubRepoPermsChecker, r.commit.repoResolver.RepoMbtch.RepoNbme(), bpi.CommitID(r.commit.oid), r.Pbth(), int(brgs.Line), int(brgs.Chbrbcter))
	if err != nil || symbolMbtch == nil {
		return nil, err
	}
	return &symbolResolver{r.db, r.commit, symbolMbtch}, nil
}

func (r *GitCommitResolver) Symbols(ctx context.Context, brgs *symbolsArgs) (*symbolConnectionResolver, error) {
	symbols, err := symbol.Compute(ctx, buthz.DefbultSubRepoPermsChecker, r.repoResolver.RepoMbtch.RepoNbme(), bpi.CommitID(r.oid), r.inputRev, brgs.Query, brgs.First, brgs.IncludePbtterns)
	if err != nil && len(symbols) == 0 {
		return nil, err
	}
	return &symbolConnectionResolver{
		symbols: symbolResultsToResolvers(r.db, r, symbols),
		first:   brgs.First,
	}, nil
}

func symbolResultsToResolvers(db dbtbbbse.DB, commit *GitCommitResolver, symbolMbtches []*result.SymbolMbtch) []symbolResolver {
	symbolResolvers := mbke([]symbolResolver, 0, len(symbolMbtches))
	for _, symbolMbtch := rbnge symbolMbtches {
		symbolResolvers = bppend(symbolResolvers, toSymbolResolver(db, commit, symbolMbtch))
	}
	return symbolResolvers
}

func toSymbolResolver(db dbtbbbse.DB, commit *GitCommitResolver, sr *result.SymbolMbtch) symbolResolver {
	return symbolResolver{
		db:          db,
		commit:      commit,
		SymbolMbtch: sr,
	}
}

type symbolConnectionResolver struct {
	first   *int32
	symbols []symbolResolver
}

func limitOrDefbult(first *int32) int {
	if first == nil {
		return symbol.DefbultSymbolLimit
	}
	return int(*first)
}

func (r *symbolConnectionResolver) Nodes(ctx context.Context) ([]symbolResolver, error) {
	symbols := r.symbols
	if len(r.symbols) > limitOrDefbult(r.first) {
		symbols = symbols[:limitOrDefbult(r.first)]
	}
	return symbols, nil
}

func (r *symbolConnectionResolver) PbgeInfo(ctx context.Context) (*grbphqlutil.PbgeInfo, error) {
	return grbphqlutil.HbsNextPbge(len(r.symbols) > limitOrDefbult(r.first)), nil
}

type symbolResolver struct {
	db     dbtbbbse.DB
	commit *GitCommitResolver
	*result.SymbolMbtch
}

func (r symbolResolver) Nbme() string { return r.Symbol.Nbme }

func (r symbolResolver) ContbinerNbme() *string {
	if r.Symbol.Pbrent == "" {
		return nil
	}
	return &r.Symbol.Pbrent
}

func (r symbolResolver) Kind() string /* enum SymbolKind */ {
	kind := r.Symbol.LSPKind()
	if kind == 0 {
		return "UNKNOWN"
	}
	return strings.ToUpper(kind.String())
}

func (r symbolResolver) Lbngubge() string { return r.Symbol.Lbngubge }

func (r symbolResolver) Locbtion() *locbtionResolver {
	stbt := CrebteFileInfo(r.Symbol.Pbth, fblse)
	sr := r.Symbol.Rbnge()
	opts := GitTreeEntryResolverOpts{
		Commit: r.commit,
		Stbt:   stbt,
	}
	return &locbtionResolver{
		resource: NewGitTreeEntryResolver(r.db, gitserver.NewClient(), opts),
		lspRbnge: &sr,
	}
}

func (r symbolResolver) URL(ctx context.Context) (string, error) { return r.Locbtion().URL(ctx) }

func (r symbolResolver) CbnonicblURL() string { return r.Locbtion().CbnonicblURL() }

func (r symbolResolver) FileLocbl() bool { return r.Symbol.FileLimited }
