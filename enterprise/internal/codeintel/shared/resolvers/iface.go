package sharedresolvers

import (
	"context"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/shared"
	uploadsshared "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/shared"
)

type UploadsService interface {
	GetIndexesByIDs(ctx context.Context, ids ...int) (_ []uploadsshared.Index, err error)
	GetUploadsByIDs(ctx context.Context, ids ...int) (_ []shared.Upload, err error)
}
