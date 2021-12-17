# Continuous integration

We have a variety of tooling on [Buildkite](https://buildkite.com/sourcegraph/sourcegraph) and [GitHub Actions](https://github.com/sourcegraph/sourcegraph/actions) for continuous integration.

> NOTE: To learn more about testing in particular, see our [testing principles](testing_principles.md).

## Buildkite pipelines

[Tests](../how-to/testing.md) are automatically run in our [various Buildkite pipelines](https://buildkite.com/sourcegraph).
Pipeline steps are generated using the [pipeline generator](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@main/-/tree/enterprise/dev/ci).

To see what checks will get run against your current branch, use [`sg`](../setup/quickstart.md):

```sh
sg ci preview
```

### Vulnerability scanning

Our CI pipeline scans uses [Trivy](https://aquasecurity.github.io/trivy/) to scan our Docker images for security vulnerabilities.

Trivy will perform scans upon commits to the following branches:

1. `main` 
2. branches prefixed by `main-dry-run/`
3. branches prefixed by `docker-images-patch/$IMAGE` (where only a single image is built)

If there are any `HIGH` or `CRITICAL` severities in a Docker image that have a known fix:

1. The CI pipeline will create an annotation that contains links to reports that describe the vulnerabilities
2. The Trivy scanning step will [soft fail](https://buildkite.com/docs/pipelines/command-step#soft-fail-attributes). Note that soft failures **do not fail builds or block deployments**. They simply highlight the failing step for further analysis.

> NOTE: Our vulnerability management process (including this workflow) is under active development and in its early stages. All of the above is subject to change. See [https://github.com/sourcegraph/sourcegraph/pull/25756](https://github.com/sourcegraph/sourcegraph/pull/25756) for more context.

### Pipeline health

Maintaining [Buildkite pipeline](#buildkite-pipelines) health is a critical part of ensuring we ship a stable product - changes that make it to the `main` branch may be deployed to various Sourcegraph instances, and having a reliable and predictable pipeline is crucial to ensuring bugs do not make it to production environments.

To enable this, we want to [address flakes as they arise](#flakes) and have tooling to mitigate the impacts of pipeline instability, such as [`buildchecker`](#buildchecker).

> NOTE: Sourcegraph teammates should refer to the [CI incidents playbook](https://handbook.sourcegraph.com/departments/product-engineering/engineering/process/incidents/playbooks/ci#scenarios) for help managing issues with pipeline health.

#### Flakes

A flake is generally characterized as one-off or rare issues that can be resolved by retrying the failed job or task. In other words: something that sometimes fails, but if you retry it enough times, it passes, *eventually*.

Tests are not the only thing that are flaky - flakes can also encompass sporadic infrastructure issues and other unreliable steps.

##### Flaky tests

Learn more about our flaky test policy in [Testing principles: Flaky tests](testing_principles.md#flaky-tests).

Use language specific functionality to skip a test. Create an issue and ping an owner about the skipping (normally on the PR skipping it).

- Go: [`testing.T.Skip`](https://pkg.go.dev/testing#hdr-Skipping)
- Typescript: [`.skip()`](https://mochajs.org/#inclusive-tests)

If the language or framework allows for a skip reason, include a link to the issue track re-enabling the test, or leave a docstring with a link.

##### Flaky steps

If a step is flaky we need to get the build back to reliable as soon as possible. If there is not already a discussion in `#buildkite-main` create one and link what step you take. Here are the recommended approaches in order:

1. Revert the PR if a recent change introduced the instability. Ping author.
2. Use `Skip` StepOpt when creating the step. Include reason and a link to context. This will still show the step on builds so we don't forget about it.

An example use of `Skip`:

```diff
--- a/enterprise/dev/ci/internal/ci/operations.go
+++ b/enterprise/dev/ci/internal/ci/operations.go
@@ -260,7 +260,9 @@ func addGoBuild(pipeline *bk.Pipeline) {
 func addDockerfileLint(pipeline *bk.Pipeline) {
        pipeline.AddStep(":docker: Lint",
                bk.Cmd("./dev/ci/docker-lint.sh"),
+               bk.Skip("2021-09-29 example message https://github.com/sourcegraph/sourcegraph/issues/123"),
        )
 }
```

#### `buildchecker`

[`buildchecker`](https://github.com/sourcegraph/sourcegraph/actions/workflows/buildchecker.yml) is a tool responding to periods of consecutive build failures on the `main` branch Sourcegraph Buildkite pipeline. If it detects a series of failures on the `main` branch, merges to `main` will be restricted to certain members of the Sourcegraph team until the issue is resolved.

To learn more, refer to the [`buildchecker` source code and documentation](https://github.com/sourcegraph/sourcegraph/tree/main/dev/buildchecker).

### Pipeline development

The source code of the pipeline generator is in [`/enterprise/dev/ci`](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@main/-/tree/enterprise/dev/ci).

To test the rendering of the entire pipeline, you can run `env BUILDKITE_BRANCH=TESTBRANCH go run ./enterprise/dev/ci/gen-pipeline.go` and inspect the YAML output. To change the behaviour set the relevant `BUILDKITE_` environment variables.

> WARNING: Sourcegraph's pipeline generator and its generated output are under the [Sourcegraph Enterprise license](https://github.com/sourcegraph/sourcegraph/blob/main/LICENSE.enterprise).

#### Pipeline operations

Pipeline steps are defined as [`Operation`s](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/enterprise/dev/ci/internal/ci/operations/operations.go) that apply changes to the given pipeline, such as adding steps and components.

```sgquery
(:[_] *bk.Pipeline) patternType:structural repo:^github\.com/sourcegraph/sourcegraph$ file:^enterprise/dev/ci/internal/ci/operations\.go
```

Within an `Operation` you will typically create one or more steps on a pipeline with `AddStep`, which can be configured with options of type `SteptOpt`.

```sqquery
(:[_]) StepOpt patternType:structural repo:^github\.com/sourcegraph/sourcegraph$ file:^enterprise/dev/ci/internal/buildkite/buildkite\.go
```

Operations are then added to a pipeline from [`GeneratePipeline`](https://sourcegraph.com/search?q=context:global+repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+file:%5Eenterprise/dev/ci/internal/ci/pipeline%5C.go+GeneratePipeline&patternType=literal).

For most basic PR checks, see [Creating PR checks](#creating-pr-checks) for how to create your own steps!

For more advanced usage for specific run types, see [Run types](#run-types).

#### Creating PR checks

To create a new check that can run on pull requests on relevant files, check the [`changed.Files`](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/enterprise/dev/ci/internal/ci/changed/changed.go) type to see if a relevant `affectsXyz` check already exists.

```sgquery
Affects type:symbol select:symbol.function repo:^github\.com/sourcegraph/sourcegraph$ file:^enterprise/dev/ci/internal/ci/changed
```

If not, you can define a new one on the `changed.Files` type.

Then, you can add a new check to [`CoreTestOperations`](https://sourcegraph.com/search?q=context:global+repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+file:%5Eenterprise/dev/ci/internal/ci+CoreTestOperations+type:symbol+&patternType=literal).
Make sure to follow the best practices outlined in docstring.

For more advanced pipelines, see [Run types](#run-types).

#### Run types

There are a variety of run types available based on branch prefixes. These generate special-purpose pipelines. For example, the `main-dry-run/` prefix is used to generate a pipeline similar to the default `main` branch. See [`RunType`](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/enterprise/dev/ci/internal/ci/runtype.go) for the various run types available, and examples for how to add more.

[`GeneratePipeline`](https://sourcegraph.com/search?q=context:global+repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+file:%5Eenterprise/dev/ci/internal/ci/pipeline%5C.go+GeneratePipeline&patternType=literal) can leverage `RunType` when generating pipelines, for example:

```sgquery
RunType.Is(:[_]) OR case :[_]: patternType:structural repo:^github\.com/sourcegraph/sourcegraph$ file:^enterprise/dev/ci/internal/ci/pipeline\.go
```

For simple PR checks, see [Creating PR checks](#creating-pr-checks).

#### Buildkite infrastructure

##### Pipeline setup

To set up Buildkite to use the rendered pipeline, add the following step in the [pipeline settings](https://buildkite.com/sourcegraph/sourcegraph/settings):

```shell
go run ./enterprise/dev/ci/gen-pipeline.go | buildkite-agent pipeline upload
```

##### Managing secrets

The term _secret_ refers to authentication credentials like passwords, API keys, tokens, etc. which are used to access a particular service. Our CI pipeline must never leak secrets:

- to add a secret, use the Secret Manager on Google Cloud and then inject it at deployment time as an environment variable in the CI agents, which will make it available to every step.
- use an environment variable name with one of the following suffixes to ensure it gets redacted in the logs: `*_PASSWORD, *_SECRET, *_TOKEN, *_ACCESS_KEY, *_SECRET_KEY, *_CREDENTIALS`
- while environment variables can be assigned when declaring steps, they should never be used for secrets, because they won't get redacted, even if they match one of the above patterns.

## GitHub Actions

### Third-party licenses

We use the [`license_finder`](https://github.com/pivotal/LicenseFinder) tool to check third-party dependencies for their licenses. It runs as a [GitHub Action on pull requests](https://github.com/sourcegraph/sourcegraph/actions?query=workflow%3A%22Licenses+Check%22), which will fail if one of the following occur:

- If the license for a dependency cannot be inferred. To resolve:
  - Use `license_finder licenses add <dep> <license>` to set the license manually
- If the license for a new or updated dependency is not on the list of approved licenses. To resolve, either:
  - Remove the dependency
  - Use `license_finder ignored_dependencies add <dep> --why="Some reason"` to ignore it
  - Use `license_finder permitted_licenses add <license> --why="Some reason"` to allow the offending license

The `license_finder` tool can be installed using `gem install license_finder`. You can run the script locally using:

```sh
# updates ThirdPartyLicenses.csv
./dev/licenses.sh

# runs the same check as the one used in CI, returning status 1
# if there are any unapproved dependencies ('action items')
LICENSE_CHECK=true ./dev/licenses.sh
```

The `./dev/licenses.sh` script will also output some `license_finder` configuration for debugging purposes - this configuration is based on the `doc/dependency_decisions.yml` file, which tracks decisions made about licenses and dependencies.

For more details, refer to the [`license_finder` documentation](https://github.com/pivotal/LicenseFinder#usage).

