package lsifstore

import (
	"context"
	"encoding/base64"
	"strings"

	"github.com/keegancsmith/sqlf"
	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// GetPackageInformation returns package information data by identifier.
func (s *store) GetPackageInformation(ctx context.Context, bundleID int, path, packageInformationID string) (_ precise.PackageInformationData, _ bool, err error) {
	ctx, _, endObservation := s.operations.getPackageInformation.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("bundleID", bundleID),
		log.String("path", path),
		log.String("packageInformationID", packageInformationID),
	}})
	defer endObservation(1, observation.Args{})

	if strings.HasPrefix(packageInformationID, "scip:") {
		packageInfo := strings.Split(packageInformationID, ":")
		if len(packageInfo) != 4 {
			return precise.PackageInformationData{}, false, errors.Newf("invalid package information ID %q", packageInformationID)
		}

		manager, err := base64.RawStdEncoding.DecodeString(packageInfo[1])
		if err != nil {
			return precise.PackageInformationData{}, false, err
		}
		name, err := base64.RawStdEncoding.DecodeString(packageInfo[2])
		if err != nil {
			return precise.PackageInformationData{}, false, err
		}
		version, err := base64.RawStdEncoding.DecodeString(packageInfo[3])
		if err != nil {
			return precise.PackageInformationData{}, false, err
		}

		return precise.PackageInformationData{
			Manager: string(manager),
			Name:    string(name),
			Version: string(version),
		}, true, nil
	}

	documentData, exists, err := s.scanFirstDocumentData(s.db.Query(ctx, sqlf.Sprintf(
		packageInformationQuery,
		bundleID,
		path,
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
(
	SELECT
		sd.id,
		sid.document_path,
		NULL AS data,
		NULL AS ranges,
		NULL AS hovers,
		NULL AS monikers,
		NULL AS packages,
		NULL AS diagnostics,
		sd.raw_scip_payload AS scip_document
	FROM codeintel_scip_document_lookup sid
	JOIN codeintel_scip_documents sd ON sd.id = sid.document_id
	WHERE
		sid.upload_id = %s AND
		sid.document_path = %s
	LIMIT 1
) UNION (
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
)
`
