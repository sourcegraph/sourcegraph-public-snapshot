# LSIF in continuous integration

## LSIF indexers

An LSIF indexer is a command line tool that analyzes your project's source code and generates a file in LSIF format containing all the definitions, references, and hover documentation in your project. That LSIF file is later uploaded to Sourcegraph to provide code intelligence.

## Generating and uploading LSIF in CI

1. Install the [Sourcegraph CLI (`src`)](https://github.com/sourcegraph/src-cli) for uploading LSIF data on your CI machines.
1. Go to https://lsif.dev to find an LSIF indexer for your language and install the command-line tool on your CI machines.
1. Add a daily step to your CI that:
  - Runs the LSIF indexer on a project within your repository and generates an LSIF file in the project directory (see docs for your LSIF indexer).
  - Uploads that generated LSIF file to your Sourcegraph instance using the following command:

Make sure the current working directory is somewhere inside your repository, then run:

```
$ src \
  -endpoint=https://sourcegraph.example.com \
  lsif upload \
  -github-token=abc... (only needed when uploading to Sourcegraph.com) \
  -file=<LSIF file (e.g. ./cmd/dump.lsif)>
```

> If you're uploading to Sourcegraph.com, you must authenticate your upload by passing a GitHub access token with [`public_repo` scope](https://developer.github.com/apps/building-oauth-apps/understanding-scopes-for-oauth-apps/#available-scopes) as `-github-token=abc...`. You can create one at https://github.com/settings/tokens

If successful, you'll see the following message:

> Upload successful, queued for processing.

If an error occurred, you'll see it in the response.

## Recommended upload frequency

Start with a periodic job (e.g. daily) in CI that generates and uploads LSIF data on the default branch for your repository.

If you're noticing a lot of stale code intel between LSIF uploads or your CI doesn't support periodic jobs, you can set up a CI job that runs on every commit (including branches). The downsides to this are: more load on CI, more load on your Sourcegraph instance, and more rapid decrease in free disk space on your Sourcegraph instance.
