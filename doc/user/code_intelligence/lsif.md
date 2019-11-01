# LSIF

[LSIF](https://github.com/Microsoft/language-server-protocol/blob/master/indexFormat/specification.md) is a file format for precomputed code intelligence data. It provides fast and precise code intelligence, but needs to be periodically generated and uploaded to your Sourcegraph instance. LSIF is opt-in: repositories for which you have not uploaded LSIF data will continue to use the out-of-the-box code intelligence.

> LSIF is supported in Sourcegraph 3.8 and up.

> For users who have a language server deployed, LSIF will take priority over the language server when LSIF data exists for a repository.

## LSIF indexers

An LSIF indexer is a command line tool that analyzes your project's source code and generates a file in LSIF format containing all the definitions, references, and hover documentation in your project. That LSIF file is later uploaded to Sourcegraph to provide code intelligence.

Several languages are currently supported:

- [TypeScript](https://github.com/Microsoft/lsif-node/tree/master/tsc)
- [Go](https://github.com/sourcegraph/lsif-go)
- [C/C++](https://github.com/sourcegraph/lsif-cpp)
- [Python](https://github.com/sourcegraph/lsif-py), [Java](https://github.com/sourcegraph/lsif-java), and [OCaml](https://github.com/sourcegraph/merlin-to-coif) are early stage
- LSIF indexers for more languages coming soon! See https://lsif.dev for more information.

## Setting up LSIF code intelligence

Install the LSIF indexer for your language (e.g. Go):

```bash
go get github.com/sourcegraph/lsif-go/cmd/lsif-go
```

Generate `data.lsif` in your project root (most LSIF indexers require a proper build environment: dependencies have been fetched, environment variables are set, etc.):

```bash
some-project-dir$ lsif-go --noContents --out=data.lsif
```

Then, upload `data.lsif` to your Sourcegraph instance via the [Sourcegraph CLI (`src`)](https://github.com/sourcegraph/src-cli):

```bash
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

Go to your global settings at `/site-admin/global-settings` and enable LSIF:

```json
{
  "codeIntel.lsif": true
}
```

After uploading LSIF files, your Sourcegraph instance will use these files to power code intelligence so that when you visit a file in that repository on your Sourcegraph instance, the code intelligence should be more precise than it was out-of-the-box.

When LSIF data does not exist for a particular file in a repository, Sourcegraph will fall back to out-of-the-box code intelligence.

## Recommended setup

Start with a periodic job (e.g. daily) in CI that generates and uploads LSIF data on the default branch for your repository.

If you're noticing a lot of stale code intel between LSIF uploads or your CI doesn't support periodic jobs, you can set up a CI job that runs on every commit (including branches). The downsides to this are: more load on CI, more load on your Sourcegraph instance, and more rapid decrease in free disk space on your Sourcegraph instance.

## Stale code intelligence

LSIF code intelligence will be out-of-sync when you're viewing a file that has changed since the LSIF data was uploaded.

## Warning about uploading too much data

Global find-references is a resource-intensive operation that's sensitive to the number of packages for which you have uploaded LSIF data into your Sourcegraph instance. Improvements to this are planned for Sourcegraph 3.10 (see the [RFC](https://docs.google.com/document/d/1VZB0Y4tWKeOUN1JvdDgo4LHwQn875MPOI9xztzqoSRc/edit#)).

**Do not upload more than 10-40 LSIF dumps to Sourcegraph instance or you risk harming other parts of Sourcegraph. We are working to validate its performance at scale and eliminate this concern.**

## More about LSIF

To learn more, check out our lightning talk about LSIF from GopherCon 2019 or the [introductory blog post](https://about.sourcegraph.com/blog/code-intelligence-with-lsif):

<iframe width="560" height="315" src="https://www.youtube.com/embed/fMIRKRj_A88" frameborder="0" allow="accelerometer; autoplay; encrypted-media; gyroscope; picture-in-picture" allowfullscreen></iframe>
