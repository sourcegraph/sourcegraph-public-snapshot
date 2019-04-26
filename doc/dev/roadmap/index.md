# Sourcegraph roadmap

We want Sourcegraph to be:

- **For developers:** the best way to answer questions and get unblocked while writing, reviewing, or reading code.
- **For organizations** (engineering leaders and internal tools teams): the infrastructure for developer tools and data.

This roadmap is a curated list of what we are working on now and the direction that we want to move in over the next 12 months. See the [Sourcegraph master plan](https://about.sourcegraph.com/plan) for our high-level product vision.

We ship a release on the [20th day of each month](../releases.md#releases-are-monthly). See [previous Sourcegraph releases](previous_releases.md).

**Next release:** [Sourcegraph 3.4](https://github.com/sourcegraph/sourcegraph/issues?q=is%3Aopen+is%3Aissue+milestone%3A3.4+label%3Aroadmap) (ships on May 20, 2019).

## Overview

We're continually improving Sourcegraph's core features for developers:

- Code search and navigation (with code intelligence)
- Integration into code review
- Robust API for automation and reporting on your code
- Code change notifications (saved searches)

We're also working toward making Sourcegraph the [**infrastructure for developer tools and data**](#infrastructure-for-developer-tools-and-data), with integration into the entire developer workflow (in the editor, code review, and anywhere else they interact with code). In addition to helping developers, this also helps engineering leaders and internal tools teams solve organization-wide problems.

| Icon | Description                                                  |
| ---- | ------------------------------------------------------------ |
| 🏃   | Features we are actively working on.                         |
| 🙋   | We are hiring for all teams, but especially for these roles. |

## Search

Owners: @ijt, @ijsnow

We're making search faster, more powerful, and more comprehensive (so it includes issues, documents, and other data sources you might want to search).

- 🏃 [Enable indexed search by default](https://github.com/sourcegraph/sourcegraph/issues/2176)
- 🏃 [Nested search queries](https://github.com/sourcegraph/sourcegraph/issues/1005) (e.g., in all repositories whose `package.json` contains `foo`, find matches of `bar`)
  - [Multi-line searches](https://github.com/sourcegraph/sourcegraph/issues/35)
- Streaming search results to reduce time to first result.
- More ways to filter queries (provided by extensions), such as by authorship, recency, and language-specific or dependency graph information
- [More types/sources of search results](https://github.com/sourcegraph/sourcegraph/issues/738) (provided by extensions), such as documentation (wiki, Markdown, and Google Docs), issues, PR comments, logs, and configuration data
- Investigate instant, as-you-type search (Livegrep-style)
- Improved search relevance and ranking

## Saved searches

Owners: @attfarhan

We want to help you stay on top of the code changes you care about, and are working to improve saved searches with better notifications, better integration with code reviews and PRs, and an easier way to add new saved searches and triage results.

- 🏃 Improve UI for adding and editing saved searches
- 🏃 More configurable notification destinations per-saved search
- Add GitHub/GitLab PR commit status integration for saved searches
- Show read/unread saved search results
- Allow creating issues (GitHub/GitLab/Jira/etc.) from saved search results

## Code navigation and intelligence

Owners: @lguychard, @chrismwendt, @felixfbecker, @vanesa

Sourcegraph helps developers navigate code by providing code intelligence when browsing code on Sourcegraph or on supported code hosts (like GitHub) via our [browser extension](../../integration/browser_extension.md). Code intelligence includes hover tooltips, go-to-definition, find-references, and other capabilities contributed by [Sourcegraph extensions](../../extensions/index.md) like [Codecov](https://sourcegraph.com/extensions?query=Codecov), [Sentry](https://sourcegraph.com/extensions?query=Sentry), [LightStep](https://sourcegraph.com/extensions?query=LightStep), and [Datadog](https://sourcegraph.com/extensions?query=Datadog).

- 🏃 [Increase the reliability and performance of the browser extension on GitHub and other code hosts](https://github.com/sourcegraph/sourcegraph/issues/3485)
- 🏃 [Improvements for Datadog integration](https://github.com/sourcegraph/sourcegraph/issues/3297) (tracing and performance monitoring)
- [Improve LightStep integration](https://github.com/sourcegraph/sourcegraph/issues/3304) (tracing and performance monitoring)
- [Improve Codecov integration](https://github.com/sourcegraph/sourcegraph/issues/2920) (code coverage)
- [Improve Sentry integration](https://github.com/sourcegraph/sourcegraph/issues/3305) (error monitoring)
- [Analyze and expose dependency graph for all major languages and build systems](https://github.com/sourcegraph/sourcegraph/issues/2928)
- [Show panel (with references/etc.) UI in code host integrations](https://github.com/sourcegraph/sourcegraph/issues/3089)
- Allow extensions to handle diffs and pull requests as a first-class concern
- [Bazel support roadmap](https://github.com/sourcegraph/sourcegraph/issues/2982)
- [Cross-language, cross-repository definitions and references support for APIs/IDLs (GraphQL, Thrift, Protobuf, etc.)](https://github.com/sourcegraph/sourcegraph/issues/981)
- [Add Slack integration](https://github.com/sourcegraph/sourcegraph/issues/2986) (team chat)
- [Add G Suite integration](https://github.com/sourcegraph/sourcegraph/issues/2987) (Google domain management)
- [Add Jira integration](https://github.com/sourcegraph/sourcegraph/issues/2930) (project planning and issue tracking)
- [Add Bazel integration](https://github.com/sourcegraph/sourcegraph/issues/2982) (builds)
- [Add LaunchDarkly integration](https://github.com/sourcegraph/sourcegraph/issues/1249) (feature flags)
- [Add FOSSA integration](https://github.com/sourcegraph/sourcegraph/issues/2988) (license compliance)
- [Add SonarQube integration](https://github.com/sourcegraph/sourcegraph/issues/2989) (static analysis)

## Core services (repositories and authentication)

Owners: @keegancsmith, @tsenart, @mrnugget

- 🏃 [Robust synchronization behavior for all supported code hosts that handles repository renames and deletions](https://github.com/sourcegraph/sourcegraph/issues/3467)
- 🏃 [Mapping local repositories in your editor to Sourcegraph](https://github.com/sourcegraph/sourcegraph/issues/462)
- [Authentication and authorization support for Bitbucket Server](https://github.com/sourcegraph/sourcegraph/issues/1108)
- [Compute and expose programming language statistics](https://github.com/sourcegraph/sourcegraph/issues/2587)
- [Improve process for adding repositories from local disk](https://github.com/sourcegraph/sourcegraph/issues/1527)
- Simpler configuration for HTTPS/SSH credentials for cloning repositories
- [Support internal CA or self-signed TLS certificates for external communication](https://github.com/sourcegraph/sourcegraph/issues/71)
- Support for non-Git version control systems (Perforce, Subversion, TFS, etc.)

## Code modification

Owners: @rvantonder

We will let you perform safe, large-scale refactors of code across repositories, services, and languages.

- 🏃 [Find and replace string literals across multiple GitHub repositories](https://github.com/sourcegraph/sourcegraph/issues/3483)
- Find and replace using sophisticated pattern matching and replacement templates (e.g. regex or comby).
- Add support for modifing code on more code hosts (e.g. Bitbucket Server, GitLab).
- Create a UI for viewing and managing the state of in progress refactors (e.g. how many PRs are merged, automatically opening new PRs to refactor matching code that just got committed).

## Deployment, configuration, and management

Owners: @beyang, @slimsag, @ggilmore

We want to make it easy to set up a self-hosted Sourcegraph instance in minutes, locally or on the most popular cloud providers. It needs to scale to the needs of organizations with thousands of developers, tens of thousands of repositories, and complex cluster and security needs.

- Sourcegraph managed instances
- Default request tracing for private instances
- Easier deployment and configuration of language servers
- Customization:
  - [Configurable welcome page](https://github.com/sourcegraph/sourcegraph/issues/2443)
  - [Configurable site admin contact info and internal helpdesk link](https://github.com/sourcegraph/sourcegraph/issues/2442)
- Improved flow for common configuration use cases (e.g., "just make everything work well with my GitHub.com organization")

## Core UX

Owners: [We're hiring](https://github.com/sourcegraph/careers/blob/master/job-descriptions/software-engineer.md)! 🙋

- Speed up page loads and reduce UI jitter
- Improve keyboard navigation and keyboard shortcuts
- Investigate more hypertext-like, less app-like UI
- Improve accessibility
- Address major pain points for mobile and tablet users

## Extension API, authoring, and registry

Owners: [We're hiring](https://github.com/sourcegraph/careers/blob/master/job-descriptions/software-engineer.md)! 🙋

The Sourcegraph extension API allows developers to enhance their code review workflow with custom data. We're working to make the extension API and extension registry even more powerful and useful.

- [Integration testing support for Sourcegraph extensions](https://github.com/sourcegraph/sourcegraph/issues/733)
- [Extension registry discovery and statistics](https://github.com/sourcegraph/sourcegraph/issues/980)
- [Using Sourcegraph extensions in the editor](https://github.com/sourcegraph/sourcegraph/issues/978)

## Editors

[Sourcegraph integrates with many editors](https://docs.sourcegraph.com/integration/editor). We want to make it super fast to get the answer you need on Sourcegraph when you're in your editor, without switching to your browser and losing focus. We're focused on solving problems for you that local editor search and navigation can't answer.

- [Add support for Sourcegraph extensions to existing editor integrations](https://github.com/sourcegraph/sourcegraph/issues/978).

---

## Infrastructure for developer tools and data

> NOTE: This section describes future capabilities that we're working toward.

We want Sourcegraph to be the infrastructure for developer tools and data inside your organization, so you can:

- [Build and adopt new developer tools organization-wide](#build-and-adopt-developer-tools-more-easily) more easily, with seamless integration into the editor, code review, and anywhere else developers interact with code
- [Provide consistent, remote-capable development environments to developers](#consistent-remote-capable-development-environments)
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
https://github.com/Microsoft/vscode/wiki/Roadmap

-->
