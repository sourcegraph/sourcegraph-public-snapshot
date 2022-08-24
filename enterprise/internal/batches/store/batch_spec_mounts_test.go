package store

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	bt "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/testing"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
)

func testStoreBatchSpecMounts(t *testing.T, ctx context.Context, s *Store, clock bt.Clock) {
	mounts := make([]*btypes.BatchSpecMount, 0, 5)

	spec := &btypes.BatchSpec{
		UserID:          int32(1234),
		NamespaceUserID: int32(1234),
	}
	err := s.CreateBatchSpec(ctx, spec)
	require.NoError(t, err)

	t.Run("Create", func(t *testing.T) {
		for i := 0; i < cap(mounts); i++ {
			mount := &btypes.BatchSpecMount{
				BatchSpecID: spec.ID,
				FileName:    fmt.Sprintf("hello-%d.txt", i),
				Path:        "foo/bar",
				Size:        12,
				Content:     []byte("hello, world!"),
				ModifiedAt:  clock.Now(),
			}
			expected := mount.Clone()

			err := s.UpsertBatchSpecMount(ctx, mount)
			require.NoError(t, err)

			expected.ID = mount.ID
			expected.RandID = mount.RandID
			expected.CreatedAt = mount.CreatedAt
			expected.UpdatedAt = mount.UpdatedAt

			diff := cmp.Diff(mount, expected)
			require.Empty(t, diff)

			mounts = append(mounts, mount)
		}
		assert.Len(t, mounts, 5)
	})

	t.Run("Update", func(t *testing.T) {
		clock.Add(1 * time.Second)

		mount := mounts[0]
		mount.Size = 20

		expected := mount.Clone()
		expected.UpdatedAt = clock.Now()

		err := s.UpsertBatchSpecMount(ctx, mount)
		require.NoError(t, err)

		diff := cmp.Diff(mount, expected)
		assert.Empty(t, diff)
	})

	t.Run("Get", func(t *testing.T) {
		t.Run("ByID", func(t *testing.T) {
			expected := mounts[0]
			actual, err := s.GetBatchSpecMount(ctx, GetBatchSpecMountOpts{ID: expected.ID})
			require.NoError(t, err)

			diff := cmp.Diff(actual, expected)
			assert.Empty(t, diff)
		})

		t.Run("ByRandID", func(t *testing.T) {
			expected := mounts[0]
			actual, err := s.GetBatchSpecMount(ctx, GetBatchSpecMountOpts{RandID: expected.RandID})
			require.NoError(t, err)

			diff := cmp.Diff(actual, expected)
			assert.Empty(t, diff)
		})

		t.Run("No Options", func(t *testing.T) {
			mount, err := s.GetBatchSpecMount(ctx, GetBatchSpecMountOpts{})
			assert.Error(t, err)
			assert.Nil(t, mount)
		})
	})

	t.Run("Count", func(t *testing.T) {
		t.Run("ByBatchSpecID", func(t *testing.T) {
			count, err := s.CountBatchSpecMounts(ctx, ListBatchSpecMountsOpts{BatchSpecID: spec.ID})
			require.NoError(t, err)
			assert.Equal(t, len(mounts), count)
		})

		t.Run("ByBatchSpecRandID", func(t *testing.T) {
			count, err := s.CountBatchSpecMounts(ctx, ListBatchSpecMountsOpts{BatchSpecRandID: spec.RandID})
			require.NoError(t, err)
			assert.Equal(t, len(mounts), count)
		})
	})

	t.Run("List", func(t *testing.T) {
		t.Run("ByBatchSpecID", func(t *testing.T) {
			actual, next, err := s.ListBatchSpecMounts(ctx, ListBatchSpecMountsOpts{BatchSpecID: spec.ID})
			require.NoError(t, err)
			assert.Zero(t, next)
			assert.Len(t, actual, len(mounts))

			for _, mount := range actual {
				expected := getExpectedMount(mount.ID, mounts)
				diff := cmp.Diff(mount, expected)
				require.Empty(t, diff)
			}
		})

		t.Run("ByBatchSpecRandID", func(t *testing.T) {
			actual, next, err := s.ListBatchSpecMounts(ctx, ListBatchSpecMountsOpts{BatchSpecRandID: spec.RawSpec})
			require.NoError(t, err)
			assert.Zero(t, next)
			assert.Len(t, actual, len(mounts))

			for _, mount := range actual {
				expected := getExpectedMount(mount.ID, mounts)
				diff := cmp.Diff(mount, expected)
				require.Empty(t, diff)
			}
		})

		t.Run("Large Limit", func(t *testing.T) {
			opts := ListBatchSpecMountsOpts{
				LimitOpts: LimitOpts{
					Limit: 100,
				},
				BatchSpecRandID: spec.RawSpec,
			}
			actual, next, err := s.ListBatchSpecMounts(ctx, opts)
			require.NoError(t, err)
			assert.Zero(t, next)
			assert.Len(t, actual, len(mounts))

			for _, mount := range actual {
				expected := getExpectedMount(mount.ID, mounts)
				diff := cmp.Diff(mount, expected)
				require.Empty(t, diff)
			}
		})

		t.Run("Small Limit", func(t *testing.T) {
			opts := ListBatchSpecMountsOpts{
				LimitOpts: LimitOpts{
					Limit: 2,
				},
				BatchSpecRandID: spec.RawSpec,
			}
			actual, next, err := s.ListBatchSpecMounts(ctx, opts)
			require.NoError(t, err)
			// Limit
			assert.Equal(t, mounts[2].ID, next)
			assert.Len(t, actual, 2)

			for _, mount := range actual {
				expected := getExpectedMount(mount.ID, mounts)
				diff := cmp.Diff(mount, expected)
				require.Empty(t, diff)
			}
		})

		t.Run("From Cursor", func(t *testing.T) {
			opts := ListBatchSpecMountsOpts{
				LimitOpts: LimitOpts{
					Limit: 3,
				},
				Cursor:          int64(3),
				BatchSpecRandID: spec.RawSpec,
			}
			actual, next, err := s.ListBatchSpecMounts(ctx, opts)
			require.NoError(t, err)
			assert.Zero(t, next)
			assert.Len(t, actual, 3)

			for _, mount := range actual {
				expected := getExpectedMount(mount.ID, mounts)
				diff := cmp.Diff(mount, expected)
				require.Empty(t, diff)
			}
		})
	})

	t.Run("Delete", func(t *testing.T) {
		t.Run("No Options", func(t *testing.T) {
			err := s.DeleteBatchSpecMount(ctx, DeleteBatchSpecMountOpts{})
			assert.Error(t, err)
		})

		t.Run("ByID", func(t *testing.T) {
			err := s.DeleteBatchSpecMount(ctx, DeleteBatchSpecMountOpts{ID: mounts[0].ID})
			require.NoError(t, err)

			deletedMount, err := s.GetBatchSpecMount(ctx, GetBatchSpecMountOpts{ID: mounts[0].ID})
			require.ErrorIs(t, err, ErrNoResults)
			assert.Nil(t, deletedMount)
		})

		t.Run("ByBatchSpecID", func(t *testing.T) {
			// Add one more mount just in case current ones have been deleted
			newMount := &btypes.BatchSpecMount{
				BatchSpecID: spec.ID,
				FileName:    "by-spec-id.txt",
				Path:        "foo/bar",
				Size:        12,
				Content:     []byte("hello, world!"),
				ModifiedAt:  clock.Now(),
			}
			err := s.UpsertBatchSpecMount(ctx, newMount)
			require.NoError(t, err)

			err = s.DeleteBatchSpecMount(ctx, DeleteBatchSpecMountOpts{BatchSpecID: spec.ID})
			require.NoError(t, err)

			for _, m := range mounts {
				deletedMount, err := s.GetBatchSpecMount(ctx, GetBatchSpecMountOpts{ID: m.ID})
				require.ErrorIs(t, err, ErrNoResults)
				assert.Nil(t, deletedMount)
			}

			// And check if the new one has also been deleted
			deletedMount, err := s.GetBatchSpecMount(ctx, GetBatchSpecMountOpts{ID: newMount.ID})
			require.ErrorIs(t, err, ErrNoResults)
			assert.Nil(t, deletedMount)
		})
	})
}

func getExpectedMount(id int64, expects []*btypes.BatchSpecMount) *btypes.BatchSpecMount {
	for _, expect := range expects {
		if id == expect.ID {
			return expect
		}
	}
	return nil
}
