package changed

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseDiff(t *testing.T) {
	t.Run("All", func(t *testing.T) {
		assert.False(t, All.Has(None))
		var lastName string
		for diff := Go; diff <= All; diff <<= 1 {
			assert.True(t, All.Has(diff))
			lastName = diff.String()
		}
		assert.Equal(t, lastName, "All")
	})

	tests := []struct {
		name             string
		files            []string
		wantAffects      []Diff
		doNotWantAffects []Diff
	}{{
		name:             "None",
		files:            []string{"asdf"},
		wantAffects:      []Diff{None},
		doNotWantAffects: []Diff{Go, Client, DatabaseSchema, All},
	}, {
		name:             "Go",
		files:            []string{"main.go", "func.go"},
		wantAffects:      []Diff{Go},
		doNotWantAffects: []Diff{Client},
	}, {
		name:             "DB schema implies Go and DB schema diff",
		files:            []string{"migrations/file1", "migrations/file2"},
		wantAffects:      []Diff{Go, DatabaseSchema},
		doNotWantAffects: []Diff{Client},
	}, {
		name:             "Or",
		files:            []string{"client/file1", "file2.graphql"},
		wantAffects:      []Diff{Client | GraphQL},
		doNotWantAffects: []Diff{Go},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diff := ParseDiff(tt.files)
			for _, want := range tt.wantAffects {
				assert.True(t, diff.Has(want))
			}
			for _, doNotWant := range tt.doNotWantAffects {
				assert.False(t, diff.Has(doNotWant))
			}
		})
	}
}

func TestDiffString(t *testing.T) {
	// Check all individual diff types have a name defined at least
	var lastName string
	for diff := Go; diff <= All; diff <<= 1 {
		assert.NotEmpty(t, diff.String(), "%d", diff)
		lastName = diff.String()
	}
	assert.Equal(t, lastName, "All")

	// Check specific names
	tests := []struct {
		name string
		diff Diff
		want string
	}{{
		name: "None",
		diff: None,
		want: "None",
	}, {
		name: "All",
		diff: All,
		want: "All",
	}, {
		name: "One diff",
		diff: Go,
		want: "Go",
	}, {
		name: "Multiple diffs",
		diff: Go | DatabaseSchema | Client,
		want: "Go, Client, DatabaseSchema",
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.diff.String())
		})
	}
}
