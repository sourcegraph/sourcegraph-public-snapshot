# Enable code intelligence auto-indexing

<aside class="experimental">
<p>
<span class="badge badge-experimental">Experimental</span> This feature is experimental and might change or be removed in the future. We've released it as an experimental feature to provide a preview of functionality we're working on.
</p>

<p><b>We're very much looking for input and feedback on this feature.</b> You can either <a href="https://about.sourcegraph.com/contact">contact us directly</a>, <a href="https://github.com/sourcegraph/sourcegraph">file an issue</a>, or <a href="https://twitter.com/sourcegraph">tweet at us</a>.</p>
</aside>

This how-to explains how to turn on [auto-indexing](../explanations/auto_indexing.md) on your Sourcegraph instance to enable [precise code intelligence](../explanations/precise_code_intelligence.md).

## Deploy executors

First, [deploy the executor service](../../../../admin/deploy_executors.md) targeting your Sourcegraph instance. This will provide the necessary compute resources that clone the target Git repository, securely analyze the code to produce a precise code intelligence index, then upload that index to your Sourcegraph instance for processing.

## Enable index job scheduling

Next, enable the precise code intelligence auto-indexing feature by enabling the following feature flag in your Sourcegraph instance's site configuration.

```yaml
{
  "codeIntelAutoIndexing.enabled": true
}
```

This step will control the scheduling of indexing jobs which are made available to the executors deployed in the previous step.

## Configure auto-indexing policies

Once auto-indexing has been enabled, [create auto-indexing policies](configure_auto_indexing.md) to control the set of repositories and commits that are eligible for indexing.

## Tune the index scheduler

The frequency of index job scheduling can be tuned via the following environment variables read by `worker` service containers running the [`codeintel-auto-indexing`](../../../admin/workers.md#codeintel-auto-indexing) task.

**`PRECISE_CODE_INTEL_AUTO_INDEXING_TASK_INTERVAL`**: The frequency with which to run periodic codeintel auto-indexing tasks. Default is every 10 minutes.

**`PRECISE_CODE_INTEL_AUTO_INDEXING_REPOSITORY_PROCESS_DELAY`**: The minimum frequency that the same repository can be considered for auto-index scheduling. Default is every 24 hours.

**`PRECISE_CODE_INTEL_AUTO_INDEXING_REPOSITORY_BATCH_SIZE`**: The number of repositories to consider for auto-indexing scheduling at a time. Default is 100.

**`PRECISE_CODE_INTEL_AUTO_INDEX_MAXIMUM_REPOSITORIES_INSPECTED_PER_SECOND`**: The maximum number of repositories inspected for auto-indexing per second. Set to zero to disable limit. Default is 0.
