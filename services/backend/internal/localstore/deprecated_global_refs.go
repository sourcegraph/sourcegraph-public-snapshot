package localstore

import (
	"sort"
	"strings"
	"time"

	"github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus"

	"gopkg.in/gorp.v1"
	"gopkg.in/inconshreveable/log15.v2"

	"context"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/repotrackutil"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/accesscontrol"
)

// dbDefKey DB-maps a DefKey (excluding commit-id) object. We keep
// this in a separate table to reduce duplication in the global_refs
// table (postgresql does not do string interning)
type dbDefKey struct {
	ID       int64  `db:"id"`
	Repo     string `db:"repo"`
	UnitType string `db:"unit_type"`
	Unit     string `db:"unit"`
	Path     string `db:"path"`
}

func init() {
	GraphSchema.Map.AddTableWithName(dbDefKey{}, "def_keys").SetKeys(true, "id").SetUniqueTogether("repo", "unit_type", "unit", "path")

	// deprecatedDBGlobalRef DB-maps a GlobalRef object.
	type deprecatedDBGlobalRef struct {
		DefKeyID  int64 `db:"def_key_id"`
		Repo      string
		File      string
		Positions [2]int
		Count     int
		UpdatedAt *time.Time `db:"updated_at"`
	}
	GraphSchema.Map.AddTableWithName(deprecatedDBGlobalRef{}, "global_refs_new")
	GraphSchema.CreateSQL = append(GraphSchema.CreateSQL,
		`ALTER TABLE global_refs_new ALTER COLUMN positions TYPE integer[] USING positions::integer[];`,
		`CREATE INDEX global_refs_new_def_key_id ON global_refs_new USING btree (def_key_id);`,
		`CREATE INDEX global_refs_new_repo ON global_refs_new USING btree (repo);`,
		`CREATE MATERIALIZED VIEW global_refs_stats AS SELECT def_key_id, count(distinct repo) AS repos, sum(count) AS refs FROM global_refs_new GROUP BY def_key_id;`,
		`CREATE UNIQUE INDEX ON global_refs_stats (def_key_id);`,
	)
	GraphSchema.DropSQL = append(GraphSchema.DropSQL,
		`DROP MATERIALIZED VIEW IF EXISTS global_refs_stats;`,
	)
}

// deprecatedGlobalRefs is a DB-backed implementation of the GlobalRefs
type deprecatedGlobalRefs struct{}

// DeprecatedTotalRefs returns the number of repos referencing the specified repo.
func (g *deprecatedGlobalRefs) DeprecatedTotalRefs(ctx context.Context, repoURI string) (int, error) {
	// Fetch an arbitrary def key for the repo.
	defKeyID, err := graphDBH(ctx).SelectInt("SELECT id FROM def_keys WHERE repo=$1 LIMIT 1", repoURI)
	if err != nil {
		return 0, err
	} else if defKeyID == 0 {
		return 0, nil
	}

	totalRepos, err := g.getRefStats(ctx, defKeyID)
	return int(totalRepos), err
}

