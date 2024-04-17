package codycontext

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/sourcegraph/log/logtest"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/schema"
	"github.com/stretchr/testify/require"
)

func TestNewEnterpriseFilter(t *testing.T) {
	t.Cleanup(func() { conf.Mock(nil) })
	logger := logtest.Scoped(t)

	_, file, _, ok := runtime.Caller(0)
	require.Equal(t, true, ok)
	content, err := os.ReadFile(filepath.Join(filepath.Dir(file), "enterprise_test_data.json"))
	require.NoError(t, err)

	type repo struct {
		Name api.RepoName
		Id   api.RepoID `json:"id"`
	}
	type fileChunk struct {
		Repo repo
		Path string
	}
	type testCase struct {
		Name              string                     `json:"name"`
		Description       string                     `json:"description"`
		IncludeByDefault  bool                       `json:"includeByDefault"`
		IncludeUnknown    bool                       `json:"includeUnknown"`
		Ccf               *schema.CodyContextFilters `json:"cody.contextFilters"`
		Repos             []repo                     `json:"repos"`
		IncludeRepos      []repo                     `json:"includeRepos"`
		FileChunks        []fileChunk                `json:"fileChunks"`
		IncludeFileChunks []fileChunk                `json:"includeFileChunks"`
	}
	var data struct {
		TestCases []testCase `json:"testCases"`
	}

	err = json.Unmarshal(content, &data)
	require.NoError(t, err)

	toRepoIDName := func(r repo) types.RepoIDName { return types.RepoIDName{ID: r.Id, Name: r.Name} }

	normalizeRepos := func(repos []repo) []types.RepoIDName {
		result := make([]types.RepoIDName, 0, len(repos))
		for _, r := range repos {
			result = append(result, toRepoIDName(r))
		}
		return result
	}

	normalizeFileChunks := func(fcc []fileChunk) []FileChunkContext {
		result := make([]FileChunkContext, 0, len(fcc))
		for _, fc := range fcc {
			r := toRepoIDName(fc.Repo)
			result = append(result, FileChunkContext{RepoName: r.Name, RepoID: r.ID, Path: fc.Path})
		}
		return result
	}

	for _, tt := range data.TestCases {
		t.Run(tt.Name, func(t *testing.T) {
			conf.Mock(&conf.Unified{
				SiteConfiguration: schema.SiteConfiguration{
					CodyContextFilters: tt.Ccf,
				},
			})

			f := newEnterpriseFilter(logger)

			var repos []types.RepoIDName
			for _, r := range tt.Repos {
				repos = append(repos, types.RepoIDName{ID: r.Id, Name: r.Name})
			}

			allowedRepos, filter, _ := f.GetFilter(context.Background(), normalizeRepos(tt.Repos))
			require.Equal(t, normalizeRepos(tt.IncludeRepos), allowedRepos)
			require.Equal(t, normalizeFileChunks(tt.IncludeFileChunks), filter(normalizeFileChunks(tt.FileChunks)))
		})
	}
}
