package symbols

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/symbols/internal/lsifstore"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/symbols/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/symbols/shared"
	uploads "github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type service interface {
	GetHover(ctx context.Context, bundleID int, path string, line, character int) (string, shared.Range, bool, error)
	GetReferences(ctx context.Context, uploadID int, path string, line int, character int, limit int, offset int) (_ []shared.Location, totalCount int, err error)
	GetImplementations(ctx context.Context, uploadID int, path string, line int, character int, limit int, offset int) (_ []shared.Location, totalCount int, err error)
	GetDefinitions(ctx context.Context, uploadID int, path string, line int, character int, limit int, offset int) (_ []shared.Location, totalCount int, err error)
	GetDiagnostics(ctx context.Context, bundleID int, prefix string, limit, offset int) (_ []shared.Diagnostic, _ int, err error)
	GetRanges(ctx context.Context, bundleID int, path string, startLine, endLine int) (_ []shared.CodeIntelligenceRange, err error)

	GetMonikersByPosition(ctx context.Context, bundleID int, path string, line, character int) (_ [][]precise.MonikerData, err error)
	GetBulkMonikerLocations(ctx context.Context, tableName string, uploadIDs []int, monikers []precise.MonikerData, limit, offset int) (_ []shared.Location, _ int, err error)
	GetPackageInformation(ctx context.Context, bundleID int, path, packageInformationID string) (_ precise.PackageInformationData, _ bool, err error)

	// Uploads Service
	GetDumpsByIDs(ctx context.Context, ids []int) (_ []shared.Dump, err error)
	GetUploadsWithDefinitionsForMonikers(ctx context.Context, monikers []precise.QualifiedMonikerData) (_ []shared.Dump, err error)
	GetUploadIDsWithReferences(ctx context.Context, orderedMonikers []precise.QualifiedMonikerData, ignoreIDs []int, repositoryID int, commit string, limit int, offset int) (ids []int, recordsScanned int, totalCount int, err error)
}

type Service struct {
	store      store.Store
	lsifstore  lsifstore.LsifStore
	uploadSvc  UploadService
	operations *operations
}

func newService(store store.Store, lsifstore lsifstore.LsifStore, uploadSvc UploadService, observationContext *observation.Context) *Service {
	return &Service{
		store:      store,
		lsifstore:  lsifstore,
		uploadSvc:  uploadSvc,
		operations: newOperations(observationContext),
	}
}

type Symbol = shared.Symbol

type SymbolOpts struct{}

func (s *Service) Symbol(ctx context.Context, opts SymbolOpts) (symbols []Symbol, err error) {
	ctx, _, endObservation := s.operations.symbol.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	// To be implemented in https://github.com/sourcegraph/sourcegraph/issues/33374
	_, _ = s.store.List(ctx, store.ListOpts{})
	return nil, errors.Newf("unimplemented: symbols.Symbol")
}

// GetHover returns the set of locations defining the symbol at the given position.
func (s *Service) GetHover(ctx context.Context, uploadID int, path string, line int, character int) (_ string, _ shared.Range, _ bool, err error) {
	ctx, _, endObservation := s.operations.getImplementations.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return s.lsifstore.GetHover(ctx, uploadID, path, line, character)
}

// GetReferences returns the list of source locations that reference the symbol at the given position.
func (s *Service) GetReferences(ctx context.Context, uploadID int, path string, line int, character int, limit int, offset int) (_ []shared.Location, totalCount int, err error) {
	ctx, _, endObservation := s.operations.getReferences.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return s.lsifstore.GetReferenceLocations(ctx, uploadID, path, line, character, limit, offset)
}

// GetImplementations returns the set of locations implementing the symbol at the given position.
func (s *Service) GetImplementations(ctx context.Context, uploadID int, path string, line int, character int, limit int, offset int) (_ []shared.Location, totalCount int, err error) {
	ctx, _, endObservation := s.operations.getImplementations.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return s.lsifstore.GetImplementationLocations(ctx, uploadID, path, line, character, limit, offset)
}

// GetDefinitions returns the set of locations defining the symbol at the given position.
func (s *Service) GetDefinitions(ctx context.Context, uploadID int, path string, line int, character int, limit int, offset int) (_ []shared.Location, totalCount int, err error) {
	ctx, _, endObservation := s.operations.getImplementations.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return s.lsifstore.GetDefinitionLocations(ctx, uploadID, path, line, character, limit, offset)
}

