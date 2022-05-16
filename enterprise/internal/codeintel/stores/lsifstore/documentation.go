package lsifstore

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/lsifstore/apidocs"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// DocumentationPage returns the documentation page with the given PathID.
func (s *Store) DocumentationPage(ctx context.Context, bundleID int, pathID string) (_ *precise.DocumentationPageData, err error) {
	ctx, _, endObservation := s.operations.documentationPage.With(ctx, &err, observation.Args{LogFields: []log.Field{
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
	ctx, _, endObservation := s.operations.documentationPathInfo.With(ctx, &err, observation.Args{LogFields: []log.Field{
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
	ctx, _, endObservation := s.operations.documentationIDsToPathIDs.With(ctx, &err, observation.Args{LogFields: []log.Field{
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
	ctx, _, endObservation := s.operations.documentationPathIDToID.With(ctx, &err, observation.Args{LogFields: []log.Field{
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
	ctx, _, endObservation := s.operations.documentationPathIDToFilePath.With(ctx, &err, observation.Args{LogFields: []log.Field{
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
	ctx, trace, endObservation := operation.With(ctx, &err, observation.Args{LogFields: []log.Field{
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

	trace.Log(log.Int("numRanges", len(documentData.Document.Ranges)))
	var found *precise.RangeData
	for _, rn := range documentData.Document.Ranges {
		if rn.DocumentationResultID == resultID {
			//nolint:exportloopref
			// We immediately break, so there are no more loop iterations, which means
			// the value of rn will not change.
			found = &rn
			break
		}
	}
	trace.Log(log.Bool("found", found == nil))
	if found == nil {
		return nil, 0, errors.New("not found")
	}

	orderedResultIDs := extractResultIDs([]precise.RangeData{*found}, extractor)
	locationsMap, totalCount, err := s.locations(ctx, bundleID, orderedResultIDs, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	trace.Log(log.Int("totalCount", totalCount))

	locations := make([]Location, 0, limit)
	for _, resultID := range orderedResultIDs {
		locations = append(locations, locationsMap[resultID]...)
	}

	return locations, totalCount, nil
}

// documentationSearchRepoNameIDs searches API documentation repositories in either the "public" or
// "private" table. It returns primary key IDs from the lsif_data_docs_search_repo_names_$SUFFIX
// table.
//
// 🚨 SECURITY: If the input tableSuffix is "private", then it is the callers responsibility to
// enforce that the user only has the ability to view results that are from repositories they have
// access to.
func (s *Store) documentationSearchRepoNameIDs(ctx context.Context, tableSuffix string, possibleRepos []string) (_ []int64, err error) {
	ctx, _, endObservation := s.operations.documentationSearchRepoNameIDs.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("table", tableSuffix),
		log.String("possibleRepos", fmt.Sprint(possibleRepos)),
	}})
	defer endObservation(1, observation.Args{})

	return basestore.ScanInt64s(s.Store.Query(ctx, sqlf.Sprintf(
		strings.ReplaceAll(documentationSearchRepoNameIDsQuery, "$SUFFIX", tableSuffix),
		// In the repo clause we forbid substring matches. Although performance would be fine, as user
		// you often do not want e.g. "go net/http" to match a repo like github.com/jane/goexploration
		// because then your search would be limited to those repos only. Note that substring matching
		// only applies to lexemes, so e.g. `sourcegraph/sourcegraph` would still match
		// `sourcegraph/sourcegraph-testing`
		apidocs.RepoSearchQuery("tsv", possibleRepos),
	)))
}

const documentationSearchRepoNameIDsQuery = `
-- source: enterprise/internal/codeintel/stores/lsifstore/documentation.go:documentationSearchRepos
SELECT id
FROM lsif_data_docs_search_repo_names_$SUFFIX
WHERE %s -- e.g. (tsv @@ '''gorilla'' <-> ''/'' <-> ''mux''')
LIMIT 100
`

// maximum candidates we'll consider: the more candidates we have the better ranking we get, but the
// slower searching is. See https://about.sourcegraph.com/blog/postgres-text-search-balancing-query-time-and-relevancy/
// 10,000 is an arbitrary choice based on current Sourcegraph.com corpus size and performance, it'll
// be tuned as we scale to more repos if perf gets worse or we find we need better relevance.
var debugAPIDocsSearchCandidates, _ = strconv.ParseInt(env.Get("DEBUG_API_DOCS_SEARCH_CANDIDATES", "10000", "maximum candidates for consideration in API docs search"), 10, 64)

// DocumentationSearch searches API documentation in either the "public" or "private" table.
//
// 🚨 SECURITY: If the input tableSuffix is "private", then it is the callers responsibility to
// enforce that the user only has the ability to view results that are from repositories they have
// access to.
func (s *Store) DocumentationSearch(ctx context.Context, tableSuffix, query string, repos []string) (_ []precise.DocumentationSearchResult, err error) {
	ctx, _, endObservation := s.operations.documentationSearch.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("table", tableSuffix),
		log.String("query", query),
		log.String("repos", fmt.Sprint(repos)),
	}})
	defer endObservation(1, observation.Args{})

	// If a search would exceed 3 seconds, just give up. We'll issue a lo of searches, so we'd
	// rather have this than the DB pressure.
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	q := apidocs.ParseQuery(query)
	resultLimit := 10

	// Can we find matching repository names? If so, we'll filter to just those. This helps reduce
	// our search space, which means we can use more fuzzy matching (like substring matching) for
	// better results - so this is ideal.
	matchingRepoNameIDs, err := s.documentationSearchRepoNameIDs(ctx, tableSuffix, q.PossibleRepos)

	var primary *sqlf.Query
	if len(matchingRepoNameIDs) > 0 {
		// Primary WHERE clauses to use when searching over a smaller subset of repositories.
		var clauses []*sqlf.Query
		clauses = append(clauses, apidocs.TextSearchQuery("result.search_key_tsv", q.MainTerms, q.SubStringMatches))
		clauses = append(clauses, apidocs.TextSearchQuery("result.search_key_reverse_tsv", apidocs.Reverse(q.MainTerms), q.SubStringMatches))
		clauses = append(clauses, apidocs.TextSearchQuery("result.label_tsv", q.MainTerms, q.SubStringMatches))
		clauses = append(clauses, apidocs.TextSearchQuery("result.label_reverse_tsv", apidocs.Reverse(q.MainTerms), q.SubStringMatches))

		// We found matching repo names, so filter to just those.
		primaryClause := sqlf.Sprintf("(%s)", sqlf.Join(clauses, "OR"))
		primary = sqlf.Sprintf("%s AND result.repo_name_id = ANY(%s)", primaryClause, pq.Array(matchingRepoNameIDs))
	} else {
		// Primary WHERE clauses to use when searching over ALL repositories. Because there is so
		// much to search over in this case, we do not do prefix/suffix matching (so no reverse
		// index lookups), and do not use ":*" tsquery prefix matching operators (which are very
		// slow).
		var clauses []*sqlf.Query
		clauses = append(clauses, apidocs.TextSearchQuery("result.search_key_tsv", q.MainTerms, false))
		clauses = append(clauses, apidocs.TextSearchQuery("result.label_tsv", q.MainTerms, false))
		primary = sqlf.Sprintf("(%s)", sqlf.Join(clauses, "OR"))
	}

	langTagsClause := apidocs.TextSearchQuery("tsv", q.MetaTerms, q.SubStringMatches)
	return s.scanDocumentationSearchResults(s.Store.Query(ctx, sqlf.Sprintf(
		strings.ReplaceAll(documentationSearchQuery, "$SUFFIX", tableSuffix),
		langTagsClause, // matching_lang_names CTE WHERE conditions
		langTagsClause, // matching_tags CTE WHERE conditions

		primary,                      // primary WHERE clause
		debugAPIDocsSearchCandidates, // maximum candidates for consideration.

		apidocs.TextSearchRank("search_key_tsv", q.MainTerms, q.SubStringMatches),                          // search_key_rank
		apidocs.TextSearchRank("search_key_reverse_tsv", apidocs.Reverse(q.MainTerms), q.SubStringMatches), // search_key_reverse_rank
		apidocs.TextSearchRank("label_tsv", q.MainTerms, q.SubStringMatches),                               // label_rank
		apidocs.TextSearchRank("label_reverse_tsv", apidocs.Reverse(q.MainTerms), q.SubStringMatches),      // label_reverse_rank

		resultLimit, // result limit
	)))
}

const documentationSearchQuery = `
-- source: enterprise/internal/codeintel/stores/lsifstore/documentation.go:DocumentationSearch
WITH
	-- Can we find a matching language name? If so, that could limit our search space greatly.
	matching_lang_names AS (
		SELECT id, lang_name
		FROM lsif_data_docs_search_lang_names_$SUFFIX
		WHERE %s -- e.g. (tsv @@ '''net'' <-> ''/'' <-> ''http'':*' OR tsv @@ '''Router'':*')
		LIMIT 1
	),

	-- Can we find a matching sequence of documentation/symbol tags? e.g. "private variable".
	-- If so, then we pick the top 10 and search only documentation nodes that have that same
	-- sequence of tags.
	matching_tags AS (
		SELECT id, tags
		FROM lsif_data_docs_search_tags_$SUFFIX
		WHERE %s -- e.g. (tsv @@ '''net'' <-> ''/'' <-> ''http'':*' OR tsv @@ '''Router'':*')
		LIMIT 10
	)
SELECT
	result_id,
	repo_id,
	path_id,
	detail,
	search_key,
	label,
	lang_name,
	repo_name,
	tags
FROM (
	SELECT result.id::bigint AS result_id, *
	FROM lsif_data_docs_search_$SUFFIX result
	JOIN lsif_data_docs_search_lang_names_$SUFFIX langnames ON langnames.id = result.lang_name_id
	JOIN lsif_data_docs_search_repo_names_$SUFFIX reponames ON reponames.id = result.repo_name_id
	JOIN lsif_data_docs_search_tags_$SUFFIX tags ON tags.id = result.tags_id
	WHERE
		-- If we found matching repo names, then we're searching over a very limited subset of repos
		-- and can afford to enable the (much) more expensive substring matching which makes search
		-- fuzzier. This is very nice, but just too expensive to use when searching over all repos.
		-- For example:
		--
		-- (
		--   (
		--     result.search_key_tsv @@ '''net'' <-> ''/'' <-> ''http'':*'
		--     OR result.search_key_tsv @@ '''Router'':*'
		--   ) OR (
		--     result.search_key_reverse_tsv @@ '''ptth'' <-> ''/'' <-> ''ten'':*'
		--     OR result.search_key_reverse_tsv @@ '''retuoR'':*'
		--   ) OR (
		--     result.label_tsv @@ '''net'' <-> ''/'' <-> ''http'':*'
		--     OR result.label_tsv @@ '''Router'':*'
		--   ) OR (
		--     result.label_reverse_tsv @@ '''ptth'' <-> ''/'' <-> ''ten'':*'
		--     OR result.label_reverse_tsv @@ '''retuoR'':*'
		--   )
		-- )
		-- AND result.repo_name_id = ANY(...)
		--
		-- If we're matching over many repos, we choose a much lighter weight query that is very
		-- strict and thus turns up worse results:
		--
		-- (
		--   (
		--     result.search_key_tsv @@ '''net'' <-> ''/'' <-> ''http'''
		--     OR result.search_key_tsv @@ '''Router'''
		--   ) OR (
		--     result.label_tsv @@ '''net'' <-> ''/'' <-> ''http'''
		--     OR result.label_tsv @@ '''Router'''
		--   )
		-- )
		--
		%s

		-- If we found matching lang names, filter to just those.
		AND (CASE WHEN (SELECT COUNT(*) FROM matching_lang_names) > 0 THEN
			result.lang_name_id = ANY(array(SELECT id FROM matching_lang_names))
		ELSE TRUE END)

		-- If we found matching tags, filter to just those.
		AND (CASE WHEN (SELECT COUNT(*) FROM matching_tags) > 0 THEN
			result.tags_id = ANY(array(SELECT id FROM matching_tags))
		ELSE TRUE END)

	-- maximum candidates we'll consider: the more candidates we have the better ranking we get,
	-- but the slower searching is. See https://about.sourcegraph.com/blog/postgres-text-search-balancing-query-time-and-relevancy/
	LIMIT %s
) sub
-- Select only results that come from the latest upload, since lsif_data_docs_search_* may have
-- results from multiple uploads (the table is cleaned up asynchronously in the background to avoid
-- lock contention at insert time.) We cannot do this in the inner CTE above as it adds ~80ms per
-- result to check, and the inner CTE selects ~10k results.
WHERE sub.dump_id = (
	SELECT dump_id FROM lsif_data_docs_search_current_$SUFFIX current
	WHERE
		current.dump_id = sub.dump_id
		AND current.dump_root = sub.dump_root
		AND lang_name_id = sub.lang_name_id
	ORDER BY current.created_at DESC, id
	LIMIT 1
)
ORDER BY
	-- First rank by search keys, as those are ideally super specific if you write the correct
	-- format.
	--
	-- e.g. for 1st arg: (ts_rank_cd(search_key_tsv, '''net'' <-> ''/'' <-> ''http'':*', 2) + ts_rank_cd(search_key_tsv, '''Router'':*', 2))
	-- e.g. for 2nd arg: (ts_rank_cd(search_key_reverse_tsv, '''retuoR'':*', 2) + ts_rank_cd(search_key_reverse_tsv, '''ptth'' <-> ''/'' <-> ''ten'':*', 2))
	GREATEST(%s, %s) DESC,

	-- Secondarily rank by label, e.g. function signature. These contain less specific info and
	-- due to e.g. containing arguments a function takes, have higher chance of collision with the
	-- desired symbol result, producing a bad match.
	--
	-- e.g. for 1st arg: (ts_rank_cd(label_tsv, '''net'' <-> ''/'' <-> ''http'':*', 2) + ts_rank_cd(label_tsv, '''Router'':*', 2))
	-- e.g. for 2nd arg: (ts_rank_cd(label_reverse_tsv, '''retuoR'':*', 2) + ts_rank_cd(label_reverse_tsv, '''ptth'' <-> ''/'' <-> ''ten'':*', 2))
	GREATEST(%s, %s) DESC,

	-- If all else failed, sort by something reasonable and deterministic.
	repo_name DESC,
	tags DESC,
	result_id DESC
LIMIT %s -- Maximum results we'll actually return.
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
			result precise.DocumentationSearchResult
			tags   string
		)
		if err := rows.Scan(
			&result.ID,
			&result.RepoID,
			&result.PathID,
			&result.Detail,
			&result.SearchKey,
			&result.Label,
			&result.Lang,
			&result.RepoName,
			&tags,
		); err != nil {
			return nil, err
		}
		result.Tags = strings.Fields(tags)
		results = append(results, result)
	}
	return results, nil
}
