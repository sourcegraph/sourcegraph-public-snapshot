package policies

import (
	"context"

	policies "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/policies/enterprise"
)

type UploadService interface {
	GetCommitsVisibleToUpload(ctx context.Context, uploadID, limit int, token *string) (_ []string, nextToken *string, err error)
}

type GitserverClient = policies.GitserverClient
