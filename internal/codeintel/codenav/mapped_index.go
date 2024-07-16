package codenav

import (
	"context"
	"sync"
	"sync/atomic"

	genslices "github.com/life4/genesis/slices"
	"github.com/sourcegraph/scip/bindings/go/scip"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav/internal/lsifstore"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav/shared"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/core"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type Mapped interface {
	IndexCommit() api.CommitID
	TargetCommit() api.CommitID
}

// MappedIndex wraps an uploaded SCIP index and a target commit and creates MappedDocument instances,
// which automatically map occurrence ranges across the target and index commit.
type MappedIndex interface {
	GetDocument(context.Context, core.RepoRelPath) (core.Option[MappedDocument], error)
	GetUploadSummary() core.UploadSummary
	// TODO: Should there be a bulk-API for getting multiple documents?
	Mapped
}

var _ MappedIndex = mappedIndex{}

// MappedDocument wraps a SCIP document from an uploaded index.
// All methods accept and return ranges in the context of the target commit.
type MappedDocument interface {
	GetOccurrences(context.Context) ([]*scip.Occurrence, error)
	GetOccurrencesAtRange(context.Context, scip.Range) ([]*scip.Occurrence, error)
	Mapped
}

var _ MappedDocument = &mappedDocument{}

// NewMappedIndex creates a MappedIndex for an uploaded index and a targetCommit
func NewMappedIndex(
	lsifStore lsifstore.LsifStore,
	repo *types.Repo,
	gitserverClient gitserver.Client,
	upload core.UploadLike,
	targetCommit api.CommitID,
) (MappedIndex, error) {
	// NOTE: No idea if 100 is a reasonable number here (the resolver wide one has a default of 1000),
	// I'll get rid of the LRU cache once I get the unified diff command output from gitserver
	hunkCache, err := NewHunkCache(100)
	if err != nil {
		return nil, err
	}
	gitTreeTranslator := NewGitTreeTranslator(gitserverClient, &TranslationBase{
		Repo:   repo,
		Commit: targetCommit,
	}, hunkCache)
	return mappedIndex{
		lsifStore:         lsifStore,
		gitTreeTranslator: gitTreeTranslator,
		upload:            upload,
		targetCommit:      targetCommit,
	}, nil
}

// NewMappedIndexFromTranslator creates a MappedIndex using the given GitTreeTranslator.
// The translators SourceCommit has to be the targetCommit, which will be mapped to upload.GetCommit()
func NewMappedIndexFromTranslator(
	lsifStore lsifstore.LsifStore,
	gitTreeTranslator GitTreeTranslator,
	upload core.UploadLike,
) MappedIndex {
	return mappedIndex{
		lsifStore:         lsifStore,
		gitTreeTranslator: gitTreeTranslator,
		upload:            upload,
		targetCommit:      gitTreeTranslator.GetSourceCommit(),
	}
}

type mappedIndex struct {
	lsifStore         lsifstore.LsifStore
	gitTreeTranslator GitTreeTranslator
	upload            core.UploadLike
	targetCommit      api.CommitID
}

func (i mappedIndex) IndexCommit() api.CommitID {
	return i.upload.GetCommit()
}

func (i mappedIndex) TargetCommit() api.CommitID {
	return i.targetCommit
}

func (i mappedIndex) GetUploadSummary() core.UploadSummary {
	return core.UploadSummary{
		ID:     i.upload.GetID(),
		Root:   i.upload.GetRoot(),
		Commit: i.upload.GetCommit(),
	}
}

func (i mappedIndex) GetDocument(ctx context.Context, path core.RepoRelPath) (core.Option[MappedDocument], error) {
	optDocument, err := i.lsifStore.SCIPDocument(ctx, i.upload.GetID(), core.NewUploadRelPath(i.upload, path))
	if err != nil {
		return core.None[MappedDocument](), err
	}
	// TODO: Should we cache the mapped document? The current usages don't request the same document twice
	// so we'd just be increasing resident memory
	if document, ok := optDocument.Get(); ok {
		return core.Some[MappedDocument](&mappedDocument{
			gitTreeTranslator: i.gitTreeTranslator,
			indexCommit:       i.upload.GetCommit(),
			targetCommit:      i.targetCommit,
			path:              path,
			document:          document,
			mapOnce:           sync.Once{},
		}), nil
	} else {
		return core.None[MappedDocument](), nil
	}
}

