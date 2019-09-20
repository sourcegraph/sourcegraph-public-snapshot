# LSIF

[LSIF](https://github.com/Microsoft/language-server-protocol/blob/master/indexFormat/specification.md) is a file format for precomputed code intelligence data. It provides fast and precise code intelligence, but needs to be periodically generated and uploaded to your Sourcegraph instance. LSIF is opt-in: repositories for which you have not uploaded LSIF data will continue to use the out-of-the-box code intelligence.

> LSIF is supported in Sourcegraph 3.8 and up.

> For users who have a language server deployed, LSIF will take priority over the language server when LSIF data exists for a repository.

## LSIF indexers

An LSIF indexer is a command line tool that analyzes your project's source code and generates a file in LSIF format containing all the definitions, references, and hover documentation in your project. That LSIF file is later uploaded to Sourcegraph to provide code intelligence.

Several languages are currently supported:

- [TypeScript](https://github.com/Microsoft/lsif-node/tree/master/tsc)
- [Go](https://github.com/sourcegraph/lsif-go)
- [Python](https://github.com/sourcegraph/lsif-py), [C/C++](https://github.com/sourcegraph/lsif-cpp), and [OCaml](https://github.com/sourcegraph/merlin-to-coif) are early stage
- LSIF indexers for more languages coming soon!

## Setting up LSIF code intelligence

Install the LSIF indexer for your language (e.g. Go):

```
$ go get github.com/sourcegraph/lsif-go/cmd/lsif-go
```

Generate `data.lsif` in your project root (most LSIF indexers require a proper build environment: dependencies have been fetched, environment variables are set, etc.):

```
some-project-dir$ lsif-go --noContents --out=data.lsif
```

Get the LSIF bash uploader:

```
some-project-dir$ curl -O https://raw.githubusercontent.com/sourcegraph/sourcegraph/master/lsif/upload.sh
```

Upload `data.lsif` to your Sourcegraph instance:

```
some-project-dir$ env \
  SRC_ENDPOINT=https://sourcegraph.example.com \
  REPOSITORY=github.com/<user>/<reponame> \
  COMMIT=$(git rev-parse HEAD | tr -d "\n") \
  bash upload.sh data.lsif
```

- `SRC_ENDPOINT` is the URL to your Sourcegraph instance
- `REPOSITORY` must match the name of the repository on your Sourcegraph instance
- `COMMIT` must be the full 40 character hash

If the upload is accepted, the response will be `null`. If an error occurred, you'll see it in the response.

After uploading LSIF files, your Sourcegraph instance will use these files to power code intelligence so that when you visit a file in that repository on your Sourcegraph instance, the code intelligence should be more precise than it was out-of-the-box.

When LSIF data does not exist for a particular file in a repository, Sourcegraph will fall back to out-of-the-box code intelligence.

## Stale code intelligence

LSIF code intelligence will be out-of-sync when you're viewing a file that has changed since the LSIF data was uploaded. You can mitigate this by setting up a periodic job that generates and uploads LSIF for the tip of your default branch (e.g. master) daily. Improvements to this are planned for Sourcegraph 3.9.

## More about LSIF

To learn more, check out our lightning talk about LSIF from GopherCon 2019 or the [introductory blog post](https://about.sourcegraph.com/blog/code-intelligence-with-lsif):

<iframe width="560" height="315" src="https://www.youtube.com/embed/fMIRKRj_A88" frameborder="0" allow="accelerometer; autoplay; encrypted-media; gyroscope; picture-in-picture" allowfullscreen></iframe>
