package graphql

import (
	"github.com/sourcegraph/sourcegraph/internal/codeintel/shared/resolvers/dataloader"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
)

type (
	UploadLoaderFactory = *dataloader.LoaderFactory[int, shared.Upload]
	IndexLoaderFactory  = *dataloader.LoaderFactory[int, shared.AutoIndexJob]
	UploadLoader        = *dataloader.Loader[int, shared.Upload]
	IndexLoader         = *dataloader.Loader[int, shared.AutoIndexJob]
)

func NewUploadLoaderFactory(uploadService UploadsService) UploadLoaderFactory {
	return dataloader.NewLoaderFactory[int, shared.Upload](dataloader.BackingServiceFunc[int, shared.Upload](uploadService.GetUploadsByIDs))
}

func NewIndexLoaderFactory(uploadService UploadsService) IndexLoaderFactory {
	return dataloader.NewLoaderFactory[int, shared.AutoIndexJob](dataloader.BackingServiceFunc[int, shared.AutoIndexJob](uploadService.GetIndexesByIDs))
}

func PresubmitAssociatedIndexes(indexLoader IndexLoader, uploads ...shared.Upload) {
	for _, upload := range uploads {
		if upload.AssociatedIndexID != nil {
			indexLoader.Presubmit(*upload.AssociatedIndexID)
		}
	}
}

func PresubmitAssociatedUploads(uploadLoader UploadLoader, indexes ...shared.AutoIndexJob) {
	for _, index := range indexes {
		if index.AssociatedUploadID != nil {
			uploadLoader.Presubmit(*index.AssociatedUploadID)
		}
	}
}
