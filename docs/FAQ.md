# Sourcegraph Project FAQ

> Note: This document primarily talks about the Sourcegraph open source project itself, not e.g. features of Sourcegraph.

## What kinds of contributions are accepted?

See [CONTRIBUTING.md](../CONTRIBUTING.md).

## What license is Sourcegraph under?

Sourcegraph is under the [Apache License v2.0](../LICENSE).

## Is all of Sourcegraph open source?

Over 95% of it is.

The Docker images that we publish (i.e. the images users normally run) are built from our enterprise codebase. This enables our users to easily upgrade to [our paid offerings](https://about.sourcegraph.com/pricing/) without having to run a separate Docker image. These images include certain non open-source features like:

- SSO
- Advanced Code Intelligence
- Deploy via G Suite
- Internal–Only Extensions
- High-Availability Cluster
- Backup and Recovery

Aside from these few features that typically only larger companies desire, Sourcegraph is completely open-source.

> Note: If you wish to build a 100% open-source image, you can do so via the `dev/dev-sourcegraph-server.sh` script! =)
