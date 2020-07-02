# Structural search

Structural search lets you match richer syntax patterns specifically in code
and structured data formats like JSON. It can be awkward or difficult to match
code blocks or nested expressions with regular expressions. To meet this
challenge we've introduced a new and easier way to search code that operates
more closely on a program's parse tree. We use [Comby
syntax](https://comby.dev/#match-syntax) for structural matching. Below you'll
find examples and notes for this new search functionality.

### Example

The `fmt.Sprintf` function is a popular print function in Go. Here is a pattern
that matches all the arguments in `fmt.Sprintf` calls in our code:

```go
fmt.Sprint(:[args])
```

[See it live on Sourcegraph's code](https://sourcegraph.com/search?q=repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24++fmt.Sprintf%28:%5Bargs%5D%29&patternType=structural)

The `:[args]` part is a hole with a descriptive name `args` that matches
code.  The important part is that this pattern understands that the parentheses
`(` `)` are balanced, and avoids character escaping and lookahead assertions
that come up in regular expressions. Let's look at two interesting variants of matches in our codebase. Here is the first:

```go
fmt.Sprintf("must be authenticated as an admin (%s)", isSiteAdminErr.Error())
```

Note that to match this code we didn't have to do any special thinking about
handling the parentheses `(%s)` that happen _inside_ the first string argument,
or the nested parentheses that form part of `Error()`. Taking care to match the
closing parentheses for the call could, in general, really complicate regular
expressions.

Here is a second match:

```go
fmt.Sprintf(
		"rest/api/1.0/projects/%s/repos/%s/pull-requests/%d",
		pr.ToRef.Repository.Project.Key,
		pr.ToRef.Repository.Slug,
		pr.ID,
	)
```

In this case, we didn't have to do any special thinking about handling multiple
lines. The hole `:[args]` by default matches across newlines, but it stops
inside balanced parentheses. This lets us match large, logical blocks or
expressions without the limitations of typical line-based regular expression
patterns. For [more examples](#examples), see below. Next, we cover
reference notes to consider when using structural search.

### Current functionality and restrictions

Structural search behaves differently to plain text search in key ways. We are
continually improving functionality of this new feature, so please note the
following:

- **Only indexed repos.** Structural search can currently only be performed on _indexed_ repositories. See [configuration](#configuration) for more details if you host your own Sourcegraph installation. Our service hosted at [sourcegraph.com](https://sourcegraph.com/search) indexes approximately 10,000 of the most popular repositories on GitHub. Other repositories are currently unsupported.

- **The `lang` keyword is semantically significant.** Adding the `lang` [keyword](queries.md) informs the parser about language-specific syntax for comments, strings, and code. This makes structural search more accurate for that language. For example, `patterntype:structural 'fmt.Sprintf(:[args])' lang:go`. If `lang` is omitted, we perform a best-effort to infer the language based on matching file extensions, or fall back to a generic structural matcher.

- **Enclosing patterns with quotes.** Prior to version 3.17, we recommended you enclose a pattern with quotes, in case the pattern conflicts with other query syntax. As of version 3.17, quotes should no longer be included, unless the intent is to match actual quotes. To avoid syntax conflicts in version 3.17 and onwards, use the `content:` parameter.

- **Saved search are not supported.** It is not currently possible to save structural searches.

- **Matching blocks in indentation-sensitive languages.** It's not currently possible to match blocks of code that are identation-sensitive. This is a feature planned for future work.

### Syntax reference

Here is a summary of syntax for structural matching, which is based on [Comby syntax](https://comby.dev/docs/syntax-reference).

- `:[hole]` matches zero or more characters (including whitespace, and across
newlines) in a lazy fashion. When `:[hole]` is used inside delimiters, as in
`{:[h1], :[h2]}` or `(:[h])`, those delimiters set a boundary for what the hole
can match, and the hole will then only match patterns within those delimiters.
Holes can be used outside of delimiters as well.

- `:[[hole]]` matches one or more alphanumeric characters and `_`.

- `:[hole.]` (with a period at the end) matches one or more alphanumeric characters and punctuation (like `.`, `;`, and `-`).

- `:[hole\n]` (with a `\n` at the end) matches one or more characters up to a newline, including the newline.

- `:[ ]` (with a space) matches only whitespace characters, excluding newlines. To assign the matched whitespace to variable, put the variable name after the space, like :`[ hole]`.


**Rules.** [Comby supports rules](https://comby.dev/#advanced-usage) to express equality constraints or pattern-based matching. Comby rules are not officially supported in Sourcegraph yet. We are in the process of making that happen and are taking care to address stable performance and usability. That said, you can explore rule functionality with an experimental `rule:` parameter. For [example](https://sourcegraph.com/search?q=repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+%22buildSearchURLQuery%28:%5Barg%5D%2C+:%5B_%5D%29%22+rule:%27where+:%5Barg%5D+%3D%3D+%22navbarQuery%22%27&patternType=structural), `"buildSearchURLQuery(:[arg], :[_])" rule:'where :[arg] == "navbarQuery"'`.

### Examples

Here are some more examples. Also see our [blog post](https://about.sourcegraph.com/blog/going-beyond-regular-expressions-with-structural-code-search) for further examples.

#### Match stringy data

Taking our [original example](#example), let's modify the original pattern
slightly to match only if the first argument is a string. We do this by adding
string quotes around a hole called `format`. Adding quotes communicates
_structural context_ and changes how the hole behaves: it will match the
contents of a single string delimited `"`. It _won't_ match multiple strings
like `"foo", "bar"`. We match remaining arguments with the hole called `args`.

```go
fmt.Sprintf(":[format]", :[args])
```

[See it live on Sourcegraph's code](https://sourcegraph.com/search?q=repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+fmt.Sprintf%28%22:%5Bformat%5D%22%2C+:%5Bargs%5D%29&patternType=structural)

We've matched some examples, like:

```go
fmt.Sprintf("external service not found: %v", e.id)
```

```go
fmt.Sprintf("%s/campaigns/%s", externalURL, string(campaignID))
```


Holes stop matching based on the first fragment of syntax that comes after it,
similar to lazy regular expression matching. So, we could write:

```go
fmt.Sprintf(:[first], :[second], :[rest])
```

to match all functions with three or more arguments, matching the the first and second arguments based on the contextual position around the commas.

#### Match equivalent expressions

Using the same identifier in multiple holes adds a constraint that both of the matched values must be syntactically equal. So, the pattern:

```go
return :[v.], :[v.]
```

will match code where a pair of identifier-like syntax in the `return` statement are the same. For example, `return true, true`, `return nil, nil`, or `return 0, 0`.

[See it live on Sourcegraph's code](https://sourcegraph.com/search?q=repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+lang:go+return+:%5Bv.%5D%2C+:%5Bv.%5D&patternType=structural)

#### Match JSON

Structural search also works on structured data, like JSON. Patterns can declaratively describe pieces of data to match. For example the pattern:

```json
"exclude": [:[items]]
```

matches all parts of a JSON document that have a member `"exclude"`, where the value is a list of items.

[See it live on Sourcegraph's code](https://sourcegraph.com/search?q=repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24++%22exclude%22:+%5B:%5Bitems%5D%5D+lang:json+file:tsconfig.json&patternType=structural)

### Configuration

**Indexed repositories.** Structural search only works for indexed repositories. To see whether a repository on your instance is indexed, visit `https://<sourcegraph-host>.com/repo-org/repo-name/-/settings/index`.

**Disabling structural search.** Disable structural search on your instace by adding the following to the site configuration:

```json
{
  "experimentalFeatures": {
      "structuralSearch": "disabled"
  }
}
```
