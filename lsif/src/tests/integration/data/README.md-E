# Generated test data

This directory contains gzipped LSIF dump data used by backend and database unit tests.

## Usage

To (re-)generate the test data, run `./generate.sh`. This will create TypeScript projects on-disk (according to the `bin/generate.sh` script in each subdirectory), run the `lsif-tsc` indexer and the `lsif-npm` moniker utility, then destroy the generated projects. The resulting LSIF dumps are gzipped.

## Requirements

The generation script requires that you install both [`lsif-tsc` and `lsif-npm`](https://github.com/microsoft/lsif-node). The `lsif-tsc` tool can be installed via npm. The `lsif-npm` tool must, for now, be installed via Sourcegraph's [fork](https://github.com/sourcegraph/lsif-node) (see [pull request](https://github.com/microsoft/lsif-node/pull/66)). The location of the `lsif-tsc` and `lsif-npm` binaries can be changed with the `LSIF_TSC` and `LSIF_NPM` environment variables.
