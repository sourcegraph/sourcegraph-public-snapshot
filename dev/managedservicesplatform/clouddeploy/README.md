```
gcloud deploy apply --project=msp-testbed-robert-7be9 --region=us-central1 --file=clouddeploy.yaml
```

https://cloud.google.com/sdk/gcloud/reference/deploy/releases/create

```
gcloud deploy releases create manual-test-15-2024-01-29 \
    --project=msp-testbed-robert-7be9 \
    --region=us-central1 \
    --delivery-pipeline=msp-testbed-us-central1-rollout \
    --source='gs://msp-testbed-robert-7be9-cloudrun-skaffold/source.tar.gz' \
    --deploy-parameters="customTarget/tag=259633_2024-01-29_5.2-736af892a393"
```
