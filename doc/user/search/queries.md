# Search query syntax

<!-- Search syntax styling overrides -->
<style>
.markdown-body tr td:nth-child(1) {
  min-width:175px;
}

.markdown-body tr td:nth-child(3) {
  min-width: 250px;
}
.markdown-body tr td:nth-child(3) code {
  word-break: break-all;
}

.markdown-body img {
  border: 1px solid;
  border-radius: 2px;
  width: 18px;
}
</style>

This page describes search pattern syntax and keywords available for code search. A typical search pattern describes content or filenames to find across all repositories. At the most basic level, a search pattern can simply be a word like `hello`. See our [search patterns](#search-patterns) documentation for detailed usage. Queries can also include keywords. For example, a typical search query will include a `repo:` keyword that filters search results for a specific repository. See our [keywords](#keywords-all-searches) documentation for more examples.

## Search pattern syntax

This section documents the available search pattern syntax and interpretation in Sourcegraph. A search pattern is _required_ to match file content. A search pattern is _optional_ and may be omitted when searching for [commits](#keywords-diff-and-commit-searches-only), [filenames](#filename-search), or [repository names](#repository-name-search).

### Literal search (default)

Literal search interprets search patterns literally to simplify searching for words or punctuation.

| Search pattern syntax | Description |
| --- | --- | 
| [`foo bar`](https://sourcegraph.com/search?q=foo+bar&patternType=literal) | Match the string `foo bar`. Matching is ordered: match `foo` followed by `bar`. Matching is case-_insensitive_ (toggle the <img src=../img/case.png> button to change). | |
| [`"foo bar"`](https://sourcegraph.com/search?q=%22foo+bar%22&patternType=literal) | Match the string `"foo bar"`. The quotes are matched literally. |

As of version 3.9.0, by default, searches are interpreted literally instead of as regexp. To change the default search, site admins and users can change their instance and personal default by setting `search.defaultPatternType` to `"literal"` or `"regexp"`. 

### Regexp search 

Click the <img src=../img/regex.png> toggle to interpret search patterns as regexps. [RE2 syntax](https://golang.org/s/re2syntax) is supported. In general, special characters may be escaped with `\`. Here is a list of valid syntax and behavior:

| Search pattern syntax | Description |
| --- | --- |
| [`foo bar`](https://sourcegraph.com/search?q=foo+bar&patternType=regexp) | Search for the regexp `foo(.*?)bar`. Spaces between non-whitespace strings is converted to `.*?` to create a fuzzy search. Matching is case _insensitive_ (toggle the <img src=../img/case.png> button to change). |
| [`foo\ bar`](https://sourcegraph.com/search?q=foo%5C+bar&patternType=regexp) or<br/>[`/foo bar/`](https://sourcegraph.com/search?q=/foo+bar/&patternType=regexp) | Search for the regexp `foo bar`. The `\` escapes the space and treats the space as part of the pattern. Using the delimiter syntax `/ ... /` avoids the need for escaping spaces. |
| [`foo\nbar`](https://sourcegraph.com/search?q=foo%5Cnbar&patternType=regexp) | Perform a multiline regexp search. `\n` is interpreted as a newline. |
| [`"foo bar"`](https://sourcegraph.com/search?q=%27foo+bar%27&patternType=regexp) | Match the _string literal_ `foo bar`. Quoting strings when regexp is active means patterns are interpreted [literally](#literal-search-default), except that special characters like `"` and `\` may be escaped, and whitespace escape sequences like `\n` are interpreted normally. |

### Structural search

Click the <img src=../img/brackets.png> toggle to activate [structural search](structural.md). Structural search is a way to match more complex syntactic structures in code, and thus only applies to matching file contents. See the dedicated [usage documentation](structural.md) for more details. Here is a  brief overview of valid syntax:

| Search pattern syntax | Description |
| --- | --- |
| [`New(:[args])`](https://sourcegraph.com/search?q=repo:github.com/sourcegraph/sourcegraph++New%28:%5Bargs%5D%29+lang:go&patternType=structural) | Match the string `New` followed by _balanced parentheses_ containing zero or more characters, including newlines. Matching is _case-sensitive_. Make the search [language-aware](structural.md#current-functionality-and-restrictions) by adding a `lang:` [keyword](#keywords-all-searches). | 
| [`"New(:[args])"`](https://sourcegraph.com/search?q=repo:github.com/sourcegraph/sourcegraph+%22New%28:%5Bargs%5D%29%22+lang:go&patternType=structural) or<br/> [`'New(:[args])'`](https://sourcegraph.com/search?q=repo:github.com/sourcegraph/sourcegraph+%27New%28:%5Bargs%5D%29%27+lang:go&patternType=structural) | Quoting the search pattern has the same meaning as `New(:[args])`, but avoids syntax errors that may conflict with [keyword syntax](#keywords-all-searches). Special characters like `"` and `\` may be escaped. |

Note: It is not possible to perform case-insensitive matching with structural search. 

## Keywords (all searches)

The following keywords can be used on all searches (using [RE2 syntax](https://golang.org/s/re2syntax) any place a regex is accepted):

| Keyword | Description | Examples |
| --- | --- | --- |
| **repo:regexp-pattern** <br> **repo:regexp-pattern@rev** <br> _alias: r_  | Only include results from repositories whose path matches the regexp. A repository's path is a string such as _github.com/myteam/abc_ or _code.example.com/xyz_ that depends on your organization's repository host. If the regexp ends in **@rev**, that revision is searched instead of the default branch (usually `master`).  | [`repo:gorilla/mux testroute`](https://sourcegraph.com/search?q=repo:gorilla/mux+testroute)<br/>`repo:alice/abc@mybranch`  |
| **-repo:regexp-pattern** <br> _alias: -r_ | Exclude results from repositories whose path matches the regexp. | `repo:alice/ -repo:old-repo` |
| **repogroup:group-name** <br> _alias: g_ | Only include results from the named group of repositories (defined by the server admin). Same as using a repo: keyword that matches all of the group's repositories. Use repo: unless you know that the group exists. | |
| **file:regexp-pattern** <br> _alias: f_ | Only include results in files whose full path matches the regexp. | [`file:\.js$ httptest`](https://sourcegraph.com/search?q=file:%5C.js%24+httptest) <br> [`file:internal/ httptest`](https://sourcegraph.com/search?q=file:internal/+httptest) |
| **-file:regexp-pattern** <br> _alias: -f_ | Exclude results from files whose full path matches the regexp. | [`file:\.js$ -file:test http`](https://sourcegraph.com/search?q=file:%5C.js%24+-file:test+http) |
| **content:"pattern"** | Explicitly override the [search pattern](#search-pattern-syntax). Useful for explicitly delineating the pattern to search for if it clashes with other parts of the query. | [`repo:sourcegraph "repo:sourcegraph"`](https://sourcegraph.com/search?q=repo:sourcegraph+content:"repo:sourcegraph"&patternType=literal) |
| **lang:language-name** <br> _alias: l_ | Only include results from files in the specified programming language. | [`lang:typescript encoding`](https://sourcegraph.com/search?q=lang:typescript+encoding) |
| **-lang:language-name** <br> _alias: -l_ | Exclude results from files in the specified programming language. | [`-lang:typescript encoding`](https://sourcegraph.com/search?q=-lang:typescript+encoding) |
| **type:symbol** | Perform a symbol search. | [`type:symbol path`](https://sourcegraph.com/search?q=type:symbol+path)  ||
| **case:yes**  | Perform a case sensitive query. Without this, everything is matched case insensitively. | [`OPEN_FILE case:yes`](https://sourcegraph.com/search?q=OPEN_FILE+case:yes) |
| **fork:yes, fork:only** | Include results from repository forks or filter results to only repository forks. Results in repository forks are exluded by default. | [`fork:yes repo:sourcegraph`](https://sourcegraph.com/search?q=fork:yes+repo:sourcegraph) |
| **archived:yes, archived:only** | Include archived repositories or filter results to only archived repositories. Results in archived repositories are excluded by default. | [`repo:sourcegraph/ archived:only`](https://sourcegraph.com/search?q=repo:%5Egithub.com/sourcegraph/+archived:only) |
| **repohasfile:regexp-pattern** | Only include results from repositories that contain a matching file. This keyword is a pure filter, so it requires at least one other search term in the query.  Note: this filter currently only works on text matches and file path matches. | [`repohasfile:\.py file:Dockerfile pip`](https://sourcegraph.com/search?q=repohasfile:%5C.py+file:Dockerfile+pip+repo:/sourcegraph/) |
| **-repohasfile:regexp-pattern** | Exclude results from repositories that contain a matching file. This keyword is a pure filter, so it requires at least one other search term in the query. Note: this filter currently only works on text matches and file path matches. | [`-repohasfile:Dockerfile docker`](https://sourcegraph.com/search?q=-repohasfile:Dockerfile+docker) |
| **repohascommitafter:"string specifying time frame"** | (Experimental) Filter out stale repositories that don't contain commits past the specified time frame. | [`repohascommitafter:"last thursday"`](https://sourcegraph.com/search?q=error+repohascommitafter:%22last+thursday%22) <br> [`repohascommitafter:"june 25 2017"`](https://sourcegraph.com/search?q=error+repohascommitafter:%22june+25+2017%22) |
| **count:_N_**<br/> | Retrieve at least <em>N</em> results. By default, Sourcegraph stops searching early and returns if it finds a full page of results. This is desirable for most interactive searches. To wait for all results, or to see results beyond the first page, use the **count:** keyword with a larger <em>N</em>. This can also be used to get deterministic results and result ordering (whose order isn't dependent on the variable time it takes to perform the search). | [`count:1000 function`](https://sourcegraph.com/search?q=count:1000+repo:sourcegraph/sourcegraph$+function) |
| **timeout:_go-duration-value_**<br/> | Customizes the timeout for searches. The value of the parameter is a string that can be parsed by the [Go time package's `ParseDuration`](https://golang.org/pkg/time/#ParseDuration) (e.g. 10s, 100ms). By default, the timeout is set to 10 seconds, and the search will optimize for returning results as soon as possible. The timeout value cannot be set longer than 1 minute. When provided, the search is given the full timeout to complete. | [`repo:^github.com/sourcegraph timeout:15s func count:10000`](https://sourcegraph.com/search?q=repo:%5Egithub.com/sourcegraph/+timeout:15s+func+count:10000) |
| **patterntype:literal, patterntype:regexp, patterntype:structural**  | Configure your query to be interpreted literally, as a regular expression, or a [structural search pattern](structural.md). Note: this keyword is available as an accessibility option in addition to the visual toggles. | [`test. patternType:literal`](https://sourcegraph.com/search?q=test.+patternType:literal)<br/>[`(open\|close)file patternType:regexp`](https://sourcegraph.com/search?q=%28open%7Cclose%29file&patternType=regexp) |
| **visibility:any, visibility:public, visibility:private** | Filter results to only public or private repositories. The default is to include both private and public repositories. | [`type:repo visibility:public`](https://sourcegraph.com/search?q=type:repo+visibility:public) |
| **stable:yes** | Ensures a deterministic result order. Applies only to file contents. Limited to at max `count:5000` results. Note this field should be removed if you're using the pagination API, which already ensures deterministic results. | [`func stable:yes count:10`](https://sourcegraph.com/search?q=func+stable:yes+count:30&patternType=literal) |


Multiple or combined **repo:** and **file:** keywords are intersected. For example, `repo:foo repo:bar` limits your search to repositories whose path contains **both** _foo_ and _bar_ (such as _github.com/alice/foobar_). To include results from repositories whose path contains **either** _foo_ or _bar_, use `repo:foo|bar`.

## Operators

Use operators to create more expressive searches.

> NOTE: Operators are available as of 3.15 and enabled with `{"experimentalFeatures": {"andOrQuery": "enabled"}}` in site settings. Built-in operator support is planned for the upcoming 3.16 release and onwards.

| Operator | Example |
| --- | --- |
| `and`, `AND` | [`conf.Get( and log15.Error(`](https://sourcegraph.com/search?q=repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+conf.Get%28+and+log15.Error%28&patternType=regexp), [`conf.Get( and log15.Error( and after`](https://sourcegraph.com/search?q=repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+conf.Get%28+and+log15.Error%28+and+after&patternType=regexp) |

Returns results for files containing matches on the left _and_ right side of the `and` (set intersection). The number of results reports the number of files containing both strings. 

| Operator | Example |
| --- | --- |
| `or`, `OR` | [`conf.Get( or log15.Error(`](https://sourcegraph.com/search?q=repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+conf.Get%28+or+log15.Error%28&patternType=regexp), [<code>conf.Get( or log15.Error( or after   </code>](https://sourcegraph.com/search?q=repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+conf.Get%28+or+log15.Error%28+or+after&patternType=regexp)| 

Returns file content matching either on the left or right side, or both (set union). The number of results reports the number of matches of both strings. 

### Operator precedence and groups

Operators may be combined. `and`-expressions have higher precedence (bind tighter) than `or`-expressions so that `a and b or c and d` means `(a and b) or (c and d)`. 

Expressions may be grouped with parentheses to change the default precedence and meaning. For example: `a and (b or c) and d`.

### Operator scope

Except for simple cases, search patterns bind tightest to scoped fields, like `file:main.c`. So, a combined query like
`file:main.c char c  or (int i and int j)` generally means `(file:main.c char c) or (int i and int j)`

Since we don't yet support search subexpressions with different scopes, the above will raise an alert. If the intent is to apply the `file` scope to the entire pattern, group it like so: `file:main.c (char c or (int i and int j))`

### Operator support

Operators are supported in regexp and structural search modes, but not literal search mode. How operators interpret search pattern syntax depends on kind of search (whether [regexp](#regexp-search) or [structural](#structural-search)). Operators currently only apply to searches for file content. Thus, expressions like `repo:npm/cli or repo:npm/npx` are not currently supported. 

---

## Keywords (diff and commit searches only)

The following keywords are only used for **commit diff** and **commit message** searches, which show changes over time:

| Keyword  | Description | Examples |
| --- | --- | --- |
| **repo:regexp-pattern@refs** | Specifies which Git refs (`:`-separated) to search for commits. Use `*refs/heads/` to include all Git branches (and `*refs/tags/` to include all Git tags). You can also prefix a Git ref name or pattern with `^` to exclude. For example, `*refs/heads/:^refs/heads/master` will match all commits that are not merged into master. | [`repo:vscode@*refs/heads/:^refs/heads/master type:diff task`](https://sourcegraph.com/search?q=repo:%5Egithub%5C.com/Microsoft/vscode%24%40*refs/heads/:%5Erefs/heads/master+type:diff+after:%221+month+ago%22+task#1) (unmerged commit diffs containing `task`) |
| **type:diff** <br> **type:commit**  | Specifies the type of search. By default, searches are executed on all code at a given point in time (a branch or a commit). Specify the `type:` if you want to search over changes to code or commit messages instead (diffs or commits).  | [`type:diff func`](https://sourcegraph.com/search?q=type:diff+func+repo:sourcegraph/sourcegraph$) <br> [`type:commit test`](https://sourcegraph.com/search?q=type:commit+test+repo:sourcegraph/sourcegraph$) |
| **author:name** | Only include results from diffs or commits authored by the user. Regexps are supported. Note that they match the whole author string of the form `Full Name <user@example.com>`, so to include only authors from a specific domain, use `author:example.com>$`.<br><br> You can also search by `committer:git-email`. _Note: there is a committer only when they are a different user than the author._ | [`type:diff author:nick`](https://sourcegraph.com/search?q=repo:sourcegraph/sourcegraph$+type:diff+author:nick) |
| **before:"string specifying time frame"** | Only include results from diffs or commits which have a commit date before the specified time frame | [`before:"last thursday"`](https://sourcegraph.com/search?q=repo:sourcegraph/sourcegraph$+type:diff+author:nick+before:%22last+thursday%22) <br> [`before:"november 1 2019"`](https://sourcegraph.com/search?q=repo:sourcegraph/sourcegraph$+type:diff+author:nick+before:%22november+1+2019%22) |
| **after:"string specifying time frame"**  | Only include results from diffs or commits which have a commit date after the specified time frame| [`after:"6 weeks ago"`](https://sourcegraph.com/search?q=repo:sourcegraph/sourcegraph$+type:diff+author:nick+after:%226+weeks+ago%22) <br> [`after:"november 1 2019"`](https://sourcegraph.com/search?q=repo:sourcegraph/sourcegraph$+type:diff+author:nick+after:%22november+1+2019%22) |
| **message:"any string"** | Only include results from diffs or commits which have commit messages containing the string | [`type:commit message:"testing"`](https://sourcegraph.com/search?q=type:commit+repo:sourcegraph/sourcegraph$+message:%22testing%22) <br> [`type:diff message:"testing"`](https://sourcegraph.com/search?q=type:diff+repo:sourcegraph/sourcegraph$+message:%22testing%22) |

## Repository name search

A query with only `repo:` filters returns a list of repositories with matching names.

Example: [`repo:docker repo:registry`](https://sourcegraph.com/search?q=repo:docker+repo:registry)

## Filename search

A query with `type:path` restricts terms to matching filenames only (not file contents).

Example: [`type:path repo:/docker/ registry`](https://sourcegraph.com/search?q=type:path+repo:/docker/+registry)
