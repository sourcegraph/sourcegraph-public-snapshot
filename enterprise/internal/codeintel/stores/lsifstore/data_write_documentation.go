package lsifstore

import (
	"context"
	"database/sql"
	"fmt"
	stdlog "log"
	"runtime/debug"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/keegancsmith/sqlf"
	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/lsifstore/apidocs"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/batch"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/lsif/protocol"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// WriteDocumentationPages is called (transactionally) from the precise-code-intel-worker.
//
// The repository name and language name identifiers should be created from a previous invocation of the
// WriteDocumentationSearchPrework method with the same parameters.
func (s *Store) WriteDocumentationPages(
	ctx context.Context,
	upload dbstore.Upload,
	repo *types.Repo,
	isDefaultBranch bool,
	documentationPages chan *precise.DocumentationPageData,
	repositoryNameID int,
	languageNameID int,
) (count uint32, err error) {
	ctx, trace, endObservation := s.operations.writeDocumentationPages.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("bundleID", upload.ID),
		log.String("repo", upload.RepositoryName),
		log.String("commit", upload.Commit),
		log.String("root", upload.Root),
	}})
	defer endObservation(1, observation.Args{})

	defer func() {
		if err := recover(); err != nil {
			stack := debug.Stack()
			stdlog.Printf("API docs panic: %v\n%s", err, stack)
			trace.Log(log.String("API docs panic error", fmt.Sprint(err)))
			trace.Log(log.String("API docs panic stack", string(stack)))
		}
	}()

	// Create temporary table symmetric to lsif_data_documentation_pages without the dump id
	if err := s.Exec(ctx, sqlf.Sprintf(writeDocumentationPagesTemporaryTableQuery)); err != nil {
		return 0, err
	}

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
		s.Handle().DB(),
		"t_lsif_data_documentation_pages",
		[]string{"path_id", "data"},
		inserter,
	); err != nil {
		return 0, err
	}
	trace.Log(log.Int("numResultChunkRecords", int(count)))

	// Note: If someone disables API docs search indexing, uploads during that time will not be
	// indexed even if it is turned back on. Only future uploads would be.
	if conf.APIDocsSearchIndexingEnabled() {
		// Perform search indexing for API docs pages.
		if err := s.WriteDocumentationSearch(ctx, upload, repo, isDefaultBranch, pages, repositoryNameID, languageNameID); err != nil {
			return 0, errors.Wrap(err, "WriteDocumentationSearch")
		}
	}

	// Insert the values from the temporary table into the target table. We select a
	// parameterized dump id here since it is the same for all rows in this operation.
	return count, s.Exec(ctx, sqlf.Sprintf(writeDocumentationPagesInsertQuery, upload.ID))
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
func (s *Store) WriteDocumentationPathInfo(ctx context.Context, bundleID int, documentationPathInfo chan *precise.DocumentationPathInfoData) (count uint32, err error) {
	ctx, trace, endObservation := s.operations.writeDocumentationPathInfo.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("bundleID", bundleID),
	}})
	defer endObservation(1, observation.Args{})

	tx, err := s.Transact(ctx)
	if err != nil {
		return 0, err
	}
	defer func() { err = tx.Done(err) }()

	// Create temporary table symmetric to lsif_data_documentation_path_info without the dump id
	if err := tx.Exec(ctx, sqlf.Sprintf(writeDocumentationPathInfoTemporaryTableQuery)); err != nil {
		return 0, err
	}

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
		return 0, err
	}
	trace.Log(log.Int("numResultChunkRecords", int(count)))

	// Insert the values from the temporary table into the target table. We select a
	// parameterized dump id here since it is the same for all rows in this operation.
	return count, tx.Exec(ctx, sqlf.Sprintf(writeDocumentationPathInfoInsertQuery, bundleID))
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
func (s *Store) WriteDocumentationMappings(ctx context.Context, bundleID int, mappings chan precise.DocumentationMapping) (count uint32, err error) {
	ctx, trace, endObservation := s.operations.writeDocumentationMappings.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("bundleID", bundleID),
	}})
	defer endObservation(1, observation.Args{})

	tx, err := s.Transact(ctx)
	if err != nil {
		return 0, err
	}
	defer func() { err = tx.Done(err) }()

	// Create temporary table symmetric to lsif_data_documentation_mappings without the dump id
	if err := tx.Exec(ctx, sqlf.Sprintf(writeDocumentationMappingsTemporaryTableQuery)); err != nil {
		return 0, err
	}

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
		return 0, err
	}
	trace.Log(log.Int("numRecords", int(count)))

	// Insert the values from the temporary table into the target table. We select a
	// parameterized dump id here since it is the same for all rows in this operation.
	return count, tx.Exec(ctx, sqlf.Sprintf(writeDocumentationMappingsInsertQuery, bundleID))
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

