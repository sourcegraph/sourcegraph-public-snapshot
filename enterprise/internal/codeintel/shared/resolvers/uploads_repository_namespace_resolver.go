package sharedresolvers

import (
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
	uploadsShared "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/shared"
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
)

type lsifUploadsWithRepositoryNamespaceResolver struct {
	uploadsSummary  uploadsShared.UploadsWithRepositoryNamespace
	uploadResolvers []resolverstubs.LSIFUploadResolver
}

func NewLSIFUploadsWithRepositoryNamespaceResolver(uploadsSummary uploadsShared.UploadsWithRepositoryNamespace, uploadResolvers []resolverstubs.LSIFUploadResolver) resolverstubs.LSIFUploadsWithRepositoryNamespaceResolver {
	return &lsifUploadsWithRepositoryNamespaceResolver{
		uploadsSummary:  uploadsSummary,
		uploadResolvers: uploadResolvers,
	}
}

func (r *lsifUploadsWithRepositoryNamespaceResolver) Root() string {
	return r.uploadsSummary.Root
}

func (r *lsifUploadsWithRepositoryNamespaceResolver) Indexer() resolverstubs.CodeIntelIndexerResolver {
	for _, indexer := range types.AllIndexers {
		if indexer.Name == r.uploadsSummary.Indexer {
			return types.NewCodeIntelIndexerResolverFrom(indexer)
		}
	}

	return types.NewCodeIntelIndexerResolver(r.uploadsSummary.Indexer)
}

func (r *lsifUploadsWithRepositoryNamespaceResolver) Uploads() []resolverstubs.LSIFUploadResolver {
	return r.uploadResolvers
}
