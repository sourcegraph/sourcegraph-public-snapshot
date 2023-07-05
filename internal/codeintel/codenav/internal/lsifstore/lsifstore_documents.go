package lsifstore

import (
	"bytes"
	"context"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	"github.com/sourcegraph/scip/bindings/go/scip"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/protobuf/proto"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/shared/symbols"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/types"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func (s *store) SCIPDocument(ctx context.Context, id int, path string) (_ *scip.Document, err error) {
	ctx, _, endObservation := s.operations.scipDocument.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.String("path", path),
		attribute.Int("uploadID", id),
	}})
	defer endObservation(1, observation.Args{})

	scanner := basestore.NewFirstScanner(func(dbs dbutil.Scanner) (*scip.Document, error) {
		var compressedSCIPPayload []byte
		if err := dbs.Scan(&compressedSCIPPayload); err != nil {
			return nil, err
		}

		scipPayload, err := shared.Decompressor.Decompress(bytes.NewReader(compressedSCIPPayload))
		if err != nil {
			return nil, err
		}

		var document scip.Document
		if err := proto.Unmarshal(scipPayload, &document); err != nil {
			return nil, err
		}
		return &document, nil
	})
	doc, _, err := scanner(s.db.Query(ctx, sqlf.Sprintf(fetchSCIPDocumentQuery, id, path)))
	return doc, err
}

const fetchSCIPDocumentQuery = `
SELECT sd.raw_scip_payload
FROM codeintel_scip_document_lookup sid
JOIN codeintel_scip_documents sd ON sd.id = sid.document_id
WHERE
	sid.upload_id = %s AND
	sid.document_path = %s
`

func (s *store) GetFullSCIPNameByDescriptor(ctx context.Context, uploadID []int, symbolNames []string) (names []*types.SCIPNames, err error) {
	ctx, _, endObservation := s.operations.getFullSCIPNameByDescriptor.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{}})
	defer endObservation(1, observation.Args{})

	symbolNamesIlike, err := formatSymbolNamesToLikeClause(symbolNames)
	if err != nil {
		return nil, err
	}

	query := sqlf.Sprintf(getFullSCIPNameByDescriptorQuery, pq.Array(symbolNamesIlike), pq.Array(uploadID))
	rows, err := s.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	// fmt.Println("This is my query >>>", query.Query(sqlf.PostgresBindVar), query.Args())

	for rows.Next() {
		var n types.SCIPNames
		if err := rows.Scan(&n.Scheme, &n.PackageManager, &n.PackageName, &n.PackageVersion, &n.Descriptor); err != nil {
			return nil, err
		}

		names = append(names, &n)
	}

	return names, nil
}

const getFullSCIPNameByDescriptorQuery = `
SELECT DISTINCT
    ssl2.name AS scheme,
    ssl3.name AS package_manager,
    ssl4.name AS package_name,
    ssl5.name AS package_version,
    ssl6.name AS descriptor
FROM codeintel_scip_symbols ss
JOIN codeintel_scip_symbols_lookup ssl1 ON ssl1.upload_id = ss.upload_id AND ssl1.id = ss.descriptor_id
JOIN codeintel_scip_symbols_lookup ssl2 ON ssl2.upload_id = ss.upload_id AND ssl2.id = ss.scheme_id
JOIN codeintel_scip_symbols_lookup ssl3 ON ssl3.upload_id = ss.upload_id AND ssl3.id = ss.package_manager_id
JOIN codeintel_scip_symbols_lookup ssl4 ON ssl4.upload_id = ss.upload_id AND ssl4.id = ss.package_name_id
JOIN codeintel_scip_symbols_lookup ssl5 ON ssl5.upload_id = ss.upload_id AND ssl5.id = ss.package_version_id
JOIN codeintel_scip_symbols_lookup ssl6 ON ssl6.upload_id = ss.upload_id AND ssl6.id = ss.descriptor_id
WHERE
    ssl1.name ILIKE ANY(%s) AND
    ssl1.scip_name_type = 'DESCRIPTOR' AND
    ssl2.scip_name_type = 'SCHEME' AND
    ssl3.scip_name_type = 'PACKAGE_MANAGER' AND
    ssl4.scip_name_type = 'PACKAGE_NAME' AND
    ssl5.scip_name_type = 'PACKAGE_VERSION' AND
    ssl6.scip_name_type = 'DESCRIPTOR' AND
	ssl1.upload_id = ANY(%s);
`

func formatSymbolNamesToLikeClause(symbolNames []string) ([]string, error) {
	explodedSymbols := make([]string, 0, len(symbolNames))
	for _, symbolName := range symbolNames {
		ex, err := symbols.NewExplodedSymbol(symbolName)
		if err != nil {
			return nil, err
		}
		explodedSymbols = append(
			explodedSymbols,
			"%"+ex.Descriptor+"%",
		)
	}

	return explodedSymbols, nil
}
