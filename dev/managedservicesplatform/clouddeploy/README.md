```
gcloud deploy apply --project=msp-testbed-robert-7be9 --region=us-central1 --file=rollout.clouddeploy.yaml
```

https://cloud.google.com/sdk/gcloud/reference/deploy/releases/create

```
gcloud deploy releases create manual-test-04-2024-01-31 \
    --project=msp-testbed-robert-7be9 \
    --region=us-central1 \
    --delivery-pipeline=msp-testbed-us-central1-rollout \
    --source='gs://msp-testbed-robert-7be9-cloudrun-skaffold/source.tar.gz' \
    --labels="commit=abc123,author=foo" \
    --deploy-parameters="customTarget/tag=dd34d1be076e_2024-01-31"
```
