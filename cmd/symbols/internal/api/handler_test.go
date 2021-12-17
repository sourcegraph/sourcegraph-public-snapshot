package api

import (
	"context"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/sourcegraph/go-ctags"

	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/database"
	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/database/writer"
	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/fetcher"
	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/parser"
	"github.com/sourcegraph/sourcegraph/internal/diskcache"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	symbolsclient "github.com/sourcegraph/sourcegraph/internal/symbols"
)

func init() {
	database.Init()
}

func TestHandler(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { os.RemoveAll(tmpDir) }()

	cache := diskcache.NewStore(tmpDir, "symbols", diskcache.WithBackgroundTimeout(20*time.Minute))

	parserFactory := func() (ctags.Parser, error) {
		return newMockParser("x", "y"), nil
	}
	parserPool, err := parser.NewParserPool(parserFactory, 15)
	if err != nil {
		t.Fatal(err)
	}

	files := map[string]string{
		"a.js": "var x = 1",
	}
	gitserverClient := NewMockGitserverClient()
	gitserverClient.FetchTarFunc.SetDefaultHook(gitserver.CreateTestFetchTarFunc(files))

	parser := parser.NewParser(parserPool, fetcher.NewRepositoryFetcher(gitserverClient, 15, &observation.TestContext), 0, 10, &observation.TestContext)
	databaseWriter := writer.NewDatabaseWriter(tmpDir, gitserverClient, parser)
	cachedDatabaseWriter := writer.NewCachedDatabaseWriter(databaseWriter, cache)
	handler := NewHandler(cachedDatabaseWriter, &observation.TestContext)

	server := httptest.NewServer(handler)
	defer server.Close()

	client := symbolsclient.Client{
		URL:        server.URL,
		HTTPClient: httpcli.InternalDoer,
	}

	x := result.Symbol{Name: "x", Path: "a.js"}
	y := result.Symbol{Name: "y", Path: "a.js"}

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
			result, err := client.Search(context.Background(), testCase.args)
			if err != nil {
				t.Fatalf("unexpected error performing search: %s", err)
			}

			if result == nil {
				if testCase.expected != nil {
					t.Errorf("unexpected search result. want=%+v, have=nil", testCase.expected)
				}
			} else if !reflect.DeepEqual(result, testCase.expected) {
				t.Errorf("unexpected search result. want=%+v, have=%+v", testCase.expected, result)
			}
		})
	}
}

type mockParser struct {
	names []string
}

func newMockParser(names ...string) ctags.Parser {
	return &mockParser{names: names}
}

func (m *mockParser) Parse(name string, content []byte) ([]*ctags.Entry, error) {
	entries := make([]*ctags.Entry, 0, len(m.names))
	for _, name := range m.names {
		entries = append(entries, &ctags.Entry{Name: name, Path: "a.js"})
	}

	return entries, nil
}

func (m *mockParser) Close() {}
