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

- Seamless use of open-source repositories from a self-hosted Sourcegraph instance (in search, navigation, and code host integrations)
- Better JIRA integration
- Support for non-Git version control systems (Perforce, Subversion, TFS, etc.)


### Search

- [Auto-fixup common mistakes in search queries](https://github.com/sourcegraph/sourcegraph/issues/2125)
- [Nested search queries](https://github.com/sourcegraph/sourcegraph/issues/1005) (e.g., in all repositories whose `package.json` contains `foo`, find matches of `bar`)
  - [Multi-line searches](https://github.com/sourcegraph/sourcegraph/issues/35)
- [Enable indexed search by default](https://github.com/sourcegraph/sourcegraph/issues/2176)
- More advanced search filters (provided by extensions), such as those using language-specific or dependency graph information <!-- TODO -->
- [More types/sources of search results](#738) (provided by extensions), such as documentation (wiki, Markdown, and Google Docs), issues, PR comments, logs, and configuration data
- Improved search relevance and ranking <!-- TODO -->

### Code navigation and intelligence

- Continually refine code intelligence (hovers, go-to-definition, find-references, etc.) for all languages
  - Search-based (non-language-server-based) code intelligence for [all languages](https://sourcegraph.com/extensions?query=category%3A%22Programming+languages%22)
  - Language-server-based (precise) code intelligence for [more languages](https://sourcegraph.com/extensions?query=tag%3Alanguage-server)
- [Analyze and expose dependency graph for all major languages and build systems](https://github.com/sourcegraph/sourcegraph/issues/2928)
- [Compute and expose programming language statistics](https://github.com/sourcegraph/sourcegraph/issues/2587)
- [Cross-language, cross-repository definitions and references support for APIs/IDLs (GraphQL, Thrift, Protobuf, etc.)](https://github.com/sourcegraph/sourcegraph/issues/981)

### Integrations

We're refining and adding Sourcegraph integration with [code hosts](#code-hosts), [editors](#editors), and [other tools and services](#other-tools-and-services).

#### Code hosts

Code host integrations have (or will have) the following feature set (in order of priority):

- Repository syncing (project metadata and Git data)
- UI integration with hovers/go-to-definition/find-references/etc. (provided as a native plugin and/or by the [Sourcegraph browser extension](../../integration/browser_extension.md))
- Repository permissions
- User authentication

We are targeting the following code hosts (many of which already support the features above):

- [GitHub integration](https://github.com/sourcegraph/sourcegraph/issues/2915) (GitHub.com and GitHub Enterprise)
- [GitLab integration](https://github.com/sourcegraph/sourcegraph/issues/2916) (GitLab.com and self-hosted GitLab instances)
- [Bitbucket Server integration](https://github.com/sourcegraph/sourcegraph/issues/2917)
- [Phabricator integration](https://github.com/sourcegraph/sourcegraph/issues/2918)
- [AWS CodeCommit integration](https://github.com/sourcegraph/sourcegraph/issues/2919)
- [Gitolite integration](https://github.com/sourcegraph/sourcegraph/issues/2922)
- Future:
  - [Gerrit integration](https://github.com/sourcegraph/sourcegraph/issues/871)
  - [Bitbucket Cloud (bitbucket.org) integration](https://github.com/sourcegraph/sourcegraph/issues/2914)

#### Editors

Editor integrations have (or will have) the following feature set:

- "View file at cursor location on Sourcegraph web interface" action
- "Search code on Sourcegraph" action (for global or repository/directory-scoped searches)
- Configurable single Sourcegraph URL (to support Sourcegraph.com or self-hosted Sourcegraph instance)
- Future:
  - [Support for Sourcegraph extensions](https://github.com/sourcegraph/sourcegraph/issues/978)

We are targeting the following editors (many of which already support the features above):

- [VS Code integration](https://github.com/sourcegraph/sourcegraph/issues/2923)
- [JetBrains IDE integration](https://github.com/sourcegraph/sourcegraph/issues/2926) (IntelliJ, WebStorm, PyCharm, GoLand, etc.)
- [Emacs integration](https://github.com/sourcegraph/sourcegraph/issues/2924)
- [Vim integration](https://github.com/sourcegraph/sourcegraph/issues/2927)
- [Sublime Text integration](https://github.com/sourcegraph/sourcegraph/issues/2929)
- Future:
  - [Eclipse IDE integration](https://github.com/sourcegraph/sourcegraph/issues/2925)

#### Other tools and services

Sourcegraph integrates (or will integrate) with the following popular tools, providing features (dependent on the tool) such as contextual links to/from Sourcegraph and contextual information overlays on code in Sourcegraph.

- [Codecov integration](https://github.com/sourcegraph/sourcegraph/issues/2920)
- [Datadog integration](TODO)
- [LightStep integration](TODO)
- [Sentry integration](TODO)
- [Slack integration](TODO)
- Future: 
  - [LaunchDarkly integration](https://github.com/sourcegraph/sourcegraph/issues/1249)
  - FOSSA
  - SonarQube
  - Figma

### Core services (repositories and authentication)

See the "[Code hosts](#code-hosts)" section above for plans related to repositories, user authentication, and permissions for specific code hosts (such as GitHub).

- [Keep repository set in sync with config](https://github.com/sourcegraph/sourcegraph/issues/2025)
- [Improve process for adding repositories from local disk](https://github.com/sourcegraph/sourcegraph/issues/1527)
- Simpler configuration for HTTPS/SSH credentials for cloning repositories
- [Support internal CA or self-signed TLS certificates for external communication](https://github.com/sourcegraph/sourcegraph/issues/71)

### Extension API, authoring, and registry

This section is only for extension API, authoring and registry improvements for [Sourcegraph extensions](../../extensions/index.md). Features that will be *provided by* extensions are listed in the other sections.

- [Integration testing support for Sourcegraph extensions](https://github.com/sourcegraph/sourcegraph/issues/733)
- [Extension registry discovery and statistics](https://github.com/sourcegraph/sourcegraph/issues/980)
- [Using Sourcegraph extensions in the editor](https://github.com/sourcegraph/sourcegraph/issues/978)

### Deployment, configuration, and management

- Better communication of license status and expiration <!-- TODO -->
- Customization:
  - [Custom branding (logo and brand icon) on the web interface](TODO)
  - [Configurable welcome/homepage message](https://github.com/sourcegraph/sourcegraph/issues/653)
  - Configurable site admin contact info and internal helpdesk link
- Improved flow for common configuration use cases <!-- TODO -->

## Features

Search

- Improvements to saved searches
- Bazel support

Sourcegraph extensions

Other

- [Checklist-based repository reviews](https://github.com/sourcegraph/sourcegraph/issues/1526)
- Enhanced notification preferences

<!--

Prior art:

https://about.gitlab.com/direction
https://docs.microsoft.com/en-us/visualstudio/productinfo/vs-roadmap

-->
