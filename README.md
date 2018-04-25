# Sourcegraph CLI [![Build Status](https://travis-ci.org/sourcegraph/src-cli.svg)](https://travis-ci.org/sourcegraph/src-cli) [![Build status](https://ci.appveyor.com/api/projects/status/fwa1bkd198hyim8a?svg=true)](https://ci.appveyor.com/project/sourcegraph/src-cli) [![Go Report Card](https://goreportcard.com/badge/sourcegraph/src-cli)](https://goreportcard.com/report/sourcegraph/src-cli)

The Sourcegraph CLI provides access to [Sourcegraph](https://sourcegraph.com) via a command line interface.

## Status of the project

Currently, the `src` CLI only provides access to Sourcegraph's GraphQL API. It lets you:

- Execute GraphQL queries against a Sourcegraph instance, and get JSON results back.
- Provide your API access token via an environment variable or file on disk.

**In the future** it may provide much more:

- Executing search queries from the command line easily and getting formatted results back,
- A command-line management interface for Sourcegraph instances: adding users, deleting users, etc.

## Installation

### Mac OS:

```bash
curl https://github.com/sourcegraph/src-cli/releases/download/latest/src_darwin_amd64 -o /usr/local/bin/src
chmod +x /usr/local/bin/src
```

### Linux:

```bash
curl https://github.com/sourcegraph/src-cli/releases/download/latest/src_linux_amd64 -o /usr/local/bin/src
chmod +x /usr/local/bin/src
```

### Windows:

Note: Windows support is still rough around the edges, but is available. If you encounter issues, please let us know by filing an issue :)

- [Download the latest src_windows_amd64.exe](https://github.com/sourcegraph/src-cli/releases/download/latest/src_windows_amd64.exe) and rename to `src.exe`.
- Place the file under e.g. `C:\Program Files\Sourcegraph\src.exe` and add that directory to your system path to access it from any command prompt.

## Usage

Consult `src -h` and `src api -h` for usage information.

## Development

If you want to develop the CLI, you can install it with `go get`:

```
go get -u github.com/sourcegraph/src-cli/cmd/src
```

## Releasing

1. Find the latest version (either via the releases tab on GitHub or via git tags) to determine which version you are releasing.
2. `VERSION=9.9.9 ./release.sh` (replace `9.9.9` with the version you are releasing)

Travis will automatically perform the release.
