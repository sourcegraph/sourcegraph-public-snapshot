package userpasswd

import (
	"strconv"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/rcache"
)

// LockoutStore provides semantics for account lockout management.
type LockoutStore interface {
	// IsLockedOut returns true if the given user has been locked along with the
	// reason.
	IsLockedOut(userID int32) (reason string, locked bool)
	// IncreaseFailedAttempt increases the failed login attempt count by 1.
	IncreaseFailedAttempt(userID int32)
	// Reset clears the failed login attempt count and releases the lockout.
	Reset(userID int32)
}

type lockoutStore struct {
	failedThreshold int
	lockouts        *rcache.Cache
	failedAttempts  *rcache.Cache
}

// NewLockoutStore returns a new LockoutStore with given durations using the
// Redis cache.
func NewLockoutStore(failedThreshold int, lockoutPeriod, consecutivePeriod time.Duration) LockoutStore {
	return &lockoutStore{
		failedThreshold: failedThreshold,
		lockouts:        rcache.NewWithTTL("account_lockout", int(lockoutPeriod.Seconds())),
		failedAttempts:  rcache.NewWithTTL("account_failed_attempts", int(consecutivePeriod.Seconds())),
	}
}

func (s *lockoutStore) IsLockedOut(userID int32) (reason string, locked bool) {
	v, locked := s.lockouts.Get(strconv.Itoa(int(userID)))
	return string(v), locked
}

func (s *lockoutStore) IncreaseFailedAttempt(userID int32) {
	key := strconv.Itoa(int(userID))
	s.failedAttempts.Increase(key)

	// Get right after Increase should make the key always exist
	v, _ := s.failedAttempts.Get(key)
	count, _ := strconv.Atoi(string(v))
	if count >= s.failedThreshold {
		s.lockouts.Set(key, []byte("too many failed attempts"))
	}
}

func (s *lockoutStore) Reset(userID int32) {
	key := strconv.Itoa(int(userID))
	s.lockouts.Delete(key)
	s.failedAttempts.Delete(key)
}
