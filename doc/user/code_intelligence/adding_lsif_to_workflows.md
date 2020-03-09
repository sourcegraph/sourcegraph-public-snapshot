# Adding LSIF to your workflows

## LSIF in continuous integration

After walking through the [LSIF quickstart guide](./lsif_quickstart.md), add a job to your CI so code intelligence keeps up with the changes to your repository.

### Generating and uploading LSIF in CI

#### Set up your CI machines

Your CI machines will need two command-line tools installed. Depending on your build system setup, you can do this as part of the CI step, or you can add it directly to your CI machines for use by the build.

1. The [Sourcegraph CLI (`src`)](https://github.com/sourcegraph/src-cli).
1. The [LSIF indexer](https://lsif.dev) for your language.

#### Add steps to your CI

1. **Generate the LSIF file** for a project within your repository by running the LSIF indexer in the project directory (see docs for your LSIF indexer).
1. **[Upload that generated LSIF file](./lsif_quickstart.md#upload-the-data)** to your Sourcegraph instance.

### Recommended upload frequency

Start with a periodic job (e.g. daily) in CI that generates and uploads LSIF data on the default branch for your repository.

If you're noticing a lot of stale code intel between LSIF uploads or your CI doesn't support periodic jobs, you can set up a CI job that runs on every commit (including branches). The downsides to this are: more load on CI, more load on your Sourcegraph instance, and more rapid decrease in free disk space on your Sourcegraph instance.

## LSIF on GitHub

You can use [GitHub Actions](https://help.github.com/en/github/automating-your-workflow-with-github-actions/about-github-actions) to (1) generate LSIF data and (2) upload it to your Sourcegraph instance.

1. Actions to **Generate LSIF index data** for each language:

    - [Go indexer action](https://github.com/marketplace/actions/sourcegraph-go-lsif-indexer)
    - ...and more coming soon!

2. Action to **[upload LSIF data](https://github.com/marketplace/actions/sourcegraph-lsif-uploader)**.

### Setup

Create a [workflow file](https://help.github.com/en/github/automating-your-workflow-with-github-actions/configuring-a-workflow#creating-a-workflow-file) `.github/workflows/lsif.yaml` in your repository.

You will need configure two actions to (1) generate the LSIF data and (2) upload it to Sourcegraph. Here's an example for generating LSIF data for a Go project:

```yaml
name: LSIF
on:
  - push
jobs:
  index:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v1
      - name: Generate LSIF data
        uses: sourcegraph/lsif-go-action@master
      - name: Upload LSIF data
        uses: sourcegraph/lsif-upload-action@master
        with:
          github_token: ${{ secrets.GITHUB_TOKEN }}
```

Depending on your code requirements, you may need to add additional steps after checkout but before indexing (installing dependencies, generating indexable assets, etc).

Once that workflow is committed to your repository, you will start to see LSIF workflows in the Actions tab of your repository (e.g. https://github.com/sourcegraph/sourcegraph/actions).

![img/workflow.png](img/workflow.png)

After the workflow succeeds, you should see LSIF-powered code intelligence on your repository on Sourcegraph.com or on GitHub with the [Sourcegraph browser extension](../../integration/browser_extension.md).
