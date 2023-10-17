package internal

import (
	"sync"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/common"
)

// RepositoryLock is returned by RepositoryLocker.TryAcquire. It allows
// updating the status of a directory lock, as well as releasing the lock.
type RepositoryLock interface {
	// SetStatus updates the status for the lock. If the lock has been released,
	// this is a noop.
	SetStatus(status string)
	// Release releases the lock.
	Release()
}

// RepositoryLocker provides locks for doing operations to a repository
// directory. When a repository is locked, only the owner of the lock is
// allowed to run commands against it.
//
// Repositories are identified by the absolute path to their $GIT_DIR.
//
// The directory's $GIT_DIR does not have to exist when locked. The owner of
// the lock may remove the directory's $GIT_DIR while holding the lock.
//
// The main use of RepositoryLocker is to prevent concurrent clones. However,
// it is also used during maintenance tasks such as recloning/migrating/etc.
type RepositoryLocker interface {
	// TryAcquire acquires the lock for dir. If it is already held, ok is false
	// and lock is nil. Otherwise a non-nil lock is returned and true. When
	// finished with the lock you must call lock.Release.
	TryAcquire(dir common.GitDir, initialStatus string) (lock RepositoryLock, ok bool)
	// Status returns the status of the locked directory dir. If dir is not
	// locked, then locked is false.
	Status(dir common.GitDir) (status string, locked bool)
}

func NewRepositoryLocker() RepositoryLocker {
	return &repositoryLocker{
		status: make(map[common.GitDir]string),
	}
}

type repositoryLocker struct {
	// mu protects status
	mu sync.RWMutex
	// status tracks directories that are locked. The value is the status. If
	// a directory is in status, the directory is locked.
	status map[common.GitDir]string
}

func (rl *repositoryLocker) TryAcquire(dir common.GitDir, initialStatus string) (lock RepositoryLock, ok bool) {
	rl.mu.Lock()
	_, failed := rl.status[dir]
	if !failed {
		if rl.status == nil {
			rl.status = make(map[common.GitDir]string)
		}
		rl.status[dir] = initialStatus
	}
	rl.mu.Unlock()

	if failed {
		return nil, false
	}

	return &repositoryLock{
		unlock: func() {
			rl.mu.Lock()
			delete(rl.status, dir)
			rl.mu.Unlock()
		},
		setStatus: func(status string) {
			rl.mu.Lock()
			rl.status[dir] = status
			rl.mu.Unlock()
		},
		dir: dir,
	}, true
}

func (rl *repositoryLocker) Status(dir common.GitDir) (status string, locked bool) {
	rl.mu.RLock()
	defer rl.mu.RUnlock()
	status, locked = rl.status[dir]
	return
}

type repositoryLock struct {
	unlock    func()
	setStatus func(status string)
	dir       common.GitDir

	mu   sync.Mutex
	done bool
}

func (l *repositoryLock) SetStatus(status string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Ensure this is still locked before updating the status
	if !l.done {
		l.setStatus(status)
	}
}

func (l *repositoryLock) Release() {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Prevent double release
	if !l.done {
		l.unlock()
		l.done = true
	}
}
