package lsifstore

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/cockroachdb/errors"
	"github.com/inconshreveable/log15"
	"github.com/keegancsmith/sqlf"
	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/batch"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
)

// WriteDocumentationPages is called (transactionally) from the precise-code-intel-worker.
func (s *Store) WriteDocumentationPages(ctx context.Context, upload dbstore.Upload, repo *types.Repo, isDefaultBranch bool, documentationPages chan *precise.DocumentationPageData) (err error) {
	ctx, traceLog, endObservation := s.operations.writeDocumentationPages.WithAndLogger(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("bundleID", upload.ID),
	}})
	defer endObservation(1, observation.Args{})

	tx, err := s.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	// Create temporary table symmetric to lsif_data_documentation_pages without the dump id
	if err := tx.Exec(ctx, sqlf.Sprintf(writeDocumentationPagesTemporaryTableQuery)); err != nil {
		return err
	}

	var count uint32
	var pages []*precise.DocumentationPageData
	inserter := func(inserter *batch.Inserter) error {
		for page := range documentationPages {
			pages = append(pages, page)
			data, err := s.serializer.MarshalDocumentationPageData(page)
			if err != nil {
				return err
			}

			if err := inserter.Insert(ctx, page.Tree.PathID, data); err != nil {
				return err
			}

			atomic.AddUint32(&count, 1)
		}
		return nil
	}

	// Bulk insert all the unique column values into the temporary table
	if err := withBatchInserter(
		ctx,
		tx.Handle().DB(),
		"t_lsif_data_documentation_pages",
		[]string{"path_id", "data"},
		inserter,
	); err != nil {
		return err
	}
	traceLog(log.Int("numResultChunkRecords", int(count)))

	// Note: If someone disables API docs search indexing, uploads during that time will not be
	// indexed even if it is turned back on. Only future uploads would be.
	if conf.APIDocsSearchIndexingEnabled() {
		// Perform search indexing for API docs pages.
		if err := tx.WriteDocumentationSearch(ctx, upload, repo, isDefaultBranch, pages); err != nil {
			return errors.Wrap(err, "WriteDocumentationSearch")
		}
	}

	// Insert the values from the temporary table into the target table. We select a
	// parameterized dump id here since it is the same for all rows in this operation.
	return tx.Exec(ctx, sqlf.Sprintf(writeDocumentationPagesInsertQuery, upload.ID))
}

const writeDocumentationPagesTemporaryTableQuery = `
-- source: enterprise/internal/codeintel/stores/lsifstore/data_write_documentation.go:WriteDocumentationPages
CREATE TEMPORARY TABLE t_lsif_data_documentation_pages (
	path_id TEXT NOT NULL,
	data bytea NOT NULL
) ON COMMIT DROP
`

const writeDocumentationPagesInsertQuery = `
-- source: enterprise/internal/codeintel/stores/lsifstore/data_write_documentation.go:WriteDocumentationPages
INSERT INTO lsif_data_documentation_pages (dump_id, path_id, data, search_indexed)
SELECT %s, source.path_id, source.data, 'true'
FROM t_lsif_data_documentation_pages source
`

// WriteDocumentationPathInfo is called (transactionally) from the precise-code-intel-worker.
func (s *Store) WriteDocumentationPathInfo(ctx context.Context, bundleID int, documentationPathInfo chan *precise.DocumentationPathInfoData) (err error) {
	ctx, traceLog, endObservation := s.operations.writeDocumentationPathInfo.WithAndLogger(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("bundleID", bundleID),
	}})
	defer endObservation(1, observation.Args{})

	tx, err := s.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	// Create temporary table symmetric to lsif_data_documentation_path_info without the dump id
	if err := tx.Exec(ctx, sqlf.Sprintf(writeDocumentationPathInfoTemporaryTableQuery)); err != nil {
		return err
	}

	var count uint32
	inserter := func(inserter *batch.Inserter) error {
		for v := range documentationPathInfo {
			data, err := s.serializer.MarshalDocumentationPathInfoData(v)
			if err != nil {
				return err
			}

			if err := inserter.Insert(ctx, v.PathID, data); err != nil {
				return err
			}

			atomic.AddUint32(&count, 1)
		}
		return nil
	}

	// Bulk insert all the unique column values into the temporary table
	if err := withBatchInserter(
		ctx,
		tx.Handle().DB(),
		"t_lsif_data_documentation_path_info",
		[]string{"path_id", "data"},
		inserter,
	); err != nil {
		return err
	}
	traceLog(log.Int("numResultChunkRecords", int(count)))

	// Insert the values from the temporary table into the target table. We select a
	// parameterized dump id here since it is the same for all rows in this operation.
	return tx.Exec(ctx, sqlf.Sprintf(writeDocumentationPathInfoInsertQuery, bundleID))
}

