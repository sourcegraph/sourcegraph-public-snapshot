package changed

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestForEachDiffType(t *testing.T) {
	var first, last Diff
	ForEachDiffType(func(d Diff) {
		if first == 0 {
			first = d
		}
		last = d
	})
	assert.Equal(t, Diff(1<<1), first, "iteration start")
	assert.Equal(t, All, last<<1, "iteration end")
}

func TestParseDiff(t *testing.T) {
	t.Run("All", func(t *testing.T) {
		assert.False(t, All.Has(None))
		assert.True(t, All.Has(All))
		ForEachDiffType(func(d Diff) {
			assert.True(t, All.Has(d))
		})
	})

	tests := []struct {
		name             string
		files            []string
		wantAffects      []Diff
		wantChangedFiles ChangedFiles
		doNotWantAffects []Diff
	}{
		{
			name:             "None",
			files:            []string{"asdf"},
			wantAffects:      []Diff{None},
			wantChangedFiles: make(ChangedFiles),
			doNotWantAffects: []Diff{Go, Client, DatabaseSchema, All},
		}, {
			name:             "Go",
			files:            []string{"main.go", "func.go"},
			wantAffects:      []Diff{Go},
			wantChangedFiles: make(ChangedFiles),
			doNotWantAffects: []Diff{Client, All},
		}, {
			name:             "go testdata",
			files:            []string{"internal/cmd/search-blitz/queries.txt"},
			wantAffects:      []Diff{Go},
			wantChangedFiles: make(ChangedFiles),
			doNotWantAffects: []Diff{Client, All},
		}, {
			name:             "DB schema implies Go and DB schema diff",
			files:            []string{"migrations/file1", "migrations/file2"},
			wantAffects:      []Diff{Go, DatabaseSchema},
			wantChangedFiles: make(ChangedFiles),
			doNotWantAffects: []Diff{Client, All},
		}, {
			name:             "Or",
			files:            []string{"client/file1", "file2.graphql"},
			wantAffects:      []Diff{Client, GraphQL},
			wantChangedFiles: make(ChangedFiles),
			doNotWantAffects: []Diff{Go, Pnpm, All},
		}, {
			name:        "Wolfi",
			files:       []string{"wolfi-images/image-test.yaml", "wolfi-packages/package-test.yaml"},
			wantAffects: []Diff{WolfiBaseImages, WolfiPackages},
			wantChangedFiles: ChangedFiles{
				WolfiBaseImages: []string{"wolfi-images/image-test.yaml"},
				WolfiPackages:   []string{"wolfi-packages/package-test.yaml"},
			},
			doNotWantAffects: []Diff{},
		}, {
			name:             "Protobuf definitions",
			files:            []string{"cmd/searcher/messages.proto"},
			wantAffects:      []Diff{Protobuf},
			wantChangedFiles: make(ChangedFiles),
			doNotWantAffects: []Diff{},
		}, {
			name:             "Protobuf generated code",
			files:            []string{"cmd/searcher/messages.pb.go"},
			wantAffects:      []Diff{Protobuf, Go},
			wantChangedFiles: make(ChangedFiles),
			doNotWantAffects: []Diff{},
		}, {
			name:             "Buf CLI module configuration",
			files:            []string{"cmd/searcher/buf.yaml"},
			wantAffects:      []Diff{Protobuf},
			wantChangedFiles: make(ChangedFiles),
			doNotWantAffects: []Diff{},
		}, {
			name:             "Buf CLI generated code configuration",
			files:            []string{"cmd/searcher/buf.gen.yaml"},
			wantAffects:      []Diff{Protobuf},
			wantChangedFiles: make(ChangedFiles),
			doNotWantAffects: []Diff{},
		}, {
			name:             "PNPM file changes",
			files:            []string{"pnpm-workspace.yaml"},
			wantAffects:      []Diff{Pnpm, Client},
			wantChangedFiles: make(ChangedFiles),
			doNotWantAffects: []Diff{},
		}, {
			name:             "PNPM sub-package file changes",
			files:            []string{"client/package.json"},
			wantAffects:      []Diff{Pnpm, Client},
			wantChangedFiles: make(ChangedFiles),
			doNotWantAffects: []Diff{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diff, changedFiles := ParseDiff(tt.files)
			for _, want := range tt.wantAffects {
				assert.True(t, diff.Has(want))
			}
			for _, doNotWant := range tt.doNotWantAffects {
				assert.False(t, diff.Has(doNotWant))
			}
			if !reflect.DeepEqual(changedFiles, tt.wantChangedFiles) {
				t.Errorf("wantedChangedFiles not equal:\nGot: %+v\nWant: %+v\n", changedFiles, tt.wantChangedFiles)
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
