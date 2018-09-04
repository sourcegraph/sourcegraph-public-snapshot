# Code Extension Protocol (CXP)

[![build](https://travis-ci.org/sourcegraph/cxp-js.svg?branch=master)](https://travis-ci.org/sourcegraph/cxp-js)
[![codecov](https://codecov.io/gh/sourcegraph/cxp-js/branch/master/graph/badge.svg?token=SLtdKY3zQx)](https://codecov.io/gh/sourcegraph/cxp-js)
[![code style: prettier](https://img.shields.io/badge/code_style-prettier-ff69b4.svg)](https://github.com/prettier/prettier)
[![sourcegraph: search](https://img.shields.io/badge/sourcegraph-search-brightgreen.svg)](https://sourcegraph.com/github.com/sourcegraph/cxp-js)
[![semantic-release](https://img.shields.io/badge/%20%20%F0%9F%93%A6%F0%9F%9A%80-semantic--release-e10079.svg)](https://github.com/semantic-release/semantic-release)

Client, server, and protocol implementations of CXP (Code Extension Protocol), a way to build cross-platform code extensions that run in multiple editors, code hosts, code review/search tools, etc.

**Status:** Alpha

## Usage

### On Sourcegraph.com

Try the [sourcegraph-codecov](https://github.com/sourcegraph/sourcegraph-codecov) extension by visiting any file that has Codecov code coverage, such as [https://sourcegraph.com/github.com/theupdateframework/notary@fb795b0bc868746ed2efa2cd7109346bc7ddf0a4/-/blob/server/storage/tuf_store.go](tuf_store.go).

### On GitHub using the Chrome extension

See [demo video](https://www.youtube.com/watch?v=j1eWBa3rWH8).

1. Install [Sourcegraph for Chrome](https://chrome.google.com/webstore/detail/sourcegraph/dgjhfomjieaadpoljlnidmbgkdffpack)
2. Open the Sourcegraph Chrome extension options page (by clicking the Sourcegraph icon in the Chrome toolbar)
3. Check the box labeled **Use Sourcegraph extensions (enable CXP)** to enable this alpha feature
4. Visit [tuf_store.go on GitHub](https://github.com/theupdateframework/notary/blob/fb795b0bc868746ed2efa2cd7109346bc7ddf0a4/server/storage/tuf_store.go)
5. Click the `Coverage: N%` button to show Codecov test coverage background colors on the file [sourcegraph-codecov](https://github.com/sourcegraph/sourcegraph-codecov), and scroll down to see them

Support for more tools will be added soon.

### The problems

1.  Editor extensions are tied to a single editor, so all the great work that goes into building editor extensions is fragmented among 10+ editors.
2.  Most dev teams use a variety of editors, which means that it's hard to configure tooling (such as code intelligence or linters) in a consistent way in everyone's editors.
3.  Other coding tools you use (such as code hosts and code review/search tools) don't support extensions, so you lose useful context and functionality when interacting with code outside your editor.

### The (proposed) solution

CXP will let you build an extension once and use it _everywhere_ you edit or view code.

### Components

- **CXP protocol:** an RPC protocol that is a rough superset of [Language Server Protocol (LSP)](https://microsoft.github.io/language-server-protocol/), spoken between CXP extensions and CXP clients (such as editors).
- **CXP extension:** like an editor extension, except written to speak CXP instead of the editor-specific extension API. It exposes functionality such as code intelligence, linting, Git line blaming, code coverage, etc., just like today's existing editor extensions.
- **CXP client:** editors, code hosts (such as GitHub, specifically the file, PR "Files changed", and diff pages), code review tools, code search tools, etc.

### Examples

TODO

## Development

```shell
npm install
npm test
```

## Implementation

This library does not depend on [Microsoft/vscode-languageserver-node](https://github.com/Microsoft/vscode-languageserver-node) (except for the types) because [Microsoft/vscode-languageserver-node](https://github.com/Microsoft/vscode-languageserver-node) uses Node.js APIs that are not available in the browser and its footprint is rather large (1+ MB).

## Acknowledgments

The CXP protocol is based on the [Language Server Protocol (LSP)](https://microsoft.github.io/language-server-protocol/). This library is based on the [Microsoft/vscode-languageserver-node](https://github.com/Microsoft/vscode-languageserver-node) implementation of LSP.
