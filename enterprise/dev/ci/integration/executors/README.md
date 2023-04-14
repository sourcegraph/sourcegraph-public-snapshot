# Executor server-side batch changes E2E test pipeline

This CI pipeline tests the executor and SSBC features end to end to detect breakages. Currently, it uses a single docker sourcegraph instance to talk to it.

## Running the test pipeline locally

You need docker-compose installed.

Currently, you need to provide an image tag of an image that has already been built in CI. We need to allow to use local dev builds here, too.
For that, grab the current version tag from a pipeline, and expose it as `CANDIDATE_VERSION`. This is usually of the format `213200_2023-04-14_5.0-c06b1549d298`.

Other required env vars:

`GITHUB_TOKEN`: A GitHub PAT that allows to clone the github.com/sourcegraph/automation-testing repository (is public).

```bash
$ GITHUB_TOKEN=<token> CANDIDATE_VERSION=<version> ./enterprise/dev/ci/integration/executors/run.sh
```

The first run will be a bit slow because it has to pull some images, but consecutive ones are reasonable fast.
