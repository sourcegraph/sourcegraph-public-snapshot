# Sourcegraph roadmap

<aside class="note visible-product"><p>View the <strong><a href="https://docs.sourcegraph.com/dev/roadmap">latest roadmap on docs.sourcegraph.com</a></strong>.</p></aside>

We want Sourcegraph to be the best way to answer questions while writing, reviewing, or planning code. This roadmap shows what's planned for upcoming Sourcegraph releases. See the [Sourcegraph master plan](https://about.sourcegraph.com/plan) for our high-level product vision.

A new Sourcegraph release ships in the first week of each month. The plans and timeframes are subject to change.

We welcome suggestions! Share feedback by using [code discussions](https://about.sourcegraph.com/blog/discuss-code-and-docs-in-repositories) on this document or the linked documents and issues.

## Future releases

### 3.0-preview (week of 2018-12-03)

- [In-product site configuration](https://github.com/sourcegraph/sourcegraph/issues/965)
- [External services UI](https://github.com/sourcegraph/sourcegraph/pull/1103)
- [Sub-query searches](https://github.com/sourcegraph/sourcegraph/issues/1005)
- [Search GitHub issues](https://github.com/sourcegraph/sourcegraph/issues/962)
- [Simpler browser extension options menu](https://github.com/sourcegraph/sourcegraph/issues/961)
- [GitHub user authentication](https://github.com/sourcegraph/sourcegraph/issues/964)
- [Go language support via extension](https://github.com/sourcegraph/sourcegraph/issues/958)
- [JavaScript/TypeScript language support via extension](https://github.com/sourcegraph/sourcegraph/issues/960)
- [Python language support via extension](https://github.com/sourcegraph/sourcegraph/issues/959)
- [Improved Sourcegraph extension documentation](https://github.com/sourcegraph/sourcegraph/issues/1151)

<small>[Draft announcement](https://github.com/sourcegraph/about/pull/49) --- [All 3.0-preview issues](https://github.com/issues?utf8=%E2%9C%93&q=is%3Aissue+is%3Aopen+archived%3Afalse+sort%3Aupdated-desc+repo%3Asourcegraph%2Fsourcegraph-extension-api+repo%3Asourcegraph%2Fsourcegraph+repo%3Asourcegraph%2Fenterprise+repo%3Asourcegraph%2Fsourcegraph-extension-api+repo%3Asourcegraph%2Fbrowser-extensions+repo%3Asourcegraph%2Fextensions-client-common+repo%3Asourcegraph%2Fsrc-cli+repo%3Asourcegraph%2Fcodeintellify+repo%3Asourcegraph%2Fgo-langserver+repo%3Asourcegraph%2Fjavascript-typescript-langserver+repo%3Asourcegraph%2Fjava-langserver+repo%3Asourcegraph%2Fdocs.sourcegraph.com+milestone%3A%223.0-preview%22)</small>

---

### 3.0 (2019-01-07)

- [Handle renames and deletions of mirrored repositories](https://github.com/sourcegraph/sourcegraph/issues/914)
- [Align internal deployment processes with customers'](https://github.com/sourcegraph/sourcegraph/issues/976)
- [Direct UI integration and deployment bundling with GitLab](https://github.com/sourcegraph/sourcegraph/issues/1000)
- [Onboarding flow for site admins](https://github.com/sourcegraph/sourcegraph/issues/975)
- Improved code host support for Sourcegraph extensions
- Codecov extension
- Java language support via extension
- [Use nginx as HTTP proxy](https://github.com/sourcegraph/sourcegraph/pull/929)

---

### 3.1 (2019-02-04)

- [Using Sourcegraph extensions in the editor](https://github.com/sourcegraph/sourcegraph/issues/978)
- [Thrift code intelligence](https://github.com/sourcegraph/sourcegraph/issues/669)
- [Extension registry discovery and statistics](https://github.com/sourcegraph/sourcegraph/issues/980)
- Configuration data search extension
- [Swift language support via extension](https://github.com/sourcegraph/sourcegraph/issues/979) (likely includes Objective-C, C, and C++)
- [Cross-language API/IDL support](https://github.com/sourcegraph/sourcegraph/issues/981) (followup from 3.0)
- [Flow (JavaScript) language support](https://github.com/sourcegraph/sourcegraph/issues/982)
- Improved code host support for Sourcegraph extensions (followup from 3.0)
- [Support and documentation for testing Sourcegraph extensions](https://github.com/sourcegraph/sourcegraph/issues/733)
- 3rd-party extensions: LightStep, Sentry, FOSSA, SonarQube, Datadog, LaunchDarkly, Figma
- PHP language support via extension
- Enhanced notification preferences
- [Direct UI integration and deployment bundling with GitLab](https://github.com/sourcegraph/sourcegraph/issues/1000) (followup from 3.0)

---

### 3.2 (2019-03-04)

- [Using Sourcegraph extensions in the editor](https://github.com/sourcegraph/sourcegraph/issues/978) (followup from 3.1)
- [Browser authorization flow for clients](https://github.com/sourcegraph/sourcegraph/pull/528)
- 3rd-party extensions (followup from 3.1)
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

## Previous releases

See [previous Sourcegraph releases](previous_releases.md).

[sourcegraph]: https://github.com/sourcegraph/sourcegraph
[sourcegraph-extension-api]: https://github.com/sourcegraph/sourcegraph/tree/master/packages/sourcegraph-extension-api
[browser-extensions]: https://github.com/sourcegraph/sourcegraph/tree/master/client/browser
[deploy-sourcegraph]: https://github.com/sourcegraph/deploy-sourcegraph
[src-cli]: https://github.com/sourcegraph/src-cli
[chrismwendt]: https://github.com/chrismwendt
[keegancsmith]: https://github.com/keegancsmith
[vanesa]: https://github.com/vanesa
[attfarhan]: https://github.com/attfarhan
[sqs]: https://github.com/sqs
[beyang]: https://github.com/beyany
[ggilmore]: https://github.com/ggilmore
[ryan-blunden]: https://github.com/ryan-blunden
[francisschmaltz]: https://github.com/francisschmaltz
[ijsnow]: https://github.com/ijsnow
[nicksnyder]: https://github.com/nicksnyder
[dadlerj]: https://github.com/dadlerj
[felixfbecker]: https://github.com/felixfbecker
[slimsag]: https://github.com/slimsag
[kattmingming]: https://github.com/kattmingming


<!--

Prior art:

https://about.gitlab.com/direction
https://docs.microsoft.com/en-us/visualstudio/productinfo/vs-roadmap

-->
