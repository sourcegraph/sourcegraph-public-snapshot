package lsifstore

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
)

// DocumentationPage returns the documentation page with the given PathID.
func (s *Store) DocumentationPage(ctx context.Context, bundleID int, pathID string) (_ *precise.DocumentationPageData, err error) {
	ctx, _, endObservation := s.operations.documentationPage.WithAndLogger(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("bundleID", bundleID),
		log.String("pathID", pathID),
	}})
	defer endObservation(1, observation.Args{})

	page, err := s.scanFirstDocumentationPageData(s.Store.Query(ctx, sqlf.Sprintf(documentationPageDataQuery, bundleID, pathID)))
	if err != nil {
		return nil, err
	}
	return page, nil
}

const documentationPageDataQuery = `
-- source: enterprise/internal/codeintel/stores/lsifstore/documentation.go:DocumentationPage
SELECT
	dump_id,
	path_id,
	data
FROM
	lsif_data_documentation_pages
WHERE
	dump_id = %s AND
	path_id = %s
`

// scanFirstDocumentationPageData reads the first DocumentationPageData row. If no rows match the
// query, a nil is returned.
func (s *Store) scanFirstDocumentationPageData(rows *sql.Rows, queryErr error) (_ *precise.DocumentationPageData, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	if !rows.Next() {
		return nil, nil
	}

	var (
		rawData  []byte
		uploadID int
		pathID   string
	)
	if err := rows.Scan(
		&uploadID,
		&pathID,
		&rawData,
	); err != nil {
		return nil, err
	}
	record, err := s.serializer.UnmarshalDocumentationPageData(rawData)
	if err != nil {
		return nil, err
	}
	return record, nil
}

// DocumentationPathInfo returns info describing what is at the given pathID.
func (s *Store) DocumentationPathInfo(ctx context.Context, bundleID int, pathID string) (_ *precise.DocumentationPathInfoData, err error) {
	ctx, _, endObservation := s.operations.documentationPathInfo.WithAndLogger(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("bundleID", bundleID),
		log.String("pathID", pathID),
	}})
	defer endObservation(1, observation.Args{})

	page, err := s.scanFirstDocumentationPathInfoData(s.Store.Query(ctx, sqlf.Sprintf(documentationPathInfoDataQuery, bundleID, pathID)))
	if err != nil {
		return nil, err
	}
	return page, nil
}

const documentationPathInfoDataQuery = `
-- source: enterprise/internal/codeintel/stores/lsifstore/documentation.go:DocumentationPathInfo
SELECT
	dump_id,
	path_id,
	data
FROM
	lsif_data_documentation_path_info
WHERE
	dump_id = %s AND
	path_id = %s
`

// scanFirstDocumentationPathInfoData reads the first DocumentationPathInfoData row. If no rows match the
// query, a nil is returned.
func (s *Store) scanFirstDocumentationPathInfoData(rows *sql.Rows, queryErr error) (_ *precise.DocumentationPathInfoData, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	if !rows.Next() {
		return nil, nil
	}

	var (
		rawData  []byte
		uploadID int
		pathID   string
	)
	if err := rows.Scan(
		&uploadID,
		&pathID,
		&rawData,
	); err != nil {
		return nil, err
	}
	record, err := s.serializer.UnmarshalDocumentationPathInfoData(rawData)
	if err != nil {
		return nil, err
	}
	return record, nil
}

// documentationIDsToPathIDs returns a mapping of the given documentationResult IDs to their
// associative path IDs. Empty result IDs ("") are ignored.
func (s *Store) documentationIDsToPathIDs(ctx context.Context, bundleID int, ids []precise.ID) (_ map[precise.ID]string, err error) {
	ctx, _, endObservation := s.operations.documentationIDsToPathIDs.WithAndLogger(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("bundleID", bundleID),
		log.String("ids", fmt.Sprint(ids)),
	}})
	defer endObservation(1, observation.Args{})

	var wantIDs []uint64
	for _, id := range ids {
		if id != "" {
			idInt, err := strconv.ParseUint(string(id), 10, 64)
			if err != nil {
				return nil, err
			}
			wantIDs = append(wantIDs, idInt)
		}
	}
	if len(wantIDs) == 0 {
		return nil, nil
	}

	pathIDs, err := s.scanDocumentationPathIDs(s.Store.Query(ctx, sqlf.Sprintf(documentationIDsToPathIDsQuery, bundleID, pq.Array(wantIDs))))
	if err != nil {
		return nil, err
	}
	return pathIDs, nil
}

