package store

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/keegancsmith/sqlf"

	ct "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/testing"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
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

	t.Run("CreateWithConflictingKey", func(t *testing.T) {
		clock.Add(1 * time.Minute)

		keyConflict := &btypes.BatchSpecExecutionCacheEntry{
			Key:   entries[0].Key,
			Value: "new value",
		}
		if err := s.CreateBatchSpecExecutionCacheEntry(ctx, keyConflict); err != nil {
			t.Fatal(err)
		}

		reloaded, err := s.GetBatchSpecExecutionCacheEntry(ctx, GetBatchSpecExecutionCacheEntryOpts{Key: keyConflict.Key})
		if err != nil {
			t.Fatal(err)
		}

		if diff := cmp.Diff(reloaded, keyConflict); diff != "" {
			t.Fatal(diff)
		}

		if reloaded.CreatedAt.Equal(entries[0].CreatedAt) {
			t.Fatal("CreatedAt not updated")
		}
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

func TestStore_CleanBatchSpecExecutionCacheEntries(t *testing.T) {
	// Separate test function because we want a clean DB

	ctx := context.Background()
	db := dbtest.NewDB(t)
	c := &ct.TestClock{Time: timeutil.Now()}
	s := NewWithClock(db, &observation.TestContext, nil, c.Now)

	maxSize := 10 * 1024 // 10kb

	for i := 0; i < 20; i += 1 {
		entry := &btypes.BatchSpecExecutionCacheEntry{
			Key:   fmt.Sprintf("cache-key-%d", i),
			Value: strings.Repeat("a", 1024),
		}

		if err := s.CreateBatchSpecExecutionCacheEntry(ctx, entry); err != nil {
			t.Fatal(err)
		}
	}

	totalSize, _, err := basestore.ScanFirstInt(s.Query(ctx, sqlf.Sprintf("SELECT sum(octet_length(value)) AS total FROM batch_spec_execution_cache_entries")))
	if err != nil {
		t.Fatal(err)
	}
	if totalSize != maxSize*2 {
		t.Fatalf("totalsize wrong=%d", totalSize)
	}

	if err := s.CleanBatchSpecExecutionCacheEntries(ctx, int64(maxSize)); err != nil {
		t.Fatal(err)
	}

	entriesLeft, _, err := basestore.ScanFirstInt(s.Query(ctx, sqlf.Sprintf("SELECT count(*) FROM batch_spec_execution_cache_entries")))
	if err != nil {
		t.Fatal(err)
	}

	wantLeft := 10
	if entriesLeft != wantLeft {
		t.Fatalf("wrong number of entries left. want=%d, have=%d", wantLeft, entriesLeft)
	}

	totalSize, _, err = basestore.ScanFirstInt(s.Query(ctx, sqlf.Sprintf("SELECT sum(octet_length(value)) AS total FROM batch_spec_execution_cache_entries")))
	if err != nil {
		t.Fatal(err)
	}
	if totalSize != maxSize {
		t.Fatalf("totalsize wrong=%d", totalSize)
	}
}
