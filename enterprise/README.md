# Sourcegraph Enterprise

[![build](https://badge.buildkite.com/f0e47ba39d32616d973b38e846f8e1aa25893920047221738e.svg?branch=master)](https://buildkite.com/sourcegraph/enterprise)
[![codecov](https://codecov.io/gh/sourcegraph/enterprise/branch/master/graph/badge.svg?token=itk6ydR7l3)](https://codecov.io/gh/sourcegraph/enterprise)
[![code style: prettier](https://img.shields.io/badge/code_style-prettier-ff69b4.svg)](https://github.com/prettier/prettier)

This directory contains all of the Sourcegraph Enterprise code.

## Dev

### Build notes

You'l need to clone https://github.com/sourcegraph/dev-private to the same directory that contains
this repository.

**IMPORTANT:** Commands that build enterprise targets (e.g., `go build`, `yarn`,
`enterprise/dev/go-install.sh`) should always be run with the `enterprise` directory as the current
working directory. Otherwise, build tools like `yarn` and `go` may try to update the root
`package.json` and `go.mod` files as a side effect, instead of updating `enterprise/package.json`
and `enterprise/go.mod`.

The OSS web app is `yarn link`ed into `enterprise/node_modules`. It will run both the build of the
enterprise webapp as well as the part of the build for the OSS repo that generates the distributed
files for the npm package.

### Updating dependencies

- `go get -u $MODULE` to update `$MODULE` and all its transitive dependencies to their latest version.
- `go mod edit -replace $MODULE@$VERSION` to update `$MODULE` to a specific version.
- `go mod tidy` if updates to `go.mod` or `go.sum` have been made as a result of other build
  invocations during development and you wish now to update `go.mod` and `go.sum` to be consistent
  with how the build will run in CI.
