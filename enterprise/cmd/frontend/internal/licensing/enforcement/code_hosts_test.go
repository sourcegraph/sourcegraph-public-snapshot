package enforcement

import (
	"context"
	"errors"
	"testing"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/licensing"
	"github.com/sourcegraph/sourcegraph/internal/db"
)

func TestNewPreCreateCodeHostHook(t *testing.T) {
	ctx := context.Background()
	if !licensing.EnforceTiers {
		licensing.EnforceTiers = true
		defer func() { licensing.EnforceTiers = false }()
	}

	t.Run("when code hosts can be added", func(t *testing.T) {
		mockGetCodeHostsLimit = func(_ context.Context) int {
			return 20
		}
		defer func() { mockGetCodeHostsLimit = nil }()

		hook := NewPreCreateCodeHostHook(&mockCodeHostsStore{codeHostCount: 10})
		if got := hook(ctx); got != nil {
			t.Fatalf("got %v, want nil", got)
		}
	})

	t.Run("when the code host limit has been reached", func(t *testing.T) {
		mockGetCodeHostsLimit = func(_ context.Context) int {
			return 5
		}
		defer func() { mockGetCodeHostsLimit = nil }()

		hook := NewPreCreateCodeHostHook(&mockCodeHostsStore{codeHostCount: 10})
		if hook(ctx) == nil {
			t.Fatalf("got nil, want error")
		}
	})

	t.Run("when there is a database error", func(t *testing.T) {
		hook := NewPreCreateCodeHostHook(&mockCodeHostsStore{err: errors.New("test fail")})
		if hook(ctx) == nil {
			t.Fatalf("got nil, want error")
		}
	})
}

// A mockCodeHostsStore implements the CodeHostsStore interface for test purposes.
type mockCodeHostsStore struct {
	codeHostCount int
	err           error
}

// Count returns the number of code hosts currently configured.
func (m *mockCodeHostsStore) Count(_ context.Context, _ db.CodeHostsListOptions) (int, error) {
	return m.codeHostCount, m.err
}
