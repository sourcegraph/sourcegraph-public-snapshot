package lsifstore

import (
	"context"
	"errors"

	"github.com/keegancsmith/sqlf"
	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
)

// GetPackageInformation returns package information data by identifier.
func (s *store) GetPackageInformation(ctx context.Context, bundleID int, path, packageInformationID string) (_ precise.PackageInformationData, _ bool, err error) {
	ctx, _, endObservation := s.operations.getPackageInformation.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("bundleID", bundleID),
		log.String("path", path),
		log.String("packageInformationID", packageInformationID),
	}})
	defer endObservation(1, observation.Args{})

	documentData, exists, err := s.scanFirstDocumentData(s.db.Query(ctx, sqlf.Sprintf(
		packageInformationQuery,
		bundleID,
		path,
	)))
	if err != nil || !exists {
		return precise.PackageInformationData{}, false, err
	}

	if documentData.SCIPData != nil {
		return precise.PackageInformationData{}, false, errors.New("unexpected SCIP payload in GetPackageInformation")
	}

	packageInformationData, exists := documentData.LSIFData.PackageInformation[precise.ID(packageInformationID)]
	return packageInformationData, exists, nil
}

const packageInformationQuery = `
SELECT
	dump_id,
	path,
	data,
	NULL AS ranges,
	NULL AS hovers,
	NULL AS monikers,
	packages,
	NULL AS diagnostics,
	NULL AS scip_document
FROM
	lsif_data_documents
WHERE
	dump_id = %s AND
	path = %s
LIMIT 1
`
