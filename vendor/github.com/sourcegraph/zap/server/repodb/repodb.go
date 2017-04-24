package repodb

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/go-kit/kit/log"
	"github.com/sourcegraph/jsonrpc2"
	"github.com/sourcegraph/zap"
	"github.com/sourcegraph/zap/pkg/fpath"
	"github.com/sourcegraph/zap/server/refdb"
)

// A SyncError occurs when a RepoDB method is called and the
// caller is not holding the necessary lock.
type SyncError struct {
	Op     string // the method that was called
	Repo   string // the repo path that the operation was being performed on
	Locked string // the repo path whose lock is held (or "" if no lock is held)
}

func (e *SyncError) Error() string {
	if e.Locked != "" {
		return fmt.Sprintf("repodb sync %s: locked repo %q != %q", e.Op, e.Locked, e.Repo)
	}
	return fmt.Sprintf("repodb sync %s: repo %q not locked", e.Op, e.Repo)
}

// RepoDB stores and retrieves repository metadata for the server. It
// uses its Backend to access the actual (Git) repository data.
//
// A RepoDB is safe for concurrent access by multiple goroutines.
//
// It provides for exclusive locks on repo paths. The holder of a repo
// name lock must release the lock before any other operation can be
// performed on that repo path.
//
// The lock is on a repo path, not a repo. This means a call to
// Get("x") acquires a lock on "x" even if there is no existing
// repo "x". This allows the caller to then create the repo "x"
// without racing other concurrent goroutines that are trying to do
// the same thing.
type RepoDB struct {
	Backend Backend

	mu    sync.Mutex
	repos map[fpath.KeyString]*Repo

	pathMu sync.Mutex
	path   map[fpath.KeyString]*sync.Mutex // repo path (case-normalized) -> lock
}

// New creates a new repository database.
func New(backend Backend) *RepoDB {
	return &RepoDB{
		Backend: backend,
		repos:   map[fpath.KeyString]*Repo{},
	}
}

// lock locks the provided repo path. The caller must not hold db.mu
// (or else there will likely be a deadlock).
func (db *RepoDB) lock(repo string) (unlock func()) {
	db.pathMu.Lock()
	if db.path == nil {
		db.path = map[fpath.KeyString]*sync.Mutex{}
	}
	mu, ok := db.path[fpath.Key(repo)]
	if !ok {
		mu = new(sync.Mutex)
		db.path[fpath.Key(repo)] = mu
	}
	db.pathMu.Unlock()
	mu.Lock()
	return mu.Unlock
}

// List returns a list of paths of all repositories on the server.
func (db *RepoDB) List() []string {
	db.mu.Lock()
	defer db.mu.Unlock()
	paths := make([]string, 0, len(db.repos))
	for _, repo := range db.repos {
		// Use repo.Path (original case) not db.repos map key
		// (normalized case).
		paths = append(paths, repo.Path)
	}
	sort.Strings(paths)
	return paths
}

// Get returns the repo with the given repo path that has already been
// added to this repodb (with the Add method). It calls the backend's
// CanAccess method to check accessibility.
//
// If no repo with the given path has been added to this repodb, an
// error is returned.
//
// If err == nil, the caller holds the exclusive lock for the repo
// path. The caller is responsible for unlocking it when it no longer
// needs exclusive access.
//
// Usage example:
//
//   repo, err := repodb.Get(ctx, logger, "x")
//   if err != nil { return err }
//   defer repo.Unlock()
func (db *RepoDB) Get(ctx context.Context, logger log.Logger, path string) (*OwnedRepo, error) {
	if err := db.canAccess(ctx, logger, path); err != nil {
		return nil, err
	}

	unlock := db.lock(path)
	db.mu.Lock()
	repo, exists := db.repos[fpath.Key(path)]
	db.mu.Unlock()
	if exists {
		return &OwnedRepo{Repo: repo, path: path, unlock: unlock}, nil
	}
	unlock()
	return nil, &jsonrpc2.Error{
		Code:    int64(zap.ErrorCodeRepoNotExists),
		Message: fmt.Sprintf("repo does not exist: %s", path),
	}
}

// Add adds the repo with the given repo path. The repo must already
// exist on the repodb backend; this Add call just makes it available
// to the Zap server. If the repo has already been added to this
// repodb, Add returns it.
//
// If err == nil, the repo is returned and the lock is held by the
// caller. The caller is responsible for unlocking the repo path when
// it no longer needs exclusive access:
//
//   repo, err := repodb.Add(ctx, logger, "x")
//   if err != nil { return err }
//   // Do some things exclusively (e.g., enqueue some request).
//   repo.Unlock()
//   // Do some more things that don't require an exclusive lock.
//
// An error is returned if the repo doesn't exist on the backend, or
// if the access check fails (using the backend's CanAccess method).
func (db *RepoDB) Add(ctx context.Context, logger log.Logger, path string) (*OwnedRepo, error) {
	if err := db.canAccess(ctx, logger, path); err != nil {
		return nil, err
	}

	unlock := db.lock(path)
	db.mu.Lock()
	defer db.mu.Unlock()
	repo, exists := db.repos[fpath.Key(path)]
	if !exists {
		repo = &Repo{
			Path:  path,
			RefDB: refdb.Sync(refdb.NewMemoryRefDB()),
		}
		db.repos[fpath.Key(path)] = repo
	}
	return &OwnedRepo{Repo: repo, path: path, unlock: unlock}, nil
}

// Delete deletes a repo from the repodb. The caller must hold the
// repo path lock for the repo to delete.
//
// The caller is still responsible for calling repo.Unlock() on the
// repo after the Delete call returns.
func (db *RepoDB) Delete(repo OwnedRepo) error {
	// No need to check canAccess because the caller must have
	// previously acquired the repo arg from a call to Get or Add
	// (which do check canAccess).

	if repo.Repo == nil {
		panic("repo.Repo == nil")
	}
	if repo.unlock == nil {
		return &SyncError{Op: "Delete", Repo: repo.Repo.Path}
	}
	if repo.Repo.Path != repo.path {
		return &SyncError{Op: "Delete", Repo: repo.Repo.Path, Locked: repo.path}
	}
	db.mu.Lock()
	defer db.mu.Unlock()
	delete(db.repos, fpath.Key(repo.path))
	return nil
}

// canAccess checks that the current context is authorized to access
// the repo by calling the backend's CanAccess method.
func (db *RepoDB) canAccess(ctx context.Context, logger log.Logger, path string) error {
	ok, err := db.Backend.CanAccess(ctx, logger, path)
	if err != nil {
		return err
	}
	if !ok {
		return &jsonrpc2.Error{
			Code:    int64(zap.ErrorCodeRepoNotExists),
			Message: fmt.Sprintf("access forbidden to repo: %s", path),
		}
	}
	return nil
}
