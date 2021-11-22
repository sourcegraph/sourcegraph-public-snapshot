package search

import (
	"context"
	"fmt"
	"os"
	"path"
	"testing"
	"time"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/parser"
	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/sqlite"
	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/diskcache"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/testutil"
)

func BenchmarkSearch(b *testing.B) {
	log15.Root().SetHandler(log15.LvlFilterHandler(log15.LvlError, log15.Root().GetHandler()))

	gitserverClient := NewMockGitserverClient()
	gitserverClient.FetchTarFunc.SetDefaultHook(testutil.FetchTarFromGithubWithPaths)

	parserPool, err := parser.NewParserPool(parser.NewCtagsParser, 15)
	if err != nil {
		b.Fatal(err)
	}
	parser := parser.NewParser(parserPool, parser.NewRepositoryFetcher(gitserverClient, 15, &observation.TestContext))

	cache := &diskcache.Store{
		Dir:               "/tmp/symbols-cache",
		Component:         "symbols",
		BackgroundTimeout: 20 * time.Minute,
	}

	databaseWriter := sqlite.NewDatabaseWriter("/tmp/symbols-cache", gitserverClient, parser)
	searcher := NewSearcher(cache, databaseWriter)

	ctx := context.Background()
	b.ResetTimer()

	indexTests := []types.SearchArgs{
		{Repo: "github.com/sourcegraph/go-langserver", CommitID: "391a062a7d9977510e7e883e412769b07fed8b5e"},
		{Repo: "github.com/moby/moby", CommitID: "6e5c2d639f67ae70f54d9f2285f3261440b074aa"},
	}

	queryTests := []types.SearchArgs{
		{Repo: "github.com/sourcegraph/go-langserver", CommitID: "391a062a7d9977510e7e883e412769b07fed8b5e", Query: "^sortedImportRecord$", First: 10},
		{Repo: "github.com/sourcegraph/go-langserver", CommitID: "391a062a7d9977510e7e883e412769b07fed8b5e", Query: "1234doesnotexist1234", First: 1},
		{Repo: "github.com/moby/moby", CommitID: "6e5c2d639f67ae70f54d9f2285f3261440b074aa", Query: "^fsCache$", First: 10},
		{Repo: "github.com/moby/moby", CommitID: "6e5c2d639f67ae70f54d9f2285f3261440b074aa", Query: "1234doesnotexist1234", First: 1},
	}

	runIndexTest := func(test types.SearchArgs) {
		b.Run(fmt.Sprintf("indexing %s@%s", path.Base(string(test.Repo)), test.CommitID[:3]), func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				tempFile, err := os.CreateTemp("", "")
				if err != nil {
					b.Fatal(err)
				}
				defer os.Remove(tempFile.Name())

				// err = sqlite.WriteAllSymbolsToNewDB(ctx, parser, tempFile.Name(), test.Repo, test.CommitID)
				err = sqlite.NewDatabaseWriter("", nil, parser).WriteDBFile(ctx, types.SearchArgs{Repo: test.Repo, CommitID: test.CommitID}, tempFile.Name())
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}

	runQueryTest := func(test types.SearchArgs) {
		b.Run(fmt.Sprintf("searching %s@%s %s", path.Base(string(test.Repo)), test.CommitID[:3], test.Query), func(b *testing.B) {
			_, err := searcher.Search(ctx, test)
			if err != nil {
				b.Fatal(err)
			}
			b.ResetTimer()
			for n := 0; n < b.N; n++ {
				_, err := searcher.Search(ctx, test)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}

	for _, test := range indexTests {
		runIndexTest(test)
	}

	for _, test := range queryTests {
		runQueryTest(test)
	}
}
