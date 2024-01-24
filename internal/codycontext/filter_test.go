package context

import (
	"context"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/stretchr/testify/require"
)

func TestNewFilter(t *testing.T) {
	repos := []types.RepoIDName{
		{ID: 1, Name: "repo1"},
		{ID: 2, Name: "repo2"},
	}

	t.Run("no ignore files", func(t *testing.T) {
		client := gitserver.NewMockClient()
		client.ReadFileFunc.SetDefaultReturn(nil, errors.Errorf("err open .cody/ignore: file does not exist"))
		f, err := NewCodyIgnoreFilter(context.Background(), client, repos)
		require.NoError(t, err)

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

		filtered := f.Filter(chunks)
		require.Equal(t, 2, len(filtered))
	})

	t.Run("filters multiple rules in ignore file", func(t *testing.T) {
		client := gitserver.NewMockClient()
		client.HeadFunc.SetDefaultReturn("abc123", true, nil)
		client.ReadFileFunc.SetDefaultHook(func(ctx context.Context, repo api.RepoName, commit api.CommitID, name string) ([]byte, error) {
			if repo == "repo2" { // filter only from repo2
				return []byte("**/file1.go\nsecret.txt"), nil
			}
			return nil, errors.Errorf("err open .cody/ignore: file does not exist")
		})

		f, err := NewCodyIgnoreFilter(context.Background(), client, repos)
		require.NoError(t, err)

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

		filtered := f.Filter(chunks)
		require.Equal(t, 1, len(filtered))
		require.Equal(t, api.RepoName("repo1"), filtered[0].RepoName)
	})

	t.Run("uses correct ignore file by repo", func(t *testing.T) {
		client := gitserver.NewMockClient()
		client.HeadFunc.SetDefaultReturn("abc123", true, nil)
		client.ReadFileFunc.SetDefaultHook(func(ctx context.Context, repo api.RepoName, commit api.CommitID, name string) ([]byte, error) {
			if repo == "repo1" { // filter file1 from repo1
				return []byte("**/file1.go"), nil
			}
			if repo == "repo2" { // filter file2 from repo2
				return []byte("**/file2.go"), nil
			}
			return nil, errors.Errorf("err open .cody/ignore: file does not exist")
		})

		f, err := NewCodyIgnoreFilter(context.Background(), client, repos)
		require.NoError(t, err)

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

		filtered := f.Filter(chunks)
		require.Equal(t, 2, len(filtered))
		require.Equal(t, api.RepoName("repo1"), filtered[0].RepoName)
		require.Equal(t, "src/file2.go", filtered[0].Path)
		require.Equal(t, api.RepoName("repo2"), filtered[1].RepoName)
		require.Equal(t, "src/file1.go", filtered[1].Path)
	})

	t.Run("empty repos don't error", func(t *testing.T) {
		client := gitserver.NewMockClient()
		client.HeadFunc.SetDefaultReturn("", false, nil)
		client.ReadFileFunc.SetDefaultHook(func(ctx context.Context, repo api.RepoName, commit api.CommitID, name string) ([]byte, error) {
			t.Errorf("repos are empty, no files should be read")
			return nil, nil
		})

		_, err := NewCodyIgnoreFilter(context.Background(), client, repos)
		require.NoError(t, err)
	})

	t.Run("errors checking head do error", func(t *testing.T) {
		client := gitserver.NewMockClient()
		client.HeadFunc.SetDefaultReturn("", false, errors.New("fail"))
		client.ReadFileFunc.SetDefaultHook(func(ctx context.Context, repo api.RepoName, commit api.CommitID, name string) ([]byte, error) {
			t.Errorf("repos are empty, no files should be read")
			return nil, nil
		})

		_, err := NewCodyIgnoreFilter(context.Background(), client, repos)
		require.Error(t, err)
	})

	t.Run("error reading ignore file does error", func(t *testing.T) {
		client := gitserver.NewMockClient()
		client.HeadFunc.SetDefaultReturn("abc123", true, nil)
		client.ReadFileFunc.SetDefaultHook(func(ctx context.Context, repo api.RepoName, commit api.CommitID, name string) ([]byte, error) {
			return nil, errors.New("fail")
		})

		_, err := NewCodyIgnoreFilter(context.Background(), client, repos)
		require.Error(t, err)
	})
}
