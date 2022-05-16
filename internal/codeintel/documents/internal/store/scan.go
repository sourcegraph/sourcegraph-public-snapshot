package store

import (
	"github.com/sourcegraph/sourcegraph/internal/codeintel/documents/shared"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

func scanDocument(s dbutil.Scanner) (document shared.Document, err error) {
	return document, s.Scan(
		&document.Path,
	)
}

var scanDocuments = basestore.NewSliceScanner(scanDocument)
