package policies

import (
	"context"
)

type UploadService interface {
	GetCommitsVisibleToUpload(ctx context.Context, uploadID, limit int, token *string) (_ []string, nextToken *string, err error)
}
