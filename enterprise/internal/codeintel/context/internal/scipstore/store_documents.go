package store

import (
	"bytes"
	"context"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/scip/bindings/go/scip"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/protobuf/proto"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func (s *store) GetSCIPDocumentsByFuzzySelector(ctx context.Context, selector, scipNameType string) (documents []*scip.Document, err error) {
	ctx, _, endObservation := s.operations.getSCIPDocumentsByFuzzySelector.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.String("selector", selector),
	}})
	defer endObservation(1, observation.Args{})

	rows, err := s.db.Query(ctx, sqlf.Sprintf(getSCIPDocumentsByFuzzySelectorQuery, selector, scipNameType))
	if err != nil {
		return nil, err
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	for rows.Next() {
		var b []byte
		if err := rows.Scan(&b); err != nil {
			return nil, err
		}

		d, err := shared.Decompressor.Decompress(bytes.NewReader(b))
		if err != nil {
			return nil, err
		}

		var doc scip.Document
		if err := proto.Unmarshal(d, &doc); err != nil {
			return nil, err
		}

		documents = append(documents, &doc)
	}

	return documents, nil
}

const getSCIPDocumentsByFuzzySelectorQuery = `
SELECT
    cd.raw_scip_payload
FROM codeintel_scip_symbols_lookup ssl
JOIN
    codeintel_scip_symbols css ON css.descriptor_no_suffix_id = ssl.id AND css.upload_id = ssl.upload_id
JOIN codeintel_scip_document_lookup cdl ON cdl.upload_id = ssl.upload_id
JOIN codeintel_scip_documents cd ON cd.id = cdl.document_id
WHERE
    ssl.name = %s AND ssl.scip_name_type = %s;
`

func (s *store) GetSCIPDocumentsByPreciseSelector(ctx context.Context, uploadID int, schemeID, packageManagerID, packageManagerName, packageVersionID, descriptorID int) (documents []*scip.Document, err error) {
	ctx, _, endObservation := s.operations.getSCIPDocumentsByPreciseSelector.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Int("uploadID", uploadID),
		attribute.Int("schemeID", schemeID),
		attribute.Int("packageManagerID", packageManagerID),
		attribute.Int("packageManagerName", packageManagerName),
		attribute.Int("packageVersionID", packageVersionID),
		attribute.Int("descriptorID", descriptorID),
	}})
	defer endObservation(1, observation.Args{})

	rows, err := s.db.Query(ctx, sqlf.Sprintf(getSCIPDocumentsByFuzzySelectorQuery, uploadID, schemeID, packageManagerID, packageManagerName, packageVersionID, descriptorID))
	if err != nil {
		return nil, err
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	for rows.Next() {
		var b []byte
		if err := rows.Scan(&b); err != nil {
			return nil, err
		}

		d, err := shared.Decompressor.Decompress(bytes.NewReader(b))
		if err != nil {
			return nil, err
		}

		var doc scip.Document
		if err := proto.Unmarshal(d, &doc); err != nil {
			return nil, err
		}

		documents = append(documents, &doc)
	}

	return documents, nil
}

const getSCIPDocumentsByPreciseSelectorQuery = `
SELECT cd.raw_scip_payload
FROM codeintel_scip_symbols css
JOIN codeintel_scip_document_lookup cdl ON cdl.upload_id = css.upload_id
JOIN codeintel_scip_documents cd ON cd.id = cdl.document_id
WHERE
   css.upload_id = %s
   AND css.scheme_id IN (SELECT id FROM codeintel_scip_symbols_lookup ssl WHERE ssl.name = %s AND ssl.scip_name_type = 'SCHEME' AND ssl.upload_id = css.upload_id)
   AND css.package_manager_id IN (SELECT id FROM codeintel_scip_symbols_lookup ssl WHERE ssl.name = %s AND ssl.scip_name_type = 'PACKAGE_MANAGER' AND ssl.upload_id = css.upload_id)
   AND css.package_name_id IN (SELECT id FROM codeintel_scip_symbols_lookup ssl WHERE ssl.name = %s AND ssl.scip_name_type = 'PACKAGE_NAME' AND ssl.upload_id = css.upload_id)
   AND css.package_version_id IN (SELECT id FROM codeintel_scip_symbols_lookup ssl WHERE ssl.name = %s AND ssl.scip_name_type = 'PACKAGE_VERSION' AND ssl.upload_id = css.upload_id)
   AND css.descriptor_id IN (SELECT id FROM codeintel_scip_symbols_lookup ssl WHERE ssl.name = %s AND ssl.scip_name_type = 'DESCRIPTOR' AND ssl.upload_id = css.upload_id);
`
