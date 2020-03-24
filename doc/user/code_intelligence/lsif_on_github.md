# LSIF on GitHub

You can use [GitHub Actions](https://help.github.com/en/github/automating-your-workflow-with-github-actions/about-github-actions) to (1) generate LSIF data and (2) upload it to a [Sourcegraph endpoint](#using-an-upload-endpoint).

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
        with:
          github_token: ${{ secrets.GITHUB_TOKEN }}
          endpoint: https://sourcegraph.com # use sourcegraph.com, or alternatively, your own instance
```

Once that workflow is committed to your repository, you will start to see LSIF workflows in the Actions tab of your repository (e.g. https://github.com/sourcegraph/sourcegraph/actions).

![img/workflow.png](img/workflow.png)

After the workflow succeeds, you should see LSIF-powered code intelligence on your repository on Sourcegraph.com or on GitHub with the [Sourcegraph browser extension](../../integration/browser_extension.md).

## Using an upload endpoint

LSIF data can be uploaded to a self-deployed Sourcegraph instance or to [sourcegraph.com](https://sourcegraph.com). Using the [sourcegraph.com](https://sourcegraph.com) endpoint will surface code intelligence for your public repositories directly on GitHub via the [Sourcegraph browser extension](https://docs.sourcegraph.com/integration/browser_extension) and at `https://sourcegraph.com/github.com/<your-username>/<your-repo>`. 

Using the [sourcegraph.com](https://sourcegraph.com) endpoint is free and your LSIF data is treated as User-Generated Content (you own it, as covered our [sourcegraph.com terms of service](https://about.sourcegraph.com/terms-dotcom#3-proprietary-rights-and-licenses)). If you run into trouble, or a situation arises where you need all of your LSIF data expunged, please reach out to us at [support@sourcegraph.com](mailto:support@sourcegraph.com).