// Get returns the names and ref counts of all repos and files within those repos
// that refer the given def.
func (g *deprecatedGlobalRefs) DeprecatedGet(ctx context.Context, op *sourcegraph.DeprecatedDefsListRefLocationsOp) (*sourcegraph.DeprecatedRefLocationsList, error) {
	if Mocks.DeprecatedGlobalRefs.DeprecatedGet != nil {
		return Mocks.DeprecatedGlobalRefs.DeprecatedGet(ctx, op)
	}

	defRepo, err := (&repos{}).Get(ctx, op.Def.Repo)
	if err != nil {
		return nil, err
	}
	defRepoPath := defRepo.URI

	trackedRepo := repotrackutil.GetTrackedRepo(defRepoPath)
	observe := func(part string, start time.Time) {
		deprecatedGlobalRefsDuration.WithLabelValues(trackedRepo, part).Observe(time.Since(start).Seconds())
	}
	defer observe("total", time.Now())

	opt := op.Opt
	if opt == nil {
		opt = &sourcegraph.DeprecatedDefListRefLocationsOptions{}
	}

	// Optimization: All our SQL operations rely on the defKeyID. Fetch
	// it once, instead of once per query
	start := time.Now()
	defKeyID, err := graphDBH(ctx).SelectInt(
		"SELECT id FROM def_keys WHERE repo=$1 AND unit_type=$2 AND unit=$3 AND path=$4",
		defRepoPath, op.Def.UnitType, op.Def.Unit, op.Def.Path)
	observe("def_keys", start)
	start = time.Now()
	if err != nil {
		return nil, err
	} else if defKeyID == 0 {
		// DefKey was not found
		return &sourcegraph.DeprecatedRefLocationsList{RepoRefs: []*sourcegraph.DeprecatedDefRepoRef{}}, nil
	}

	// Optimization: fetch ref stats in parallel to fetching ref locations.
	var totalRepos int64
	statsDone := make(chan error)
	go func() {
		var err error
		statsStart := time.Now()
		totalRepos, err = g.getRefStats(ctx, defKeyID)
		observe("stats", statsStart)
		statsDone <- err
	}()

	// dbRefLocationsResult holds the result of the SELECT query for fetching global refs.
	type dbRefLocationsResult struct {
		Repo      string
		RepoCount int `db:"repo_count"`
		File      string
		Positions *pq.Int64Array
		Count     int
	}

	var args []interface{}
	arg := func(a interface{}) string {
		v := gorp.PostgresDialect{}.BindVar(len(args))
		args = append(args, a)
		return v
	}

	var sql string
	innerSelectSQL := `SELECT repo, file, positions, count FROM global_refs_new`
	innerSelectSQL += ` WHERE def_key_id=` + arg(defKeyID)
	if len(opt.Repos) > 0 {
		repoBindVars := make([]string, len(opt.Repos))
		for i, r := range opt.Repos {
			repoBindVars[i] = arg(r)
		}
		innerSelectSQL += " AND repo in (" + strings.Join(repoBindVars, ",") + ")"
	}
	innerSelectSQL += " LIMIT 65536" // TODO is this a sufficient/sane limit?

	sql = "SELECT repo, SUM(count) OVER(PARTITION BY repo) AS repo_count, file, positions, count FROM (" + innerSelectSQL + ") res"
	orderBySQL := " ORDER BY repo=" + arg(defRepoPath) + " DESC, repo_count DESC, count DESC"
	sql += orderBySQL
	sql += " LIMIT 512"

	var dbRefResult []*dbRefLocationsResult
	if _, err := graphDBH(ctx).Select(&dbRefResult, sql, args...); err != nil {
		return nil, err
	}

	// repoRefs holds the ordered list of repos referencing this def. The list is sorted by
	// decreasing ref counts per repo, and the file list in each individual DefRepoRef is
	// also sorted by descending ref counts.
	var repoRefs []*sourcegraph.DeprecatedDefRepoRef
	defRepoIdx := -1
	// refsByRepo groups each referencing file by repo.
	refsByRepo := make(map[string]*sourcegraph.DeprecatedDefRepoRef)
	missingRepos := make(map[string]struct{})
	for _, r := range dbRefResult {
		if _, ok := missingRepos[r.Repo]; ok {
			continue
		}

		if _, ok := refsByRepo[r.Repo]; !ok {
			// HACK: check if the repo really exists in the DB or not. This is
			// because some number of repos in the table do not exist in the DB
			// (are Go import paths accidently inserted somehow), so the later
			// VerifyUserHasReadAccessToDefRepoRefs will outright fail due to the
			// repos not existing
			if _, err := Repos.getByURI(ctx, r.Repo); err != nil {
				log15.Warn("GlobalRefs.Get found missing repo", "repo", r.Repo)
				missingRepos[r.Repo] = struct{}{}
				continue
			}

			refsByRepo[r.Repo] = &sourcegraph.DeprecatedDefRepoRef{
				Repo:  r.Repo,
				Count: int32(r.RepoCount),
			}
			repoRefs = append(repoRefs, refsByRepo[r.Repo])
			// Note the position of the def's own repo in the slice.
			if defRepoPath == r.Repo {
				defRepoIdx = len(repoRefs) - 1
			}
		}
		if r.File != "" && r.Count != 0 {
			var pos []int64
			if r.Positions != nil {
				pos = []int64(*r.Positions)
			}
			refsByRepo[r.Repo].Files = append(refsByRepo[r.Repo].Files, &sourcegraph.DeprecatedDefFileRef{
				Path:      r.File,
				Positions: deprecatedDeinterlacePositions(pos),
				Count:     int32(r.Count),
			})
		}
	}

	// Place the def's own repo at the head of the slice, if it exists in the
	// slice and is not at the head already.
	if defRepoIdx > 0 {
		repoRefs[0], repoRefs[defRepoIdx] = repoRefs[defRepoIdx], repoRefs[0]
	}

	observe("locations", start)
	start = time.Now()

	// SECURITY: filter private repos user doesn't have access to.
	repoRefs, err = accesscontrol.VerifyUserHasReadAccessToDefRepoRefs(ctx, "GlobalRefs.Get", repoRefs)
	if err != nil {
		return nil, err
	}
	observe("access", start)

	// Return Files in a consistent order
	for _, r := range repoRefs {
		sort.Sort(deprecatedDefFileRefByScore(r.Files))
	}

	select {
	case err := <-statsDone:
		if err != nil {
			return nil, err
		}
	}

	return &sourcegraph.DeprecatedRefLocationsList{
		RepoRefs:   repoRefs,
		TotalRepos: int32(totalRepos),
	}, nil
}

