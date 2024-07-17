package internal

import (
	"context"
	"sync"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/tenant"
)

// RepositoryLock is returned by RepositoryLocker.TryAcquire. It allows
// updating the status of a repo lock, as well as releasing the lock.
type RepositoryLock interface {
	// SetStatus updates the status for the lock. If the lock has been released,
	// this is a noop.
	SetStatus(status string)
	// Release releases the lock.
	Release()
}

// RepositoryLocker provides locks for doing operations to a repository.
// When a repository is locked, only the owner of the lock is allowed to perform
// writing operations against it.
//
// The main use of RepositoryLocker is to prevent concurrent fetches.
type RepositoryLocker interface {
	// TryAcquire acquires the lock for repo. If it is already held, ok is false
	// and lock is nil. Otherwise a non-nil lock is returned and true. When
	// finished with the lock you must call lock.Release.
	TryAcquire(ctx context.Context, repo api.RepoName, initialStatus string) (lock RepositoryLock, ok bool)
	// Status returns the status of the locked repo. If repo is not locked, then
	// locked is false.
	Status(ctx context.Context, repo api.RepoName) (status string, locked bool)
	// AllStatuses returns the status of all locked repositories.
	AllStatuses(ctx context.Context) map[api.RepoName]string
}

func NewRepositoryLocker() RepositoryLocker {
	return &repositoryLocker{
		status: make(map[int]map[api.RepoName]string),
	}
}

type repositoryLocker struct {
	// mu protects status
	mu sync.RWMutex
	// status tracks repos that are locked by tenant ID. The value is the status. If
	// a repo is in status, the repo is locked.
	status map[int]map[api.RepoName]string
}

func (rl *repositoryLocker) TryAcquire(ctx context.Context, repo api.RepoName, initialStatus string) (lock RepositoryLock, ok bool) {
	tnt := tenant.FromContext(ctx)

	rl.mu.Lock()
	_, found := rl.status[tnt.ID()]
	if !found {
		if rl.status == nil {
			rl.status = make(map[int]map[api.RepoName]string)
		}
	}
	_, found = rl.status[tnt.ID()][repo]
	if !found {
		if rl.status[tnt.ID()] == nil {
			rl.status[tnt.ID()] = make(map[api.RepoName]string)
		}
		rl.status[tnt.ID()][repo] = initialStatus
	}
	rl.mu.Unlock()

	if found {
		return nil, false
	}

	return &repositoryLock{
		unlock: func() {
			rl.mu.Lock()
			delete(rl.status[tnt.ID()], repo)
			rl.mu.Unlock()
		},
		setStatus: func(status string) {
			rl.mu.Lock()
			rl.status[tnt.ID()][repo] = status
			rl.mu.Unlock()
		},
		repo: repo,
	}, true
}

func (rl *repositoryLocker) Status(ctx context.Context, repo api.RepoName) (status string, locked bool) {
	rl.mu.RLock()
	defer rl.mu.RUnlock()
	tnt := tenant.FromContext(ctx)
	_, locked = rl.status[tnt.ID()]
	if !locked {
		return "", false
	}
	status, locked = rl.status[tnt.ID()][repo]
	return
}

func (rl *repositoryLocker) AllStatuses(ctx context.Context) map[api.RepoName]string {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	tnt := tenant.FromContext(ctx)

	statuses := make(map[api.RepoName]string, len(rl.status))
	for repo, status := range rl.status[tnt.ID()] {
		statuses[repo] = status
	}

	return statuses
}

type repositoryLock struct {
	unlock    func()
	setStatus func(status string)
	repo      api.RepoName

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
