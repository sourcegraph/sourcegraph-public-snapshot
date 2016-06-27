package localstore

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"

	"gopkg.in/gorp.v1"
	"gopkg.in/inconshreveable/log15.v2"

	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/dbutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/search"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/store"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/accesscontrol"
	"sourcegraph.com/sourcegraph/srclib/graph"
	sstore "sourcegraph.com/sourcegraph/srclib/store"
)

type dbRepoRev struct {
	ID     int64  `db:"id"`
	Repo   string `db:"repo"`
	Commit string `db:"commit"`

	// State is either 0, 1, or 2. 0 indicates that data for the revision is in
	// the process of being uploaded but is not yet available for querying. 1
	// indicates that the revision is the latest indexed revision of the
	// repository. 2 indicates that the revision is indexed, but not the latest
	// indexed revision. In the global_defs table, definitions whose revisions
	// have state 2 are garbage collected.
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

type dbDefInsert struct {
	dbDefShared

	ToksD string `db:"toks_c"`
	ToksC string `db:"toks_c"`
	ToksB string `db:"toks_b"`
	ToksA string `db:"toks_a"`
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

var dbDefCreateTmpSQL = `CREATE TEMPORARY TABLE defnew (
rid bigint,
defid bigint,
name TEXT,
kind TEXT,
language smallint,
updated_at timestamp with time zone,
toks_a TEXT,
toks_b TEXT,
toks_c TEXT,
toks_d TEXT)
ON COMMIT DROP;`

var dbDefDeleteSQL = `DELETE FROM defs2 WHERE rid=$1`

var dbDefInsertSQL = `INSERT INTO defs2(rid, defid, name, kind, language, bow, updated_at, state, ref_ct)
SELECT d.rid, d.defid, d.name, d.kind, d.language,
	setweight(to_tsvector('english', d.toks_a), 'A') || setweight(to_tsvector('english', d.toks_b), 'B') || setweight(to_tsvector('english', d.toks_c), 'C') || setweight(to_tsvector('english', d.toks_d), 'D'), d.updated_at, 0, 0
FROM defnew d;
`

func execDBDefInsert(tx gorp.SqlExecutor, repoRevID int64, dbDefs []*dbDefInsert) error {
	if _, err := tx.Exec(dbDefCreateTmpSQL); err != nil {
		return err
	}
	copy, err := dbutil.Prepare(tx, pq.CopyIn("defnew", "rid", "defid", "name", "kind", "language", "updated_at", "toks_a", "toks_b", "toks_c", "toks_d"))
	if err != nil {
		return fmt.Errorf("defnew copy prepare failed: %s", err)
	}
	for _, df := range dbDefs {
		if _, err := copy.Exec(df.Rev, df.DefKey, df.Name, df.Kind, df.Language, df.UpdatedAt, df.ToksA, df.ToksB, df.ToksC, df.ToksD); err != nil {
			return fmt.Errorf("defnew copy failed: %s", err)
		}
	}
	if _, err := copy.Exec(); err != nil {
		return fmt.Errorf("defnew final copy failed: %s", err)
	}

	if _, err := tx.Exec(dbDefDeleteSQL, repoRevID); err != nil {
		return fmt.Errorf("defs2 delete failed: %s", err)
	}
	if _, err := tx.Exec(dbDefInsertSQL); err != nil {
		return fmt.Errorf("defs2 insert failed: %s", err)
	}
	return nil
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
			`CREATE INDEX `+table+`_bow_idx ON `+table+` USING gin(bow);`,
			`CREATE INDEX `+table+`_bow_latest_idx ON `+table+` USING gin(bow) WHERE state=1;`,
			`CREATE INDEX `+table+`_name ON `+table+` USING btree (lower(name));`,
			`CREATE INDEX `+table+`_rid_idx ON `+table+` using btree (rid);`,
			`CREATE INDEX `+table+`_defid_idx ON `+table+` using btree (defid);`,
		)
	}
}

type defs struct{}

var _ store.Defs = (*defs)(nil)

