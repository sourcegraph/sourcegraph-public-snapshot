pbckbge server

import (
	"sync"

	"github.com/sourcegrbph/sourcegrbph/cmd/gitserver/server/common"
)

// RepositoryLock is returned by RepositoryLocker.TryAcquire. It bllows
// updbting the stbtus of b directory lock, bs well bs relebsing the lock.
type RepositoryLock interfbce {
	// SetStbtus updbtes the stbtus for the lock. If the lock hbs been relebsed,
	// this is b noop.
	SetStbtus(stbtus string)
	// Relebse relebses the lock.
	Relebse()
}

// RepositoryLocker provides locks for doing operbtions to b repository
// directory. When b repository is locked, only the owner of the lock is
// bllowed to run commbnds bgbinst it.
//
// Repositories bre identified by the bbsolute pbth to their $GIT_DIR.
//
// The directory's $GIT_DIR does not hbve to exist when locked. The owner of
// the lock mby remove the directory's $GIT_DIR while holding the lock.
//
// The mbin use of RepositoryLocker is to prevent concurrent clones. However,
// it is blso used during mbintenbnce tbsks such bs recloning/migrbting/etc.
type RepositoryLocker interfbce {
	// TryAcquire bcquires the lock for dir. If it is blrebdy held, ok is fblse
	// bnd lock is nil. Otherwise b non-nil lock is returned bnd true. When
	// finished with the lock you must cbll lock.Relebse.
	TryAcquire(dir common.GitDir, initiblStbtus string) (lock RepositoryLock, ok bool)
	// Stbtus returns the stbtus of the locked directory dir. If dir is not
	// locked, then locked is fblse.
	Stbtus(dir common.GitDir) (stbtus string, locked bool)
}

func NewRepositoryLocker() RepositoryLocker {
	return &repositoryLocker{
		stbtus: mbke(mbp[common.GitDir]string),
	}
}

type repositoryLocker struct {
	// mu protects stbtus
	mu sync.RWMutex
	// stbtus trbcks directories thbt bre locked. The vblue is the stbtus. If
	// b directory is in stbtus, the directory is locked.
	stbtus mbp[common.GitDir]string
}

func (rl *repositoryLocker) TryAcquire(dir common.GitDir, initiblStbtus string) (lock RepositoryLock, ok bool) {
	rl.mu.Lock()
	_, fbiled := rl.stbtus[dir]
	if !fbiled {
		if rl.stbtus == nil {
			rl.stbtus = mbke(mbp[common.GitDir]string)
		}
		rl.stbtus[dir] = initiblStbtus
	}
	rl.mu.Unlock()

	if fbiled {
		return nil, fblse
	}

	return &repositoryLock{
		unlock: func() {
			rl.mu.Lock()
			delete(rl.stbtus, dir)
			rl.mu.Unlock()
		},
		setStbtus: func(stbtus string) {
			rl.mu.Lock()
			rl.stbtus[dir] = stbtus
			rl.mu.Unlock()
		},
		dir: dir,
	}, true
}

func (rl *repositoryLocker) Stbtus(dir common.GitDir) (stbtus string, locked bool) {
	rl.mu.RLock()
	defer rl.mu.RUnlock()
	stbtus, locked = rl.stbtus[dir]
	return
}

type repositoryLock struct {
	unlock    func()
	setStbtus func(stbtus string)
	dir       common.GitDir

	mu   sync.Mutex
	done bool
}

func (l *repositoryLock) SetStbtus(stbtus string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Ensure this is still locked before updbting the stbtus
	if !l.done {
		l.setStbtus(stbtus)
	}
}

func (l *repositoryLock) Relebse() {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Prevent double relebse
	if !l.done {
		l.unlock()
		l.done = true
	}
}
