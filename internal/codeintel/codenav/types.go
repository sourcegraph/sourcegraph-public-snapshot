package codenav

import (
	"strings"
	"sync"

	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav/shared"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
)

type RequestState struct {
	dataLoader                     *UploadsDataLoader
	GitTreeTranslator              GitTreeTranslator
	maximumIndexesPerMonikerSearch int
	commitCache                    CommitCache
	authChecker                    authz.SubRepoPermissionChecker
}

func (r *RequestState) GetCacheUploads() []shared.Dump {
	return r.dataLoader.uploads
}

func (r *RequestState) SetAuthChecker(authChecker authz.SubRepoPermissionChecker) {
	r.authChecker = authChecker
}

func (r *RequestState) SetUploadsDataLoader(uploads []dbstore.Dump) {
	r.dataLoader = NewUploadsDataLoader()
	for _, upload := range uploads {
		r.dataLoader.AddUpload(upload)
	}
}

func (r *RequestState) SetLocalGitTreeTranslator(client gitserver.Client, repo *types.Repo, commit, path string, hunkCacheSize int) error {
	hunkCache, err := NewHunkCache(hunkCacheSize)
	if err != nil {
		return err
	}

	args := &requestArgs{
		repo:   repo,
		commit: commit,
		path:   path,
	}

	r.GitTreeTranslator = NewGitTreeTranslator(client, args, hunkCache)

	return nil
}

func (r *RequestState) SetLocalCommitCache(client shared.GitserverClient) {
	r.commitCache = newCommitCache(client)
}

func (r *RequestState) SetMaximumIndexesPerMonikerSearch(maxNumber int) {
	r.maximumIndexesPerMonikerSearch = maxNumber
}

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

func (l *UploadsDataLoader) GetUploadFromCacheMap(id int) (shared.Dump, bool) {
	l.cacheMutex.RLock()
	defer l.cacheMutex.RUnlock()

	upload, ok := l.uploadsByID[id]
	return upload, ok
}

func (l *UploadsDataLoader) SetUploadInCacheMap(uploads []shared.Dump) {
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

// visibleUpload pairs an upload visible from the current target commit with the
// current target path and position matched to the data within the underlying index.
type visibleUpload struct {
	Upload                shared.Dump
	TargetPath            string
	TargetPosition        shared.Position
	TargetPathWithoutRoot string
}

type qualifiedMonikerSet struct {
	monikers       []precise.QualifiedMonikerData
	monikerHashMap map[string]struct{}
}

func newQualifiedMonikerSet() *qualifiedMonikerSet {
	return &qualifiedMonikerSet{
		monikerHashMap: map[string]struct{}{},
	}
}

// add the given qualified moniker to the set if it is distinct from all elements
// currently in the set.
func (s *qualifiedMonikerSet) add(qualifiedMoniker precise.QualifiedMonikerData) {
	monikerHash := strings.Join([]string{
		qualifiedMoniker.PackageInformationData.Name,
		qualifiedMoniker.PackageInformationData.Version,
		qualifiedMoniker.MonikerData.Scheme,
		qualifiedMoniker.MonikerData.Identifier,
	}, ":")

	if _, ok := s.monikerHashMap[monikerHash]; ok {
		return
	}

	s.monikerHashMap[monikerHash] = struct{}{}
	s.monikers = append(s.monikers, qualifiedMoniker)
}
