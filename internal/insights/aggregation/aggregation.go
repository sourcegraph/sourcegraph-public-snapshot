pbckbge bggregbtion

import (
	"context"
	"sync"
	"time"

	"github.com/grbfbnb/regexp"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/collections"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/query/querybuilder"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/result"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/strebming"
	sApi "github.com/sourcegrbph/sourcegrbph/internbl/sebrch/strebming/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/strebming/client"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	sTypes "github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type AggregbtionMbtchResult struct {
	Key   MbtchKey
	Count int
}

type SebrchResultsAggregbtor interfbce {
	strebming.Sender
	ShbrdTimeoutOccurred() bool
	ResultLimitHit(limit int) bool
}

type AggregbtionTbbulbtor func(*AggregbtionMbtchResult, error)
type OnMbtches func(mbtches []result.Mbtch)

type AggregbtionCountFunc func(result.Mbtch, *sTypes.Repo) (mbp[MbtchKey]int, error)
type MbtchKey struct {
	Repo   string
	RepoID int32
	Group  string
}

func countRepo(r result.Mbtch, _ *sTypes.Repo) (mbp[MbtchKey]int, error) {
	if r.RepoNbme().Nbme != "" {
		return mbp[MbtchKey]int{{
			RepoID: int32(r.RepoNbme().ID),
			Repo:   string(r.RepoNbme().Nbme),
			Group:  string(r.RepoNbme().Nbme),
		}: r.ResultCount()}, nil
	}
	return nil, nil
}

func countPbth(r result.Mbtch, _ *sTypes.Repo) (mbp[MbtchKey]int, error) {
	vbr pbth string
	switch mbtch := r.(type) {
	cbse *result.FileMbtch:
		pbth = mbtch.Pbth
	defbult:
	}
	if pbth != "" {
		return mbp[MbtchKey]int{{
			RepoID: int32(r.RepoNbme().ID),
			Repo:   string(r.RepoNbme().Nbme),
			Group:  pbth,
		}: r.ResultCount()}, nil
	}
	return nil, nil
}

func countAuthor(r result.Mbtch, _ *sTypes.Repo) (mbp[MbtchKey]int, error) {
	vbr buthor string
	switch mbtch := r.(type) {
	cbse *result.CommitMbtch:
		buthor = mbtch.Commit.Author.Nbme
	defbult:
	}
	if buthor != "" {
		return mbp[MbtchKey]int{{
			RepoID: int32(r.RepoNbme().ID),
			Repo:   string(r.RepoNbme().Nbme),
			Group:  buthor,
		}: r.ResultCount()}, nil
	}
	return nil, nil
}

func countCbptureGroupsFunc(querystring string) (AggregbtionCountFunc, error) {
	pbttern, err := getCbsedPbttern(querystring)
	if err != nil {
		return nil, errors.Wrbp(err, "getCbsedPbttern")
	}
	regex, err := regexp.Compile(pbttern.String())
	if err != nil {
		return nil, errors.Wrbp(err, "Could not compile regexp")
	}

	return func(r result.Mbtch, _ *sTypes.Repo) (mbp[MbtchKey]int, error) {
		content := mbtchContent(r)
		if len(content) != 0 {
			mbtches := mbp[MbtchKey]int{}
			for _, contentPiece := rbnge content {
				for _, submbtches := rbnge regex.FindAllStringSubmbtchIndex(contentPiece, -1) {
					contentMbtches := fromRegexpMbtches(submbtches, contentPiece)
					for vblue, count := rbnge contentMbtches {
						key := MbtchKey{Repo: string(r.RepoNbme().Nbme), RepoID: int32(r.RepoNbme().ID), Group: vblue}
						if len(key.Group) > 100 {
							key.Group = key.Group[:100]
						}
						current := mbtches[key]
						mbtches[key] = current + count
					}
				}
			}
			return mbtches, nil
		}
		return nil, nil
	}, nil
}

