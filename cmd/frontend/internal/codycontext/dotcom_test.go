package codycontext

import (
	"context"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
	"github.com/sourcegraph/sourcegraph/schema"

	"github.com/stretchr/testify/require"
)

func TestNewDotcomFilter(t *testing.T) {
	repos := []types.RepoIDName{
		{ID: 1, Name: "repo1"},
		{ID: 2, Name: "repo2"},
	}
	conf.Mock(&conf.Unified{
		SiteConfiguration: schema.SiteConfiguration{
			ExperimentalFeatures: &schema.ExperimentalFeatures{
				CodyContextIgnore: pointers.Ptr(true),
			},
		},
	})
	logger := logtest.Scoped(t)
	t.Cleanup(func() { conf.Mock(nil) })

	t.Run("no ignore files", func(t *testing.T) {
		client := gitserver.NewMockClient()
		client.GetDefaultBranchFunc.SetDefaultReturn("main", api.CommitID("abc123"), nil)
		client.NewFileReaderFunc.SetDefaultReturn(nil, os.ErrNotExist)
		f := newDotcomFilter(logger, client)

		chunks := []FileChunkContext{
			{
				RepoName: "repo1",
				RepoID:   1,
				Path:     "/file1.go",
			},
			{
				RepoName: "repo2",
				RepoID:   2,
				Path:     "/file2.go",
			},
		}
		_, filter, _ := f.getMatcher(context.Background(), repos)
		filtered := filterChunks(chunks, filter)
		require.Equal(t, 2, len(filtered))
	})

	t.Run("filters multiple rules in ignore file", func(t *testing.T) {
		client := gitserver.NewMockClient()
		client.GetDefaultBranchFunc.SetDefaultReturn("main", api.CommitID("abc123"), nil)
		client.NewFileReaderFunc.SetDefaultHook(func(ctx context.Context, rn api.RepoName, ci api.CommitID, s string) (io.ReadCloser, error) {
			if rn == "repo2" {
				return io.NopCloser(strings.NewReader("**/file1.go\nsecret.txt")), nil
			}
			return nil, os.ErrNotExist
		})

		f := newDotcomFilter(logger, client)

		chunks := []FileChunkContext{
			{
				RepoName: "repo1",
				RepoID:   1,
				Path:     "file1.go",
			},
			{
				RepoName: "repo2",
				RepoID:   2,
				Path:     "folder1/file1.go",
			},
			{
				RepoName: "repo2",
				RepoID:   2,
				Path:     "folder1/folder2/file1.go",
			},
			{
				RepoName: "repo2",
				RepoID:   2,
				Path:     "secret.txt",
			},
		}

		_, filter, _ := f.getMatcher(context.Background(), repos)
		filtered := filterChunks(chunks, filter)
		require.Equal(t, 1, len(filtered))
		require.Equal(t, api.RepoName("repo1"), filtered[0].RepoName)
	})

	t.Run("uses correct ignore file by repo", func(t *testing.T) {
		client := gitserver.NewMockClient()
		client.GetDefaultBranchFunc.SetDefaultReturn("main", api.CommitID("abc123"), nil)
		client.NewFileReaderFunc.SetDefaultHook(func(ctx context.Context, rn api.RepoName, ci api.CommitID, s string) (io.ReadCloser, error) {
			switch rn {
			case "repo1":
				return io.NopCloser(strings.NewReader("**/file1.go")), nil
			case "repo2":
				return io.NopCloser(strings.NewReader("**/file2.go")), nil
			default:
				return nil, os.ErrNotExist
			}
		})
		f := newDotcomFilter(logger, client)

		chunks := []FileChunkContext{
			{
				RepoName: "repo1",
				RepoID:   1,
				Path:     "src/file1.go",
			},
			{
				RepoName: "repo1",
				RepoID:   1,
				Path:     "src/file2.go",
			},
			{
				RepoName: "repo2",
				RepoID:   2,
				Path:     "src/file1.go",
			},
			{
				RepoName: "repo2",
				RepoID:   2,
				Path:     "src/file2.go",
			},
		}

		_, filter, _ := f.getMatcher(context.Background(), repos)
		filtered := filterChunks(chunks, filter)

		require.Equal(t, 2, len(filtered))
		require.Equal(t, api.RepoName("repo1"), filtered[0].RepoName)
		require.Equal(t, "src/file2.go", filtered[0].Path)
		require.Equal(t, api.RepoName("repo2"), filtered[1].RepoName)
		require.Equal(t, "src/file1.go", filtered[1].Path)
	})

	t.Run("empty repos don't error", func(t *testing.T) {
		client := gitserver.NewMockClient()
		client.GetDefaultBranchFunc.SetDefaultReturn("", api.CommitID(""), nil)
		client.NewFileReaderFunc.SetDefaultHook(func(ctx context.Context, rn api.RepoName, ci api.CommitID, s string) (io.ReadCloser, error) {
			t.Errorf("repos are empty, no files should be read")
			return nil, nil
		})

		f := newDotcomFilter(logger, client)
		filterableRepos, _, _ := f.getMatcher(context.Background(), repos)
		require.Len(t, filterableRepos, 0)
	})

	t.Run("errors checking head do error", func(t *testing.T) {
		client := gitserver.NewMockClient()
		client.GetDefaultBranchFunc.SetDefaultReturn("", api.CommitID(""), errors.New("fail"))
		client.NewFileReaderFunc.SetDefaultHook(func(ctx context.Context, rn api.RepoName, ci api.CommitID, s string) (io.ReadCloser, error) {
			t.Errorf("failed checking head should not continue")
			return nil, nil
		})

		f := newDotcomFilter(logger, client)
		filterableRepos, _, _ := f.getMatcher(context.Background(), repos)
		require.Len(t, filterableRepos, 0)
	})

	t.Run("error reading ignore file causes error", func(t *testing.T) {
		client := gitserver.NewMockClient()
		client.GetDefaultBranchFunc.SetDefaultReturn("main", api.CommitID("abc123"), nil)
		client.NewFileReaderFunc.SetDefaultHook(func(ctx context.Context, rn api.RepoName, ci api.CommitID, s string) (io.ReadCloser, error) {
			return nil, errors.New("fail")
		})

		f := newDotcomFilter(logger, client)
		filterableRepos, _, _ := f.getMatcher(context.Background(), repos)
		require.Len(t, filterableRepos, 0)
	})
}

