package api

import (
	"context"
	"net/http/httptest"
	"sort"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/sourcegraph/go-ctags"
	"github.com/sourcegraph/log/logtest"
	"golang.org/x/sync/semaphore"

	symbolsdatabase "github.com/sourcegraph/sourcegraph/cmd/symbols/internal/database"
	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/database/writer"
	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/fetcher"
	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/parser"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/ctags_config"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/sourcegraph/sourcegraph/internal/diskcache"
	"github.com/sourcegraph/sourcegraph/internal/endpoint"
	internalgrpc "github.com/sourcegraph/sourcegraph/internal/grpc/defaults"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	symbolsclient "github.com/sourcegraph/sourcegraph/internal/symbols"
	types "github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func init() {
	symbolsdatabase.Init()
}

func TestHandler(t *testing.T) {
	tmpDir := t.TempDir()

	cache := diskcache.NewStore(tmpDir, "symbols", diskcache.WithBackgroundTimeout(20*time.Minute))
	// This ensures the ctags config is initialized properly
	conf.MockAndNotifyWatchers(&conf.Unified{})
	parserFactory := func(source ctags_config.ParserType) (ctags.Parser, error) {
		var pathToEntries map[string][]*ctags.Entry
		if source == ctags_config.UniversalCtags {
			pathToEntries = map[string][]*ctags.Entry{
				"a.pl": {
					{
						Name:     "x",
						Path:     "a.pl",
						Language: "Perl",
						Line:     1, // ctags line numbers are 1-based
					},
					{
						Name:     "y",
						Path:     "a.pl",
						Language: "Perl",
						Line:     2,
					},
				},
				".zshrc": {
					{
						Name:     "z",
						Path:     ".zshrc",
						Language: "Zsh",
						Line:     1,
					},
				},
			}
		} else if source == ctags_config.ScipCtags {
			pathToEntries = map[string][]*ctags.Entry{
				"b.magik": {
					{
						Name:     "v",
						Path:     "b.magik",
						Language: "Magik",
						Line:     1, // ctags line numbers are 1-based
					},
					{
						Name:     "w",
						Path:     "b.magik",
						Language: "Magik",
						Line:     2,
					},
				},
			}
		} else {
			t.Errorf("Invalid ctags type %d", source)
		}

		return newMockParser(pathToEntries), nil
	}

	parserPool, err := parser.NewParserPool(observation.TestContextTB(t), "test", parserFactory, 15, parser.DefaultParserTypes)
	if err != nil {
		t.Fatal(err)
	}

	files := map[string]string{
		"a.pl":    "$x = 1\n$y = 2",
		".zshrc":  "z=42",
		"b.magik": "v << 1\nw<<2",
	}
	gitserverClient := NewMockGitserverClient()
	gitserverClient.FetchTarFunc.SetDefaultHook(gitserver.CreateTestFetchTarFunc(files))

	symbolParser := parser.NewParser(observation.TestContextTB(t), parserPool, fetcher.NewRepositoryFetcher(observation.TestContextTB(t), gitserverClient, 1000, 1_000_000), 0, 10)
	databaseWriter := writer.NewDatabaseWriter(observation.TestContextTB(t), tmpDir, gitserverClient, symbolParser, semaphore.NewWeighted(1))
	cachedDatabaseWriter := writer.NewCachedDatabaseWriter(databaseWriter, cache)
	handler := NewHandler(MakeSqliteSearchFunc(observation.TestContextTB(t), cachedDatabaseWriter, dbmocks.NewMockDB()), func(ctx context.Context, rcp types.RepoCommitPath) ([]byte, error) { return nil, nil }, nil)

	server := httptest.NewServer(handler)
	defer server.Close()

	connectionCache := internalgrpc.NewConnectionCache(logtest.Scoped(t))
	t.Cleanup(connectionCache.Shutdown)

	client := symbolsclient.Client{
		Endpoints:           endpoint.Static(server.URL),
		GRPCConnectionCache: connectionCache,
	}

	x := result.Symbol{Name: "x", Path: "a.pl", Language: "Perl", Line: 0, Character: 1}
	y := result.Symbol{Name: "y", Path: "a.pl", Language: "Perl", Line: 1, Character: 1}
	z := result.Symbol{Name: "z", Path: ".zshrc", Language: "Zsh", Line: 0, Character: 0}
	v := result.Symbol{Name: "v", Path: "b.magik", Language: "Magik", Line: 0, Character: 0}
	w := result.Symbol{Name: "w", Path: "b.magik", Language: "Magik", Line: 1, Character: 0}

	testCases := map[string]struct {
		args     search.SymbolsParameters
		expected result.Symbols
	}{
		"simple": {
			args:     search.SymbolsParameters{IncludePatterns: []string{"^a.pl$"}, First: 10},
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
			args:     search.SymbolsParameters{IncludePatterns: []string{"^A.pl$"}, First: 10},
			expected: []result.Symbol{x, y},
		},
		"casesensitiveexactpathmatch": {
			args:     search.SymbolsParameters{IncludePatterns: []string{"^a.pl$"}, IsCaseSensitive: true, First: 10},
			expected: []result.Symbol{x, y},
		},
		"casesensitivenoexactpathmatch": {
			args:     search.SymbolsParameters{IncludePatterns: []string{"^A.pl$"}, IsCaseSensitive: true, First: 10},
			expected: nil,
		},
		"exclude": {
			args:     search.SymbolsParameters{ExcludePattern: "a.pl", IsCaseSensitive: true, First: 10},
			expected: []result.Symbol{z, v, w},
		},
		"include lang filters": {
			args:     search.SymbolsParameters{Query: ".*", IncludeLangs: []string{"Perl"}, IsCaseSensitive: true, First: 10},
			expected: []result.Symbol{x, y},
		},
		"include lang filters with ctags conversion": {
			args:     search.SymbolsParameters{Query: ".*", IncludeLangs: []string{"Shell"}, IsCaseSensitive: true, First: 10},
			expected: []result.Symbol{z},
		},
		"exclude lang filters": {
			args:     search.SymbolsParameters{Query: ".*", ExcludeLangs: []string{"Perl", "Magik"}, IsCaseSensitive: true, First: 10},
			expected: []result.Symbol{z},
		},
		"scip-ctags only language": {
			args:     search.SymbolsParameters{Query: ".*", IncludeLangs: []string{"Magik"}, IsCaseSensitive: true, First: 10},
			expected: []result.Symbol{v, w},
		},
	}

	for label, testCase := range testCases {
		t.Run(label, func(t *testing.T) {
			resultSymbols, limitHit, err := client.Search(context.Background(), testCase.args)

			// Sort to ensure consistent comparisons
			sort.Slice(resultSymbols, func(i, j int) bool {
				return resultSymbols[i].Path < resultSymbols[j].Path
			})

			if err != nil {
				t.Fatalf("unexpected error performing search: %s", err)
			}
			if limitHit {
				t.Fatalf("unexpected limitHit")
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
