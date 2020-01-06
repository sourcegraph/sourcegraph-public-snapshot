# User documentation

Welcome to Sourcegraph! As a developer, you can use Sourcegraph to get help while writing and reviewing code.

- [Tour](tour.md): A walkthrough of Sourcegraph's features, with real-world example use cases.
- Integrations:
  - [Browser extension](../integration/browser_extension.md) (adds go-to-definition, hover tooltips, etc., to your code host and review tool)
  - [Browser search engine](../integration/browser_search_engine.md)
  - [Editor extension](../integration/editor.md)
- [Code search](search/index.md)
  - [Query syntax](search/queries.md): Supported query operators.
  - [Search examples](search/examples.md)
  - Types of searches:
    - Cross-repository search
    - Full-text search (with regular expression support)
    - Repository name search
    - Filename search
    - Diff search
    - Commit message search
    - Multi-branch search
- [Code intelligence](code_intelligence/index.md)
  - Hover tooltips (type signatures, docs, etc.)
  - Jump-to-definition
  - Find-references
  - Cross-repository jump-to-definition and find-references
  - Supports [Go](https://sourcegraph.com/extensions/sourcegraph/go), [TypeScript](https://sourcegraph.com/extensions/sourcegraph/typescript), [Python](https://sourcegraph.com/extensions/sourcegraph/python) - check the [extension registry](https://sourcegraph.com/extensions?query=category%3A%22Programming+languages%22) for more
- [GraphQL API](../api/graphql/index.md)
- [Repository badges](repository/badges.md)

All features:

- [Usage statistics](usage_statistics.md)
- [User surveys](user_surveys.md)
- [Color themes](themes.md)
- [Automation preview](automation.md)
- [Quick links](quick_links.md)

## What is Sourcegraph?

Sourcegraph is a code search and intelligence tool for developers. It lets you search and explore all of your organization's code on the web, with integrations into your existing tools.

## What does Sourcegraph do?

Sourcegraph's main features are:

- [Code search](#code-search): fast, up-to-date, and scalable, with regexp support on any branch or commit without an indexing delay (and diff search)
- [Code intelligence](#code-intelligence): jump-to-definition, find references, and other smart, IDE-like code browsing features on any branch, commit, or PR/code review
- Easy and secure self-hosted installation (your code never touches our servers)
- [Integrations](#integrations) with code hosts, code review tools, editors, web browsers, etc.

## What do I use Sourcegraph for?

Sourcegraph helps you:

- Find example code
- Explore/read code (including during a code review)
- Debug issues
- Locate a specific piece of code
- Determine the impact of changes

Sourcegraph makes it faster and easier to perform these tasks, for you and everyone else at your organization.

## Who should use Sourcegraph?

All developers, except:

- Sourcegraph is more useful to developers working with larger codebases or teams (15+ developers).
- If you rarely search, read, or review code, you probably won't find Sourcegraph useful. (But you should start reading and reviewing more code!)

## Who else uses Sourcegraph?

In addition to the [companies listed on about.sourcegraph.com](https://about.sourcegraph.com), many large technology companies with 500-3,000+ engineers use Sourcegraph internally.

Also, both Facebook and Google provide an in-house Sourcegraph-like code search and intelligence tool to their employees. A [published research paper from Google](https://static.googleusercontent.com/media/research.google.com/en//pubs/archive/43835.pdf) and a [Google developer survey](https://docs.google.com/document/d/1LQxLk4E3lrb3fIsVKlANu_pUjnILteoWMMNiJQmqNVU/edit#heading=h.xxziwxixfqq3) showed that 98% of developers surveyed consider their Sourcegraph-like internal tool to be critical. Developers use it on average for 5.3 sessions each day. (Facebook's and Google's in-house tools are not available to other companies; use Sourcegraph instead.)

## How do I start using Sourcegraph?

1.  [Install Sourcegraph](../admin/install/index.md) inside your organization on your internal code, if nobody else has yet
1.  Install and configure the [web browser code host integrations](../integration/browser_extension.md) (recommended)
1.  Start searching and browsing code on Sourcegraph by visiting the URL of your organization's internal Sourcegraph instance

You can also try [Sourcegraph.com](https://sourcegraph.com/search), which is a public instance of Sourcegraph for use on open-source code only.

---

## Code intelligence

Sourcegraph gives your development team cross-repository IDE-like features on your code:

- Hover tooltips
- Go-to-definition
- Find references
- Symbol search

Sourcegraph gives you code intelligence in:

- **code files in Sourcegraph's web UI**

![Hover tooltip](../user/code_intelligence/img/hover-tooltip.png)

- **diffs in your code review tool**, via [integrations](../integration/index.md)

![GitHub pull request integration](../integration/img/GitHubDiff.png)

- **code files on your code host**, via [integrations](../integration/index.md)

![GitHub file integration](img/GitHubFile.png)

Read the [code intelligence documentation](../user/code_intelligence/index.md) to learn more and to set it up.

---

## Code search

Sourcegraph code search is fast, works across all your repositories at any commit, and has no indexing delay. Code search also includes advanced features, including:

- Powerful, flexible query syntax
- Commit diff search
- Commit message search
- Custom search scopes
- Saved search monitoring

Read the [code search documentation](search/index.md) to learn more and discover the full feature set.

---

## Integrations

Sourcegraph allows you to get code intelligence and code search on code files and code review diffs in your code host and review tool. See our [integrations documentation](../integration/index.md) to set up Sourcegraph with your tools and roll it out to your entire team.
