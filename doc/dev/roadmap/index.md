# Sourcegraph roadmap

<aside class="note visible-product"><p>View the <strong><a href="https://docs.sourcegraph.com/dev/roadmap">latest roadmap on docs.sourcegraph.com</a></strong>.</p></aside>

We want Sourcegraph to be the best way to answer questions while writing, reviewing, or planning code. This roadmap shows what's planned for upcoming Sourcegraph releases. See the [Sourcegraph master plan](https://about.sourcegraph.com/plan) for our high-level product vision.

A new Sourcegraph release ships in the first week of each month. The plans and timeframes are subject to change.

We welcome suggestions! Share feedback by using [code discussions](https://about.sourcegraph.com/blog/discuss-code-and-docs-in-repositories) on this document or the linked documents and issues.

## Future releases

### 3.1

Release date: 2019-02-20

Core Services:

- [Upgrade to PostgreSQL 10.6](https://github.com/sourcegraph/sourcegraph/issues/1404)

Code search and navigation:

- Non-precise code intelligence improvements
- [Code navigation UX and robustness](https://github.com/sourcegraph/sourcegraph/issues/2070)
- [Search UX](https://github.com/sourcegraph/sourcegraph/issues/2081)


<small>[All 3.1 issues](https://github.com/issues?utf8=%E2%9C%93&q=is%3Aissue+is%3Aopen+archived%3Afalse+sort%3Aupdated-desc+org%3Asourcegraph+milestone%3A3.1)</small>

---

### 3.2

TODO: We are currently planning 3.2.

Release date: 2019-03-20

Core services:

- [Keep repository set in sync with config](https://github.com/sourcegraph/sourcegraph/issues/2025)

Distribution:

- Automate cloud provider deployment process
- [Improve scaling documentation](https://github.com/sourcegraph/sourcegraph/issues/2019)
- [Instance activation flow for site admins](https://github.com/sourcegraph/sourcegraph/issues/2069)
- Docs UX (TBD)
- [Better docs for setting up an instance ready for sharing](https://github.com/sourcegraph/sourcegraph/issues/2277) (TBD)

Code search:

- [Search UX](https://github.com/sourcegraph/sourcegraph/issues/2081)

Code navigation:

- Integrations foundation (details TBD)
- Code intelligence (details TBD)

---

### Future

Search

- [Multi-line searches](https://github.com/sourcegraph/sourcegraph/issues/35)
- Improvements to saved searches

Code intelligence and navigation

- [Java language support via extension](https://github.com/sourcegraph/sourcegraph/issues/1400)
- [Python dependency fetching and cross repository references](https://github.com/sourcegraph/sourcegraph/issues/1401)
- [Swift language support via extension](https://github.com/sourcegraph/sourcegraph/issues/979) (likely includes Objective-C, C, and C++)
- [Thrift code intelligence](https://github.com/sourcegraph/sourcegraph/issues/669)
- [Cross-language API/IDL support](https://github.com/sourcegraph/sourcegraph/issues/981) (followup from 3.0)
- [Flow (JavaScript) language support](https://github.com/sourcegraph/sourcegraph/issues/982)
- [Scoped symbols sidebar](https://github.com/sourcegraph/sourcegraph/issues/1967)
- PHP language support via extension

Sourcegraph extensions

- [Extension registry discovery and statistics](https://github.com/sourcegraph/sourcegraph/issues/980)
- Codecov extension
- More 3rd-party extensions: Sentry, FOSSA, SonarQube, [LaunchDarkly](https://github.com/sourcegraph/sourcegraph/issues/1249), Figma
- [Configuration data search extension](https://github.com/sourcegraph/sourcegraph/issues/670)
- Improved code host support for Sourcegraph extensions
- [Using Sourcegraph extensions in the editor](https://github.com/sourcegraph/sourcegraph/issues/978)
- [Sourcegraph extension testing](https://github.com/sourcegraph/sourcegraph/issues/733)

Other

- [Direct UI integration and deployment bundling with GitLab](https://github.com/sourcegraph/sourcegraph/issues/1000)
- [Checklist-based repository reviews](https://github.com/sourcegraph/sourcegraph/issues/1526)
- [Browser authorization flow for clients](https://github.com/sourcegraph/sourcegraph/pull/528)
- Enhanced notification preferences
- API access logging

---

## Themes

We want Sourcegraph to be the best way to answer questions while writing, reviewing, or planning code. See the [Sourcegraph master plan](https://about.sourcegraph.com/plan) for our high-level product vision.

Our work generally falls into the following categories:

- **Search and browsing:** quickly showing you the code you're looking for and making it easy to navigate around
- **Code intelligence:** go-to-definition, hover tooltips, references, symbols, etc., for code in many languages, including real-time and cross-repository support
- **Integrations:** making Sourcegraph work well with code hosts, review tools, editors, and other tools in your dev workflow (e.g., repository syncing from your code host, browser extensions, and editor extensions)
- **Extensibility:** supporting Sourcegraph extensions that add code intelligence and other information (e.g., tracing, logging, and security annotations from 3rd-party tools) to Sourcegraph and external tools that Sourcegraph integrates with
- **Deployment:** making it easy to run and maintain a self-hosted Sourcegraph instance
- **Enterprise:** features that larger companies need (e.g., scaling, authentication, authorization, auditing, etc.)

## [Previous releases](previous_releases.md)

See [previous Sourcegraph releases](previous_releases.md).


<!--

Prior art:

https://about.gitlab.com/direction
https://docs.microsoft.com/en-us/visualstudio/productinfo/vs-roadmap

-->
