# `prometheus-gcp`

This is a patched version of Sourcegraoh's Prometheus image using an alternative base image provided by GCP to allow automatically shipping metrics to the Google Cloud Managed Prometheus Service.

Please note: using this image will automatically begin shipping Prometheus metrics to the Google Cloud Managed Prometheus Service and may incur additional cloud hosting costs.

This image is built by calling the [`build.sh` script in `prometheus`](../prometheus/build.sh) with some additional parameters.

For more details about Google's forked Prometheus server, see: https://cloud.google.com/stackdriver/docs/managed-prometheus/setup-unmanaged#run-gmp
