# LSIF: Fast and precise code intelligence

[LSIF](https://github.com/Microsoft/language-server-protocol/blob/master/indexFormat/specification.md) is a file format for precomputed code intelligence data. It provides fast and precise code intelligence, but needs to be periodically generated and uploaded to your Sourcegraph instance. LSIF is opt-in: repositories for which you have not uploaded LSIF data will continue to use the out-of-the-box code intelligence.

> Precise code intelligence using LSIF is supported in Sourcegraph 3.8 and up.

> For users who have a language server deployed, LSIF will take priority over the language server when LSIF data exists for a repository.

## Getting started

Follow our [LSIF quickstart guide](lsif_quickstart.md) to manually generate and upload LSIF data for your repository. After you are satisfied with the result, you can upload LSIF data to a Sourcegraph instance using your existing [continuous integration infrastructure](lsif_in_ci.md), or using [GitHub Actions](lsif_on_github.md).

## Enabling LSIF on your Sourcegraph instance

Go to your global settings at https://sourcegraph.example.com/site-admin/global-settings and enable LSIF:

```json
  "codeIntel.lsif": true
```

After uploading LSIF files, your Sourcegraph instance will use these files to power code intelligence so that when you visit a file in that repository on your Sourcegraph instance, the code intelligence should be more precise than it was out-of-the-box.

When LSIF data does not exist for a particular file in a repository, Sourcegraph will fall back to out-of-the-box code intelligence.

## Stale code intelligence

LSIF code intelligence will be out-of-sync when you're viewing a file that has changed since the LSIF data was uploaded.

## Warning about uploading too much data

Global find-references is a resource-intensive operation that's sensitive to the number of packages for which you have uploaded LSIF data into your Sourcegraph instance. Improvements to this are planned for Sourcegraph 3.10 (see the [RFC](https://docs.google.com/document/d/1VZB0Y4tWKeOUN1JvdDgo4LHwQn875MPOI9xztzqoSRc/edit#)).

**Do not upload more than 10-40 LSIF dumps to Sourcegraph instance or you risk harming other parts of Sourcegraph. We are working to validate its performance at scale and eliminate this concern.**

## More about LSIF

To learn more, check out our lightning talk about LSIF from GopherCon 2019 or the [introductory blog post](https://about.sourcegraph.com/blog/code-intelligence-with-lsif):

<iframe width="560" height="315" src="https://www.youtube.com/embed/fMIRKRj_A88" frameborder="0" allow="accelerometer; autoplay; encrypted-media; gyroscope; picture-in-picture" allowfullscreen></iframe>
