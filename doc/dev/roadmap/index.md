# Sourcegraph roadmap

We want Sourcegraph to be:

- **For developers:** the best way to answer questions and get unblocked while writing, reviewing, or reading code.
- **For organizations** (engineering leaders and internal tools teams): the infrastructure for developer tools and data.

This roadmap shows what's planned for the next 12 months. See the [Sourcegraph master plan](https://about.sourcegraph.com/plan) for our high-level product vision.

**Next release:** [Sourcegraph 3.3 release plan](https://github.com/sourcegraph/sourcegraph/issues/2931) (ships on April 20, 2019).

We ship a release on the [20th day of each month](../releases.md#releases-are-monthly). See [previous Sourcegraph releases](previous_releases.md).

---

## Overview

We're continually improving Sourcegraph's core features for developers:

- Code search and navigation (with code intelligence)
- Integration into code review
- Robust API for automation and reporting on your code
- Code change notifications (saved searches)

We're also working toward making Sourcegraph the [**infrastructure for developer tools and data**](#future-infrastructure-for-developer-tools-and-data), with integration into the entire developer workflow (in the editor, code review, and anywhere else they interact with code). In addition to helping developers, this also helps engineering leaders and internal tools teams solve organization-wide problems.

## Search

<!-- You'll be able to refine your search to find what you need more quickly, and search across more types of things (beyond code). -->

We're making search faster, more accurate, and more comprehensive (so it includes issues, documents, and other data sources you might want to search).

- [Auto-fixup common mistakes in search queries](https://github.com/sourcegraph/sourcegraph/issues/2125)
- [Nested search queries](https://github.com/sourcegraph/sourcegraph/issues/1005) (e.g., in all repositories whose `package.json` contains `foo`, find matches of `bar`)
  - [Multi-line searches](https://github.com/sourcegraph/sourcegraph/issues/35)
- [Enable indexed search by default](https://github.com/sourcegraph/sourcegraph/issues/2176)
- More advanced search filters (provided by extensions), such as those using language-specific or dependency graph information <!-- TODO -->
- [More types/sources of search results](https://github.com/sourcegraph/sourcegraph/issues/738) (provided by extensions), such as documentation (wiki, Markdown, and Google Docs), issues, PR comments, logs, and configuration data
- Investigate instant, as-you-type search (Livegrep-style)
- Improved search relevance and ranking <!-- TODO -->

## Saved searches

TODO

- Enhanced notification preferences
- TODO: Start with better search results and saved searches interface

## Code navigation and intelligence

We're continually refining code intelligence (hovers, go-to-definition, find-references, etc.) for all languages and making our integration into code hosts (such as GitHub) more robust. We're also working toward features that let you navigate the dependency graph of imports, libraries, and services.

- Search-based (non-language-server-based) code intelligence for [all languages](https://sourcegraph.com/extensions?query=category%3A%22Programming+languages%22)
- Language-server-based (precise) code intelligence for [more languages](https://sourcegraph.com/extensions?query=tag%3Alanguage-server)
- Continually ensure code navigation/intelligence works on code hosts using our [browser extension](../../integration/browser_extension.md) and [native code host integrations](#code-hosts).
- [Analyze and expose dependency graph for all major languages and build systems](https://github.com/sourcegraph/sourcegraph/issues/2928)
- [Compute and expose programming language statistics](https://github.com/sourcegraph/sourcegraph/issues/2587)
- Show panel (with references/etc.) UI in code host integrations
- Allow extensions to handle diffs and pull requests as a first-class concern
- [Cross-language, cross-repository definitions and references support for APIs/IDLs (GraphQL, Thrift, Protobuf, etc.)](https://github.com/sourcegraph/sourcegraph/issues/981)

## Integrations

We're refining and adding Sourcegraph integration with [code hosts](#code-hosts), [editors](#editors), and [other tools and services](#other-tools-and-services), so that Sourcegraph searches across all your code and gives you contextual information from your favorite developer tools in your workflow.

### Code hosts

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

### Editors

We want to make it super fast to get the answer you need on Sourcegraph when you're in your editor, without switching to your browser and losing focus. We're focused on solving problems for you that local editor search and navigation can't answer. Editor integrations have (or will have) the following feature set:

- "View file at cursor location on Sourcegraph web interface" action
- "Search code on Sourcegraph" action (for global searches or searches scoped to the current repository and its transitive dependencies/dependents)
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

### Other tools and services

Sourcegraph integrations enhance your developer workflow, giving you vital information you need while coding. These integrations add features such as contextual links to/from Sourcegraph and contextual information overlays on code in Sourcegraph.

- [Codecov integration](https://github.com/sourcegraph/sourcegraph/issues/2920) (code coverage)
- [Datadog integration](TODO) (tracing and performance monitoring)
- [LightStep integration](TODO) (tracing and performance monitoring)
- [Sentry integration](TODO) (error monitoring)
- [Slack integration](TODO) (team chat)
- [G Suite integration](TODO) (Google domain management)
- Future:
  - [JIRA integration](https://github.com/sourcegraph/sourcegraph/issues/2930) (project planning and issue tracking)
  - [LaunchDarkly integration](https://github.com/sourcegraph/sourcegraph/issues/1249) (feature flags)
  - FOSSA integration (license compliance)
  - SonarQube integration (static analysis)

## Core UX

<!-- TODO: we have no owner for this stuff right now -->

- Speed up page loads and reduce UI jitter
- Improve keyboard navigation and keyboard shortcuts
- Investigate more hypertext-like, less app-like UI
- Improve accessibility
- Address major pain points for mobile and tablet users

## Core services (repositories and authentication)

See the "[Code hosts](#code-hosts)" section above for plans related to repositories, user authentication, and permissions for specific code hosts (such as GitHub).

- [Keep repository set in sync with config](https://github.com/sourcegraph/sourcegraph/issues/2025)
- [Improve process for adding repositories from local disk](https://github.com/sourcegraph/sourcegraph/issues/1527)
- Simpler configuration for HTTPS/SSH credentials for cloning repositories
- [Support internal CA or self-signed TLS certificates for external communication](https://github.com/sourcegraph/sourcegraph/issues/71)
- Support for non-Git version control systems (Perforce, Subversion, TFS, etc.)

## Extension API, authoring, and registry

This section is only for extension API, authoring and registry improvements for [Sourcegraph extensions](../../extensions/index.md). Features that will be *provided by* extensions are listed in the other sections.

- [Integration testing support for Sourcegraph extensions](https://github.com/sourcegraph/sourcegraph/issues/733)
- [Extension registry discovery and statistics](https://github.com/sourcegraph/sourcegraph/issues/980)
- [Using Sourcegraph extensions in the editor](https://github.com/sourcegraph/sourcegraph/issues/978)

## Deployment, configuration, and management

We want to make it easy to set up a self-hosted Sourcegraph instance in minutes, locally or on the most popular cloud providers. It needs to scale to the needs of organizations with thousands of developers, tens of thousands of repositories, and complex cluster and security needs.

- Better communication of license status and expiration <!-- TODO -->
- Customization:
  - [Custom branding (logo and brand icon) on the web interface](TODO)
  - [Configurable welcome/homepage message](https://github.com/sourcegraph/sourcegraph/issues/653)
  - Configurable site admin contact info and internal helpdesk link
- Improved flow for common configuration use cases (e.g., "just make everything work well with my GitHub.com organization")

## Other

- [Seamless use of open-source repositories from a self-hosted Sourcegraph instance](https://github.com/sourcegraph/sourcegraph/issues/2954)

---

## Infrastructure for developer tools and data

> NOTE: This section describes future capabilities that we're working toward.

We want Sourcegraph to be the infrastructure for developer tools and data inside your organization, so you can:

- [Build and adopt new developer tools organization-wide](#build-and-adopt-developer-tools-more-easily) more easily, with seamless integration into the editor, code review, and anywhere else developers interact with code
- [Provide consistent, remote-capable development environments to developers](#consistent-remote-capable-development-environments)
- [Perform automated refactors programmatically across repositories and languages](#automated-refactoring)
- [Make data-driven decisions about development processes](#data-driven-development-insights-and-reporting)
- [Enforce rules around security, compliance, and licensing in a developer-friendly way](#security-compliance-and-licensing)

By "developer tools and data", we mean things like:

- Repositories and Git data
- Git history
- Code ownership and review
- Build process and environment
- Test execution, status, and coverage
- Dependency graph (for imports/libraries and runtime services)
- Static code analysis and metrics (such as code churn)
- Cross-references
- Deployment (where is this code running?)
- Runtime monitoring and logging

Your existing tools to handle these things would integrate with Sourcegraph; no need to switch.

### Build and adopt developer tools more easily

We will let you build and roll out developer tools to every developer in your organization, with seamless integration into the editor, code review, and anywhere else developers interact with code.

- You can build new tools that analyze code and automate processes without cloning, building, and executing code in each repository, and without needing to integrate one-by-one with all of your other tools and data. They can access all of your developer tools and data in one central place on your Sourcegraph instance, and they can perform actions (such as creating diffs/PRs or sending notifications to code owners) via Sourcegraph.
- A [Sourcegraph extension](../../extensions/index.md) (which provides integration with the developer tools and data listed above) can be rolled out to every editor, code review, and code host used by anyone at your organization.

> Example: You want to encourage better testing among developers in your organization, so you've started measuring test coverage in CI. With Sourcegraph, you'd also be able to roll out (optional) test coverage visual overlays to all developers in their editor and code review tool. This would make test coverage more useful and accessible to developers while coding and reviewing, which will yield much faster improvements to your testing culture.

> Example: You want to automatically ingest crash logs from your application and create issues from them. To do so, you'd build a tool that parses the stack traces and uses the Sourcegraph API to gather the information to include in the issue: assignment to the code owner of the most recently changed method in the stack trace, a link to the associated code review, logs and traces related to the stack trace, and a list of other services that depend on this code (to help triage). Without the infrastructure Sourcegraph will provide, this kind of automation would be extremely hard to build and fragile.

### Consistent, remote-capable development environments

We will let you connect your local editor to your organization's Sourcegraph instance and immediately get code intelligence, builds, tests, and other development tools on all of your code.

- The development environment would be centrally configured from your Sourcegraph instance, so every developer would have a consistent environment (subject to user customizations).
- Computationally intensive tasks (such as builds, code intelligence, and tests) could be offloaded to the remote server. (Some will be harder to make remote than others, and we expect to make gradual progress here.)
- You could also choose to launch a web-based cloud IDE session with a fully configured developer environment.

### Automated refactoring

We will let you perform safe, large-scale refactors of code across repositories, services, and languages.

- For simple large-scale edits that can be merged independently, you'll be able to preview and create multiple linked diffs/PRs for the edit, each assigned to the right code reviewer. You can track progress of merging all of the individual edits and monitor new candidates for the edit to be applied (such as when a developer merges code after you created the initial batch of diffs/PRs).
- For refactors that need coordinated or staged deployment, you'll be able to define templates that codify how to make the change in multiple steps. For example, to change the signature of a service method, a template might codify a multi-step process where first the implementation changes to handle both the old and new arguments, then all callers are updated, and finally (1 month after all calling services are deployed) support for the old arguments is removed. This process relies on Sourcegraph's knowledge of the dependency graph, code generation steps, and deployment status.

### Data-driven development insights and reporting

We will let you use data derived from development to make better decisions. You will be able to export this data to your preferred reporting dashboard.

- Built-in reporting of usage trends for programming languages, dependencies, and services
- Advanced reporting using information supplied by the developer tools you integrate into Sourcegraph (such as test coverage, code churn, etc.)

### Security, compliance, and licensing

We will let you enforce policies and processes around security, compliance, and licensing in a developer-friendly way.

- Existing tools for security and license analysis will provide the raw list of possible problems.
- Sourcegraph will provide the workflow for triage, reporting, and fixing the issues (and ensuring they don't reoccur).
  - Triage and reporting will use Sourcegraph's understanding of code ownership and review (for assignment) and the dependency graph (for prioritization, such as highlighting when a security issue is found in code used by a user-facing service).
  - Fixing the issue and preventing regressions will rely on [automated refactoring](#automated-refactoring).

> Example: For companies needing to be PCI compliant, [PCI DSS 3.0 rule 6.3.2](https://pcinetwork.org/forum/index.php?threads/pci-dss-3-0-6-3-2-review-custom-code-prior-to-release-to-production-or-customers-in-order-to-identify-any-potent.645/) requires someone other than the author to review each code change. Using Sourcegraph's knowledge of the dependency graph for libraries and services, you'll be able to guarantee that all code used by a PCI-compliant system has been properly reviewed (prior to merging and before release), even if it's in a separate repository or tree.

<!--

Prior art:

https://about.gitlab.com/direction
https://docs.microsoft.com/en-us/visualstudio/productinfo/vs-roadmap

-->