func TestDotcomFilterDisabled(t *testing.T) {
	repos := []types.RepoIDName{
		{ID: 1, Name: "repo1"},
		{ID: 2, Name: "repo2"},
	}
	conf.Mock(&conf.Unified{
		SiteConfiguration: schema.SiteConfiguration{},
	})
	logger := logtest.Scoped(t)
	t.Cleanup(func() { conf.Mock(nil) })

	t.Run("Does not exclude files when disabled", func(t *testing.T) {
		client := gitserver.NewMockClient()
		client.GetDefaultBranchFunc.SetDefaultHook(func(ctx context.Context, rn api.RepoName, b bool) (string, api.CommitID, error) {
			t.Fatalf("filter should be disabled no git commands should be called")
			return "", api.CommitID(""), errors.New("should not be called")
		})
		client.NewFileReaderFunc.SetDefaultHook(func(ctx context.Context, rn api.RepoName, ci api.CommitID, s string) (io.ReadCloser, error) {
			t.Fatalf("filter should be disabled no git commands should be called")
			return nil, errors.New("should not be called")
		})

		f := newDotcomFilter(logger, client)

		chunks := []FileChunkContext{
			{
				RepoName: "repo1",
				RepoID:   1,
				Path:     "file1.go",
			},
			{
				RepoName: "repo2",
				RepoID:   2,
				Path:     "folder1/file1.go",
			},
			{
				RepoName: "repo2",
				RepoID:   2,
				Path:     "folder1/folder2/file1.go",
			},
			{
				RepoName: "repo2",
				RepoID:   2,
				Path:     "secret.txt",
			},
		}

		_, matcher, _ := f.getMatcher(context.Background(), repos)
		filtered := filterChunks(chunks, matcher)
		require.Equal(t, 4, len(filtered))
	})
}

func filterChunks(chunks []FileChunkContext, matcher fileMatcher) []FileChunkContext {
	filtered := make([]FileChunkContext, 0, len(chunks))
	for _, chunk := range chunks {
		if matcher(chunk.RepoID, chunk.Path) {
			filtered = append(filtered, chunk)
		}
	}
	return filtered
}
