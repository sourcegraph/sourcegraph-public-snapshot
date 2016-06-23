package localstore

import (
	"fmt"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"gopkg.in/gorp.v1"
	"gopkg.in/inconshreveable/log15.v2"

	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/repotrackutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/search"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/store"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/accesscontrol"
	"sourcegraph.com/sourcegraph/srclib/graph"
	sstore "sourcegraph.com/sourcegraph/srclib/store"
	"sourcegraph.com/sqs/pbtypes"
)

// dbLang is a numerical identifier that identifies the language of a definition
// in the database. NOTE: this values of existing dbLang constants should NOT be
// changed. Doing so would require a database migration.
type dbLang uint16

const (
	dbLangGo     = 1
	dbLangJava   = 2
	dbLangPython = 3
)

var toDBLang_ = map[string]dbLang{
	"go":     dbLangGo,
	"java":   dbLangJava,
	"python": dbLangPython,
}

func toDBLang(lang string) (dbLang, error) {
	if l, exists := toDBLang_[lang]; exists {
		return l, nil
	}
	return 0, fmt.Errorf("unrecognized language %s", lang)
}

func init() {
	fields := []string{"Repo", "CommitID", "UnitType", "Unit", "Path"}
	GraphSchema.Map.AddTableWithName(dbGlobalDef{}, "global_defs").SetKeys(false, fields...)
	for _, table := range []string{"global_defs"} {
		GraphSchema.CreateSQL = append(GraphSchema.CreateSQL,
			`ALTER TABLE `+table+` ALTER COLUMN updated_at TYPE timestamp with time zone USING updated_at::timestamp with time zone;`,
			`ALTER TABLE `+table+` ALTER COLUMN ref_ct SET DEFAULT 0;`,
			`ALTER TABLE `+table+` ALTER COLUMN language TYPE smallint`,
			`CREATE INDEX `+table+`_bow_idx ON `+table+` USING gin(to_tsvector('english', bow));`,
			`CREATE INDEX `+table+`_doc_idx ON `+table+` USING gin(to_tsvector('english', doc));`,
			`CREATE INDEX `+table+`_name ON `+table+` USING btree (lower(name));`,
			`CREATE INDEX `+table+`_name_hi_ref_ct ON `+table+` USING btree (lower(name)) WHERE ref_ct >= 3;`,
			`CREATE INDEX `+table+`_repo ON `+table+` USING btree (repo text_pattern_ops);`,
			`CREATE INDEX `+table+`_updater ON `+table+` USING btree (repo, unit_type, unit, path);`,
		)
	}
}

type dbGlobalDefLanguages struct {
	ID       int16  `db:"id"`
	Language string `db:"language"`
}

// dbGlobalDef DB-maps a GlobalDef object.
type dbGlobalDef struct {
	Repo     string `db:"repo"`
	CommitID string `db:"commit_id"`
	UnitType string `db:"unit_type"`
	Unit     string `db:"unit"`
	Path     string `db:"path"`

	Name     string `db:"name"`
	Kind     string `db:"kind"`
	File     string `db:"file"`
	Language int16  `db:"language"`

	RefCount  int        `db:"ref_ct"`
	UpdatedAt *time.Time `db:"updated_at"`

	Data []byte `db:"data"`

	BoW string `db:"bow"`
	Doc string `db:"doc"`
}

func fromDBDef(d *dbGlobalDef) *sourcegraph.Def {
	if d == nil {
		return nil
	}

	var data pbtypes.RawMessage
	data.Unmarshal(d.Data)
	def := &sourcegraph.Def{
		Def: graph.Def{
			DefKey: graph.DefKey{
				Repo:     d.Repo,
				CommitID: d.CommitID,
				UnitType: d.UnitType,
				Unit:     d.Unit,
				Path:     d.Path,
			},

			Name: d.Name,
			Kind: d.Kind,
			File: d.File,

			Data: data,
		},
	}
	if d.Doc != "" {
		def.Docs = []*graph.DefDoc{{Format: "text/plain", Data: d.Doc}}
	}
	return def
}

func toDBDef(d *sourcegraph.Def) *dbGlobalDef {
	if d == nil {
		return nil
	}
	data, err := d.Data.Marshal()
	if err != nil {
		data = []byte{}
	}
	return &dbGlobalDef{
		Repo:     d.Repo,
		UnitType: d.UnitType,
		Unit:     d.Unit,
		Path:     d.Path,

		Name: d.Name,
		Kind: d.Kind,
		File: d.File,

		Data: data,
	}
}

// dbGlobalSearchResult holds the result of the SELECT query for global def search.
type dbGlobalSearchResult struct {
	dbGlobalDef

	Score float64 `db:"score"`
}

// globalDefs is a DB-backed implementation of the GlobalDefs store.
type globalDefs struct{}

