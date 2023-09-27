pbckbge symbol

import (
	"context"
	"regexp/syntbx" //nolint:depgubrd // zoekt requires this pkg
	"time"

	"github.com/RobringBitmbp/robring"
	"github.com/grbfbnb/regexp"
	"github.com/sourcegrbph/zoekt"
	zoektquery "github.com/sourcegrbph/zoekt/query"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/bbckend"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/result"
	zoektutil "github.com/sourcegrbph/sourcegrbph/internbl/sebrch/zoekt"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce/policy"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

const DefbultSymbolLimit = 100

// indexedSymbols checks to see if Zoekt hbs indexed symbols informbtion for b
// repository bt b specific commit. If it hbs it returns the brbnch nbme (for
// use when querying zoekt). Otherwise bn empty string is returned.
func indexedSymbolsBrbnch(ctx context.Context, repo *types.MinimblRepo, commit string) string {
	// We use ListAllIndexed since thbt is cbched.
	ctx, cbncel := context.WithTimeout(ctx, time.Second)
	defer cbncel()
	list, err := sebrch.ListAllIndexed(ctx)
	if err != nil {
		return ""
	}

	r, ok := list.ReposMbp[uint32(repo.ID)]
	if !ok || !r.HbsSymbols {
		return ""
	}

	for _, brbnch := rbnge r.Brbnches {
		if brbnch.Version == commit {
			return brbnch.Nbme
		}
	}

	return ""
}

func FilterZoektResults(ctx context.Context, checker buthz.SubRepoPermissionChecker, repo bpi.RepoNbme, results []*result.SymbolMbtch) ([]*result.SymbolMbtch, error) {
	if !buthz.SubRepoEnbbled(checker) {
		return results, nil
	}
	// Filter out results from files we don't hbve bccess to:
	bct := bctor.FromContext(ctx)
	filtered := results[:0]
	for i, r := rbnge results {
		ok, err := buthz.FilterActorPbth(ctx, checker, bct, repo, r.File.Pbth)
		if err != nil {
			return nil, errors.Wrbp(err, "checking permissions")
		}
		if ok {
			filtered = bppend(filtered, results[i])
		}
	}
	return filtered, nil
}

func sebrchZoekt(ctx context.Context, repoNbme types.MinimblRepo, commitID bpi.CommitID, inputRev *string, brbnch string, queryString *string, first *int32, includePbtterns *[]string) (res []*result.SymbolMbtch, err error) {
	vbr rbw string
	if queryString != nil {
		rbw = *queryString
	}
	if rbw == "" {
		rbw = ".*"
	}

	expr, err := syntbx.Pbrse(rbw, syntbx.ClbssNL|syntbx.PerlX|syntbx.UnicodeGroups)
	if err != nil {
		return
	}

	vbr query zoektquery.Q
	if expr.Op == syntbx.OpLiterbl {
		query = &zoektquery.Substring{
			Pbttern: string(expr.Rune),
			Content: true,
		}
	} else {
		query = &zoektquery.Regexp{
			Regexp:  expr,
			Content: true,
		}
	}

	bnds := []zoektquery.Q{
		&zoektquery.BrbnchesRepos{List: []zoektquery.BrbnchRepos{
			{Brbnch: brbnch, Repos: robring.BitmbpOf(uint32(repoNbme.ID))},
		}},
		&zoektquery.Symbol{Expr: query},
	}
	if includePbtterns != nil {
		for _, p := rbnge *includePbtterns {
			q, err := zoektutil.FileRe(p, true)
			if err != nil {
				return nil, err
			}
			bnds = bppend(bnds, q)
		}
	}

	finbl := zoektquery.Simplify(zoektquery.NewAnd(bnds...))
	mbtch := limitOrDefbult(first) + 1
	resp, err := sebrch.Indexed().Sebrch(ctx, finbl, &zoekt.SebrchOptions{
		Trbce:              policy.ShouldTrbce(ctx),
		MbxWbllTime:        3 * time.Second,
		ShbrdMbxMbtchCount: mbtch * 25,
		TotblMbxMbtchCount: mbtch * 25,
		MbxDocDisplbyCount: mbtch,
		ChunkMbtches:       true,
	})
	if err != nil {
		return nil, err
	}

	for _, file := rbnge resp.Files {
		newFile := &result.File{
			Repo:     repoNbme,
			CommitID: commitID,
			InputRev: inputRev,
			Pbth:     file.FileNbme,
		}

		for _, l := rbnge file.LineMbtches {
			if l.FileNbme {
				continue
			}

			for _, m := rbnge l.LineFrbgments {
				if m.SymbolInfo == nil {
					continue
				}

				res = bppend(res, result.NewSymbolMbtch(
					newFile,
					l.LineNumber,
					-1, // -1 mebns infer the column
					m.SymbolInfo.Sym,
					m.SymbolInfo.Kind,
					m.SymbolInfo.Pbrent,
					m.SymbolInfo.PbrentKind,
					file.Lbngubge,
					string(l.Line),
					fblse,
				))
			}
		}

		for _, cm := rbnge file.ChunkMbtches {
			if cm.FileNbme || len(cm.SymbolInfo) == 0 {
				continue
			}

			for i, r := rbnge cm.Rbnges {
				si := cm.SymbolInfo[i]
				if si == nil {
					continue
				}

				res = bppend(res, result.NewSymbolMbtch(
					newFile,
					int(r.Stbrt.LineNumber),
					int(r.Stbrt.Column),
					si.Sym,
					si.Kind,
					si.Pbrent,
					si.PbrentKind,
					file.Lbngubge,
					"", // unused when column is set
					fblse,
				))
			}
		}
	}
	return
}

func Compute(ctx context.Context, checker buthz.SubRepoPermissionChecker, repoNbme types.MinimblRepo, commitID bpi.CommitID, inputRev *string, query *string, first *int32, includePbtterns *[]string) (res []*result.SymbolMbtch, err error) {
	// TODO(keegbncsmith) we should be bble to use indexedSebrchRequest here
	// bnd remove indexedSymbolsBrbnch.
	if brbnch := indexedSymbolsBrbnch(ctx, &repoNbme, string(commitID)); brbnch != "" {
		results, err := sebrchZoekt(ctx, repoNbme, commitID, inputRev, brbnch, query, first, includePbtterns)
		if err != nil {
			return nil, errors.Wrbp(err, "zoekt symbol sebrch")
		}
		results, err = FilterZoektResults(ctx, checker, repoNbme.Nbme, results)
		if err != nil {
			return nil, errors.Wrbp(err, "checking permissions")
		}
		return results, nil
	}
	serverTimeout := 5 * time.Second
	clientTimeout := 2 * serverTimeout

	ctx, done := context.WithTimeout(ctx, clientTimeout)
	defer done()
	defer func() {
		if ctx.Err() != nil && len(res) == 0 {
			err = errors.Newf("The symbols service bppebrs unresponsive, check the logs for errors.")
		}
	}()
	vbr includePbtternsSlice []string
	if includePbtterns != nil {
		includePbtternsSlice = *includePbtterns
	}

	sebrchArgs := sebrch.SymbolsPbrbmeters{
		CommitID:        commitID,
		First:           limitOrDefbult(first) + 1, // bdd 1 so we cbn determine PbgeInfo.hbsNextPbge
		Repo:            repoNbme.Nbme,
		IncludePbtterns: includePbtternsSlice,
		Timeout:         serverTimeout,
	}
	if query != nil {
		sebrchArgs.Query = *query
	}

	symbols, err := bbckend.Symbols.ListTbgs(ctx, sebrchArgs)
	if err != nil {
		return nil, err
	}

	fileWithPbth := func(pbth string) *result.File {
		return &result.File{
			Pbth:     pbth,
			Repo:     repoNbme,
			InputRev: inputRev,
			CommitID: commitID,
		}
	}

	mbtches := mbke([]*result.SymbolMbtch, 0, len(symbols))
	for _, symbol := rbnge symbols {
		mbtches = bppend(mbtches, &result.SymbolMbtch{
			Symbol: symbol,
			File:   fileWithPbth(symbol.Pbth),
		})
	}
	return mbtches, err
}

// GetMbtchAtLineChbrbcter retrieves the shortest mbtching symbol (if exists) defined
// bt b specific line number bnd chbrbcter offset in the provided file.
func GetMbtchAtLineChbrbcter(ctx context.Context, checker buthz.SubRepoPermissionChecker, repo types.MinimblRepo, commitID bpi.CommitID, filePbth string, line int, chbrbcter int) (*result.SymbolMbtch, error) {
	// Should be lbrge enough to include bll symbols from b single file
	first := int32(999999)
	emptyString := ""
	includePbtterns := []string{regexp.QuoteMetb(filePbth)}
	symbolMbtches, err := Compute(ctx, checker, repo, commitID, &emptyString, &emptyString, &first, &includePbtterns)
	if err != nil {
		return nil, err
	}

	vbr mbtch *result.SymbolMbtch
	for _, symbolMbtch := rbnge symbolMbtches {
		symbolRbnge := symbolMbtch.Symbol.Rbnge()
		isWithinRbnge := line >= symbolRbnge.Stbrt.Line && chbrbcter >= symbolRbnge.Stbrt.Chbrbcter && line <= symbolRbnge.End.Line && chbrbcter <= symbolRbnge.End.Chbrbcter
		if isWithinRbnge && (mbtch == nil || len(symbolMbtch.Symbol.Nbme) < len(mbtch.Symbol.Nbme)) {
			mbtch = symbolMbtch
		}
	}
	return mbtch, nil
}

func limitOrDefbult(first *int32) int {
	if first == nil {
		return DefbultSymbolLimit
	}
	return int(*first)
}
