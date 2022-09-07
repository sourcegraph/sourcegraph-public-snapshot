# Developing Sourcegraph

<style>
.markdown-body h2 {
  margin-top: 2em;
}
.markdown-body ul {
  list-style:none;
  padding-left: 1em;
}
.markdown-body ul li {
  margin: 0.5em 0;
}
.markdown-body ul li:before {
  content: '';
  display: inline-block;
  height: 1.2em;
  width: 1em;
  background-size: contain;
  background-repeat: no-repeat;
  background-image: url(../batch_changes/file-icon.svg);
  margin-right: 0.5em;
  margin-bottom: -0.29em;
}
body.theme-dark .markdown-body ul li:before {
  filter: invert(50%);
}
</style>

<p class="subtitle">Documentation for <b>developers contributing to the Sourcegraph code base</b></p>

<div class="cta-group">
<a class="btn btn-primary" href="setup/quickstart">★ Quickstart: develop Sourcegraph on your machine</a>
<a class="btn" href="https://github.com/sourcegraph/sourcegraph">GitHub repository</a>
<a class="btn" href="https://github.com/sourcegraph/sourcegraph/issues">Issue Tracker</a>
</div>


## [Setup](setup/index.md)

<p class="subtitle">Learn how to develop Sourcegraph on your machine.</p>

<div class="getting-started">
  <a href="setup/quickstart" class="btn" alt="Run through the Quickstart guide">
   <span>★ Quickstart</span>
   </br>
   Run through the <b>step by step guide</b> and get your local environment ready.
  </a>

  <a href="../dev/how-to" class="btn" alt="How-to guides">
   <span>How-to guides</span>
   </br>
  <b>Context specific</b> guides: debugging live code
  </a>

  <a href="setup/troubleshooting" class="btn" alt="Troubleshooting">
   <span>Troubleshooting</span>
   </br>
  Help for the <b>most common</b> problems.
  </a>
</div>

## [Background information](background-information/index.md)

Clarification and discussion about key concepts, architecture, and development stack.

### Overview

- [Tech stack](background-information/tech_stack.md)
- [Current Sourcegraph tech radar](https://radar.thoughtworks.com/?sheetId=https%3A%2F%2Fraw.githubusercontent.com%2Fsourcegraph%2Fsourcegraph%2Fmain%2Fdoc%2Fdev%2Fradar%2Ftech-radar.csv) (also see [how to maintain the radar](how-to/maintain-tech-radar.md))

### [Architecture](background-information/architecture/index.md)

- [Overview](background-information/architecture/index.md)
- [Introducing a new service](background-information/architecture/introducing_a_new_service.md)

### Development

- [`sg` - the Sourcegraph developer tool](background-information/sg/index.md)
  - [Full `sg` reference](background-information/sg/reference.md)
- [Developing the web clients](background-information/web/index.md)
  - [Developing the web app](background-information/web/web_app.md)
  - [Developing the code host integrations](background-information/web/code_host_integrations.md)
  - [Working with GraphQL](background-information/web/graphql.md)
  - [Wildcard Component Library](background-information/web/wildcard.md)
  - [Styling UI](background-information/web/styling.md)
  - [Accessibility](background-information/web/accessibility.md)
  - [Temporary settings](background-information/web/temporary_settings.md)
  - [Build process](background-information/web/build.md)
- [Developing the GraphQL API](background-information/graphql_api.md)
- [Developing batch changes](background-information/batch_changes/index.md)
- [Developing code navigation](background-information/codeintel/index.md)
- [Developing code insights](background-information/insights/index.md)
- [Developing code monitoring](background-information/codemonitoring/index.md)
- [Developing observability](background-information/observability/index.md)
- [Dependencies and generated code](background-information/dependencies_and_codegen.md)
- [Code reviews](background-information/pull_request_reviews.md)
- [Commit messages](background-information/commit_messages.md)
- [Exposing services](background-information/exposing-services.md)
- [Developing a store](background-information/basestore.md)
- [Developing a worker](background-information/workers.md)
- [Developing an out-of-band migration](background-information/oobmigrations.md)
- [Developing a background routine](background-information/backgroundroutine.md)
- [Code host connections on local dev environment](background-information/code-host.md)
- [Testing](#testing)
  - [Testing principles and guidelines](background-information/testing_principles.md)
  - [Continuous integration](background-information/ci/index.md)
  - [How to write and run tests](how-to/testing.md)
    - [Testing Go code](background-information/languages/testing_go_code.md)
    - [Testing web code](background-information/testing_web_code.md)

### [Languages](background-information/languages/index.md)

- [Go](background-information/languages/go.md)
- [TypeScript](background-information/languages/typescript.md)
- [Bash](background-information/languages/bash.md)
- [Terraform](background-information/languages/terraform.md)

### [SQL](background-information/sql/index.md)

- [Using PostgreSQL](background-information/postgresql.md)
- [Migrations](background-information/sql/migrations.md)
- High-performance guides
  - [Batch operations](background-information/sql/batch_operations.md)
  - [Materialized cache](background-information/sql/materialized_cache.md)

### Security

- [Security Patterns](background-information/security_patterns.md)
- [Security policy](https://about.sourcegraph.com/security/)
- [How to disclose vulnerabilities](https://about.sourcegraph.com/handbook/engineering/security/reporting-vulnerabilities).
- [CSRF security model](security/csrf_security_model.md)

### Tools

- [Renovate dependency updates](background-information/renovate.md)
- [Honeycomb](background-information/honeycomb.md)
- [GoLand](background-information/goland.md)

### Other

- [Telemetry](background-information/telemetry.md)
- [Adding, changing and debugging pings](background-information/adding_ping_data.md)

## Guidelines

- [Code reviews](background-information/pull_request_reviews.md)
- [Open source FAQ](https://about.sourcegraph.com/community/faq)
- [Code of conduct](https://about.sourcegraph.com/community/code_of_conduct)

## [How-to guides](how-to/index.md)

Guides to help with troubleshooting, configuring test instances, debugging, and more.

### New features

- [How to add support for a language](how-to/add_support_for_a_language.md)
- [How to use feature flags](how-to/use_feature_flags.md)

### Observability

- [How to add observability](how-to/add_observability.md)
- [How to add logging](how-to/add_logging.md)
- [How to find monitoring](how-to/find_monitoring.md)
- [How to add monitoring](how-to/add_monitoring.md)
- [Set up local monitoring development](how-to/monitoring_local_dev.md)
- [Set up local OpenTelemetry development](how-to/otel_local_dev.md)

### Documentation

- [Developing the product documentation](how-to/documentation_implementation.md)
- [Architecture Decision Records (ADRs)](adr/index.md)

### Executors

- [How to deploy a new executor image](how-to/deploy_executor_image.md)

### Testing

- [How to write and run tests](how-to/testing.md)
- Testing against code hosts
  - [Configure a test instance of Phabricator and Gitolite](how-to/configure_phabricator_gitolite.md)
  - [Test a Phabricator and Gitolite instance](how-to/test_phabricator.md)
  - [Sync repositories from gitolite.sgdev.org](how-to/sync_repositories_from_gitolite_sgdev_org.md)

## [Contributing](./contributing/index.md)

- [Project setup and CI checks for frontend issues](./contributing/frontend_contribution.md).
- [Accepting an external contribution guide](./contributing/accepting_contribution.md).