var _ store.GlobalDefs = (*globalDefs)(nil)

func (g *globalDefs) Search(ctx context.Context, op *store.GlobalDefSearchOp) (*sourcegraph.SearchResultsList, error) {
	var args []interface{}
	arg := func(v interface{}) string {
		args = append(args, v)
		return gorp.PostgresDialect{}.BindVar(len(args) - 1)
	}

	if op.Opt == nil {
		op.Opt = &sourcegraph.SearchOptions{}
	}

	if len(op.TokQuery) == 0 && len(op.Opt.Repos) == 0 {
		return &sourcegraph.SearchResultsList{}, nil
	}

	bowQuery := search.UserQueryToksToTSQuery(op.TokQuery)
	lastTok := ""
	if len(op.TokQuery) > 0 {
		lastTok = op.TokQuery[len(op.TokQuery)-1]
	}

	var scoreSQL string
	if bowQuery != "" {
		// The ranking critieron is the weighted sum of xref count,
		// text similarity score, and whether the last term matches
		// the name.
		scoreSQL = `5.0*log(10 + ref_ct) + 100.0*ts_rank(to_tsvector('english', bow), to_tsquery('english', ` + arg(bowQuery) + `)) + 100.0*((LOWER(name)=LOWER(` + arg(lastTok) + `))::int) score`
	} else {
		scoreSQL = `ref_ct score`
	}
	selectSQL := `SELECT repo, commit_id, unit_type, unit, path, name, kind, file, data, doc, ref_ct, ` + scoreSQL + ` FROM global_defs`
	var whereSQL, prefixSQL string
	{
		var wheres []string

		// Repository filtering.
		if len(op.Opt.Repos) > 0 {
			reposURIs, err := repoURIs(ctx, op.Opt.Repos)
			if err != nil {
				return nil, err
			}
			var r []string
			for _, repo := range reposURIs {
				r = append(r, arg(repo))
			}
			wheres = append(wheres, `repo IN (`+strings.Join(r, ", ")+`)`)
		}
		if len(op.Opt.NotRepos) > 0 {
			notReposURIs, err := repoURIs(ctx, op.Opt.NotRepos)
			if err != nil {
				return nil, err
			}
			var r []string
			for _, repo := range notReposURIs {
				r = append(r, arg(repo))
			}
			wheres = append(wheres, `repo NOT IN (`+strings.Join(r, ", ")+`)`)
		}

		// Language filtering.
		if len(op.Opt.Languages) > 0 {
			var l []string
			for _, language := range op.Opt.Languages {
				id, err := toDBLang(language)
				if err != nil {
					continue
				}
				l = append(l, arg(id))
			}
			wheres = append(wheres, `language IN (`+strings.Join(l, ", ")+`)`)
		}
		if len(op.Opt.NotLanguages) > 0 {
			var l []string
			for _, language := range op.Opt.NotLanguages {
				id, err := toDBLang(language)
				if err != nil {
					continue
				}
				l = append(l, arg(id))
			}
			wheres = append(wheres, `language NOT IN (`+strings.Join(l, ", ")+`)`)
		}

		if len(op.Opt.Kinds) > 0 {
			var kindList []string
			for _, kind := range op.Opt.Kinds {
				kindList = append(kindList, arg(kind))
			}
			wheres = append(wheres, `kind IN (`+strings.Join(kindList, ", ")+`)`)
		}
		if len(op.Opt.NotKinds) > 0 {
			var notKindList []string
			for _, kind := range op.Opt.NotKinds {
				notKindList = append(notKindList, arg(kind))
			}
			wheres = append(wheres, `kind NOT IN (`+strings.Join(notKindList, ", ")+`)`)
		}

		if op.UnitQuery != "" {
			wheres = append(wheres, `unit=`+arg(op.UnitQuery))
		}
		if op.UnitTypeQuery != "" {
			wheres = append(wheres, `lower(unit_type)=lower(`+arg(op.UnitTypeQuery)+`)`)
		}

		if len(op.TokQuery) == 1 { // special-case single token queries for performance
			wheres = append(wheres, `lower(name)=lower(`+arg(op.TokQuery[0])+`)`)

			if op.Opt.Fast && len(op.TokQuery) < 10 {
				// Require ref_ct be above a threshold for single-token queries that have short length.
				// NOTE: the RHS of this predicate ("3") must match the conditional of the partial index *exactly*.
				wheres = append(wheres, `ref_ct >= 3`)
			}

			// Skip prefix matching for too few characters.
			if op.Opt.PrefixMatch && len(op.TokQuery[0]) > 2 {
				prefixSQL = ` OR to_tsquery('english', ` + arg(op.TokQuery[0]+":*") + `) @@ to_tsvector('english', bow)`
			}
		} else if bowQuery != "" {
			wheres = append(wheres, "bow != ''")
			if op.Opt.PrefixMatch {
				wheres = append(wheres, `to_tsquery('english', `+arg(bowQuery+":*")+`) @@ to_tsvector('english', bow)`)
			} else {
				wheres = append(wheres, `to_tsquery('english', `+arg(bowQuery)+`) @@ to_tsvector('english', bow)`)
			}
		}

		whereSQL = fmt.Sprint(`WHERE (`+strings.Join(wheres, ") AND (")+`)`) + prefixSQL
	}
	orderSQL := `ORDER BY score DESC`
	limitSQL := `LIMIT ` + arg(op.Opt.PerPageOrDefault())

	sql := strings.Join([]string{selectSQL, whereSQL, orderSQL, limitSQL}, "\n")

	var dbSearchResults []*dbGlobalSearchResult
	if _, err := graphDBH(ctx).Select(&dbSearchResults, sql, args...); err != nil {
		return nil, err
	}

	// Critical permissions check. DO NOT REMOVE.
	var results []*sourcegraph.DefSearchResult
	for _, d := range dbSearchResults {
		if err := accesscontrol.VerifyUserHasReadAccess(ctx, "GlobalDefs.Search", d.Repo); err != nil {
			continue
		}
		def := fromDBDef(&d.dbGlobalDef)
		results = append(results, &sourcegraph.DefSearchResult{
			Def:      *def,
			RefCount: int32(d.RefCount),
			Score:    float32(d.Score),
		})
	}
	return &sourcegraph.SearchResultsList{DefResults: results}, nil
}