type mappedDocument struct {
	gitTreeTranslator GitTreeTranslator
	indexCommit       api.CommitID
	targetCommit      api.CommitID
	path              core.RepoRelPath
	document          *scip.Document

	mapOnce    sync.Once
	mapErrored error
	isMapped   atomic.Bool
}

func (d *mappedDocument) IndexCommit() api.CommitID {
	return d.indexCommit
}

func (d *mappedDocument) TargetCommit() api.CommitID {
	return d.targetCommit
}

func (d *mappedDocument) mapAllOccurrences(ctx context.Context) ([]*scip.Occurrence, error) {
	newOccurrences := make([]*scip.Occurrence, 0)
	for _, occ := range d.document.Occurrences {
		scipRange := scip.NewRangeUnchecked(occ.Range)
		sharedRange := shared.TranslateRange(scipRange)
		mappedRange, ok, err := d.gitTreeTranslator.GetTargetCommitRangeFromSourceRange(ctx, string(d.indexCommit), d.path.RawValue(), sharedRange, true)
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}
		newOccurrences = append(newOccurrences, &scip.Occurrence{
			Range:                 mappedRange.ToSCIPRange().SCIPRange(),
			Symbol:                occ.Symbol,
			SymbolRoles:           occ.SymbolRoles,
			OverrideDocumentation: occ.OverrideDocumentation,
			SyntaxKind:            occ.SyntaxKind,
			Diagnostics:           occ.Diagnostics,
			EnclosingRange:        occ.EnclosingRange,
		})
	}
	return newOccurrences, nil
}

func (d *mappedDocument) GetOccurrences(ctx context.Context) ([]*scip.Occurrence, error) {
	// It's important we don't remap the occurrences twice
	d.mapOnce.Do(func() {
		newOccurrences, err := d.mapAllOccurrences(ctx)
		if err != nil {
			d.mapErrored = err
			return
		}
		d.document.Occurrences = newOccurrences
		d.isMapped.Store(true)
	})

	if d.mapErrored != nil {
		return nil, d.mapErrored
	}
	return d.document.Occurrences, nil
}

func (d *mappedDocument) GetOccurrencesAtRange(ctx context.Context, rg scip.Range) ([]*scip.Occurrence, error) {
	pastOccurrences := d.document.Occurrences
	if d.isMapped.Load() {
		// if isMapped is true we know d.document.Occurrences is mapped,
		// we can _not_ use pastOccurrences here
		// (in case mapping finished between taking that reference and checking isMapped)
		return FindOccurrencesWithEqualRange(d.document.Occurrences, rg), nil
	}
	// We know pastOccurrences is not mapped at this point, and `GetOccurrence` will not modify it
	mappedRg, ok, err := d.gitTreeTranslator.GetTargetCommitRangeFromSourceRange(
		ctx, string(d.indexCommit), d.path.RawValue(), shared.TranslateRange(rg), false,
	)
	if err != nil {
		return nil, err
	}
	if !ok {
		// The range was changed/removed in the target commit, so return no occurrences
		return nil, nil
	}
	pastMatchingOccurrences := FindOccurrencesWithEqualRange(pastOccurrences, mappedRg.ToSCIPRange())
	Range := rg.SCIPRange()
	return genslices.Map(pastMatchingOccurrences, func(occ *scip.Occurrence) *scip.Occurrence {
		return &scip.Occurrence{
			// Return the range in the target commit, instead of the index commit
			Range:                 Range,
			Symbol:                occ.Symbol,
			SymbolRoles:           occ.SymbolRoles,
			OverrideDocumentation: occ.OverrideDocumentation,
			SyntaxKind:            occ.SyntaxKind,
			Diagnostics:           occ.Diagnostics,
			EnclosingRange:        occ.EnclosingRange,
		}
	}), nil
}
