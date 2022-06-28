package store

import (
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

func scanUpload(s dbutil.Scanner) (upload shared.Upload, err error) {
	return upload, s.Scan(
		&upload.ID,
	)
}

var scanUploads = basestore.NewSliceScanner(scanUpload)