// scanDocumentationPathIDs reads documentation path IDs from the given row object.
func (s *Store) scanDocumentationPathIDs(rows *sql.Rows, queryErr error) (_ map[precise.ID]string, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	values := map[precise.ID]string{}
	for rows.Next() {
		var (
			resultID uint64
			pathID   string
		)
		if err := rows.Scan(&resultID, &pathID); err != nil {
			return nil, err
		}
		values[precise.ID(strconv.FormatUint(resultID, 10))] = pathID
	}
	return values, nil
}

const documentationIDsToPathIDsQuery = `
-- source: enterprise/internal/codeintel/stores/lsifstore/ranges.go:documentationIDsToPathIDs
SELECT
	result_id,
	path_id
FROM
	lsif_data_documentation_mappings
WHERE
	dump_id = %s AND
	result_id = ANY (%s)
`

func (s *Store) documentationPathIDToID(ctx context.Context, bundleID int, pathID string) (_ precise.ID, err error) {
	ctx, _, endObservation := s.operations.documentationPathIDToID.WithAndLogger(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("bundleID", bundleID),
		log.String("pathID", pathID),
	}})
	defer endObservation(1, observation.Args{})

	resultID, err := s.scanFirstDocumentationResultID(s.Store.Query(ctx, sqlf.Sprintf(documentationPathIDToIDQuery, bundleID, pathID)))
	if err != nil {
		return "", err
	}
	if resultID == -1 {
		return "", err
	}
	return precise.ID(strconv.FormatInt(resultID, 10)), nil
}

const documentationPathIDToIDQuery = `
-- source: enterprise/internal/codeintel/stores/lsifstore/ranges.go:documentationPathIDToID
SELECT
	result_id
FROM
	lsif_data_documentation_mappings
WHERE
	dump_id = %s AND
	path_id = %s
LIMIT 1
`

// scanFirstDocumentationResultID reads the first result_id row. If no rows match the query, an empty string is returned.
func (s *Store) scanFirstDocumentationResultID(rows *sql.Rows, queryErr error) (_ int64, err error) {
	if queryErr != nil {
		return -1, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	if !rows.Next() {
		return -1, nil
	}

	var resultID int64
	if err := rows.Scan(&resultID); err != nil {
		return -1, err
	}
	return resultID, nil
}

// documentationPathIDToFilePath queries the file path associated with a documentation path ID,
// e.g. the file where the documented symbol is located - if the path ID is describing such a
// symbol, or nil otherwise.
func (s *Store) documentationPathIDToFilePath(ctx context.Context, bundleID int, pathID string) (_ *string, err error) {
	ctx, _, endObservation := s.operations.documentationPathIDToFilePath.WithAndLogger(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("bundleID", bundleID),
		log.String("pathID", pathID),
	}})
	defer endObservation(1, observation.Args{})

	return s.scanFirstDocumentationFilePath(s.Store.Query(ctx, sqlf.Sprintf(documentationPathIDToFilePathQuery, bundleID, pathID)))
}

const documentationPathIDToFilePathQuery = `
-- source: enterprise/internal/codeintel/stores/lsifstore/ranges.go:documentationPathIDToFilePath
SELECT
	file_path
FROM
	lsif_data_documentation_mappings
WHERE
	dump_id = %s AND
	path_id = %s
LIMIT 1
`

