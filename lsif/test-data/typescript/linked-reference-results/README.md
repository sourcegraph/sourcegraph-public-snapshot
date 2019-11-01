# LSIF data for TypeScript linked reference result tests

The `./bin/generate.sh` to create an LSIF dumps for TypeScript project that produces linked reference results (namely when implementing methods from an interface or base class). This will create a gzipped LSIF dump file which is used in the tests found in `typescript-linked-reference-results.test.ts`.

The dump files used for testing are under version control, but can be regenerated to test changes in the indexer utilities.

### Requirements

This script requires you install [`lsif-tsc` and `lsif-npm`](https://github.com/microsoft/lsif-node). The `lsif-tsc` tool can be installed via npm. Unfortunately, Microsoft's implementation of `lsif-npm` is currently broken (but fixed with this [pull request](https://github.com/microsoft/lsif-node/pull/66)). The script must, for now, be run with Sourcegraph's fork of [lsif-npm](https://github.com/sourcegraph/lsif-node), which contains the update from the pull request. The location of the `lsif-npm` binary can be switched as follows.

```bash
LSIF_NPM=~/path/to/lsif-npm ./bin/generate.sh`
```
