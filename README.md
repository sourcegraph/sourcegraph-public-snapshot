# EXPERIMENTAL: Code Extension Protocol (CXP) support for JavaScript/TypeScript

[![build](https://badge.buildkite.com/2eb2e1c2cca148c26ce8d9571f4a8a1d21e3989d10c518feb9.svg)](https://buildkite.com/sourcegraph/cxp-js)
[![codecov](https://codecov.io/gh/sourcegraph/cxp-js/branch/master/graph/badge.svg?token=SLtdKY3zQx)](https://codecov.io/gh/sourcegraph/cxp-js)
[![code style: prettier](https://img.shields.io/badge/code_style-prettier-ff69b4.svg)](https://github.com/prettier/prettier)
[![sourcegraph: search](https://img.shields.io/badge/sourcegraph-search-brightgreen.svg)](https://sourcegraph.com/github.com/sourcegraph/cxp-js)
[![semantic-release](https://img.shields.io/badge/%20%20%F0%9F%93%A6%F0%9F%9A%80-semantic--release-e10079.svg)](https://github.com/semantic-release/semantic-release)

Client, server, and protocol implementations of CXP (Code Extension Protocol), a way to build cross-platform code extensions that run in multiple editors, code hosts, code review/search tools, etc.

**Status:** Experimental

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

## Protocol

TODO

## Usage

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
