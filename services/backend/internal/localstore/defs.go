package localstore

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"gopkg.in/gorp.v1"
	"gopkg.in/inconshreveable/log15.v2"

	"context"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/accesscontrol"
	"sourcegraph.com/sourcegraph/srclib/graph"
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
	if l, exists := toDBLang_[strings.ToLower(lang)]; exists {
		return l, nil
	}
	return 0, fmt.Errorf("unrecognized language %s", lang)
}

type dbRepoRev struct {
	ID     int64  `db:"id"`
	Repo   string `db:"repo"`
	Commit string `db:"commit"`

	// State is either 0, 1, or 2. 0 indicates that data for the revision is in
	// the process of being uploaded but is not yet available for querying. 1
	// indicates that the revision is the latest indexed revision of the
	// repository. 2 indicates that the revision is indexed, but not the latest
	// indexed revision. In the defs table, definitions whose revisions have state
	// 2 are garbage collected. State increases monotically over time (i.e.,
	// something in state 2 never transitions to state 1)
	State uint8 `db:"state"`
}

type dbDef struct {
	dbDefShared

	// State mirrors dbRepoRev.State. This value should always be kept in sync
	// with the value of dbRepoRev.State.
	State uint8 `db:"state"`

	// RefCount is the computed number of references (internal + external) that refer to this definition.
	RefCount int `db:"ref_ct"`
}

type dbDefShared struct {
	// Rev is the foreign key to the repo_rev table
	Rev int64 `db:"rid"`

	// DefKey is the foreign key to the def_keys table
	DefKey int64 `db:"defid"`

	// These fields mirror the fields in sourcegraph.Def
	Name string `db:"name"`
	Kind string `db:"kind"`

	// Language is the deteced language of the definition. This is used for language-filtered queries.
	Language dbLang `db:"language"`

	// UpdatedAt is the last time at which this row as updated in the DB.
	UpdatedAt *time.Time `db:"updated_at"`
}

type dbDefSearchResult struct {
	dbDef
	Score float64 `db:"score"`
}

func init() {
	GraphSchema.Map.AddTableWithName(dbRepoRev{}, "repo_revs").SetKeys(true, "id").SetUniqueTogether("repo", "commit")
	GraphSchema.CreateSQL = append(GraphSchema.CreateSQL,
		`ALTER TABLE repo_revs ALTER COLUMN state TYPE smallint`,
	)
	GraphSchema.Map.AddTableWithName(dbDef{}, "defs2").SetKeys(false, "rid", "defid")
	for _, table := range []string{"defs2"} {
		GraphSchema.CreateSQL = append(GraphSchema.CreateSQL,
			`ALTER TABLE `+table+` ADD COLUMN bow tsvector;`,
			`ALTER TABLE `+table+` ALTER COLUMN updated_at TYPE timestamp with time zone USING updated_at::timestamp with time zone;`,
			`ALTER TABLE `+table+` ALTER COLUMN ref_ct SET DEFAULT 0;`,
			`ALTER TABLE `+table+` ALTER COLUMN ref_ct SET NOT NULL;`,
			`ALTER TABLE `+table+` ALTER COLUMN language TYPE smallint`,
			`ALTER TABLE `+table+` ALTER COLUMN state TYPE smallint`,
			`CREATE INDEX `+table+`_bow_latest_idx ON `+table+` USING gin(bow) WHERE state=1;`,
			`CREATE INDEX `+table+`_bow_fast_idx ON `+table+` USING gin(bow) WHERE ref_ct > 10;`,
			`CREATE INDEX `+table+`_name ON `+table+` USING btree (lower(name));`,
			`CREATE INDEX `+table+`_name_fast ON `+table+` USING btree (lower(name)) WHERE ref_ct > 10;`,
			`CREATE INDEX `+table+`_rid_idx ON `+table+` using btree (rid);`,
			`CREATE INDEX `+table+`_defid_idx ON `+table+` using btree (defid);`,
		)
	}
}

type defs struct{}

type DefSearchOp struct {
	// TokQuery is a list of tokens that describe the user's text
	// query. Order matter, as the last token is given especial weight.
	TokQuery []string
	Opt      *sourcegraph.SearchOptions
}

