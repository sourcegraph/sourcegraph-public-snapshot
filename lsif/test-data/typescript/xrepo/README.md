# LSIF data for TypeScript cross repo tests

The `./bin/generate.sh` to create LSIF dumps for a set of TypeScript projects that reference each other. This will create seven repositories and gzipped LSIF dump files which is used in the tests found in `typescript-xrepo.test.ts`.

The dump files used for testing are under version control, but can be regenerated to test changes in the indexer utilities.

### Dump Layout

The repository `a` defines the `math-util` package containing functions `add` and `mul`. The latter function is defined in terms of the former (and thus contains a eference to it).

The repositories `b{1,2,3}` have a dependency on `math-util` and import both `add` and `mul` functions.

The repositories `c{1,2,3}` have a dependency on `math-util` and import only the `add` function.

The TypeScript source for each project is contained in the script that generates the project, `./bin/generate-{a,b,c}.sh`.

### Requirements

This script requires you install [`lsif-tsc` and `lsif-npm`](https://github.com/microsoft/lsif-node). The `lsif-tsc` tool can be installed via npm. Unfortunately, Microsoft's implementation of `lsif-npm` is currently broken (but fixed with this [pull request](https://github.com/microsoft/lsif-node/pull/66)). The script must, for now, be run with Sourcegraph's fork of [lsif-npm](https://github.com/sourcegraph/lsif-node), which contains the update from the pull request. The location of the `lsif-npm` binary can be switched as follows.

```bash
LSIF_NPM=~/path/to/lsif-npm ./bin/generate.sh`
```