func (s *defs) Search(ctx context.Context, op store.DefSearchOp) (*sourcegraph.SearchResultsList, error) {
	// Params checking
	if !op.Opt.Latest {
		if len(op.Opt.Repos) != 1 {
			return nil, fmt.Errorf("Repos must have exactly one element if not searching latest")
		}
		if len(op.Opt.NotRepos) > 0 {
			return nil, fmt.Errorf("NotRepos unsupported if not searching latest")
		}
		if op.Opt.CommitID == "" {
			return nil, fmt.Errorf("CommitID must be specified if not searching latest")
		}
	}

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
		scoreSQL = `5.0*log(10 + ref_ct) + 1000.0*ts_rank('{0.1, 0.1, 0.1, 1.0}', bow, to_tsquery('english', ` + arg(bowQuery) + `)) + 100.0*((LOWER(name)=LOWER(` + arg(lastTok) + `))::int) score`
	} else {
		scoreSQL = `ref_ct score`
	}
	selectSQL := `SELECT rid, defid, name, kind, language, updated_at, state, ref_ct, ` + scoreSQL + ` FROM defs2`
	var whereSQL, prefixSQL string
	{
		var wheres []string

		if op.Opt.Latest {
			wheres = append(wheres, "state=1")

			if len(op.Opt.NotRepos) > 0 {
				notRIDs := make([]int64, len(op.Opt.NotRepos))
				for i, r := range op.Opt.NotRepos {
					notRepo, err := store.ReposFromContext(ctx).Get(ctx, r)
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
			}
		}

		// Repository/commit filtering.
		if len(op.Opt.Repos) == 1 {
			rp, err := store.ReposFromContext(ctx).Get(ctx, op.Opt.Repos[0])
			if err != nil {
				return nil, fmt.Errorf("error getting included repository: %s", err)
			}

			var rid int64
			if op.Opt.CommitID == "" {
				rr, err := getRepoRevLatest(graphDBH(ctx), rp.URI)
				if err == repoRevUnindexedErr {
					return &sourcegraph.SearchResultsList{}, nil
				} else if err != nil {
					return nil, fmt.Errorf("error getting latest repo revision: %s", err)
				}
				rid = rr.ID
			} else {
				var err error
				rid, err = getRID(ctx, graphDBH(ctx), rp.URI, op.Opt.CommitID)
				if err != nil {
					return nil, fmt.Errorf("could not get repo rev ID: %s", err)
				}
			}
			wheres = append(wheres, `rid=`+arg(rid))
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

		if len(op.TokQuery) == 1 { // special-case single token queries for performance
			wheres = append(wheres, `lower(name)=lower(`+arg(op.TokQuery[0])+`)`)

			// Skip matching for too less characters.
			if op.Opt.PrefixMatch && len(op.TokQuery[0]) > 2 {
				prefixSQL = ` OR to_tsquery('english', ` + arg(op.TokQuery[0]+":*") + `) @@ bow`
			}
		} else if bowQuery != "" {
			wheres = append(wheres, "bow != ''")
			if op.Opt.PrefixMatch {
				wheres = append(wheres, `to_tsquery('english', `+arg(bowQuery+":*")+`) @@ bow`)
			} else {
				wheres = append(wheres, `to_tsquery('english', `+arg(bowQuery)+`) @@ bow`)
			}
		}

		whereSQL = fmt.Sprint(`WHERE (`+strings.Join(wheres, ") AND (")+`)`) + prefixSQL
	}
	orderSQL := `ORDER BY score DESC`
	limitSQL := `LIMIT ` + arg(op.Opt.PerPageOrDefault())

	sql := strings.Join([]string{selectSQL, whereSQL, orderSQL, limitSQL}, "\n")

	var dbSearchResults []*dbDefSearchResult
	if _, err := graphDBH(ctx).Select(&dbSearchResults, sql, args...); err != nil {
		return nil, fmt.Errorf("error fetching from defs2: %s", err)
	}

	var results []*sourcegraph.DefSearchResult
	for _, d := range dbSearchResults {
		// convert dbDef to Def
		var def sourcegraph.Def
		{
			dk, err := getDefKey(ctx, graphDBH(ctx), d.DefKey)
			if err != nil {
				return nil, fmt.Errorf("error getting def key: %s", err)
			}

			rv, err := getRepoRev(ctx, graphDBH(ctx), d.Rev)
			if err != nil {
				return nil, fmt.Errorf("error getting repo revision repo_rev.id %d: %s", d.Rev, err)
			}

			def = sourcegraph.Def{
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
		}

		// Critical permissions check. DO NOT REMOVE.
		if err := accesscontrol.VerifyUserHasReadAccess(ctx, "GlobalDefs.Search", def.Repo); err != nil {
			continue
		}

		results = append(results, &sourcegraph.DefSearchResult{
			Def:      def,
			RefCount: int32(d.RefCount),
			Score:    float32(d.Score),
		})
	}

	return &sourcegraph.SearchResultsList{DefResults: results}, nil
}

// UpdateFromSrclibStore is a stop-gap method. Eventually, defs will replace
// srclib store as the canonical storage for defs. Until then, the canonical
// storage is srclib store. UpdateFromSrclibStore will sync the data in defs to
// reflect what is in srclib store for a given (repo, commit).
func (s *defs) UpdateFromSrclibStore(ctx context.Context, op store.DefUpdateOp) error {
	if err := accesscontrol.VerifyUserHasWriteAccess(ctx, "Defs.UpdateFromSrclibStore", op.Repo); err != nil {
		return err
	}

	repo, err := store.ReposFromContext(ctx).Get(ctx, op.Repo)
	if err != nil {
		return err
	}

	if len(op.CommitID) == 0 {
		rr, err := getRepoRevLatest(graphDBH(ctx), repo.URI)
		if err != nil {
			return err
		}
		op.CommitID = rr.Commit
	} else if len(op.CommitID) != 40 {
		return fmt.Errorf("commit ID must be 40 characters long, was: %q", op.CommitID)
	}

	defs_, err := store.GraphFromContext(ctx).Defs(
		sstore.ByRepoCommitIDs(sstore.Version{Repo: repo.URI, CommitID: op.CommitID}),
	)
	if err != nil {
		return err
	}

	op.Defs = defs_
	return s.Update(ctx, op)
}

func (s *defs) Update(ctx context.Context, op store.DefUpdateOp) error {
	if err := accesscontrol.VerifyUserHasWriteAccess(ctx, "Defs.Update", op.Repo); err != nil {
		return err
	}

	repo, err := store.ReposFromContext(ctx).Get(ctx, op.Repo)
	if err != nil {
		return err
	}

	// Validate input
	if op.Repo == 0 || op.CommitID == "" {
		return fmt.Errorf("both op.Repo and op.CommitID must be non-empty")
	}
	if len(op.CommitID) != 40 {
		return fmt.Errorf("commit must be 40 characters long, was: %q", op.CommitID)
	}

	for _, def := range op.Defs {
		if def.Repo != "" && def.Repo != repo.URI {
			return fmt.Errorf("cannot update def with non-matching repo (%s != %s)", def.Repo, repo.URI)
		}
		if def.CommitID != "" && def.CommitID != op.CommitID {
			return fmt.Errorf("cannot update def with non-matching revision (%s != %s)", def.CommitID, op.CommitID)
		}
	}

	// KLUDGE to improve search quality. This info ideally would be emitted by srclib toolchains.
	var chosenDefs []*graph.Def
	for _, d := range op.Defs {
		if shouldIndex(d) {
			chosenDefs = append(chosenDefs, d)
		}
	}

	dbh := graphDBH(ctx)

	// Update def_keys
	if err := updateDefKeys(ctx, dbh, repo.URI, chosenDefs); err != nil {
		return err
	}

	defKeys := make(map[graph.DefKey]struct{})
	for _, def := range chosenDefs {
		var dk graph.DefKey = def.DefKey
		dk.Repo, dk.CommitID = repo.URI, ""
		defKeys[dk] = struct{}{}
	}
	dbDefKeys, err := getDefKeys(ctx, dbh, defKeys)
	if err != nil {
		return err
	}
	defKeyIDs := make(map[graph.DefKey]int64)
	for _, dk := range dbDefKeys {
		defKeyIDs[graph.DefKey{Repo: dk.Repo, UnitType: dk.UnitType, Unit: dk.Unit, Path: dk.Path}] = dk.ID
	}

	// Update repo_revs
	repoRevsInsertSQL := `INSERT INTO repo_revs(repo, commit, state) (SELECT $1 AS repo, $2 AS commit, 0 AS state WHERE NOT EXISTS (SELECT 1 FROM repo_revs WHERE repo=$1 AND commit=$2))`
	if _, err := dbh.Exec(repoRevsInsertSQL, repo.URI, op.CommitID); err != nil {
		return fmt.Errorf("repo_rev update failed: %s", err)
	}
	repoRevID, err := dbh.SelectInt(`SELECT id FROM repo_revs WHERE repo=$1 AND commit=$2`, repo.URI, op.CommitID)
	if err != nil {
		return fmt.Errorf("repo_rev id fetch failed: %s", err)
	}

	// Compute defs to insert
	langWarnCount := 0
	dbDefs := make([]*dbDefInsert, len(chosenDefs))
	now := time.Now()
	for i, def := range chosenDefs {
		dk := def.DefKey
		dk.Repo, dk.CommitID = repo.URI, ""

		// TODO(beyang): kludge. Should not rely on def formatter for this information.
		languageID, err := toDBLang(strings.ToLower(graph.PrintFormatter(def).Language()))
		if err != nil {
			langWarnCount++
		}

		tsvector := search.ToTSVector(def)

		dbDefs[i] = &dbDefInsert{
			dbDefShared: dbDefShared{
				Rev:       repoRevID,
				DefKey:    defKeyIDs[dk],
				Name:      def.Name,
				Kind:      def.Kind,
				Language:  languageID,
				UpdatedAt: &now,
			},
			ToksA: strings.Join(search.BagOfWordsToTokens2(tsvector.AToks, tsvector.A), " "),
			ToksB: strings.Join(search.BagOfWordsToTokens2(tsvector.BToks, tsvector.B), " "),
			ToksC: strings.Join(search.BagOfWordsToTokens2(tsvector.CToks, tsvector.C), " "),
			ToksD: strings.Join(search.BagOfWordsToTokens2(tsvector.DToks, tsvector.D), " "),
		}
	}
	if langWarnCount > 0 {
		log15.Warn("could not determine language for all defs", "noLang", langWarnCount, "allDefs", len(chosenDefs))
	}

	if err := dbutil.Transact(dbh, func(tx gorp.SqlExecutor) error {
		return execDBDefInsert(tx, repoRevID, dbDefs)
	}); err != nil {
		return err
	}

	// Update state column
	if err := dbutil.Transact(dbh, func(tx gorp.SqlExecutor) error {
		if op.Latest {
			var repoRevs []*dbRepoRev
			if _, err := tx.Select(&repoRevs, `SELECT * FROM repo_revs WHERE repo=$1 AND state=1`, repo.URI); err != nil {
				return err
			}
			oldLatestRIDs := make([]int64, len(repoRevs))
			for i, repoRev := range repoRevs {
				oldLatestRIDs[i] = repoRev.ID
			}

			if _, err := tx.Exec(`UPDATE repo_revs SET state=2 WHERE repo=$1 AND state=1`, repo.URI); err != nil {
				return err
			}
			if _, err := tx.Exec(`UPDATE repo_revs SET state=1 WHERE id=$1`, repoRevID); err != nil {
				return err
			}
			if len(oldLatestRIDs) > 0 {
				var params = make([]string, len(oldLatestRIDs))
				var args []interface{}
				arg := func(v interface{}) string {
					args = append(args, v)
					return gorp.PostgresDialect{}.BindVar(len(args) - 1)
				}
				for i, rid := range oldLatestRIDs {
					params[i] = arg(rid)
				}
				s := `UPDATE defs2 SET state=2 WHERE rid IN (` + strings.Join(params, ",") + `)`
				if _, err := tx.Exec(s, args...); err != nil {
					return err
				}
			}
			if _, err := tx.Exec(`UPDATE defs2 SET state=1 WHERE rid=$1`, repoRevID); err != nil {
				return err
			}
		} else {
			if _, err := tx.Exec("UPDATE repo_revs SET state=2 WHERE id=$1", repoRevID); err != nil {
				return err
			}
			if _, err := tx.Exec("UPDATE defs2 SET state=2 WHERE rid=$1", repoRevID); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return err
	}

	if op.RefreshCounts {
		if err := s.UpdateRefCounts(ctx, repo.URI); err != nil {
			return err
		}
	}
	return nil
}

var updateRefCountSQL = `UPDATE defs2 SET ref_ct = ref_counts.ref_ct
FROM (
	SELECT d.defid defid, coalesce(sum(gr.count), 0) ref_ct
	FROM defs2 d LEFT JOIN global_refs_new gr ON d.defid=gr.def_key_id
	WHERE d.rid=$1 AND gr.repo != $2
	GROUP BY defid
	ORDER BY defid
) ref_counts WHERE defs2.defid=ref_counts.defid;
`

// UpdateRefCounts updates the ref_ct column to reflect the number of xrefs
// referencing defs in the latest built revision of the specified repository
// (repo).
func (s *defs) UpdateRefCounts(ctx context.Context, repo string) error {
	rr, err := getRepoRevLatest(graphDBH(ctx), repo)
	if err != nil {
		return err
	}

	if _, err := graphDBH(ctx).Exec(updateRefCountSQL, rr.ID, rr.Repo); err != nil {
		return err
	}
	return nil
}

func getDefKey(ctx context.Context, dbh gorp.SqlExecutor, id int64) (*dbDefKey, error) {
	var d dbDefKey
	if err := dbh.SelectOne(&d, `SELECT * FROM def_keys WHERE id=$1`, id); err != nil {
		return nil, err
	}
	return &d, nil
}

func getRID(ctx context.Context, dbh gorp.SqlExecutor, repo string, commit string) (int64, error) {
	return dbh.SelectInt(`SELECT id FROM repo_revs WHERE repo=$1 AND commit=$2`, repo, commit)
}

func getRepoRev(ctx context.Context, dbh gorp.SqlExecutor, id int64) (*dbRepoRev, error) {
	var d dbRepoRev
	if err := dbh.SelectOne(&d, `SELECT * FROM repo_revs WHERE id=$1`, id); err != nil {
		return nil, err
	}
	return &d, nil
}

func getDefKeys(ctx context.Context, dbh gorp.SqlExecutor, defKeys map[graph.DefKey]struct{}) ([]*dbDefKey, error) {
	var dbDefKeys []*dbDefKey
	createTmpSQL := `CREATE TEMPORARY TABLE def_keys_tmp (
repo TEXT,
unit_type TEXT,
unit TEXT,
path TEXT)
ON COMMIT DROP;`
	selectSQL := `SELECT * FROM def_keys WHERE (repo, unit_type, unit, path) IN (SELECT * from def_keys_tmp)`
	err := dbutil.Transact(dbh, func(tx gorp.SqlExecutor) error {
		if _, err := tx.Exec(createTmpSQL); err != nil {
			return err
		}
		copy, err := dbutil.Prepare(tx, pq.CopyIn("def_keys_tmp", "repo", "unit_type", "unit", "path"))
		if err != nil {
			return fmt.Errorf("def_keys_tmp copy prepare failed: %s", err)
		}
		for defKey, _ := range defKeys {
			if _, err := copy.Exec(defKey.Repo, defKey.UnitType, defKey.Unit, defKey.Path); err != nil {
				return fmt.Errorf("def_keys_tmp copy failed: %s", err)
			}
		}
		if _, err := copy.Exec(); err != nil {
			return fmt.Errorf("def_keys_tmp final copy failed: %s", err)
		}
		if _, err := tx.Select(&dbDefKeys, selectSQL); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return dbDefKeys, nil
}

// updateDefKeys, given a list of defs, inserts all missing defs into def_keys.
// It never deletes anything.
func updateDefKeys(ctx context.Context, dbh gorp.SqlExecutor, repo string, defs []*graph.Def) error {
	createTmpSQL := `CREATE TEMPORARY TABLE def_keys_tmp (
repo TEXT,
unit_type TEXT,
unit TEXT,
path TEXT)
ON COMMIT DROP;`
	insertSQL := `INSERT INTO def_keys(repo, unit_type, unit, path)
SELECT repo, unit_type, unit, path
FROM def_keys_tmp dkt
WHERE NOT EXISTS (
	SELECT id FROM def_keys dk where dk.repo=dkt.repo AND dk.unit_type=dkt.unit_type AND dk.unit=dkt.unit AND dk.path=dkt.path
)`

	return dbutil.Transact(dbh, func(tx gorp.SqlExecutor) error {
		if _, err := tx.Exec(createTmpSQL); err != nil {
			return fmt.Errorf("def_keys_tmp create failed: %s", err)
		}
		copy, err := dbutil.Prepare(tx, pq.CopyIn("def_keys_tmp", "repo", "unit_type", "unit", "path"))
		if err != nil {
			return fmt.Errorf("def_keys_tmp prepare failed: %s", err)
		}
		for _, def := range defs {
			rp := def.Repo
			if rp == "" {
				rp = repo
			}
			if _, err := copy.Exec(rp, def.UnitType, def.Unit, def.Path); err != nil {
				return fmt.Errorf("def_keys_tmp copy failed: %s", err)
			}
		}
		if _, err := copy.Exec(); err != nil {
			return fmt.Errorf("def_keys_tmp final copy failed: %s", err)
		}
		if _, err := tx.Exec(insertSQL); err != nil {
			return fmt.Errorf("def_keys insert failed: %s", err)
		}
		return nil
	})
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
