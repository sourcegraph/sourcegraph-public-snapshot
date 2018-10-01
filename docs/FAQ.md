# Sourcegraph Project FAQ

> Note: This document primarily talks about the Sourcegraph open source project itself, not e.g. features of Sourcegraph.

## What kinds of contributions are accepted?

See [CONTRIBUTING.md](../CONTRIBUTING.md).

## What license is Sourcegraph under?

Sourcegraph is under the [Apache License v2.0](../LICENSE).

## Is all of Sourcegraph open source?

This repository is 100% open source and builds a product known as Sourcegraph OSS.

Sourcegraph OSS omits certain trademarks, logos, and [paid enterprise features](https://about.sourcegraph.com/pricing/). Sourcegraph Core and Enterprise include these additions, which make them not open source.

The Docker images that users normally run are Sourcegraph Core (including paid enterprise features behind a paywall) which enables a smooth upgrade process for our users. Sourcegraph Core is completely free, however.

> Note: If you wish to build a 100% open-source image, you can do so via the `dev/dev-sourcegraph-server.sh` script! =)
