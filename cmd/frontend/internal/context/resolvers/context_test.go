package resolvers

import (
	"context"
	"io/fs"
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	codycontext "github.com/sourcegraph/sourcegraph/internal/codycontext"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/embeddings"
	"github.com/sourcegraph/sourcegraph/internal/featureflag"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/client"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestContextResolver(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Background()

	db := database.NewDB(logger, dbtest.NewDB(t))
	repo1 := types.Repo{Name: "repo1"}
	repo2 := types.Repo{Name: "repo2"}
	// Create populates the IDs in the passed in types.Repo
	err := db.Repos().Create(ctx, &repo1, &repo2)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, "INSERT INTO repo_embedding_jobs (state, repo_id, revision) VALUES ('completed', $1, 'HEAD');", int32(repo1.ID))
	require.NoError(t, err)

	files := map[api.RepoName]map[string][]byte{
		"repo1": {
			"testcode1.go": []byte("testcode1"),
			"testtext1.md": []byte("testtext1"),
		},
		"repo2": {
			"testcode2.go": []byte("testcode2"),
			"testtext2.md": []byte("testtext2"),
		},
	}

	mockGitserver := gitserver.NewMockClient()
	mockGitserver.StatFunc.SetDefaultHook(func(_ context.Context, repo api.RepoName, _ api.CommitID, fileName string) (fs.FileInfo, error) {
		return fakeFileInfo{path: fileName}, nil
	})
	mockGitserver.ReadFileFunc.SetDefaultHook(func(_ context.Context, repo api.RepoName, _ api.CommitID, fileName string) ([]byte, error) {
		if content, ok := files[repo][fileName]; ok {
			return content, nil
		}
		return nil, os.ErrNotExist
	})

	mockEmbeddingsClient := embeddings.NewMockClient()
	mockEmbeddingsClient.SearchFunc.SetDefaultHook(func(_ context.Context, params embeddings.EmbeddingsSearchParameters) (*embeddings.EmbeddingCombinedSearchResults, error) {
		require.Equal(t, params.RepoNames, []api.RepoName{"repo1"})
		require.Equal(t, params.TextResultsCount, 1)
		require.Equal(t, params.CodeResultsCount, 1)
		return &embeddings.EmbeddingCombinedSearchResults{
			CodeResults: embeddings.EmbeddingSearchResults{{
				FileName: "testcode1.go",
			}},
			TextResults: embeddings.EmbeddingSearchResults{{
				FileName: "testtext1.md",
			}},
		}, nil
	})

	lineRange := func(start, end int) result.ChunkMatches {
		return result.ChunkMatches{{
			Ranges: result.Ranges{{
				Start: result.Location{Line: start},
				End:   result.Location{Line: end},
			}},
		}}
	}

	mockSearchClient := client.NewMockSearchClient()
	mockSearchClient.PlanFunc.SetDefaultHook(func(_ context.Context, _ string, _ *string, query string, _ search.Mode, _ search.Protocol) (*search.Inputs, error) {
		return &search.Inputs{OriginalQuery: query}, nil
	})
	mockSearchClient.ExecuteFunc.SetDefaultHook(func(_ context.Context, stream streaming.Sender, inputs *search.Inputs) (*search.Alert, error) {
		if strings.Contains(inputs.OriginalQuery, "-file") {
			stream.Send(streaming.SearchEvent{
				Results: result.Matches{&result.FileMatch{
					File: result.File{
						Path: "testcode2.go",
						Repo: types.MinimalRepo{ID: repo2.ID, Name: repo2.Name},
					},
					ChunkMatches: lineRange(0, 4),
				}, &result.FileMatch{
					File: result.File{
						Path: "testcode2again.go",
						Repo: types.MinimalRepo{ID: repo2.ID, Name: repo2.Name},
					},
					ChunkMatches: lineRange(0, 4),
				}},
			})
		} else {
			stream.Send(streaming.SearchEvent{
				Results: result.Matches{&result.FileMatch{
					File: result.File{
						Path: "testtext2.md",
						Repo: types.MinimalRepo{ID: repo2.ID, Name: repo2.Name},
					},
					ChunkMatches: lineRange(0, 4),
				}},
			})
		}
		return nil, nil
	})

	contextClient := codycontext.NewCodyContextClient(
		observation.NewContext(logger),
		db,
		mockEmbeddingsClient,
		mockSearchClient,
		nil,
	)

	resolver := NewResolver(
		db,
		mockGitserver,
		contextClient,
	)

	truePtr := true
	conf.Mock(&conf.Unified{
		SiteConfiguration: schema.SiteConfiguration{
			CodyEnabled: &truePtr,
		},
	})

	ctx = actor.WithActor(ctx, actor.FromMockUser(1))
	ffs := featureflag.NewMemoryStore(map[string]bool{"cody": true}, nil, nil)
	ctx = featureflag.WithFlags(ctx, ffs)

	results, err := resolver.GetCodyContext(ctx, graphqlbackend.GetContextArgs{
		Repos:            graphqlbackend.MarshalRepositoryIDs([]api.RepoID{1, 2}),
		Query:            "my test query",
		TextResultsCount: 2,
		CodeResultsCount: 2,
	})
	require.NoError(t, err)

	paths := make([]string, len(results))
	for i, result := range results {
		paths[i] = result.(*graphqlbackend.FileChunkContextResolver).Blob().Path()
	}
	// One code result and text result from each repo
	expected := []string{"testcode1.go", "testtext1.md", "testcode2.go", "testtext2.md"}
	sort.Strings(expected)
	sort.Strings(paths)
	require.Equal(t, expected, paths)
}

type fakeFileInfo struct {
	path string
	fs.FileInfo
}

func (f fakeFileInfo) Name() string {
	return f.path
}
