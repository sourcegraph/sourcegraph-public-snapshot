package codenav

import (
	"context"
	"sync"

	genslices "github.com/life4/genesis/slices"
	"github.com/sourcegraph/scip/bindings/go/scip"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav/internal/lsifstore"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav/shared"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/core"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type MappedIndex interface {
	GetDocument(context.Context, core.RepoRelPath) (core.Option[MappedDocument], error)
	GetUploadSummary() core.UploadSummary
	// TODO: Should there be a bulk-API for getting multiple documents?
}

var _ MappedIndex = mappedIndex{}

// MappedDocument wraps a SCIP document from an uploaded index.
// All methods accept and return ranges in the context of the target commit.
type MappedDocument interface {
	GetOccurrences(context.Context) ([]*scip.Occurrence, error)
	GetOccurrencesAtRange(context.Context, scip.Range) ([]*scip.Occurrence, error)
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
	// NOTE(issue: GRAPH-742): No idea if 100 is a reasonable number here (the resolver wide one has a default of 1000),
	// This will go away once the linked issue is fixed.
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
			document: &lockedDocument{
				inner:      document,
				isMapped:   false,
				mapErrored: nil,
				lock:       sync.RWMutex{},
			},
			mapOnce: sync.Once{},
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

	document *lockedDocument
	mapOnce  sync.Once
}

type lockedDocument struct {
	inner      *scip.Document
	isMapped   bool
	mapErrored error
	lock       sync.RWMutex
}

// Concurrency: Only call this while you're holding a read lock on the document
func (d *mappedDocument) mapAllOccurrences(ctx context.Context) ([]*scip.Occurrence, error) {
	newOccurrences := make([]*scip.Occurrence, 0)
	for _, occ := range d.document.inner.Occurrences {
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
		d.document.lock.RLock()
		newOccurrences, err := d.mapAllOccurrences(ctx)
		d.document.lock.RUnlock()

		d.document.lock.Lock()
		defer d.document.lock.Unlock()
		if err != nil {
			d.document.mapErrored = err
			return
		}
		d.document.inner.Occurrences = newOccurrences
		d.document.isMapped = true
	})

	d.document.lock.RLock()
	defer d.document.lock.RUnlock()
	if d.document.mapErrored != nil {
		return nil, d.document.mapErrored
	}
	return d.document.inner.Occurrences, nil
}

func (d *mappedDocument) GetOccurrencesAtRange(ctx context.Context, rg scip.Range) ([]*scip.Occurrence, error) {
	d.document.lock.RLock()
	defer d.document.lock.RUnlock()
	if d.document.isMapped {
		return FindOccurrencesWithEqualRange(d.document.inner.Occurrences, rg), nil
	}
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
	pastMatchingOccurrences := FindOccurrencesWithEqualRange(d.document.inner.Occurrences, mappedRg.ToSCIPRange())
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
