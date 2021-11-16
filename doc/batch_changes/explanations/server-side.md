# Using executors to compute Batch Changes server-side

<aside class="experimental"></aside>

Instead of computing Batch Changes [diffs](how_src_executes_a_batch_spec.md) locally using `src-cli`, it's possible to offload this task to [executors](../../admin/deploy_executors.md). This feature is [experimental](../../admin/beta_and_experimental_features#experimental-features.md). Executors are also required to enable Code Intelligence [auto-indexing](https://docs.sourcegraph.com/code_intelligence/explanations/auto_indexing).

If enabled, server-side Batch Changes computing allows to:

- run large-scale batch changes that would be impractical to compute locally
- speed up batch change creation time by distributing batch change computing over several executors
- reduce the setup time required to onboard new users to Batch Changes

## Limitations

This feature is experimental. In particular, it comes with limitations:

- only site admins can run batch changes server-side
- the server side batch changes UI is minimal and will change a lot before the GA release
- documentation is minimal and will change a lot before the GA release
- batch change execution is not optimized
- executors can only be deployed on AWS and GCP, with Terraform (see [deploying executors](../../admin/deploy_executors.md))
- step-wise caching is not included server side
- steps cannot include [files](batch_spec_yaml_reference.md#steps-files)


Server-side Batch Changes has been tested to run a simple 20k changeset batch change. Actual performance and setup requirements depend on the complexity of the batch change.

## How to enable computing Batch Changes server-side

1. TODO
1. 
1. 
1. 
