package lsifstore

import (
	"context"

	"github.com/keegancsmith/sqlf"
	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
)

// PackageInformation looks up package information data by identifier.
func (s *Store) PackageInformation(ctx context.Context, bundleID int, path, packageInformationID string) (_ precise.PackageInformationData, _ bool, err error) {
	ctx, _, endObservation := s.operations.packageInformation.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("bundleID", bundleID),
		log.String("path", path),
		log.String("packageInformationID", packageInformationID),
	}})
	defer endObservation(1, observation.Args{})

	documentData, exists, err := s.scanFirstDocumentData(s.Store.Query(ctx, sqlf.Sprintf(packageInformationQuery, bundleID, path)))
	if err != nil || !exists {
		return precise.PackageInformationData{}, false, err
	}

	packageInformationData, exists := documentData.Document.PackageInformation[precise.ID(packageInformationID)]
	return packageInformationData, exists, nil
}

const packageInformationQuery = `
-- source: internal/codeintel/stores/lsifstore/packages.go:PackageInformation
SELECT
	dump_id,
	path,
	data,
	NULL AS ranges,
	NULL AS hovers,
	NULL AS monikers,
	packages,
	NULL AS diagnostics
FROM
	lsif_data_documents
WHERE
	dump_id = %s AND
	path = %s
LIMIT 1
`