// deprecatedDeinterlacePositions deinterlaces the interlaced [line, column] slice p into
// their non-interlaced form. We store the positions in the DB interlaced
// because github.com/lib/pq does not support multidimensional array types
// (and implementing this is tedious). Thus they are stored interlaced, i.e.
//
//  [line0, col0, line1, col1, line2, col2]
//
func deprecatedDeinterlacePositions(p []int64) (out []*sourcegraph.DeprecatedFilePosition) {
	if len(p)%2 != 0 {
		panic("deprecatedDeinterlacePositions: unequal length array (bad data?)")
	}
	for i := 0; i < len(p); i += 2 {
		out = append(out, &sourcegraph.DeprecatedFilePosition{
			Line:   int32(p[i]),
			Column: int32(p[i+1]),
		})
	}
	return
}

type deprecatedDefFileRefByScore []*sourcegraph.DeprecatedDefFileRef

func (v deprecatedDefFileRefByScore) Len() int      { return len(v) }
func (v deprecatedDefFileRefByScore) Swap(i, j int) { v[i], v[j] = v[j], v[i] }
func (v deprecatedDefFileRefByScore) Less(i, j int) bool {
	if v[i].Score != v[j].Score {
		return v[i].Score > v[j].Score
	}
	if v[i].Count != v[j].Count {
		return v[i].Count > v[j].Count
	}
	return v[i].Path < v[j].Path
}

// getRefStats fetches global ref aggregation stats pagination and display
// purposes.
func (g *deprecatedGlobalRefs) getRefStats(ctx context.Context, defKeyID int64) (int64, error) {
	// Our strategy is to defer to the potentially stale materialized view
	// if there are a large number of distinct repos. Otherwise we can
	// calculate the exact value since it should be fast to do
	count, err := graphDBH(ctx).SelectInt("SELECT repos FROM global_refs_stats WHERE def_key_id=$1", defKeyID)
	if err != nil {
		return 0, err
	}
	if count > 1000 {
		return count, nil
	}

	return graphDBH(ctx).SelectInt("SELECT COUNT(DISTINCT repo) AS Repos FROM global_refs_new WHERE def_key_id=$1", defKeyID)
}

var deprecatedGlobalRefsDuration = prometheus.NewSummaryVec(prometheus.SummaryOpts{
	Namespace: "src",
	Subsystem: "global_refs",
	Name:      "duration_seconds",
	Help:      "Duration for querying global_refs_new",
	MaxAge:    time.Hour,
}, []string{"repo", "part"})

func init() {
	prometheus.MustRegister(deprecatedGlobalRefsDuration)
}

type DeprecatedMockGlobalRefs struct {
	DeprecatedGet func(ctx context.Context, op *sourcegraph.DeprecatedDefsListRefLocationsOp) (*sourcegraph.DeprecatedRefLocationsList, error)
}
