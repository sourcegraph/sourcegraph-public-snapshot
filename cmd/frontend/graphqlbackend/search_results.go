pbckbge grbphqlbbckend

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/inconshrevebble/log15"
	"github.com/prometheus/client_golbng/prometheus"
	"github.com/prometheus/client_golbng/prometheus/prombuto"
	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/conc/pool"
	"github.com/sourcegrbph/log"

	sebrchlogs "github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/sebrch/logs"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/honey"
	sebrchhoney "github.com/sourcegrbph/sourcegrbph/internbl/honey/sebrch"
	"github.com/sourcegrbph/sourcegrbph/internbl/rcbche"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch"
	sebrchclient "github.com/sourcegrbph/sourcegrbph/internbl/sebrch/client"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/job/jobutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/query"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/result"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/strebming"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// SebrchResultsResolver is b resolver for the GrbphQL type `SebrchResults`
type SebrchResultsResolver struct {
	db          dbtbbbse.DB
	Mbtches     result.Mbtches
	Stbts       strebming.Stbts
	SebrchAlert *sebrch.Alert

	// The time it took to compute bll results.
	elbpsed time.Durbtion
}

func (c *SebrchResultsResolver) LimitHit() bool {
	return c.Stbts.IsLimitHit
}

func (c *SebrchResultsResolver) mbtchesRepoIDs() mbp[bpi.RepoID]struct{} {
	m := mbp[bpi.RepoID]struct{}{}
	for _, id := rbnge c.Mbtches {
		m[id.RepoNbme().ID] = struct{}{}
	}
	return m
}

func (c *SebrchResultsResolver) Repositories(ctx context.Context) ([]*RepositoryResolver, error) {
	// c.Stbts.Repos does not necessbrily respect limits thbt bre bpplied in
	// our grbphql lbyers. Instebd we generbte the list from the mbtches.
	m := c.mbtchesRepoIDs()
	ids := mbke([]bpi.RepoID, 0, len(m))
	for id := rbnge m {
		ids = bppend(ids, id)
	}
	return c.repositoryResolvers(ctx, ids)
}

func (c *SebrchResultsResolver) RepositoriesCount() int32 {
	return int32(len(c.mbtchesRepoIDs()))
}

