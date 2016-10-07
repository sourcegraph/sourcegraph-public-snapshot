package localstore

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus"

	"gopkg.in/gorp.v1"
	"gopkg.in/inconshreveable/log15.v2"

	"context"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/dbutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/inventory/filelang"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/langp"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/repotrackutil"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/accesscontrol"
	"sourcegraph.com/sourcegraph/srclib/graph"
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

	// dbGlobalRef DB-maps a GlobalRef object.
	type dbGlobalRef struct {
		DefKeyID  int64 `db:"def_key_id"`
		Repo      string
		File      string
		Positions [2]int
		Count     int
		UpdatedAt *time.Time `db:"updated_at"`
	}
	GraphSchema.Map.AddTableWithName(dbGlobalRef{}, "global_refs_new")
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

	type dbGlobalRefVersion struct {
		Repo      string     `db:"repo"`
		CommitID  string     `db:"commit_id"`
		UpdatedAt *time.Time `db:"updated_at"`
	}
	GraphSchema.Map.AddTableWithName(dbGlobalRefVersion{}, "global_refs_version").SetKeys(false, "repo")
}

// globalRefs is a DB-backed implementation of the GlobalRefs
type globalRefs struct{}

// Get returns the names and ref counts of all repos and files within those repos
// that refer the given def.
func (g *globalRefs) Get(ctx context.Context, op *sourcegraph.DefsListRefLocationsOp) (*sourcegraph.RefLocationsList, error) {
	if Mocks.GlobalRefs.Get != nil {
		return Mocks.GlobalRefs.Get(ctx, op)
	}

	defRepo, err := (&repos{}).Get(ctx, op.Def.Repo)
	if err != nil {
		return nil, err
	}
	defRepoPath := defRepo.URI

	trackedRepo := repotrackutil.GetTrackedRepo(defRepoPath)
	observe := func(part string, start time.Time) {
		globalRefsDuration.WithLabelValues(trackedRepo, part).Observe(time.Since(start).Seconds())
	}
	defer observe("total", time.Now())

	opt := op.Opt
	if opt == nil {
		opt = &sourcegraph.DefListRefLocationsOptions{}
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
		return &sourcegraph.RefLocationsList{RepoRefs: []*sourcegraph.DefRepoRef{}}, nil
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
	var repoRefs []*sourcegraph.DefRepoRef
	defRepoIdx := -1
	// refsByRepo groups each referencing file by repo.
	refsByRepo := make(map[string]*sourcegraph.DefRepoRef)
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

			refsByRepo[r.Repo] = &sourcegraph.DefRepoRef{
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
			refsByRepo[r.Repo].Files = append(refsByRepo[r.Repo].Files, &sourcegraph.DefFileRef{
				Path:      r.File,
				Positions: deinterlacePositions(pos),
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
		sort.Sort(defFileRefByScore(r.Files))
	}

	select {
	case err := <-statsDone:
		if err != nil {
			return nil, err
		}
	}

	return &sourcegraph.RefLocationsList{
		RepoRefs:   repoRefs,
		TotalRepos: int32(totalRepos),
	}, nil
}

// deinterlacePositions deinterlaces the interlaced [line, column] slice p into
// their non-interlaced form. We store the positions in the DB interlaced
// because github.com/lib/pq does not support multidimensional array types
// (and implementing this is tedious). Thus they are stored interlaced, i.e.
//
//  [line0, col0, line1, col1, line2, col2]
//
func deinterlacePositions(p []int64) (out []*sourcegraph.FilePosition) {
	if len(p)%2 != 0 {
		panic("deinterlacePositions: unequal length array (bad data?)")
	}
	for i := 0; i < len(p); i += 2 {
		out = append(out, &sourcegraph.FilePosition{
			Line:   int32(p[i]),
			Column: int32(p[i+1]),
		})
	}
	return
}

type defFileRefByScore []*sourcegraph.DefFileRef

func (v defFileRefByScore) Len() int      { return len(v) }
func (v defFileRefByScore) Swap(i, j int) { v[i], v[j] = v[j], v[i] }
func (v defFileRefByScore) Less(i, j int) bool {
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
func (g *globalRefs) getRefStats(ctx context.Context, defKeyID int64) (int64, error) {
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

// Update takes the graph output of a repo at the latest commit and
// updates the set of refs in the global ref store that originate from
// it.
func (g *globalRefs) Update(ctx context.Context, op RefreshIndexOp) error {
	if Mocks.GlobalRefs.Update != nil {
		return Mocks.GlobalRefs.Update(ctx, op)
	}

	if err := accesscontrol.VerifyUserHasWriteAccess(ctx, "GlobalRefs.Update", op.Repo); err != nil {
		return err
	}
	dbh := graphDBH(ctx)

	commitID := op.CommitID
	repoObj, err := Repos.Get(ctx, op.Repo)
	if err != nil {
		return err
	}

	repo := repoObj.URI
	trackedRepo := repotrackutil.GetTrackedRepo(repo)
	observe := func(part string, start time.Time) {
		// We also add logs because update operations are relatively
		// slow and low volume compared to queries
		d := time.Since(start)
		log15.Debug("TRACE GlobalRefs.Update", "repo", repo, "part", part, "duration", d)
		globalRefsUpdateDuration.WithLabelValues(trackedRepo, part).Observe(d.Seconds())
	}
	defer observe("total", time.Now())

	start := time.Now()
	oldCommitID, err := g.version(dbh, repo)
	observe("version", start)
	if err != nil {
		return err
	}
	if commitID == oldCommitID {
		log15.Debug("GlobalRefs.Update re-indexing commit", "repo", repo, "commitID", commitID)
	}

	start = time.Now()
	external, err := langp.DefaultClient.ExternalRefs(ctx, &langp.RepoRev{Repo: repo, Commit: commitID})
	observe("langp", start)
	if err != nil {
		return err
	}
	refs := make([]*langp.Ref, 0, len(external.Refs))
	for _, r := range external.Refs {
		if shouldIndexRef(r) {
			refs = append(refs, r)
		}
	}

	log15.Info("GlobalRefs.Update", "repo", repo, "commitID", commitID, "oldCommitID", oldCommitID, "numRefs", len(refs))

	if len(refs) == 0 {
		return fmt.Errorf("no indexable refs for %s %s", repo, commitID)
	}

	// Get all defKeyIDs outside of the transaction, since doing it inside
	// of the transaction can lead to conflicts with other imports
	start = time.Now()
	defKeyIDs := map[graph.DefKey]int64{}
	for _, r := range refs {
		k := graph.DefKey{Repo: r.Def.Repo, UnitType: r.Def.UnitType, Unit: r.Def.Unit, Path: r.Def.Path}
		defKeyIDs[k] = -1
	}
	getOrInsertDefKeys(ctx, dbh, defKeyIDs)
	observe("defkeys", start)

	// Create a list of ref positions by file.
	type fileDefKey struct {
		graph.DefKey
		File string
	}
	posByFileDefKey := make(map[fileDefKey][]int)
	for _, r := range refs {
		k := fileDefKey{
			DefKey: graph.DefKey{Repo: r.Def.Repo, UnitType: r.Def.UnitType, Unit: r.Def.Unit, Path: r.Def.Path},
			File:   r.File,
		}
		if len(posByFileDefKey[k]) > 10 { // Only keep 10 ref positions to avoid pollution
			continue
		}
		posByFileDefKey[k] = append(posByFileDefKey[k], r.Line)
		posByFileDefKey[k] = append(posByFileDefKey[k], r.Column)
	}

	tmpCreateSQL := `CREATE TEMPORARY TABLE global_refs_tmp (
	def_key_id bigint,
	repo TEXT,
	file TEXT,
	positions integer[],
	count integer default 1
)
ON COMMIT DROP;`
	finalDeleteSQL := `DELETE FROM global_refs_new WHERE repo=$1;`
	finalInsertSQL := `INSERT INTO global_refs_new(def_key_id, repo, file, positions, count, updated_at)
	SELECT def_key_id, repo, file, positions, sum(count) as count, now() as updated_at
	FROM global_refs_tmp
	GROUP BY def_key_id, repo, file, positions;`

	// Do actual update in one transaction, to ensure we don't have concurrent
	// updates to repo
	defer observe("transaction", time.Now())
	return dbutil.Transact(dbh, func(tx gorp.SqlExecutor) error {
		// Create a temporary table to load all new ref data.
		if _, err := tx.Exec(tmpCreateSQL); err != nil {
			return err
		}

		stmt, err := dbutil.Prepare(tx, pq.CopyIn("global_refs_tmp", "def_key_id", "repo", "file", "positions"))
		if err != nil {
			return fmt.Errorf("global_refs_tmp prepare failed: %s", err)
		}

		// Insert refs into temporary table
		start = time.Now()
		for _, r := range refs {
			defKey := graph.DefKey{Repo: r.Def.Repo, UnitType: r.Def.UnitType, Unit: r.Def.Unit, Path: r.Def.Path}
			defKeyID := defKeyIDs[defKey]
			pos := posByFileDefKey[fileDefKey{DefKey: defKey, File: r.File}]
			if _, err := stmt.Exec(defKeyID, repo, r.File, pq.Array(pos)); err != nil {
				return fmt.Errorf("global_refs_tmp stmt insert failed: %s", err)
			}
		}

		// We need to do an empty Exec() to flush any remaining
		// inserts that are in the drivers buffer
		if _, err = stmt.Exec(); err != nil {
			return fmt.Errorf("global_refs_tmp stmt final insert failed: %s", err)
		}
		observe("copyin", start)

		// Purge all existing ref data for files in this source unit.
		start = time.Now()
		if _, err := tx.Exec(finalDeleteSQL, repo); err != nil {
			return err
		}
		observe("purge", start)

		// Insert refs into global refs table
		start = time.Now()
		if _, err := tx.Exec(finalInsertSQL); err != nil {
			return err
		}
		observe("insert", start)

		// Update version row again, to have a more relevant timestamp
		start = time.Now()
		if err := g.versionUpdate(tx, repo, commitID); err != nil {
			return err
		}
		observe("versionupdate", start)
		return nil
	})
}

// version returns the commit_id that global_refs has indexed for repo
func (g *globalRefs) version(tx gorp.SqlExecutor, repoPath string) (string, error) {
	commitID, err := tx.SelectNullStr("SELECT commit_id FROM global_refs_version WHERE repo=$1 FOR UPDATE", repoPath)
	if err != nil {
		return "", err
	}
	// Insert a value into the table so we just have to run an UPDATE + can lock the row.
	if !commitID.Valid {
		_, err = tx.Exec("INSERT INTO global_refs_version VALUES ($1, '', now())", repoPath)
		if err != nil {
			return "", err
		}
	}
	return commitID.String, nil
}

// versionUpdate will update the version for repo to commitID in the transaction tx
func (g *globalRefs) versionUpdate(tx gorp.SqlExecutor, repoPath string, commitID string) error {
	_, err := tx.Exec("UPDATE global_refs_version SET commit_id=$1, updated_at=now() WHERE repo=$2;", commitID, repoPath)
	return err
}

// StatRefresh refreshes the global_refs_stats tables. This should ONLY be called from test code.
func (g *globalRefs) StatRefresh(ctx context.Context) error {
	_, err := graphDBH(ctx).Exec("REFRESH MATERIALIZED VIEW CONCURRENTLY global_refs_stats;")
	return err
}

// shouldIndexRef helps filter out refs we do not want in the index.
func shouldIndexRef(r *langp.Ref) bool {
	// Ignore vendored refs.
	if filelang.IsVendored(r.File, false) {
		return false
	}
	// Ignore ref to builtin defs of golang/go repo (string, int, bool, etc) as this
	// doesn't add significant value; yet it adds up to a lot of space in the db,
	// and queries for refs of builtin defs take long to finish.
	if r.Def.UnitType == "GoPackage" && r.Def.Repo == "github.com/golang/go" && r.Def.Unit == "builtin" {
		return false
	}
	return true
}

var globalRefsDuration = prometheus.NewSummaryVec(prometheus.SummaryOpts{
	Namespace: "src",
	Subsystem: "global_refs",
	Name:      "duration_seconds",
	Help:      "Duration for querying global_refs_new",
	MaxAge:    time.Hour,
}, []string{"repo", "part"})
var globalRefsUpdateDuration = prometheus.NewSummaryVec(prometheus.SummaryOpts{
	Namespace: "src",
	Subsystem: "global_refs",
	Name:      "update_duration_seconds",
	Help:      "Duration for updating global_refs_new",
	MaxAge:    time.Hour,
}, []string{"repo", "part"})

func init() {
	prometheus.MustRegister(globalRefsDuration)
	prometheus.MustRegister(globalRefsUpdateDuration)
}

type MockGlobalRefs struct {
	Get    func(ctx context.Context, op *sourcegraph.DefsListRefLocationsOp) (*sourcegraph.RefLocationsList, error)
	Update func(ctx context.Context, op RefreshIndexOp) error
}
