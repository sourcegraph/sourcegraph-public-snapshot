<p align="center">
<a href="https://about.sourcegraph.com/" target="_blank">
<picture>
  <source media="(prefers-color-scheme: dark)" srcset="https://p21.p4.n0.cdn.getcloudapp.com/items/6qub2y6g/8c25cf68-2715-4f0e-9de6-26292fad604f.svg" width="50%">
  <img src="https://p21.p4.n0.cdn.getcloudapp.com/items/12u7NWXL/5e21725d-6e84-4ccd-8300-27bf9a050416.svg" width="50%">
</picture></a>
</p>

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

&nbsp;

---

## 5.0 Features

* Cody, your code-aware programmer's assistant
* Sourcegraph Own preview
* Improved code exploration experience
* A completely re-imagined search input
* Intelligent search ranking
* Improved auto-indexing setup experience
* Integrate Batch Changes with other tools with outgoing webhooks
* Limit access to batch changes
* Improved Code Insights support for instances with a large number of repositories
* Impoved Gerrit support with repository permissions
* Improved support for the Azure DevOps code host
* Improved rate limiting for GitHub and GitLab
* Permissions center
* Request account for unauthenticated users
* SCIM support

[Read more](https://about.sourcegraph.com/blog/release/5.0)

<p align="center">
<img src="https://storage.googleapis.com/sourcegraph-assets/blog/5.0/reimagined-search-input.png" width="75%">
</p>

## Deploy Sourcegraph

### Recommended

- [Sourcegraph App Beta](https://srcgr.ph/ufA1R): lightweight, single-binary version of Sourcegraph
- [Sourcegraph Cloud](https://docs.sourcegraph.com/cloud): create a single-tenant instance managed by Sourcegraph

### Self-hosted

- [AWS](https://docs.sourcegraph.com/admin/deploy/machine-images/aws-ami)
- [Azure](https://docs.sourcegraph.com/admin/deploy/docker-compose/azure)
- [DigitalOcean](https://docs.sourcegraph.com/admin/deploy/docker-compose/digitalocean)
- [Docker Compose](https://docs.sourcegraph.com/admin/deploy/docker-compose)
- [Google Cloud (GCP)](https://docs.sourcegraph.com/admin/deploy/images/gce)
- [Private Cloud](https://docs.sourcegraph.com/admin/deploy)
- [Kubernetes (Enterprise-only)](https://docs.sourcegraph.com/admin/deploy/kubernetes)

### Local machine

- [Docker](https://docs.sourcegraph.com/admin/deploy/docker-single-container)

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
