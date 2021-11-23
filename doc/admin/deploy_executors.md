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

### Google

The GCE auto-scaling groups configured by the Sourcegraph Terraform module to respond to changes in metric values written to Cloud Monitoring. The target Sourcegraph instance is expected to continuously write these values.

To write the metric to Cloud Monitoring, the `worker` service must define the following environment variables:

- `EXECUTOR_METRIC_ENVIRONMENT_LABEL`: Must use same value as [`metrics_environment_label`](https://sourcegraph.com/search?q=context:global+repo:%5Egithub.com/sourcegraph/terraform-google-executors%24+variable+%22metrics_environment_label%22&patternType=literal)
- `EXECUTOR_METRIC_GCP_PROJECT_ID`
- `EXECUTOR_METRIC_GOOGLE_APPLICATION_CREDENTIALS_FILE`

### AWS

The EC2 auto-scaling groups configured by the Sourcegraph Terraform module to respond to changes in metric values written to CloudWatch. The target Sourcegraph instance is expected to continuously write these values.

To write the metric to CloudWatch, the `worker` service must define the following environment variables:

- `EXECUTOR_METRIC_ENVIRONMENT_LABEL`: Must use same value as [`metrics_environment_label`](https://sourcegraph.com/search?q=context:global+repo:%5Egithub.com/sourcegraph/terraform-aws-executors%24+variable+%22metrics_environment_label%22&patternType=literal)
- `EXECUTOR_METRIC_AWS_NAMESPACE`: Must be set to `sourcegraph-executor`
- `EXECUTOR_METRIC_AWS_REGION`
- `EXECUTOR_METRIC_AWS_ACCESS_KEY_ID`
- `EXECUTOR_METRIC_AWS_SECRET_ACCESS_KEY`

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
