# Sourcegraph extension API

[![build](https://travis-ci.org/sourcegraph/sourcegraph-extension-api.svg?branch=master)](https://travis-ci.org/sourcegraph/sourcegraph-extension-api)
[![codecov](https://codecov.io/gh/sourcegraph/sourcegraph-extension-api/branch/master/graph/badge.svg?token=SLtdKY3zQx)](https://codecov.io/gh/sourcegraph/sourcegraph-extension-api)
[![code style: prettier](https://img.shields.io/badge/code_style-prettier-ff69b4.svg)](https://github.com/prettier/prettier)
[![sourcegraph: search](https://img.shields.io/badge/sourcegraph-search-brightgreen.svg)](https://sourcegraph.com/github.com/sourcegraph/sourcegraph-extension-api)
[![semantic-release](https://img.shields.io/badge/%20%20%F0%9F%93%A6%F0%9F%9A%80-semantic--release-e10079.svg)](https://github.com/semantic-release/semantic-release)

**Status:** Alpha

> [The extension API you wish your code host had.](https://twitter.com/sqs/status/1037759040233779200)

Sourcegraph extensions enhance reading and reviewing code in your existing tools. They show useful info from services such as Codecov (for test coverage), logging/monitoring/performance tools, and more on GitHub (more code hosts coming soon) and Sourcegraph. It's like being able to use editor extensions when viewing and reviewing code on GitHub and your other favorite tools.

[**🎥 Demo video**](https://www.youtube.com/watch?v=j1eWBa3rWH8)

## Usage

Try the first extension! 🚀

- [**Codecov extension**](https://github.com/sourcegraph/sourcegraph-codecov#readme): See code coverage information from Codecov on GitHub and Sourcegraph

Coming soon: extensions that overlay information (from popular 3rd-party dev tools) on GitHub and Sourcegraph related to:

- logging
- monitoring
- performance
- linting
- profiling
- security
- PR and commit comments

Sourcegraph extensions are in alpha. Please [file issues](https://github.com/sourcegraph/sourcegraph-extension-api/issues) for extension requests, problems, and feedback.

## Authoring extensions

Check out the [hello world tutorial](https://github.com/sourcegraph/sourcegraph-extension-docs/blob/master/tutorials/hello-world.md) and other [documentation](https://github.com/sourcegraph/sourcegraph-extension-docs)

Writing a Sourcegraph extension is very similar to writing an editor extension for [VS Code](https://code.visualstudio.com/docs/extensions/overview). See the [Codecov extension's main `extension.ts` file](https://sourcegraph.com/github.com/sourcegraph/sourcegraph-codecov/-/blob/src/extension.ts) for an example.

## Background

[Sourcegraph](https://sourcegraph.com) provides IDE-like code intelligence (definitions, references, hover tooltips, and search) anywhere you view and review code, such as in GitHub files and pull requests. To do this, we built a way to show information from [language servers](http://langserver.org) (for 20+ languages) in the places you view and review code.

Sometimes that info is enough. But sometimes you need to see other kinds of info: error logs, test coverage, performance, authorship, reachability, lint warnings, etc. For example, if you're reviewing a PR that changes code for which there is [OpenTracing](https://opentracing.io/) trace information or that adds calls to untested code, you should be aware of that while reviewing without having to check an external system.

Existing APIs for code hosts and editors don't make this easy or possible.

#### The state of extension APIs for code hosts

The APIs of code hosts and review tools don't generally provide the necessary hooks to customize the code/diff view and other parts of the UI and workflow. Commit statuses, bot comments, and pipelines are fantastic, but they are solving other problems. Developers need more information and code actions while viewing and reviewing code (just as they do while editing code), and that means providing deeper extensibility.

Even if a single code host implemented a comprehensive extension API, it would be tied to that single code host. Extensions would only be usable by a fraction of developers unless the author created separate extensions for many popular code hosts: GitHub, GitHub Enterprise, Bitbucket, Bitbucket Server, GitLab, Phabricator, etc.

#### The state of extension APIs for editors

Editors generally have comprehensive extension APIs, but they are each tied to a single editor. Initiatives like [Language Server Protocol (LSP)](https://microsoft.github.io/language-server-protocol/) are partial solutions, but LSP still requires an editor extension for each editor (such as [vscode-go](https://github.com/Microsoft/vscode-go)).

As a result, all of the great work that goes into building editor extensions is fragmented among 10+ editors. That makes it less likely that your favorite dev tools have great extensions for your editor of choice. It also makes it difficult to get your entire team using the same editor extensions and configuring them consistently (if your teammates use a variety of editors).

### How Sourcegraph extensions work

To solve this problem (of seeing important contextual info when you're viewing and reviewing code), we made a way for you to write something like an editor extension--except that it can be used on GitHub, Sourcegraph, and (soon) other tools. Check out the code for [an example extension that shows a hello message in hover tooltips](https://sourcegraph.com/github.com/sourcegraph/sourcegraph-hello-world-hover@master/-/blob/src/extension.ts).

After writing and publishing an extension (using `src extensions publish` in the [`src` CLI tool](https://github.com/sourcegraph/src-cli)), anyone can add it and use it on GitHub or Sourcegraph. Behind the scenes (but in open source), 2 other components make this possible:

#### Sourcegraph extension API

The Sourcegraph extension API (this repository) is the open-source extension API that Sourcegraph extensions are written against. It exposes concepts present in all clients, such as notifications of opened/closed documents, adding buttons to toolbars, changing the background or gutter color of lines, showing hover tooltips, etc.

The [sourcegraph](https://npmjs.com/package/sourcegraph) npm package (published from this repository) exposes this as a TypeScript/JavaScript API that is familiar to anyone who has written a [VS Code](https://code.visualstudio.com/) extension.

#### Adapters to "polyfill" existing tools to run Sourcegraph extensions

Adapters make it so that Sourcegraph extensions can modify the UI of existing tools, such as GitHub. They run Sourcegraph extensions and know how to detect when the user hovers over a specific line and column in a file, how to insert buttons and actions seamlessly in the tool's UI, etc. All Sourcegraph client adapters are open source.

- The open-source [Sourcegraph for Chrome](https://github.com/sourcegraph/browser-extensions) supports using Sourcegraph extensions on GitHub (support for other code hosts and browsers coming soon). It runs extensions by executing their JavaScript code on the client in a Web Worker.
- Support for GitHub PRs and diffs, other code hosts, other code review tools, and more browsers is coming soon.
- Sourcegraph for `$EDITOR` is coming soon (for VS Code, Vim, Emacs, Atom, Sublime, JetBrains, etc.).

## Development

```shell
yarn
yarn test
```

## Acknowledgments

The Sourcegraph extension API and architecture is inspired by the [VS Code extension API](https://code.visualstudio.com/docs/extensions/overview) and the [Language Server Protocol (LSP)](https://microsoft.github.io/language-server-protocol/).
