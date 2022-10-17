package sharedresolvers

import (
	"github.com/sourcegraph/sourcegraph/internal/codeintel/shared/types"
	uploadsShared "github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
)

type LSIFUploadsWithRepositoryNamespaceResolver interface {
	Root() string
	Indexer() types.CodeIntelIndexerResolver
	Uploads() []LSIFUploadResolver
}

type lsifUploadsWithRepositoryNamespaceResolver struct {
	uploadsSummary  uploadsShared.UploadsWithRepositoryNamespace
	uploadResolvers []LSIFUploadResolver
}

func NewLSIFUploadsWithRepositoryNamespaceResolver(uploadsSummary uploadsShared.UploadsWithRepositoryNamespace, uploadResolvers []LSIFUploadResolver) LSIFUploadsWithRepositoryNamespaceResolver {
	return &lsifUploadsWithRepositoryNamespaceResolver{
		uploadsSummary:  uploadsSummary,
		uploadResolvers: uploadResolvers,
	}
}

func (r *lsifUploadsWithRepositoryNamespaceResolver) Root() string {
	return r.uploadsSummary.Root
}

func (r *lsifUploadsWithRepositoryNamespaceResolver) Indexer() types.CodeIntelIndexerResolver {
	for _, indexer := range types.AllIndexers {
		if indexer.Name == r.uploadsSummary.Indexer {
			return types.NewCodeIntelIndexerResolverFrom(indexer)
		}
	}

	return types.NewCodeIntelIndexerResolver(r.uploadsSummary.Indexer)
}

func (r *lsifUploadsWithRepositoryNamespaceResolver) Uploads() []LSIFUploadResolver {
	return r.uploadResolvers
}