const writeDocumentationPathInfoTemporaryTableQuery = `
-- source: enterprise/internal/codeintel/stores/lsifstore/data_write_documentation.go:WriteDocumentationPathInfo
CREATE TEMPORARY TABLE t_lsif_data_documentation_path_info (
	path_id TEXT NOT NULL,
	data bytea NOT NULL
) ON COMMIT DROP
`

const writeDocumentationPathInfoInsertQuery = `
-- source: enterprise/internal/codeintel/stores/lsifstore/data_write_documentation.go:WriteDocumentationPathInfo
INSERT INTO lsif_data_documentation_path_info (dump_id, path_id, data)
SELECT %s, source.path_id, source.data
FROM t_lsif_data_documentation_path_info source
`

// WriteDocumentationMappings is called (transactionally) from the precise-code-intel-worker.
func (s *Store) WriteDocumentationMappings(ctx context.Context, bundleID int, mappings chan precise.DocumentationMapping) (err error) {
	ctx, traceLog, endObservation := s.operations.writeDocumentationMappings.WithAndLogger(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("bundleID", bundleID),
	}})
	defer endObservation(1, observation.Args{})

	tx, err := s.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	// Create temporary table symmetric to lsif_data_documentation_mappings without the dump id
	if err := tx.Exec(ctx, sqlf.Sprintf(writeDocumentationMappingsTemporaryTableQuery)); err != nil {
		return err
	}

	var count uint32
	inserter := func(inserter *batch.Inserter) error {
		for mapping := range mappings {
			if err := inserter.Insert(ctx, mapping.PathID, mapping.ResultID, mapping.FilePath); err != nil {
				return err
			}
			atomic.AddUint32(&count, 1)
		}
		return nil
	}

	// Bulk insert all the unique column values into the temporary table
	if err := withBatchInserter(
		ctx,
		tx.Handle().DB(),
		"t_lsif_data_documentation_mappings",
		[]string{"path_id", "result_id", "file_path"},
		inserter,
	); err != nil {
		return err
	}
	traceLog(log.Int("numRecords", int(count)))

	// Insert the values from the temporary table into the target table. We select a
	// parameterized dump id here since it is the same for all rows in this operation.
	return tx.Exec(ctx, sqlf.Sprintf(writeDocumentationMappingsInsertQuery, bundleID))
}

const writeDocumentationMappingsTemporaryTableQuery = `
-- source: enterprise/internal/codeintel/stores/lsifstore/data_write_documentation.go:WriteDocumentationMappings
CREATE TEMPORARY TABLE t_lsif_data_documentation_mappings (
	path_id TEXT NOT NULL,
	result_id integer NOT NULL,
	file_path text
) ON COMMIT DROP
`

const writeDocumentationMappingsInsertQuery = `
-- source: enterprise/internal/codeintel/stores/lsifstore/data_write_documentation.go:WriteDocumentationMappings
INSERT INTO lsif_data_documentation_mappings (dump_id, path_id, result_id, file_path)
SELECT %s, source.path_id, source.result_id, source.file_path
FROM t_lsif_data_documentation_mappings source
`

