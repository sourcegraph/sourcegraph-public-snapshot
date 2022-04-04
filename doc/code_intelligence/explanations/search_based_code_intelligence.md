# Search-based code intelligence

Sourcegraph comes with built-in code intelligence provided by search-based heuristics.

If you are interested in enabling precise code intelligence for your repository, see our [how-to guides](../how-to/index.md).

## How does it work?

[Search-based code intelligence](https://github.com/sourcegraph/sourcegraph-basic-code-intel) is able to provide 3 core code intelligence features:

- Jump to definition: it performs a [symbol search](../../code_search/explanations/features.md#symbol-search)
- Hover documentation: it first finds the definition then extracts documentation from comments near the definition
- Find references: it performs a case-sensitive word-boundary cross-repository [plain text search](../../code_search/explanations/features.md#powerful-flexible-queries) for the given symbol

Search-based code intelligence also filters results by file extension and by imports at the top of the file for some languages.

## What languages are supported?

Search-based code intelligence supports all of [the most popular programming languages](https://sourcegraph.com/extensions?category=Programming+languages).

Are you using a language we don't support? [File a GitHub issue](https://github.com/sourcegraph/sourcegraph/issues/new/choose) or [submit a PR](https://github.com/sourcegraph/sourcegraph-basic-code-intel#adding-a-new-sourcegraphsourcegraph-lang-extension).

## Why are my results sometimes incorrect?

Search-based code intelligence uses search-based heuristics, rather than parsing the code into an [abstract syntax tree](https://en.wikipedia.org/wiki/Abstract_syntax_tree) (AST). Incorrect results occur more often for tokens with common names (such as `Get`) than for tokens with more unique names simply because those tokens appear more often in the search index.

If you require 100% confidence in accuracy for a definition or reference results for a symbol you hovered over we recommend utilizing precise code intelligence. Scenarios where you may still get search-based code intelligence results even with precision on are described in more detail in the [precise code intelligence docs](./precise_code_intelligence.md).

## Why does it sometimes time out?

The [symbol search performance](./features.md#symbol-search-behavior-and-performance) section describes query paths and performance. Consider using [Rockskip](rockskip.md) if you're experiencing frequent timeouts.

## What configuration settings can I apply?

The symbols container recognizes these environment variables:

- `CTAGS_COMMAND`: defaults to `universal-ctags`, ctags command (should point to universal-ctags executable compiled with JSON and seccomp support)
- `CTAGS_PATTERN_LENGTH_LIMIT`: defaults to `250`, the maximum length of the patterns output by ctags
- `LOG_CTAGS_ERRORS`: defaults to `false`, log ctags errors
- `SANITY_CHECK`: defaults to `false`, check that go-sqlite3 works then exit 0 if it's ok or 1 if not
- `CACHE_DIR`: defaults to `/tmp/symbols-cache`, directory in which to store cached symbols
- `SYMBOLS_CACHE_SIZE_MB`: defaults to `100000`, maximum size of the disk cache (in megabytes)
- `CTAGS_PROCESSES`: defaults to `strconv.Itoa(runtime.GOMAXPROCS(0))`, number of concurrent parser processes to run
- `REQUEST_BUFFER_SIZE`: defaults to `8192`, maximum size of buffered parser request channel
- `PROCESSING_TIMEOUT`: defaults to `2h`, maximum time to spend processing a repository
- `MAX_TOTAL_PATHS_LENGTH`: defaults to `100000`, maximum sum of lengths of all paths in a single call to git archive
- `USE_ROCKSKIP`: defaults to `false`, enables [Rockskip](rockskip.md) for fast symbol searches and search-based code intelligence on big repositories specified in `ROCKSKIP_REPOS`
- `ROCKSKIP_REPOS`: no default, in combination with `USE_ROCKSKIP=true` this specifies a comma separated list of repositories to index using [Rockskip](rockskip.md) (e.g. `github.com/torvalds/linux,github.com/pallets/flask`)
- `MAX_CONCURRENTLY_INDEXING`: defaults to `4`, maximum number of repositories being indexed at a time by [Rockskip](rockskip.md) (also limits ctags processes)

The defaults come from [`config.go`](https://github.com/sourcegraph/sourcegraph/blob/eea895ae1a8acef08370a5cc6f24bdc7c66cb4ed/cmd/symbols/config.go#L42-L59).
