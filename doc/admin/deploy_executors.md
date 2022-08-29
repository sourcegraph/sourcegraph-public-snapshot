# Deploying Sourcegraph executors

<aside class="beta">
<p>
<span class="badge badge-beta">Beta</span> This feature is in beta and might change in the future.
</p>

<p><b>We're very much looking for input and feedback on this feature.</b> You can either <a href="https://about.sourcegraph.com/contact">contact us directly</a>, <a href="https://github.com/sourcegraph/sourcegraph">file an issue</a>, or <a href="https://twitter.com/sourcegraph">tweet at us</a>.</p>
</aside>

[Executors](executors.md) provide a sandbox that can run resource-intensive or untrusted tasks on behalf of the Sourcegraph instance, such as:

- [automatically indexing a repository for precise code navigation](../code_navigation/explanations/auto_indexing.md)
- [running batch changes](../batch_changes/explanations/server_side.md)

## Requirements

Executors by default use KVM-based micro VMs powered by [Firecracker](https://github.com/firecracker-microvm/firecracker) in accordance with our [sandboxing model](executors.md#how-it-works) to isolate jobs from each other and the host.
This requires executors to be run on machines capable of running Linux KVM extensions. On the most popular cloud providers, this either means running executors on bare-metal machines or machines capable of nested virtualization.

Optionally, executors can be run without using KVM-based isolation, which is less secure but might be easier to run on common machines.

## Installation

In order to deploy executors that can talk to the Sourcegraph instance, you need to do the following:

1. [Configure a shared secret in the Sourcegraph instance](#configure-sourcegraph)
1. [Run executors](#run-executors)
  - [Using Terraform](#terraform)
  - [Using binaries](#binaries)
1. [Confirm executors can reach the Sourcegraph instance](#confirm-executors-are-working)
1. Optional: [Configure auto scaling](#configuring-auto-scaling)

### Configure Sourcegraph

Executors must be run separately from your Sourcegraph instance.

Since they must still be able to reach the Sourcegraph instance in order to dequeue and perform work, requests between the Sourcegraph instance and the executors are authenticated via a shared secret.

Before starting any executors, generate an arbitrary secret string (with at least 20 characters) and [set it as the `executors.accessToken` key in your Sourcegraph instance's site-config](config/site_config.md#view-and-edit-site-configuration).

Once the access token is set, executors can use that access token to talk to the Sourcegraph instance.

### Run executors

There are currently two supported ways to install and run executors:

1. [Using our Terraform modules to provision infrastructure on Google Cloud or AWS that runs executors](#terraform)
2. [Downloading and running executor binaries yourself](#binaries)

#### Terraform

We supply [Terraform modules](https://learn.hashicorp.com/tutorials/terraform/module-use?in=terraform/modules) to provision machines running executors on common cloud providers ([Google Cloud](https://github.com/sourcegraph/terraform-google-executors) or [AWS](https://github.com/sourcegraph/terraform-aws-executors)).

A Terraform definition of executor compute resources will look similar to the following basic, minimal usage. Here, we configure the use of a Terraform module defined in the public registry - no explicit installation or clone step is required to use the modules provided by Sourcegraph.

```hcl
module "executors" {
  source  = "sourcegraph/executors/<cloud>"
  version = "<version>"

  executor_sourcegraph_external_url            = "<sourcegraph_external_url>"
  executor_sourcegraph_executor_proxy_password = "<shared_secret_configured_in_sourcegraph_instance>"
  executor_queue_name                          = "codeintel" # Type of work (e.g., codeintel, batches)
  executor_instance_tag                        = "codeintel"
  executor_metrics_environment_label           = "prod"
  docker_mirror_static_ip                      = "10.0.1.4"
}
```

Two variables must be supplied to the module in order for it to contact your Sourcegraph instance:

- `sourcegraph_external_url` ([Google](https://sourcegraph.com/search?q=context:global+repo:%5Egithub.com/sourcegraph/terraform-google-executors%24+variable+%22sourcegraph_external_url%22&patternType=literal); [AWS](https://sourcegraph.com/search?q=context:global+repo:%5Egithub.com/sourcegraph/terraform-aws-executors%24+variable+%22sourcegraph_external_url%22&patternType=literal)): The **public** URL of your Sourcegraph instance. This corresponds to the `externalURL` value in your Sourcegraph instance's site-config and must be resolvable from the provisioned executor compute resources.
- `sourcegraph_executor_proxy_password` ([Google](https://sourcegraph.com/search?q=context:global+repo:%5Egithub.com/sourcegraph/terraform-google-executors%24+variable+%22sourcegraph_executor_proxy_password%22&patternType=literal); [AWS](https://sourcegraph.com/search?q=context:global+repo:%5Egithub.com/sourcegraph/terraform-aws-executors%24+variable+%22sourcegraph_executor_proxy_password%22&patternType=literal)): The access token chosen and configured above.

For Google Cloud, make sure the [IAM API](https://console.cloud.google.com/apis/api/iam.googleapis.com/overview) is enabled.

Additional values may need to be supplied for a specific cloud provider. Refer to the relevant Terraform module documentation for specifics.

To deploy executor compute resources defined in the Terraform file above, simply run `terraform apply`.

See the [Examples](#examples) for more information on how to configure and deploy single or multiple executors.

#### Binaries

You can also download and run the executor binaries yourself, without using Terraform.

<aside class="warning">
<p>
<strong>Caution: This process is currently very manual and might change between releases. We will mature this process as we move towards GA of executors, but for now expect to do manual work for upgrades and for the initial setup.</strong>
</p>
</aside>

The Sourcegraph-provided executor VM images for Cloud providers contain a lot of things that executors currently expect to be present and configured like outlined in there. For your currently running release of Sourcegraph, please refer to our [install script setting up the machine for executors use](https://github.com/sourcegraph/sourcegraph/blob/main/enterprise/cmd/executor/vm-image/install.sh) for your Sourcegraph version. This script also handles some security concerns, like configuring iptables, so make sure to follow this script when you derive your own machine setup.

Once dependencies are met, you can download and run executor binaries:

**Step 1:** Confirm that virtualization is enabled

KVM (virtualization) support is required for [our sandboxing model](executors.md#how-it-works) with Firecracker. The following command checks whether virtualization is enabled on the machine (it should print something):

```bash
$ lscpu | grep Virtualization

Virtualization:      VT-x
```

On Ubuntu-based distributions, you can also use the tool `kvm-ok` available in the `cpu-checker` package to reliably validate KVM support on your host:

```bash
# Install cpu-checker
$ apt-get update && apt-get install -y cpu-checker

# Check for KVM support
$ kvm-ok
INFO: /dev/kvm exists
KVM acceleration can be used
```

**Step 2:** Download latest executor binary

Below are the download links for the *insiders* release (`latest`) of executors:

* [`info.txt`](https://storage.googleapis.com/sourcegraph-artifacts/executor/latest/info.txt)
* [`linux-amd64/executor`](https://storage.googleapis.com/sourcegraph-artifacts/executor/latest/linux-amd64/executor)
* [`linux-amd64/executor_SHA256SUM`](https://storage.googleapis.com/sourcegraph-artifacts/executor/latest/linux-amd64/executor_SHA256SUM)

Download and setup the `executor` binary:

```bash
curl -sfLo executor https://storage.googleapis.com/sourcegraph-artifacts/executor/latest/linux-amd64/executor
chmod +x executor
mv executor /usr/local/bin
```

**Step 3:** Configure your machine

The executor makes a lot assumptions about it's environment today:
- It has to have a working installation of ignite that is configured
- and expects the `sourcegraph/ignite-ubuntu` base image for the VMs ready.
- In addition, we also do some hardening of regular ignite in our pre-built environments to restrict networking further and use up-to-date kernels.
Until we automate more of this outside of our pre-built images, refer to [the script that our pre-built images use](https://github.com/sourcegraph/sourcegraph/blob/main/enterprise/cmd/executor/vm-image/install.sh)
for how we do it. Usually, 95% of this file should be portable to any environment.

**Step 4:** Setup required environment variables and start

The executor binary is configured through environment variables. Those need to be passed to it when you start the executor.

| Env var                       | Example value | Description |
| ------------------------------| ------------- | ----------- |
| `EXECUTOR_FRONTEND_URL`       | `http://sourcegraph.example.com` | The external URL of the Sourcegraph instance. |
| `EXECUTOR_FRONTEND_PASSWORD`  | `our-shared-secret` | The shared secret configured in the Sourcegraph instannce under `executors.accessToken` |
| `EXECUTOR_QUEUE_NAME`         | `batches`     | The name of the queue to pull jobs from to. Possible values: `batches` and `codeintel` |

```bash
# Example:
export EXECUTOR_QUEUE_NAME=batches
export EXECUTOR_FRONTEND_URL=http://sourcegraph.example.com
export EXECUTOR_FRONTEND_PASSWORD=hunter2hunter2hunter2
```

Done! You can start your executor now.

### Confirm executors are working

If executor instances boot correctly and can authenticate with the Sourcegraph frontend, they will show up in the _Executors_ page under _Site Admin_ > _Maintenance_.

![Executor list in UI](https://storage.googleapis.com/sourcegraph-assets/docs/images/code-intelligence/sg-3.34/executor-ui-test.png)

### Examples

The following are complete examples of provisioning a _single_ executor type using each of the provided Terraform modules. In these examples, executors pull from the queue `codeintel`, which controls [auto-indexing](../code_navigation/explanations/auto_indexing.md).

- [AWS example](https://github.com/sourcegraph/terraform-aws-executors/tree/master/examples/single-executor)
- [Google example](https://github.com/sourcegraph/terraform-google-executors/tree/master/examples/single-executor)

The following are complete examples of provisioning _multiple_ executor types using the provided Terraform submodules. In these examples, two pools of executors pull from the `codeintel` and `batches` queues, which control [auto-indexing](../code_navigation/explanations/auto_indexing.md) and  [running batch changes server-side](../batch_changes/explanations/server_side.md), respectively.

- [AWS example](https://github.com/sourcegraph/terraform-aws-executors/tree/master/examples/multiple-executors)
- [Google example](https://github.com/sourcegraph/terraform-google-executors/tree/master/examples/multiple-executors)

#### Example step-by-step installation

Let's walk through setting up a single executor VM on GCP using the [Google Cloud Terraform Module for Sourcegraph executors](https://github.com/sourcegraph/terraform-google-executors) and indexing a repository.

1. Install Terraform (must match the version listed in [`.tool-versions`](https://github.com/sourcegraph/terraform-google-executors/blob/master/providers.tf)):

```
brew install tfenv
tfenv install 1.1.5
tfenv use 1.1.5
```

2. Install the [`gcloud CLI`](https://cloud.google.com/sdk/docs/install)
3. Run `gcloud auth application-default login`
4. Open your Sourcegraph instance in your browser, click your profile in the top right, click **Site admin**, expand **Configuration**, click **Site configuration**, and set:
  - `"externalURL": "<URL>"` to a URL that is accessible from the GCP VM that will be created later (e.g. a public URL such as `https://sourcegraph.example.com`)
  - `"executors.accessToken": "<new long secret>"` to a new long secret (e.g. `cat /dev/random | base64 | head -c 20`)
  - `"codeIntelAutoIndexing.enabled": true`
5. Download the [example files directory](https://github.com/sourcegraph/terraform-google-executors/blob/master/examples/single-executor) and change these terraform variables:
  - `project`: your GCP project name and change `region` and `zone` if needed
  - `executor_sourcegraph_external_url`: this must match `externalURL` you set in your site config
  - `executor_sourcegraph_executor_proxy_password`: this must match `executors.accessToken` you set in your site config
6. Run `terraform init` to download the Sourcegraph executor modules
7. Run `terraform apply` and enter "yes" after reviewing the proposed changes to create the executor VM
8. Go back to the site admin page, expand **Maintenance**, click **Executors**, and check to see if your executor shows up in the list with a green dot 🟢. If it's not there:
  - Make sure `terraform apply` exited with code 0 and did not print any errors
  - Make sure a GCP VM was created:

```
$ gcloud compute instances list
NAME                                          ZONE           MACHINE_TYPE   PREEMPTIBLE  INTERNAL_IP  EXTERNAL_IP    STATUS
sourcegraph-executor-h0rv                     us-central1-c  n1-standard-4               10.0.1.16                   RUNNING
sourcegraph-executors-docker-registry-mirror  us-central1-c  n1-standard-2               10.0.1.2                    RUNNING
```

  - Make sure the `executor` service is running:

```
you@sourcegraph-executor-h0rv:~$ systemctl status executor
🟢 executor.service - User code executor
     Loaded: loaded (/etc/systemd/system/executor.service; enabled; vendor preset: enabled)
     Active: active (running) since Thu 2021-11-18 02:28:48 UTC; 19s ago
```

  - Make sure there are no errors in the `executor` service logs:

```
you@sourcegraph-executor-h0rv:~$ journalctl -u executor | less
Nov 18 02:31:01 sourcegraph-executor-h0rv executor[2465]: t=2021-11-18T02:31:01+0000 lvl=dbug msg="TRACE internal" host=... path=/.executors/queue/codeintel/dequeue code=204 duration=92.131237ms
Nov 18 02:31:01 sourcegraph-executor-h0rv executor[2465]: t=2021-11-18T02:31:01+0000 lvl=dbug msg="TRACE internal" host=... path=/.executors/queue/codeintel/canceled code=200 duration=90.630467ms
Nov 18 02:31:02 sourcegraph-executor-h0rv executor[2465]: t=2021-11-18T02:31:02+0000 lvl=dbug msg="TRACE internal" host=... path=/.executors/queue/codeintel/dequeue code=204 duration=91.269106ms
Nov 18 02:31:02 sourcegraph-executor-h0rv executor[2465]: t=2021-11-18T02:31:02+0000 lvl=dbug msg="TRACE internal" host=... path=/.executors/queue/codeintel/canceled code=200 duration=161.469685ms
```

  - Make sure the `EXECUTOR_FRONTEND_URL` and `EXECUTOR_FRONTEND_PASSWORD` in `/etc/systemd/system/executor.env` are correct
  - Make sure the VM can hit your `externalURL`:

```
you@sourcegraph-executor-h0rv:~$ curl <your externalURL here>
<a href="/sign-in?returnTo=%2F">Found</a>
```

9. Go back to the site admin page, expand **Code graph**, click **Configuration**, click **Create new policy**, and fill in:
  - Name: `TEST`
  - Click **add a repository pattern**
  - Repository pattern #1: set this to an existing repository on your Sourcegraph instance (e.g. `github.com/gorilla/mux`)
  - Type: `HEAD`
  - Auto-indexing: Enabled
10. Go to that repository's page, click **Code graph**, click **Auto-indexing**, and check to see if an indexing job has appeared. If nothing is there:
  - Try clicking **Enqueue**
  - Try setting a higher update frequency: `PRECISE_CODE_INTEL_AUTO_INDEXING_TASK_INTERVAL=10s`
  - Try setting a lower delay: `PRECISE_CODE_INTEL_AUTO_INDEXING_REPOSITORY_PROCESS_DELAY=10s`
11. Once you have a completed indexing job, click **Uploads** and check to see that an index has been uploaded.
12. Once the index has been uploaded, you should see the **`PRECISE`** badge in the hover popover! 🎉 

## Configuring auto scaling

> NOTE: Auto scaling is currently not supported when [downloading and running executor binaries yourself](#binaries), and on managed instances when using self-hosted executors, since it requires deployment adjustments.

Auto scaling of executor instances can help to increase concurrency of jobs, without paying for unused resources. With auto scaling, you can scale down to 0 instances when no workload exist and scale up as far as you like and your cloud provider can support. Auto scaling needs to be configured separately.

Auto scaling makes use of the auto-scaling capabilities of the respective cloud provider (AutoScalingGroups on AWS and Instance Groups on GCP). Sourcegraph's `worker` service publishes a scaling metric (that is, the number of jobs in queue) to the cloud providers. Then, based on that reported value, the auto scalers add and remove compute resources to match the required amount of compute. The autoscaler will attempt to hold 1 instance running per each [`jobs_per_instance_scaling`](https://sourcegraph.com/search?q=context:global+repo:%5Egithub.com/sourcegraph/terraform-.*-executors%24+variable+%22jobs_per_instance_scaling%22&patternType=literal) items in queue.

For example, if `jobs_per_instance_scaling` is set to `20` and the queue size is currently `400`, then `20` instances would be determined as required to handle the load. You might want to tweak this number based on the [machine type](https://sourcegraph.com/search?q=context:global+repo:%5Egithub.com/sourcegraph/terraform-.*-executors%24+variable+%22machine_type%22+-f:docker-mirror&patternType=literal), [concurrency per machine](https://sourcegraph.com/search?q=context:global+repo:%5Egithub.com/sourcegraph/terraform-.*-executors%24+variable+%22maximum_num_jobs%22&patternType=literal) and desired processing speed.

With the Terraform variables [`min_replicas`](https://sourcegraph.com/search?q=context:global+repo:%5Egithub.com/sourcegraph/terraform-.*-executors%24+variable+%22min_replicas%22&patternType=literal) and [`max_replicas`](https://sourcegraph.com/search?q=context:global+repo:%5Egithub.com/sourcegraph/terraform-.*-executors%24+variable+%22max_replicas%22&patternType=literal) in the Terraform modules linked to above, you can configure the minimum and maximum number of compute machines to be run at a given time.

For auto scaling to work, two things must be true:

1. `min_replicas` must be `>= 0` and `max_replicas` must be `> min_replicas`.
2. The Sourcegraph instance (its `worker` service, specifically) needs to publish scaling metrics to the used cloud provider.

For the latter to work, the Sourcegraph instance needs to be configured with the correct credentials that allow it to access the cloud provider.

The `credentials` submodule in both our [AWS](https://sourcegraph.com/github.com/sourcegraph/terraform-aws-executors/-/tree/modules/credentials) and [GCP](https://sourcegraph.com/github.com/sourcegraph/terraform-google-executors/-/tree/modules/credentials) executor modules exists for that purpose. When used, the `credentials` module sets up the credentials on the cloud provider and returns them in the Terraform outputs.

Here's an example of how one would use the `credentials` submodule:

```terraform
module "my-credentials" {
  source  = "sourcegraph/executors/<cloud>//modules/credentials"

  # Find the latest version matching your Sourcegraph version here:
  # - https://github.com/sourcegraph/terraform-google-executors/tags
  # - https://github.com/sourcegraph/terraform-aws-executors/tags
  version = "<version>"

  region          = <region>
  resource_prefix = ""
}

# For Google:
output "metric_writer_credentials_file" {
  value = module.my-credentials.metric_writer_credentials_file
}

# For AWS:
output "metric_writer_access_key_id" {
  value = module.my-credentials.metric_writer_access_key_id
}
output "metric_writer_secret_key" {
  value = module.my-credentials.metric_writer_secret_key
}
```

After a `terraform apply`, a `terraform output` should yield something like this:

```
# For AWS:
metric_writer_access_key_id = <THE_ACCESS_KEY_TO_CONFIGURE>
metric_writer_secret_key    = <THE_SECRET_KEY_TO_CONFIGURE>

# For Google:
metric_writer_credentials_file = <THE_CREDENTIALS_FILE_CONTENT_BASE64_ENCODED>
```

These outputs can then be used to configure the Sourcegraph instance, which is explained in the following sections for the different cloud providers.

### Google

The GCE auto-scaling groups configured by the Sourcegraph Terraform module respond to changes in metric values written to Cloud Monitoring. The target Sourcegraph instance is expected to continuously write these values.

To write the scaling metric to Cloud Monitoring, the `worker` service must have defined the following environment variables:

- `EXECUTOR_METRIC_ENVIRONMENT_LABEL`: Must use the value that was set as [`metrics_environment_label`](https://sourcegraph.com/search?q=context:global+repo:%5Egithub.com/sourcegraph/terraform-google-executors%24+variable+%22metrics_environment_label%22&patternType=literal) when provisioning `executors`
- `EXECUTOR_METRIC_GCP_PROJECT_ID`: The GCP project ID
- Option 1: `EXECUTOR_METRIC_GOOGLE_APPLICATION_CREDENTIALS_FILE_CONTENT`: The base64-decoded output of the `metric_writer_credentials_file` from above.
- Option 2: `EXECUTOR_METRIC_GOOGLE_APPLICATION_CREDENTIALS_FILE`: Path to a file containing the base64-decoded `metric_writer_credentials_file` output from above.

### AWS

The EC2 auto-scaling groups configured by the Sourcegraph Terraform module respond to changes in metric values written to CloudWatch. The target Sourcegraph instance is expected to continuously write these values.

To write the scaling metric to CloudWatch, the `worker` service must have defined the following environment variables:

- `EXECUTOR_METRIC_ENVIRONMENT_LABEL`: Must use the same value as [`metrics_environment_label`](https://sourcegraph.com/search?q=context:global+repo:%5Egithub.com/sourcegraph/terraform-aws-executors%24+variable+%22metrics_environment_label%22&patternType=literal)
- `EXECUTOR_METRIC_AWS_NAMESPACE`: Must be set to `sourcegraph-executor`
- `EXECUTOR_METRIC_AWS_REGION`
- `EXECUTOR_METRIC_AWS_ACCESS_KEY_ID`
- `EXECUTOR_METRIC_AWS_SECRET_ACCESS_KEY`

### Testing auto scaling

Once these are set, and the worker service has been restarted, you should be able to find the scaling metrics in your cloud providers dashboards.

To test if the metric is correctly reported into the Cloud provider:

- On Google Cloud, this can be found in the **Metrics explorer**. Select **Resource type: Global** and then **Metric: `custom/executors/queue/size`**. You should see values reported here. 0 is also an indicator that it works correct.

- On AWS, this can be found in the CloudWatch metrics section. Under **All metrics**, select the namespace `sourcegraph-executor` and then the metric `environment, queueName`. Make sure there are entries returned.

Next, you can test whether the number of executors rises and shrinks as load spikes occur. Keep in mind that auto-scaling is not a real-time operation on most cloud providers and usually takes a short moment and can have some delays between the metric going down and the desired machine count adjusting.
