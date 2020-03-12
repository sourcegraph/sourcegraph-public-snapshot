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

[See it live on Sourcegraph's code](https://sourcegraph.com/search?q=repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24++%27fmt.Sprintf%28:%5Bargs%5D%29%27&patternType=structural)

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
patterns. For [more examples](#more-examples), see below. Next, we cover a
reference notes to consider when using structural search.

### Current functionality and restrictions

Structural search behaves differently to plain text search in key ways. We are
continually improving functionality of this new feature, so please note the
following:

- **Only indexed repos.** Structural search can currently only be performed on _indexed_ repositories. See [configuration](#configuration) for more details if you host your own Sourcegraph installation. Our service hosted at [sourcegraph.com](https://sourcegraph.com/search) indexes approximately 10,000 of the most popular repositories on GitHub. Other repositories are currently unsupported.

- **Enclose patterns with quotes.** When entering the pattern in the browser search bar or `src-cli` command line, always enclose the pattern with quotes: `'fmt.Sprintf(:[args])'`. Quotes that are part of the pattern can be escaped with `\`.

- **The `lang` keyword is semantically significant.** Adding the `lang` [keyword](queries.md) informs the parser about language-specific syntax for comments, strings, and code. This makes structural search more accurate for that language. For example, `patterntype:structural 'fmt.Sprintf(:[args])' lang:go`. If `lang` is omitted, we perform a best-effort to infer the language based on matching file extensions, or fall back to a generic structural matcher.

- **Saved search are not supported.** It is not currently possible to save structural searches.

- **Matching blocks in indentation-sensitive languages.** It's not currently possible to match blocks of code that are identation-sensitive. This is a feature planned for future work.

### Syntax reference

Here is a summary of syntax for structural matching, which is based on [Comby syntax](https://comby.dev/#match-syntax).

- `:[hole]` matches zero or more characters (including whitespace, and across
newlines) in a lazy fashion. When `:[hole]` is used inside delimiters, as in
`{:[h1], :[h2]}` or `(:[h])`, those delimiters set a boundary for what the hole
can match, and the hole will then only match patterns within those delimiters.
Holes can be used outside of delimiters as well.

- `:[[hole]]` matches one or more alphanumeric characters and `_`.

- `:[hole.]` (with a period at the end) matches one or more alphanumeric characters and punctuation (like `.`, `;`, and `-`).

- `:[hole\n]` (with a `\n` at the end) matches one or more characters up to a newline, including the newline.

- `:[ ]` (with a space) matches only whitespace characters, excluding newlines. To assign the matched whitespace to variable, put the variable name after the space, like :`[ hole]`.


**Rules** [Comby supports rules](https://comby.dev/#advanced-usage) to express equality constraints or pattern-based matching. Comby rules are not officially supported in Sourcegraph yet. We are in the process of making that happen and are taking care to address stable performance and usability. That said, you can explore rule functionality with an experimental `rule:` parameter. For [example](https://sourcegraph.com/search?q=repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+%22buildSearchURLQuery%28:%5Barg%5D%2C+:%5B_%5D%29%22+rule:%27where+:%5Barg%5D+%3D%3D+%22navbarQuery%22%27&patternType=structural), `"buildSearchURLQuery(:[arg], :[_])" rule:'where :[arg] == "navbarQuery"'`.

### Examples

Here are some more examples. Also see our [blog post](https://about.sourcegraph.com/blog/going-beyond-regular-expressions-with-structural-code-search) for further examples.

#### Match stringy data

Taking our [original example](#example), let's modify the original pattern
slightly to match only if the first (and only) argument is a string. We do this
by adding string quotes. Adding quotes communicates _structural context_ and
changes how the hole behaves: it will match the contents of a single string
delimited `"`. It _won't_ match multiple strings like `"foo", "bar"`.

```go
fmt.Sprintf(":[str]")
```

[See it live on Sourcegraph's code](https://sourcegraph.com/search?q=repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24++%27fmt.Sprintf%28%22:%5Bargs%5D%22%29%27&patternType=structural)

We've matched some interesting examples, like:

```go
q.WriteString(fmt.Sprintf("nodes{ ... pr }\n"))
```

In fact, a single string is passed to `fmt.Sprintf` here without any format
specifiers, so this `fmt.Sprintf` call is unnecessary. We could just write `q.WriteString("nodes{ ... pr }\n")`. Looks like we have some
cleaning up to do. 

#### Match function arguments contextually

If we wanted to instead match on the first argument of `fmt.Sprintf` calls with more than one argument, we could write:

```go
fmt.Sprint(:[first], :[rest])
```

This pattern matches all of the code leading up to the comma `,` in
`:[first]`. _All_ of the rest of the arguments match to `:[rest]`. Holes stop
matching based on the first fragment of syntax that comes after it,
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

[See it live on Sourcegraph's code](https://sourcegraph.com/search?q=+lang:go+%27return+:%5Bv.%5D%2C+:%5Bv.%5D%27&patternType=structural)

#### Match JSON

Structural search also works on structured data, like JSON. Patterns can declaratively describe pieces of data to match. For example the pattern:

```json
"exclude": [:[items]]
```

matches all parts of a JSON document that have a member `"exclude"`, where the value is a list of items. 

[See it live on Sourcegraph's code](https://sourcegraph.com/search?q=repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24++%27%22exclude%22:+%5B:%5Bitems%5D%5D%27+lang:json&patternType=structural)

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
