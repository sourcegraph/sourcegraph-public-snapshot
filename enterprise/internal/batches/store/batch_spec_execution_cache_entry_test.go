package store

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"

	ct "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/testing"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
)

func testStoreBatchSpecExecutionCacheEntries(t *testing.T, ctx context.Context, s *Store, clock ct.Clock) {
	entries := make([]*btypes.BatchSpecExecutionCacheEntry, 0, 3)
	for i := 0; i < cap(entries); i++ {
		job := &btypes.BatchSpecExecutionCacheEntry{
			Key:   fmt.Sprintf("check-out-this-cache-key-%d", i),
			Value: fmt.Sprintf("what-about-this-cache-value-huh-%d", i),
		}

		entries = append(entries, job)
	}

	t.Run("Create", func(t *testing.T) {
		for _, job := range entries {
			if err := s.CreateBatchSpecExecutionCacheEntry(ctx, job); err != nil {
				t.Fatal(err)
			}

			have := job
			if have.ID == 0 {
				t.Fatal("ID should not be zero")
			}

			want := have
			want.CreatedAt = clock.Now()

			if diff := cmp.Diff(have, want); diff != "" {
				t.Fatal(diff)
			}
		}
	})

	t.Run("Get", func(t *testing.T) {
		t.Run("GetByKey", func(t *testing.T) {
			for i, job := range entries {
				t.Run(strconv.Itoa(i), func(t *testing.T) {
					have, err := s.GetBatchSpecExecutionCacheEntry(ctx, GetBatchSpecExecutionCacheEntryOpts{Key: job.Key})

					if err != nil {
						t.Fatal(err)
					}

					if diff := cmp.Diff(have, job); diff != "" {
						t.Fatal(diff)
					}
				})
			}
		})

		t.Run("NoResults", func(t *testing.T) {
			opts := GetBatchSpecExecutionCacheEntryOpts{
				Key: "if-this-returns-something-lol-i-will-eat-a-hat",
			}

			_, have := s.GetBatchSpecExecutionCacheEntry(ctx, opts)
			want := ErrNoResults

			if have != want {
				t.Fatalf("have err %v, want %v", have, want)
			}
		})
	})

	t.Run("MarkUsedBatchSpecExecutionCacheEntry", func(t *testing.T) {
		entry := &btypes.BatchSpecExecutionCacheEntry{
			Key:   "the-amazing-cache-key",
			Value: "the-mysterious-cache-value",
		}

		if err := s.CreateBatchSpecExecutionCacheEntry(ctx, entry); err != nil {
			t.Fatal(err)
		}

		if err := s.MarkUsedBatchSpecExecutionCacheEntry(ctx, entry.ID); err != nil {
			t.Fatal(err)
		}

		reloaded, err := s.GetBatchSpecExecutionCacheEntry(ctx, GetBatchSpecExecutionCacheEntryOpts{Key: entry.Key})
		if err != nil {
			t.Fatal(err)
		}

		if want, have := clock.Now(), reloaded.LastUsedAt; !have.Equal(want) {
			t.Fatalf("entry.LastUsedAt is wrong.\n\twant=%s\n\thave=%s", want, have)
		}
	})
}
