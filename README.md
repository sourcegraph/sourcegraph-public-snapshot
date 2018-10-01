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

### Updating dependencies

Use `dev/update-oss-dep.sh` to update the webapp and backend `github.com/sourcegraph/sourcegraph`
dependencies to their latest revisions. In general, `dev/update-oss-dep.sh` should only be run on
the `master` branch and its changes pushed directly to `master`.

- If you only want to update the backend `github.com/sourcegraph/sourcegraph` dependency, run `./dev/go-mod-update.sh`.
- If you only want to update the webapp dependency, run `yarn upgrade @sourcegraph/webapp@latest`.
- For other backend dependency updates, use `dev/go $args` instead of `go $args`. That script will
  wrap the `go` invocation by removing the `replace github.com/sourcegraph/sourcegraph` directive
  from `go.mod` and calling `go mod tidy` afterward. For instance,
  - `dev/go get -u $MODULE` to update `$MODULE` and all its transitive dependencies to their latest version.
  - `dev/go mod edit -replace $MODULE@$VERSION` to update `$MODULE` to a specific version.
  - `dev/go mod tidy` if updates to `go.mod` or `go.sum` have been made as a result of other build
    invocations during development and you wish now to update `go.mod` and `go.sum` to be consistent
    with how the build will run in CI.
