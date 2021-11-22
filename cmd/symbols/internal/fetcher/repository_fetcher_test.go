package fetcher

import (
	"archive/tar"
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func TestRepositoryFetcher(t *testing.T) {
	validParseRequests := map[string]string{
		"a.txt": strings.Repeat("payload a", 1<<8),
		"b.txt": strings.Repeat("payload b", 1<<9),
		"c.txt": strings.Repeat("payload c", 1<<10),
		"d.txt": strings.Repeat("payload d", 1<<11),
		"e.txt": strings.Repeat("payload e", 1<<12),
		"f.txt": strings.Repeat("payload f", 1<<13),
		"g.txt": strings.Repeat("payload g", 1<<14),
	}

	tarContents := map[string]string{}
	for name, content := range validParseRequests {
		tarContents[name] = content
	}

	// JSON is ignored
	tarContents["ignored.json"] = "{}"

	// Large files are ignored
	tarContents["payloads.txt"] = strings.Repeat("oversized load", maxFileSize)

	gitserverClient := NewMockGitserverClient()
	gitserverClient.FetchTarFunc.SetDefaultHook(createFetchTarFunc(tarContents))

	repositoryFetcher := NewRepositoryFetcher(gitserverClient, 15, &observation.TestContext)
	args := types.SearchArgs{Repo: api.RepoName("foo"), CommitID: api.CommitID("deadbeef")}

	t.Run("all paths", func(t *testing.T) {
		paths := []string(nil)
		ch := repositoryFetcher.FetchRepositoryArchive(context.Background(), args, paths)
		parseRequests := consumeParseRequests(t, ch)

		expectedParseRequests := validParseRequests
		if diff := cmp.Diff(expectedParseRequests, parseRequests); diff != "" {
			t.Errorf("unexpected parse requests (-want +got):\n%s", diff)
		}
	})

	t.Run("selected paths", func(t *testing.T) {
		paths := []string{"a.txt", "b.txt", "c.txt"}
		ch := repositoryFetcher.FetchRepositoryArchive(context.Background(), args, paths)
		parseRequests := consumeParseRequests(t, ch)

		expectedParseRequests := map[string]string{
			"a.txt": validParseRequests["a.txt"],
			"b.txt": validParseRequests["b.txt"],
			"c.txt": validParseRequests["c.txt"],
		}
		if diff := cmp.Diff(expectedParseRequests, parseRequests); diff != "" {
			t.Errorf("unexpected parse requests (-want +got):\n%s", diff)
		}
	})
}

func consumeParseRequests(t *testing.T, ch <-chan parseRequestOrError) map[string]string {
	parseRequests := map[string]string{}
	for v := range ch {
		if v.Err != nil {
			t.Fatalf("unexpected fetch error: %s", v.Err)
		}

		parseRequests[v.ParseRequest.Path] = string(v.ParseRequest.Data)
	}

	return parseRequests
}

func createFetchTarFunc(tarContents map[string]string) func(context.Context, api.RepoName, api.CommitID, []string) (io.ReadCloser, error) {
	return func(ctx context.Context, repo api.RepoName, commit api.CommitID, paths []string) (io.ReadCloser, error) {
		var buffer bytes.Buffer
		tarWriter := tar.NewWriter(&buffer)

		for name, content := range tarContents {
			if paths != nil {
				found := false
				for _, path := range paths {
					if path == name {
						found = true
					}
				}
				if !found {
					continue
				}
			}

			tarHeader := &tar.Header{
				Name: name,
				Mode: 0o600,
				Size: int64(len(content)),
			}
			if err := tarWriter.WriteHeader(tarHeader); err != nil {
				return nil, err
			}
			if _, err := tarWriter.Write([]byte(content)); err != nil {
				return nil, err
			}
		}

		if err := tarWriter.Close(); err != nil {
			return nil, err
		}

		return io.NopCloser(bytes.NewReader(buffer.Bytes())), nil
	}
}
