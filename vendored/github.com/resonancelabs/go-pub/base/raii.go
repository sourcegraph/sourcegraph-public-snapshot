/*
Provides a form of RAII for acquiring and releasing resources (only once).

A example when used with sync.Mutex might be :
defer base.Release(base.Acquire(&mutex))
*/
package base

import "sync"

// Defines the interface for one-time-releasable resources.
// acquiredLock, as created by Acquire{,R,W} implements this.
type OneTimeReleaser interface {
	ReleaseOnce()
}

func Release(acquired OneTimeReleaser) {
	acquired.ReleaseOnce()
}

// sync.Mutex and sync.RWMutex
type acquiredLock struct {
	locker   sync.Locker
	isLocked bool
}

// Unlocks the aquired lock only once.
func (l *acquiredLock) ReleaseOnce() {
	if l.isLocked {
		l.isLocked = false
		l.locker.Unlock()
	}
}

func Acquire(mutex *sync.Mutex) OneTimeReleaser {
	mutex.Lock()
	return &acquiredLock{locker: mutex, isLocked: true}
}

func AcquireR(mutex *sync.RWMutex) OneTimeReleaser {
	mutex.RLock()
	return &acquiredLock{locker: mutex.RLocker(), isLocked: true}
}

func AcquireW(mutex *sync.RWMutex) OneTimeReleaser {
	mutex.Lock()
	return &acquiredLock{locker: mutex, isLocked: true}
}
