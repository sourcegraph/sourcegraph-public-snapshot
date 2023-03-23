# Sourcegraph search GraphQL API

This page adds some additional depth and context to Sourcegraph's search GraphQL API.

For general information about the GraphQL API and how to use it, see [this page instead](index.md).

## `src` CLI usage (easier than GraphQL)

Putting together a comprehensive GraphQL search query can be difficult. For this reason, we created the [`src` CLI tool](https://sourcegraph.com/github.com/sourcegraph/src-cli) which allows you to simply run a search query and get the JSON results without constructing the GraphQL query:

```
export SRC_ENDPOINT=https://sourcegraph.com
export SRC_ACCESS_TOKEN=secret

src search -json 'repo:pallets/flask error'
```

You can then consume the JSON output directly, add `--get-curl` to get a `curl` execution line, and more. See [the `src` CLI tool](https://sourcegraph.com/github.com/sourcegraph/src-cli) for more details.

