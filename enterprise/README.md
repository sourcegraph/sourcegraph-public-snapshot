# Sourcegraph Enterprise

[![build](https://badge.buildkite.com/f0e47ba39d32616d973b38e846f8e1aa25893920047221738e.svg?branch=master)](https://buildkite.com/sourcegraph/enterprise)
[![codecov](https://codecov.io/gh/sourcegraph/enterprise/branch/master/graph/badge.svg?token=itk6ydR7l3)](https://codecov.io/gh/sourcegraph/enterprise)
[![code style: prettier](https://img.shields.io/badge/code_style-prettier-ff69b4.svg)](https://github.com/prettier/prettier)

This repository contains all of the Sourcegraph Enterprise code.

## Project layout

- The main Sourcegraph codebase is open source, see [github.com/sourcegraph/sourcegraph](https://github.com/sourcegraph/sourcegraph).
- This codebase just wraps the open source codebase and links in some private code for enterprise features.
- Only the enterprise codebase is published to e.g. Docker Hub. Enterprise features are behind paywalls (or good-faith). The open-source codebase is not published on Docker Hub (this avoids confusion and keeps the upgrade/downgrade process from open source <-> enterprise easy).

## Dev

See [README.dev.md](README.dev.md).

### Updating dependencies

- `go get -u $MODULE` to update `$MODULE` and all its transitive dependencies to their latest version.
- `go mod edit -replace $MODULE@$VERSION` to update `$MODULE` to a specific version.
- `go mod tidy` if updates to `go.mod` or `go.sum` have been made as a result of other build
  invocations during development and you wish now to update `go.mod` and `go.sum` to be consistent
  with how the build will run in CI.
