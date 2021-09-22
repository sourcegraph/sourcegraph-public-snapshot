package lsifstore

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

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

	const (
		tableNamePublic  = "lsif_data_documentation_search_public"
		tableNamePrivate = "lsif_data_documentation_search_private"
	)
	tableName := tableNamePublic
	if repo.Private {
		tableName = tableNamePrivate
	}

	// This upload is for a commit on the default branch of the repository, so it is eligible for API
	// docs search indexing. It will replace any existing data that we have or this unique (repo_id, lang, root)
	// tuple in either table so we go ahead and purge the old data now.
	for _, tableName := range []string{tableNamePublic, tableNamePrivate} {
		if err := s.Exec(ctx, sqlf.Sprintf(
			strings.ReplaceAll(purgeDocumentationSearchOldData, "$TABLE_NAME", tableName),
			upload.RepositoryID,
			upload.Root,
			languageOrIndexerName,
		)); err != nil {
			return errors.Wrap(err, "purging old data")
		}
	}

	var index func(node *precise.DocumentationNode) error
	index = func(node *precise.DocumentationNode) error {
		if node.Documentation.SearchKey != "" {
			tags := []string{}
			for _, tag := range node.Documentation.Tags {
				tags = append(tags, string(tag))
			}
			err := s.Exec(ctx, sqlf.Sprintf(
				strings.ReplaceAll(writeDocumentationSearchInsertQuery, "$TABLE_NAME", tableName),
				upload.ID,
				node.PathID,
				languageOrIndexerName,
				upload.RepositoryName,
				node.Documentation.SearchKey,
				truncate(node.Label.String(), 256),     // 256 bytes, enough for ~100 characters in all languages
				truncate(node.Detail.String(), 5*1024), // 5 KiB - just for sanity
				strings.Join(tags, " "),
				upload.RepositoryID,
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
	for _, tableName := range []string{tableNamePublic, tableNamePrivate} {
		if err := s.truncateDocumentationSearchIndexSize(ctx, tableName); err != nil {
			return errors.Wrap(err, "truncating documentation search index size")
		}
	}
	return nil
}

const purgeDocumentationSearchOldData = `
-- source: enterprise/internal/codeintel/stores/lsifstore/data_write_documentation.go:WriteDocumentationSearch
WITH candidates AS (
	SELECT id FROM lsif_dumps
	WHERE repository_id=%s
	AND root=%s

	-- Lock these rows in a deterministic order so that we don't deadlock with other processes
	-- updating the lsif_data_documentation_search_* tables.
	ORDER BY id FOR UPDATE
)
DELETE FROM $TABLE_NAME
WHERE dump_id IN (SELECT dump_id FROM candidates)
AND lang=%s
RETURNING dump_id
`

const writeDocumentationSearchInsertQuery = `
-- source: enterprise/internal/codeintel/stores/lsifstore/data_write_documentation.go:WriteDocumentationSearch
INSERT INTO $TABLE_NAME (dump_id, path_id, lang, repo_name, search_key, label, detail, tags, repo_id)
VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s)
`

var (
	lastTruncationWarningMu   sync.Mutex
	lastTruncationWarningTime time.Time
)

// truncateDocumentationSearchIndexSize is called (within a transaction) to truncate the
// documentation search index size according to the site config apidocs.search-index-limit-factor.
func (s *Store) truncateDocumentationSearchIndexSize(ctx context.Context, tableName string) error {
	totalRows, exists, err := basestore.ScanFirstInt64(s.Query(ctx, sqlf.Sprintf(
		strings.ReplaceAll(countDocumentationSearchRowsQuery, "$TABLE_NAME", tableName),
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
		strings.ReplaceAll(truncateDocumentationSearchRowsQuery, "$TABLE_NAME", tableName),
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
SELECT count(*)::bigint FROM $TABLE_NAME
`

const truncateDocumentationSearchRowsQuery = `
-- source: enterprise/internal/codeintel/stores/lsifstore/data_write_documentation.go:truncateDocumentationSearchIndexSize
WITH candidates AS (
	SELECT ctid FROM $TABLE_NAME

	-- Lock these rows in a deterministic order so that we don't deadlock with other processes
	-- updating the lsif_data_documentation_search_* tables.
	ORDER BY dump_id FOR UPDATE
	LIMIT %s
)
DELETE FROM $TABLE_NAME
WHERE ctid IN (SELECT ctid FROM candidates)
RETURNING ctid
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
