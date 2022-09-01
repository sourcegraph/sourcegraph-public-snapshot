<p align="center">
  <a href="https://sourcegraph.com" target="_blank">
    <img src="https://p21.p4.n0.cdn.getcloudapp.com/items/YEuWmEJA/38872827-37f4-4d2f-992d-c6870d794f57.svg" alt="Sourcegraph 4.0" width="300px">
  </a>

<p align="center">
    <a href="https://docs.sourcegraph.com">Docs</a> •
    <a href="https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/CONTRIBUTING.md">Contributing</a> •
    <a href="https://twitter.com/sourcegraph">Twitter</a>
    <br /><br />
    <a href="https://buildkite.com/sourcegraph/sourcegraph">
        <img src="https://badge.buildkite.com/00bbe6fa9986c78b8e8591cffeb0b0f2e8c4bb610d7e339ff6.svg?branch=main" alt="Build status" />
    </a>
    <a href="https://api.securityscorecards.dev/projects/github.com/sourcegraph/sourcegraph">
        <img src="https://img.shields.io/ossf-scorecard/github.com/sourcegraph/sourcegraph?label=openssf%20scorecard" alt="Scorecard" />
    </a>
    <a href="https://github.com/sourcegraph/sourcegraph/releases/">
        <img src="https://img.shields.io/github/release/sourcegraph/Sourcegraph.svg" alt="Latest release" />
    </a>
    <a href="https://srcgr.ph/discord">
        <img src="https://img.shields.io/discord/969688426372825169?color=5765F2" alt="Discord" />
    </a>
    <a href="https://github.com/sourcegraph/sourcegraph/contributors/">
        <img src="https://img.shields.io/github/contributors/sourcegraph/Sourcegraph.svg?color=000000" alt="Contributors" />
    </a>
</p>
<br />
<p align="center">
  <b>Understand, fix, and automate across your codebase with Sourcegraph's code intelligence platform</b>
</p>

---

## Use cases

- [Improve code security](https://about.sourcegraph.com/use-cases#code-security)
- [Accelerate developer onboarding](https://about.sourcegraph.com/use-cases#onboarding)
- [Resolve incidents faster](https://about.sourcegraph.com/use-cases#incident-response)
- [Streamline code reuse](https://about.sourcegraph.com/use-cases#code-reuse)
- [Boost code health](https://about.sourcegraph.com/use-cases#code-health)

## Features

- Fast global code search with a hybrid backend that combines a trigram index with in-memory streaming.
- Code intelligence for many languages via [SCIP](https://about.sourcegraph.com/blog/announcing-scip).
- Enhances GitHub, GitLab, Phabricator, and other code hosts and code review tools via the [Sourcegraph browser extension](https://docs.sourcegraph.com/integration/browser_extension).

## Try it now

Sourcegraph has three editions:

- [Sourcegraph Cloud](https://sourcegraph.com) lets you search over the open-source ecosystem plus your private code. [Search strings](https://sourcegraph.com/search?q=context:global+failed+to+ensure+HEAD+exists&patternType=literal), [search patterns](https://sourcegraph.com/search?q=context:global+lang:python+range%28len%28:%5B1%5D%29%29&patternType=structural), [search symbols](https://sourcegraph.com/search?q=context:global+type:symbol+lang:typescript+%28OA%7Coa%7COa%29uth+%5BHh%5Dandler+-file:%28%5E%7C/%29node_modules/+&patternType=regexp&case=yes) and [find references](https://sourcegraph.com/github.com/spf13/cobra@a684a6d7f5e37385d954dd3b5a14fc6912c6ab9d/-/blob/command.go?L221:19&subtree=true#tab=references) across your entire codebase and the open-source world.
- [Sourcegraph Enterprise](https://docs.sourcegraph.com/#getting-started) lets you run your own Sourcegraph instance in your own environment.
- [Sourcegraph OSS](#sourcegraph-oss) is an open-source version of Sourcegraph that provides the core functionality of Sourcegraph (code search, code browsing, basic code navigation), but lacks more advanced features (enterprise authentication, repository permissions, admin controls, advanced code navigation, etc.)

> Source code for all three editions is contained in this repository. See the [License section](#license) for more details.

More:

- Install the open-source [browser extension](https://chrome.google.com/webstore/detail/sourcegraph/dgjhfomjieaadpoljlnidmbgkdffpack?hl=en) to add Sourcegraph-like powers to your code review and code host.
- File feature requests and bug reports in [our issue tracker](https://github.com/sourcegraph/sourcegraph/issues).
- Visit [about.sourcegraph.com](https://about.sourcegraph.com) for more information about features, use cases, and organizations that use Sourcegraph.

## Self-hosted installation

### Sourcegraph Enterprise (free up to 10 users)

The fastest way to run Sourcegraph self-hosted is with the Docker container. See the [quickstart installation guide](https://docs.sourcegraph.com/#getting-started). There are also several additional ways of running a [production instance](https://docs.sourcegraph.com/admin/install).

### Sourcegraph OSS

1. Go through [Quickstart](https://docs.sourcegraph.com/dev/setup/quickstart) to install `sg` and dependencies
1. Start the development environment in OSS mode:

   ```sh
   sg start oss
   ```

Sourcegraph should now be running at https://sourcegraph.test:3443.

For detailed instructions and troubleshooting, see the [local development documentation](https://docs.sourcegraph.com/dev).

## Development

Refer to the [Developing Sourcegraph guide](https://docs.sourcegraph.com/dev) to get started.

### Documentation

The `doc` directory has additional documentation for developing and understanding Sourcegraph:

- [Project FAQ](./doc/admin/faq.md)
- [Architecture](./doc/dev/background-information/architecture/index.md): high-level architecture
- [Database setup](./doc/dev/background-information/postgresql.md): database best practices
- [Go style guide](./doc/dev/background-information/languages/go.md)
- [Documentation style guide](https://handbook.sourcegraph.com/engineering/product_documentation)
- [GraphQL API](./doc/api/graphql/index.md): useful tips when modifying the GraphQL API
- [Contributing](./CONTRIBUTING.md)

## License

This repository contains both OSS-licensed and non-OSS-licensed files. We maintain one repository rather than two separate repositories mainly for development convenience.

All files in the `enterprise` and `client/web/src/enterprise` fall under [LICENSE.enterprise](LICENSE.enterprise).

The remaining files fall under the [Apache 2 license](LICENSE.apache). Sourcegraph OSS is built only from the Apache-licensed files in this repository.
