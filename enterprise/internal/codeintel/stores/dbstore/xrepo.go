package dbstore

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/keegancsmith/sqlf"
	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
)

// DefinitionDumpsLimit is the maximum number of records that can be returned from DefinitionDumps.
var DefinitionDumpsLimit, _ = strconv.ParseInt(env.Get("PRECISE_CODE_INTEL_DEFINITION_DUMPS_LIMIT", "100", "The maximum number of dumps that can define the same package."), 10, 64)

// DefinitionDumps returns the set of dumps that define at least one of the given monikers.
func (s *Store) DefinitionDumps(ctx context.Context, monikers []precise.QualifiedMonikerData) (_ []Dump, err error) {
	ctx, trace, endObservation := s.operations.definitionDumps.WithAndLogger(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("numMonikers", len(monikers)),
		log.String("monikers", monikersToString(monikers)),
	}})
	defer endObservation(1, observation.Args{})

	if len(monikers) == 0 {
		return nil, nil
	}

	qs := make([]*sqlf.Query, 0, len(monikers))
	for _, moniker := range monikers {
		qs = append(qs, sqlf.Sprintf("(%s, %s, %s)", moniker.Scheme, moniker.Name, moniker.Version))
	}

	authzConds, err := database.AuthzQueryConds(ctx, s.Store.Handle().DB())
	if err != nil {
		return nil, err
	}

	dumps, err := scanDumps(s.Query(ctx, sqlf.Sprintf(definitionDumpsQuery, sqlf.Join(qs, ", "), authzConds, DefinitionDumpsLimit)))
	if err != nil {
		return nil, err
	}
	trace.Log(log.Int("numDumps", len(dumps)))

	return dumps, nil
}

const definitionDumpsQuery = `
-- source: enterprise/internal/codeintel/stores/dbstore/xrepo.go:DefinitionDumps
WITH
ranked_uploads AS (
	SELECT
		u.id,
		-- Rank each upload providing the same package from the same directory
		-- within a repository by commit date. We'll choose the oldest commit
		-- date as the canonical choice used to resolve the current definitions
		-- request.
		` + packageRankingQueryFragment + ` AS rank
	FROM lsif_uploads u
	JOIN lsif_packages p ON p.dump_id = u.id
	JOIN repo ON repo.id = u.repository_id
	WHERE
		-- Don't match deleted uploads
		u.state = 'completed' AND
		(p.scheme, p.name, p.version) IN (%s) AND
		%s -- authz conds
),
canonical_uploads AS (
	SELECT ru.id
	FROM ranked_uploads ru
	WHERE ru.rank = 1
	ORDER BY ru.id
	LIMIT %s
)
SELECT
	u.id,
	u.commit,
	u.root,
	EXISTS (` + visibleAtTipSubselectQuery + `) AS visible_at_tip,
	u.uploaded_at,
	u.state,
	u.failure_message,
	u.started_at,
	u.finished_at,
	u.process_after,
	u.num_resets,
	u.num_failures,
	u.repository_id,
	u.repository_name,
	u.indexer,
	u.associated_index_id
FROM lsif_dumps_with_repository_name u
WHERE u.id IN (SELECT id FROM canonical_uploads)
`

// packageRankingQueryFragment uses `lsif_uploads u` JOIN `lsif_packages p` to return a rank
// for each row grouped by package and source code location and ordered by the associated Git
// commit date.
const packageRankingQueryFragment = `
rank() OVER (
	PARTITION BY
		-- Group providers of the same package together
		p.scheme, p.name, p.version,
		-- Defined by the same directory within a repository
		u.repository_id, u.indexer, u.root
	ORDER BY
		-- Rank each grouped upload by the associated commit date
		u.committed_at,
		-- Break ties via the unique identifier
		u.id
)
`

