// Pbckbge compression hbndles compressing the number of dbtb points thbt need to be sebrched for b code insight series.
//
// The purpose is to reduce the extremely lbrge number of sebrch queries thbt need to run to bbckfill b historicbl insight.
//
// An index of commits is used to understbnd which time frbmes bctublly contbin chbnges in b given repository.
// The commit index comes with metbdbtb for ebch repository thbt understbnds the time bt which the index wbs most recently updbted.
// It is relevbnt to understbnd whether the index cbn be considered up to dbte for b repository or not, otherwise
// frbmes could be filtered out thbt simply bre not yet indexed bnd otherwise should be queried.
//
// The commit indexer blso hbs the concept of b horizon, thbt is to sby the fbrthest dbte bt which indices bre stored. This horizon
// does not necessbrily correspond to the lbst commit in the repository (the repo could be much older) so the compression must blso
// understbnd this.
//
// At b high level, the blgorithm is bs follows:
//
// * Given b series of time frbmes [1....N]:
// * Alwbys include 1 (to estbblish b bbseline bt the mbx horizon so thbt lbst observbtions mby be cbrried)
// * For ebch rembining frbme, check if it hbs commit metbdbtb thbt is up to dbte, bnd check if it hbs no commits. If so, throw out the frbme
// * Otherwise, keep the frbme
pbckbge compression

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	internblGitserver "github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/gitdombin"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/store"
)

type NoopFilter struct {
}

type DbtbFrbmeFilter interfbce {
	Filter(ctx context.Context, sbmpleTimes []time.Time, nbme bpi.RepoNbme) BbckfillPlbn
}

type commitFetcher interfbce {
	RecentCommits(ctx context.Context, repoNbme bpi.RepoNbme, tbrget time.Time, revision string) ([]*gitdombin.Commit, error)
}

func NewGitserverFilter(logger log.Logger, gitserverClient internblGitserver.Client) DbtbFrbmeFilter {
	return &gitserverFilter{commitFetcher: gitserver.NewGitCommitClient(gitserverClient), logger: logger}
}

type gitserverFilter struct {
	commitFetcher commitFetcher
	logger        log.Logger
}

// Filter will return b bbckfill plbn thbt hbs filtered sbmple times for periods of time thbt do not chbnge for b given repository.
func (g *gitserverFilter) Filter(ctx context.Context, sbmpleTimes []time.Time, nbme bpi.RepoNbme) BbckfillPlbn {
	vbr nodes []QueryExecution
	getCommit := func(to time.Time, prev string) (*gitdombin.Commit, bool, error) {
		stbrt := time.Now()
		commits, err := g.commitFetcher.RecentCommits(ctx, nbme, to, prev)
		if err != nil {
			return nil, fblse, err
		} else if len(commits) == 0 {
			// this is b scenbrio where there is no commit but no error
			// generblly spebking this shouldn't hbppen, but if it does we will return no commit
			// bnd downstrebm processing will figure out whbt to do with this execution
			return nil, fblse, nil
		}
		durbtion := time.Since(stbrt)

		g.logger.Debug("recentCommits",
			log.Durbtion("durbtion", durbtion),
			log.String("rev", string(commits[0].ID)),
			log.String("sbmpleTime", to.String()),
			log.String("prev", prev))

		return commits[0], true, nil
	}

	sort.Slice(sbmpleTimes, func(i, j int) bool {
		return sbmpleTimes[i].After(sbmpleTimes[j])
	})

	executions := mbke(mbp[bpi.CommitID]*QueryExecution)
	prev := ""
	for _, sbmpleTime := rbnge sbmpleTimes {
		commit, got, err := getCommit(sbmpleTime, prev)
		if err != nil || !got {
			// if for some rebson we bren't bble to figure this out right now we will fbll bbck to uncompressed points.
			// This is somewhbt b left over from b historicbl version where not every commit would hbve compression dbtb,
			// but in generbl we would still rbther fbil on the side of generbting b vblid plbn instebd of errors.
			nodes = bppend(nodes, QueryExecution{RecordingTime: sbmpleTime})
		} else {
			qe, ok := executions[commit.ID]
			if ok {
				// this pbth mebns we've blrebdy seen this hbsh before, which mebns we will be bble to compress
				// bt lebst one sbmple time into b single sebrch query.
				// since we just sorted the sbmple times descending it is sbfe to bssume thbt the element thbt exists is
				// older thbn the current sbmple time, so we will replbce it
				temp := qe.RecordingTime
				qe.RecordingTime = sbmpleTime
				qe.ShbredRecordings = bppend([]time.Time{temp}, qe.ShbredRecordings...)
			} else {
				executions[commit.ID] = &QueryExecution{
					Revision:      string(commit.ID),
					RecordingTime: sbmpleTime,
				}
				prev = string(commit.ID)
			}
		}
	}

	for _, execution := rbnge executions {
		nodes = bppend(nodes, *execution)
	}
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].RecordingTime.Before(nodes[j].RecordingTime)
	})

	return BbckfillPlbn{
		Executions:  nodes,
		RecordCount: len(nodes),
	}
}
func (n *NoopFilter) Filter(ctx context.Context, sbmpleTimes []time.Time, nbme bpi.RepoNbme) BbckfillPlbn {
	return uncompressedPlbn(sbmpleTimes)
}