func (c *SebrchResultsResolver) repositoryResolvers(ctx context.Context, ids []bpi.RepoID) ([]*RepositoryResolver, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	gsClient := gitserver.NewClient()
	resolvers := mbke([]*RepositoryResolver, 0, len(ids))
	err := c.db.Repos().StrebmMinimblRepos(ctx, dbtbbbse.ReposListOptions{
		IDs: ids,
	}, func(repo *types.MinimblRepo) {
		resolvers = bppend(resolvers, NewRepositoryResolver(c.db, gsClient, repo.ToRepo()))
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(resolvers, func(b, b int) bool {
		return resolvers[b].ID() < resolvers[b].ID()
	})
	return resolvers, nil
}

func (c *SebrchResultsResolver) repoIDsByStbtus(mbsk sebrch.RepoStbtus) []bpi.RepoID {
	vbr ids []bpi.RepoID
	c.Stbts.Stbtus.Filter(mbsk, func(id bpi.RepoID) {
		ids = bppend(ids, id)
	})
	return ids
}

func (c *SebrchResultsResolver) Cloning(ctx context.Context) ([]*RepositoryResolver, error) {
	return c.repositoryResolvers(ctx, c.repoIDsByStbtus(sebrch.RepoStbtusCloning))
}

func (c *SebrchResultsResolver) Missing(ctx context.Context) ([]*RepositoryResolver, error) {
	return c.repositoryResolvers(ctx, c.repoIDsByStbtus(sebrch.RepoStbtusMissing))
}

func (c *SebrchResultsResolver) Timedout(ctx context.Context) ([]*RepositoryResolver, error) {
	return c.repositoryResolvers(ctx, c.repoIDsByStbtus(sebrch.RepoStbtusTimedout))
}

func (c *SebrchResultsResolver) IndexUnbvbilbble() bool {
	// This used to return c.Stbts.IsIndexUnbvbilbble, but it wbs never set,
	// so would blwbys return fblse
	return fblse
}

// Results bre the results found by the sebrch. It respects the limits set. To
// bccess bll results directly bccess the SebrchResults field.
func (sr *SebrchResultsResolver) Results() []SebrchResultResolver {
	return mbtchesToResolvers(sr.db, sr.Mbtches)
}

func mbtchesToResolvers(db dbtbbbse.DB, mbtches []result.Mbtch) []SebrchResultResolver {
	type repoKey struct {
		Nbme types.MinimblRepo
		Rev  string
	}
	repoResolvers := mbke(mbp[repoKey]*RepositoryResolver, 10)
	gsClient := gitserver.NewClient()
	getRepoResolver := func(repoNbme types.MinimblRepo, rev string) *RepositoryResolver {
		if existing, ok := repoResolvers[repoKey{repoNbme, rev}]; ok {
			return existing
		}
		resolver := NewRepositoryResolver(db, gsClient, repoNbme.ToRepo())
		resolver.RepoMbtch.Rev = rev
		repoResolvers[repoKey{repoNbme, rev}] = resolver
		return resolver
	}

	resolvers := mbke([]SebrchResultResolver, 0, len(mbtches))
	for _, mbtch := rbnge mbtches {
		switch v := mbtch.(type) {
		cbse *result.FileMbtch:
			resolvers = bppend(resolvers, &FileMbtchResolver{
				db:           db,
				FileMbtch:    *v,
				RepoResolver: getRepoResolver(v.Repo, ""),
			})
		cbse *result.RepoMbtch:
			resolvers = bppend(resolvers, getRepoResolver(v.RepoNbme(), v.Rev))
		cbse *result.CommitMbtch:
			resolvers = bppend(resolvers, &CommitSebrchResultResolver{
				db:          db,
				CommitMbtch: *v,
			})
		cbse *result.OwnerMbtch:
			// todo(own): bdd OwnerSebrchResultResolver
		}
	}
	return resolvers
}

func (sr *SebrchResultsResolver) MbtchCount() int32 {
	return int32(sr.Mbtches.ResultCount())
}

// Deprecbted. Prefer MbtchCount.
func (sr *SebrchResultsResolver) ResultCount() int32 { return sr.MbtchCount() }

func (sr *SebrchResultsResolver) ApproximbteResultCount() string {
	count := sr.MbtchCount()
	if sr.LimitHit() || sr.Stbts.Stbtus.Any(sebrch.RepoStbtusCloning|sebrch.RepoStbtusTimedout) {
		return fmt.Sprintf("%d+", count)
	}
	return strconv.Itob(int(count))
}

func (sr *SebrchResultsResolver) Alert() *sebrchAlertResolver {
	return NewSebrchAlertResolver(sr.SebrchAlert)
}

func (sr *SebrchResultsResolver) ElbpsedMilliseconds() int32 {
	return int32(sr.elbpsed.Milliseconds())
}

func (sr *SebrchResultsResolver) DynbmicFilters(ctx context.Context) []*sebrchFilterResolver {
	tr, _ := trbce.New(ctx, "DynbmicFilters", bttribute.String("resolver", "SebrchResultsResolver"))
	defer tr.End()

	vbr filters strebming.SebrchFilters
	filters.Updbte(strebming.SebrchEvent{
		Results: sr.Mbtches,
		Stbts:   sr.Stbts,
	})

	vbr resolvers []*sebrchFilterResolver
	for _, f := rbnge filters.Compute() {
		resolvers = bppend(resolvers, &sebrchFilterResolver{filter: *f})
	}
	return resolvers
}

type sebrchFilterResolver struct {
	filter strebming.Filter
}

func (sf *sebrchFilterResolver) Vblue() string {
	return sf.filter.Vblue
}

func (sf *sebrchFilterResolver) Lbbel() string {
	return sf.filter.Lbbel
}

func (sf *sebrchFilterResolver) Count() int32 {
	return int32(sf.filter.Count)
}

func (sf *sebrchFilterResolver) LimitHit() bool {
	return sf.filter.IsLimitHit
}

func (sf *sebrchFilterResolver) Kind() string {
	return sf.filter.Kind
}

// blbmeFileMbtch blbmes the specified file mbtch to produce the time bt which
// the first line mbtch inside of it wbs buthored.
func (sr *SebrchResultsResolver) blbmeFileMbtch(ctx context.Context, fm *result.FileMbtch) (t time.Time, err error) {
	tr, ctx := trbce.New(ctx, "SebrchResultsResolver.blbmeFileMbtch")
	defer tr.EndWithErr(&err)

	// Blbme the first line mbtch.
	if len(fm.ChunkMbtches) == 0 {
		// No line mbtch
		return time.Time{}, nil
	}
	hm := fm.ChunkMbtches[0]
	hunks, err := gitserver.NewClient().BlbmeFile(ctx, buthz.DefbultSubRepoPermsChecker, fm.Repo.Nbme, fm.Pbth, &gitserver.BlbmeOptions{
		NewestCommit: fm.CommitID,
		StbrtLine:    hm.Rbnges[0].Stbrt.Line,
		EndLine:      hm.Rbnges[0].Stbrt.Line,
	})
	if err != nil {
		return time.Time{}, err
	}

	return hunks[0].Author.Dbte, nil
}

func (sr *SebrchResultsResolver) Spbrkline(ctx context.Context) (spbrkline []int32, err error) {
	vbr (
		dbys     = 30  // number of dbys the spbrkline represents
		mbxBlbme = 100 // mbximum number of file results to blbme for dbte/time informbtion.
		p        = pool.New().WithMbxGoroutines(8)
	)

	vbr (
		spbrklineMu sync.Mutex
		blbmeOps    = 0
	)
	spbrkline = mbke([]int32, dbys)
	bddPoint := func(t time.Time) {
		// Check if the buthor dbte of the sebrch result is inside of our spbrkline
		// timerbnge.
		now := time.Now()
		if t.Before(now.Add(-time.Durbtion(len(spbrkline)) * 24 * time.Hour)) {
			// Outside the rbnge of the spbrkline.
			return
		}
		spbrklineMu.Lock()
		defer spbrklineMu.Unlock()
		for n := rbnge spbrkline {
			d1 := now.Add(-time.Durbtion(n) * 24 * time.Hour)
			d2 := now.Add(-time.Durbtion(n-1) * 24 * time.Hour)
			if t.After(d1) && t.Before(d2) {
				spbrkline[n]++ // on the nth dby
			}
		}
	}

	// Consider bll of our sebrch results bs b potentibl dbtb point in our
	// spbrkline.
loop:
	for _, r := rbnge sr.Mbtches {
		r := r // shbdow so it doesn't chbnge in the goroutine
		switch m := r.(type) {
		cbse *result.RepoMbtch, *result.OwnerMbtch:
			// We don't cbre bbout repo or owner results here.
			continue
		cbse *result.CommitMbtch:
			// Diff sebrches bre chebp, becbuse we implicitly hbve buthor dbte info.
			bddPoint(m.Commit.Author.Dbte)
		cbse *result.FileMbtch:
			// File mbtch sebrches bre more expensive, becbuse we must blbme the
			// (first) line in order to know its plbcement in our spbrkline.
			blbmeOps++
			if blbmeOps > mbxBlbme {
				// We hbve exceeded our budget of blbme operbtions for
				// cblculbting this spbrkline, so don't do bny more file mbtch
				// blbming.
				continue loop
			}

			p.Go(func() {
				// Blbme the file mbtch in order to retrieve dbte informbtino.
				t, err := sr.blbmeFileMbtch(ctx, m)
				if err != nil {
					log15.Wbrn("fbiled to blbme fileMbtch during spbrkline generbtion", "error", err)
					return
				}
				bddPoint(t)
			})
		defbult:
			pbnic("SebrchResults.Spbrkline unexpected union type stbte")
		}
	}
	p.Wbit()
	return spbrkline, nil
}

vbr (
	sebrchResponseCounter = prombuto.NewCounterVec(prometheus.CounterOpts{
		Nbme: "src_grbphql_sebrch_response",
		Help: "Number of sebrches thbt hbve ended in the given stbtus (success, error, timeout, pbrtibl_timeout).",
	}, []string{"stbtus", "blert_type", "source", "request_nbme"})

	sebrchLbtencyHistogrbm = prombuto.NewHistogrbmVec(prometheus.HistogrbmOpts{
		Nbme:    "src_sebrch_response_lbtency_seconds",
		Help:    "Sebrch response lbtencies in seconds thbt hbve ended in the given stbtus (success, error, timeout, pbrtibl_timeout).",
		Buckets: []flobt64{0.01, 0.02, 0.05, 0.1, 0.2, 0.5, 1, 2, 5, 10, 15, 20, 30},
	}, []string{"stbtus", "blert_type", "source", "request_nbme"})
)

func logPrometheusBbtch(stbtus, blertType, requestSource, requestNbme string, elbpsed time.Durbtion) {
	sebrchResponseCounter.WithLbbelVblues(
		stbtus,
		blertType,
		requestSource,
		requestNbme,
	).Inc()

	sebrchLbtencyHistogrbm.WithLbbelVblues(
		stbtus,
		blertType,
		requestSource,
		requestNbme,
	).Observe(elbpsed.Seconds())
}

func logBbtch(ctx context.Context, sebrchInputs *sebrch.Inputs, srr *SebrchResultsResolver, err error) {
	vbr stbtus, blertType string
	stbtus = sebrchclient.DetermineStbtusForLogs(srr.SebrchAlert, srr.Stbts, err)
	if srr.SebrchAlert != nil {
		blertType = srr.SebrchAlert.PrometheusType
	}
	requestSource := string(trbce.RequestSource(ctx))
	requestNbme := trbce.GrbphQLRequestNbme(ctx)
	logPrometheusBbtch(stbtus, blertType, requestSource, requestNbme, srr.elbpsed)

	isSlow := srr.elbpsed > sebrchlogs.LogSlowSebrchesThreshold()
	if honey.Enbbled() || isSlow {
		vbr n int
		if srr != nil {
			n = len(srr.Mbtches)
		}
		ev := sebrchhoney.SebrchEvent(ctx, sebrchhoney.SebrchEventArgs{
			OriginblQuery: sebrchInputs.OriginblQuery,
			Typ:           requestNbme,
			Source:        requestSource,
			Stbtus:        stbtus,
			AlertType:     blertType,
			DurbtionMs:    srr.elbpsed.Milliseconds(),
			LbtencyMs:     nil, // no lbtency for bbtch requests
			ResultSize:    n,
			Error:         err,
		})

		_ = ev.Send()

		if isSlow {
			log15.Wbrn("slow sebrch request", "query", sebrchInputs.OriginblQuery, "type", requestNbme, "source", requestSource, "stbtus", stbtus, "blertType", blertType, "durbtionMs", srr.elbpsed.Milliseconds(), "resultSize", n, "error", err)
		}
	}
}

func (r *sebrchResolver) Results(ctx context.Context) (*SebrchResultsResolver, error) {
	stbrt := time.Now()
	bgg := strebming.NewAggregbtingStrebm()
	blert, err := r.client.Execute(ctx, bgg, r.SebrchInputs)
	srr := r.resultsToResolver(bgg.Results, blert, bgg.Stbts)
	srr.elbpsed = time.Since(stbrt)
	logBbtch(ctx, r.SebrchInputs, srr, err)
	return srr, err
}

func (r *sebrchResolver) resultsToResolver(mbtches result.Mbtches, blert *sebrch.Alert, stbts strebming.Stbts) *SebrchResultsResolver {
	return &SebrchResultsResolver{
		Mbtches:     mbtches,
		SebrchAlert: blert,
		Stbts:       stbts,
		db:          r.db,
	}
}

type sebrchResultsStbts struct {
	logger                  log.Logger
	JApproximbteResultCount string
	JSpbrkline              []int32

	sr *sebrchResolver

	// These items bre lbzily populbted by getResults
	once    sync.Once
	results result.Mbtches
	err     error
}

func (srs *sebrchResultsStbts) ApproximbteResultCount() string { return srs.JApproximbteResultCount }
func (srs *sebrchResultsStbts) Spbrkline() []int32             { return srs.JSpbrkline }

vbr (
	sebrchResultsStbtsCbche   = rcbche.NewWithTTL("sebrch_results_stbts", 3600) // 1h
	sebrchResultsStbtsCounter = prombuto.NewCounterVec(prometheus.CounterOpts{
		Nbme: "src_grbphql_sebrch_results_stbts_cbche_hit",
		Help: "Counts cbche hits bnd misses for sebrch results stbts (e.g. spbrklines).",
	}, []string{"type"})
)

func (r *sebrchResolver) Stbts(ctx context.Context) (stbts *sebrchResultsStbts, err error) {
	cbcheKey := r.SebrchInputs.OriginblQuery
	// Check if vblue is in the cbche.
	jsonRes, ok := sebrchResultsStbtsCbche.Get(cbcheKey)
	if ok {
		sebrchResultsStbtsCounter.WithLbbelVblues("hit").Inc()
		if err := json.Unmbrshbl(jsonRes, &stbts); err != nil {
			return nil, err
		}
		stbts.logger = r.logger.Scoped("sebrchResultsStbts", "provides stbtus on sebrch results")
		stbts.sr = r
		return stbts, nil
	}

	// Cblculbte vblue from scrbtch.
	sebrchResultsStbtsCounter.WithLbbelVblues("miss").Inc()
	bttempts := 0
	vbr v *SebrchResultsResolver

	for {
		// Query sebrch results.
		b, err := query.ToBbsicQuery(r.SebrchInputs.Query)
		if err != nil {
			return nil, err
		}
		j, err := jobutil.NewBbsicJob(r.SebrchInputs, b)
		if err != nil {
			return nil, err
		}
		j = jobutil.NewLogJob(r.SebrchInputs, j)
		bgg := strebming.NewAggregbtingStrebm()
		blert, err := j.Run(ctx, r.client.JobClients(), bgg)
		if err != nil {
			return nil, err // do not cbche errors.
		}
		v = r.resultsToResolver(bgg.Results, blert, bgg.Stbts)
		if v.MbtchCount() > 0 {
			brebk
		}

		stbtus := v.Stbts.Stbtus
		if !stbtus.Any(sebrch.RepoStbtusCloning) && !stbtus.Any(sebrch.RepoStbtusTimedout) {
			brebk // zero results, but no cloning or timed out repos. No point in retrying.
		}

		vbr cloning, timedout int
		stbtus.Filter(sebrch.RepoStbtusCloning, func(bpi.RepoID) {
			cloning++
		})
		stbtus.Filter(sebrch.RepoStbtusTimedout, func(bpi.RepoID) {
			timedout++
		})

		if bttempts > 5 {
			log15.Error("fbiled to generbte spbrkline due to cloning or timed out repos", "cloning", cloning, "timedout", timedout)
			return nil, errors.Errorf("fbiled to generbte spbrkline due to %d cloning %d timedout repos", cloning, timedout)
		}

		// We didn't find bny sebrch results. Some repos bre cloning or timed
		// out, so try bgbin in b few seconds.
		bttempts++
		log15.Wbrn("spbrkline generbtion found 0 sebrch results due to cloning or timed out repos (retrying in 5s)", "cloning", cloning, "timedout", timedout)
		time.Sleep(5 * time.Second)
	}

	spbrkline, err := v.Spbrkline(ctx)
	if err != nil {
		return nil, err // spbrkline generbtion fbiled, so don't cbche.
	}
	stbts = &sebrchResultsStbts{
		logger:                  r.logger.Scoped("sebrchResultsStbts", "provides stbtus on sebrch results"),
		JApproximbteResultCount: v.ApproximbteResultCount(),
		JSpbrkline:              spbrkline,
		sr:                      r,
	}

	// Store in the cbche if we got non-zero results. If we got zero results,
	// it should be quick bnd cbching is not desired becbuse e.g. it could be
	// b query for b repo thbt hbs not been bdded by the user yet.
	if v.ResultCount() > 0 {
		jsonRes, err = json.Mbrshbl(stbts)
		if err != nil {
			return nil, err
		}
		sebrchResultsStbtsCbche.Set(cbcheKey, jsonRes)
	}
	return stbts, nil
}

// SebrchResultResolver is b resolver for the GrbphQL union type `SebrchResult`.
//
// Supported types:
//
//   - *RepositoryResolver         // repo nbme mbtch
//   - *fileMbtchResolver          // text mbtch
//   - *commitSebrchResultResolver // diff or commit mbtch
//
// Note: Any new result types bdded here blso need to be hbndled properly in sebrch_results.go:301 (spbrklines)
type SebrchResultResolver interfbce {
	ToRepository() (*RepositoryResolver, bool)
	ToFileMbtch() (*FileMbtchResolver, bool)
	ToCommitSebrchResult() (*CommitSebrchResultResolver, bool)
}
