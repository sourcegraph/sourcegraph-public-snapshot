package localstore

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"strings"

	log15 "gopkg.in/inconshreveable/log15.v2"

	"github.com/lib/pq"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/pkg/errors"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/go-langserver/pkg/lspext"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/dbutil"
	"sourcegraph.com/sourcegraph/sourcegraph/xlang"
)

// The goal of this global_ref DB is to take the top level [URI, Name, ContainerName]
// which effectively describes a symbol being referenced, and map it back into one or
// more respective Location which describes where the reference was made (thus
// implementing 'Find External References'):
//
// 	{
// 		Location:{
// 			URI: "git://github.com/leak-test/mux?757bef944d0f21880861c2dd9c871ca543023cba#regexp.go",
// 			Range: {
// 				Start:{Line:313, Character:41},
// 				End:{Line:313 Character:47}
// 			},
// 		}
// 		Name:"Request"
// 		ContainerName:""
// 		URI:"git://github.com/golang/go?go1.7.1#src/net/http"
// 	}
//
// Name and ContainerName fields are deduplicated using a separate
// `global_ref_name` and `global_ref_container` tables, respectively. i.e.
// although we store the same Name/ContainerName many times, we do so only
// using an ID and one single textual representation.
//
// URI fields are not stored as string literals. Instead, they are parsed and
// then stored in a DB-friendly way.
//
// 	URI.Scheme -> uriSchemeID mapping is used to store only one smallint / int16
// 	URI.Host+u.Path -> Deduplicated via a separate `global_ref_source` table.
// 	URI.Query -> Deduplicated via a separate `global_ref_version` table.
//  URI.Fragment -> Deduplicated via a separate `global_ref_file` table.
//
// This allows language servers to use many URI schemes for input and output
// such as:
//
//  (Go) git://github.com/golang/go?go1.7.1#src/net/http
//  (Rust) cargo://winapi?0.2.8#src/d3d9.rs
//  (JS) npm://react?16.1.0
//  ...
//
// It is important to note that the URI scheme does not itself identify the
// language. For example, a Rust or JS library may also have git:// URIs. For
// this reason (and others), a language field constant is stored among the
// results.

// Query Examples
//
// You may be intimidated at first by the fact that global_ref_by_file and
// global_ref_by_source tables are mostly IDs, all you need to know how to do in
// order to write effective queries is use JOIN. It lets you relate an ID from
// one of these tables to a (string) value from another. For example:
//
// 	-- This query will return to you rows like `(27,149)`, use `\d+ global_ref_by_file`
// 	-- in a psql prompt to discover what table an ID like this `REFERENCES`.
// 	SELECT (def_source, def_file) FROM global_ref_by_file;
//
// 	-- To get the textual representation for the above, use JOIN in order to
// 	-- receive a result. Here I have used the full `<table>.<field>` syntax to
// 	-- make it more clear which table is being referenced. This returns rows
// 	-- like `(github.com/golang/go,src/testing)`:
// 	SELECT (global_ref_source.source, global_ref_file.file)
// 	FROM global_ref_by_file
// 	JOIN global_ref_source
// 	ON (global_ref_source.id = global_ref_by_file.def_source)
// 	JOIN global_ref_file
// 	ON (global_ref_file.id = global_ref_by_file.def_file);
//
// 	-- If you need to select the textual representation for both e.g. source
// 	-- and def_source fields, you can use `JOIN global_ref_source myalias`, for
// 	-- example to get `(github.com/golang/go,github.com/leak-test/mux)`:
// 	SELECT (def_source.source, source.source)
// 	FROM global_ref_by_file
// 	JOIN global_ref_source def_source
// 	ON (def_source.id = global_ref_by_file.def_source)
// 	JOIN global_ref_source source
// 	ON (source.id = global_ref_by_file.source);
//
// To delete all global refs data for a specific language, find it's ID in defs.go:
//
// 	DELETE FROM global_ref_by_source WHERE language=<ID>;
// 	DELETE FROM global_ref_by_file WHERE language=<ID>;
//
// Data inside the other tables for the language does not need to be deleted
// manually, it will be automatically garbage collected at the next repository
// index.

// TODO: Long term, support more URI schemes (svn, hg, npm, mvn, etc). For now,
// you can rest assured these are all that can be in the DB.
var uriSchemeID = map[string]int16{
	"git": 1, // 0 reserved for detecting zero-value errors.
}

var uriSchemeIDLookup = make(map[int16]string, len(uriSchemeID))

func init() {
	for scheme, id := range uriSchemeID {
		uriSchemeIDLookup[id] = scheme
	}
}

// dbGlobalRefSource represents the 'global_ref_source' table.
type dbGlobalRefSource struct{}

func (*dbGlobalRefSource) CreateTable() string {
	return `CREATE table global_ref_source (
		id serial primary key NOT NULL,
		source text NOT NULL,
		UNIQUE(source)
	);
	CREATE INDEX global_ref_source_source ON global_ref_source USING btree (source);`
}

func (*dbGlobalRefSource) DropTable() string {
	return `DROP TABLE IF EXISTS global_ref_source CASCADE;`
}

// dbGlobalRefVersion represents the 'global_ref_version' table.
type dbGlobalRefVersion struct{}

func (*dbGlobalRefVersion) CreateTable() string {
	return `CREATE table global_ref_version (
		id serial primary key NOT NULL,
		version text NOT NULL,
		UNIQUE(version)
	);
	CREATE INDEX global_ref_version_version ON global_ref_version USING btree (version);`
}

