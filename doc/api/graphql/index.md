# Sourcegraph GraphQL API

The Sourcegraph GraphQL API is a rich API that exposes data related to the code available on a Sourcegraph instance.

> NOTE: The API is under active development so it may change in backward incompatible ways. These types of changes will be documented in our changelog.

# API overview

The Sourcegraph API has two components: the GraphQL API and the Language Server Protocol Gateway (LSP Gateway).

The GraphQL API supports the following types of queries:

- Full-text and regexp code search
- Rich git-level metadata, including commits, branches, blame information, and file tree data
- Dependency graph queries ("Which repositories use library X?")
- Repository and user metadata

The LSP gateway uses the [Language Server Protocol](https://github.com/Microsoft/language-server-protocol) to provide Code Intelligence features like jump-to-definition, find-references, and auto-completion. The LSP community has created [open-source language servers](http://langserver.org/) for nearly every major language. The LSP Gateway sits in front of these language servers and provides a single endpoint for developer tools and services wishing to offer Code Intelligence for many languages at once.

## Quickstart

Generate an access token on your Sourcegraph instance at:

```none
https://sourcegraph.example.com/user/account/tokens
```

Then run this query to echo your username back:

```shell
curl \
  -H 'Authorization: token YOUR_TOKEN' \
  -d '{"query": "query { currentUser { username } }"}' \
 https://sourcegraph.example.com/.api/graphql
```

You should see a response like this:

```json
{ "data": { "currentUser": { "username": "YOUR_USERNAME" } } }
```

## Documentation & tooling

### API Console

Sourcegraph includes a built-in API console that lets you write queries and view API documentation in your browser.

You can find the API console at any time by clicking **your profile picture** and clicking **API console** in the bottom left of the page (next to **Sign out**), or by visiting it directly at `https://sourcegraph.example.com/api/console`.

If you have not yet set up a Sourcegraph server, you can also test out the API on the [Sourcegraph.com API console](https://sourcegraph.com/api/console) (which always uses the latest version of the API).

### Documentation

Sourcegraph's GraphQL API documentation is available directly in the API console itself. To access the documentation, click **Docs** on the right-hand side of the API console page.

### Sudo access tokens

Site admins may create access tokens with the special `site-admin:sudo` scope, which allows the holder to perform any action as any other user.

```shell
curl \
  -H 'Authorization: token-sudo user="SUDO_TO_USERNAME",token="YOUR_TOKEN"' \
  -d '{"query": "query { currentUser { username } }"}' \
 https://sourcegraph.example.com/.api/graphql
```

This scope is useful when building Sourcegraph integrations with external services where the service needs to communicate with Sourcegraph and does not want to force each user to individually authenticate to Sourcegraph.

## Examples

See ["Sourcegraph GraphQL API examples](examples.md)".

### Using the API via the Sourcegraph CLI

A command line interface to Sourcegraph's API is available. Today, it is roughly the same as using the API via `curl` (see below), but it offers a few nice things:

- Allows you to easily compose queries from scripts, e.g. without worrying about escaping JSON input to `curl` properly.
- Reads your access token and Sourcegraph server endpoint from a config file (or env var).
- Pipe multi-line GraphQL queries into it easily.
- Get any API query written using the CLI as a `curl` command using the `src api -get-curl` flag.

To learn more, see [github.com/sourcegraph/src-cli](https://github.com/sourcegraph/src-cli)

### Using the API via curl

The entire API can be used via `curl` (or any HTTP library), just the same as any other GraphQL API. For example:

```shell
curl \
  -H 'Authorization: token YOUR_TOKEN' \
  -d '{"query":"query($query: String!) { search(query: $query) { results { resultCount } } }","variables":{"query":"Router"}}' \
  https://sourcegraph.com/.api/graphql
```

i.e. you just need to send the `Authorization` header and a JSON object like `{"query": "my query string", "variables": {"var1": "val1"}}`.

## Quickstart for the LSP gateway

To get hover information of the token at [line 321 column 9 of route.go in gorilla/mux](https://sourcegraph.com/github.com/gorilla/mux@c85619274f5d55a67491010f883a91f568b3a9b0/-/blob/route.go#L322:9):

```html
curl 'https://sourcegraph.com/.api/xlang/textDocument/hover' -d @- <<EOF
[
  { "id": 0, "method": "initialize", "params": { "rootUri": "git://github.com/gorilla/mux?c85619274f5d55a67491010f883a91f568b3a9b0", "mode": "go" } },
  { "id": 1, "method": "textDocument/hover", "params": { "textDocument": { "uri": "git://github.com/gorilla/mux?c85619274f5d55a67491010f883a91f568b3a9b0#route.go" }, "position": { "character": 8, "line": 321 } } },
  { "id": 2, "method": "shutdown" },
  { "method": "exit" }
]
EOF
```

The response is a JSON array that looks like this:

```
[
  ...,
  {
    "id": 1,
    "result": {
      "contents": [
        {
          "language": "go",
          "value": "func matchInArray(arr []string, value string) bool"
        },
        "matchInArray returns true if the given string value is in the array. \n\n"
      ],
      "range": {
        "end": {
          "character": 20,
          "line": 321
        },
        "start": {
          "character": 8,
          "line": 321
        }
      }
    },
    "jsonrpc": "2.0"
  },
  ...
]
```

You can use the LSP gateway to run LSP queries against a language server that's running on the Sourcegraph instance. The LSP gateway shares connections between multiple users and dynamically opens/closes connections as needed.
