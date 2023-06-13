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
JOIN codeintel_scip_symbols css
	ON css.descriptor_no_suffix_id = ssl.id
	AND css.upload_id = ssl.upload_id
JOIN codeintel_scip_document_lookup cdl
	ON cdl.upload_id = ssl.upload_id
JOIN codeintel_scip_documents cd
	ON cd.id = cdl.document_id
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

	rows, err := s.db.Query(ctx, sqlf.Sprintf(getSCIPDocumentsByPreciseSelectorQuery, uploadID, schemeID, packageManagerID, packageManagerName, packageVersionID, descriptorID))
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

// func (s *store) GetFullSCIPNameByDescriptor(ctx context.Context, uploadID []int, symbolNames []string) (names []*contextshared.SCIPNames, err error) {
// 	// TODO: CHANGE operations
// 	ctx, _, endObservation := s.operations.getFullSCIPNameByDescriptor.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{}})
// 	defer endObservation(1, observation.Args{})

// 	symbolNamesIlike, err := formatSymbolNamesToLikeClause(symbolNames)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// fmt.Println("This is my symbolNamesIlike >>>", symbolNamesIlike)

// 	query := sqlf.Sprintf(getFullSCIPNameByDescriptorQuery, pq.Array(symbolNamesIlike), pq.Array(uploadID))
// 	rows, err := s.db.Query(ctx, query)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer func() { err = basestore.CloseRows(rows, err) }()

// 	// fmt.Println("This is my query >>>", query.Query(sqlf.PostgresBindVar), query.Args())

// 	for rows.Next() {
// 		var n contextshared.SCIPNames
// 		if err := rows.Scan(&n.Scheme, &n.PackageManager, &n.PackageName, &n.PackageVersion, &n.Descriptor); err != nil {
// 			return nil, err
// 		}

// 		names = append(names, &n)
// 	}

// 	return names, nil
// }

// const getFullSCIPNameByDescriptorQuery = `
// SELECT DISTINCT
//     ssl2.name AS scheme,
//     ssl3.name AS package_manager,
//     ssl4.name AS package_name,
//     ssl5.name AS package_version,
//     ssl6.name AS descriptor
// FROM codeintel_scip_symbols ss
// JOIN codeintel_scip_symbols_lookup ssl1 ON ssl1.upload_id = ss.upload_id AND ssl1.id = ss.descriptor_id
// JOIN codeintel_scip_symbols_lookup ssl2 ON ssl2.upload_id = ss.upload_id AND ssl2.id = ss.scheme_id
// JOIN codeintel_scip_symbols_lookup ssl3 ON ssl3.upload_id = ss.upload_id AND ssl3.id = ss.package_manager_id
// JOIN codeintel_scip_symbols_lookup ssl4 ON ssl4.upload_id = ss.upload_id AND ssl4.id = ss.package_name_id
// JOIN codeintel_scip_symbols_lookup ssl5 ON ssl5.upload_id = ss.upload_id AND ssl5.id = ss.package_version_id
// JOIN codeintel_scip_symbols_lookup ssl6 ON ssl6.upload_id = ss.upload_id AND ssl6.id = ss.descriptor_id
// WHERE
//     ssl1.name ILIKE ANY(%s) AND
//     ssl1.scip_name_type = 'DESCRIPTOR' AND
//     ssl2.scip_name_type = 'SCHEME' AND
//     ssl3.scip_name_type = 'PACKAGE_MANAGER' AND
//     ssl4.scip_name_type = 'PACKAGE_NAME' AND
//     ssl5.scip_name_type = 'PACKAGE_VERSION' AND
//     ssl6.scip_name_type = 'DESCRIPTOR' AND
// 	ssl1.upload_id = ANY(%s);
// `

// func formatSymbolNamesToLikeClause(symbolNames []string) ([]string, error) {
// 	explodedSymbols := make([]string, 0, len(symbolNames))
// 	for _, symbolName := range symbolNames {
// 		ex, err := symbols.NewExplodedSymbol(symbolName)
// 		if err != nil {
// 			return nil, err
// 		}
// 		explodedSymbols = append(
// 			explodedSymbols,
// 			"%"+ex.Descriptor+"%",
// 		)
// 	}

// 	return explodedSymbols, nil
// }

// func (s *store) GetSCIPDocumentsBySymbolNames(ctx context.Context, uploadID int, symbolNames []string) (documents []*scip.Document, err error) {
// 	ctx, _, endObservation := s.operations.getSCIPDocumentsByPreciseSelector.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
// 		attribute.Int("uploadID", uploadID),
// 	}})
// 	defer endObservation(1, observation.Args{})

// 	q := sqlf.Sprintf(
// 		getDocumentsBySymbolNameQuery,
// 		pq.Array(formatSymbolNamesToLikeClause(symbolNames)),
// 		uploadID,
// 	)

// 	rows, err := s.db.Query(ctx, q)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer func() { err = basestore.CloseRows(rows, err) }()

// 	for rows.Next() {
// 		var b []byte
// 		if err := rows.Scan(&b); err != nil {
// 			return nil, err
// 		}

// 		d, err := shared.Decompressor.Decompress(bytes.NewReader(b))
// 		if err != nil {
// 			return nil, err
// 		}

// 		var doc scip.Document
// 		if err := proto.Unmarshal(d, &doc); err != nil {
// 			return nil, err
// 		}

// 		documents = append(documents, &doc)
// 	}

// 	return documents, nil
// }

// const getDocumentsBySymbolNameQuery = `
// SELECT
//     raw_scip_payload
// FROM codeintel_scip_symbols_lookup ssl
// JOIN codeintel_scip_symbols ss ON ss.upload_id = ssl.upload_id AND ss.descriptor_id = ssl.id
// JOIN codeintel_scip_document_lookup sdl ON sdl.id = ss.document_lookup_id
// JOIN codeintel_scip_documents sd ON sd.id = sdl.document_id
// WHERE
//     ssl.name ILIKE ANY(%s)
//     AND ssl.scip_name_type = 'DESCRIPTOR'
//     AND ssl.upload_id = %s;
// `

// func formatSymbolNamesToLikeClause(symbolNames []string) []string {
// 	explodedSymbols := make([]string, 0, len(symbolNames))
// 	for _, symbolName := range symbolNames {
// 		ex := symbols.NewExplodedSymbol(symbolName)
// 		explodedSymbols = append(
// 			explodedSymbols,
// 			"%"+ex.Descriptor+"%",
// 		)
// 	}

// 	return explodedSymbols
// }

/*

context: [
	{

		symbol: "New()."
		repository: "github.com/sourcegraph/sourcegraph",
		type: DEFINITION,
		text: "func New() *Hello {\n\tm := world.New()\n\treturn &Hello{World: m}\n}",
	},
	{
		repository: "github.com/sourcegraph/sourcegraph",
		type: DEFINITION,
		text: "func World",
	},
]

*/
