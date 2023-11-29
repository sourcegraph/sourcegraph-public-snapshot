package resolvers

import (
	"context"
	"os"
	"testing"

	"github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/sourcegraph/sourcegraph/internal/embeddings"
	"github.com/sourcegraph/sourcegraph/internal/embeddings/background/repo"
	"github.com/sourcegraph/sourcegraph/internal/featureflag"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/licensing"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestEmbeddingSearchResolver(t *testing.T) {
	logger := logtest.Scoped(t)

	oldMock := licensing.MockCheckFeature
	licensing.MockCheckFeature = func(feature licensing.Feature) error {
		return nil
	}
	t.Cleanup(func() {
		licensing.MockCheckFeature = oldMock
	})

	mockDB := dbmocks.NewMockDB()
	mockRepos := dbmocks.NewMockRepoStore()
	mockRepos.GetByIDsFunc.SetDefaultReturn([]*types.Repo{{ID: 1, Name: "repo1"}}, nil)
	mockDB.ReposFunc.SetDefaultReturn(mockRepos)

	mockGitserver := gitserver.NewMockClient()
	mockGitserver.ReadFileFunc.SetDefaultHook(func(_ context.Context, _ api.RepoName, _ api.CommitID, fileName string) ([]byte, error) {
		if fileName == "testfile" {
			return []byte("test\nfirst\nfour\nlines\nplus\nsome\nmore"), nil
		}
		return nil, os.ErrNotExist
	})

	mockEmbeddingsClient := embeddings.NewMockClient()
	mockEmbeddingsClient.SearchFunc.SetDefaultReturn(&embeddings.EmbeddingCombinedSearchResults{
		CodeResults: embeddings.EmbeddingSearchResults{{
			FileName:  "testfile",
			StartLine: 0,
			EndLine:   4,
		}, {
			FileName:  "censored",
			StartLine: 0,
			EndLine:   4,
		}},
	}, nil)

	repoEmbeddingJobsStore := repo.NewMockRepoEmbeddingJobsStore()

	resolver := NewResolver(
		mockDB,
		logger,
		mockGitserver,
		mockEmbeddingsClient,
		repoEmbeddingJobsStore,
	)

	conf.Mock(&conf.Unified{
		SiteConfiguration: schema.SiteConfiguration{
			CodyEnabled: pointers.Ptr(true),
			LicenseKey:  "asdf",
		},
	})

	ctx := actor.WithActor(context.Background(), actor.FromMockUser(1))
	ffs := featureflag.NewMemoryStore(map[string]bool{"cody": true}, nil, nil)
	ctx = featureflag.WithFlags(ctx, ffs)

	results, err := resolver.EmbeddingsMultiSearch(ctx, graphqlbackend.EmbeddingsMultiSearchInputArgs{
		Repos:            []graphql.ID{graphqlbackend.MarshalRepositoryID(3)},
		Query:            "test",
		CodeResultsCount: 1,
		TextResultsCount: 1,
	})
	require.NoError(t, err)

	codeResults, err := results.CodeResults(ctx)
	require.NoError(t, err)
	require.Len(t, codeResults, 1)
	require.Equal(t, "test\nfirst\nfour\nlines", codeResults[0].Content(ctx))
}

func Test_extractLineRange(t *testing.T) {
	cases := []struct {
		input      []byte
		start, end int
		output     []byte
	}{{
		[]byte("zero\none\ntwo\nthree\n"),
		0, 2,
		[]byte("zero\none"),
	}, {
		[]byte("zero\none\ntwo\nthree\n"),
		1, 2,
		[]byte("one"),
	}, {
		[]byte("zero\none\ntwo\nthree\n"),
		1, 2,
		[]byte("one"),
	}, {
		[]byte(""),
		1, 2,
		[]byte(""),
	}}

	for _, tc := range cases {
		t.Run("", func(t *testing.T) {
			got := extractLineRange(tc.input, tc.start, tc.end)
			require.Equal(t, tc.output, got)
		})
	}
}