// WriteDocumentationSearch is called (within a transaction) to write the search index for a given documentation page.
func (s *Store) WriteDocumentationSearch(ctx context.Context, upload dbstore.Upload, repo *types.Repo, isDefaultBranch bool, pages []*precise.DocumentationPageData) (err error) {
	if !isDefaultBranch {
		// We do not index non-default branches for API docs search.
		return nil
	}

	ctx, traceLog, endObservation := s.operations.writeDocumentationSearch.WithAndLogger(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("repo", upload.RepositoryName),
		log.Int("bundleID", upload.ID),
		log.Int("pages", len(pages)),
	}})
	defer endObservation(1, observation.Args{})

	// This will not always produce a proper language name, e.g. if an indexer is not named after
	// the language or is not in "lsif-$LANGUAGE" format. That's OK: in that case, the "language"
	// is the indexer name which is likely good enough since we use fuzzy search / partial text matching
	// over it.
	languageOrIndexerName := strings.ToLower(strings.TrimPrefix(upload.Indexer, "lsif-"))

	tableSuffix := "public"
	if repo.Private {
		tableSuffix = "private"
	}

	// This upload is for a commit on the default branch of the repository, so it is eligible for API
	// docs search indexing. It will replace any existing data that we have or this unique (repo_id, lang, root)
	// tuple in either table so we go ahead and purge the old data now.
	for _, suffix := range []string{"public", "private"} {
		if err := s.Exec(ctx, sqlf.Sprintf(
			strings.ReplaceAll(purgeDocumentationSearchOldData, "$SUFFIX", suffix),
			textSearchVector(languageOrIndexerName), // langs CTE tsv
			upload.RepositoryID,
			upload.Root,
		)); err != nil {
			return errors.Wrap(err, "purging old data")
		}
	}

	// Upsert the language name.
	langNameID, exists, err := basestore.ScanFirstInt(s.Query(ctx, sqlf.Sprintf(
		strings.ReplaceAll(writeDocumentationSearchLangNames, "$SUFFIX", tableSuffix),
		languageOrIndexerName,                   // lang_name
		textSearchVector(languageOrIndexerName), // tsv
		textSearchVector(languageOrIndexerName), // union tsv query
	)))
	if err != nil {
		return errors.Wrap(err, "upserting language name")
	}
	if !exists {
		return fmt.Errorf("failed to upsert language name")
	}

	// Upsert the repo name.
	repoNameID, exists, err := basestore.ScanFirstInt(s.Query(ctx, sqlf.Sprintf(
		strings.ReplaceAll(writeDocumentationSearchRepoNames, "$SUFFIX", tableSuffix),
		upload.RepositoryName,                            // repo_name
		textSearchVector(upload.RepositoryName),          // tsv
		textSearchVector(reverse(upload.RepositoryName)), // reverse_tsv
		textSearchVector(upload.RepositoryName),          // union tsv query
	)))
	if err != nil {
		return errors.Wrap(err, "upserting repo name")
	}
	if !exists {
		return fmt.Errorf("failed to upsert repo name")
	}

	var index func(node *precise.DocumentationNode) error
	index = func(node *precise.DocumentationNode) error {
		if node.Documentation.SearchKey != "" {
			// Upsert the tags sequence.
			tagsSlice := []string{}
			for _, tag := range node.Documentation.Tags {
				tagsSlice = append(tagsSlice, string(tag))
			}
			tags := strings.Join(tagsSlice, " ")
			tagsID, exists, err := basestore.ScanFirstInt(s.Query(ctx, sqlf.Sprintf(
				strings.ReplaceAll(writeDocumentationSearchTags, "$SUFFIX", tableSuffix),
				tags,                   // tags
				textSearchVector(tags), // tsv
				textSearchVector(tags), // union tsv query
			)))
			if err != nil {
				return errors.Wrap(err, "upserting tags")
			}
			if !exists {
				return fmt.Errorf("failed to upsert tags")
			}

			// Insert the search result.
			label := truncate(node.Label.String(), 256)      // 256 bytes, enough for ~100 characters in all languages
			detail := truncate(node.Detail.String(), 5*1024) // 5 KiB - just for sanity
			err = s.Exec(ctx, sqlf.Sprintf(
				strings.ReplaceAll(writeDocumentationSearchInsertQuery, "$SUFFIX", tableSuffix),
				upload.RepositoryID, // repo_id
				upload.ID,           // dump_id
				upload.Root,         // dump_root
				node.PathID,         // path_id
				detail,              // detail
				langNameID,          // lang_name_id
				repoNameID,          // repo_name_id
				tagsID,              // tags_id

				node.Documentation.SearchKey,                            // search_key
				textSearchVector(node.Documentation.SearchKey),          // search_key_tsv
				textSearchVector(reverse(node.Documentation.SearchKey)), // search_key_reverse_tsv

				label,                            // label
				textSearchVector(label),          // label_tsv
				textSearchVector(reverse(label)), // label_reverse_tsv
			))
			if err != nil {
				return err
			}
		}

		// Index descendants.
		for _, child := range node.Children {
			if child.Node != nil {
				if err := index(child.Node); err != nil {
					return err
				}
			}
		}
		return nil
	}

	// Index each page.
	for _, page := range pages {
		traceLog(log.String("page", page.Tree.PathID))
		if err := index(page.Tree); err != nil {
			return err
		}
	}

	// Truncate the search index size if it exceeds our configured limit now.
	for _, suffix := range []string{"public", "private"} {
		if err := s.truncateDocumentationSearchIndexSize(ctx, suffix); err != nil {
			return errors.Wrap(err, "truncating documentation search index size")
		}
	}
	return nil
}

