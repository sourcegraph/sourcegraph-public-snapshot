package uploads

import "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/internal/shared"

type (
	AvailableIndexer                = shared.AvailableIndexer
	JobsOrHints                     = shared.JobsOrHints
	IndexesWithRepositoryNamespace  = shared.IndexesWithRepositoryNamespace
	RepositoryWithCount             = shared.RepositoryWithCount
	RepositoryWithAvailableIndexers = shared.RepositoryWithAvailableIndexers
	DirtyRepository                 = shared.DirtyRepository
	GetIndexersOptions              = shared.GetIndexersOptions
	GetUploadsOptions               = shared.GetUploadsOptions
	ReindexUploadsOptions           = shared.ReindexUploadsOptions
	DeleteUploadsOptions            = shared.DeleteUploadsOptions
	Package                         = shared.Package
	PackageReference                = shared.PackageReference
	PackageReferenceScanner         = shared.PackageReferenceScanner
	UploadsWithRepositoryNamespace  = shared.UploadsWithRepositoryNamespace
	GetIndexesOptions               = shared.GetIndexesOptions
	DeleteIndexesOptions            = shared.DeleteIndexesOptions
	ReindexIndexesOptions           = shared.ReindexIndexesOptions
	ExportedUpload                  = shared.ExportedUpload
)

var (
	GetKeyForLookup = shared.GetKeyForLookup
	Compressor      = shared.Compressor
	Decompressor    = shared.Decompressor
)

func PopulateInferredAvailableIndexers[J JobsOrHints](jobsOrHints []J, blocklist map[string]struct{}, inferredAvailableIndexers map[string]AvailableIndexer) map[string]AvailableIndexer {
	return shared.PopulateInferredAvailableIndexers(jobsOrHints, blocklist, inferredAvailableIndexers)
}