// WriteDocumentationSearchPrework upserts the repository name and language name identifiers into the
// appropriate tables. These values should be passed to a later invocation of WriteDocumentationSearch.
//
// Since these values are interned and heavily shared, we recommended upserting both of these values
// outside of a long-running transaction to reduce lock contention between shared rows being held longer
// than necessary.
func (s *Store) WriteDocumentationSearchPrework(ctx context.Context, upload dbstore.Upload, repo *types.Repo, isDefaultBranch bool) (_ int, _ int, err error) {
	ctx, _, endObservation := s.operations.writeDocumentationSearchPrework.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("repo", upload.RepositoryName),
		log.Int("bundleID", upload.ID),
	}})
	defer endObservation(1, observation.Args{})

	if !isDefaultBranch {
		// We do not index non-default branches for API docs search.
		return 0, 0, nil
	}

	if !conf.APIDocsSearchIndexingEnabled() {
		// We will not use these values within WriteDocumentationPages
		return 0, 0, nil
	}

	tableSuffix := "public"
	if repo.Private {
		tableSuffix = "private"
	}

	repositoryNameID, err := s.upsertRepositoryName(ctx, upload.RepositoryName, tableSuffix)
	if err != nil {
		return 0, 0, errors.Wrap(err, "upsertRepositoryName")
	}

	languageNameID, err := s.upsertLanguageName(ctx, upload.Indexer, tableSuffix)
	if err != nil {
		return 0, 0, errors.Wrap(err, "upsertLanguageName")
	}

	return repositoryNameID, languageNameID, nil
}

// WriteDocumentationSearch is called to write the search index for a given documentation page. This method
// is called from transactionally from the precise-code-intel worker as well as from the apiDocsSearchMigrator.
//
// The repository name and language name identifiers should be created from a previous invocation of the
// WriteDocumentationSearchPrework method with the same parameters.
func (s *Store) WriteDocumentationSearch(
	ctx context.Context,
	upload dbstore.Upload,
	repo *types.Repo,
	isDefaultBranch bool,
	pages []*precise.DocumentationPageData,
	repositoryNameID int,
	languageNameID int,
) (err error) {
	ctx, _, endObservation := s.operations.writeDocumentationSearch.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("repo", upload.RepositoryName),
		log.Int("bundleID", upload.ID),
		log.Int("pages", len(pages)),
	}})
	defer endObservation(1, observation.Args{})

	if !isDefaultBranch {
		// We do not index non-default branches for API docs search.
		return nil
	}

	tableSuffix := "public"
	if repo.Private {
		tableSuffix = "private"
	}

	tx, err := s.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	// gatherTags sorts the tags we'll be dealing with, This ensures that we will always
	// try to bulk-insert the tags in the same order, which should avoid deadlock situations
	// where overlapping tags are updated in different orders from different processors.
	tagIDs, err := tx.upsertTags(ctx, gatherTags(pages), tableSuffix)
	if err != nil {
		return errors.Wrap(err, "upsertTags")
	}

	if err := tx.replaceSearchRecords(ctx, upload, repositoryNameID, languageNameID, pages, tagIDs, tableSuffix); err != nil {
		return errors.Wrap(err, "replaceSearchRecords")
	}

	// Truncate the search index size if it exceeds our configured limit now.
	if err := tx.truncateDocumentationSearchIndexSize(ctx, tableSuffix); err != nil {
		return errors.Wrap(err, "truncateDocumentationSearchIndexSize")
	}

	return nil
}

func (s *Store) upsertRepositoryName(ctx context.Context, name, tableSuffix string) (int, error) {
	id, _, err := basestore.ScanFirstInt(s.Query(ctx, sqlf.Sprintf(
		strings.ReplaceAll(upsertRepositoryNameQuery, "$SUFFIX", tableSuffix),
		name,
		apidocs.TextSearchVector(name),
		apidocs.TextSearchVector(apidocs.Reverse(name)),
		name,
	)))

	return id, err
}

