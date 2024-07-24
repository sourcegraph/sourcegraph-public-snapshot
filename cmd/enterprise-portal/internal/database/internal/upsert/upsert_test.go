package upsert

import (
	"testing"
	"time"

	"github.com/hexops/autogold/v2"
	"github.com/hexops/valast"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

func TestBuilder(t *testing.T) {
	mockTime := time.Date(2024, 7, 8, 16, 39, 16, 4277000, time.Local)
	for _, tc := range []struct {
		name string

		forceUpdate  bool
		upsertFields func(b *Builder)

		wantQuery autogold.Value
		wantArgs  autogold.Value
	}{
		{
			name:        "not force-update",
			forceUpdate: false,
			upsertFields: func(b *Builder) {
				Field[*string](b, "col1", nil)

				// WithIgnoreOnForceUpdate() does nothing because the we are not
				// in a force-update.
				Field(b, "col2", pointers.Ptr("value2"), WithIgnoreZeroOnForceUpdate())

				// WithColumnDefault() does nothing because the time is not zero
				Field(b, "time", mockTime, WithColumnDefault())

				// Do not set, it should use the default value.
				Field(b, "should_be_ignored", "", WithColumnDefault())
			},
			wantQuery: autogold.Expect(`
INSERT INTO table
(col1, col2, time)
VALUES
(@col1, @col2, @time)
ON CONFLICT
(id)
DO UPDATE SET
col2 = EXCLUDED.col2,
time = EXCLUDED.time`),
			wantArgs: autogold.Expect(pgx.NamedArgs{
				"col1": nil, "col2": valast.Ptr("value2"),
				"time": time.Date(2024,
					7,
					8,
					16,
					39,
					16,
					4277000,
					time.Local),
			}),
		},
		{
			name:        "force-update",
			forceUpdate: true,
			upsertFields: func(b *Builder) {
				Field[*string](b, "col1", nil)
				Field(b, "col2", pointers.Ptr("value2"))
				Field(b, "time", mockTime)

				// Do not set, it cannot be updated in a force-update.
				Field(b, "should_be_ignored", "", WithIgnoreZeroOnForceUpdate())
			},
			wantQuery: autogold.Expect(`
INSERT INTO table
(col1, col2, time)
VALUES
(@col1, @col2, @time)
ON CONFLICT
(id)
DO UPDATE SET
col1 = EXCLUDED.col1,
col2 = EXCLUDED.col2,
time = EXCLUDED.time`),
			wantArgs: autogold.Expect(pgx.NamedArgs{
				"col1": nil, "col2": valast.Ptr("value2"),
				"time": time.Date(2024,
					7,
					8,
					16,
					39,
					16,
					4277000,
					time.Local),
			}),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			b := New("table", "id", tc.forceUpdate)
			tc.upsertFields(b)

			q, ok := b.buildQuery()
			if tc.wantQuery == nil && tc.wantArgs == nil {
				assert.False(t, ok)
			} else {
				assert.True(t, ok)
				tc.wantQuery.Equal(t, q)
				tc.wantArgs.Equal(t, b.args)
			}
		})
	}
}
