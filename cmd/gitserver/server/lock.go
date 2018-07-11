package server

import (
	"path/filepath"
	"sync"
)

// RepositoryLocker provides locks for doing operations to a repository
// directory. When a repository is locked, only the owner of the lock is
// allowed to run commands against it.
//
// Repositories are identified by the absolute path to their directories. Note
// the directories are the parent of the $GIT_DIR (ie excluding the /.git
// suffix). All operations affect the $GIT_DIR, but for legacy reasons we
// identify repositories by the parent.
//
// The directory's $GIT_DIR does not have to exist when locked. The owner of
// the lock may remove the directory's $GIT_DIR while holding the lock.
//
// The main use of RepositoryLocker is to prevent concurrent clones. However,
// it is also used during maintance tasks such as recloning/migrating/etc.
type RepositoryLocker struct {
	// mu protects status
	mu sync.Mutex
	// status tracks directories that are locked. The value is the status. If
	// a directory is in status, the directory is locked.
	status map[string]string
}

// TryAcquire acquires the lock for dir. If it is already held, ok is false
// and lock is nil. Otherwise a non-nil lock is returned and true. When
// finished with the lock you must call lock.Release.
func (rl *RepositoryLocker) TryAcquire(dir string, initialStatus string) (lock *RepositoryLock, ok bool) {
	dir = rl.normalize(dir)

	rl.mu.Lock()
	_, failed := rl.status[dir]
	if !failed {
		if rl.status == nil {
			rl.status = make(map[string]string)
		}
		rl.status[dir] = initialStatus
	}
	rl.mu.Unlock()

	if failed {
		return nil, false
	}

	return &RepositoryLock{
		locker: rl,
		dir:    dir,
	}, true
}

// Status returns the status of the locked directory dir. If dir is not
// locked, then locked is false.
func (rl *RepositoryLocker) Status(dir string) (status string, locked bool) {
	dir = rl.normalize(dir)

	rl.mu.Lock()
	status, locked = rl.status[dir]
	rl.mu.Unlock()
	return
}

// normalize cleans dir and ensures dir is not pointing to the GIT_DIR, but
// rather the parent. ie it will translate
// /data/repos/example.com/foo/bar/.git to /data/repos/example.com/foo/bar
func (rl *RepositoryLocker) normalize(dir string) string {
	dir = filepath.Clean(dir)

	// Use parent if we are passed a $GIT_DIR
	if name := filepath.Base(dir); name == ".git" {
		return filepath.Dir(dir)
	}

	return dir
}

// RepositoryLock is returned by RepositoryLocker.TryAcquire. It allows
// updating the status of a directory lock, as well as releasing the lock.
type RepositoryLock struct {
	locker *RepositoryLocker
	dir    string

	// done is protected by locker.mu
	done bool
}

// SetStatus updates the status for the lock. If the lock has been released,
// this is a noop.
func (l *RepositoryLock) SetStatus(status string) {
	l.locker.mu.Lock()
	// Ensure this is still locked before updating the status
	if !l.done {
		l.locker.status[l.dir] = status
	}
	l.locker.mu.Unlock()
}

// Release releases the lock.
func (l *RepositoryLock) Release() {
	l.locker.mu.Lock()
	// Prevent double release
	if !l.done {
		delete(l.locker.status, l.dir)
		l.done = true
	}
	l.locker.mu.Unlock()
}
