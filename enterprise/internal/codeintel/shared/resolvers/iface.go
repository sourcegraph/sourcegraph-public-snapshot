package sharedresolvers

import (
	"context"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
)

type UploadsService interface {
	GetIndexesByIDs(ctx context.Context, ids ...int) (_ []types.Index, err error)
	GetUploadsByIDs(ctx context.Context, ids ...int) (_ []types.Upload, err error)
}
