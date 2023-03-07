# Bazel for client/*

See [bazel.md](./bazel.md) for general bazel info/setup.

## Tools

The sourcegraph client projects are setup to compile, bundle and test with Bazel. The tools used within Bazel include:
* Babel for ts\[x\] transpilation
* Webpack for bundling
* Jest, Mocha for testing
* Node tools such as graphql-codegen for generating graphql schema types

The Bazel rulesets used to support these include:
* [rules_js](https://github.com/aspect-build/rules_js)
* [rules_ts](https://github.com/aspect-build/rules_ts)
* [rules_jest](https://github.com/aspect-build/rules_jest)
* [rules_webpack](https://github.com/aspect-build/rules_webpack)
* [rules_esbuild](https://github.com/aspect-build/rules_esbuild)

See [Aspect rules docs](https://docs.aspect.build/rules/) for more information on the Bazel rulesets used.

## Targets

The primary Bazel targets have been configured roughly aligning with the pnpm workspace projects, while often composed of many sub-targets. The primary targets for `client/*` pnpm projects are generated using `bazel configure`. The primary targets include:
* `:{name}_pkg` for the npm package representing the pnpm project
* `:test` for the Jest unit tests of the project
* `:{name}_lib` for compiling non-test `*.ts\[x\]` files
* `:{name}_tests` for compiling unit test `*.ts\[x\]` files

Other targets may be configured per project such as sass compilation, graphql schema generation etc. See `client/*/BUILD.bazel` files for more details and examples.

Additional `BUILD.bazel` files may exist throughout subdirectories and is encouraged to create many smaller independently cacheable targets.

## Testing

All client tests (of all types such as jest and mocha) can be invoked by `bazel test //client/...` or individual tests can be specified such as `bazel test //client/common:test` or `bazel test //client/web/src/end-to-end:e2e`. Jest tests can be debugged using `bazel run --config=debug //client/common:test`.

## Bundling

The primary `client/web` bundle targets are:
* `//client/web:bundle`
* `//client/web:bundle-dev`
* `//client/web:bundle-enterprise`
See `client/web/BUILD`.

The `client/web` devserver can be run using `bazel run //client/web:devserver`

## Rule configuration

Most rules used throughout `client/*` are macros defined in `tools/*.bzl` to provide a consistent configuration used throughout the repo.
