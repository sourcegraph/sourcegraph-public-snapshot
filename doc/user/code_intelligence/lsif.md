# LSIF

LSIF is a file format for precomputed code intelligence data. It's fast and precise, but needs to be periodically generated and uploaded to your Sourcegraph instance.

> LSIF is supported in Sourcegraph 3.8 and up.

## Language support

- [TypeScript](https://github.com/Microsoft/lsif-node/tree/master/tsc)
- [Go](https://github.com/sourcegraph/lsif-go)
- [Python](https://github.com/sourcegraph/lsif-py), [C/C++](https://github.com/sourcegraph/lsif-cpp), and [OCaml](https://github.com/sourcegraph/merlin-to-coif) are early stage
- LSIF indexers for more languages coming soon!

## Setting up LSIF code intelligence

TODO overview of how indexers work (the general pipeline of generating + moniker generation)

Install the LSIF indexer for your language (e.g. Go):

```
$ go get github.com/sourcegraph/lsif-go/cmd/lsif-go
```

Generate `data.lsif` in your project root:

TODO what's required in the CI environment

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
  COMMIT=<40 char hash> \
  bash upload.sh data.lsif
```

If the upload is accepted, the response will be `null`.

Now, when you visit a file in that repository on your Sourcegraph instance, the code intelligence should be more precise than it was out-of-the-box.

## Troubleshooting

When uploading:

- `REPOSITORY` must match the name of the repository on your Sourcegraph instance
- `COMMIT` must be the full 40 character hash
- Set http or https to whichever you use when visiting your Sourcegraph instance in your browser

## Stale code intelligence

LSIF code intelligence will be out-of-sync when you're viewing a file that has changed since the LSIF data was uploaded. You can fix this by setting up a periodic job that generates and uploads LSIF for the tip of your default branch (e.g. master) daily. Improvements to this are slated for Sourcegraph 3.9.

TODO mention how basic code intel falls back

TODO recommend against mixing LSIF and LSP

## More about LSIF

To learn more, check out our lightning talk about LSIF from GopherCon 2019 or the [introductory blog post](https://about.sourcegraph.com/blog/code-intelligence-with-lsif):

<iframe width="560" height="315" src="https://www.youtube.com/embed/fMIRKRj_A88" frameborder="0" allow="accelerometer; autoplay; encrypted-media; gyroscope; picture-in-picture" allowfullscreen></iframe>

TODO comparison with LSP
