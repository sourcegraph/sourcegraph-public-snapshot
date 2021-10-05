# Deploying Sourcegraph executors

>NOTE: **Sourcegraph executors are currently experimental.** We're exploring this feature set. 
>Let us know what you think! [File an issue](https://github.com/sourcegraph/sourcegraph/issues/new/choose)
>with feedback/problems/questions, or [contact us directly](https://about.sourcegraph.com/contact).

TODO

- [Google Cloud](https://github.com/sourcegraph/terraform-google-executors)
- [AWS](https://github.com/sourcegraph/terraform-aws-executors)

## Configuring executors and instance communication

TODO

Frontend deployment config:

```yaml
- name: EXECUTOR_FRONTEND_USERNAME
  value: TODO
- name: EXECUTOR_FRONTEND_PASSWORD
  value: TODO
```

TODO

## Configuring auto scaling

TODO

Worker deployment config:

```yaml
- name: EXECUTOR_METRIC_GOOGLE_APPLICATION_CREDENTIALS_FILE
  value: /secrets/codeintel/service_account.json
- name: EXECUTOR_METRIC_GCP_PROJECT_ID
  value: sourcegraph-code-intel
- name: EXECUTOR_METRIC_ENVIRONMENT_LABEL
  value: cloud
- name: EXECUTOR_METRIC_AWS_NAMESPACE
  value: sourcegraph-executor
- name: EXECUTOR_METRIC_AWS_REGION
  value: us-west-2
- name: EXECUTOR_METRIC_AWS_ACCESS_KEY_ID
  valueFrom:
    secretKeyRef:
      name: frontend-secrets
      key: EXECUTOR_METRIC_AWS_ACCESS_KEY_ID
- name: EXECUTOR_METRIC_AWS_SECRET_ACCESS_KEY
  valueFrom:
    secretKeyRef:
      name: frontend-secrets
      key: EXECUTOR_METRIC_AWS_SECRET_ACCESS_KEY
```

TODO

## Configuring observability

Prometheus deployment config:

```yaml
    # Kubernetes-external `services
    # Configuration reference: https://prometheus.io/docs/prometheus/latest/configuration/configuration/#gce_sd_config
    # Indexer services in the sourcegraph-code-intel project
    - job_name: 'sourcegraph-code-intel-indexers'
      gce_sd_configs:
        - project: sourcegraph-code-intel
          port: 6060
          zone: us-central1-c
          filter: '(labels.executor_tag = codeintel-cloud)'
      ec2_sd_configs:
        - region: us-west-2
          port: 6060
          filters:
            - name: tag:executor_tag
              values: [codeintel-cloud]
      relabel_configs:
        - source_labels: [__meta_gce_public_ip, __meta_ec2_public_ip]
          target_label: __address__
          replacement: "${1}${2}:6060"
          separator: ''
        - source_labels: [__meta_gce_zone, __meta_ec2_availability_zone]
          regex: ".+/([^/]+)"
          target_label: zone
          separator: ''
        - source_labels: [__meta_gce_project]
          target_label: project
        - source_labels: [__meta_gce_instance_name, __meta_ec2_instance_id]
          target_label: instance
          separator: ''
        - regex: "__meta_gce_metadata_(image_.+)"
          action: labelmap
        - source_labels: [__meta_ec2_ami]
          target_label: version
```

TODO
