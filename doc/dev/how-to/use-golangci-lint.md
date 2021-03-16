# Linting with `golangci-lint`

## How it works

Linting is the process of running static checks on the codebase to catch common mistakes and provide an automatically-enforceable set of best practices. We use a tool called [`golangci-lint`](https://golangci-lint.run/), which bundles a large number of common linters into a single binary.

`golangci-lint` is configured using the `.golangci.yml` and `.golangci.enforced.yml` files in the root of the repository. `.golangci.enforced.yml` contains a subset of the lints in `.golangci.yml`. In CI, the only lints that will cause a failure are the ones in `.golangci.enforced.yml`. By default, `golangci-lint` uses `.golangci.yml`, so you may see warnings in your editor or by running `golangci-lint run` that will not cause failures in CI. Eventually, we hope to unify these two configurations. A tracking issue with the progress can be found here: [#18720](https://github.com/sourcegraph/sourcegraph/issues/18720).

## Running the linters

The easiest way to check locally if your changes will pass the lint step in CI is to run `./dev/check/go-lint.sh`. This is run as part of `./dev/check/all.sh`, so if it passes, linting should be good in CI as well.

To run the extended set of linters (all enforced lints + some currently unenforced but recommended lints), you can run `golangci-lint run`, which will automatically pick up the config in `.golangci.yml`.

## Ignoring lints

We do our best to only enable lints that either reduce minor change churn (like `goimports`) or find common issues in our code (like `ineffassign`), but occasionally there will be false positives from the linters. In these cases, `golangci-lint` provides a way to ignore lint issues with a comment: `//nolint:lintname`. For more information on how to use `//nolint`, see the [`golangci-lint` false positives page](https://golangci-lint.run/usage/false-positives/).

If a lint is routinely ignored or is just adding to development noise, consider disabling it.

## Recommendations for making linting less annoying

The goal of linting is to provide value without excess development noise, but it can quickly become annoying to push up a change, make a PR, and see CI fail because of a misplaced import statement. Here are some ideas to make working with a linter less painful:

_Integrate the linter with your editor_

Most editors provide a way to provide in-line lints with `golangci-lint` as a source. This can significantly reduce development friction relative to running the linter outside of your editor since it will show you issues as you develop.

See the [integrations](https://golangci-lint.run/usage/integrations/) page in the `golangci-lint` docs for more info on integrating `golangci-lint` into your editor. 

_Run the lints before you push_

Depending on CI to run all your tests when you create a PR increase friction linting since you might wait 5-10 minutes for a failure. Running the lints in advance can alleviate that. 

Rather than attempting to remember to run `./dev/check/go-lint.sh` every time, some people prefer to add git hooks that run automatically on either commit or push.

_Use the linters for fast, iterative feedback_

Depending on your workflow, `golangci-lint` can often replace `go vet` or (sometimes) `go build` in the `write code -> build -> fix errors` cycle. One of the configured linters is the `typecheck` linter, which will catch most type errors that you'd normally get with `go build`. Using `golangci-lint` instead of `go vet` gives you compile warnings and lint warnings as part of your development cycle, cutting out the disconnect between developing and linting.
