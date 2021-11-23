# Deploying Sourcegraph executors

<aside class="experimental">
<p>
<span class="badge badge-experimental">Experimental</span> This feature is experimental and might change or be removed in the future. We've released it as an experimental feature to provide a preview of functionality we're working on.
</p>

<p><b>We're very much looking for input and feedback on this feature.</b> You can either <a href="https://about.sourcegraph.com/contact">contact us directly</a>, <a href="https://github.com/sourcegraph/sourcegraph">file an issue</a>, or <a href="https://twitter.com/sourcegraph">tweet at us</a>.</p>
</aside>

Executors provide a sandbox that can run resource-intensive or untrusted tasks on behalf of the Sourcegraph instance, such as:

- [automatically indexing a repository for precise code intelligence](../code_intelligence/explanations/auto_indexing.md)
- [computing batch changes](../batch_changes/explanations/server_side.md)

## Installation

Executors operate outside of your Sourcegraph instance and must be provisioned separately from your Sourcegraph server deployment method.

In order for the executors to dequeue and perform work, they must be able to reach the target Sourcegraph instance. These requests are authenticated via a shared secret value known by the Sourcegraph instance. Before provisioning executor compute resources, generate an arbitrary secret string (with at least 20 characters) and [set it as the `executors.accessToken` key in your Sourcegraph instance's site-config](config/site_config.md#view-and-edit-site-configuration).

Once the access token is set, executor compute resources can be provisioned. We supply [Terraform modules](https://learn.hashicorp.com/tutorials/terraform/module-use?in=terraform/modules) to provision such resources on common cloud providers ([Google Cloud](https://github.com/sourcegraph/terraform-google-executors) and [AWS](https://github.com/sourcegraph/terraform-aws-executors)).

A Terraform definition of executor compute resources will look similar to the following basic, minimal usage. Here, we configure the use of a Terraform module defined in the public registry - no explicit installation or clone step is required to use the modules provided by Sourcegraph.

```hcl
module "executors" {
  source  = "sourcegraph/executors/<cloud>"
  version = "<version>"

  executor_sourcegraph_external_url            = "<sourcegraph_external_url>"
  executor_sourcegraph_executor_proxy_password = "<sourcegraph_executor_proxy_password>"
  executor_queue_name                          = "codeintel" # Type of work (e.g., codeintel, batches)
  executor_instance_tag                        = "codeintel"
  executor_metrics_environment_label           = "prod"
  docker_mirror_static_ip                      = "10.0.1.4"
}
```

Two variables must be supplied to the module in order for it to contact your Sourcegraph instance:

- `sourcegraph_external_url` ([Google](https://sourcegraph.com/search?q=context:global+repo:%5Egithub.com/sourcegraph/terraform-google-executors%24+variable+%22sourcegraph_external_url%22&patternType=literal); [AWS](https://sourcegraph.com/search?q=context:global+repo:%5Egithub.com/sourcegraph/terraform-aws-executors%24+variable+%22sourcegraph_external_url%22&patternType=literal)): The **public** URL of your Sourcegraph instance.
- `sourcegraph_executor_proxy_password` ([Google](https://sourcegraph.com/search?q=context:global+repo:%5Egithub.com/sourcegraph/terraform-google-executors%24+variable+%22sourcegraph_executor_proxy_password%22&patternType=literal); [AWS](https://sourcegraph.com/search?q=context:global+repo:%5Egithub.com/sourcegraph/terraform-aws-executors%24+variable+%22sourcegraph_executor_proxy_password%22&patternType=literal)): The access token chosen and configured above.

Additional values may need to be supplied for a specific cloud provider. Refer to the relevant Terraform module documentation for specifics.

To deploy executor compute resources defined in the Terraform file above, simply run `terraform apply`.

If executor instances boot correctly and can authenticate with the Sourcegraph frontend, they will show up in the _Executors_ page under _Site Admin_ > _Maintenance_.

![Executor list in UI](https://storage.googleapis.com/sourcegraph-assets/docs/images/code-intelligence/sg-3.34/executor-ui-test.png)

### Examples

The following are complete examples of provisioning a _single_ executor type using each of the provided Terraform modules. In these examples, executors pull from the queue `codeintel`, which controls auto-indexing.

- [AWS example](https://github.com/sourcegraph/terraform-aws-executors/tree/master/examples/single-executor)
- [Google example](https://github.com/sourcegraph/terraform-google-executors/tree/master/examples/single-executor)

The following are complete examples of provisioning _multiple_ executor types using the provided Terraform submodules. In these examples, two pools of executors pull from the `codeintel` and `batches` queues, which control auto-indexing and server-side batch changes, respectively.

- [AWS example](https://github.com/sourcegraph/terraform-aws-executors/tree/master/examples/multiple-executors)
- [Google example](https://github.com/sourcegraph/terraform-google-executors/tree/master/examples/multiple-executors)

## Configuring auto scaling

Auto scaling of executor instances can help to increase concurrency of jobs, without paying for unused resources. With auto scaling, you can scale down to 0 instances when no workload exist and scale up as far as you like and your cloud provider can support. Auto scaling needs to be configured separately.

Auto scaling makes use of the auto-scaling capabilities of the respective cloud provider (AutoScalingGroups on AWS and Instance Groups on GCP). Sourcegraph's `worker` service publishes a scaling metric (that is, the number of jobs in queue) to the cloud providers. Then, based on that reported value, the auto scalers add and remove compute resources to match the required amount of compute.

With the Terraform variables [`min-replicas`](https://sourcegraph.com/search?q=context:global+repo:%5Egithub.com/sourcegraph/terraform-.*-executors%24+variable+%22min_replicas%22&patternType=literal) and [`max-replicas`](https://sourcegraph.com/search?q=context:global+repo:%5Egithub.com/sourcegraph/terraform-.*-executors%24+variable+%22max_replicas%22&patternType=literal) in the Terraform modules linked to above, you can configure the minimum and maximum number of compute machines to be run at a given time. `min-replicas` must be `>= 0`.

For auto scaling to work, the Sourcegraph instance (its `worker` service, specifically) needs to be configured to have credentials to publish a scaling metric to the cloud provider used. Therefore, the `credentials` submodule exists in both our [AWS](https://sourcegraph.com/github.com/sourcegraph/terraform-aws-executors/-/tree/modules/credentials) and [GCP](https://sourcegraph.com/github.com/sourcegraph/terraform-google-executors/-/tree/modules/credentials) executor modules. Using them, you get properly configured credentials in the Terraform outputs.

```terraform
module "credentials" {
  source  = "sourcegraph/executors/<cloud>//modules/credentials"
  version = "<version>"

  region          = <region>
  resource_prefix = ""
}
```

When applied, this will yield something like

```
metric_writer_access_key_id = <THE_ACCESS_KEY_TO_CONFIGURE>
metric_writer_secret_key    = <THE_SECRET_KEY_TO_CONFIGURE>
```

Use these credentials in the following way for the different cloud providers:

### Google

The GCE auto-scaling groups configured by the Sourcegraph Terraform module respond to changes in metric values written to Cloud Monitoring. The target Sourcegraph instance is expected to continuously write these values.

To write the scaling metric to Cloud Monitoring, the `worker` service must have defined the following environment variables:

- `EXECUTOR_METRIC_ENVIRONMENT_LABEL`: Must use the same value as [`metrics_environment_label`](https://sourcegraph.com/search?q=context:global+repo:%5Egithub.com/sourcegraph/terraform-google-executors%24+variable+%22metrics_environment_label%22&patternType=literal)
- `EXECUTOR_METRIC_GCP_PROJECT_ID`
- `EXECUTOR_METRIC_GOOGLE_APPLICATION_CREDENTIALS_FILE`

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

- On Google Cloud, this can be found in the Metrics explorer. Select Resource type = "Global" and then Metric = "custom/executors/queue/size". You should see some values reported here, 0 is also an indicator that it works correct.

- On AWS, this can be found in the CloudWatch metrics section. Under "All metrics", select the namespace "sourcegraph-executor" and then the metric "environment, queueName". Make sure there are entries returned.

Next, you can test whether the number of executors rises and shrinks as load spikes occur. Keep in mind that auto-scaling is not a real-time operation on most cloud providers and usually takes a short moment and can have some delays between the metric going down and the desired machine count adjusting.

## Configuring observability

Sourcegraph ships with dashboards to display executor metrics. To populate these dashboards, the target Prometheus instance must be able to scrape the executor metrics endpoint.

### Google

The Prometheus configuration must add the following scraping job that uses [GCE service discovery configuration](https://prometheus.io/docs/prometheus/latest/configuration/configuration/#gce_sd_config):

```yaml
- job_name: 'sourcegraph-executors'
  metrics_path: /proxy
  params:
    module: [executor]
  gce_sd_configs: &executor_gce_config
    - project: {GCP_PROJECT}
      port: 9999
      zone: {GCP_ZONE}
      filter: '(labels.executor_tag = {INSTANCE_TAG})'
  relabel_configs: &executor_relabel_config
    - source_labels: [__meta_gce_public_ip]
      target_label: __address__
      replacement: "${1}${2}:9999"
      separator: ''
    - source_labels: [__meta_gce_zone]
      regex: ".+/([^/]+)"
      target_label: zone
      separator: ''
    - source_labels: [__meta_gce_project]
      target_label: project
    - source_labels: [__meta_gce_instance_name]
      target_label: instance
      separator: ''
    - regex: "__meta_gce_metadata_(image_.+)"
      action: labelmap
- job_name: 'sourcegraph-executor-nodes'
  metrics_path: /proxy
  params:
    module: [node]
  gce_sd_configs: *executor_gce_config
  relabel_configs: *executor_relabel_config
- job_name: 'sourcegraph-executors-docker-registry-mirrors'
  metrics_path: /proxy
  params:
    module: [registry]
  gce_sd_configs: &gce_executor_mirror_config
    - project: {GCP_PROJECT}
      port: 9999
      zone: {GCP_ZONE}
      filter: '(labels.executor_tag = {INSTANCE_TAG}-docker-mirror)'
  relabel_configs: *executor_relabel_config
- job_name: 'sourcegraph-executors-docker-registry-mirror-nodes'
  metrics_path: /proxy
  params:
    module: [node]
  gce_sd_configs: *gce_executor_mirror_config
  relabel_configs: *executor_relabel_config
```

The `{INSTANCE_TAG}` value above must be the same as [`instance_tag`](https://sourcegraph.com/search?q=context:global+repo:%5Egithub.com/sourcegraph/terraform-google-executors%24+variable+%22instance_tag%22&patternType=literal).

### AWS

The Prometheus configuration must add the following scraping job that uses [EC2 service discovery configuration](https://prometheus.io/docs/prometheus/latest/configuration/configuration/#ec2_sd_config).

```yaml
- job_name: 'sourcegraph-executors'
  metrics_path: /proxy
  params: 
    module: [executor]
  ec2_sd_configs: &executor_ec2_config
    - region: {AWS_REGION}
      port: 9999
      filters:
        - name: tag:executor_tag
          values: [{INSTANCE_TAG}]
  relabel_configs: &executor_relabel_config
    - source_labels: [__meta_ec2_public_ip]
      target_label: __address__
      replacement: "${1}${2}:9999"
      separator: ''
    - source_labels: [__meta_ec2_availability_zone]
      regex: ".+/([^/]+)"
      target_label: zone
      separator: ''
    - source_labels: [__meta_ec2_instance_id]
      target_label: instance
      separator: ''
    - source_labels: [__meta_ec2_ami]
      target_label: version
- job_name: 'sourcegraph-executor-nodes'
  metrics_path: /proxy
  params:
    module: [node]
  ec2_sd_configs: *executor_ec2_config
  relabel_configs: *executor_relabel_config
- job_name: 'sourcegraph-executors-docker-registry-mirrors'
  metrics_path: /proxy
  params:
    module: [registry]
  ec2_sd_configs: &ec2_executor_mirror_config
    - region: {AWS_REGION}
      port: 9999
      filters:
        - name: tag:executor_tag
          values: [{INSTANCE_TAG}-docker-mirror]
  relabel_configs: *executor_relabel_config
- job_name: 'sourcegraph-executors-docker-registry-mirror-nodes'
  metrics_path: /proxy
  params:
    module: [node]
  ec2_sd_configs: *ec2_executor_mirror_config
  relabel_configs: *executor_relabel_config
```

The `{INSTANCE_TAG}` value above must be the same as [`instance_tag`](https://sourcegraph.com/search?q=context:global+repo:%5Egithub.com/sourcegraph/terraform-aws-executors%24+variable+%22instance_tag%22&patternType=literal).
