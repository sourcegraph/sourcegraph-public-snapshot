<a href="https://about.sourcegraph.com/sourcegraph-4" target="_blank">
  <picture>
    <source srcset="https://p21.p4.n0.cdn.getcloudapp.com/items/xQuxo7AA/a9e1873b-c7b2-4295-b96a-dd21c992eb63.svg" media="(prefers-color-scheme: dark)">
    <img src="https://p21.p4.n0.cdn.getcloudapp.com/items/ApuDmzGr/822118fe-e5e6-4f37-8c17-8b406437ad03.svg" alt="Sourcegraph 4.0" width="100%">
  </picture>
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

&nbsp;

---

## 4.0 Features

### 🧠 Code intelligence: uplevel your code search

- Understand usage and search structure with high-level aggregations of search results
- A faster, simpler search experience
- Configure precise code navigation for 9 languages (Ruby, Rust, Go, Java, Scala, Kotlin, Python, TypeScript, JavaScript) in a matter of minutes with auto-indexing
- Your favorite extensions are now available by default
- Quickly access answers within your codebase with a revamped reference panel

<p align="center">
<img src="https://storage.googleapis.com/sourcegraph-assets/blog/release-post/4.0/New-Search-UI.png" width="75%">
</p>

### 🏗️ High-leverage ways to improve your entire codebase

- Make changes across all of your codebase at enterprise scale with server-side Batch Changes (beta)
  - Run large-scale or resource-intensive batch changes without clogging your local machine
  - Run large batch changes quickly by distributing them across an autoscaled pool of compute instances
  - Get a better debugging experience with the streaming of logs directly into Sourcegraph.

### ☁️ Dedicated Sourcegraph Cloud instances for enterprise

- Sourcegraph Cloud now offers dedicated, single-tenant instances of Sourcegraph

### 📈 Advanced admin capabilities

- Save time upgrading to Sourcegraph 4.0 with multi-version upgrades
- View usage and measure the value of our platform with new and enhanced in-product analytics
- Uncover developer time saved using Browser and IDE extensions
- Easily export traces using OpenTelemetry
- Quickly see the status on your repository and permissions syncing
- Measure precise code navigation coverage with an enhanced analytics dashboard

<p align="center">
<img src="https://storage.googleapis.com/sourcegraph-assets/blog/release-post/4.0/Search.png" width="75%">
</p>

## Deploy Sourcegraph

### Recommended

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