const purgeDocumentationSearchOldData = `
-- source: enterprise/internal/codeintel/stores/lsifstore/data_write_documentation.go:WriteDocumentationSearch
WITH
	langs AS (
		SELECT id FROM lsif_data_docs_search_lang_names_$SUFFIX
		WHERE tsv = %s
	),
	candidates AS (
		SELECT dump_id FROM lsif_data_docs_search_$SUFFIX
		WHERE repo_id=%s
		AND dump_root=%s

		-- Lock these rows in a deterministic order so that we don't deadlock with other processes
		-- updating the lsif_data_docs_search_* tables.
		ORDER BY dump_id FOR UPDATE
	)
DELETE FROM lsif_data_docs_search_$SUFFIX
WHERE dump_id = ANY(SELECT dump_id FROM candidates)
AND lang_name_id = ANY(SELECT id FROM langs)
RETURNING dump_id
`

const writeDocumentationSearchLangNames = `
WITH e AS(
	INSERT INTO lsif_data_docs_search_lang_names_$SUFFIX (lang_name, tsv)
	VALUES (%s, %s)
	ON CONFLICT DO NOTHING
	RETURNING id
)
SELECT * FROM e
UNION
	-- Fallback union for when there is a conflict in the upsert CTE. We only scan the first item.
    SELECT id FROM lsif_data_docs_search_lang_names_$SUFFIX WHERE tsv = %s
`

const writeDocumentationSearchRepoNames = `
WITH e AS(
	INSERT INTO lsif_data_docs_search_repo_names_$SUFFIX (repo_name, tsv, reverse_tsv)
	VALUES (%s, %s, %s)
	ON CONFLICT DO NOTHING
	RETURNING id
)
SELECT * FROM e
UNION
	-- Fallback union for when there is a conflict in the upsert CTE. We only scan the first item.
	SELECT id FROM lsif_data_docs_search_repo_names_$SUFFIX WHERE tsv = %s
`

const writeDocumentationSearchTags = `
WITH e AS(
	INSERT INTO lsif_data_docs_search_tags_$SUFFIX (tags, tsv)
	VALUES (%s, %s)
	ON CONFLICT DO NOTHING
	RETURNING id
)
SELECT * FROM e
UNION
	-- Fallback union for when there is a conflict in the upsert CTE. We only scan the first item.
	SELECT id FROM lsif_data_docs_search_tags_$SUFFIX WHERE tsv = %s
`

const writeDocumentationSearchInsertQuery = `
-- source: enterprise/internal/codeintel/stores/lsifstore/data_write_documentation.go:WriteDocumentationSearch
INSERT INTO lsif_data_docs_search_$SUFFIX (
	repo_id,
	dump_id,
	dump_root,
	path_id,
	detail,
	lang_name_id,
	repo_name_id,
	tags_id,
	search_key,
	search_key_tsv,
	search_key_reverse_tsv,
	label,
	label_tsv,
	label_reverse_tsv
) VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
`

var (
	lastTruncationWarningMu   sync.Mutex
	lastTruncationWarningTime time.Time
)

