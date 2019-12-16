# LSIF: Fast and precise code intelligence

[LSIF](https://github.com/Microsoft/language-server-protocol/blob/master/indexFormat/specification.md) is a file format for precomputed code intelligence data. It provides fast and precise code intelligence but needs to be periodically generated and uploaded to your Sourcegraph instance. LSIF is opt-in: repositories for which you have not uploaded LSIF data will continue to use the built-in code intelligence.

> Precise code intelligence using LSIF is supported in Sourcegraph 3.8 and up.

> For users who have a language server deployed, LSIF will take priority over the language server when LSIF data exists for a repository.

## Getting started

Follow our [LSIF quickstart guide](lsif_quickstart.md) to manually generate and upload LSIF data for your repository. After you are satisfied with the result, you can upload LSIF data to a Sourcegraph instance using your existing [continuous integration infrastructure](lsif_in_ci.md), or using [GitHub Actions](lsif_on_github.md).

## Enabling LSIF on your Sourcegraph instance

Go to your global settings at https://sourcegraph.example.com/site-admin/global-settings and enable LSIF:

```json
  "codeIntel.lsif": true
```

After uploading LSIF files, your Sourcegraph instance will use these files to respond to code intelligence requests (such as for hovers, definitions, and references). When LSIF data does not exist for a particular file in a repository, Sourcegraph will fall back to built-in code intelligence.

## Stale code intelligence

LSIF code intelligence will be out of sync when you're viewing a file that has changed since the LSIF data was uploaded.

## Data retention policy

The bulk of LSIF data is stored on-disk, and as code intelligence data for a commit ages it becomes less useful. Sourcegraph will automatically remove the least recently uploaded data if the amount of disk space falls above a threshold. This value can be changed via the `DBS_DIR_MAXIMUM_SIZE_BYTES` environment variable. The default value of this variable is `10737418240`, which is `1024 * 1024 * 1024 * 10` bytes, or `10` gigabytes.

## Warning about uploading too much data

Global find-references is a resource-intensive operation that's sensitive to the number of packages for which you have uploaded LSIF data into your Sourcegraph instance. Improvements to this are planned for Sourcegraph 3.10 (see the [RFC](https://docs.google.com/document/d/1VZB0Y4tWKeOUN1JvdDgo4LHwQn875MPOI9xztzqoSRc/edit#)).

**Do not upload more than 10-40 LSIF dumps to your Sourcegraph instance or you risk harming other parts of Sourcegraph. We are working to validate its performance at scale and eliminate this concern.**

The following table gives a rough estimate for the space and time requirements for indexing and conversion. These repositories are a representative sample of public Go repositories available on GitHub. The clone size is the total size of source files (without history) of the clone at the given commit. The index size gives the size of the uncompressed LSIF output of the indexer, and the conversion size gives the total amount of disk space occupied after uploading the dump to a Sourcegraph instance.

| Repository | Working tree size | Index time | Index size | Processing time | Post-processing size |
| ---------------------------------------------------------------------------------------------------- | ----------------- | ------ | ---- | ------- | ------ |
| [bigcache](https://github.com/allegro/bigcache/tree/b7689f7c33374d4c67c011eaa0a5b345ddb1a99c)        | 216KB   (32 files) |  1.18s | 3.5MB |   0.45s | 0.564MB |
| [sqlc](https://github.com/kyleconroy/sqlc/tree/16cc4e9c378341b5496af784b25422d1ed4c7fd9)             | 396K   (24 files) |  1.53s | 7.2M |   1.62s | 1.6M   |
| [nebula](https://github.com/slackhq/nebula/tree/a680ac29f5b7ce13d4007d090776e983cd3c1e76)            | 700K   (71 files) |  2.48s | 16M  |   1.63s | 2.9M   |
| [cayley](https://github.com/cayleygraph/cayley/tree/4d89b8a1806203c5c09e16bfc405bc3d64d74236)        | 5.6M  (226 files) |  5.58s | 51M  |   4.68s | 11M    |
| [go-ethereum](https://github.com/ethereum/go-ethereum/tree/275cd4988dbef4b81e856a6c6ae8cb12242e3976) | 27M   (945 files) | 20.53s | 255M |  77.40s | 50M    |
| [kubernetes](https://github.com/kubernetes/kubernetes/tree/e680ad7156f263a6d8129cc0117fda58602e50ad) | 301M (4577 files) | 34.81m | 910M |  80.06s | 162M   |
| [aws-sdk-go](https://github.com/aws/aws-sdk-go/tree/18a2d30ffcef68a1d1bed6a4a9cd6b34bfac049a)        | 119M (1759 files) |  8.20m | 1.3G | 155.82s | 358M   |


## Cross-repository code intelligence

Cross-repository code intelligence will only be powered by LSIF when **both** repositories have LSIF data. When the current file has LSIF data and the other repository doesn't, there will be no code intelligence results (we're working on fallback to fuzzy code intelligence for 3.10).

## More about LSIF

To learn more, check out our lightning talk about LSIF from GopherCon 2019 or the [introductory blog post](https://about.sourcegraph.com/blog/code-intelligence-with-lsif):

<iframe width="560" height="315" src="https://www.youtube.com/embed/fMIRKRj_A88" frameborder="0" allow="accelerometer; autoplay; encrypted-media; gyroscope; picture-in-picture" allowfullscreen></iframe>
