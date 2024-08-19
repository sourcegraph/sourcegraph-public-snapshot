package lsifstore

import (
	"bytes"
	"context"

	"github.com/keegancsmith/sqlf"
	genslices "github.com/life4/genesis/slices"
	"github.com/sourcegraph/scip/bindings/go/scip"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/protobuf/proto"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/core"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func (s *store) SCIPDocument(ctx context.Context, uploadID int, path core.UploadRelPath) (_ core.Option[*scip.Document], err error) {
	ctx, _, endObservation := s.operations.scipDocument.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.String("path", path.RawValue()),
		attribute.Int("uploadID", uploadID),
	}})
	defer endObservation(1, observation.Args{})

	documents, err := s.SCIPDocuments(ctx, uploadID, []core.UploadRelPath{path})
	if err != nil {
		return core.None[*scip.Document](), err
	}
	if doc, ok := documents[path]; ok {
		return core.Some(doc), nil
	} else {
		return core.None[*scip.Document](), nil
	}
}

func (s *store) SCIPDocuments(ctx context.Context, uploadID int, paths []core.UploadRelPath) (_ map[core.UploadRelPath]*scip.Document, err error) {
	stringPaths := genslices.Map(paths, func(p core.UploadRelPath) string { return p.RawValue() })
	ctx, _, endObservation := s.operations.scipDocuments.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Int("uploadID", uploadID),
		attribute.StringSlice("paths", stringPaths),
	}})
	defer endObservation(1, observation.Args{})

	scanner := basestore.NewMapScanner(func(dbs dbutil.Scanner) (core.UploadRelPath, *scip.Document, error) {
		var compressedSCIPPayload []byte
		var path string
		emptyPath := core.NewUploadRelPathUnchecked("")
		if err := dbs.Scan(&path, &compressedSCIPPayload); err != nil {
			return emptyPath, nil, err
		}

		scipPayload, err := shared.Decompressor.Decompress(bytes.NewReader(compressedSCIPPayload))
		if err != nil {
			return emptyPath, nil, err
		}

		var document scip.Document
		if err := proto.Unmarshal(scipPayload, &document); err != nil {
			return emptyPath, nil, err
		}
		return core.NewUploadRelPathUnchecked(path), &document, nil
	})
	searchPaths := make([]*sqlf.Query, 0, len(paths))
	for _, path := range stringPaths {
		searchPaths = append(searchPaths, sqlf.Sprintf("%s", path))
	}
	docs, err := scanner(s.db.Query(ctx, sqlf.Sprintf(fetchSCIPDocumentsQuery, uploadID, sqlf.Join(searchPaths, ","))))
	if err != nil {
		return nil, err
	}
	return docs, nil
}

const fetchSCIPDocumentsQuery = `
SELECT
	sid.document_path,
	sd.raw_scip_payload
FROM codeintel_scip_document_lookup sid
JOIN codeintel_scip_documents sd ON sd.id = sid.document_id
WHERE
	sid.upload_id = %s AND
	sid.document_path IN (%s)
`

func (s *store) FindDocumentIDs(ctx context.Context, uploadIDToLookupPath map[int]core.UploadRelPath) (uploadIDToDocumentID map[int]int, err error) {
	ctx, _, endObservation := s.operations.findDocumentIDs.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.Int("numUploadIDs", len(uploadIDToLookupPath)),
	}})
	defer endObservation(1, observation.Args{})

	if len(uploadIDToLookupPath) == 0 {
		return nil, nil
	}

	searchTuples := []*sqlf.Query{}
	for uploadID, path := range uploadIDToLookupPath {
		searchTuples = append(searchTuples, sqlf.Sprintf("(%d, %s)", uploadID, path.RawValue()))
	}

	finalQuery := sqlf.Sprintf(findDocumentIDsQuery, sqlf.Join(searchTuples, ","))

	scanner := basestore.NewMapScanner(func(dbs dbutil.Scanner) (uploadID int, documentID int, err error) {
		err = dbs.Scan(&uploadID, &documentID)
		return uploadID, documentID, err
	})
	return scanner(s.db.Query(ctx, finalQuery))
}

const findDocumentIDsQuery = `
SELECT sid.upload_id, sid.document_id
FROM codeintel_scip_document_lookup sid
WHERE (sid.upload_id, sid.document_path) IN (%s)
`