const upsertRepositoryNameQuery = `
-- source: enterprise/internal/codeintel/stores/lsifstore/data_write_documentation.go:upsertRepositoryName
WITH inserted AS (
	INSERT INTO lsif_data_docs_search_repo_names_$SUFFIX (repo_name, tsv, reverse_tsv)
	VALUES (%s, %s, %s)
	ON CONFLICT DO NOTHING
	RETURNING id
)
SELECT id FROM inserted
UNION
SELECT id FROM lsif_data_docs_search_repo_names_$SUFFIX WHERE repo_name = %s
LIMIT 1
`

func (s *Store) upsertLanguageName(ctx context.Context, indexerName, tableSuffix string) (int, error) {
	// This will not always produce a proper language name, e.g. if an indexer is not named after
	// the language or is not in "lsif-$LANGUAGE" format. That's OK: in that case, the "language"
	// is the indexer name which is likely good enough since we use fuzzy search / partial text
	// matching over it.
	languageName := strings.ToLower(strings.TrimPrefix(indexerName, "lsif-"))

	id, _, err := basestore.ScanFirstInt(s.Query(ctx, sqlf.Sprintf(
		strings.ReplaceAll(upsertLanguageNameQuery, "$SUFFIX", tableSuffix),
		languageName,
		apidocs.TextSearchVector(languageName),
		languageName,
	)))

	return id, err
}

const upsertLanguageNameQuery = `
-- source: enterprise/internal/codeintel/stores/lsifstore/data_write_documentation.go:upsertLanguageName
WITH inserted AS (
	INSERT INTO lsif_data_docs_search_lang_names_$SUFFIX (lang_name, tsv)
	VALUES (%s, %s)
	ON CONFLICT DO NOTHING
	RETURNING id
)
SELECT id FROM inserted
UNION
SELECT id FROM lsif_data_docs_search_lang_names_$SUFFIX WHERE lang_name = %s
LIMIT 1
`