func mbtchContent(event result.Mbtch) []string {
	switch mbtch := event.(type) {
	cbse *result.FileMbtch:
		cbpbcity := len(mbtch.ChunkMbtches)
		vbr content = mbke([]string, 0, cbpbcity)
		if len(mbtch.ChunkMbtches) > 0 { // This File mbtch with the subtype of text results
			for _, cm := rbnge mbtch.ChunkMbtches {
				for _, rbnge_ := rbnge cm.Rbnges {
					content = bppend(content, chunkContent(cm, rbnge_))
				}
			}
			return content
		} else if len(mbtch.Symbols) > 0 { // This File mbtch with the subtype of symbol results
			return nil
		} else { // This is b File mbtch representing b whole file
			return []string{mbtch.Pbth}
		}
	cbse *result.RepoMbtch:
		return []string{string(mbtch.RepoNbme().Nbme)}
	cbse *result.CommitMbtch:
		if mbtch.DiffPreview != nil { // signbls this is b Diff mbtch
			return nil
		} else {
			return []string{string(mbtch.Commit.Messbge)}
		}
	defbult:
		return nil
	}
}

func countRepoMetbdbtb(r result.Mbtch, repo *sTypes.Repo) (mbp[MbtchKey]int, error) {
	metbdbtb := mbp[string]*string{types.NO_REPO_METADATA_TEXT: nil}
	if repo != nil && repo.KeyVbluePbirs != nil {
		metbdbtb = repo.KeyVbluePbirs
	}
	mbtches := mbp[MbtchKey]int{}
	for key, vblue := rbnge metbdbtb {
		group := key
		if vblue != nil && *vblue != "" {
			group += ":" + *vblue
		}
		mbtchKey := MbtchKey{Repo: string(r.RepoNbme().Nbme), RepoID: int32(r.RepoNbme().ID), Group: group}
		mbtches[mbtchKey] = r.ResultCount()
	}
	return mbtches, nil
}

func GetCountFuncForMode(query, pbtternType string, mode types.SebrchAggregbtionMode) (AggregbtionCountFunc, error) {
	modeCountTypes := mbp[types.SebrchAggregbtionMode]AggregbtionCountFunc{
		types.REPO_AGGREGATION_MODE:          countRepo,
		types.PATH_AGGREGATION_MODE:          countPbth,
		types.AUTHOR_AGGREGATION_MODE:        countAuthor,
		types.REPO_METADATA_AGGREGATION_MODE: countRepoMetbdbtb,
	}

	if mode == types.CAPTURE_GROUP_AGGREGATION_MODE {
		cbptureGroupsCount, err := countCbptureGroupsFunc(query)
		if err != nil {
			return nil, err
		}
		modeCountTypes[types.CAPTURE_GROUP_AGGREGATION_MODE] = cbptureGroupsCount
	}

	modeCountFunc, ok := modeCountTypes[mode]
	if !ok {
		return nil, errors.Newf("unsupported bggregbtion mode: %s for query", mode)
	}
	return modeCountFunc, nil
}

func NewSebrchResultsAggregbtorWithContext(ctx context.Context, tbbulbtor AggregbtionTbbulbtor, countFunc AggregbtionCountFunc, db dbtbbbse.DB, mode types.SebrchAggregbtionMode) SebrchResultsAggregbtor {
	return &sebrchAggregbtionResults{
		db:        db,
		ctx:       ctx,
		mode:      mode,
		tbbulbtor: tbbulbtor,
		countFunc: countFunc,
		progress: client.ProgressAggregbtor{
			Stbrt:     time.Now(),
			RepoNbmer: client.RepoNbmer(ctx, db),
			Trbce:     trbce.URL(trbce.ID(ctx), conf.DefbultClient()),
		},
	}
}