func (*dbGlobalRefVersion) DropTable() string {
	return `DROP TABLE IF EXISTS global_ref_version CASCADE;`
}

// dbGlobalRefFile represents the 'global_ref_file' table.
type dbGlobalRefFile struct{}

func (*dbGlobalRefFile) CreateTable() string {
	return `CREATE table global_ref_file (
		id serial primary key NOT NULL,
		file text NOT NULL,
		UNIQUE(file)
	);
	CREATE INDEX global_ref_file_file ON global_ref_file USING btree (file);`
}

func (*dbGlobalRefFile) DropTable() string {
	return `DROP TABLE IF EXISTS global_ref_file CASCADE;`
}

// dbGlobalRefName represents the 'global_ref_name' table.
type dbGlobalRefName struct{}

func (*dbGlobalRefName) CreateTable() string {
	return `CREATE table global_ref_name (
		id serial primary key NOT NULL,
		name text NOT NULL,
		UNIQUE(name)
	);
	CREATE INDEX global_ref_name_name ON global_ref_name USING btree (name);`
}

func (*dbGlobalRefName) DropTable() string {
	return `DROP TABLE IF EXISTS global_ref_name CASCADE;`
}

// dbGlobalRefContainer represents the 'global_ref_container' table.
type dbGlobalRefContainer struct{}

func (*dbGlobalRefContainer) CreateTable() string {
	return `CREATE table global_ref_container (
		id serial primary key NOT NULL,
		container text NOT NULL,
		UNIQUE(container)
	);
	CREATE INDEX global_ref_container_container ON global_ref_container USING btree (container);`
}

func (*dbGlobalRefContainer) DropTable() string {
	return `DROP TABLE IF EXISTS global_ref_container CASCADE;`
}

// dbGlobalRefBySource represents the 'global_ref_by_source' table. This table
// tells us which sources (e.g. repos) reference a definition.
type dbGlobalRefBySource struct{}

func (*dbGlobalRefBySource) CreateTable() string {
	return `CREATE table global_ref_by_source (
		-- language represents the language that this reference was created by.
		-- See defs.go for a complete list of IDs.
		language smallint NOT NULL,

		-- def_name represents the name of the definition being referenced by
		-- this source.
		def_name integer references global_ref_name(id) NOT NULL,

		-- def_container represents the container name of the definition being
		-- referenced by this source.
		def_container integer references global_ref_container(id) NOT NULL,

		-- def_scheme, def_source, def_version, and def_file all speak about
		-- the URI where the definition itself can be located.
		def_scheme smallint NOT NULL,
		def_source integer references global_ref_source(id) NOT NULL,
		def_version integer references global_ref_version(id) NOT NULL,
		def_file integer references global_ref_file(id) NOT NULL,

		-- scheme, source and version speak about the source URI that is making
		-- a reference to the def.
		scheme smallint NOT NULL,
		source integer references global_ref_source(id) NOT NULL,
		version integer references global_ref_version(id) NOT NULL,

		-- files is how many files in the source are referencing the definition.
		files smallint NOT NULL,

		-- refs is how many total references to def there are in the source.
		refs smallint NOT NULL,

		-- score is the score for this reference. It is arbitrarily set and
		-- used for sorting. For example, upvote/downvote buttons may
		-- increment and decrement this field, or we may assign specific
		-- sources higher values (like the Go stdlib).
		score smallint NOT NULL,
		UNIQUE(def_name, def_container, def_scheme, def_source, def_version, def_file, scheme, source, version)
	);
	CREATE INDEX global_ref_by_source_language ON global_ref_by_source USING btree (language);
	CREATE INDEX global_ref_by_source_def_name ON global_ref_by_source USING btree (def_name);
	CREATE INDEX global_ref_by_source_def_container ON global_ref_by_source USING btree (def_container);
	CREATE INDEX global_ref_by_source_def_scheme ON global_ref_by_source USING btree (def_scheme);
	CREATE INDEX global_ref_by_source_def_source ON global_ref_by_source USING btree (def_source);
	CREATE INDEX global_ref_by_source_def_version ON global_ref_by_source USING btree (def_version);
	CREATE INDEX global_ref_by_source_scheme ON global_ref_by_source USING btree (scheme);
	CREATE INDEX global_ref_by_source_source ON global_ref_by_source USING btree (source);
	CREATE INDEX global_ref_by_source_version ON global_ref_by_source USING btree (version);
	CREATE INDEX global_ref_by_source_refs ON global_ref_by_source USING btree (refs);
	CREATE INDEX global_ref_by_source_score ON global_ref_by_source USING btree (score);`
}

func (*dbGlobalRefBySource) DropTable() string {
	return `DROP TABLE IF EXISTS global_ref_by_source CASCADE;`
}

// dbGlobalRefByFile represents the 'global_ref_by_file' table. This table
// tells us which sources (typically repositories) reference a definition
// and exactly where they do (file/line/col).
type dbGlobalRefByFile struct{}