// truncateDocumentationSearchIndexSize is called (within a transaction) to truncate the
// documentation search index size according to the site config apidocs.search-index-limit-factor.
func (s *Store) truncateDocumentationSearchIndexSize(ctx context.Context, tableSuffix string) error {
	totalRows, exists, err := basestore.ScanFirstInt64(s.Query(ctx, sqlf.Sprintf(
		strings.ReplaceAll(countDocumentationSearchRowsQuery, "$SUFFIX", tableSuffix),
	)))
	if !exists {
		return fmt.Errorf("failed to count table size")
	}
	if err != nil {
		return errors.Wrap(err, "counting table size")
	}

	searchIndexLimitFactor := conf.Get().ApidocsSearchIndexSizeLimitFactor
	if searchIndexLimitFactor <= 0 {
		searchIndexLimitFactor = 1.0
	}
	searchIndexRowsLimit := int64(searchIndexLimitFactor * 250_000_000)
	rowsToDelete := totalRows - searchIndexRowsLimit
	if rowsToDelete <= 0 {
		return nil
	}

	lastTruncationWarningMu.Lock()
	if lastTruncationWarningTime.IsZero() || time.Since(lastTruncationWarningTime) > 5*time.Minute {
		lastTruncationWarningTime = time.Now()
		log15.Warn(
			"API docs search index size exceeded configured limit, truncating index",
			"apidocs.search-index-limit-factor", searchIndexLimitFactor,
			"rows_limit", searchIndexRowsLimit,
			"total_rows", totalRows,
			"deleting", rowsToDelete,
		)
	}
	lastTruncationWarningMu.Unlock()

	// Delete the first (oldest) N rows
	if err := s.Exec(ctx, sqlf.Sprintf(
		strings.ReplaceAll(truncateDocumentationSearchRowsQuery, "$SUFFIX", tableSuffix),
		rowsToDelete,
	)); err != nil {
		return errors.Wrap(err, "truncating search index rows")
	}
	return nil
}

// TODO(apidocs): future: introduce materialized count for this table and for other interesting API
// docs data points in general. https://github.com/sourcegraph/sourcegraph/pull/25206#discussion_r714270738
const countDocumentationSearchRowsQuery = `
-- source: enterprise/internal/codeintel/stores/lsifstore/data_write_documentation.go:truncateDocumentationSearchIndexSize
SELECT count(*)::bigint FROM lsif_data_docs_search_$SUFFIX
`

const truncateDocumentationSearchRowsQuery = `
-- source: enterprise/internal/codeintel/stores/lsifstore/data_write_documentation.go:truncateDocumentationSearchIndexSize
WITH candidates AS (
	SELECT id FROM lsif_data_docs_search_$SUFFIX

	-- Lock these rows in a deterministic order so that we don't deadlock with other processes
	-- updating the lsif_data_docs_search_* tables.
	ORDER BY id FOR UPDATE
	LIMIT %s
)
DELETE FROM lsif_data_docs_search_$SUFFIX
WHERE id IN (SELECT id FROM candidates)
RETURNING id
`

// truncate truncates a string to limitBytes, taking into account multi-byte runes. If the string
// is truncated, an ellipsis "…" is added to the end.
func truncate(s string, limitBytes int) string {
	runes := []rune(s)
	bytes := 0
	for i, r := range runes {
		if bytes+len(string(r)) >= limitBytes {
			return string(runes[:i-1]) + "…"
		}
		bytes += len(string(r))
	}
	return s
}