// scanFirstDocumentationFilePath reads the first file_path row. If no rows match the query, an empty string is returned.
func (s *Store) scanFirstDocumentationFilePath(rows *sql.Rows, queryErr error) (_ *string, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	if !rows.Next() {
		return nil, nil
	}

	var filePath *string
	if err := rows.Scan(&filePath); err != nil {
		return nil, err
	}
	return filePath, nil
}

// DocumentationDefinitions returns the set of locations defining the symbol found at the given path ID, if any.
func (s *Store) DocumentationDefinitions(ctx context.Context, bundleID int, pathID string, limit, offset int) (_ []Location, _ int, err error) {
	resultID, err := s.documentationPathIDToID(ctx, bundleID, pathID)
	if err != nil || resultID == "" {
		return nil, 0, err
	}
	extractor := func(r precise.RangeData) precise.ID { return r.DefinitionResultID }
	operation := s.operations.documentationDefinitions
	return s.documentationDefinitions(ctx, extractor, operation, bundleID, pathID, resultID, limit, offset)
}

func (s *Store) documentationDefinitions(
	ctx context.Context,
	extractor func(r precise.RangeData) precise.ID,
	operation *observation.Operation,
	bundleID int,
	pathID string,
	resultID precise.ID,
	limit,
	offset int,
) (_ []Location, _ int, err error) {
	ctx, traceLog, endObservation := operation.WithAndLogger(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("bundleID", bundleID),
		log.String("resultID", string(resultID)),
	}})
	defer endObservation(1, observation.Args{})

	filePath, err := s.documentationPathIDToFilePath(ctx, bundleID, pathID)
	if err != nil {
		return nil, 0, errors.Wrap(err, "documentationPathIDToFilePath")
	}
	if filePath == nil {
		// The documentation result is not attached to a file, it cannot have references.
		return nil, 0, nil
	}

	documentData, exists, err := s.scanFirstDocumentData(s.Store.Query(ctx, sqlf.Sprintf(locationsDocumentQuery, bundleID, filePath)))
	if err != nil || !exists {
		return nil, 0, err
	}

	traceLog(log.Int("numRanges", len(documentData.Document.Ranges)))
	var found *precise.RangeData
	for _, rn := range documentData.Document.Ranges {
		if rn.DocumentationResultID == resultID {
			found = &rn
			break
		}
	}
	traceLog(log.Bool("found", found == nil))
	if found == nil {
		return nil, 0, errors.New("not found")
	}

	orderedResultIDs := extractResultIDs([]precise.RangeData{*found}, extractor)
	locationsMap, totalCount, err := s.locations(ctx, bundleID, orderedResultIDs, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	traceLog(log.Int("totalCount", totalCount))

	locations := make([]Location, 0, limit)
	for _, resultID := range orderedResultIDs {
		locations = append(locations, locationsMap[resultID]...)
	}

	return locations, totalCount, nil
}

// DocumentationSearch searches API documentation in either the "public" or "private" table.
func (s *Store) DocumentationSearch(ctx context.Context, table, query string, repos []string) (_ []precise.DocumentationSearchResult, err error) {
	ctx, _, endObservation := s.operations.documentationSearch.WithAndLogger(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("query", query),
		log.String("repos", fmt.Sprint(repos)),
	}})
	defer endObservation(1, observation.Args{})

	metaQuery := query
	mainQuery := query
	if i := strings.Index(query, ":"); i != -1 {
		metaQuery = strings.TrimSpace(query[:i])
		mainQuery = strings.TrimSpace(query[i+len(":"):])
	}
	fmt.Printf("QUERY META %q MAIN %q\n", metaQuery, mainQuery)
	return s.scanDocumentationSearchResults(s.Store.Query(ctx, sqlf.Sprintf(
		strings.ReplaceAll(documentationSearchQuery, "$TABLE_NAME", "lsif_data_documentation_search_"+table),
		metaQuery,
		metaQuery,
		metaQuery,
		metaQuery,
		metaQuery,
		metaQuery,
		mainQuery,
		mainQuery,
		mainQuery,
		mainQuery,
	)))
}

