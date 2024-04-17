package codycontext

import (
	"context"
	"encoding/json"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"testing"

	"github.com/sourcegraph/log/logtest"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/schema"
	"github.com/stretchr/testify/require"
)

func TestNewEnterpriseFilter(t *testing.T) {
	t.Cleanup(func() { conf.Mock(nil) })

	tests := []struct {
		name       string
		ccf        *schema.CodyContextFilters
		repos      []types.RepoIDName
		chunks     []FileChunkContext
		wantRepos  []types.RepoIDName
		wantChunks []FileChunkContext
	}{
		{
			name: "Cody context filters not set",
			ccf:  nil,
			repos: []types.RepoIDName{
				{Name: "github.com/sourcegraph/about", ID: 1},
				{Name: "github.com/sourcegraph/annotate", ID: 2},
				{Name: "github.com/sourcegraph/sourcegraph", ID: 3},
				{Name: "github.com/docker/compose", ID: 4},
			},
			chunks: []FileChunkContext{
				{
					RepoName: "github.com/sourcegraph/about",
					RepoID:   1,
					Path:     "/file1.go",
				},
				{
					RepoName: "github.com/sourcegraph/annotate",
					RepoID:   2,
					Path:     "/file2.go",
				},
				{
					RepoName: "github.com/sourcegraph/sourcegraph",
					RepoID:   3,
					Path:     "/file3.go",
				},
				{
					RepoName: "github.com/docker/compose",
					RepoID:   4,
					Path:     "/file4.go",
				},
			},
			wantRepos: []types.RepoIDName{
				{Name: "github.com/sourcegraph/about", ID: 1},
				{Name: "github.com/sourcegraph/annotate", ID: 2},
				{Name: "github.com/sourcegraph/sourcegraph", ID: 3},
				{Name: "github.com/docker/compose", ID: 4},
			},
			wantChunks: []FileChunkContext{
				{
					RepoName: "github.com/sourcegraph/about",
					RepoID:   1,
					Path:     "/file1.go",
				},
				{
					RepoName: "github.com/sourcegraph/annotate",
					RepoID:   2,
					Path:     "/file2.go",
				},
				{
					RepoName: "github.com/sourcegraph/sourcegraph",
					RepoID:   3,
					Path:     "/file3.go",
				},
				{
					RepoName: "github.com/docker/compose",
					RepoID:   4,
					Path:     "/file4.go",
				},
			},
		},
		{
			// This scenario shouldn't happen.
			// "cody.contextFilters" if defined in the site config, should have at least one property.
			// Thus, either "include" or "exclude" should be defined.
			// We rely on site config schema validation.
			name: "include and exclude rules are not defined",
			ccf:  &schema.CodyContextFilters{},
			repos: []types.RepoIDName{
				{Name: "github.com/sourcegraph/about", ID: 1},
				{Name: "github.com/sourcegraph/annotate", ID: 2},
				{Name: "github.com/sourcegraph/sourcegraph", ID: 3},
				{Name: "github.com/docker/compose", ID: 4},
			},
			chunks: []FileChunkContext{
				{
					RepoName: "github.com/sourcegraph/about",
					RepoID:   1,
					Path:     "/file1.go",
				},
				{
					RepoName: "github.com/sourcegraph/annotate",
					RepoID:   2,
					Path:     "/file2.go",
				},
				{
					RepoName: "github.com/sourcegraph/sourcegraph",
					RepoID:   3,
					Path:     "/file3.go",
				},
				{
					RepoName: "github.com/docker/compose",
					RepoID:   4,
					Path:     "/file4.go",
				},
			},
			wantRepos: []types.RepoIDName{
				{Name: "github.com/sourcegraph/about", ID: 1},
				{Name: "github.com/sourcegraph/annotate", ID: 2},
				{Name: "github.com/sourcegraph/sourcegraph", ID: 3},
				{Name: "github.com/docker/compose", ID: 4},
			},
			wantChunks: []FileChunkContext{
				{
					RepoName: "github.com/sourcegraph/about",
					RepoID:   1,
					Path:     "/file1.go",
				},
				{
					RepoName: "github.com/sourcegraph/annotate",
					RepoID:   2,
					Path:     "/file2.go",
				},
				{
					RepoName: "github.com/sourcegraph/sourcegraph",
					RepoID:   3,
					Path:     "/file3.go",
				},
				{
					RepoName: "github.com/docker/compose",
					RepoID:   4,
					Path:     "/file4.go",
				},
			},
		},
		{
			// This scenario shouldn't happen. If either "include" or "exclude" field is defined, it should have at least one item.
			// We rely on site config schema validation.
			name: "include and exclude rules empty",
			ccf: &schema.CodyContextFilters{
				Include: []*schema.CodyContextFilterItem{},
				Exclude: []*schema.CodyContextFilterItem{},
			},
			repos: []types.RepoIDName{
				{Name: "github.com/sourcegraph/about", ID: 1},
				{Name: "github.com/sourcegraph/annotate", ID: 2},
				{Name: "github.com/sourcegraph/sourcegraph", ID: 3},
				{Name: "github.com/docker/compose", ID: 4},
			},
			chunks: []FileChunkContext{
				{
					RepoName: "github.com/sourcegraph/about",
					RepoID:   1,
					Path:     "/file1.go",
				},
				{
					RepoName: "github.com/sourcegraph/annotate",
					RepoID:   2,
					Path:     "/file2.go",
				},
				{
					RepoName: "github.com/sourcegraph/sourcegraph",
					RepoID:   3,
					Path:     "/file3.go",
				},
				{
					RepoName: "github.com/docker/compose",
					RepoID:   4,
					Path:     "/file4.go",
				},
			},
			wantRepos: []types.RepoIDName{
				{Name: "github.com/sourcegraph/about", ID: 1},
				{Name: "github.com/sourcegraph/annotate", ID: 2},
				{Name: "github.com/sourcegraph/sourcegraph", ID: 3},
				{Name: "github.com/docker/compose", ID: 4},
			},
			wantChunks: []FileChunkContext{
				{
					RepoName: "github.com/sourcegraph/about",
					RepoID:   1,
					Path:     "/file1.go",
				},
				{
					RepoName: "github.com/sourcegraph/annotate",
					RepoID:   2,
					Path:     "/file2.go",
				},
				{
					RepoName: "github.com/sourcegraph/sourcegraph",
					RepoID:   3,
					Path:     "/file3.go",
				},
				{
					RepoName: "github.com/docker/compose",
					RepoID:   4,
					Path:     "/file4.go",
				},
			},
		},
		{
			name: "only include rules defined",
			ccf: &schema.CodyContextFilters{
				Include: []*schema.CodyContextFilterItem{
					{RepoNamePattern: "^github\\.com\\/sourcegraph\\/.+"},
				},
			},
			repos: []types.RepoIDName{
				{Name: "github.com/sourcegraph/about", ID: 1},
				{Name: "github.com/sourcegraph/annotate", ID: 2},
				{Name: "github.com/sourcegraph/sourcegraph", ID: 3},
				{Name: "github.com/docker/compose", ID: 4},
			},
			chunks: []FileChunkContext{
				{
					RepoName: "github.com/sourcegraph/about",
					RepoID:   1,
					Path:     "/file1.go",
				},
				{
					RepoName: "github.com/sourcegraph/annotate",
					RepoID:   2,
					Path:     "/file2.go",
				},
				{
					RepoName: "github.com/sourcegraph/sourcegraph",
					RepoID:   3,
					Path:     "/file3.go",
				},
				{
					RepoName: "github.com/docker/compose",
					RepoID:   4,
					Path:     "/file4.go",
				},
			},
			wantRepos: []types.RepoIDName{
				{Name: "github.com/sourcegraph/about", ID: 1},
				{Name: "github.com/sourcegraph/annotate", ID: 2},
				{Name: "github.com/sourcegraph/sourcegraph", ID: 3},
			},
			wantChunks: []FileChunkContext{
				{
					RepoName: "github.com/sourcegraph/about",
					RepoID:   1,
					Path:     "/file1.go",
				},
				{
					RepoName: "github.com/sourcegraph/annotate",
					RepoID:   2,
					Path:     "/file2.go",
				},
				{
					RepoName: "github.com/sourcegraph/sourcegraph",
					RepoID:   3,
					Path:     "/file3.go",
				},
			},
		},
		{
			name: "only exclude rules defined",
			ccf: &schema.CodyContextFilters{
				Exclude: []*schema.CodyContextFilterItem{
					{RepoNamePattern: "^github\\.com\\/sourcegraph\\/.+"},
				},
			},
			repos: []types.RepoIDName{
				{Name: "github.com/sourcegraph/about", ID: 1},
				{Name: "github.com/sourcegraph/annotate", ID: 2},
				{Name: "github.com/sourcegraph/sourcegraph", ID: 3},
				{Name: "github.com/docker/compose", ID: 4},
			},
			chunks: []FileChunkContext{
				{
					RepoName: "github.com/sourcegraph/about",
					RepoID:   1,
					Path:     "/file1.go",
				},
				{
					RepoName: "github.com/sourcegraph/annotate",
					RepoID:   2,
					Path:     "/file2.go",
				},
				{
					RepoName: "github.com/sourcegraph/sourcegraph",
					RepoID:   3,
					Path:     "/file3.go",
				},
				{
					RepoName: "github.com/docker/compose",
					RepoID:   4,
					Path:     "/file4.go",
				},
			},
			wantRepos: []types.RepoIDName{
				{Name: "github.com/docker/compose", ID: 4},
			},
			wantChunks: []FileChunkContext{
				{
					RepoName: "github.com/docker/compose",
					RepoID:   4,
					Path:     "/file4.go",
				},
			},
		},
		{
			name: "include and exclude rules defined",
			ccf: &schema.CodyContextFilters{
				Include: []*schema.CodyContextFilterItem{
					{RepoNamePattern: "^github\\.com\\/sourcegraph\\/.+"},
				},
				Exclude: []*schema.CodyContextFilterItem{
					{RepoNamePattern: ".*cody.*"},
				},
			},
			repos: []types.RepoIDName{
				{Name: "github.com/sourcegraph/about", ID: 1},
				{Name: "github.com/sourcegraph/annotate", ID: 2},
				{Name: "github.com/sourcegraph/sourcegraph", ID: 3},
				{Name: "github.com/docker/compose", ID: 4},
				{Name: "github.com/sourcegraph/cody", ID: 5},
			},
			chunks: []FileChunkContext{
				{
					RepoName: "github.com/sourcegraph/about",
					RepoID:   1,
					Path:     "/file1.go",
				},
				{
					RepoName: "github.com/sourcegraph/annotate",
					RepoID:   2,
					Path:     "/file2.go",
				},
				{
					RepoName: "github.com/sourcegraph/sourcegraph",
					RepoID:   3,
					Path:     "/file3.go",
				},
				{
					RepoName: "github.com/docker/compose",
					RepoID:   4,
					Path:     "/file4.go",
				},
				{
					RepoName: "github.com/sourcegraph/cody",
					RepoID:   5,
					Path:     "/index.ts",
				},
			},
			wantRepos: []types.RepoIDName{
				{Name: "github.com/sourcegraph/about", ID: 1},
				{Name: "github.com/sourcegraph/annotate", ID: 2},
				{Name: "github.com/sourcegraph/sourcegraph", ID: 3},
			},
			wantChunks: []FileChunkContext{
				{
					RepoName: "github.com/sourcegraph/about",
					RepoID:   1,
					Path:     "/file1.go",
				},
				{
					RepoName: "github.com/sourcegraph/annotate",
					RepoID:   2,
					Path:     "/file2.go",
				},
				{
					RepoName: "github.com/sourcegraph/sourcegraph",
					RepoID:   3,
					Path:     "/file3.go",
				},
			},
		},
		{
			name: "multiple include and exclude rules defined",
			ccf: &schema.CodyContextFilters{
				Include: []*schema.CodyContextFilterItem{
					{RepoNamePattern: "^github\\.com\\/sourcegraph\\/.+"},
					{RepoNamePattern: "^github\\.com\\/docker\\/compose$"},
					{RepoNamePattern: "^github\\.com\\/.+\\/react"},
				},
				Exclude: []*schema.CodyContextFilterItem{
					{RepoNamePattern: ".*cody.*"},
					{RepoNamePattern: ".+\\/docker\\/.+"},
				},
			},
			repos: []types.RepoIDName{
				{Name: "github.com/sourcegraph/about", ID: 1},
				{Name: "github.com/sourcegraph/annotate", ID: 2},
				{Name: "github.com/sourcegraph/sourcegraph", ID: 3},
				{Name: "github.com/docker/compose", ID: 4},
				{Name: "github.com/sourcegraph/cody", ID: 5},
				{Name: "github.com/facebook/react", ID: 6},
			},
			chunks: []FileChunkContext{
				{
					RepoName: "github.com/sourcegraph/about",
					RepoID:   1,
					Path:     "/file1.go",
				},
				{
					RepoName: "github.com/sourcegraph/annotate",
					RepoID:   2,
					Path:     "/file2.go",
				},
				{
					RepoName: "github.com/sourcegraph/sourcegraph",
					RepoID:   3,
					Path:     "/file3.go",
				},
				{
					RepoName: "github.com/docker/compose",
					RepoID:   4,
					Path:     "/file4.go",
				},
				{
					RepoName: "github.com/sourcegraph/cody",
					RepoID:   5,
					Path:     "/index.ts",
				},
				{
					RepoName: "github.com/facebook/react",
					RepoID:   6,
					Path:     "/hooks.ts",
				},
			},
			wantRepos: []types.RepoIDName{
				{Name: "github.com/sourcegraph/about", ID: 1},
				{Name: "github.com/sourcegraph/annotate", ID: 2},
				{Name: "github.com/sourcegraph/sourcegraph", ID: 3},
				{Name: "github.com/facebook/react", ID: 6},
			},
			wantChunks: []FileChunkContext{
				{
					RepoName: "github.com/sourcegraph/about",
					RepoID:   1,
					Path:     "/file1.go",
				},
				{
					RepoName: "github.com/sourcegraph/annotate",
					RepoID:   2,
					Path:     "/file2.go",
				},
				{
					RepoName: "github.com/sourcegraph/sourcegraph",
					RepoID:   3,
					Path:     "/file3.go",
				},
				{
					RepoName: "github.com/facebook/react",
					RepoID:   6,
					Path:     "/hooks.ts",
				},
			},
		},
		{
			name: "exclude everything",
			ccf: &schema.CodyContextFilters{
				Include: []*schema.CodyContextFilterItem{
					{RepoNamePattern: "^github\\.com\\/sourcegraph\\/.+"},
					{RepoNamePattern: "^github\\.com\\/docker\\/compose$"},
					{RepoNamePattern: "^github\\.com\\/.+\\/react"},
				},
				Exclude: []*schema.CodyContextFilterItem{
					{RepoNamePattern: ".*cody.*"},
					{RepoNamePattern: ".*"},
				},
			},
			repos: []types.RepoIDName{
				{Name: "github.com/sourcegraph/about", ID: 1},
				{Name: "github.com/sourcegraph/annotate", ID: 2},
				{Name: "github.com/sourcegraph/sourcegraph", ID: 3},
				{Name: "github.com/docker/compose", ID: 4},
				{Name: "github.com/sourcegraph/cody", ID: 5},
				{Name: "github.com/facebook/react", ID: 6},
			},
			chunks: []FileChunkContext{
				{
					RepoName: "github.com/sourcegraph/about",
					RepoID:   1,
					Path:     "/file1.go",
				},
				{
					RepoName: "github.com/sourcegraph/annotate",
					RepoID:   2,
					Path:     "/file2.go",
				},
				{
					RepoName: "github.com/sourcegraph/sourcegraph",
					RepoID:   3,
					Path:     "/file3.go",
				},
				{
					RepoName: "github.com/docker/compose",
					RepoID:   4,
					Path:     "/file4.go",
				},
				{
					RepoName: "github.com/sourcegraph/cody",
					RepoID:   5,
					Path:     "/index.ts",
				},
				{
					RepoName: "github.com/facebook/react",
					RepoID:   6,
					Path:     "/hooks.ts",
				},
			},
			wantRepos:  []types.RepoIDName{},
			wantChunks: []FileChunkContext{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conf.Mock(&conf.Unified{
				SiteConfiguration: schema.SiteConfiguration{
					CodyContextFilters: tt.ccf,
				},
			})

			f := newEnterpriseFilter(logtest.Scoped(t))
			allowedRepos, filter, _ := f.GetFilter(context.Background(), tt.repos)

			require.Equal(t, tt.wantRepos, allowedRepos)
			require.Equal(t, tt.wantChunks, filter(tt.chunks))
		})
	}
}

