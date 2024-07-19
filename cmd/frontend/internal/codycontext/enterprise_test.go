package codycontext

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/api"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestNewEnterpriseFilter(t *testing.T) {
	t.Cleanup(func() {
		conf.Mock(nil)
	})

	content, err := os.ReadFile(filepath.Join("testdata", "enterprise_filter_test_data.json"))
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
		Name               string                     `json:"name"`
		Description        string                     `json:"description"`
		Ccf                *schema.CodyContextFilters `json:"cody.contextFilters"`
		Repos              []repo                     `json:"repos"`
		IncludedRepos      []repo                     `json:"includedRepos"`
		FileChunks         []fileChunk                `json:"fileChunks"`
		IncludedFileChunks []fileChunk                `json:"includedFileChunks"`
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

			f := newEnterpriseFilter(logtest.Scoped(t))
			includedRepos, matcher, _ := f.getMatcher(context.Background(), normalizeRepos(tt.Repos))
			includedFileChunks := make([]fileChunk, 0, len(tt.FileChunks))
			for _, chunk := range tt.FileChunks {
				if matcher(chunk.Repo.Id, chunk.Path) {
					includedFileChunks = append(includedFileChunks, chunk)
				}
			}

			require.Equal(t, normalizeRepos(tt.IncludedRepos), includedRepos)
			require.Equal(t, normalizeFileChunks(tt.IncludedFileChunks), normalizeFileChunks(includedFileChunks))
		})
	}
}