func (*dbGlobalRefByFile) CreateTable() string {
	return `CREATE table global_ref_by_file (
		-- language represents the language that this reference was created by.
		-- See defs.go for a complete list of IDs.
		language smallint NOT NULL,

		-- def_name represents the name of the definition being referenced by
		-- this source.
		def_name integer references global_ref_name(id) NOT NULL,

		-- def_container represents the container name of the definition being
		-- referenced by this source.
		def_container integer references global_ref_container(id) NOT NULL,

		-- def_scheme, def_source, def_version, and def_file all speak about
		-- the URI where the definition itself can be located.
		def_scheme smallint NOT NULL,
		def_source integer references global_ref_source(id) NOT NULL,
		def_version integer references global_ref_version(id) NOT NULL,
		def_file integer references global_ref_file(id) NOT NULL,

		-- scheme, source, version and file speak about the source URI that is
		-- making a reference to the def.
		scheme smallint NOT NULL,
		source integer references global_ref_source(id) NOT NULL,
		version integer references global_ref_version(id) NOT NULL,
		file integer references global_ref_file(id) NOT NULL,

		-- Positions is an interleaved array of positions in the file where a
		-- reference to def is made. It is in the form of:
		--
		-- 	[start_line_1, start_col_1, end_line_1, end_col_1, start_line_2, start_col_2, ...]
		--
		-- Together, four values (start_line, start_col, end_line, end_col)
		-- describe the exact range in which a reference to def is made.
		positions integer[] NOT NULL,

		-- score is the score for this reference. It is arbitrarily set and
		-- used for sorting. For example, upvote/downvote buttons may
		-- increment and decrement this field, or we may assign specific
		-- sources higher values (like the Go stdlib).
		score smallint NOT NULL,
		UNIQUE(def_name, def_container, def_scheme, def_source, def_version, def_file, scheme, source, version, file)
	);
	CREATE INDEX global_ref_by_file_language ON global_ref_by_file USING btree (language);
	CREATE INDEX global_ref_by_file_def_name ON global_ref_by_file USING btree (def_name);
	CREATE INDEX global_ref_by_file_def_container ON global_ref_by_file USING btree (def_container);
	CREATE INDEX global_ref_by_file_def_scheme ON global_ref_by_file USING btree (def_scheme);
	CREATE INDEX global_ref_by_file_def_source ON global_ref_by_file USING btree (def_source);
	CREATE INDEX global_ref_by_file_def_version ON global_ref_by_file USING btree (def_version);
	CREATE INDEX global_ref_by_file_scheme ON global_ref_by_file USING btree (scheme);
	CREATE INDEX global_ref_by_file_source ON global_ref_by_file USING btree (source);
	CREATE INDEX global_ref_by_file_version ON global_ref_by_file USING btree (version);
	CREATE INDEX global_ref_by_file_score ON global_ref_by_file USING btree (score);
	`
}

func (*dbGlobalRefByFile) DropTable() string {
	return `DROP TABLE IF EXISTS global_ref_by_file CASCADE;`
}

type globalRefs struct{}

// RefreshIndex refreshes the global refs index for the specified repository.
func (g *globalRefs) RefreshIndex(ctx context.Context, repoID int32, commit string) error {
	// Determine the repo's URI.
	repo, err := Repos.Get(ctx, repoID)
	if err != nil {
		return errors.Wrap(err, "Repos.Get")
	}

	// TODO(slimsag): use inventory
	languages := []dbLang{dbLangGo}
	var errs []string
	for _, language := range languages {
		if err := g.refreshIndexForLanguage(ctx, language, repo.URI, commit); err != nil {
			log15.Crit("refreshing index failed", "language", language, "error", err)
			errs = append(errs, fmt.Sprintf("refreshing index failed language=%s error=%v", language, err))
		}
	}
	if len(errs) == 1 {
		return errors.New(errs[0])
	} else if len(errs) > 1 {
		return errors.New(strings.Join(errs, "\n"))
	}
	return nil
}

func (g *globalRefs) TotalRefs(ctx context.Context, source string) (int, error) {
	row := globalGraphDBH.Db.QueryRow(`SELECT count(DISTINCT source)
		FROM global_ref_by_source
		WHERE def_source IN (SELECT id FROM global_ref_source WHERE source=$1);
	`, source)
	var totalSources int
	return totalSources, row.Scan(&totalSources)
}