func (s *Service) GetDiagnostics(ctx context.Context, bundleID int, prefix string, limit, offset int) (_ []shared.Diagnostic, _ int, err error) {
	ctx, _, endObservation := s.operations.getImplementations.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return s.lsifstore.GetDiagnostics(ctx, bundleID, prefix, limit, offset)
}

func (s *Service) GetRanges(ctx context.Context, bundleID int, path string, startLine, endLine int) (_ []shared.CodeIntelligenceRange, err error) {
	ctx, _, endObservation := s.operations.getImplementations.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return s.lsifstore.GetRanges(ctx, bundleID, path, startLine, endLine)
}

// GetStencil returns the set of locations defining the symbol at the given position.
func (s *Service) GetStencil(ctx context.Context, uploadID int, path string) (_ []shared.Range, err error) {
	ctx, _, endObservation := s.operations.getImplementations.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return s.lsifstore.GetStencil(ctx, uploadID, path)
}

func (s *Service) GetMonikersByPosition(ctx context.Context, bundleID int, path string, line, character int) (_ [][]precise.MonikerData, err error) {
	ctx, _, endObservation := s.operations.getMonikersByPosition.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return s.lsifstore.GetMonikersByPosition(ctx, bundleID, path, line, character)
}

func (s *Service) GetBulkMonikerLocations(ctx context.Context, tableName string, uploadIDs []int, monikers []precise.MonikerData, limit, offset int) (_ []shared.Location, _ int, err error) {
	ctx, _, endObservation := s.operations.getBulkMonikerLocations.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return s.lsifstore.GetBulkMonikerLocations(ctx, tableName, uploadIDs, monikers, limit, offset)
}

func (s *Service) GetUploadsWithDefinitionsForMonikers(ctx context.Context, monikers []precise.QualifiedMonikerData) (_ []shared.Dump, err error) {
	ctx, _, endObservation := s.operations.getUploadsWithDefinitionsForMonikers.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	uploadDumps, err := s.uploadSvc.GetDumpsWithDefinitionsForMonikers(ctx, monikers)
	if err != nil {
		return nil, err
	}
	dumps := updateSvcDumpToSharedDump(uploadDumps)

	return dumps, nil
}

func (s *Service) GetDumpsByIDs(ctx context.Context, ids []int) (_ []shared.Dump, err error) {
	ctx, _, endObservation := s.operations.getDumpsByIDs.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	uploadDumps, err := s.uploadSvc.GetDumpsByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	dumps := updateSvcDumpToSharedDump(uploadDumps)

	return dumps, nil
}

func (s *Service) GetUploadIDsWithReferences(
	ctx context.Context,
	orderedMonikers []precise.QualifiedMonikerData,
	ignoreIDs []int,
	repositoryID int,
	commit string,
	limit int,
	offset int,
) (ids []int, recordsScanned int, totalCount int, err error) {
	ctx, _, endObservation := s.operations.getUploadIDsWithReferences.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return s.uploadSvc.GetUploadIDsWithReferences(ctx, orderedMonikers, ignoreIDs, repositoryID, commit, limit, offset)
}

func (s *Service) GetPackageInformation(ctx context.Context, bundleID int, path, packageInformationID string) (_ precise.PackageInformationData, _ bool, err error) {
	ctx, _, endObservation := s.operations.getPackageInformation.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return s.lsifstore.GetPackageInformation(ctx, bundleID, path, packageInformationID)
}

func updateSvcDumpToSharedDump(uploadDumps []uploads.Dump) []shared.Dump {
	dumps := make([]shared.Dump, 0, len(uploadDumps))
	for _, d := range uploadDumps {
		dumps = append(dumps, shared.Dump{
			ID:                d.ID,
			Commit:            d.Commit,
			Root:              d.Root,
			VisibleAtTip:      d.VisibleAtTip,
			UploadedAt:        d.UploadedAt,
			State:             d.State,
			FailureMessage:    d.FailureMessage,
			StartedAt:         d.StartedAt,
			FinishedAt:        d.FinishedAt,
			ProcessAfter:      d.ProcessAfter,
			NumResets:         d.NumResets,
			NumFailures:       d.NumFailures,
			RepositoryID:      d.RepositoryID,
			RepositoryName:    d.RepositoryName,
			Indexer:           d.Indexer,
			IndexerVersion:    d.IndexerVersion,
			AssociatedIndexID: d.AssociatedIndexID,
		})
	}
	return dumps
}