func (g *globalDefs) Update(ctx context.Context, op store.GlobalDefUpdateOp) error {
	for _, repoUnit := range op.RepoUnits {
		if err := accesscontrol.VerifyUserHasWriteAccess(ctx, "GlobalDefs.Update", repoUnit.Repo); err != nil {
			return err
		}
	}

	observe := func(part string, start time.Time) {
		log15.Debug("TRACE GlobalDefsMulti.Update", "part", part, "duration", time.Since(start))
	}
	if len(op.RepoUnits) == 1 {
		// If we just have 1 repo (the usual case), we can do some
		// more reasonable instrumentation
		repo, err := store.ReposFromContext(ctx).Get(ctx, op.RepoUnits[0].Repo)
		if err != nil {
			return err
		}
		observe = func(part string, start time.Time) {
			since := time.Since(start)
			log15.Debug("TRACE GlobalDefs.Update", "repo", repo.URI, "part", part, "duration", since)
			trackedRepo := repotrackutil.GetTrackedRepo(repo.URI)
			globalDefsUpdateDuration.WithLabelValues(trackedRepo, part).Observe(since.Seconds())
		}
	}
	defer observe("total", time.Now())

	start := time.Now()
	repoUnits, err := resolveUnits(ctx, op.RepoUnits)
	observe("resolveUnits", start)
	if err != nil {
		return err
	}

	start = time.Now()
	for _, repoUnit := range repoUnits {
		repoPath, commitID, err := resolveRevisionDefaultBranch(ctx, repoUnit.Repo)
		if err != nil {
			return err
		}

		if err := updateDefs(ctx, false, repoPath, commitID, repoUnit.UnitType, repoUnit.Unit); err != nil {
			return err
		}
	}
	observe("updateDefs", start)

	return nil
}

func (g *globalDefs) RefreshRefCounts(ctx context.Context, op store.GlobalDefUpdateOp) error {
	for _, r := range op.RepoUnits {
		if err := accesscontrol.VerifyUserHasWriteAccess(ctx, "GlobalDefs.RefreshRefCounts", r.Repo); err != nil {
			return err
		}
	}

	observe := func(part string, start time.Time) {
		log15.Debug("TRACE GlobalDefsMulti.RefreshRefCounts", "part", part, "duration", time.Since(start))
	}
	if len(op.RepoUnits) == 1 {
		// If we just have 1 repo (the usual case), we can do some
		// more reasonable instrumentation
		repo, err := store.ReposFromContext(ctx).Get(ctx, op.RepoUnits[0].Repo)
		if err != nil {
			return err
		}
		observe = func(part string, start time.Time) {
			since := time.Since(start)
			log15.Debug("TRACE GlobalDefs.RefreshRefCounts", "repo", repo.URI, "part", part, "duration", since)
			trackedRepo := repotrackutil.GetTrackedRepo(repo.URI)
			globalDefsRefreshRefCountsDuration.WithLabelValues(trackedRepo, part).Observe(since.Seconds())
		}
	}
	defer observe("total", time.Now())

	start := time.Now()
	repoUnits, err := resolveUnits(ctx, op.RepoUnits)
	observe("resolveUnits", start)
	if err != nil {
		return err
	}

	start = time.Now()
	for _, repoUnit := range repoUnits {
		updateSQL := `UPDATE global_defs d
SET ref_ct = refs.ref_ct
FROM (SELECT def_keys.repo def_repo, def_keys.unit_type def_unit_type, def_keys.unit def_unit, def_keys.path def_path, sum(global_refs_new.count) ref_ct
      FROM global_refs_new
      INNER JOIN def_keys
      ON global_refs_new.def_key_id = def_keys.id
      WHERE def_keys.repo=$1 AND def_keys.unit_type=$2 AND def_keys.unit=$3
      GROUP BY def_repo, def_unit_type, def_unit, def_path) refs
WHERE repo=def_repo AND unit_type=refs.def_unit_type AND unit=refs.def_unit AND path=refs.def_path;`
		_, err := graphDBH(ctx).Exec(updateSQL, repoUnit.RepoURI, repoUnit.UnitType, repoUnit.Unit)
		if err != nil {
			return err
		}
	}
	log15.Debug("GlobalRefs.RefreshRefCounts finished", "units", len(repoUnits), "duration", time.Since(start))
	observe("update", start)
	return nil
}

