# LSIF Typescript test data

Run `./generate.sh` to create LSIF dumps for a set of Typescript projects that reference each other. This will create gzipped dump files that are then used by `xrepo-typescript.test.ts`.

**Note**: Microsoft's implementation of lsif-npm is currently broken (see [this issue](https://github.com/microsoft/lsif-node/pull/66)). Therefore, the generate script must be run with Sourcegraph's version of [lsif-npm](https://github.com/sourcegraph/lsif-node). The binary can be switched via `LSIF_NPM=~/path/to/lsif-npm ./generate.sh`.

### Project structure

The repository `a` defines the npm package `math-util` that provides an `add` and a `mul` function. Refer to `./generate-a.sh` for the source.

The repositories `b1`, `b2`, and `b3` are identical except for their name. They depend on the `math-util` package and use both the `add` and `mul` functions. Refer to `./generate-b.sh` for the source.

The repositories `c1`, `c2`, and `c3` are identical except for their name. They depend on the `math-util` package and use only the `add` function. Refer to `./generate-c.sh` for the source.
