package codenav

import (
	"context"
	"sync"

	genslices "github.com/life4/genesis/slices"
	"github.com/sourcegraph/scip/bindings/go/scip"
	"google.golang.org/protobuf/proto"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/codegraph"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav/internal/lsifstore"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/core"
)

type MappedIndex interface {
	// GetDocument returns None if the index does not contain a document at the given path.
	// There is no caching here, every call to GetDocument re-fetches the full document from the database.
	GetDocument(context.Context, core.RepoRelPath) (core.Option[MappedDocument], error)
	GetUploadSummary() core.UploadSummary
	GetDocuments(context.Context, []core.RepoRelPath) (map[core.RepoRelPath]MappedDocument, error)
	// TODO: Should there be a bulk-API for getting multiple documents?
}

var _ MappedIndex = mappedIndex{}

// MappedDocument wraps a SCIP document from an uploaded index.
// All methods accept and return ranges in the context of the target commit.
type MappedDocument interface {
	// GetOccurrences returns shared slices. Do not modify the returned slice or Occurrences without copying them first
	GetOccurrences(context.Context) ([]*scip.Occurrence, error)
	// GetOccurrencesAtRange returns shared slices. Do not modify the returned slice or Occurrences without copying them first
	GetOccurrencesAtRange(context.Context, scip.Range) ([]*scip.Occurrence, error)
	GetPath() core.RepoRelPath
}

var _ MappedDocument = &mappedDocument{}

// NewMappedIndexFromTranslator creates a MappedIndex using the given GitTreeTranslator.
// The translators SourceCommit has to be the targetCommit, which will be mapped to upload.GetCommit()
func NewMappedIndexFromTranslator(
	lsifStore lsifstore.LsifStore,
	gitTreeTranslator GitTreeTranslator,
	upload core.UploadLike,
	targetCommit api.CommitID,
) MappedIndex {
	return mappedIndex{
		lsifStore:         lsifStore,
		gitTreeTranslator: gitTreeTranslator,
		upload:            upload,
		targetCommit:      targetCommit,
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

func (i mappedIndex) makeMappedDocument(path core.RepoRelPath, scipDocument *scip.Document) MappedDocument {
	return &mappedDocument{
		gitTreeTranslator: i.gitTreeTranslator,
		indexCommit:       i.upload.GetCommit(),
		targetCommit:      i.targetCommit,
		path:              path,
		document: &lockedDocument{
			inner:      scipDocument,
			isMapped:   false,
			mapErrored: nil,
			lock:       sync.RWMutex{},
		},
		mapOnce: sync.Once{},
	}
}

func (i mappedIndex) GetDocuments(ctx context.Context, paths []core.RepoRelPath) (map[core.RepoRelPath]MappedDocument, error) {
	// We're fetching the targetCommit -> indexCommit here, because this is currently used in syntactic usages
	// which calls GetOccurrencesAtRange, which in turn needs to map ranges in that direction.
	//
	// NOTE(id: mapped-index-over-fetching-diffs) We will end up over-fetching diffs here, as it's not guaranteed that we'll get a SCIP document for every
	// path in paths. If we reduce concurrency here and fetch after the SQL query returns we could avoid that.
	i.gitTreeTranslator.Prefetch(ctx, i.targetCommit, i.upload.GetCommit(), paths)
	documentMap, err := i.lsifStore.SCIPDocuments(ctx, i.upload.GetID(), genslices.Map(paths, func(p core.RepoRelPath) core.UploadRelPath {
		return core.NewUploadRelPath(i.upload, p)
	}))
	if err != nil {
		return nil, err
	}
	resultMap := make(map[core.RepoRelPath]MappedDocument, len(documentMap))
	for path, scipDocument := range documentMap {
		repoRelPath := core.NewRepoRelPath(i.upload, path)
		resultMap[repoRelPath] = i.makeMappedDocument(repoRelPath, scipDocument)
	}
	return resultMap, nil
}

func (i mappedIndex) GetDocument(ctx context.Context, path core.RepoRelPath) (core.Option[MappedDocument], error) {
	optDocument, err := i.lsifStore.SCIPDocument(ctx, i.upload.GetID(), core.NewUploadRelPath(i.upload, path))
	if err != nil {
		return core.None[MappedDocument](), err
	}
	if document, ok := optDocument.Get(); ok {
		return core.Some[MappedDocument](i.makeMappedDocument(path, document)), nil
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

func cloneOccurrence(occ *scip.Occurrence) *scip.Occurrence {
	occCopy, ok := proto.Clone(occ).(*scip.Occurrence)
	if !ok {
		panic("impossible! proto.Clone changed the type of message")
	}
	return occCopy
}

func (d *mappedDocument) GetPath() core.RepoRelPath {
	return d.path
}

// Concurrency: Only call this while you're holding a read lock on the document
func (d *mappedDocument) mapAllOccurrences(ctx context.Context) ([]*scip.Occurrence, error) {
	newOccurrences := make([]*scip.Occurrence, 0)
	for _, occ := range d.document.inner.Occurrences {
		mappedRangeOpt, err := d.gitTreeTranslator.TranslateRange(
			ctx, d.indexCommit, d.targetCommit, d.path, scip.NewRangeUnchecked(occ.Range),
		)
		if err != nil {
			return nil, err
		}
		if mappedRange, ok := mappedRangeOpt.Get(); ok {
			newOccurrence := cloneOccurrence(occ)
			newOccurrence.Range = mappedRange.SCIPRange()
			newOccurrences = append(newOccurrences, newOccurrence)
		}
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

	if d.document.mapErrored != nil {
		return nil, d.document.mapErrored
	}
	return d.document.inner.Occurrences, nil
}

func (d *mappedDocument) GetOccurrencesAtRange(ctx context.Context, range_ scip.Range) ([]*scip.Occurrence, error) {
	d.document.lock.RLock()
	occurrences := d.document.inner.Occurrences
	if d.document.isMapped {
		d.document.lock.RUnlock()
		return codegraph.FindOccurrencesWithEqualRange(occurrences, range_), nil
	}
	d.document.lock.RUnlock()

	mappedRgOpt, err := d.gitTreeTranslator.TranslateRange(ctx, d.targetCommit, d.indexCommit, d.path, range_)
	if err != nil {
		return nil, err
	}
	mappedRg, ok := mappedRgOpt.Get()
	if !ok {
		// The range was changed/removed in the target commit, so return no occurrences
		return nil, nil
	}
	pastMatchingOccurrences := codegraph.FindOccurrencesWithEqualRange(occurrences, mappedRg)
	scipRange := range_.SCIPRange()
	return genslices.Map(pastMatchingOccurrences, func(occ *scip.Occurrence) *scip.Occurrence {
		newOccurrence := cloneOccurrence(occ)
		newOccurrence.Range = scipRange
		return newOccurrence
	}), nil
}
