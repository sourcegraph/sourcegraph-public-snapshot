package sharedresolvers

import (
	"context"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/shared"
)

type UploadsService interface {
	GetIndexesByIDs(ctx context.Context, ids ...int) (_ []types.Index, err error)
	GetUploadsByIDs(ctx context.Context, ids ...int) (_ []shared.Upload, err error)
}
