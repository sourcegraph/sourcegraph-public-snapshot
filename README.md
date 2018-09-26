# Sourcegraph Enterprise

[![build](https://badge.buildkite.com/f0e47ba39d32616d973b38e846f8e1aa25893920047221738e.svg?branch=master)](https://buildkite.com/sourcegraph/enterprise)
[![codecov](https://codecov.io/gh/sourcegraph/enterprise/branch/master/graph/badge.svg?token=itk6ydR7l3)](https://codecov.io/gh/sourcegraph/enterprise)
[![code style: prettier](https://img.shields.io/badge/code_style-prettier-ff69b4.svg)](https://github.com/prettier/prettier)

This repository contains all of the Sourcegraph Enterprise code.

## Project layout

- The main Sourcegraph codebase is open source, see [github.com/sourcegraph/sourcegraph](https://github.com/sourcegraph/sourcegraph).
- This codebase just wraps the open source codebase and links in some private code for enterprise features.
- Only the enterprise codebase is published to e.g. Docker Hub. Enterprise features are behind paywalls (or good-faith). The open source codebase is not published on Docker Hub (this avoids confusion and keeps the upgrade/downgrade process from open source <-> enterprise easy).

## Webapp

The webapp code pulls in the latest OSS webapp in CI from the `@sourcegraph/webapp` npm package, which gets published on every commit to the OSS repo.
This means the declared version in yarn.lock is expected to be outdated, and will be overridden on every build with the latest version.

## Dev

When developing locally, use `./dev/start.sh`. This will ensure you build against your locally checked out OSS repository by ensuring the following:

- The OSS web app is `npm link`ed into `node_modules`. It will run both the build of the enterprise
  webapp as well as the part of the build for the OSS repo that generates the distributed files for
  the npm package.
- The Go binaries are built with `replace github.com/sourcegraph/sourcegraph => ../sourcegraph` (via
  Go modules).

### Updating vendored dependencies

Vendored dependencies should be kept in sync with changes to `go.mod`. However, before you run `go mod vendor`, you must first unlink the `github.com/sourcegraph/sourcegraph` dependency from its local source to ensure the vendored files exist in a revision in the remote OSS repository:

```bash
GO111MODULE=on go mod edit -dropreplace github.com/sourcegraph/sourcegraph
GO111MODULE=on go mod vendor
GO111MODULE=on go mod tidy
GO111MODULE=on go mod edit -replace github.com/sourcegraph/sourcegraph=../sourcegraph
./dev/check/all.sh
```