func TestFiltersConfig(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	require.Equal(t, true, ok)
	content, err := os.ReadFile(filepath.Join(filepath.Dir(file), "enterprise_test_data.json"))
	require.NoError(t, err)

	type repo struct {
		Name api.RepoName
		Id   api.RepoID `json:"id"`
	}
	type testCase struct {
		Name             string                     `json:"name"`
		Description      string                     `json:"description"`
		IncludeByDefault bool                       `json:"includeByDefault"`
		IncludeUnknown   bool                       `json:"includeUnknown"`
		Ccf              *schema.CodyContextFilters `json:"cody.contextFilters"`
		Repos            []repo                     `json:"repos"`
		IncludeRepos     []repo                     `json:"includeRepos"`
	}
	var data struct {
		TestCases []testCase `json:"testCases"`
	}

	err = json.Unmarshal(content, &data)
	require.NoError(t, err)

	for _, tt := range data.TestCases {
		t.Run(tt.Name, func(t *testing.T) {
			fc, _ := parseCodyContextFilters(tt.Ccf)
			n := 0
			for _, r := range tt.Repos {
				want := slices.ContainsFunc(tt.IncludeRepos, func(p repo) bool { return r.Id == p.Id })
				got := fc.isRepoAllowed(types.RepoIDName{ID: r.Id, Name: r.Name})
				if got {
					n++
				}
				require.Equal(t, want, got)
			}

			require.Equal(t, len(tt.IncludeRepos), n)
		})
	}
}