// textSearchVector constructs an ordered tsvector from the given string.
//
// Postgres' built in to_tsvector configurations (`simple`, `english`, etc.) work well for human
// language search but for multiple reasons produces tsvectors that are not appropriate for our
// use case of code search.
//
// By default, tsvectors perform word deduplication and normalization of words (Rat -> rat for
// example.) They also get sorted alphabetically:
//
// ```
// SELECT 'a fat cat sat on a mat and ate a fat rat'::tsvector;
//                       tsvector
// ----------------------------------------------------
//  'a' 'and' 'ate' 'cat' 'fat' 'mat' 'on' 'rat' 'sat'
// ```
//
// In the context of general document search, this doesn't matter. But in our context of API docs
// search, the order in which words (in the general computing sense) appear matters. For example,
// when searching `mux.router` it's important we match (package mux, symbol router) and not
// (package router, symbol mux).
//
// Another critical reason to_tsvector's configurations are not suitable for codes search is that
// they explicitly drop most punctuation (excluding periods) and don't split words between periods:
//
// ```
// select to_tsvector('foo::bar mux.Router const Foo* = Bar<T<X>>');
//                  to_tsvector
// ----------------------------------------------
//  'bar':2,6 'const':4 'foo':1,5 'mux.router':3
// ```
//
// Luckily, Postgres allows one to construct tsvectors manually by providing a sorted list of lexemes
// with optional integer _positions_ and weights, or:
//
// ```
// SELECT $$'foo':1 '::':2 'bar':3 'mux':4 'Router':5 'const':6 'Foo':7 '*':8 '=':9 'Bar':10 '<':11 'T':12 '<':13 'X':14 '>':15 '>':16$$::tsvector;
//                                                       tsvector
// --------------------------------------------------------------------------------------------------------------------
//  '*':8 '::':2 '<':11,13 '=':9 '>':15,16 'Bar':10 'Foo':7 'Router':5 'T':12 'X':14 'bar':3 'const':6 'foo':1 'mux':4
// ```
//
// Ordered, case-sensitive, punctuation-inclusive tsquery matches against the tsvector are then possible.
//
// For example, a query for `bufio.` would then match a tsvector ("bufio", ".", "Reader", ".", "writeBuf"):
//
// ```
// SELECT $$'bufio':1 '.':2 'Reader':3 '.':4 'writeBuf':5$$::tsvector @@ tsquery_phrase($$'bufio'$$::tsquery, $$'.'$$::tsquery) AS matches;
//  matches
// ---------
//  t
// ```
//
func textSearchVector(s string) string {
	// We need to emit a string in the Postgres tsvector format, roughly:
	//
	//     lexeme1:1 lexeme2:2 lexeme3:3
	//
	lexemes := lexemes(s)
	pairs := make([]string, 0, len(lexemes))
	for i, lexeme := range lexemes {
		pairs = append(pairs, fmt.Sprintf("%s:%v", lexeme, i+1))
	}
	return strings.Join(pairs, " ")
}

// lexemes splits the string into lexemes, each will be any contiguous section of Unicode digits,
// numbers, and letters. All other unicode runes, such as punctuation, are considered their own
// individual lexemes - and spaces are removed and considered boundaries.
func lexemes(s string) []string {
	var (
		lexemes       []string
		currentLexeme []rune
	)
	for _, r := range s {
		if unicode.IsDigit(r) || unicode.IsNumber(r) || unicode.IsLetter(r) {
			currentLexeme = append(currentLexeme, r)
			continue
		}
		if len(currentLexeme) > 0 {
			lexemes = append(lexemes, string(currentLexeme))
			currentLexeme = currentLexeme[:0]
		}
		if !unicode.IsSpace(r) {
			lexemes = append(lexemes, string(r))
		}
	}
	if len(currentLexeme) > 0 {
		lexemes = append(lexemes, string(currentLexeme))
	}
	return lexemes
}

// reverses a UTF-8 string accounting for Unicode and combining characters. This is not a part of
// the Go standard library or any of the golang.org/x packages. Note that reversing a slice of
// runes is not enough (would not handle combining characters.)
//
// See http://rosettacode.org/wiki/Reverse_a_string#Go
func reverse(s string) string {
	if s == "" {
		return ""
	}
	p := []rune(s)
	r := make([]rune, len(p))
	start := len(r)
	for i := 0; i < len(p); {
		// quietly skip invalid UTF-8
		if p[i] == utf8.RuneError {
			i++
			continue
		}
		j := i + 1
		for j < len(p) && (unicode.Is(unicode.Mn, p[j]) ||
			unicode.Is(unicode.Me, p[j]) || unicode.Is(unicode.Mc, p[j])) {
			j++
		}
		for k := j - 1; k >= i; k-- {
			start--
			r[start] = p[k]
		}
		i = j
	}
	return (string(r[start:]))
}
