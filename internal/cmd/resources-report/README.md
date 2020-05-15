# Resources report tool

This tool reports on the status of various resources in AWS and GCP accounts. Credentials are expected to be set up beforehand, and leverage default credentials of each supported platform.

## Google Cloud Platform

Enable reporting with `--gcp`.

Credentials should be a GCP service account with access to the following permissions in all relevant projects:

- `Viewer`
- `Cloud Asset Viewer`

The key should be accessible in `GOOGLE_APPLICATION_CREDENTIALS`.
