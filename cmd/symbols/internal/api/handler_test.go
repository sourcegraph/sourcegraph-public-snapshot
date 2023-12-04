package api

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/sourcegraph/go-ctags"
	"github.com/sourcegraph/log/logtest"
	"golang.org/x/sync/semaphore"

	"github.com/sourcegraph/sourcegraph/cmd/symbols/fetcher"
	"github.com/sourcegraph/sourcegraph/cmd/symbols/gitserver"
	symbolsdatabase "github.com/sourcegraph/sourcegraph/cmd/symbols/internal/database"
	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/database/writer"
	"github.com/sourcegraph/sourcegraph/cmd/symbols/parser"
	"github.com/sourcegraph/sourcegraph/internal/ctags_config"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/sourcegraph/sourcegraph/internal/diskcache"
	"github.com/sourcegraph/sourcegraph/internal/endpoint"
	internalgrpc "github.com/sourcegraph/sourcegraph/internal/grpc/defaults"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	symbolsclient "github.com/sourcegraph/sourcegraph/internal/symbols"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func init() {
	symbolsdatabase.Init()
}

func TestHandler(t *testing.T) {
	tmpDir := t.TempDir()

	cache := diskcache.NewStore(tmpDir, "symbols", diskcache.WithBackgroundTimeout(20*time.Minute))

	parserFactory := func(source ctags_config.ParserType) (ctags.Parser, error) {
		pathToEntries := map[string][]*ctags.Entry{
			"a.js": {
				{
					Name: "x",
					Path: "a.js",
					Line: 1, // ctags line numbers are 1-based
				},
				{
					Name: "y",
					Path: "a.js",
					Line: 2,
				},
			},
		}
		return newMockParser(pathToEntries), nil
	}
	parserPool, err := parser.NewParserPool(parserFactory, 15, parser.DefaultParserTypes)
	if err != nil {
		t.Fatal(err)
	}

	files := map[string]string{
		"a.js": "var x = 1\nvar y = 2",
	}
	gitserverClient := NewMockGitserverClient()
	gitserverClient.FetchTarFunc.SetDefaultHook(gitserver.CreateTestFetchTarFunc(files))

	symbolParser := parser.NewParser(&observation.TestContext, parserPool, fetcher.NewRepositoryFetcher(&observation.TestContext, gitserverClient, 1000, 1_000_000), 0, 10)
	databaseWriter := writer.NewDatabaseWriter(observation.TestContextTB(t), tmpDir, gitserverClient, symbolParser, semaphore.NewWeighted(1))
	cachedDatabaseWriter := writer.NewCachedDatabaseWriter(databaseWriter, cache)
	handler := NewHandler(MakeSqliteSearchFunc(observation.TestContextTB(t), cachedDatabaseWriter, dbmocks.NewMockDB()), gitserverClient.ReadFile, nil, "")

	server := httptest.NewServer(handler)
	defer server.Close()

	connectionCache := internalgrpc.NewConnectionCache(logtest.Scoped(t))
	t.Cleanup(connectionCache.Shutdown)

	client := symbolsclient.Client{
		Endpoints:           endpoint.Static(server.URL),
		GRPCConnectionCache: connectionCache,
		HTTPClient:          httpcli.InternalDoer,
	}

	x := result.Symbol{Name: "x", Path: "a.js", Line: 0, Character: 4}
	y := result.Symbol{Name: "y", Path: "a.js", Line: 1, Character: 4}

	testCases := map[string]struct {
		args     search.SymbolsParameters
		expected result.Symbols
	}{
		"simple": {
			args:     search.SymbolsParameters{First: 10},
			expected: []result.Symbol{x, y},
		},
		"onematch": {
			args:     search.SymbolsParameters{Query: "x", First: 10},
			expected: []result.Symbol{x},
		},
		"nomatches": {
			args:     search.SymbolsParameters{Query: "foo", First: 10},
			expected: nil,
		},
		"caseinsensitiveexactmatch": {
			args:     search.SymbolsParameters{Query: "^X$", First: 10},
			expected: []result.Symbol{x},
		},
		"casesensitiveexactmatch": {
			args:     search.SymbolsParameters{Query: "^x$", IsCaseSensitive: true, First: 10},
			expected: []result.Symbol{x},
		},
		"casesensitivenoexactmatch": {
			args:     search.SymbolsParameters{Query: "^X$", IsCaseSensitive: true, First: 10},
			expected: nil,
		},
		"caseinsensitiveexactpathmatch": {
			args:     search.SymbolsParameters{IncludePatterns: []string{"^A.js$"}, First: 10},
			expected: []result.Symbol{x, y},
		},
		"casesensitiveexactpathmatch": {
			args:     search.SymbolsParameters{IncludePatterns: []string{"^a.js$"}, IsCaseSensitive: true, First: 10},
			expected: []result.Symbol{x, y},
		},
		"casesensitivenoexactpathmatch": {
			args:     search.SymbolsParameters{IncludePatterns: []string{"^A.js$"}, IsCaseSensitive: true, First: 10},
			expected: nil,
		},
		"exclude": {
			args:     search.SymbolsParameters{ExcludePattern: "a.js", IsCaseSensitive: true, First: 10},
			expected: nil,
		},
	}

	for label, testCase := range testCases {
		t.Run(label, func(t *testing.T) {
			resultSymbols, err := client.Search(context.Background(), testCase.args)
			if err != nil {
				t.Fatalf("unexpected error performing search: %s", err)
			}

			if resultSymbols == nil {
				if testCase.expected != nil {
					t.Errorf("unexpected search result. want=%+v, have=nil", testCase.expected)
				}
			} else if diff := cmp.Diff(resultSymbols, testCase.expected, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("unexpected search result. diff: %s", diff)
			}
		})
	}
}

type mockParser struct {
	pathToEntries map[string][]*ctags.Entry
}

func newMockParser(pathToEntries map[string][]*ctags.Entry) ctags.Parser {
	return &mockParser{pathToEntries: pathToEntries}
}

func (m *mockParser) Parse(path string, content []byte) ([]*ctags.Entry, error) {
	if entries, ok := m.pathToEntries[path]; ok {
		return entries, nil
	}
	return nil, errors.Newf("no mock entries for %s", path)
}

func (m *mockParser) Close() {}