func (s *defs) Search(ctx context.Context, op DefSearchOp) (*sourcegraph.SearchResultsList, error) {
	if Mocks.Defs.Search != nil {
		return Mocks.Defs.Search(ctx, op)
	}

	startTime := time.Now()
	var args []interface{}
	arg := func(v interface{}) string {
		args = append(args, v)
		return gorp.PostgresDialect{}.BindVar(len(args) - 1)
	}

	if op.Opt == nil {
		op.Opt = &sourcegraph.SearchOptions{}
	}

	if len(op.TokQuery) == 0 && len(op.Opt.Repos) == 0 && !op.Opt.AllowEmpty {
		return &sourcegraph.SearchResultsList{}, nil
	}

	obs := newDefsSearchObserver("defs2", "")
	totalEnd := obs.start("search_total")
	defer totalEnd()

	bowQuery := strings.Join(op.TokQuery, " & ")
	lastTok := ""
	if len(op.TokQuery) > 0 {
		lastTok = op.TokQuery[len(op.TokQuery)-1]
	}

	var scoreSQL string
	if bowQuery != "" {
		// The ranking critieron is the weighted sum of xref count,
		// text similarity score, and whether the last term matches
		// the name.
		scoreSQL = `5.0*log(10 + ref_ct) + 1000.0*ts_rank('{0.1, 0.1, 0.1, 1.0}', bow, to_tsquery('english', ` + arg(bowQuery) + `)) + 100.0*((LOWER(name)=LOWER(` + arg(lastTok) + `))::int) score`
	} else {
		scoreSQL = `ref_ct score`
	}
	selectSQL := `SELECT rid, defid, name, kind, language, updated_at, state, ref_ct, ` + scoreSQL + ` FROM defs2`
	var whereSQL, fastWhereSQL, prefixSQL string
	{
		var wheres []string
		wheres = append(wheres, "state=1")

		if len(op.Opt.NotRepos) > 0 {
			end := obs.start("notrepos")
			notRIDs := make([]int64, len(op.Opt.NotRepos))
			for i, r := range op.Opt.NotRepos {
				notRepo, err := Repos.Get(ctx, r)
				if err != nil {
					return nil, fmt.Errorf("error getting excluded repository: %s", err)
				}
				// NOTE(beyang): there's a race condition here as the latest repository
				// revision could change between here and when the query to the defs
				// table is made. In this case, we will fail to exclude results from
				// repositories in NotRepos. Updates are infrequent enough that we
				// accept this possibility.
				rr, err := getRepoRevLatest(graphDBH(ctx), notRepo.URI)
				if err == repoRevUnindexedErr {
					return &sourcegraph.SearchResultsList{}, nil
				} else if err != nil {
					return nil, err
				}
				notRIDs[i] = rr.ID
			}
			nrArgs := make([]string, len(notRIDs))
			for i, r := range notRIDs {
				nrArgs[i] = arg(r)
			}
			wheres = append(wheres, "rid NOT IN ("+strings.Join(nrArgs, ",")+")")
			end()
		}

		// Repository/commit filtering.
		if len(op.Opt.Repos) > 0 {
			end := obs.start("repos")
			var repoArgs []string
			for _, repo := range op.Opt.Repos {
				rp, err := Repos.Get(ctx, repo)
				if err != nil {
					return nil, fmt.Errorf("error getting included repository: %s", err)
				}

				rr, err := getRepoRevLatest(graphDBH(ctx), rp.URI)
				if err == repoRevUnindexedErr {
					continue
				} else if err != nil {
					return nil, fmt.Errorf("error getting latest repo revision: %s", err)
				}
				repoArgs = append(repoArgs, arg(rr.ID))
			}
			if len(repoArgs) == 0 {
				log15.Warn("All repos specified in def search are unindexed; no results may be returned.", "repos", op.Opt.Repos)
				return &sourcegraph.SearchResultsList{}, nil
			}
			if len(repoArgs) > 0 {
				wheres = append(wheres, `rid IN (`+strings.Join(repoArgs, ",")+`)`)
			}
			end()
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

		if bowQuery != "" {
			wheres = append(wheres, "bow != ''")
			wheres = append(wheres, `to_tsquery('english', `+arg(bowQuery)+`) @@ bow`)
		}

		whereSQL = fmt.Sprint(`WHERE (`+strings.Join(wheres, ") AND (")+`)`) + prefixSQL
		fastWheres := append(wheres, "ref_ct > 10") // this corresponds to a partial index
		fastWhereSQL = fmt.Sprint(`WHERE (`+strings.Join(fastWheres, ") AND (")+`)`) + prefixSQL
	}
	orderSQL := `ORDER BY score DESC`
	limitSQL := `LIMIT ` + arg(op.Opt.PerPageOrDefault())

	sql := strings.Join([]string{selectSQL, whereSQL, orderSQL, limitSQL}, "\n")
	fastSQL := strings.Join([]string{selectSQL, fastWhereSQL, orderSQL, limitSQL}, "\n")

	var dbSearchResults []*dbDefSearchResult
	end := obs.start("select_fast")
	if _, err := graphDBH(ctx).Select(&dbSearchResults, fastSQL, args...); err != nil {
		end()
		return nil, fmt.Errorf("error fast-fetching from defs2: %s", err)
	}
	end()
	if len(dbSearchResults) == 0 { // if no fast results, search for slow results
		end := obs.start("select_slow")
		if _, err := graphDBH(ctx).Select(&dbSearchResults, sql, args...); err != nil {
			end()
			return nil, fmt.Errorf("error fetching from defs2: %s", err)
		}
		end()
	}

	end = obs.start("resolveresults")
	defer end()
	var results []*sourcegraph.DefSearchResult
	for _, d := range dbSearchResults {
		// convert dbDef to Def
		dk, err := getDefKey(ctx, graphDBH(ctx), d.DefKey)
		if err != nil {
			return nil, fmt.Errorf("error getting def key: %s", err)
		}
		rv, err := getRepoRev(ctx, graphDBH(ctx), d.Rev)
		if err != nil {
			return nil, fmt.Errorf("error getting repo revision repo_rev.id %d: %s", d.Rev, err)
		}
		def := sourcegraph.Def{
			Def: graph.Def{
				DefKey: graph.DefKey{
					Repo:     dk.Repo,
					CommitID: rv.Commit,
					UnitType: dk.UnitType,
					Unit:     dk.Unit,
					Path:     dk.Path,
				},
				Name: d.Name,
				Kind: d.Kind,
			},
		}

		// Critical permissions check. DO NOT REMOVE.
		if err := accesscontrol.VerifyUserHasReadAccess(ctx, "Defs.Search", def.Repo); err != nil {
			continue
		}

		results = append(results, &sourcegraph.DefSearchResult{
			Def:      def,
			RefCount: int32(d.RefCount),
			Score:    float32(d.Score),
		})
	}

	defsSearchResultsLength.Observe(float64(len(results)))
	if len(results) == 0 {
		defsSearchResultsNone.Inc()
	}
	log15.Debug("TRACE defs.Search", "tokens", strings.Join(op.TokQuery, ","), "opts", fmt.Sprintf("%+v", op.Opt), "results_len", len(results), "duration", time.Since(startTime))

	return &sourcegraph.SearchResultsList{DefResults: results}, nil
}

func getDefKey(ctx context.Context, dbh gorp.SqlExecutor, id int64) (*dbDefKey, error) {
	var d dbDefKey
	if err := dbh.SelectOne(&d, `SELECT * FROM def_keys WHERE id=$1`, id); err != nil {
		return nil, err
	}
	return &d, nil
}

func getRepoRev(ctx context.Context, dbh gorp.SqlExecutor, id int64) (*dbRepoRev, error) {
	var d dbRepoRev
	if err := dbh.SelectOne(&d, `SELECT * FROM repo_revs WHERE id=$1`, id); err != nil {
		return nil, err
	}
	return &d, nil
}

var repoRevUnindexedErr = errors.New("repository has not been indexed yet, so no latest revision exists")

func getRepoRevLatest(dbh gorp.SqlExecutor, repo string) (*dbRepoRev, error) {
	var rr []*dbRepoRev
	if _, err := dbh.Select(&rr, `select * from repo_revs where repo=$1 and state=1`, repo); err != nil {
		return nil, err
	}
	if len(rr) == 0 {
		return nil, repoRevUnindexedErr
	} else if len(rr) > 1 {
		return nil, fmt.Errorf("repo %s has more than one latest version", repo)
	}
	return rr[0], nil
}

var defsSearchResultsLength = prometheus.NewSummary(prometheus.SummaryOpts{
	Namespace: "src",
	Subsystem: "defs",
	Name:      "search_results_length",
	Help:      "Number of results returned for a search",
	MaxAge:    time.Hour,
})
var defsSearchResultsNone = prometheus.NewCounter(prometheus.CounterOpts{
	Namespace: "src",
	Subsystem: "defs",
	Name:      "search_results_none_total",
	Help:      "Number of times we returned no results",
})

func init() {
	prometheus.MustRegister(defsSearchResultsLength)
	prometheus.MustRegister(defsSearchResultsNone)
}

type MockDefs struct {
	Search func(ctx context.Context, op DefSearchOp) (*sourcegraph.SearchResultsList, error)
}
