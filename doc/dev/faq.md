# Sourcegraph open-source project FAQ

This document is about the Sourcegraph _open source project_, not the product or features of Sourcegraph.

## What kinds of contributions are accepted?

See [CONTRIBUTING.md](https://github.com/sourcegraph/sourcegraph/blob/master/CONTRIBUTING.md).

## What license is Sourcegraph released under?

Sourcegraph is released under the [Apache License v2.0](https://github.com/sourcegraph/sourcegraph/blob/master/LICENSE).

## Is all of Sourcegraph open source?

This repository is 100% open source and builds a product known as Sourcegraph OSS. Sourcegraph OSS omits certain trademarks, logos, and [paid enterprise features](https://about.sourcegraph.com/pricing/) from the official Sourcegraph build to make it open-source.

The official, free Sourcegraph build is called Sourcegraph Core and includes these additions, so it's not open source. The reason why we include these features in the official build is to provide a smooth upgrade path to Sourcegraph Enterprise (you can just supply a license key to activate the features, with no migration necessary).

> NOTE: To build a 100% open-source `sourcegraph/server` image, use `dev/dev-sourcegraph-server.sh`.
