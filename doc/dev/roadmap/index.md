# Sourcegraph roadmap

We want Sourcegraph to be the best way to answer questions while writing, reviewing, or planning code. This roadmap shows what's planned for upcoming Sourcegraph releases. See the [Sourcegraph master plan](https://about.sourcegraph.com/plan) for our high-level product vision.

A new Sourcegraph release ships on the [20th day of each month](../releases.md#releases-are-monthly). The plans and timeframes are subject to change.

## Releases

### 3.3

Release date: 2019-04-20 ([draft announcement](https://docs.google.com/document/d/19SsZ00UdA7WZFIXSCOaJVgP1Ngu7l8HgGeArT_iDbhg/edit))

- Core services
  - [Keep repository set in sync with config](https://github.com/sourcegraph/sourcegraph/issues/2025)
- [Distribution](https://github.com/sourcegraph/sourcegraph/issues/2809)
- [Documentation](https://github.com/sourcegraph/sourcegraph/issues/2848)
- [Code search](https://github.com/sourcegraph/sourcegraph/issues/2740)
  - Working toward [subquery search](https://github.com/sourcegraph/sourcegraph/issues/1005)
  - [Saved search improvements](https://github.com/sourcegraph/sourcegraph/issues/2824)
- Code navigation
  - [Integrations quality](https://github.com/sourcegraph/sourcegraph/issues/2834)
  - [Code intelligence](https://github.com/sourcegraph/sourcegraph/issues/2856)

### [Previous releases](previous_releases.md)

See [previous Sourcegraph releases](previous_releases.md).

---

## Themes

We want Sourcegraph to be the best way to answer questions while writing, reviewing, and reading code. See the [Sourcegraph master plan](https://about.sourcegraph.com/plan) for our high-level product vision.

Our work generally falls into the following categories:

- **Code search and navigation:** quickly showing you the code you're looking for and making it easy to navigate around
- **Code intelligence:** go-to-definition, hover tooltips, references, symbols, etc., for code in many languages, including real-time and cross-repository support
- **Integrations:** making Sourcegraph work well with code hosts, review tools, editors, and other tools in your dev workflow (e.g., repository syncing from your code host, browser extensions, and editor extensions)
- **Extensibility:** supporting [Sourcegraph extensions](../../extensions/index.md) that add code intelligence and other information (e.g., tracing, logging, and security annotations from 3rd-party tools) to Sourcegraph and external tools that Sourcegraph integrates with
- **Distribution:** making it easy to deploy and manage a self-hosted Sourcegraph instance
- **Enterprise:** features that larger companies need (e.g., scaling, authentication, authorization, auditing, etc.)

---

## Feature areas

### Cross-feature-area

- Better support for searching open-source repositories AND internal code

### Search

- "Did you mean ...?" for common mistakes in search queries [#2772](https://github.com/sourcegraph/sourcegraph/issues/2772)
- Nested search queries (e.g., in all repositories whose `package.json` contains `foo`, find matches of `bar`) [#1005](https://github.com/sourcegraph/sourcegraph/issues/1005)
  - Multi-line searches (#35)(https://github.com/sourcegraph/sourcegraph/issues/35)
- More advanced search filters (provided by extensions), such as those using language-specific or dependency graph information <!-- TODO -->
- More types/sources of search results (provided by extensions), such as documentation (wiki, Markdown, and Google Docs), issues, PR comments, and configuration data [#738](https://github.com/sourcegraph/sourcegraph/issues/738)
- Improved search relevance and ranking <!-- TODO -->

### Code navigation and intelligence

- Refine search-based (non-language-server-based) code intelligence for [all languages](https://sourcegraph.com/extensions?query=category%3A%22Programming+languages%22)
  - 
- Analyze and expose dependency graph for all major languages and build systems <!-- TODO -->
- Programming language statistics <!-- TODO -->
- Cross-language, cross-repository definitions and references support for APIs/IDLs (GraphQL, Thrift, Protobuf, etc) [#981](https://github.com/sourcegraph/sourcegraph/issues/981)
- Language-server-based (precise) code intelligence for [more languages](https://sourcegraph.com/extensions?query=tag%3Alanguage-server)

## Features

Search

- Improvements to saved searches

Code intelligence and navigation

- [Java language support via extension](https://github.com/sourcegraph/sourcegraph/issues/1400)
- [Python dependency fetching and cross repository references](https://github.com/sourcegraph/sourcegraph/issues/1401)
- [Swift language support via extension](https://github.com/sourcegraph/sourcegraph/issues/979) (likely includes Objective-C, C, and C++)
- [Thrift code intelligence](https://github.com/sourcegraph/sourcegraph/issues/669)
- [Flow (JavaScript) language support](https://github.com/sourcegraph/sourcegraph/issues/982)
- [Scoped symbols sidebar](https://github.com/sourcegraph/sourcegraph/issues/1967)
- PHP language support via extension
- Bazel support

Sourcegraph extensions

- [Extension registry discovery and statistics](https://github.com/sourcegraph/sourcegraph/issues/980)
- Enhanced extensions for Codecov and Datadog
- New 3rd-party extensions: Sentry, LightStep, FOSSA, SonarQube, [LaunchDarkly](https://github.com/sourcegraph/sourcegraph/issues/1249), Figma, etc.
- [Configuration data search extension](https://github.com/sourcegraph/sourcegraph/issues/670)
- Improved code host support for Sourcegraph extensions
- [Using Sourcegraph extensions in the editor](https://github.com/sourcegraph/sourcegraph/issues/978)
- [Sourcegraph extension testing](https://github.com/sourcegraph/sourcegraph/issues/733)

Other

- [Direct UI integration and deployment bundling with GitLab](https://github.com/sourcegraph/sourcegraph/issues/1000)
- [Checklist-based repository reviews](https://github.com/sourcegraph/sourcegraph/issues/1526)
- [Browser authorization flow for clients](https://github.com/sourcegraph/sourcegraph/pull/528)
- Enhanced notification preferences
- Support for non-Git version control systems (Perforce, Subversion, TFS, etc.)
- API access logging

<!--

Prior art:

https://about.gitlab.com/direction
https://docs.microsoft.com/en-us/visualstudio/productinfo/vs-roadmap

-->
