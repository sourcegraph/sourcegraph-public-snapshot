# Frontends with Bazel

This folder contains various examples for writing JavaScript applications with Bazel.

Bazel's [rules_js] uses the pnpm package manager. This folder is the root of a pnpm workspace.
This allows npm packages within this monorepo to depend on each other.

See the README.md file under each folder here to understand more about that package.

## Linting

We demonstrate the usage of [rules_lint]. There are a few ways to wire this up, we show two:
- *build failure*: in the `next.js` folder, `npm run lint` does a `bazel build` with a config setting that makes the build fail when lint violations are found.
- *test failure*: in the `react/src` folder, an `eslint_test` target results in test failures when lint violations are found.

However, in both cases this inhibits creation of new lint rules or even upgrading the linter, because it requires updating the entire repository to fix or suppress
new lint violations at the same time.
We recommend showing lint results during code review instead.
See <https://github.com/aspect-build/rules_lint/blob/main/docs/linting.md>

[rules_js]: https://docs.aspect.build/rules/aspect_rules_js
[rules_lint]: https://github.com/aspect-build/rules_lint
