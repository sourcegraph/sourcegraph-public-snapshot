package codenav

import (
	"sync"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	sgTypes "github.com/sourcegraph/sourcegraph/internal/types"
)

type RequestState struct {
	// Local Caches
	dataLoader        *UploadsDataLoader
	GitTreeTranslator GitTreeTranslator
	commitCache       CommitCache
	// maximumIndexesPerMonikerSearch configures the maximum number of reference upload identifiers
	// that can be passed to a single moniker search query. Previously this limit was meant to keep
	// the number of SQLite files we'd have to open within a single call relatively low. Since we've
	// migrated to Postgres this limit is not a concern. Now we only want to limit these values
	// based on the number of elements we can pass to an IN () clause in the codeintel-db, as well
	// as the size required to encode them in a user-facing pagination cursor.
	maximumIndexesPerMonikerSearch int

	authChecker authz.SubRepoPermissionChecker

	RepositoryID int
	Commit       string
	Path         string
}

func NewRequestState(
	uploads []types.Dump,
	authChecker authz.SubRepoPermissionChecker,
	gitclient GitserverClient,
	repo *sgTypes.Repo,
	commit string,
	path string,
	maxIndexes int,
	hunkCacheSize int,
) RequestState {
	r := &RequestState{
		RepositoryID: int(repo.ID),
		Commit:       commit,
		Path:         path,
	}
	r.SetUploadsDataLoader(uploads)
	r.SetAuthChecker(authChecker)
	r.SetLocalGitTreeTranslator(gitclient, repo, commit, path, hunkCacheSize)
	r.SetLocalCommitCache(gitclient)
	r.SetMaximumIndexesPerMonikerSearch(maxIndexes)

	return *r
}

func (r RequestState) GetCacheUploads() []types.Dump {
	return r.dataLoader.uploads
}

func (r RequestState) GetCacheUploadsAtIndex(index int) types.Dump {
	if index < 0 || index >= len(r.dataLoader.uploads) {
		return types.Dump{}
	}

	return r.dataLoader.uploads[index]
}

func (r *RequestState) SetAuthChecker(authChecker authz.SubRepoPermissionChecker) {
	r.authChecker = authChecker
}

func (r *RequestState) SetUploadsDataLoader(uploads []types.Dump) {
	r.dataLoader = NewUploadsDataLoader()
	for _, upload := range uploads {
		r.dataLoader.AddUpload(upload)
	}
}

func (r *RequestState) SetLocalGitTreeTranslator(client GitserverClient, repo *sgTypes.Repo, commit, path string, hunkCacheSize int) error {
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

func (r *RequestState) SetLocalCommitCache(client GitserverClient) {
	r.commitCache = NewCommitCache(client)
}

func (r *RequestState) SetMaximumIndexesPerMonikerSearch(maxNumber int) {
	r.maximumIndexesPerMonikerSearch = maxNumber
}

type UploadsDataLoader struct {
	uploads     []types.Dump
	uploadsByID map[int]types.Dump
	cacheMutex  sync.RWMutex
}

func NewUploadsDataLoader() *UploadsDataLoader {
	return &UploadsDataLoader{
		uploadsByID: make(map[int]types.Dump),
	}
}

func (l *UploadsDataLoader) GetUploadFromCacheMap(id int) (types.Dump, bool) {
	l.cacheMutex.RLock()
	defer l.cacheMutex.RUnlock()

	upload, ok := l.uploadsByID[id]
	return upload, ok
}

func (l *UploadsDataLoader) SetUploadInCacheMap(uploads []types.Dump) {
	l.cacheMutex.Lock()
	defer l.cacheMutex.Unlock()

	for i := range uploads {
		l.uploadsByID[uploads[i].ID] = uploads[i]
	}
}

func (l *UploadsDataLoader) AddUpload(dump types.Dump) {
	l.cacheMutex.Lock()
	defer l.cacheMutex.Unlock()

	l.uploads = append(l.uploads, dump)
	l.uploadsByID[dump.ID] = dump
}