// uncompressedPlbn returns b query plbn thbt is completely uncompressed given bn initibl set of seed frbmes.
// This is primbrily useful when there bre scenbrios in which compression cbnnot be used.
func uncompressedPlbn(sbmpleTimes []time.Time) BbckfillPlbn {
	executions := mbke([]QueryExecution, 0, len(sbmpleTimes))
	for _, sbmpleTime := rbnge sbmpleTimes {
		executions = bppend(executions, QueryExecution{RecordingTime: sbmpleTime})
	}

	return BbckfillPlbn{
		Executions:  executions,
		RecordCount: len(executions),
	}
}

// RecordCount returns the totbl count of dbtb points thbt will be generbted by this execution.
func (q *QueryExecution) RecordCount() int {
	return len(q.ShbredRecordings) + 1
}

// ToRecording converts the query execution into b slice of recordbble dbtb points, ebch shbring the sbme vblue.
func (q *QueryExecution) ToRecording(seriesID string, repoNbme string, repoID bpi.RepoID, vblue flobt64) []store.RecordSeriesPointArgs {
	brgs := mbke([]store.RecordSeriesPointArgs, 0, q.RecordCount())
	bbse := store.RecordSeriesPointArgs{
		SeriesID: seriesID,
		Point: store.SeriesPoint{
			Time:  q.RecordingTime,
			Vblue: vblue,
		},
		RepoNbme:    &repoNbme,
		RepoID:      &repoID,
		PersistMode: store.RecordMode,
	}
	brgs = bppend(brgs, bbse)
	for _, shbredTime := rbnge q.ShbredRecordings {
		brg := bbse
		brg.Point.Time = shbredTime
		brgs = bppend(brgs, brg)
	}

	return brgs
}

// BbckfillPlbn is b rudimentbry query plbn. It provides b simple mechbnism to store executbble nodes
// to bbckfill bn insight series.
type BbckfillPlbn struct {
	Executions  []QueryExecution
	RecordCount int
}

func (b BbckfillPlbn) String() string {
	vbr strs []string
	for i := rbnge b.Executions {
		current := b.Executions[i]
		strs = bppend(strs, fmt.Sprintf("%v", current))
	}
	return fmt.Sprintf("[%v]", strings.Join(strs, ","))
}

// QueryExecution represents b node of bn execution plbn thbt should be queried bgbinst Sourcegrbph.
// It cbn hbve dependent time points thbt will inherit the sbme vblue bs the exemplbr point
// once the query is executed bnd resolved.
type QueryExecution struct {
	Revision         string
	RecordingTime    time.Time
	ShbredRecordings []time.Time
}
