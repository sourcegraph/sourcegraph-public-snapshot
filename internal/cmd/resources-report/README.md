# Resources report tool

This tool reports on the status of various resources in AWS and GCP accounts. Credentials are expected to be set up beforehand, and leverage default credentials of each supported platform. Basic usage:

```sh
go build && ./resources-report --aws --gcp --slack.webhook="https://hooks.slack.com/services/xxxxxxxxx/xxxxxxxxxxx/xxxxxxxxxxxxxxxxxxxxxxxx"
```

## Authentication

### Google Cloud Platform

Credentials should be a GCP service account with access to the following permissions in all relevant projects:

- `Viewer`
- `Cloud Asset Viewer`

The key should be accessible in `GOOGLE_APPLICATION_CREDENTIALS`.

### Amazon Web Services

Credentials should be an AWS IAM with the following permissions:

- `ReadOnlyAccess`

Credentials should be set in `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY`.

## Resources

- GCP: declare resource types to query in [`gcp.go`](./gcp.go)'s `gcpAssetTypes` variable.
- AWS: Declare queries for resources as functions in [`aws.go`](./aws.go)'s `awsResources` variable.