// ReferenceIDsAndFilters returns visible uploads that refer (via package information) to any of the
// given monikers' packages. Each upload identifier in the result set is paired with one or more
// compressed bloom filters that encode more precisely the set of identifiers imported from dependent
// packages.
//
// Visibility is determined in two parts: if the index belongs to the given repository, it is visible if
// it can be seen from the given index; otherwise, an index is visible if it can be seen from the tip of
// the default branch of its own repository.
func (s *Store) ReferenceIDsAndFilters(ctx context.Context, repositoryID int, commit string, monikers []precise.QualifiedMonikerData, limit, offset int) (_ PackageReferenceScanner, _ int, err error) {
	ctx, trace, endObservation := s.operations.referenceIDsAndFilters.WithAndLogger(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("repositoryID", repositoryID),
		log.String("commit", commit),
		log.Int("numMonikers", len(monikers)),
		log.String("monikers", monikersToString(monikers)),
		log.Int("limit", limit),
		log.Int("offset", offset),
	}})
	defer endObservation(1, observation.Args{})

	if len(monikers) == 0 {
		return PackageReferenceScannerFromSlice(), 0, nil
	}

	qs := make([]*sqlf.Query, 0, len(monikers))
	for _, moniker := range monikers {
		qs = append(qs, sqlf.Sprintf("(%s, %s, %s)", moniker.Scheme, moniker.Name, moniker.Version))
	}

	visibleUploadsQuery := makeVisibleUploadsQuery(repositoryID, commit)

	authzConds, err := database.AuthzQueryConds(ctx, s.Store.Handle().DB())
	if err != nil {
		return nil, 0, err
	}

	totalCount, _, err := basestore.ScanFirstInt(s.Query(ctx, sqlf.Sprintf(
		referenceIDsAndFiltersCountQuery,
		visibleUploadsQuery,
		repositoryID,
		sqlf.Join(qs, ", "),
		authzConds,
	)))
	if err != nil {
		return nil, 0, err
	}
	trace.Log(log.Int("totalCount", totalCount))

	rows, err := s.Query(ctx, sqlf.Sprintf(
		referenceIDsAndFiltersQuery,
		visibleUploadsQuery,
		repositoryID,
		sqlf.Join(qs, ", "),
		authzConds,
		limit,
		offset,
	))
	if err != nil {
		return nil, 0, err
	}

	return packageReferenceScannerFromRows(rows), totalCount, nil
}

const referenceIDsAndFiltersCTEDefinitions = `
-- source: enterprise/internal/codeintel/stores/dbstore/xrepo.go:ReferenceIDsAndFilters
WITH
visible_uploads AS (
	(%s)
	UNION
	(SELECT uvt.upload_id FROM lsif_uploads_visible_at_tip uvt WHERE uvt.repository_id != %s AND uvt.is_default_branch)
)
`

const referenceIDsAndFiltersBaseQuery = `
FROM lsif_references r
LEFT JOIN lsif_dumps u ON u.id = r.dump_id
JOIN repo ON repo.id = u.repository_id
WHERE
	(r.scheme, r.name, r.version) IN (%s) AND
	r.dump_id IN (SELECT * FROM visible_uploads) AND
	%s -- authz conds
`

const referenceIDsAndFiltersQuery = referenceIDsAndFiltersCTEDefinitions + `
SELECT r.dump_id, r.scheme, r.name, r.version, r.filter
` + referenceIDsAndFiltersBaseQuery + `
ORDER BY dump_id
LIMIT %s OFFSET %s
`

const referenceIDsAndFiltersCountQuery = referenceIDsAndFiltersCTEDefinitions + `
SELECT COUNT(distinct r.dump_id)
` + referenceIDsAndFiltersBaseQuery

func monikersToString(vs []precise.QualifiedMonikerData) string {
	strs := make([]string, 0, len(vs))
	for _, v := range vs {
		strs = append(strs, fmt.Sprintf("%s:%s:%s:%s", v.Kind, v.Scheme, v.Identifier, v.Version))
	}

	return strings.Join(strs, ", ")
}

// ReferencesForUpload returns the set of import monikers attached to the given upload identifier. The
// scanner will return nulls for the Filter field as it's expected to be unused (and rather heavy) by
// callers.
func (s *Store) ReferencesForUpload(ctx context.Context, uploadID int) (_ PackageReferenceScanner, err error) {
	ctx, endObservation := s.operations.referencesForUpload.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("uploadID", uploadID),
	}})
	defer endObservation(1, observation.Args{})

	rows, err := s.Query(ctx, sqlf.Sprintf(referencesForUploadQuery, uploadID))
	if err != nil {
		return nil, err
	}

	return packageReferenceScannerFromRows(rows), nil
}

const referencesForUploadQuery = `
-- source: enterprise/internal/codeintel/stores/dbstore/xrepo.go:ReferencesForUpload
SELECT r.dump_id, r.scheme, r.name, r.version, NULL as filter
FROM lsif_references r
WHERE dump_id = %s
ORDER BY r.scheme, r.name, r.version
`
