package store

import (
	"context"
	"strconv"
	"testing"

	"github.com/keegancsmith/sqlf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	bt "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/testing"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

func testCursor(t *testing.T, ctx context.Context, s *Store, _ bt.Clock) {
	t.Run("without values", func(t *testing.T) {
		for _, limit := range []int{0, 5} {
			t.Run(strconv.Itoa(limit), func(t *testing.T) {
				have, cursor, err := s.ListPaginationTests(ctx, ListPaginationTestOpts{
					CursorOpts: CursorOpts{Limit: limit},
				})
				assert.NoError(t, err)
				assert.Empty(t, have)
				assert.Zero(t, cursor)
			})
		}
	})

	t.Run("with values", func(t *testing.T) {
		tests := []*PaginationTest{}
		for i := 0; i < 5; i++ {
			pt := PaginationTest{}
			require.NoError(t, s.CreatePaginationTest(ctx, &pt))
			tests = append(tests, &pt)
		}

		reversed := make([]*PaginationTest, 0, len(tests))
		for i := len(tests) - 1; i >= 0; i-- {
			reversed = append(reversed, tests[i])
		}

		t.Run("single page results", func(t *testing.T) {
			for _, limit := range []int{0, 5, 10} {
				t.Run(strconv.Itoa(limit), func(t *testing.T) {
					have, cursor, err := s.ListPaginationTests(ctx, ListPaginationTestOpts{
						CursorOpts: CursorOpts{Limit: limit},
						// Default direction should be ascending.
					})
					assert.NoError(t, err)
					assert.Equal(t, tests, have)
					assert.Zero(t, cursor)

					have, cursor, err = s.ListPaginationTests(ctx, ListPaginationTestOpts{
						CursorOpts: CursorOpts{Limit: limit},
						Direction:  CursorDirectionAscending, // Explicit direction.
					})
					assert.NoError(t, err)
					assert.Equal(t, tests, have)
					assert.Zero(t, cursor)

					have, cursor, err = s.ListPaginationTests(ctx, ListPaginationTestOpts{
						CursorOpts: CursorOpts{Limit: limit},
						Direction:  CursorDirectionDescending,
					})
					assert.NoError(t, err)
					assert.Equal(t, reversed, have)
					assert.Zero(t, cursor)
				})
			}
		})

		t.Run("paginated results", func(t *testing.T) {
			for dir, have := range map[CursorDirection][]*PaginationTest{
				CursorDirectionAscending:  tests,
				CursorDirectionDescending: reversed,
			} {
				t.Run(dir.String()+", homogeneous", func(t *testing.T) {
					page, cursor, err := s.ListPaginationTests(ctx, ListPaginationTestOpts{
						CursorOpts: CursorOpts{Limit: 2},
						Direction:  dir,
					})
					assert.NoError(t, err)
					assert.Equal(t, have[0:2], page)
					assert.EqualValues(t, have[2].ID, cursor)

					page, cursor, err = s.ListPaginationTests(ctx, ListPaginationTestOpts{
						CursorOpts: CursorOpts{Limit: 2, Cursor: cursor},
						Direction:  dir,
					})
					assert.NoError(t, err)
					assert.Equal(t, have[2:4], page)
					assert.EqualValues(t, have[4].ID, cursor)

					page, cursor, err = s.ListPaginationTests(ctx, ListPaginationTestOpts{
						CursorOpts: CursorOpts{Limit: 2, Cursor: cursor},
						Direction:  dir,
					})
					assert.NoError(t, err)
					assert.Equal(t, have[4:], page)
					assert.Zero(t, cursor)
				})

				t.Run(dir.String()+", heterogeneous", func(t *testing.T) {
					page, cursor, err := s.ListPaginationTests(ctx, ListPaginationTestOpts{
						CursorOpts: CursorOpts{Limit: 1},
						Direction:  dir,
					})
					assert.NoError(t, err)
					assert.Equal(t, have[0:1], page)
					assert.EqualValues(t, have[1].ID, cursor)

					page, cursor, err = s.ListPaginationTests(ctx, ListPaginationTestOpts{
						CursorOpts: CursorOpts{Limit: 4, Cursor: cursor},
						Direction:  dir,
					})
					assert.NoError(t, err)
					assert.Equal(t, have[1:], page)
					assert.Zero(t, cursor)
				})
			}
		})
	})
}

type PaginationTest struct {
	ID int
}

func (pt *PaginationTest) Cursor() int64 { return int64(pt.ID) }

const paginationTestCreateQuery = `
CREATE TABLE pagination_test (
  id SERIAL PRIMARY KEY NOT NULL
)
`

func (s *Store) CreatePaginationTest(ctx context.Context, pt *PaginationTest) error {
	q := sqlf.Sprintf("INSERT INTO pagination_test DEFAULT VALUES RETURNING id")
	return createOrUpdateRecord(ctx, s, q, scanPaginationTest, pt)
}

type ListPaginationTestOpts struct {
	CursorOpts
	Direction CursorDirection
}

func (s *Store) ListPaginationTests(ctx context.Context, opts ListPaginationTestOpts) ([]*PaginationTest, int64, error) {
	// Although this test demonstrates using the direction to dynamically build a
	// query, this isn't required: you can also hard code the direction into the
	// ORDER BY statement in the statement, provide the correct direction constant
	// when calling WhereDB, and everything will be just fine.
	q := sqlf.Sprintf(
		"SELECT id FROM pagination_test WHERE %s ORDER BY id %s %s",
		opts.WhereDB("id", opts.Direction),
		sqlf.Sprintf(opts.Direction.String()),
		opts.LimitDB(),
	)
	return listRecords(ctx, s, q, opts.CursorOpts, scanPaginationTest)
}

func scanPaginationTest(pt *PaginationTest, sc dbutil.Scanner) error {
	return sc.Scan(&pt.ID)
}