func (s *Store) upsertTags(ctx context.Context, tags []string, tableSuffix string) (map[string]int, error) {
	tx, err := s.Transact(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { err = tx.Done(err) }()

	// Create temporary table symmetric to lsif_data_docs_search_tags_$SUFFIX. We'll use this
	// as a staging table so that we can insert only the values that don't already exist in the
	// target table..
	if err := tx.Exec(ctx, sqlf.Sprintf(strings.ReplaceAll(upsertTagsTemporaryTableQuery, "$SUFFIX", tableSuffix))); err != nil {
		return nil, errors.Wrap(err, "creating temporary table")
	}

	inserter := func(inserter *batch.Inserter) error {
		for _, tags := range tags {
			if err := inserter.Insert(ctx, tags, apidocs.TextSearchVector(tags)); err != nil {
				return err
			}
		}

		return nil
	}

	// Bulk insert tag values into the temporary table
	if err := batch.WithInserter(
		ctx,
		tx.Handle().DB(),
		"t_lsif_data_docs_search_tags_"+tableSuffix,
		batch.MaxNumPostgresParameters,
		[]string{"tags", "tsv"},
		inserter,
	); err != nil {
		return nil, errors.Wrap(err, "bulk inserting tags")
	}

	// Upsert the values from the temporary table into the target table. Here we insert
	// only values that are not yet present. This query also selects the ids of each of
	// the pre-existing tags so that we don't need a second fetch.
	tagIDs, err := scanStringIntPairs(tx.Query(ctx, sqlf.Sprintf(strings.ReplaceAll(upsertTagsInsertQuery, "$SUFFIX", tableSuffix))))
	if err != nil {
		return nil, errors.Wrap(err, "committing staged tags")
	}

	return tagIDs, nil
}

const upsertTagsTemporaryTableQuery = `
-- source: enterprise/internal/codeintel/stores/lsifstore/data_write_documentation.go:upsertTags
CREATE TEMPORARY TABLE t_lsif_data_docs_search_tags_$SUFFIX (
	tags TEXT NOT NULL,
	tsv TSVECTOR NOT NULL
) ON COMMIT DROP
`

const upsertTagsInsertQuery = `
-- source: enterprise/internal/codeintel/stores/lsifstore/data_write_documentation.go:upsertTags
WITH
inserted AS (
	INSERT INTO lsif_data_docs_search_tags_$SUFFIX (tags, tsv)
	SELECT source.tags, source.tsv
	FROM t_lsif_data_docs_search_tags_$SUFFIX source
	WHERE source.tags NOT IN (SELECT tags FROM lsif_data_docs_search_tags_$SUFFIX)
	RETURNING id, tags
),
existing AS (
	SELECT source.id, source.tags
	FROM lsif_data_docs_search_tags_$SUFFIX source
	WHERE source.tags IN (SELECT tags FROM t_lsif_data_docs_search_tags_$SUFFIX)
)
SELECT tags, id FROM inserted
UNION
SELECT tags, id FROM existing
`

func scanStringIntPairs(rows *sql.Rows, queryErr error) (_ map[string]int, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	values := map[string]int{}
	for rows.Next() {
		var value1 string
		var value2 int
		if err := rows.Scan(&value1, &value2); err != nil {
			return nil, err
		}

		values[value1] = value2
	}

	return values, nil
}

func (s *Store) replaceSearchRecords(
	ctx context.Context,
	upload dbstore.Upload,
	repositoryNameID int,
	languageNameID int,
	pages []*precise.DocumentationPageData,
	tagIDs map[string]int,
	tableSuffix string,
) error {
	tx, err := s.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	// Create temporary table symmetric to lsif_data_docs_search_$SUFFIX without the fields that would have
	// the same value for the same upload. We'll insert these shared values all at once to save on query
	// bandwidth.
	if err := tx.Exec(ctx, sqlf.Sprintf(strings.ReplaceAll(insertSearchRecordsTemporaryTableQuery, "$SUFFIX", tableSuffix))); err != nil {
		return errors.Wrap(err, "creating temporary table")
	}

	inserter := func(inserter *batch.Inserter) error {
		handler := func(node *precise.DocumentationNode) error {
			if node.Documentation.SearchKey == "" {
				return nil
			}

			detail := apidocs.Truncate(node.Detail.String(), 5*1024) // 5 KiB - just for sanity
			label := apidocs.Truncate(node.Label.String(), 256)      // 256 bytes, enough for ~100 characters in all languages
			tagsID := tagIDs[normalizeTags(node.Documentation.Tags)]

			if err := inserter.Insert(
				ctx,
				node.PathID, // path_id
				detail,      // detail
				tagsID,      // tags_id

				node.Documentation.SearchKey,                                            // search_key
				apidocs.TextSearchVector(node.Documentation.SearchKey),                  // search_key_tsv
				apidocs.TextSearchVector(apidocs.Reverse(node.Documentation.SearchKey)), // search_key_reverse_tsv

				label,                           // label
				apidocs.TextSearchVector(label), // label_tsv
				apidocs.TextSearchVector(apidocs.Reverse(label)), // label_reverse_tsv
			); err != nil {
				return err
			}

			return nil
		}

		for _, page := range pages {
			if err := walkDocumentationNode(page.Tree, handler); err != nil {
				return err
			}
		}

		return nil
	}

	// Bulk insert all the unique column values into the temporary table
	if err := batch.WithInserter(
		ctx,
		tx.Handle().DB(),
		"t_lsif_data_docs_search_"+tableSuffix,
		batch.MaxNumPostgresParameters,
		[]string{
			"path_id",
			"detail",
			"tags_id",
			"search_key",
			"search_key_tsv",
			"search_key_reverse_tsv",
			"label",
			"label_tsv",
			"label_reverse_tsv",
		},
		inserter,
	); err != nil {
		return errors.Wrap(err, "bulk inserting search records")
	}

	// Insert the values from the temporary table into the target table. Here we insert
	// the values that are the same for every row instead of sending them on each of the
	// batched insert calls.
	if err := tx.Exec(ctx, sqlf.Sprintf(
		strings.ReplaceAll(insertSearchRecordsInsertQuery, "$SUFFIX", tableSuffix),
		upload.RepositoryID, // repo_id
		upload.ID,           // dump_id
		upload.Root,         // dump_root
		repositoryNameID,    // repo_name_id
		languageNameID,      // lang_name_id
	)); err != nil {
		return errors.Wrap(err, "committing staged search records")
	}

	// Insert a current marker for the recently inserted search records
	if err := tx.Exec(ctx, sqlf.Sprintf(
		strings.ReplaceAll(insertSearchRecordsInsertCurrentMarkerQuery, "$SUFFIX", tableSuffix),
		upload.RepositoryID, // repo_id
		upload.Root,         // dump_root
		languageNameID,      // lang_name_id
		upload.ID,           // dump_id
	)); err != nil {
		return errors.Wrap(err, "inserting current marker")
	}

	return nil
}

const insertSearchRecordsTemporaryTableQuery = `
-- source: enterprise/internal/codeintel/stores/lsifstore/data_write_documentation.go:insertSearchRecords
CREATE TEMPORARY TABLE t_lsif_data_docs_search_$SUFFIX (
	path_id TEXT NOT NULL,
	detail TEXT NOT NULL,
	tags_id INTEGER NOT NULL,
	search_key TEXT NOT NULL,
	search_key_tsv TSVECTOR NOT NULL,
	search_key_reverse_tsv TSVECTOR NOT NULL,
	label TEXT NOT NULL,
	label_tsv TSVECTOR NOT NULL,
	label_reverse_tsv TSVECTOR NOT NULL
) ON COMMIT DROP
`

const insertSearchRecordsInsertQuery = `
-- source: enterprise/internal/codeintel/stores/lsifstore/data_write_documentation.go:insertSearchRecords
INSERT INTO lsif_data_docs_search_$SUFFIX (
	repo_id,
	dump_id,
	dump_root,
	repo_name_id,
	lang_name_id,
	path_id,
	detail,
	tags_id,
	search_key,
	search_key_tsv,
	search_key_reverse_tsv,
	label,
	label_tsv,
	label_reverse_tsv
)
SELECT
	%s, -- repo_id
	%s, -- dump_id
	%s, -- dump_root
	%s, -- repo_name_id
	%s, -- lang_name_id
	source.path_id,
	source.detail,
	source.tags_id,
	source.search_key,
	source.search_key_tsv,
	source.search_key_reverse_tsv,
	source.label,
	source.label_tsv,
	source.label_reverse_tsv
FROM t_lsif_data_docs_search_$SUFFIX source
`

const insertSearchRecordsInsertCurrentMarkerQuery = `
-- source: enterprise/internal/codeintel/stores/lsifstore/data_write_documentation.go:insertSearchRecords
INSERT INTO lsif_data_docs_search_current_$SUFFIX (
	repo_id,
	dump_root,
	lang_name_id,
	dump_id
)
VALUES (%s, %s, %s, %s)
`

func walkDocumentationNode(node *precise.DocumentationNode, f func(node *precise.DocumentationNode) error) error {
	if err := f(node); err != nil {
		return err
	}

	for _, child := range node.Children {
		if child.Node != nil {
			if err := walkDocumentationNode(child.Node, f); err != nil {
				return err
			}
		}
	}

	return nil
}

func gatherTags(pages []*precise.DocumentationPageData) []string {
	tagMap := map[string]struct{}{}
	for _, page := range pages {
		_ = walkDocumentationNode(page.Tree, func(node *precise.DocumentationNode) error {
			if node.Documentation.SearchKey != "" {
				tagMap[normalizeTags(node.Documentation.Tags)] = struct{}{}
			}

			return nil
		})
	}

	tags := make([]string, 0, len(tagMap))
	for normalizedTags := range tagMap {
		tags = append(tags, normalizedTags)
	}
	sort.Strings(tags)

	return tags
}

func normalizeTags(tags []protocol.Tag) string {
	s := make([]string, 0, len(tags))
	for _, tag := range tags {
		s = append(s, string(tag))
	}

	return strings.Join(s, " ")
}

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
	if err != nil {
		return errors.Wrap(err, "counting table size")
	}
	if !exists {
		return errors.Newf("failed to count table size")
	}

	searchIndexLimitFactor := s.config.SiteConfig().ApidocsSearchIndexSizeLimitFactor
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

const countDocumentationSearchRowsQuery = `
-- source: enterprise/internal/codeintel/stores/lsifstore/data_write_documentation.go:truncateDocumentationSearchIndexSize
SELECT count::bigint FROM lsif_data_apidocs_num_search_results_$SUFFIX
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
