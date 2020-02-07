# LSIF: Fast and precise code intelligence

## Background

To efficiently answer precise code intelligence queries such as Go-to-Definition and Find References, it is usually necessary to perform an indexing step. Resolving the target of a referring expression in code generally requires knowledge of the dependencies of the code. That, in turn, requires information about the build and/or deployment environment of the code.

Historically, sourcegraph used [language servers](https://microsoft.github.io/language-server-protocol/implementors/servers) to answer these queries, but the performance and operational costs of maintaining a fleet of such services per repository were deemed unacceptable. Instead, we now promote a separate offline indexing strategy: For each language to be indexed, run a standalone language-specific precise code indexer program over each designated repository. These indexers emit data in the [Language Server Index Format (LSIF)](https://code.visualstudio.com/blogs/2019/02/19/lsif), and those data are captured (uploaded) to persistent storage.

[LSIF](https://github.com/Microsoft/language-server-protocol/blob/master/indexFormat/specification.md) is a file format for precomputed code intelligence data. It allows precise code intelligence queries to be answered quickly, but needs to be periodically (re-)generated and uploaded to your Sourcegraph instance. Sourcegraph can then use the stored index data to answer precise code intelligence queries without the need to run, maintain, and call out to a separate language server process.  Precise code intelligence is currently opt-in: repositories for which you have not uploaded LSIF data will continue to use the built-in code intelligence.

To automate building and updating code intelligence data, we advocate running precise code indexers (colloquially: “LSIF indexers”) as part of testing, continuous integration/deployment, or other code-host automation tasks (using, for example, GitHub Actions, Travis CI, Bitbucket Pipelines, etc.). This approach ensures the indexers run in the same (or similar) environment as a production build or deployment, so that the correct tools and dependencies are visible to the indexer. It also gives the customer control over what code to index, and how often.

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

## Why are my results sometimes incorrect?

You may occasionally see results from [basic code intelligence](basic_code_intelligence.md) even when you have uploaded LSIF data. Such results are indicated with a ![tooltip](img/basic-code-intel-tooltip.svg) tooltip. This can happen in the following scenarios:

- The symbol has LSIF data, but it is defined in a repository which does not have LSIF data.
- The nearest commit that has LSIF data is too far away from your browsing commit. [The limit is 100 commits](https://github.com/sourcegraph/sourcegraph/blob/e7803474dbac8021e93ae2af930269045aece079/lsif/src/shared/constants.ts#L25) ahead/behind.
- The current file doesn't exist in the nearest LSIF dump or has been changed between the LSIF dump and the browsing commit.
- The _Find references_ panel will always include search-based results, but only after all of the precise results have been displayed. This ensures every symbol has code intelligence.

## Data retention policy

The bulk of LSIF data is stored on-disk, and as code intelligence data for a commit ages it becomes less useful. Sourcegraph will automatically remove the least recently uploaded data if the amount of disk space falls above a threshold. This value can be changed via the `DBS_DIR_MAXIMUM_SIZE_BYTES` environment variable. The default value of this variable is `10737418240`, which is `1024 * 1024 * 1024 * 10` bytes, or `10` gigabytes.

## Warning about uploading too much data

Global find-references is a resource-intensive operation that's sensitive to the number of packages for which you have uploaded LSIF data into your Sourcegraph instance. Improvements to this are planned for Sourcegraph 3.10 (see the [RFC](https://docs.google.com/document/d/1VZB0Y4tWKeOUN1JvdDgo4LHwQn875MPOI9xztzqoSRc/edit#)).

**Do not upload more than 10-40 LSIF dumps to your Sourcegraph instance or you risk harming other parts of Sourcegraph. We are working to validate its performance at scale and eliminate this concern.**

The following table gives a rough estimate for the space and time requirements for indexing and conversion. These repositories are a representative sample of public Go repositories available on GitHub. The working tree size is the size of the clone at the given commit (without git history), the number of files indexed, and the number of lines of Go code in the repository. The index size gives the size of the uncompressed LSIF output of the indexer. The conversion size gives the total amount of disk space occupied after uploading the dump to a Sourcegraph instance.

| Repository | Working tree size | Index time | Index size | Processing time | Post-processing size |
| ------------------------------------------------------------------- | ------------------------------- | ------ | ----- | ------- | ----- |
| [bigcache](https://github.com/allegro/bigcache/tree/b7689f7)        | 216KB,   32 files,   2.585k loc |  1.18s | 3.5MB |   0.45s | 0.6MB |
| [sqlc](https://github.com/kyleconroy/sqlc/tree/16cc4e9)             | 396KB,   24 files,   7.041k loc |  1.53s | 7.2MB |   1.62s | 1.6MB |
| [nebula](https://github.com/slackhq/nebula/tree/a680ac2)            | 700KB,   71 files,  10.704k loc |  2.48s |  16MB |   1.63s | 2.9MB |
| [cayley](https://github.com/cayleygraph/cayley/tree/4d89b8a)        | 5.6MB,  226 files,  36.346k loc |  5.58s |  51MB |   4.68s |  11MB |
| [go-ethereum](https://github.com/ethereum/go-ethereum/tree/275cd49) |  27MB,  945 files, 317.664k loc | 20.53s | 255MB |  77.40s |  50MB |
| [kubernetes](https://github.com/kubernetes/kubernetes/tree/e680ad7) | 301MB, 4577 files,   1.550m loc |  1.21m | 910MB |  80.06s | 162MB |
| [aws-sdk-go](https://github.com/aws/aws-sdk-go/tree/18a2d30)        | 119MB, 1759 files,   1.067m loc |  8.20m | 1.3GB | 155.82s | 358MB |






## Cross-repository code intelligence

Cross-repository code intelligence will only be powered by LSIF when **both** repositories have LSIF data. When the current file has LSIF data and the other repository doesn't, there will be no code intelligence results (we're working on fallback to fuzzy code intelligence for 3.10).

## More about LSIF

To learn more, check out our lightning talk about LSIF from GopherCon 2019 or the [introductory blog post](https://about.sourcegraph.com/blog/code-intelligence-with-lsif):

<iframe width="560" height="315" src="https://www.youtube.com/embed/fMIRKRj_A88" frameborder="0" allow="accelerometer; autoplay; encrypted-media; gyroscope; picture-in-picture" allowfullscreen></iframe>
