package lsifstore

import (
	"bytes"
	"context"
	"sort"
	"strings"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	"github.com/sourcegraph/scip/bindings/go/scip"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/protobuf/proto"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/shared/symbols"
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

func (s *store) GetFullSCIPNameByDescriptor(ctx context.Context, uploadID []int, symbolNames []string) (names []*symbols.ExplodedSymbol, err error) {
	ctx, _, endObservation := s.operations.getFullSCIPNameByDescriptor.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{}})
	defer endObservation(1, observation.Args{})

	symbolNamesIlike, err := formatSymbolNamesToLikeClause(symbolNames)
	if err != nil {
		return nil, err
	}

	return scanExplodedSymbols(s.db.Query(ctx, sqlf.Sprintf(
		getFullSCIPNameByDescriptorQuery,
		pq.Array(uploadID),
		pq.Array(symbolNamesIlike),
	)))
}

const getFullSCIPNameByDescriptorQuery = `
SELECT DISTINCT
    l1.name AS scheme,
    l2.name AS package_manager,
    l3.name AS package_name,
    l4.name AS package_version,
    l5.name AS descriptor
-- Initially fuzzy search (see WHERE clause)
FROM codeintel_scip_symbols_lookup l6

-- Join to symbols table, which will bridge DESCRIPTOR_NO_SUFFIX (syntect) and DESCRIPTOR (precise)
JOIN codeintel_scip_symbols_lookup_leaves ll ON ll.upload_id = l6.upload_id AND ll.descriptor_no_suffix_id = l6.id

-- Follow parent path from descriptor l5->l4->l3->l2->l1
JOIN codeintel_scip_symbols_lookup l5 ON l5.upload_id = l6.upload_id AND l5.id = ll.descriptor_id
JOIN codeintel_scip_symbols_lookup l4 ON l4.upload_id = l6.upload_id AND l4.id = l5.parent_id
JOIN codeintel_scip_symbols_lookup l3 ON l3.upload_id = l6.upload_id AND l3.id = l4.parent_id
JOIN codeintel_scip_symbols_lookup l2 ON l2.upload_id = l6.upload_id AND l2.id = l3.parent_id
JOIN codeintel_scip_symbols_lookup l1 ON l1.upload_id = l6.upload_id AND l1.id = l2.parent_id
WHERE
	l6.upload_id = ANY(%s) AND
	l6.scip_name_type = 'DESCRIPTOR_NO_SUFFIX' AND
	l6.name ILIKE ANY(%s)
`

var scanExplodedSymbols = basestore.NewSliceScanner(func(s dbutil.Scanner) (*symbols.ExplodedSymbol, error) {
	var n symbols.ExplodedSymbol
	err := s.Scan(&n.Scheme, &n.PackageManager, &n.PackageName, &n.PackageVersion, &n.Descriptor)
	return &n, err
})

func formatSymbolNamesToLikeClause(symbolNames []string) ([]string, error) {
	trimmedDescriptorMap := make(map[string]struct{}, len(symbolNames))
	for _, symbolName := range symbolNames {
		ex, err := symbols.NewExplodedSymbol(symbolName)
		if err != nil {
			continue
		}

		// TODO: BIT OF A HACK
		parts := strings.Split(ex.Descriptor, "/")
		trimmedDescriptorMap[parts[len(parts)-1]] = struct{}{}
	}

	descriptorWildcards := make([]string, 0, len(trimmedDescriptorMap))
	for symbol := range trimmedDescriptorMap {
		if symbol != "" {
			descriptorWildcards = append(descriptorWildcards, "%"+symbol)
		}
	}
	sort.Strings(descriptorWildcards)

	return descriptorWildcards, nil
}
