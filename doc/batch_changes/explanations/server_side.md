# Using executors to compute Batch Changes server-side

<aside class="experimental">This feature is experimental</aside>

By default, Batch Changes uses a command line interface in your local environment to [compute diffs](how_src_executes_a_batch_spec.md) and create changesets. This can be impractical for creating batch changes affecting hundreds or thousands of repositories, with large numbers of workspaces, or if the batch change steps require CPU, memory, or disk resources that are unavailable locally.

Instead of computing Batch Changes locally using `src-cli`, it's possible to offload this task to [executors](../../admin/deploy_executors.md). Executors are also required to enable Code Intelligence [auto-indexing](../../code_intelligence/explanations/auto_indexing.md).

If enabled, server-side Batch Changes computing allows to:

- run large-scale batch changes that would be impractical to compute locally
- speed up batch change creation time by distributing batch change computing over several executors
- reduce the setup time required to onboard new users to Batch Changes
- get a GUI-only experience if you want to

## Limitations

This feature is experimental. In particular, it comes with the following limitations, that we plan to resolve before GA.

- Server-side Batch Changes is only available for self-hosted deployments. It is not available on Sourcegraph Cloud and on managed instances.
- Only site admins can run batch changes server-side.
- The server side batch changes UI is minimal and will change a lot before the GA release.
- Documentation is minimal and will change a lot before the GA release.
- Batch change execution is not optimized.
- Executors can only be deployed using Terraform (AWS or GCP) or using pre-built binaries (see [deploying executors](../../admin/deploy_executors.md)).
- Steps cannot include [files](../references/batch_spec_yaml_reference.md#steps-files).

Server-side Batch Changes has been tested to run a simple 20k changeset batch change. Actual performance and setup requirements depend on the complexity of the batch change.

Feedback on server side Batch Changes is very welcome, feel free to open an [issue](https://github.com/sourcegraph/sourcegraph/issues), reach out through your usual support channel, or send a [direct message](https://twitter.com/MaloMarrec).

## FAQ

### Can large batch changes execution be distributed on multiple executors?

They can! Each changeset that is computed can be assigned to a separate executor, provided there are enough executors available.

### What additional resources do I need to provision to run server-side batch changes?

See [deploying executors](../../admin/deploy_executors.md) page. The short answer is: as little as a single compute instance and a docker registry mirror if you just want to process batch changes at a small scale; an autoscaling group of instances if you want to process large batch changes very fast.

### Can someone accidentally take down the Sourcegraph instance if they run too big a batch change?

No. Executors have been designed for the Sourcegraph instance to offload resource-intensive tasks. The Sourcegraph instance itself only queues up batch changes for processing, tracks execution, then uses the resulting diffs to open and track changesets just as it would for batch changes created locally using the `src-cli`.

### I have several machines configured as executors, and they don't have the same specs (eg. memory). Can I submit some batch changes specifically to a given machine?

No, all executors are equal in the eyes of Sourcegraph. We suggest using only one type of machine.

### What happens if the execution of a step fails?

If the execution of a step on a given repository fails, that repository will be skipped, and execution on the other repositories will continue. Standard error and output will be available to the user for debugging purposes.

### How do executors interact with code hosts? Will they clone repos directly? 

Executors do not interact directly with code hosts. They behave in a way [similar to src CLI](how_src_executes_a_batch_spec.md) today: executors interact with the Sourcegraph instance, the Sourcegraph instance interacts with the code host. In particular, executors download code from the Sourcegraph instance and executors do not need to access code hosts credentials directly.
