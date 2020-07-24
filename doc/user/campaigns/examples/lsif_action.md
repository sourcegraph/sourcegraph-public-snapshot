# Example: Adding a GitHub action to upload LSIF data to Sourcegraph

> NOTE: This documentation describes the current work-in-progress version of campaigns. [Click here](https://docs.sourcegraph.com/@3.18/user/campaigns) to read the documentation for campaigns in Sourcegraph 3.18.

<!-- TODO(sqs): update for new campaigns flow -->

Our goal for this campaign is to add a GitHub Action that generates and uploads LSIF data to Sourcegraph by adding a `.github/workflows/lsif.yml` file to each repository that doesn't have it yet.

The first thing we need is an action definition that we can execute with the [`src` CLI tool](https://github.com/sourcegraph/src-cli) and its `src actions exec` subcommand.

Here is an `action.json` file that runs a Docker container based on the Docker image called `add-lsif-to-build-pipeline-action` in each repository that has a `go.mod` file, `github` in its name and no `.github/workflows/lsif.yml` file:

```json
{
  "scopeQuery": "repohasfile:go.mod repo:github -repohasfile:.github/workflows/lsif.yml",
  "steps": [
    {
      "type": "docker",
      "image": "add-lsif-to-build-pipeline-action"
    }
  ]
}
```

Save that as `action.json`.

In order to build the Docker image, we first need to create a file called `github-action-workflow-golang.yml` with the following content:

```yaml
name: LSIF
on:
  - push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v1
      - name: Generate LSIF data
        uses: sourcegraph/lsif-go-action@master
        with:
          verbose: "true"
      - name: Upload LSIF data
        uses: sourcegraph/lsif-upload-action@master
        with:
          github_token: ${{ secrets.GITHUB_TOKEN }}
```

This is the definition of the GitHub action that we want to add to every repository returned by our `"scopeQuery"`.

Next we create the `Dockerfile`:

```Dockerfile
FROM alpine:3
ADD ./github-action-workflow-golang.yml /tmp/workflows/

CMD mkdir -p .github/workflows && \
  DEST=.github/workflows/lsif.yml; \
  if [ ! -f .github/workflows/lsif.yml ]; then \
    cp /tmp/workflows/github-action-workflow-golang.yml $DEST; \
  else \
    echo Doing nothing because existing LSIF workflow found at $DEST; \
  fi
```

Now we're ready to run the action and create the campaign:

1. Build the Docker image:

  ```
  docker build -t add-lsif-to-build-pipeline-action
  ```
1. Run the action and create a patch set:

  ```
  src actions exec -f action.json | src campaign patchset create-from-patches
  ```
1. Follow the printed instructions to create and run the campaign on Sourcegraph.
