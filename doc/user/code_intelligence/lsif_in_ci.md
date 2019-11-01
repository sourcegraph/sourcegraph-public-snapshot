# LSIF in continuous integration

*(Documentation Incomplete)*

## Setting up LSIF code intelligence

Install the LSIF indexer for your language (e.g. Go):

```
$ go get github.com/sourcegraph/lsif-go/cmd/lsif-go
```

Generate `data.lsif` in your project root (most LSIF indexers require a proper build environment: dependencies have been fetched, environment variables are set, etc.):

```
some-project-dir$ lsif-go --noContents --out=data.lsif
```

Then, upload `data.lsif` to your Sourcegraph instance via the [Sourcegraph CLI (`src`)](https://github.com/sourcegraph/src-cli):

```
some-project-dir$ src \
  -endpoint=https://sourcegraph.example.com \
  lsif upload \
  -repo=github.com/<user>/<reponame> \
  -commit=$(git rev-parse HEAD | tr -d "\n") \
  -file=data.lsif
```

> If you're uploading to Sourcegraph.com, you must authenticate your upload by passing a GitHub access token with [`public_repo` scope](https://developer.github.com/apps/building-oauth-apps/understanding-scopes-for-oauth-apps/#available-scopes). You can create one at https://github.com/settings/tokens

If successful, you'll see the following message:

> Upload successful, queued for processing.

If an error occurred, you'll see it in the response.

## Recommended setup

Start with a periodic job (e.g. daily) in CI that generates and uploads LSIF data on the default branch for your repository.

If you're noticing a lot of stale code intel between LSIF uploads or your CI doesn't support periodic jobs, you can set up a CI job that runs on every commit (including branches). The downsides to this are: more load on CI, more load on your Sourcegraph instance, and more rapid decrease in free disk space on your Sourcegraph instance.
