# LSIF on GitHub

You can use [GitHub Actions](https://help.github.com/en/github/automating-your-workflow-with-github-actions/about-github-actions) to (1) generate LSIF data and (2) upload it to your Sourcegraph instance.

1. Actions to **Generate LSIF index data** for each language:

    - [Go indexer action](https://github.com/marketplace/actions/sourcegraph-go-lsif-indexer)
    - ...and more coming soon!

2. Action to **[upload LSIF data](https://github.com/marketplace/actions/sourcegraph-lsif-uploader)**.

## Setup

Create a [workflow file](https://help.github.com/en/github/automating-your-workflow-with-github-actions/configuring-a-workflow#creating-a-workflow-file) `.github/workflows/lsif.yaml` in your repository.

You will need configure two actions to (1) generate the LSIF data and (2) upload it to Sourcegraph. Here's an example for generating LSIF data for a Go project:

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
```

Once that workflow is committed to your repository, you will start to see LSIF workflows in the Actions tab of your repository (e.g. https://github.com/sourcegraph/sourcegraph/actions).

![img/workflow.png](img/workflow.png)

After the workflow succeeds, you should see LSIF-powered code intelligence on your repository on Sourcegraph.com or on GitHub with the [Sourcegraph browser extension](../../integration/browser_extension.md).
