# How to deploy a new executor image

This guide documents how to deploy a new image of [executors](../../../admin/executors.md) to the following [Sourcegraph instances](https://handbook.sourcegraph.com/departments/product-engineering/engineering/process/deployments/instances):

* [Sourcegraph Cloud](https://sourcegraph.com)
* [k8s.sgdev.org](https://k8s.sgdev.org)

## Requirements

* Clone of [`infrastructure`](https://github.com/sourcegraph/infrastructure) repository
* `terraform` in the version specified in [executors/.tool-versions](https://github.com/sourcegraph/infrastructure/blob/main/executors/.tool-versions)
  * Using `asdf`: `asdf install terraform x.x.x`
* Authenticated with GCP: `gcloud auth application-default login`
* AWS credentials set as env vars: `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`

## Steps

1. Make a change to the `executor` code so that the buildkite build results in new `executor` AWS/GCP images being built & published. Or push to a branch with the [`executor-patch-notest/`](https://github.com/sourcegraph/sourcegraph/blob/882ed49014bc470a3be17ce74764a856f82bee4e/enterprise/dev/ci/internal/ci/runtype.go#L65-L67) prefix to trigger the build.
1. Look at the buildkite build. Example: https://buildkite.com/sourcegraph/sourcegraph/builds/116966

    There are two steps with `executor-image` in here. The first one builds and uploads image. Second step releases it.

    Try to find the GCP image name, i.e. `executor-cc28c728e5-116966`

    Try to find the AWS AMI name, i.e. `ami-0fb21656aeba5eb7c`

2. In the `infrastructure` repository:
  * Open `executors/gcp.tf` and update the image at the top in `gcp_executor_machine_image`.
  * open `executors/aws.tf` and update the image at the top in `aws_executor_ami`.
3. Create a pull request with that change.
4. Get approval for PR.
5. In that PR branch, in the `executors` folder, run: `terraform apply`.
6. Open https://k8s.sgdev.org/site-admin/executors to see that the new version is used.
7. Merge PR.