const documentationSearchQuery = `
-- source: enterprise/internal/codeintel/stores/lsifstore/documentation.go:DocumentationSearch
WITH
    matching_langs AS (
        SELECT DISTINCT
            lang,
            lang <<<-> %s AS lang_dist
        FROM $TABLE_NAME
		WHERE lang <<%% %s
        ORDER BY lang_dist
        LIMIT 1
    ),
    matching_repos AS (
        SELECT DISTINCT
            repo_name,
            repo_name <<-> %s as repo_name_dist
        FROM $TABLE_NAME
        WHERE repo_name <<%% %s
        ORDER BY repo_name_dist
        LIMIT 100
    ),
    matching_tags AS (
        SELECT DISTINCT
            tags,
            tags <<-> %s as tags_dist
        FROM $TABLE_NAME
        WHERE tags <<%% %s
        ORDER BY tags_dist
        LIMIT 100
    )
SELECT
	lang,
    repo_name,
    search_key,
    search_key_dist,
	label,
	label_dist,
	detail,
    tags,
	path_id
FROM (
    SELECT
        lang,
        repo_name,
        search_key,
        search_key <<<-> %s as search_key_dist,
        label,
        label <<<-> %s as label_dist,
		detail,
        tags,
		path_id
    FROM $TABLE_NAME
    WHERE
        search_key <<%% %s
        OR
        label <<%% %s
) sub
WHERE
    CASE
        -- filter by (tags, langs, repos)
        WHEN
            (SELECT COUNT(*) FROM matching_tags) > 0
            AND (SELECT COUNT(*) FROM matching_langs) > 0
            AND (SELECT COUNT(*) FROM matching_repos) > 0
        THEN
            tags IN (select tags from matching_tags)
            AND lang IN (SELECT lang FROM matching_langs)
            AND repo_name IN (SELECT repo_name FROM matching_repos)

        -- filter by (tags, langs)
        WHEN
            (SELECT COUNT(*) FROM matching_tags) > 0
            AND (SELECT COUNT(*) FROM matching_langs) > 0
        THEN
            tags IN (select tags from matching_tags)
            AND lang IN (SELECT lang FROM matching_langs)

		-- filter by (langs, repos)
        WHEN
            (SELECT COUNT(*) FROM matching_langs) > 0
            AND (SELECT COUNT(*) FROM matching_repos) > 0
        THEN
            lang IN (SELECT lang FROM matching_langs)
            AND repo_name IN (SELECT repo_name FROM matching_repos)

		-- filter by (tags)
        WHEN (SELECT COUNT(*) FROM matching_tags) > 0
        THEN tags IN (SELECT tags FROM matching_tags)

		-- filter by (langs)
        WHEN (SELECT COUNT(*) FROM matching_langs) > 0
        THEN lang IN (SELECT lang FROM matching_langs)

		-- filter by (repos)
        WHEN (SELECT COUNT(*) FROM matching_repos) > 0
        THEN repo_name IN (SELECT repo_name FROM matching_repos)

		-- Need a truthy where condition in default case.
		ELSE search_key IS NOT NULL
    END
ORDER BY LEAST(search_key_dist, label_dist), repo_name, tags, lang;
`

// scanDocumentationSearchResults reads the documentation search results rows. If no rows matched, (nil, nil) is returned.
func (s *Store) scanDocumentationSearchResults(rows *sql.Rows, queryErr error) (_ []precise.DocumentationSearchResult, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	var results []precise.DocumentationSearchResult
	for rows.Next() {
		var (
			result                      precise.DocumentationSearchResult
			tags                        string
			searchKeyDist, labelDist float32
		)
		if err := rows.Scan(
			&result.Lang,
			&result.RepoName,
			&result.SearchKey,
			&searchKeyDist,
			&result.Label,
			&labelDist,
			&result.Detail,
			&tags,
			&result.PathID,
		); err != nil {
			return nil, err
		}
		result.Tags = strings.Fields(tags)
		results = append(results, result)
	}
	return results, nil
}
