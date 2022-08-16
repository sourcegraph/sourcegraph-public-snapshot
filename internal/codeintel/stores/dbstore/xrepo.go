package dbstore

import (
	"context"
	"fmt"
	"strings"

	"github.com/keegancsmith/sqlf"
	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
)

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

// ReferenceIDs returns visible uploads that refer (via package information) to any of the
// given monikers' packages.
//
// Visibility is determined in two parts: if the index belongs to the given repository, it is visible if
// it can be seen from the given index; otherwise, an index is visible if it can be seen from the tip of
// the default branch of its own repository.
func (s *Store) ReferenceIDs(ctx context.Context, repositoryID int, commit string, monikers []precise.QualifiedMonikerData, limit, offset int) (_ PackageReferenceScanner, _ int, err error) {
	ctx, trace, endObservation := s.operations.referenceIDs.With(ctx, &err, observation.Args{LogFields: []log.Field{
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

	authzConds, err := database.AuthzQueryConds(ctx, database.NewDBWith(s.logger, s.Store))
	if err != nil {
		return nil, 0, err
	}

	totalCount, _, err := basestore.ScanFirstInt(s.Query(ctx, sqlf.Sprintf(
		referenceIDsCountQuery,
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
		referenceIDsQuery,
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

const referenceIDsCTEDefinitions = `
-- source: internal/codeintel/stores/dbstore/xrepo.go:ReferenceIDs
WITH
visible_uploads AS (
	(%s)
	UNION
	(SELECT uvt.upload_id FROM lsif_uploads_visible_at_tip uvt WHERE uvt.repository_id != %s AND uvt.is_default_branch)
)
`

const referenceIDsBaseQuery = `
FROM lsif_references r
LEFT JOIN lsif_dumps u ON u.id = r.dump_id
JOIN repo ON repo.id = u.repository_id
WHERE
	(r.scheme, r.name, r.version) IN (%s) AND
	r.dump_id IN (SELECT * FROM visible_uploads) AND
	%s -- authz conds
`

const referenceIDsQuery = referenceIDsCTEDefinitions + `
SELECT r.dump_id, r.scheme, r.name, r.version
` + referenceIDsBaseQuery + `
ORDER BY dump_id
LIMIT %s OFFSET %s
`

const referenceIDsCountQuery = referenceIDsCTEDefinitions + `
SELECT COUNT(distinct r.dump_id)
` + referenceIDsBaseQuery

func monikersToString(vs []precise.QualifiedMonikerData) string {
	strs := make([]string, 0, len(vs))
	for _, v := range vs {
		strs = append(strs, fmt.Sprintf("%s:%s:%s:%s", v.Kind, v.Scheme, v.Identifier, v.Version))
	}

	return strings.Join(strs, ", ")
}

// ReferencesForUpload returns the set of import monikers attached to the given upload identifier.
func (s *Store) ReferencesForUpload(ctx context.Context, uploadID int) (_ PackageReferenceScanner, err error) {
	ctx, _, endObservation := s.operations.referencesForUpload.With(ctx, &err, observation.Args{LogFields: []log.Field{
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
-- source: internal/codeintel/stores/dbstore/xrepo.go:ReferencesForUpload
SELECT r.dump_id, r.scheme, r.name, r.version
FROM lsif_references r
WHERE dump_id = %s
ORDER BY r.scheme, r.name, r.version
`
