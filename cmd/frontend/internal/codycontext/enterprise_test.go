package codycontext

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/sourcegraph/sourcegraph/internal/featureflag"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/schema"
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

	newFF := func(v bool) *featureflag.FeatureFlag {
		return &featureflag.FeatureFlag{
			Name:      "cody-context-filters-enabled",
			Bool:      &featureflag.FeatureFlagBool{Value: v},
			Rollout:   nil,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			DeletedAt: nil,
		}
	}
	featureFlagValues := []*featureflag.FeatureFlag{newFF(true), newFF(false), nil}

	for _, tt := range tests {
		for _, ff := range featureFlagValues {
			name := tt.name
			if ff != nil {
				name = name + fmt.Sprintf(" (%q feature flag value: %t)", ff.Name, ff.Bool.Value)
			}
			t.Run(name, func(t *testing.T) {
				conf.Mock(&conf.Unified{
					SiteConfiguration: schema.SiteConfiguration{
						CodyContextFilters: tt.ccf,
					},
				})

				featureFlags := dbmocks.NewMockFeatureFlagStore()
				if ff != nil {
					featureFlags.GetFeatureFlagFunc.SetDefaultReturn(ff, nil)
				}
				db := dbmocks.NewMockDB()
				db.FeatureFlagsFunc.SetDefaultReturn(featureFlags)

				f := newEnterpriseFilter(logtest.Scoped(t), db)
				allowedRepos, matcher, _ := f.getMatcher(context.Background(), tt.repos)
				filtered := make([]FileChunkContext, 0, len(tt.chunks))
				for _, chunk := range tt.chunks {
					if matcher(chunk.RepoID, chunk.Path) {
						filtered = append(filtered, chunk)
					}
				}

				if ff != nil && ff.Bool.Value {
					require.Equal(t, tt.wantRepos, allowedRepos)
					require.Equal(t, tt.wantChunks, filtered)
				} else {
					// If feature flag is not set or is set to false, the Cody context filters are disabled.
					require.Equal(t, tt.repos, tt.repos)
					require.Equal(t, tt.wantChunks, tt.wantChunks)
				}
			})
		}
	}
}
