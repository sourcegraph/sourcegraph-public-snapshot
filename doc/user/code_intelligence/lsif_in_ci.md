# LSIF in continuous integration

After walking through the [LSIF quickstart guide](./lsif_quickstart.md), add a job to your CI so code intelligence keeps up with the changes to your repository.

## Generating and uploading LSIF in CI

### Setup your CI machines

Your CI machines will need two command-line tools installed. Depending on your build system setup, you can do this as part of the CI step, or you can add it directly to your CI machines for use by the build.

1. The [Sourcegraph CLI (`src`)](https://github.com/sourcegraph/src-cli).
1. The [LSIF indexer](https://lsif.dev) for your language.

### Add steps to your CI

1. **Generate the LSIF file** for a project within your repository by running the LSIF indexer in the project directory (see docs for your LSIF indexer).
1. **[Upload that generated LSIF file](./lsif_quickstart.md#upload-the-data)** to your Sourcegraph instance.

## Recommended upload frequency

Start with a periodic job (e.g. daily) in CI that generates and uploads LSIF data on the default branch for your repository.

If you're noticing a lot of stale code intel between LSIF uploads or your CI doesn't support periodic jobs, you can set up a CI job that runs on every commit (including branches). The downsides to this are: more load on CI, more load on your Sourcegraph instance, and more rapid decrease in free disk space on your Sourcegraph instance.
