# LSIF data for TypeScript linked reference result tests

The `./bin/generate.sh` to create an LSIF dumps for TypeScript project that produces linked reference results (namely when implementing methods from an interface or base class). This will create a gzipped LSIF dump file which is used in the tests found in `typescript-linked-reference-results.test.ts`.

The dump files used for testing are under version control, but can be regenerated to test changes in the indexer utilities.

### Requirements

This script requires you install [`lsif-tsc`](https://github.com/microsoft/lsif-node). The `lsif-tsc` tool can be installed via npm.
