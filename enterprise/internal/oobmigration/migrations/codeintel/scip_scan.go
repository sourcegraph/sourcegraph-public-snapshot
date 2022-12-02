package codeintel

import (
	"database/sql"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

func makeDocumentScanner(serializer *serializer) func(rows *sql.Rows, queryErr error) (map[string]DocumentData, error) {
	return basestore.NewMapScanner(func(s dbutil.Scanner) (string, DocumentData, error) {
		var path string
		var data MarshalledDocumentData
		if err := s.Scan(&path, &data.Ranges, &data.HoverResults, &data.Monikers, &data.PackageInformation, &data.Diagnostics); err != nil {
			return "", DocumentData{}, err
		}

		document, err := serializer.UnmarshalDocumentData(data)
		if err != nil {
			return "", DocumentData{}, err
		}

		return path, document, nil
	})
}

func scanResultChunksIntoMap(serializer *serializer, f func(idx int, resultChunk ResultChunkData) error) func(rows *sql.Rows, queryErr error) error {
	return basestore.NewCallbackScanner(func(s dbutil.Scanner) error {
		var idx int
		var rawData []byte
		if err := s.Scan(&idx, &rawData); err != nil {
			return err
		}

		data, err := serializer.UnmarshalResultChunkData(rawData)
		if err != nil {
			return err
		}

		return f(idx, data)
	})
}
