package fetcher

import (
	"context"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/search"
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

	gitserverClient := NewMockGitserverClient()
	gitserverClient.FetchTarFunc.SetDefaultHook(gitserver.CreateTestFetchTarFunc(tarContents))

	repositoryFetcher := NewRepositoryFetcher(observation.TestContextTB(t), gitserverClient, 1000, 1_000_000)
	args := search.SymbolsParameters{Repo: api.RepoName("foo"), CommitID: api.CommitID("deadbeef")}

	t.Run("all paths", func(t *testing.T) {
		paths := []string(nil)
		ch := repositoryFetcher.FetchRepositoryArchive(context.Background(), args.Repo, args.CommitID, paths)
		parseRequests := consumeParseRequests(t, ch)

		expectedParseRequests := validParseRequests
		if diff := cmp.Diff(expectedParseRequests, parseRequests); diff != "" {
			t.Errorf("unexpected parse requests (-want +got):\n%s", diff)
		}
	})

	t.Run("selected paths", func(t *testing.T) {
		paths := []string{"a.txt", "b.txt", "c.txt"}
		ch := repositoryFetcher.FetchRepositoryArchive(context.Background(), args.Repo, args.CommitID, paths)
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

func consumeParseRequests(t *testing.T, ch <-chan ParseRequestOrError) map[string]string {
	parseRequests := map[string]string{}
	for v := range ch {
		if v.Err != nil {
			t.Fatalf("unexpected fetch error: %s", v.Err)
		}

		parseRequests[v.ParseRequest.Path] = string(v.ParseRequest.Data)
	}

	return parseRequests
}

func TestBatching(t *testing.T) {
	// When all strings fit in a single batch, they should be sent in a single batch.
	if diff := cmp.Diff([][]string{{"foo", "bar", "baz"}}, batchByTotalLength([]string{"foo", "bar", "baz"}, 10)); diff != "" {
		t.Errorf("unexpected batches (-want +got):\n%s", diff)
	}

	// When not all strings fit into a single batch, they should be sent in multiple batches.
	if diff := cmp.Diff([][]string{{"foo", "bar"}, {"baz"}}, batchByTotalLength([]string{"foo", "bar", "baz"}, 7)); diff != "" {
		t.Errorf("unexpected batches (-want +got):\n%s", diff)
	}

	// When the max is smaller than each string, they should be put into their own batches.
	if diff := cmp.Diff([][]string{{"foo"}, {"bar"}, {"baz"}}, batchByTotalLength([]string{"foo", "bar", "baz"}, 2)); diff != "" {
		t.Errorf("unexpected batches (-want +got):\n%s", diff)
	}
}