func (g *globalRefs) TopDefs(ctx context.Context, op sourcegraph.TopDefsOptions) (*sourcegraph.TopDefs, error) {
	// There is one row entry for every source referencing a def. i.e. there
	// will be multiple rows for the same (def_name, def_container) pair: one
	// per source (e.g. repo). We need all of them, so we can identify the sum
	// of the refs field (total number of refs to the def across all sources).
	//
	// To do this, we first perform our op.Limit by selecting the
	// (def_name, def_container) pairs that we are interested in:
	//
	// 	SELECT def_name, def_container FROM global_ref_by_source WHERE def_source IN (
	// 		SELECT id FROM global_ref_source WHERE source='github.com/golang/go'
	// 	) ORDER BY refs DESC, score DESC LIMIT 3;
	//
	// Next, we would query all rows for the (def_name, def_container) pairs,
	// since any other way or writing it would exceed LIMIT and not return the
	// right total refs count. To avoid roundtripping, we double-down and use
	// two of the above subqueries (one for a `def_name` WHERE IN clause, and
	// another for a `def_container` WHERE IN clause). The subquery represents
	// the exact amount of data we expect, so it should always be very quick.
	rows, err := globalGraphDBH.Db.Query(`SELECT def_scheme, global_ref_source.source as def_source, global_ref_version.version as def_version, global_ref_name.name as def_name, global_ref_container.container as def_container, array_agg(files) as files, array_agg(refs) as refs
		FROM global_ref_by_source
		JOIN global_ref_name ON (global_ref_name.id = global_ref_by_source.def_name)
		JOIN global_ref_container ON (global_ref_container.id = global_ref_by_source.def_container)
		JOIN global_ref_source ON (global_ref_source.id = global_ref_by_source.def_source)
		JOIN global_ref_version ON (global_ref_version.id = global_ref_by_source.def_version)

		-- Omit references to unnamed symbols. I.e., when a source references
		-- another source in an unnamed way (a Go import statement, JS require,
		-- etc) since these do not match our expectation of 'top defs'.
		WHERE global_ref_name.name != ''
		AND def_name in (
			SELECT def_name FROM global_ref_by_source WHERE def_source IN (
					SELECT id FROM global_ref_source WHERE source=$1
			) ORDER BY refs DESC, score DESC LIMIT $2
		)
		AND def_container in (
			SELECT def_container FROM global_ref_by_source WHERE def_source IN (
					SELECT id FROM global_ref_source WHERE source=$1
			) ORDER BY refs DESC, score DESC LIMIT $2
		)
		GROUP BY global_ref_name.name, global_ref_container.container, def_scheme, global_ref_source.source, global_ref_version.version
		ORDER BY sum(refs) DESC, sum(score) DESC;`, op.Source, op.Limit)
	if err != nil {
		return nil, errors.Wrap(err, "query")
	}
	defer rows.Close()

	topDefs := &sourcegraph.TopDefs{}
	for rows.Next() {
		var (
			scheme                           int16
			source, version, name, container string

			// Counts per source.
			files = make(pq.Int64Array, op.Limit)
			refs  = make(pq.Int64Array, op.Limit)
		)
		if err := rows.Scan(&scheme, &source, &version, &name, &container, &files, &refs); err != nil {
			return nil, errors.Wrap(err, "Scan")
		}
		topDefs.SourceDefs = append(topDefs.SourceDefs, &sourcegraph.SourceDef{
			DefScheme:        uriSchemeIDLookup[scheme],
			DefSource:        source,
			DefVersion:       version,
			DefName:          name,
			DefContainerName: container,
			Sources:          len(refs),
			Files:            int(sumInt64(files)),
			Refs:             int(sumInt64(refs)),
		})
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows error")
	}
	return topDefs, nil
}

func (g *globalRefs) RefLocations(ctx context.Context, op sourcegraph.RefLocationsOptions) (*sourcegraph.RefLocations, error) {
	refLocations := &sourcegraph.RefLocations{}

	// Note: It's impossible to run these queries in a transaction because we
	// are running two queries: https://github.com/lib/pq/issues/81
	var err error
	refLocations.TotalSources, err = g.queryTotalReposReferencing(globalGraphDBH.Db, op)
	if err != nil {
		return nil, errors.Wrap(err, "queryTotalReposReferencing")
	}
	if refLocations.TotalSources == 0 {
		// In the case of no references, abort early to avoid doing more work
		// than we have to.
		return refLocations, nil
	}

	refLocations.SourceRefs, err = g.queryRefsBySource(globalGraphDBH.Db, op)
	if err != nil {
		return nil, errors.Wrap(err, "queryRefsBySource")
	}
	return refLocations, nil
}

func (g *globalRefs) queryTotalReposReferencing(db *sql.DB, op sourcegraph.RefLocationsOptions) (int, error) {
	row := db.QueryRow(`SELECT count(source)
		FROM global_ref_by_source
		WHERE def_source IN (SELECT id FROM global_ref_source WHERE source=$1)
		AND def_name IN (SELECT id FROM global_ref_name WHERE name=$2)
		AND def_container IN (SELECT id FROM global_ref_container WHERE container=$3);
	`, op.Source, op.Name, op.ContainerName)
	var totalSources int
	return totalSources, row.Scan(&totalSources)
}

