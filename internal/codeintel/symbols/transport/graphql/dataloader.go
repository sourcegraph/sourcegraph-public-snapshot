package graphql

import (
	"sync"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/symbols/shared"
)

type UploadsDataLoader struct {
	uploads     []shared.Dump
	uploadsByID map[int]shared.Dump
	cacheMutex  sync.RWMutex
}

func NewUploadsDataLoader() *UploadsDataLoader {
	return &UploadsDataLoader{
		uploadsByID: make(map[int]shared.Dump),
	}
}

func (l *UploadsDataLoader) getUploadFromCacheMap(id int) (shared.Dump, bool) {
	l.cacheMutex.RLock()
	defer l.cacheMutex.RUnlock()

	upload, ok := l.uploadsByID[id]
	return upload, ok
}

func (l *UploadsDataLoader) setUploadInCacheMap(uploads []shared.Dump) {
	l.cacheMutex.Lock()
	defer l.cacheMutex.Unlock()

	for i := range uploads {
		l.uploadsByID[uploads[i].ID] = uploads[i]
	}
}

func (l *UploadsDataLoader) AddUpload(d dbstore.Dump) {
	l.cacheMutex.Lock()
	defer l.cacheMutex.Unlock()

	dump := shared.Dump{
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
	}
	l.uploads = append(l.uploads, dump)
	l.uploadsByID[dump.ID] = dump
}
