# Sourcegraph CLI [![Build Status](https://travis-ci.org/sourcegraph/src-cli.svg)](https://travis-ci.org/sourcegraph/src-cli) [![Go Report Card](https://goreportcard.com/badge/sourcegraph/src-cli)](https://goreportcard.com/report/sourcegraph/src-cli)

The Sourcegraph CLI provides access to [Sourcegraph](https://sourcegraph.com) via a command line interface.

## Status of the project

Currently, the `src` CLI only provides access to Sourcegraph's GraphQL API. It lets you:

- Execute GraphQL queries against a Sourcegraph instance, and get JSON results back.
- Provide your API access token via an environment variable or file on disk.

**In the future** it may provide much more:

- Executing search queries from the command line easily and getting formatted results back,
- A command-line management interface for Sourcegraph instances: adding users, deleting users, etc.

## Installation

WIP

## Usage

WIP

## Development

If you want to develop the CLI, you can install it with `go get`:

```
go get -u github.com/sourcegraph/src-cli/cmd/src
```