func (g *globalRefs) queryRefsBySource(db *sql.DB, op sourcegraph.RefLocationsOptions) ([]*sourcegraph.SourceRef, error) {
	rows, err := db.Query(`SELECT scheme, global_ref_source.source, global_ref_version.version, files, refs, score
		FROM global_ref_by_source
		JOIN global_ref_source ON (global_ref_source.id = global_ref_by_source.source)
		JOIN global_ref_version ON (global_ref_version.id = global_ref_by_source.version)
		WHERE def_source IN (SELECT id FROM global_ref_source WHERE source=$1)
		AND def_name IN (SELECT id FROM global_ref_name WHERE name=$2)
		AND def_container IN (SELECT id FROM global_ref_container WHERE container=$3)
		ORDER BY refs DESC, score DESC LIMIT $4;
	`, op.Source, op.Name, op.ContainerName, op.Sources)
	if err != nil {
		return nil, errors.Wrap(err, "Query")
	}
	defer rows.Close()
	var refsBySource []*sourcegraph.SourceRef
	for rows.Next() {
		var (
			scheme, score   int16
			source, version string
			files, refs     int
		)
		if err := rows.Scan(&scheme, &source, &version, &files, &refs, &score); err != nil {
			return nil, errors.Wrap(err, "Scan")
		}
		fileRefs, err := g.queryRefsByFile(db, source, op)
		if err != nil {
			return nil, errors.Wrap(err, "queryRefsByFile")
		}
		refsBySource = append(refsBySource, &sourcegraph.SourceRef{
			Scheme:   uriSchemeIDLookup[scheme],
			Source:   source,
			Version:  version,
			Files:    files,
			Refs:     refs,
			Score:    score,
			FileRefs: fileRefs,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows error")
	}
	return refsBySource, nil
}

func (g *globalRefs) queryRefsByFile(db *sql.DB, source string, op sourcegraph.RefLocationsOptions) ([]*sourcegraph.FileRef, error) {
	rows, err := db.Query(`SELECT scheme, global_ref_source.source, global_ref_version.version, global_ref_file.file, positions, score
		FROM global_ref_by_file
		JOIN global_ref_source ON (global_ref_source.id = global_ref_by_file.source)
		JOIN global_ref_version ON (global_ref_version.id = global_ref_by_file.version)
		JOIN global_ref_file ON (global_ref_file.id = global_ref_by_file.file)
		WHERE def_source IN (SELECT id FROM global_ref_source WHERE source=$1)
		AND def_name IN (SELECT id FROM global_ref_name WHERE name=$2)
		AND def_container IN (SELECT id FROM global_ref_container WHERE container=$3)
		AND global_ref_source.source = $4
		ORDER BY score DESC, global_ref_by_file.source LIMIT $5;
	`, op.Source, op.Name, op.ContainerName, source, op.Files)
	if err != nil {
		return nil, errors.Wrap(err, "Query")
	}
	defer rows.Close()
	var refsByFile []*sourcegraph.FileRef
	for rows.Next() {
		var (
			scheme, score         int16
			source, version, file string
			positions             = make(pq.Int64Array, op.Files)
		)
		if err := rows.Scan(&scheme, &source, &version, &file, &positions, &score); err != nil {
			return nil, err
		}
		refsByFile = append(refsByFile, &sourcegraph.FileRef{
			Scheme:    uriSchemeIDLookup[scheme],
			Source:    source,
			Version:   version,
			File:      file,
			Positions: deinterlacePositions(positions),
			Score:     score,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows error")
	}
	return refsByFile, nil
}

func (g *globalRefs) refreshIndexForLanguage(ctx context.Context, language dbLang, repoURI, commit string) error {
	// Query all external references for the repository.
	var refs []lspext.ReferenceInformation
	rootPath := "git://" + repoURI + "?" + commit
	err := xlang.OneShotClientRequest(ctx, language.String(), rootPath, "workspace/reference", lspext.WorkspaceReferenceParams{}, &refs)
	if err != nil {
		return errors.Wrap(err, "workspaceReference")
	}

	err = dbutil.Transaction(ctx, globalGraphDBH.Db, func(tx *sql.Tx) error {
		// Update the global_ref_source table.
		sourceIDs, err := g.updateKey(ctx, tx, refs, "global_ref_source", "source", func(u *parsedURI) string {
			return u.source
		})
		if err != nil {
			return errors.Wrap(err, "update global_ref_source")
		}

		// Update the global_ref_version table.
		versionIDs, err := g.updateKey(ctx, tx, refs, "global_ref_version", "version", func(u *parsedURI) string {
			return u.version
		})
		if err != nil {
			return errors.Wrap(err, "update global_ref_version")
		}

		// Update the global_ref_file table.
		fileIDs, err := g.updateKey(ctx, tx, refs, "global_ref_file", "file", func(u *parsedURI) string {
			return u.file
		})
		if err != nil {
			return errors.Wrap(err, "update global_ref_file")
		}

		// Update the global_ref_name table.
		nameIDs, err := g.updateRefField(ctx, tx, refs, "global_ref_name", "name", func(r lspext.ReferenceInformation) string {
			return r.Name
		})
		if err != nil {
			return errors.Wrap(err, "update global_ref_name")
		}

		// Update the global_ref_name table.
		containerIDs, err := g.updateRefField(ctx, tx, refs, "global_ref_container", "container", func(r lspext.ReferenceInformation) string {
			return r.ContainerName
		})
		if err != nil {
			return errors.Wrap(err, "update global_ref_container")
		}

		// Update the global_ref_by_file table.
		err = g.updateRefByFile(ctx, tx, language, refs, sourceIDs, versionIDs, fileIDs, nameIDs, containerIDs)
		if err != nil {
			return errors.Wrap(err, "update global_ref_by_file")
		}

		// Update the global_ref_by_source table.
		err = g.updateRefBySource(ctx, tx, language, refs, sourceIDs, versionIDs, fileIDs, nameIDs, containerIDs)
		if err != nil {
			return errors.Wrap(err, "update global_ref_by_source")
		}

		// Evict unused table data.
		err = g.evictUnusedData(ctx, tx)
		if err != nil {
			return errors.Wrap(err, "evicting unused data")
		}
		return nil
	})
	if err != nil {
		return errors.Wrap(err, "executing transaction")
	}
	return nil
}

// updateKey updates a global_ref_(source,version,file) table.
func (g *globalRefs) updateKey(ctx context.Context, tx *sql.Tx, refs []lspext.ReferenceInformation, table, field string, uriToKey func(u *parsedURI) string) (ids map[string]int32, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "update "+table)
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.SetTag("err", err.Error())
		}
		span.Finish()
	}()

	// Keys are often repeated across symbols, so keeping a mapping allows
	// us to omit a great number of query executions and provides us the
	// ability of looking up IDs by name later on.
	m := make(map[string]int32, 128)
	for _, r := range refs {
		for _, uri := range []string{r.Location.URI, r.URI} {
			u, err := parseURI(uri)
			if err != nil {
				return nil, err
			}
			key := uriToKey(u)
			if _, alreadyUpserted := m[key]; alreadyUpserted {
				continue
			}
			var id int32
			row := tx.QueryRow(`INSERT INTO `+table+`(`+field+`) VALUES($1) ON CONFLICT(`+field+`) DO UPDATE SET `+field+`=$1 RETURNING id`, key)
			if err := row.Scan(&id); err != nil {
				return nil, errors.Wrap(err, "scanning row")
			}
			m[key] = id
		}
	}
	return m, nil
}

// updateRefField updates a global_ref_(name,container) table.
func (g *globalRefs) updateRefField(ctx context.Context, tx *sql.Tx, refs []lspext.ReferenceInformation, table, field string, fieldToKey func(r lspext.ReferenceInformation) string) (ids map[string]int32, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "update "+table)
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.SetTag("err", err.Error())
		}
		span.Finish()
	}()

	// Keys are often repeated across symbols, so keeping a mapping allows
	// us to omit a great number of query executions and provides us the
	// ability of looking up IDs by name later on.
	m := make(map[string]int32, 128)
	for _, r := range refs {
		key := fieldToKey(r)
		if _, alreadyUpserted := m[key]; alreadyUpserted {
			continue
		}
		var id int32
		row := tx.QueryRow(`INSERT INTO `+table+`(`+field+`) VALUES($1) ON CONFLICT(`+field+`) DO UPDATE SET `+field+`=$1 RETURNING id`, key)
		if err := row.Scan(&id); err != nil {
			return nil, errors.Wrap(err, "scanning row")
		}
		m[key] = id
	}
	return m, nil
}

