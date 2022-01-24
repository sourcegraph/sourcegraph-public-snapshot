# Continuous integration <span class="badge badge-note">SOC2/GN-105</span> <span class="badge badge-note">SOC2/GN-106</span>

Sourcegraph uses a continuous integration and delivery tool, [Buildkite](#buildkite-pipelines), to help ensure a [consistent](#pipeline-health) build, test and deploy process. Software changes are systematically required to complete all steps within the continuous integration tool workflow prior to production deployment, in addition to being [peer reviewed](pull_request_reviews.md).

Sourcegraph also maintains a variety of tooling on [GitHub Actions](#github-actions) for continuous integration and repository maintainence purposes.

> NOTE: To learn more about testing in particular, see our [testing principles](testing_principles.md).

## Buildkite pipelines

[Tests](../how-to/testing.md) are automatically run in our [various Buildkite pipelines](https://buildkite.com/sourcegraph).
Pipeline steps are generated using the [pipeline generator](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@main/-/tree/enterprise/dev/ci).

To see what checks will get run against your current branch, use [`sg`](../setup/quickstart.md):

```sh
sg ci preview
```

To learn about making changes to our Buildkite pipelines, see [Pipeline development](#pipeline-development).

### Pipeline steps

A complete reference of all available pipeline steps is not yet available ([#30203](https://github.com/sourcegraph/sourcegraph/issues/30203)). This section contains a high-level documentation about what runs in our pipeline.

#### Soft failures

<span class="badge badge-note">SOC2/GN-106</span>

Many steps in Sourcegraph's Buildkite pipelines allow for [soft failures](https://buildkite.com/changelog/56-command-steps-can-now-be-made-to-soft-fail), which means that even if they fail they do not cause the entire build to be failed.

In the Buildkite UI, soft failures currently look like the following, with a _triangular_ warning sign (not to be mistaken for a hard failure!):

![soft fail in Buildkite UI](https://user-images.githubusercontent.com/23356519/150558751-d8e0da19-0b6f-4645-aa12-7547d375330f.png)

We use soft failures for the following reasons only:

- Steps that determine whether a subsequent step should run, where soft failures are the only technical way to communicate that a later step should be skipped in this manner using Buildkite.
  - Examples: [hash comparison steps that determine if a build should run](https://sourcegraph.com/search?q=context:%40sourcegraph/all+repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+file:%5Eenterprise/dev/ci/internal/ci/operations%5C.go+compare-hash.sh&patternType=literal)
- Regular analysis tasks, where soft failures serve as an monitoring indicator to warn the team responsible for fixing issues.
  - Examples: [image vulnerability scanning](#image-vulnerability-scanning), linting tasks for catching deprecation warnings
- Temporary exceptions to accommodate experimental or in-progress work.

You can find all usages of soft failures [with the following queries](https://sourcegraph.com/notebooks/Tm90ZWJvb2s6NTc=):

- [Soft failures in the Sourcegraph pipeline generator](https://sourcegraph.com/search?q=repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+%7B...bk.SoftFail...%7D+OR+%28...bk.SoftFail...%29+count:all&patternType=structural)
- [Soft failures in Buildkite YAML pipelines](https://sourcegraph.com/search?q=repo:%5Egithub%5C.com/sourcegraph/.*+soft_fail+lang:yaml+count:all&patternType=literal)

All other failures are hard failures.

#### Image vulnerability scanning

Our CI pipeline scans uses [Trivy](https://aquasecurity.github.io/trivy/) to scan our Docker images for security vulnerabilities.

Trivy will perform scans upon commits to the following branches:

1. `main`
2. branches prefixed by `main-dry-run/`
3. branches prefixed by `docker-images-patch/$IMAGE` (where only a single image is built)

If there are any `HIGH` or `CRITICAL` severities in a Docker image that have a known fix:

1. The CI pipeline will create an annotation that contains links to reports that describe the vulnerabilities
2. The Trivy scanning step will [soft fail](#soft-failures). Note that soft failures **do not fail builds or block deployments**. They simply highlight the failing step for further analysis.

> NOTE: Our vulnerability management process (including this workflow) is under active development and in its early stages. All of the above is subject to change. See [https://github.com/sourcegraph/sourcegraph/pull/25756](https://github.com/sourcegraph/sourcegraph/pull/25756) for more context.

We also run [separate vulnerability scans for our infrastructure](https://handbook.sourcegraph.com/departments/product-engineering/engineering/cloud/security/checkov).

### Pipeline health

Maintaining [Buildkite pipeline](#buildkite-pipelines) health is a critical part of ensuring we ship a stable product - changes that make it to the `main` branch may be deployed to various Sourcegraph instances, and having a reliable and predictable pipeline is crucial to ensuring bugs do not make it to production environments.

To enable this, we [address flakes as they arise](#flakes) and mitigate the impacts of pipeline instability with [branch locks](#branch-locks).

> NOTE: Sourcegraph teammates should refer to the [CI incidents playbook](https://handbook.sourcegraph.com/departments/product-engineering/engineering/process/incidents/playbooks/ci#scenarios) for help managing issues with pipeline health.

#### Branch locks

> WARNING: **A red `main` build is not okay and must be fixed.** Learn more about our `main` branch policy in [Testing principles: Failures on the `main` branch](testing_principles.md#failures-on-the-main-branch).

[`buildchecker`](#buildchecker) is a tool responding to periods of consecutive build failures on the `main` branch Sourcegraph Buildkite pipeline. If it detects a series of failures on the `main` branch, merges to `main` will be restricted to members of the Sourcegraph team who authored the failing commits until the issue is resolved - this is referred to as a "branch lock". When a build passes on `main` again, `buildchecker` will automatically unlock the branch.

**Authors of the most recent failed builds are responsible for investigating failures.** Please refer to the [Continuous integration playbook](https://handbook.sourcegraph.com/departments/product-engineering/engineering/process/incidents/playbooks/ci#build-has-failed-on-the-main-branch) for step-by-step guides on what to do in various scenarios.

#### Flakes

A *flake* is defined as a test or script that is unreliable or non-deterministic, i.e. it exhibits both a passing and a failing result with the same code. In other words: something that sometimes fails, but if you retry it enough times, it passes, *eventually*.

Tests are not the only thing that are flaky - flakes can also encompass [sporadic infrastructure issues](#flaky-infrastructure) and [unreliable steps](#flaky-steps).

##### Flaky tests

> WARNING: **We do not tolerate flaky tests of any kind.** Learn more about our flaky test policy in [Testing principles: Flaky tests](testing_principles.md#flaky-tests).

Typical reasons why a test may be flaky:

- Race conditions or timing issues
- Caching or inconsistent state between tests
- Unreliable test infrastructure (such as CI)
- Reliance on third-party services that are inconsistent

If a flaky test is discovered, immediately use language-specific functionality to skip a test and open a PR to disable the test:

- Go: [`testing.T.Skip`](https://pkg.go.dev/testing#hdr-Skipping)
- Typescript: [`.skip()`](https://mochajs.org/#inclusive-tests)

If the language or framework allows for a skip reason, include a link to the issue track re-enabling the test, or leave a docstring with a link.

Then open an issue to investigate the flaky test (use the [flaky test issue template](https://github.com/sourcegraph/sourcegraph/issues/new/choose)), and assign it to the most likely owner.

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

##### Flaky infrastructure

If the [build or test infrastructure itself is flaky](https://handbook.sourcegraph.com/departments/product-engineering/engineering/enablement/dev-experience#build-pipeline-support), then [open an issue with the `team/devx` label](https://github.com/sourcegraph/sourcegraph/issues/new?labels=team/devx) and notify the [Developer Experience team](https://handbook.sourcegraph.com/departments/product-engineering/engineering/enablement/dev-experience#contact).

Also see [Buildkite infrastructure](#buildkite-infrastructure).

### Pipeline development

The source code of the pipeline generator is in [`/enterprise/dev/ci`](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@main/-/tree/enterprise/dev/ci).

To test the rendering of the entire pipeline, you can run `env BUILDKITE_BRANCH=TESTBRANCH go run ./enterprise/dev/ci/gen-pipeline.go` and inspect the YAML output. To change the behaviour set the relevant `BUILDKITE_` environment variables.

> WARNING: Sourcegraph's pipeline generator and its generated output are under the [Sourcegraph Enterprise license](https://github.com/sourcegraph/sourcegraph/blob/main/LICENSE.enterprise).

#### Pipeline operations

[Pipeline steps](#pipeline-steps) are defined as [`Operation`s](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/enterprise/dev/ci/internal/ci/operations/operations.go) that apply changes to the given pipeline, such as adding steps and components.

```sgquery
(:[_] *bk.Pipeline) patternType:structural repo:^github\.com/sourcegraph/sourcegraph$ file:^enterprise/dev/ci/internal/ci/operations\.go
```

Within an `Operation` you will typically create one or more steps on a pipeline with `AddStep`, which can be configured with options of type `SteptOpt`.

```sqquery
(:[_]) StepOpt patternType:structural repo:^github\.com/sourcegraph/sourcegraph$ file:^enterprise/dev/ci/internal/buildkite/buildkite\.go
```

Operations are then added to a pipeline from [`GeneratePipeline`](https://sourcegraph.com/search?q=context:global+repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+file:%5Eenterprise/dev/ci/internal/ci/pipeline%5C.go+GeneratePipeline&patternType=literal).

For most basic PR checks, see [Developing PR checks](#developing-pr-checks) for how to create your own steps!

For more advanced usage for specific run types, see [Developing run types](#developing-run-types).

For caching artefacts to speed up builds, see [How to cache CI artefacts](../how-to/cache_ci_artefacts.md).

#### Developing PR checks

To create a new check that can run on pull requests on relevant files, check the [`changed.Files`](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/enterprise/dev/ci/internal/ci/changed/changed.go) type to see if a relevant `affectsXyz` check already exists.

```sgquery
Affects type:symbol select:symbol.function repo:^github\.com/sourcegraph/sourcegraph$ file:^enterprise/dev/ci/internal/ci/changed
```

If not, you can define a new one on the `changed.Files` type.

Then, you can add a new check to [`CoreTestOperations`](https://sourcegraph.com/search?q=context:global+repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+file:%5Eenterprise/dev/ci/internal/ci+CoreTestOperations+type:symbol+&patternType=literal).
Make sure to follow the best practices outlined in docstring.

For more advanced pipelines, see [Run types](#run-types).

#### Developing run types

There are a variety of run types available based on branch prefixes. These generate special-purpose pipelines. For example, the `main-dry-run/` prefix is used to generate a pipeline similar to the default `main` branch. See [`RunType`](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/enterprise/dev/ci/internal/ci/runtype.go) for the various run types available, and examples for how to add more.

[`GeneratePipeline`](https://sourcegraph.com/search?q=context:global+repo:%5Egithub%5C.com/sourcegraph/sourcegraph%24+file:%5Eenterprise/dev/ci/internal/ci/pipeline%5C.go+GeneratePipeline&patternType=literal) can leverage `RunType` when generating pipelines, for example:

```sgquery
RunType.Is(:[_]) OR case :[_]: patternType:structural repo:^github\.com/sourcegraph/sourcegraph$ file:^enterprise/dev/ci/internal/ci/pipeline\.go
```

For simple PR checks, see [Creating PR checks](#creating-pr-checks).

#### Buildkite infrastructure

Also see [Flaky infrastructure](#flaky-infrastructure), [Continous integration infrastructure](https://handbook.sourcegraph.com/departments/product-engineering/engineering/tools/infrastructure/ci), and the [Continuous integration changelog](https://handbook.sourcegraph.com/departments/product-engineering/engineering/tools/infrastructure/ci/changelog).

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

### `buildchecker`

[`buildchecker`](https://github.com/sourcegraph/sourcegraph/actions/workflows/buildchecker.yml), our [branch lock management tool](#branch-locks), runs in GitHub actions - see the [workflow specification](https://github.com/sourcegraph/sourcegraph/blob/main/.github/workflows/buildchecker.yml).

To learn more about `buildchecker`, refer to the [`buildchecker` source code and documentation](https://github.com/sourcegraph/sourcegraph/tree/main/dev/buildchecker).

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
