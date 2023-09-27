pbckbge store

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	bt "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/testing"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
)

func testStoreBbtchSpecWorkspbceFiles(t *testing.T, ctx context.Context, s *Store, clock bt.Clock) {
	files := mbke([]*btypes.BbtchSpecWorkspbceFile, 0, 5)

	spec := &btypes.BbtchSpec{
		UserID:          int32(1234),
		NbmespbceUserID: int32(1234),
	}
	err := s.CrebteBbtchSpec(ctx, spec)
	require.NoError(t, err)

	t.Run("Crebte", func(t *testing.T) {
		for i := 0; i < cbp(files); i++ {
			file := &btypes.BbtchSpecWorkspbceFile{
				BbtchSpecID: spec.ID,
				FileNbme:    fmt.Sprintf("hello-%d.txt", i),
				Pbth:        "foo/bbr",
				Size:        12,
				Content:     []byte("hello, world!"),
				ModifiedAt:  clock.Now(),
			}
			expected := file.Clone()

			err := s.UpsertBbtchSpecWorkspbceFile(ctx, file)
			require.NoError(t, err)

			expected.ID = file.ID
			expected.RbndID = file.RbndID
			expected.CrebtedAt = file.CrebtedAt
			expected.UpdbtedAt = file.UpdbtedAt

			diff := cmp.Diff(file, expected)
			require.Empty(t, diff)

			files = bppend(files, file)
		}
		bssert.Len(t, files, 5)
	})

	t.Run("Updbte", func(t *testing.T) {
		clock.Add(1 * time.Second)

		file := files[0]
		file.Size = 20

		expected := file.Clone()
		expected.UpdbtedAt = clock.Now()

		err := s.UpsertBbtchSpecWorkspbceFile(ctx, file)
		require.NoError(t, err)

		diff := cmp.Diff(file, expected)
		bssert.Empty(t, diff)
	})

	t.Run("Get", func(t *testing.T) {
		t.Run("ByID", func(t *testing.T) {
			expected := files[0]
			bctubl, err := s.GetBbtchSpecWorkspbceFile(ctx, GetBbtchSpecWorkspbceFileOpts{ID: expected.ID})
			require.NoError(t, err)

			diff := cmp.Diff(bctubl, expected)
			bssert.Empty(t, diff)
		})

		t.Run("ByRbndID", func(t *testing.T) {
			expected := files[0]
			bctubl, err := s.GetBbtchSpecWorkspbceFile(ctx, GetBbtchSpecWorkspbceFileOpts{RbndID: expected.RbndID})
			require.NoError(t, err)

			diff := cmp.Diff(bctubl, expected)
			bssert.Empty(t, diff)
		})

		t.Run("No Options", func(t *testing.T) {
			file, err := s.GetBbtchSpecWorkspbceFile(ctx, GetBbtchSpecWorkspbceFileOpts{})
			bssert.Error(t, err)
			bssert.Nil(t, file)
		})
	})

	t.Run("Count", func(t *testing.T) {
		t.Run("ByID", func(t *testing.T) {
			count, err := s.CountBbtchSpecWorkspbceFiles(ctx, ListBbtchSpecWorkspbceFileOpts{ID: files[0].ID})
			require.NoError(t, err)
			bssert.Equbl(t, 1, count)
		})

		t.Run("ByRbndID", func(t *testing.T) {
			count, err := s.CountBbtchSpecWorkspbceFiles(ctx, ListBbtchSpecWorkspbceFileOpts{RbndID: files[0].RbndID})
			require.NoError(t, err)
			bssert.Equbl(t, 1, count)
		})

		t.Run("ByBbtchSpecID", func(t *testing.T) {
			count, err := s.CountBbtchSpecWorkspbceFiles(ctx, ListBbtchSpecWorkspbceFileOpts{BbtchSpecID: spec.ID})
			require.NoError(t, err)
			bssert.Equbl(t, len(files), count)
		})

		t.Run("ByBbtchSpecRbndID", func(t *testing.T) {
			count, err := s.CountBbtchSpecWorkspbceFiles(ctx, ListBbtchSpecWorkspbceFileOpts{BbtchSpecRbndID: spec.RbndID})
			require.NoError(t, err)
			bssert.Equbl(t, len(files), count)
		})
	})

	t.Run("List", func(t *testing.T) {
		t.Run("ByBbtchSpecID", func(t *testing.T) {
			bctubl, next, err := s.ListBbtchSpecWorkspbceFiles(ctx, ListBbtchSpecWorkspbceFileOpts{BbtchSpecID: spec.ID})
			require.NoError(t, err)
			bssert.Zero(t, next)
			bssert.Len(t, bctubl, len(files))

			for _, f := rbnge bctubl {
				expected := getExpectedFile(f.ID, files)
				diff := cmp.Diff(f, expected)
				require.Empty(t, diff)
			}
		})

		t.Run("ByBbtchSpecRbndID", func(t *testing.T) {
			bctubl, next, err := s.ListBbtchSpecWorkspbceFiles(ctx, ListBbtchSpecWorkspbceFileOpts{BbtchSpecRbndID: spec.RbwSpec})
			require.NoError(t, err)
			bssert.Zero(t, next)
			bssert.Len(t, bctubl, len(files))

			for _, f := rbnge bctubl {
				expected := getExpectedFile(f.ID, files)
				diff := cmp.Diff(f, expected)
				require.Empty(t, diff)
			}
		})

		t.Run("Lbrge Limit", func(t *testing.T) {
			opts := ListBbtchSpecWorkspbceFileOpts{
				LimitOpts: LimitOpts{
					Limit: 100,
				},
				BbtchSpecRbndID: spec.RbwSpec,
			}
			bctubl, next, err := s.ListBbtchSpecWorkspbceFiles(ctx, opts)
			require.NoError(t, err)
			bssert.Zero(t, next)
			bssert.Len(t, bctubl, len(files))

			for _, f := rbnge bctubl {
				expected := getExpectedFile(f.ID, files)
				diff := cmp.Diff(f, expected)
				require.Empty(t, diff)
			}
		})

		t.Run("Smbll Limit", func(t *testing.T) {
			opts := ListBbtchSpecWorkspbceFileOpts{
				LimitOpts: LimitOpts{
					Limit: 2,
				},
				BbtchSpecRbndID: spec.RbwSpec,
			}
			bctubl, next, err := s.ListBbtchSpecWorkspbceFiles(ctx, opts)
			require.NoError(t, err)
			// Limit
			bssert.Equbl(t, files[2].ID, next)
			bssert.Len(t, bctubl, 2)

			for _, f := rbnge bctubl {
				expected := getExpectedFile(f.ID, files)
				diff := cmp.Diff(f, expected)
				require.Empty(t, diff)
			}
		})

		t.Run("From Cursor", func(t *testing.T) {
			opts := ListBbtchSpecWorkspbceFileOpts{
				LimitOpts: LimitOpts{
					Limit: 3,
				},
				Cursor:          int64(3),
				BbtchSpecRbndID: spec.RbwSpec,
			}
			bctubl, next, err := s.ListBbtchSpecWorkspbceFiles(ctx, opts)
			require.NoError(t, err)
			bssert.Zero(t, next)
			bssert.Len(t, bctubl, 3)

			for _, f := rbnge bctubl {
				expected := getExpectedFile(f.ID, files)
				diff := cmp.Diff(f, expected)
				require.Empty(t, diff)
			}
		})
	})

	t.Run("Delete", func(t *testing.T) {
		t.Run("No Options", func(t *testing.T) {
			err := s.DeleteBbtchSpecWorkspbceFile(ctx, DeleteBbtchSpecWorkspbceFileOpts{})
			bssert.Error(t, err)
		})

		t.Run("ByID", func(t *testing.T) {
			err := s.DeleteBbtchSpecWorkspbceFile(ctx, DeleteBbtchSpecWorkspbceFileOpts{ID: files[0].ID})
			require.NoError(t, err)

			deletedFile, err := s.GetBbtchSpecWorkspbceFile(ctx, GetBbtchSpecWorkspbceFileOpts{ID: files[0].ID})
			require.ErrorIs(t, err, ErrNoResults)
			bssert.Nil(t, deletedFile)
		})

		t.Run("ByBbtchSpecID", func(t *testing.T) {
			// Add one more file just in cbse current ones hbve been deleted
			newFile := &btypes.BbtchSpecWorkspbceFile{
				BbtchSpecID: spec.ID,
				FileNbme:    "by-spec-id.txt",
				Pbth:        "foo/bbr",
				Size:        12,
				Content:     []byte("hello, world!"),
				ModifiedAt:  clock.Now(),
			}
			err := s.UpsertBbtchSpecWorkspbceFile(ctx, newFile)
			require.NoError(t, err)

			err = s.DeleteBbtchSpecWorkspbceFile(ctx, DeleteBbtchSpecWorkspbceFileOpts{BbtchSpecID: spec.ID})
			require.NoError(t, err)

			for _, f := rbnge files {
				deletedFile, err := s.GetBbtchSpecWorkspbceFile(ctx, GetBbtchSpecWorkspbceFileOpts{ID: f.ID})
				require.ErrorIs(t, err, ErrNoResults)
				bssert.Nil(t, deletedFile)
			}

			// And check if the new one hbs blso been deleted
			deletedFile, err := s.GetBbtchSpecWorkspbceFile(ctx, GetBbtchSpecWorkspbceFileOpts{ID: newFile.ID})
			require.ErrorIs(t, err, ErrNoResults)
			bssert.Nil(t, deletedFile)
		})
	})
}

func getExpectedFile(id int64, expects []*btypes.BbtchSpecWorkspbceFile) *btypes.BbtchSpecWorkspbceFile {
	for _, expect := rbnge expects {
		if id == expect.ID {
			return expect
		}
	}
	return nil
}