// updateRefByFile updates the global_ref_by_file table.
func (g *globalRefs) updateRefByFile(ctx context.Context, tx *sql.Tx, language dbLang, refs []lspext.ReferenceInformation, sourceIDs, versionIDs, fileIDs, nameIDs, containerIDs map[string]int32) (err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "updateRefByFile")
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.SetTag("err", err.Error())
		}
		span.Finish()
	}()
	span.SetTag("refs", len(refs))

	// First, create a temporary table. Note that this must reflect the same
	// exact fields present in the global_refs_by_file table.
	_, err = tx.Exec(`CREATE TEMPORARY TABLE new_refs_by_file (
		language smallint,
		def_name integer,
		def_container integer,
		def_scheme smallint,
		def_source integer,
		def_version integer,
		def_file integer,
		scheme smallint,
		source integer,
		version integer,
		file integer,
		positions integer[],
		score smallint
	) ON COMMIT DROP;`)
	if err != nil {
		return errors.Wrap(err, "create temp table")
	}

	// Copy the new refs into the temporary table.
	copy, err := tx.Prepare(pq.CopyIn("new_refs_by_file",
		"language",
		"def_name",
		"def_container",
		"def_scheme",
		"def_source",
		"def_version",
		"def_file",
		"scheme",
		"source",
		"version",
		"file",
		"positions",
		"score",
	))
	if err != nil {
		return errors.Wrap(err, "prepare copy in")
	}
	for _, r := range byFile(refs) {
		// Parse URIs.
		defURI, err := parseURI(r.URI)
		if err != nil {
			return errors.Wrap(err, "parse r.URI")
		}
		uri, err := parseURI(r.Location.URI)
		if err != nil {
			return errors.Wrap(err, "parse r.Location.URI")
		}

		// Translate respective fields into IDs and copy into the temp table.
		if _, err := copy.Exec(
			language,                      // language
			nameIDs[r.Name],               // def_name
			containerIDs[r.ContainerName], // def_container
			defURI.scheme,                 // def_scheme
			sourceIDs[defURI.source],      // def_source
			versionIDs[defURI.version],    // def_version
			fileIDs[defURI.file],          // def_file
			uri.scheme,                    // scheme
			sourceIDs[uri.source],         // source
			versionIDs[uri.version],       // version
			fileIDs[uri.file],             // file
			pq.Array(r.positions),         // positions
			0, // score
		); err != nil {
			return errors.Wrap(err, "executing ref copy")
		}
	}
	if _, err := copy.Exec(); err != nil {
		return errors.Wrap(err, "executing copy")
	}

	// Delete existing refs from the real table.
	deletions := make(map[parsedURI]struct{}, 64)
	for _, r := range refs {
		uri, err := parseURI(r.Location.URI)
		if err != nil {
			return errors.Wrap(err, "parse r.Location.URI")
		}
		// Build a map of URIs to be deleted. We delete all references that are
		// in the same scheme://source, regardless of version or file. For
		// example, delete all refs in git://github.com/gorilla/mux or all refs
		// in npm://react.
		//
		// We don't actually index npm etc yet, so this isn't a concern right
		// now, but in the future we will need to make this deletion scheme
		// smarter. For example, index only the default branch of git repos,
		// while in contrast indexing any major (but not minor or patch)
		// version of an npm package or Rust crate.
		deletions[parsedURI{
			scheme: uri.scheme,
			source: uri.source,
		}] = struct{}{}
	}
	for refURI := range deletions {
		if _, err := tx.Exec(`DELETE FROM global_ref_by_file WHERE language=$1 AND scheme=$2 AND source=$3`, language, refURI.scheme, sourceIDs[refURI.source]); err != nil {
			return errors.Wrap(err, "executing global_ref_by_file deletion")
		}
	}

	// Insert from temporary table into the real table.
	_, err = tx.Exec(`INSERT INTO global_ref_by_file(
		language,
		def_name,
		def_container,
		def_scheme,
		def_source,
		def_version,
		def_file,
		scheme,
		source,
		version,
		file,
		positions,
		score
	) SELECT d.language,
		d.def_name,
		d.def_container,
		d.def_scheme,
		d.def_source,
		d.def_version,
		d.def_file,
		d.scheme,
		d.source,
		d.version,
		d.file,
		d.positions,
		d.score
	FROM new_refs_by_file d;`)
	if err != nil {
		return errors.Wrap(err, "executing final insertion from temp table")
	}
	return nil
}

