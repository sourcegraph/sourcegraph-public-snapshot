# Adding LSIF to your workflows

## [Language-specific guides](./languages/index.md)

We are working on creating guides for each language with an LSIF indexer, so make sure to check for the documentation for your language! If there is not a guide for your language, this general guide will help you through the LSIF setup process.

## LSIF in continuous integration

After walking through the [LSIF quickstart guide](./lsif_quickstart.md), add a job to your CI so code intelligence keeps up with the changes to your repository. Because of how many CI frameworks and languages there are, we may not have documented any specific advice for your use case. Feel free to [file an issue](https://github.com/sourcegraph/sourcegraph/issues/new) if you're having difficulties and we can help troubleshoot your setup.

## Index

- [Using indexer containers](#using-indexer-containers)
- [CI from scratch](#ci-from-scratch)
- [Recommended upload frequency](#recommended-upload-frequency)
- [Uploading LSIF data to Sourcegraph.com](#uploading-lsif-data-to-sourcegraphcom)

## Using indexer containers

We're working on creating containers that bundle indexers and the [Sourcegraph CLI (`src`)](https://github.com/sourcegraph/src-cli). All the languages in our [language-specific guides](languages/index.md) are supported, and this section provides a general guide to using those containers. We've pinned the containers to the `latest` tag in these samples so they stay up to date, but make sure you pin them before going live to ensure reproducibility.

### Basic usage

Some indexers will work out of the box by just running them in your project root. Here's an example using GitHub actions and Go:

```yaml
jobs:
  lsif-go:
    # this line will prevent forks of this repo from uploading lsif indexes
    if: github.repository == '<insert your repo name>'
    runs-on: ubuntu-latest
    container: sourcegraph/lsif-go:latest
    steps:
      - uses: actions/checkout@v1
      - name: Generate LSIF data
        run: lsif-go
      - name: Upload LSIF data
        # this will upload to Sourcegraph.com, you may need to substitute a different command
        # by default, we ignore failures to avoid disrupting CI pipelines with non-critical errors.
        run: src lsif upload -github-token=${{ secrets.GITHUB_TOKEN }} -ignore-upload-failures
```

### Sub-projects

If your repository contains multiple code projects in different folders, your CI framework likely provides a "working directory" option so that you can do something like:
```yaml
jobs:
  lsif-go:
    # this line will prevent forks of this repo from uploading lsif indexes
    if: github.repository == '<insert your repo name>'
    runs-on: ubuntu-latest
    container: sourcegraph/lsif-go:latest
    steps:
      - uses: actions/checkout@v1
      - name: Generate LSIF data
        working-directory: backend/
        run: lsif-go
      - name: Upload LSIF data
        # note that the upload command also needs to happen in the same directory!
        working-directory: backend/
        # this will upload to Sourcegraph.com, you may need to substitute a different command
        # by default, we ignore failures to avoid disrupting CI pipelines with non-critical errors.
        run: src lsif upload -github-token=${{ secrets.GITHUB_TOKEN }} -ignore-upload-failures
```

### Custom build environments

Depending on which language you're using, you may need to perform some pre-indexing steps in a custom build environment. You have a couple options for this:

- Add the indexer and [Sourcegraph CLI (`src`)](https://github.com/sourcegraph/src-cli) to your environment. To make this easier, you could either directly build a Docker image from our indexer container, or use its Dockerfile as inspiration for getting the tools installed in your environment.
- Do the pre-indexing steps in your environment, and then make those artifacts available to our container.

This second step is easy in GitHub actions because our container can be used as an action. Here's an example for a TypeScript project:
```yaml
jobs:
  lsif-node:
    # this line will prevent forks of this repo from uploading lsif indexes
    if: github.repository == '<insert your repo name>'
    runs-on: ubuntu-latest
    container: my-awesome-container
    steps:
      - uses: actions/checkout@v1
      - name: Install dependencies
        run: <install dependencies>
      - name: Generate LSIF data
        uses: docker://sourcegraph/lsif-node:latest
        with:
          args: lsif-tsc -p .
      - name: Upload LSIF data
        uses: docker://sourcegraph/src-cli:latest
        with:
          # this will upload to Sourcegraph.com, you may need to substitute a different command
          # by default, we ignore failures to avoid disrupting CI pipelines with non-critical errors.
          args: src lsif upload -github-token=${{ secrets.GITHUB_TOKEN }} -ignore-upload-failures
```

Other frameworks may require you to explicitly cache artifacts between jobs. In CircleCI this might look like:
```yaml
jobs:
  install-deps:
    docker:
      - image: my-awesome-container
    steps:
      - checkout
      - <install dependencies>
      - save_cache:
          paths:
            - node_modules
          key: dependencies

jobs:
  lsif-node:
    docker:
      - image: sourcegraph/lsif-node:latest
    steps:
      - checkout
      - restore_cache:
          keys:
            - dependencies
      - run: lsif-tsc -p .
        # this will upload to Sourcegraph.com, you may need to substitute a different command
        # by default, we ignore failures to avoid disrupting CI pipelines with non-critical errors.
      - run: src lsif upload -github-token=<<parameters.github-token>> -ignore-upload-failures

workflows:
  lsif-node:
    jobs:
      - install-deps
      - lsif-node:
          requires:
            - install-deps
```

## CI from scratch

If you're indexing a language we haven't documented yet in our [language-specific guides](languages/index.md), follow the instructions in this section. We want to have containers available for every language with a robust indexer, so please also consider [filing an issue](https://github.com/sourcegraph/sourcegraph/issues/new) to let us know we're missing one.

### Set up your CI machines

Your CI machines will need two command-line tools installed. Depending on your build system setup, you can do this as part of the CI step, or you can add it directly to your CI machines for use by the build.

1. The [Sourcegraph CLI (`src`)](https://github.com/sourcegraph/src-cli).
1. The [LSIF indexer](https://lsif.dev) for your language.

### Add steps to your CI

1. **Generate the LSIF file** for a project within your repository by running the LSIF indexer in the project directory (see docs for your LSIF indexer).
1. **[Upload that generated LSIF file](./lsif_quickstart.md#upload-the-data)** to your Sourcegraph instance.

## Recommended upload frequency

We recommend that you start with a CI job that runs on every commit (including branches).

If you see too much load on your CI, your Sourcegraph instance, or a rapid decrease in free disk space on your Sourcegraph instance, you can instead index only the default branch, or set up a periodic job (e.g. daily) in CI that indexes your default branch.

With periodic jobs, you should still receive precise code intelligence on non-indexed commits on lines that are unchanged since the nearest indexed commit. This requires that the indexed commit be a direct ancestor or descendant no more than [100 commits](https://github.com/sourcegraph/sourcegraph/blob/e7803474dbac8021e93ae2af930269045aece079/lsif/src/shared/constants.ts#L25) away. If your commit frequency is too high and your index frequency is too low, you may find commits with no precise code intelligence at all. In this case, we recommend you try to increase your index frequency if possible.

## Uploading LSIF data to Sourcegraph.com

LSIF data can be uploaded to a self-hosted Sourcegraph instance or to [Sourcegraph.com](https://sourcegraph.com). Using the [Sourcegraph.com](https://sourcegraph.com) endpoint will surface code intelligence for your public repositories directly on GitHub via the [Sourcegraph browser extension](https://docs.sourcegraph.com/integration/browser_extension) and at `https://sourcegraph.com/github.com/<your-username>/<your-repo>`. 

Using the [Sourcegraph.com](https://sourcegraph.com) endpoint is free and your LSIF data is treated as User-Generated Content (you own it, as covered in our [Sourcegraph.com terms of service](https://about.sourcegraph.com/terms-dotcom#3-proprietary-rights-and-licenses)). If you run into trouble, or a situation arises where you need all of your LSIF data expunged, please reach out to us at [support@sourcegraph.com](mailto:support@sourcegraph.com).
