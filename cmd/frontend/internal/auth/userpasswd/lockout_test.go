package userpasswd

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/internal/rcache"
)

func TestLockoutStore(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Run("explicit reset", func(t *testing.T) {
		rcache.SetupForTest(t)

		s := NewLockoutStore(1, time.Minute, time.Minute)

		_, locked := s.IsLockedOut(1)
		assert.False(t, locked)

		// Should be locked out after one failed attempt
		s.IncreaseFailedAttempt(1)
		_, locked = s.IsLockedOut(1)
		assert.True(t, locked)

		// Should be unlocked after reset
		s.Reset(1)
		_, locked = s.IsLockedOut(1)
		assert.False(t, locked)
	})

	t.Run("automatically released", func(t *testing.T) {
		rcache.SetupForTest(t)

		s := NewLockoutStore(1, 2*time.Second, time.Minute)

		_, locked := s.IsLockedOut(1)
		assert.False(t, locked)

		// Should be locked out after one failed attempt
		s.IncreaseFailedAttempt(1)
		_, locked = s.IsLockedOut(1)
		assert.True(t, locked)

		// Should be unlocked after three seconds, wait for an extra second to eliminate flakiness
		time.Sleep(3 * time.Second)
		_, locked = s.IsLockedOut(1)
		assert.False(t, locked)
	})

	t.Run("failed attempts far apart", func(t *testing.T) {
		rcache.SetupForTest(t)

		s := NewLockoutStore(2, time.Minute, time.Second)

		_, locked := s.IsLockedOut(1)
		assert.False(t, locked)

		// Should not be locked out after the consecutive period
		s.IncreaseFailedAttempt(1)
		time.Sleep(2 * time.Second) // Wait for an extra second to eliminate flakiness
		s.IncreaseFailedAttempt(1)

		_, locked = s.IsLockedOut(1)
		assert.False(t, locked)
	})
}
