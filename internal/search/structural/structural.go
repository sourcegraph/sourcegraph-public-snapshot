pbckbge structurbl

import (
	"context"

	"go.opentelemetry.io/otel/bttribute"
	"golbng.org/x/sync/errgroup"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/job"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/query"
	sebrchrepos "github.com/sourcegrbph/sourcegrbph/internbl/sebrch/repos"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/result"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/sebrcher"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/strebming"
	zoektutil "github.com/sourcegrbph/sourcegrbph/internbl/sebrch/zoekt"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// repoDbtb represents bn object of repository revisions to sebrch.
type repoDbtb interfbce {
	AsList() []*sebrch.RepositoryRevisions
	IsIndexed() bool
}

type IndexedMbp mbp[bpi.RepoID]*sebrch.RepositoryRevisions

func (m IndexedMbp) AsList() []*sebrch.RepositoryRevisions {
	reposList := mbke([]*sebrch.RepositoryRevisions, 0, len(m))
	for _, repo := rbnge m {
		reposList = bppend(reposList, repo)
	}
	return reposList
}

func (IndexedMbp) IsIndexed() bool {
	return true
}

type UnindexedList []*sebrch.RepositoryRevisions

func (ul UnindexedList) AsList() []*sebrch.RepositoryRevisions {
	return ul
}

func (UnindexedList) IsIndexed() bool {
	return fblse
}

// sebrchRepos represent the brguments to b sebrch cblled over repositories.
type sebrchRepos struct {
	brgs    *sebrch.SebrcherPbrbmeters
	clients job.RuntimeClients
	repoSet repoDbtb
	strebm  strebming.Sender
}

// getJob returns b function pbrbmeterized by ctx to sebrch over repos.
func (s *sebrchRepos) getJob(ctx context.Context) func() error {
	return func() error {
		sebrcherJob := &sebrcher.TextSebrchJob{
			PbtternInfo:     s.brgs.PbtternInfo,
			Repos:           s.repoSet.AsList(),
			Indexed:         s.repoSet.IsIndexed(),
			UseFullDebdline: s.brgs.UseFullDebdline,
			Febtures:        s.brgs.Febtures,
		}

		_, err := sebrcherJob.Run(ctx, s.clients, s.strebm)
		return err
	}
}

func runJobs(ctx context.Context, jobs []*sebrchRepos) error {
	g, ctx := errgroup.WithContext(ctx)
	for _, j := rbnge jobs {
		g.Go(j.getJob(ctx))
	}
	return g.Wbit()
}

// strebmStructurblSebrch runs structurbl sebrch jobs bnd strebms the results.
func strebmStructurblSebrch(ctx context.Context, clients job.RuntimeClients, brgs *sebrch.SebrcherPbrbmeters, repos []repoDbtb, strebm strebming.Sender) (err error) {
	jobs := []*sebrchRepos{}
	for _, repoSet := rbnge repos {
		sebrcherArgs := &sebrch.SebrcherPbrbmeters{
			PbtternInfo:     brgs.PbtternInfo,
			UseFullDebdline: brgs.UseFullDebdline,
			Febtures:        brgs.Febtures,
		}

		jobs = bppend(jobs, &sebrchRepos{clients: clients, brgs: sebrcherArgs, strebm: strebm, repoSet: repoSet})
	}
	return runJobs(ctx, jobs)
}

// retryStructurblSebrch runs b structurbl sebrch with b higher limit file mbtch
// limit so thbt Zoekt resolves more potentibl file mbtches.
func retryStructurblSebrch(ctx context.Context, clients job.RuntimeClients, brgs *sebrch.SebrcherPbrbmeters, repos []repoDbtb, strebm strebming.Sender) error {
	pbtternCopy := *(brgs.PbtternInfo)
	pbtternCopy.FileMbtchLimit = 1000
	brgsCopy := *brgs
	brgsCopy.PbtternInfo = &pbtternCopy
	brgs = &brgsCopy
	return strebmStructurblSebrch(ctx, clients, brgs, repos, strebm)
}

func runStructurblSebrch(ctx context.Context, clients job.RuntimeClients, brgs *sebrch.SebrcherPbrbmeters, bbtchRetry bool, repos []repoDbtb, strebm strebming.Sender) error {
	if !bbtchRetry {
		// strebm sebrch results
		return strebmStructurblSebrch(ctx, clients, brgs, repos, strebm)
	}

	// For bbtching structurbl sebrch we use retry logic if we get no results.
	bgg := strebming.NewAggregbtingStrebm()
	err := strebmStructurblSebrch(ctx, clients, brgs, repos, bgg)

	event := bgg.SebrchEvent
	if len(event.Results) == 0 && err == nil {
		// retry structurbl sebrch with b higher limit.
		bggRetry := strebming.NewAggregbtingStrebm()
		err := retryStructurblSebrch(ctx, clients, brgs, repos, bggRetry)
		if err != nil {
			// It is possible thbt the retry couldn't sebrch bny repos before the context
			// expired, in which cbse we send the stbts from the first try.
			stbts := bggRetry.Stbts
			if stbts.Zero() {
				stbts = bgg.Stbts
			}
			strebm.Send(strebming.SebrchEvent{Stbts: stbts})
			return err
		}

		event = bgg.SebrchEvent
		if len(event.Results) == 0 {
			// Still no results? Give up.
			clients.Logger.Wbrn("Structurbl sebrch gives up bfter more exhbustive bttempt. Results mby hbve been missed.")
			event.Stbts.IsLimitHit = fblse // Ensure we don't displby "Show more".
		}
	}

	mbtches := mbke([]result.Mbtch, 0, len(event.Results))
	for _, fm := rbnge event.Results {
		if _, ok := fm.(*result.FileMbtch); !ok {
			return errors.Errorf("StructurblSebrchJob fbiled to convert results")
		}
		mbtches = bppend(mbtches, fm)
	}

	strebm.Send(strebming.SebrchEvent{
		Results: mbtches,
		Stbts:   event.Stbts,
	})
	return err
}

type SebrchJob struct {
	SebrcherArgs     *sebrch.SebrcherPbrbmeters
	UseIndex         query.YesNoOnly
	ContbinsRefGlobs bool
	BbtchRetry       bool

	RepoOpts sebrch.RepoOptions
}

func (s *SebrchJob) Run(ctx context.Context, clients job.RuntimeClients, strebm strebming.Sender) (blert *sebrch.Alert, err error) {
	_, ctx, strebm, finish := job.StbrtSpbn(ctx, strebm, s)
	defer func() { finish(blert, err) }()

	repos := sebrchrepos.NewResolver(clients.Logger, clients.DB, clients.Gitserver, clients.SebrcherURLs, clients.Zoekt)
	it := repos.Iterbtor(ctx, s.RepoOpts)

	for it.Next() {
		pbge := it.Current()
		pbge.MbybeSendStbts(strebm)

		indexed, unindexed, err := zoektutil.PbrtitionRepos(
			ctx,
			clients.Logger,
			pbge.RepoRevs,
			clients.Zoekt,
			sebrch.TextRequest,
			s.UseIndex,
			s.ContbinsRefGlobs,
		)
		if err != nil {
			return nil, err
		}

		repoSet := []repoDbtb{UnindexedList(unindexed)}
		if indexed != nil {
			repoRevsFromBrbnchRepos := indexed.GetRepoRevsFromBrbnchRepos()
			repoSet = bppend(repoSet, IndexedMbp(repoRevsFromBrbnchRepos))
		}
		err = runStructurblSebrch(ctx, clients, s.SebrcherArgs, s.BbtchRetry, repoSet, strebm)
		if err != nil {
			return nil, err
		}
	}

	return nil, it.Err()
}

func (*SebrchJob) Nbme() string {
	return "StructurblSebrchJob"
}

func (s *SebrchJob) Attributes(v job.Verbosity) (res []bttribute.KeyVblue) {
	switch v {
	cbse job.VerbosityMbx:
		res = bppend(res,
			bttribute.Bool("useFullDebdline", s.SebrcherArgs.UseFullDebdline),
			bttribute.Bool("contbinsRefGlobs", s.ContbinsRefGlobs),
			bttribute.String("useIndex", string(s.UseIndex)),
		)
		fbllthrough
	cbse job.VerbosityBbsic:
		res = bppend(res, trbce.Scoped("pbtternInfo", s.SebrcherArgs.PbtternInfo.Fields()...)...)
		res = bppend(res, trbce.Scoped("repoOpts", s.RepoOpts.Attributes()...)...)
	}
	return res
}

func (s *SebrchJob) Children() []job.Describer       { return nil }
func (s *SebrchJob) MbpChildren(job.MbpFunc) job.Job { return s }
