# Sourcegraph roadmap

> NOTE: If you're reading this on a Sourcegraph instance's `/help` area, [view the latest roadmap on docs.sourcegraph.com](https://docs.sourcegraph.com/dev/roadmap).

This roadmap shows what's next for Sourcegraph. The projects and timeframes are subject to change.

A new Sourcegraph release [ships in the first week of each month](https://about.sourcegraph.com/blog). For example, the October 2018 items will ship in the first week of November 2018.

We welcome suggestions! Share feedback by using [code discussions](https://about.sourcegraph.com/blog/discuss-code-and-docs-in-repositories) on this document or the linked feature documents.

## Themes

We want Sourcegraph to be the best way to answer questions while writing, reviewing, or planning code. See the [Sourcegraph master plan](https://about.sourcegraph.com/plan). Our work generally falls into the following categories:

- **Search and browsing:** quickly showing you the code you're looking for and making it easy to navigate around
- **Code intelligence:** go-to-definition, hover tooltips, references, symbols, etc., for code in many languages, including real-time and cross-repository support
- **Integrations:** making Sourcegraph work well with code hosts, review tools, editors, and other tools in your dev workflow (e.g., repository syncing from your code host, browser extensions, and editor extensions)
- **Extensibility:** supporting Sourcegraph extensions that add code intelligence and other information (e.g., tracing, logging, and security annotations from 3rd-party tools) to Sourcegraph and external tools that Sourcegraph integrates with
- **Deployment:** making it easy to run and maintain a self-hosted Sourcegraph instance
- **Enterprise:** features that larger companies need (e.g., scaling, authentication, authorization, auditing, etc.)

## Key

🌞 = pull request or code<br>
📣 = draft blog post<br>
🐞 = issues<br>
📖 = draft docs<br>
📽 = demo or screencast video<br>
💡 = high-level sketch<br>
🚢 = shipped and ready to use<br>

<style>
tr td:nth-child(2) {
width: 0;
white-space: nowrap;
padding: 6px;
}
tr td:nth-child(3), tr td:nth-child(4) {
width: 20%;
font-size: 80%;
}
</style>

## October 2018

Feature | 🔗 | Themes/products | Team
------- | --- | -------- | ----
Language support via Sourcegraph extensions | [🐍📽](https://slack-files.com/T02FSM7DL-FDVNVV83G-1af26916ad)[📣](https://sourcegraph.com/github.com/sourcegraph/about/-/blob/projects/simplified-language-support.md) | Code intelligence, extensibility, [Sourcegraph][sourcegraph], [browser extension][browser-extensions], [extension API][sourcegraph-extension-api] | [@chrismwendt][chrismwendt] [@felixfbecker][felixfbecker]
Sub-query support in search | [📣](https://github.com/sourcegraph/about/pull/8)[💡](https://sourcegraph.sgdev.org/github.com/sourcegraph/docs-private/-/blob/201809/hierarchical-search-2.md) | Search, [Sourcegraph][sourcegraph] | [@keegancsmith][keegancsmith]
Indexed search enabled via config | 🚢[🌞](https://github.com/sourcegraph/sourcegraph/pull/502)[📣](https://github.com/sourcegraph/sourcegraph/pull/502) | Search, [Sourcegraph][sourcegraph] | [@keegancsmith][keegancsmith]
Search query transformations via Sourcegraph extensions | [🌞](https://github.com/sourcegraph/sourcegraph/pull/499)[📣](https://github.com/sourcegraph/about/pull/40)[📽️](https://cl.ly/5159339a6016) | Search, extensibility, [Sourcegraph][sourcegraph], [extension API][sourcegraph-extension-api] | [@attfarhan][attfarhan]
Primary workflow UX improvements | [📣](https://github.com/sourcegraph/about/pull/39)[🐞](https://github.com/sourcegraph/sourcegraph/issues?q=is%3Aopen+is%3Aissue+assignee%3Avanesa+milestone%3A%22October+2018%22) | Search and browsing, integrations, [Sourcegraph][sourcegraph] | [@vanesa][vanesa] [@francisschmaltz][francisschmaltz]
Pure Docker cluster deployment | [🌞](https://github.com/sourcegraph/deploy-sourcegraph-docker)[📖](https://github.com/sourcegraph/deploy-sourcegraph-docker#readme)[📣](https://github.com/sourcegraph/about/pull/37) | Deployment, [deploy-sourcegraph-docker](https://github.com/sourcegraph/deploy-sourcegraph-docker) | [@slimsag][slimsag]
Repository permissions | [🌞](https://github.com/sourcegraph/sourcegraph/pull/557)[📣](https://github.com/sourcegraph/about/pull/47)[📖](https://sourcegraph.com/github.com/sourcegraph/about/-/blob/projects/acls.md) | Enterprise, integrations, [Sourcegraph][sourcegraph] | [@beyang][beyang]
Explore page | [🚢](https://sourcegraph.com/explore)[📣](https://github.com/sourcegraph/about/pull/51) | [Sourcegraph][sourcegraph] | [@francisschmaltz][francisschmaltz] [@sqs][sqs]
Product documentation | [📣](https://github.com/sourcegraph/about/pull/43)[📖](https://docs.sourcegraph.com/dev/documentation)[📖](https://github.com/sourcegraph/docs.sourcegraph.com#readme) | All, [Sourcegraph][sourcegraph] | [@sqs][sqs]
Sourcegraph extensions usage and authoring experience | [📽](https://drive.google.com/file/d/1lguzuXbKYuSFwIvM7KK6FW8p6jMibxGF/view)[📖](https://docs.sourcegraph.com/extensions) | Extensibility, [Sourcegraph][sourcegraph], [extension API][sourcegraph-extension-api] | [@slimsag][slimsag] [@ryan-blunden][ryan-blunden]

<small>[📣 2.13 announcement](https://about.sourcegraph.com/blog/announcing-sourcegraph-2.13) (week of 5 November 2018) and [📣 3.0-preview announcement](https://github.com/sourcegraph/about/pull/49) (week of 19 November 2018) --- [All October 2018 issues](https://github.com/issues?utf8=%E2%9C%93&q=is%3Aissue+is%3Aopen+author%3Asqs+archived%3Afalse+sort%3Aupdated-desc+repo%3Asourcegraph%2Fsourcegraph-extension-api+repo%3Asourcegraph%2Fsourcegraph+repo%3Asourcegraph%2Fenterprise+repo%3Asourcegraph%2Fsourcegraph-extension-api+repo%3Asourcegraph%2Fbrowser-extensions+repo%3Asourcegraph%2Fextensions-client-common+repo%3Asourcegraph%2Fsrc-cli+repo%3Asourcegraph%2Fcodeintellify+repo%3Asourcegraph%2Fgo-langserver+repo%3Asourcegraph%2Fjavascript-typescript-langserver+repo%3Asourcegraph%2Fjava-langserver+repo%3Asourcegraph%2Fdocs.sourcegraph.com+milestone%3A%22October+2018%22)</small>

---

## November 2018

> NOTE: *Tentative.* Not all features have their blog posts and docs linked yet.

Feature | 🔗 | Themes/products | Team
------- | --- | -------- | ----
More robust code host repository syncing | | Integrations, [Sourcegraph][sourcegraph] | [@nicksnyder][nicksnyder]
Unified site config editing and management console | [🌞](https://github.com/sourcegraph/sourcegraph/pull/498)[📣](https://github.com/sourcegraph/about/pull/36) | Deployment, [Sourcegraph][sourcegraph], [deploy-sourcegraph][deploy-sourcegraph] | [@ggilmore][ggilmore]
Onboarding flow for code host integrations | [📣](https://github.com/sourcegraph/about/pull/38) | Integrations, [Sourcegraph][sourcegraph], [browser extension][browser-extensions] | [@francisschmaltz][francisschmaltz] and T.B.D.
GitHub issue search | [💡](https://docs.google.com/document/d/1OTXPlVxSDNC37hlEVnNmtO1s-doA6O3S1210UWl55tY/edit) [📣](https://github.com/sourcegraph/about/pull/53) [📽](https://sourcegraph.slack.com/archives/C89KCDK5J/p1541753225044700) | Search, extensibility, [Sourcegraph][sourcegraph], [extension API][sourcegraph-extension-api] | [@vanesa][vanesa] [@attfarhan][attfarhan] [@keegancsmith][keegancsmith] [@francisschmaltz][francisschmaltz]
JavaScript/TypeScript extension | | Code intelligence, sourcegraph-typescript | [@felixfbecker][felixfbecker]
LDAP and Active Directory user authentication | | Enterprise, [Sourcegraph][sourcegraph] | [@beyang][beyang]
Go extension | [💡](https://docs.google.com/document/d/1j6X6Flw9_GT0QsCv1XVD1_zx0VjWFcD8pvl5uhzazMU/edit) | Code intelligence, sourcegraph-go | [@chrismwendt][chrismwendt]
Simpler browser extension options menu | [🌞](https://github.com/sourcegraph/browser-extensions/pull/271)[📣](https://github.com/sourcegraph/about/pull/46) | Integrations, [browser extension][browser-extensions] | [@ijsnow][ijsnow] [@francisschmaltz][francisschmaltz]
Python extension | [📽](https://slack-files.com/T02FSM7DL-FDXV2DM3J-ecc49122bd) | Code intelligence, sourcegraph-python | [@sqs][sqs]
Codecov and other dev tool extensions | | Integrations, extensibility | T.B.D.

<small>Release: week of 3 December 2018 --- [All November 2018 issues](https://github.com/issues?utf8=%E2%9C%93&q=is%3Aissue+is%3Aopen+author%3Asqs+archived%3Afalse+sort%3Aupdated-desc+repo%3Asourcegraph%2Fsourcegraph-extension-api+repo%3Asourcegraph%2Fsourcegraph+repo%3Asourcegraph%2Fenterprise+repo%3Asourcegraph%2Fsourcegraph-extension-api+repo%3Asourcegraph%2Fbrowser-extensions+repo%3Asourcegraph%2Fextensions-client-common+repo%3Asourcegraph%2Fsrc-cli+repo%3Asourcegraph%2Fcodeintellify+repo%3Asourcegraph%2Fgo-langserver+repo%3Asourcegraph%2Fjavascript-typescript-langserver+repo%3Asourcegraph%2Fjava-langserver+repo%3Asourcegraph%2Fdocs.sourcegraph.com+milestone%3A%22November+2018%22)</small>

---

## December 2018

> NOTE: *Tentative.* Not all features have their blog posts and docs linked yet.

Feature | 🔗 | Themes/products | Team
------- | --- | -------- | ----
Using Sourcegraph extensions in the editor | [📣](https://docs.google.com/document/d/1_NTon70WY6uHzogGPBG06FRatNCVrKvSbHbZUEKY9xM/edit) | Integrations, extensibility, [Sourcegraph][sourcegraph], [extension API][sourcegraph-extension-api] | [@slimsag][slimsag]
Extension registry discovery and statistics | [📣](https://github.com/sourcegraph/docs-private/blob/master/201809/tentative/social-cxp-registry.md) | Extensibility, [Sourcegraph][sourcegraph] | [@slimsag][slimsag] [@vanesa][vanesa] [@francisschmaltz][francisschmaltz]
[Direct UI integration and deployment bundling with GitLab](https://github.com/sourcegraph/about/pull/41) | | Integrations, [Sourcegraph][sourcegraph], [browser extension][browser-extensions] | [@ggilmore][ggilmore] [@ijsnow][ijsnow] [@francisschmaltz][francisschmaltz]
Doc site integrations | [💡](https://sourcegraph.sgdev.org/github.com/sourcegraph/docs-private/-/blob/201808/docs-code-intel.md) | Integrations, [Sourcegraph][sourcegraph] | [@vanesa][vanesa] [@ijsnow][ijsnow]
Browser authorization flow for clients | [🌞](https://github.com/sourcegraph/sourcegraph/pull/528)[🐞](https://github.com/sourcegraph/src-cli/issues/28) [📖](https://github.com/sourcegraph/about/pull/42) | Integrations, [Sourcegraph][sourcegraph], [`src`][src-cli] | [@sqs][sqs]
Swift language support | | Code intelligence, sourcegraph-swift | | T.B.D. ([@nicksnyder][nicksnyder] or [@chrismwendt][chrismwendt]?)
Cross-language API/IDL support ([GraphQL](https://sourcegraph.com/github.com/sourcegraph/about/-/blob/projects/graphql-sourcegraph-extension.md), Thrift, Protobuf) | | Code intelligence, sourcegraph-{graphql,thrift,protobuf} | T.B.D.
Ruby language support | | Code intelligence, sourcegraph-ruby | T.B.D.
Flow (JS) language support | | Code intelligence, sourcegraph-flow | T.B.D.
Rust language support *(tentative)* | | Code intelligence, sourcegraph-go | [@slimsag][slimsag]

<small>Release: week of 7 January 2019 --- [All December 2018 issues](https://github.com/issues?utf8=%E2%9C%93&q=is%3Aissue+is%3Aopen+author%3Asqs+archived%3Afalse+sort%3Aupdated-desc+repo%3Asourcegraph%2Fsourcegraph-extension-api+repo%3Asourcegraph%2Fsourcegraph+repo%3Asourcegraph%2Fenterprise+repo%3Asourcegraph%2Fsourcegraph-extension-api+repo%3Asourcegraph%2Fbrowser-extensions+repo%3Asourcegraph%2Fextensions-client-common+repo%3Asourcegraph%2Fsrc-cli+repo%3Asourcegraph%2Fcodeintellify+repo%3Asourcegraph%2Fgo-langserver+repo%3Asourcegraph%2Fjavascript-typescript-langserver+repo%3Asourcegraph%2Fjava-langserver+repo%3Asourcegraph%2Fdocs.sourcegraph.com+milestone%3A%22November+2018%22)</small>

---

## January 2019

<small>Release: week of 4 February 2019</small>

<!-- TODO: Standardized code host UI integration points for Sourcegraph extensions | | Integrations, [Sourcegraph][sourcegraph], [extension API][sourcegraph-extension-api] [browser extension][browser-extensions] | [@francisschmaltz][francisschmaltz] [@ijsnow][ijsnow] -->

---

## February 2019

<small>Release: week of 4 March 2019</small>

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

https://docs.microsoft.com/en-us/visualstudio/productinfo/vs-roadmap

-->
