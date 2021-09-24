# Monitoring

As executors are not part of the default deployment, Prometheus must be additionally configured to also discovery and add the executor compute machines as scrape targets.
To do this, the relevant configuration, as outlined below depending on your deployment, will need to be added to the `prometheus.yml` file in the [Prometheus Kubernetes ConfigMap](https://github.com/sourcegraph/deploy-sourcegraph/blob/master/base/prometheus/prometheus.ConfigMap.yaml#L3).

These configurations assume that GCE/EC2 Service Discovery will be used to dynamically add the machines as scrape targets. The [Prometheus Kubernetes Deployment](https://github.com/sourcegraph/deploy-sourcegraph/blob/master/base/prometheus/prometheus.Deployment.yaml) may require additional environment variables/volume mounts in order for service discovery to be authorized by your cloud provider. Documentation on this is linked below in each of the cloud provider subheadings.

## Google Cloud Deployment

See the [following Prometheus documentation](https://prometheus.io/docs/prometheus/latest/configuration/configuration/#gce_sd_config) on how to allow GCE Service Discovery.

The below snippet is an example target configuration. Adjust according to any changed parameters such as GCE Zone, Project Name etc.
```yaml
- job_name: 'sourcegraph-code-intel-indexers'
  gce_sd_configs:
    - project: sourcegraph-code-intel
      port: 6060
      zone: us-central1-c
      filter: '(labels.executor_tag = codeintel-cloud)'
  relabel_configs:
    - source_labels: [__meta_gce_public_ip]
      target_label: __address__
      replacement: "${1}${2}:6060"
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
```

## AWS Deployment

See the [following Prometheus documentation](https://prometheus.io/docs/prometheus/latest/configuration/configuration/#ec2_sd_config) on how to allow EC2 Service Discovery.

The below snippet is an example target configuration. Adjust according to any changed parameters such as EC2 Region, Tag Values etc.
```yaml
- job_name: 'sourcegraph-code-intel-indexers'
  ec2_sd_configs:
    - region: us-west-2
      port: 6060
      filters:
        - name: tag:executor_tag
          values: [codeintel-cloud]
  relabel_configs:
    - source_labels: [__meta_ec2_public_ip]
      target_label: __address__
      replacement: "${1}${2}:6060"
      separator: ''
    - source_labels: [__meta_ec2_availability_zone]
      regex: ".+/([^/]+)"
      target_label: zone
      separator: ''
    - source_labels: [__meta_ec2_instance_id]
      target_label: instance
      separator: ''
    - regex: "__meta_gce_metadata_(image_.+)"
      action: labelmap
    - source_labels: [__meta_ec2_ami]
      target_label: version
```