type resolvedRepoUnit struct {
	store.RepoUnit
	RepoURI string
}

// resolveUnits resolves RepoUnits without a source unit specified to
// their underlying source units
func resolveUnits(ctx context.Context, repoUnits []store.RepoUnit) ([]resolvedRepoUnit, error) {
	var resolved []resolvedRepoUnit
	for _, repoUnit := range repoUnits {
		repo, err := store.ReposFromContext(ctx).Get(ctx, repoUnit.Repo)
		if err != nil {
			return nil, err
		}

		if repoUnit.Unit != "" {
			resolved = append(resolved, resolvedRepoUnit{RepoUnit: repoUnit, RepoURI: repo.URI})
			continue
		}

		start := time.Now()
		units_, err := store.GraphFromContext(ctx).Units(sstore.ByRepos(repo.URI))
		if err != nil {
			return nil, err
		}
		for _, u := range units_ {
			resolved = append(resolved, resolvedRepoUnit{
				RepoUnit: store.RepoUnit{
					Repo:     repoUnit.Repo,
					Unit:     u.Name,
					UnitType: u.Type,
				},
				RepoURI: repo.URI,
			})
		}
		since := time.Since(start)
		log15.Debug("TRACE GlobalDefs", "repo", repo.URI, "part", "resolveUnits", "units", len(units_), "duration", since)
		trackedRepo := repotrackutil.GetTrackedRepo(repo.URI)
		globalDefsResolveUnitDuration.WithLabelValues(trackedRepo).Observe(since.Seconds())
	}
	return resolved, nil
}

func resolveRevisionDefaultBranch(ctx context.Context, repo int32) (repoPath, commitID string, err error) {
	repoObj, err := store.ReposFromContext(ctx).Get(ctx, repo)
	if err != nil {
		return
	}
	vcsrepo, err := store.RepoVCSFromContext(ctx).Open(ctx, repoObj.ID)
	if err != nil {
		return
	}
	c, err := vcsrepo.ResolveRevision(repoObj.DefaultBranch)
	if err != nil {
		return
	}
	return repoObj.URI, string(c), nil
}

func repoURIs(ctx context.Context, repoIDs []int32) (uris []string, err error) {
	for _, repoID := range repoIDs {
		repo, err := store.ReposFromContext(ctx).Get(ctx, repoID)
		if err != nil {
			return nil, err
		}
		uris = append(uris, repo.URI)
	}
	return uris, nil
}

var globalDefsResolveUnitDuration = prometheus.NewSummaryVec(prometheus.SummaryOpts{
	Namespace: "src",
	Subsystem: "global_defs",
	Name:      "resolve_unit_duration_seconds",
	Help:      "Duration for resolving a unit in global_defs",
	MaxAge:    time.Hour,
}, []string{"repo"})
var globalDefsUpdateDuration = prometheus.NewSummaryVec(prometheus.SummaryOpts{
	Namespace: "src",
	Subsystem: "global_defs",
	Name:      "update_duration_seconds",
	Help:      "Duration for updating global_defs",
	MaxAge:    time.Hour,
}, []string{"repo", "part"})
var globalDefsRefreshRefCountsDuration = prometheus.NewSummaryVec(prometheus.SummaryOpts{
	Namespace: "src",
	Subsystem: "global_defs",
	Name:      "ref_counts_duration_seconds",
	Help:      "Duration for refreshing RefCounts for global_defs",
	MaxAge:    time.Hour,
}, []string{"repo", "part"})

func init() {
	prometheus.MustRegister(globalDefsResolveUnitDuration)
	prometheus.MustRegister(globalDefsUpdateDuration)
	prometheus.MustRegister(globalDefsRefreshRefCountsDuration)
}