// updateRefBySource updates the global_ref_by_source table.
func (g *globalRefs) updateRefBySource(ctx context.Context, tx *sql.Tx, language dbLang, refs []lspext.ReferenceInformation, sourceIDs, versionIDs, fileIDs, nameIDs, containerIDs map[string]int32) (err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "updateRefBySource")
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.SetTag("err", err.Error())
		}
		span.Finish()
	}()
	span.SetTag("refs", len(refs))

	// First, create a temporary table. Note that this must reflect the same
	// exact fields present in the global_refs_by_source table.
	_, err = tx.Exec(`CREATE TEMPORARY TABLE new_refs_by_source (
		language smallint,
		def_name integer,
		def_container integer,
		def_scheme smallint,
		def_source integer,
		def_version integer,
		def_file integer,
		scheme smallint,
		source integer,
		version integer,
		files smallint,
		refs smallint,
		score smallint
	) ON COMMIT DROP;`)
	if err != nil {
		return errors.Wrap(err, "create temp table")
	}

	// Copy the new refs into the temporary table.
	copy, err := tx.Prepare(pq.CopyIn("new_refs_by_source",
		"language",
		"def_name",
		"def_container",
		"def_scheme",
		"def_source",
		"def_version",
		"def_file",
		"scheme",
		"source",
		"version",
		"files",
		"refs",
		"score",
	))
	if err != nil {
		return errors.Wrap(err, "prepare copy in")
	}
	srcRefs, err := bySource(refs)
	if err != nil {
		return errors.Wrap(err, "bySource")
	}
	for _, r := range srcRefs {
		// Parse URIs.
		defURI, err := parseURI(r.URI)
		if err != nil {
			return errors.Wrap(err, "parse r.URI")
		}
		uri, err := parseURI(r.Location.URI)
		if err != nil {
			return errors.Wrap(err, "parse r.Location.URI")
		}

		// Translate respective fields into IDs and copy into the temp table.
		if _, err := copy.Exec(
			language,                      // language
			nameIDs[r.Name],               // def_name
			containerIDs[r.ContainerName], // def_container
			defURI.scheme,                 // def_scheme
			sourceIDs[defURI.source],      // def_source
			versionIDs[defURI.version],    // def_version
			fileIDs[defURI.file],          // def_file
			uri.scheme,                    // scheme
			sourceIDs[uri.source],         // source
			versionIDs[uri.version],       // version
			len(r.files),                  // files
			r.refs,                        // refs
			0,                             // score
		); err != nil {
			return errors.Wrap(err, "executing ref copy")
		}
	}
	if _, err := copy.Exec(); err != nil {
		return errors.Wrap(err, "executing copy")
	}

	// Delete existing refs from the real table.
	deletions := make(map[parsedURI]struct{}, 64)
	for _, r := range refs {
		uri, err := parseURI(r.Location.URI)
		if err != nil {
			return errors.Wrap(err, "parse r.Location.URI")
		}
		// Build a map of URIs to be deleted. We delete all references that are
		// in the same scheme://source, regardless of version or file. For
		// example, delete all refs in git://github.com/gorilla/mux or all refs
		// in npm://react.
		//
		// We don't actually index npm etc yet, so this isn't a concern right
		// now, but in the future we will need to make this deletion scheme
		// smarter. For example, index only the default branch of git repos,
		// while in contrast indexing any major (but not minor or patch)
		// version of an npm package or Rust crate.
		deletions[parsedURI{
			scheme: uri.scheme,
			source: uri.source,
		}] = struct{}{}
	}
	for refURI := range deletions {
		if _, err := tx.Exec(`DELETE FROM global_ref_by_source WHERE language=$1 AND scheme=$2 AND source=$3`, language, refURI.scheme, sourceIDs[refURI.source]); err != nil {
			return errors.Wrap(err, "executing global_ref_by_source deletion")
		}
	}

	// Insert from temporary table into the real table.
	_, err = tx.Exec(`INSERT INTO global_ref_by_source(
			language,
			def_name,
			def_container,
			def_scheme,
			def_source,
			def_version,
			def_file,
			scheme,
			source,
			version,
			files,
			refs,
			score
		) SELECT d.language,
			d.def_name,
			d.def_container,
			d.def_scheme,
			d.def_source,
			d.def_version,
			d.def_file,
			d.scheme,
			d.source,
			d.version,
			d.files,
			d.refs,
			d.score
		FROM new_refs_by_source d;`)
	if err != nil {
		return errors.Wrap(err, "executing final insertion from temp table")
	}
	return nil
}

// evictUnusedData is responsible for evicting unused textual data from the
// global_ref_(source,version,file,container) tables. The data can become
// unreferenced after global_ref_by_* are updated, and as such we sweep them
// afterwards.
func (g *globalRefs) evictUnusedData(ctx context.Context, tx *sql.Tx) (err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "evictUnusedData")
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.SetTag("err", err.Error())
		}
		span.Finish()
	}()

	result, err := tx.Exec(`
-- evict global_ref_source when there are no more references to it
DELETE FROM global_ref_source
WHERE NOT EXISTS(
	SELECT NULL FROM global_ref_by_file f WHERE f.source = id OR f.def_source = id
) AND NOT EXISTS(
	SELECT NULL FROM global_ref_by_source f WHERE f.source = id OR f.def_source = id
);

-- evict global_ref_version when there are no more references to it
DELETE FROM global_ref_version
WHERE NOT EXISTS(
	SELECT NULL FROM global_ref_by_file f WHERE f.version = id OR f.def_version = id
) AND NOT EXISTS(
	SELECT NULL FROM global_ref_by_source f WHERE f.version = id OR f.def_version = id
);

-- evict global_ref_file when there are no more references to it
DELETE FROM global_ref_file
WHERE NOT EXISTS(
	SELECT NULL FROM global_ref_by_file f WHERE f.file = id OR f.def_file = id
) AND NOT EXISTS(
	SELECT NULL FROM global_ref_by_source f WHERE f.def_file = id
);

-- evict global_ref_name when there are no more references to it
DELETE FROM global_ref_name
WHERE NOT EXISTS(
	SELECT NULL FROM global_ref_by_file f WHERE f.def_name = id
) AND NOT EXISTS(
	SELECT NULL FROM global_ref_by_source f WHERE f.def_name = id
);

-- evict global_ref_container when there are no more references to it
DELETE FROM global_ref_container
WHERE NOT EXISTS(
	SELECT NULL FROM global_ref_by_file f WHERE f.def_container = id
) AND NOT EXISTS(
	SELECT NULL FROM global_ref_by_source f WHERE f.def_container = id
);
`)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	span.SetTag("RowsAffected", n)
	return nil
}

