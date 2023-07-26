package graphql

import (
	"github.com/sourcegraph/sourcegraph/internal/codeintel/shared/resolvers/dataloader"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
)

type (
	UploadLoaderFactory = *dataloader.LoaderFactory[int, shared.Upload]
	IndexLoaderFactory  = *dataloader.LoaderFactory[int, shared.Index]
	UploadLoader        = *dataloader.Loader[int, shared.Upload]
	IndexLoader         = *dataloader.Loader[int, shared.Index]
)

func NewUploadLoaderFactory(uploadService UploadsService) UploadLoaderFactory {
	return dataloader.NewLoaderFactory[int, shared.Upload](dataloader.BackingServiceFunc[int, shared.Upload](uploadService.GetUploadsByIDs))
}

func NewIndexLoaderFactory(uploadService UploadsService) IndexLoaderFactory {
	return dataloader.NewLoaderFactory[int, shared.Index](dataloader.BackingServiceFunc[int, shared.Index](uploadService.GetIndexesByIDs))
}

func PresubmitAssociatedIndexes(indexLoader IndexLoader, uploads ...shared.Upload) {
	for _, upload := range uploads {
		if upload.AssociatedIndexID != nil {
			indexLoader.Presubmit(*upload.AssociatedIndexID)
		}
	}
}

func PresubmitAssociatedUploads(uploadLoader UploadLoader, indexes ...shared.Index) {
	for _, index := range indexes {
		if index.AssociatedUploadID != nil {
			uploadLoader.Presubmit(*index.AssociatedUploadID)
		}
	}
}
