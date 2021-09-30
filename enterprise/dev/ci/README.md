# Buildkite Pipeline for sourcegraph/sourcegraph

We dynamically generate our CI pipeline for https://buildkite.com/sourcegraph/sourcegraph based on the output of [gen-pipeline.go](./gen-pipeline.go).

## Setup

In the [pipeline settings](https://buildkite.com/sourcegraph/sourcegraph/settings) ensure there is the following step:

```shell
go run ./enterprise/dev/ci/gen-pipeline.go | buildkite-agent pipeline upload
```

## Testing

To test this you can run `env BUILDKITE_BRANCH=TESTBRANCH go run ./enterprise/dev/ci/gen-pipeline.go` and inspect the YAML output. To change the behaviour set the relevant `BUILDKITE_` environment variables.

## Flaky Tests

Use language specific functionality to skip a test. If the language allows for a skip reason, include a link to track re-enabling the test.

- Go :: [testing.T.Skip](https://pkg.go.dev/testing#hdr-Skipping).
- Typescript :: [.skip()](https://mochajs.org/#inclusive-tests)

Ping an owner about the skipping (normally on the PR skipping it).

## Flaky Step

If a step is flaky we need to get the build back to reliable as soon as possible. If there is not already a discussion in `#buildkite-main` create one and link what step you take. Here are the recommended approaches in order:

1. Revert the PR if a recent change introduced the instability. Ping author.
2. Use `Skip` StepOpt when creating the step. Include reason and a link to context. This will still show the step on builds so we don't forget about it.
3. Use `SoftFail` StepOpt. This will still run the step, but won't block the build. Note: we don't yet have a convenient way to collect reliability information on a step.

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