// interlaceRange is a small helper to convert an LSP range into its interlaced
// form. See deinterlacePositions below for more information.
func interlaceRange(r lsp.Range) []int {
	return []int{
		r.Start.Line,
		r.Start.Character,
		r.End.Line,
		r.End.Character,
	}
}

// deinterlacePositions deinterlaces the interlaced:
//
// 	[start_line_1, start_col_1, end_line_1, end_col_1, start_line_2, start_col_2, ...]
//
// slice p into its non-interlaced form. We store the positions in the DB
// interlaced because github.com/lib/pq does not support multidimensional array
// types (and implementing this is tedious).
func deinterlacePositions(p []int64) (out []lsp.Range) {
	if len(p)%4 != 0 {
		panic("deprecatedDeinterlacePositions: unequal length array (bad data?)")
	}
	for i := 0; i < len(p); i += 4 {
		out = append(out, lsp.Range{
			Start: lsp.Position{
				Line:      int(p[i]),
				Character: int(p[i+1]),
			},
			End: lsp.Position{
				Line:      int(p[i+2]),
				Character: int(p[i+3]),
			},
		})
	}
	return
}

// parsedURI represents a parsed URI, see the parseURI function below for more
// details.
type parsedURI struct {
	scheme                int16
	source, version, file string
}

// clearFileFromURI clears the file portion of the URI. See parseURI
// for what is meant by the file portion.
func clearFileFromURI(uri string) (string, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("parsing URI %q", uri))
	}
	u.Fragment = ""
	return u.String(), nil
}

// parseURI parses a URI like e.g.:
//
//  (Go) git://github.com/golang/go?go1.7.1#src/net/http
//  (Rust) cargo://winapi?0.2.8#src/d3d9.rs
//  (JS) npm://react?16.1.0#something.js
//
// And returns its scheme, source ("github.com/golang/go", "winapi", "react"),
// version ("go1.7.1", "0.2.8", "16.1.0") and file ("src/net/http",
// "src/d3d9.rs", "something.js").
func parseURI(uri string) (*parsedURI, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("parsing URI %q", uri))
	}
	scheme, ok := uriSchemeID[u.Scheme]
	if !ok {
		return nil, fmt.Errorf("parseURI %q: scheme %q is not registered for global refs", uri, u.Scheme)
	}
	return &parsedURI{
		scheme:  scheme,
		source:  u.Host + u.Path,
		version: u.RawQuery,
		file:    u.Fragment,
	}, nil
}

// refByFile represents one or more external references to a single symbol,
// such that positions accurately describes all of the locations in
// .Location.URI where the symbol is referenced.
type refByFile struct {
	lspext.ReferenceInformation
	positions []int
}

// byFile aggregates all references in the same file to the same external
// symbol.
func byFile(refs []lspext.ReferenceInformation) map[lspext.ReferenceInformation]*refByFile {
	out := make(map[lspext.ReferenceInformation]*refByFile, len(refs))
	for _, r := range refs {
		key := r
		key.Location.Range = lsp.Range{}
		if refPos, ok := out[key]; ok {
			refPos.positions = append(refPos.positions, interlaceRange(r.Location.Range)...)
			continue
		}
		out[key] = &refByFile{
			ReferenceInformation: r,
			positions:            interlaceRange(r.Location.Range),
		}
	}
	return out
}

// refBySource represents all of the external references to a single symbol
// within an entire source (e.g. a repository).
type refBySource struct {
	lspext.ReferenceInformation
	files map[string]struct{}
	refs  int
}

// bySource aggregates all references in the same source to the same external
// symbol. It is guaranteed that each slice in the returned map will have at
// least one element.
func bySource(refs []lspext.ReferenceInformation) (map[lspext.ReferenceInformation]*refBySource, error) {
	out := make(map[lspext.ReferenceInformation]*refBySource, len(refs))
	for _, r := range refs {
		u, err := parseURI(r.Location.URI)
		if err != nil {
			return nil, err
		}

		key := r
		key.Location.Range = lsp.Range{}
		key.Location.URI, err = clearFileFromURI(key.Location.URI)
		if err != nil {
			return nil, err
		}

		if refSrc, ok := out[key]; ok {
			refSrc.refs += 1
			refSrc.files[u.file] = struct{}{}
			continue
		}
		out[key] = &refBySource{
			ReferenceInformation: r,
			files: map[string]struct{}{
				u.file: struct{}{},
			},
			refs: 1,
		}
	}
	return out, nil
}

func sumInt64(s []int64) (sum int64) {
	for _, v := range s {
		sum += v
	}
	return
}