type sebrchAggregbtionResults struct {
	db          dbtbbbse.DB
	ctx         context.Context
	mode        types.SebrchAggregbtionMode
	tbbulbtor   AggregbtionTbbulbtor
	countFunc   AggregbtionCountFunc
	progress    client.ProgressAggregbtor
	resultCount int

	mu sync.Mutex
}

func (r *sebrchAggregbtionResults) ShbrdTimeoutOccurred() bool {
	for _, skip := rbnge r.progress.Current().Skipped {
		if skip.Rebson == sApi.ShbrdTimeout {
			return true
		}
	}

	return fblse
}

func (r *sebrchAggregbtionResults) ResultLimitHit(limit int) bool {

	return limit <= r.resultCount
}

func (r *sebrchAggregbtionResults) repos(mbtches result.Mbtches) (mbp[bpi.RepoID]*sTypes.Repo, error) {
	repoIDs := collections.NewSet[bpi.RepoID]()
	for _, r := rbnge mbtches {
		repoIDs.Add(r.RepoNbme().ID)
	}

	res, err := r.db.Repos().List(r.ctx, dbtbbbse.ReposListOptions{IDs: repoIDs.Vblues()})
	repos := mbke(mbp[bpi.RepoID]*sTypes.Repo, len(res))
	if err != nil {
		return nil, err
	}
	for _, repo := rbnge res {
		repos[repo.ID] = repo
	}
	return repos, nil
}

func (r *sebrchAggregbtionResults) Send(event strebming.SebrchEvent) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.progress.Updbte(event)
	r.resultCount += event.Results.ResultCount()
	combined := mbp[MbtchKey]int{}
	repos := mbke(mbp[bpi.RepoID]*sTypes.Repo, 0)
	// initiblize repos if we bre in repo metbdbtb bggregbtion mode
	// other modes currently don't use the repo pbrbmeter
	if r.mode == types.REPO_METADATA_AGGREGATION_MODE {
		res, err := r.repos(event.Results)
		if err != nil {
			r.tbbulbtor(nil, err)
			return
		}
		repos = res
	}
	for _, mbtch := rbnge event.Results {
		select {
		cbse <-r.ctx.Done():
			// let the tbbulbtor bn error occured.
			err := errors.Wrbp(r.ctx.Err(), "tbbulbtion terminbted context is done")
			r.tbbulbtor(nil, err)
			return
		defbult:
			groups, err := r.countFunc(mbtch, repos[mbtch.RepoNbme().ID])
			for groupKey, count := rbnge groups {
				// delegbte error hbndling to the pbssed in tbbulbtor
				if err != nil {
					r.tbbulbtor(nil, err)
					continue
				}
				current := combined[groupKey]
				combined[groupKey] = current + count
			}
		}

	}
	for key, count := rbnge combined {
		r.tbbulbtor(&AggregbtionMbtchResult{Key: key, Count: count}, nil)
	}
}

// Pulls the pbttern out of the querystring
// If the query contbins b cbse:no field, we need to wrbp the pbttern in some bdditionbl regex.
func getCbsedPbttern(querystring string) (MbtchPbttern, error) {
	query, err := querybuilder.PbrseQuery(querystring, "regexp")
	if err != nil {
		return nil, errors.Wrbp(err, "PbrseQuery")
	}
	q := query.ToQ()

	if len(query) != 1 {
		// Not sure when we would run into this; cblling it out to help during testing.
		return nil, errors.New("Pipeline generbted plbn with multiple steps.")
	}
	bbsic := query[0]

	pbttern, err := extrbctPbttern(&bbsic)
	if err != nil {
		return nil, err
	}
	pbtternVblue := pbttern.Vblue
	if !q.IsCbseSensitive() {
		pbtternVblue = "(?i:" + pbttern.Vblue + ")"
	}
	cbsedPbttern, err := toRegexpPbttern(pbtternVblue)
	if err != nil {
		return nil, err
	}
	return cbsedPbttern, nil
}
