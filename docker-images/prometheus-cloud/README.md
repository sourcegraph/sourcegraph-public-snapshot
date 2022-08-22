# `prometheus-cloud`
This is a patched version of Sourcegraoh's Prometheus image using an alternative base image provided by GCP to allow automatically shipping metrics to the Google Cloud Managed Prometheus Service.

This image is built by calling the `build.sh` function in [`prometheus`](../prometheus/README.md).

https://cloud.google.com/stackdriver/docs/managed-prometheus/setup-unmanaged#run-gmp
