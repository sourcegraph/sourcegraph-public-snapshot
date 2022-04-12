# Precise code intelligence

<style>
  .video-container {
    position: relative;
    padding-bottom: 56.25%; /* 16:9 */
    height: 0;
  }
  .video-container iframe {
    position: absolute;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
  }
</style>

Precise code intelligence relies on [LSIF](https://github.com/Microsoft/language-server-protocol/blob/master/indexFormat/specification.md)
(Language Server Index Format) data to deliver precomputed code intelligence. It provides fast and highly accurate code intelligence but needs to be periodically generated and uploaded to your Sourcegraph instance. Precise code intelligence is an opt-in feature: repositories for which you have not uploaded LSIF data will continue to use the search-based code intelligence.

> NOTE: Precise code intelligence using LSIF is supported in Sourcegraph 3.8 and up.

## Getting started

See the [how-to guides](../how-to/index.md) to get started with precise code intelligence.

## Cross-repository code intelligence

Cross-repository code intelligence works out-of-the-box when both the dependent repository and the dependency repository has LSIF data _at the correct commits or versions_. We are working on relaxing this constraint so that nearest-commit functionality works on a cross-repository basis as well.

When the current repository has LSIF data and a dependent doesn't, the missing precise results will be supplemented with imprecise search-based code intelligence. This also applies when both repositories have LSIF data, but for a different set of versions. For example, if repository A@v1 depends on B@v2, then we will get precise cross-repository intelligence when we have LSIF data for both A@v1 and B@v2, but would not get a precise result we instead have LISF data for A@v1 and B@v1.

## Why are my results sometimes incorrect?

If LSIF data is not found for a particular file in a repository, Sourcegraph will fall back to search-based code intelligence. You may occasionally see results from [search-based code intelligence](search_based_code_intelligence.md) even when you have uploaded LSIF data. This can happen in the following scenarios:

- The symbol has LSIF data, but it is defined in a repository which does not have LSIF data.
- The line containing the symbol was created or edited between the nearest indexed commit and the commit being browsed.
- The _Find references_ panel may include search-based results, but only after all of the precise results have been displayed. This ensures every symbol has useful code intelligence.

## More about LSIF

- [Writing an LSIF indexer](writing_an_indexer.md)
- [Adding LSIF to many repositories](../how-to/adding_lsif_to_many_repos.md)

To learn more, check out our lightning talk about LSIF from GopherCon 2019 or the [introductory blog post](https://about.sourcegraph.com/go/code-intelligence-with-lsif):

<div class="video-container">
  <iframe width="560" height="315" src="https://www.youtube.com/embed/fMIRKRj_A88" frameborder="0" allow="accelerometer; autoplay; encrypted-media; gyroscope; picture-in-picture" allowfullscreen></iframe>
</div>
