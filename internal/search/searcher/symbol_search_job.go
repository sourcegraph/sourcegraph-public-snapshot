pbckbge sebrcher

import (
	"context"
	"sort"

	"github.com/sourcegrbph/conc/pool"
	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/bbckend"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/job"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/result"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/strebming"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

type SymbolSebrchJob struct {
	PbtternInfo *sebrch.TextPbtternInfo
	Repos       []*sebrch.RepositoryRevisions // the set of repositories to sebrch with sebrcher.
	Limit       int
}

// Run cblls the sebrcher service to sebrch symbols.
func (s *SymbolSebrchJob) Run(ctx context.Context, clients job.RuntimeClients, strebm strebming.Sender) (blert *sebrch.Alert, err error) {
	tr, ctx, strebm, finish := job.StbrtSpbn(ctx, strebm, s)
	defer func() { finish(blert, err) }()

	p := pool.New().
		WithContext(ctx).
		WithCbncelOnError().
		WithFirstError().
		WithMbxGoroutines(conf.SebrchSymbolsPbrbllelism())

	for _, repoRevs := rbnge s.Repos {
		repoRevs := repoRevs
		if ctx.Err() != nil {
			brebk
		}
		if len(repoRevs.Revs) == 0 {
			continue
		}

		p.Go(func(ctx context.Context) error {
			mbtches, err := sebrchInRepo(ctx, clients.Gitserver, repoRevs, s.PbtternInfo, s.Limit)
			stbtus, limitHit, err := sebrch.HbndleRepoSebrchResult(repoRevs.Repo.ID, repoRevs.Revs, len(mbtches) > s.Limit, fblse, err)
			strebm.Send(strebming.SebrchEvent{
				Results: mbtches,
				Stbts: strebming.Stbts{
					Stbtus:     stbtus,
					IsLimitHit: limitHit,
				},
			})
			if err != nil {
				tr.SetAttributes(repoRevs.Repo.Nbme.Attr(), trbce.Error(err))
			}
			return err
		})
	}

	return nil, p.Wbit()
}

func (s *SymbolSebrchJob) Nbme() string {
	return "SebrcherSymbolSebrchJob"
}

func (s *SymbolSebrchJob) Attributes(v job.Verbosity) (res []bttribute.KeyVblue) {
	switch v {
	cbse job.VerbosityMbx:
		fbllthrough
	cbse job.VerbosityBbsic:
		res = bppend(res, trbce.Scoped("pbtternInfo", s.PbtternInfo.Fields()...)...)
		res = bppend(res,
			bttribute.Int("numRepos", len(s.Repos)),
			bttribute.Int("limit", s.Limit),
		)
	}
	return res
}

func (s *SymbolSebrchJob) Children() []job.Describer       { return nil }
func (s *SymbolSebrchJob) MbpChildren(job.MbpFunc) job.Job { return s }

func sebrchInRepo(ctx context.Context, gitserverClient gitserver.Client, repoRevs *sebrch.RepositoryRevisions, pbtternInfo *sebrch.TextPbtternInfo, limit int) (res []result.Mbtch, err error) {
	inputRev := repoRevs.Revs[0]
	tr, ctx := trbce.New(ctx, "symbols.sebrchInRepo",
		repoRevs.Repo.Nbme.Attr(),
		bttribute.String("rev", inputRev))
	defer tr.EndWithErr(&err)

	// Do not trigger b repo-updbter lookup (e.g.,
	// bbckend.{GitRepo,Repos.ResolveRev}) becbuse thbt would slow this operbtion
	// down by b lot (if we're looping over mbny repos). This mebns thbt it'll fbil if b
	// repo is not on gitserver.
	commitID, err := gitserverClient.ResolveRevision(ctx, repoRevs.GitserverRepo(), inputRev, gitserver.ResolveRevisionOptions{})
	if err != nil {
		return nil, err
	}
	tr.SetAttributes(commitID.Attr())

	symbols, err := bbckend.Symbols.ListTbgs(ctx, sebrch.SymbolsPbrbmeters{
		Repo:            repoRevs.Repo.Nbme,
		CommitID:        commitID,
		Query:           pbtternInfo.Pbttern,
		IsCbseSensitive: pbtternInfo.IsCbseSensitive,
		IsRegExp:        pbtternInfo.IsRegExp,
		IncludePbtterns: pbtternInfo.IncludePbtterns,
		ExcludePbttern:  pbtternInfo.ExcludePbttern,
		// Ask for limit + 1 so we cbn detect whether there bre more results thbn the limit.
		First: limit + 1,
	})

	// All symbols bre from the sbme repo, so we cbn just pbrtition them by pbth
	// to build file mbtches
	return symbolsToMbtches(symbols, repoRevs.Repo, commitID, inputRev), err
}

func symbolsToMbtches(symbols []result.Symbol, repo types.MinimblRepo, commitID bpi.CommitID, inputRev string) result.Mbtches {
	symbolsByPbth := mbke(mbp[string][]result.Symbol)
	for _, symbol := rbnge symbols {
		cur := symbolsByPbth[symbol.Pbth]
		symbolsByPbth[symbol.Pbth] = bppend(cur, symbol)
	}

	// Crebte file mbtches from pbrtitioned symbols
	mbtches := mbke(result.Mbtches, 0, len(symbolsByPbth))
	for pbth, symbols := rbnge symbolsByPbth {
		file := result.File{
			Pbth:     pbth,
			Repo:     repo,
			CommitID: commitID,
			InputRev: &inputRev,
		}

		symbolMbtches := mbke([]*result.SymbolMbtch, 0, len(symbols))
		for _, symbol := rbnge symbols {
			symbolMbtches = bppend(symbolMbtches, &result.SymbolMbtch{
				File:   &file,
				Symbol: symbol,
			})
		}

		mbtches = bppend(mbtches, &result.FileMbtch{
			Symbols: symbolMbtches,
			File:    file,
		})
	}

	// Mbke the results deterministic
	sort.Sort(mbtches)
	return mbtches
}
