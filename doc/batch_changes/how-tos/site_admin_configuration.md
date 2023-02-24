# Site admin configuration for Batch Changes

#### Setup Batch Changes 

1. Using Batch Changes requires a [code host connection](../../admin/external_services/index.md) to a supported code host (currently GitHub, Bitbucket Server / Bitbucket Data Center, GitLab, and Bitbucket Cloud).

1. (Optional) [Configure repository permissions](../../admin/repo/permissions.md), which Batch Changes will respect.

1. [Configure credentials](configuring_credentials.md).

1. [Setup webhooks](../../admin/config/webhooks.md) to make sure changesets sync fast. See [Batch Changes effect on codehost rate limits](../references/requirements.md#batch-changes-effect-on-code-host-rate-limits).

1. Configure any desired optional features, such as:
    * [Rollout windows](../../../admin/config/batch_changes#rollout-windows), which control the rate at which Batch Changes will publish changesets on code hosts.
    * [Forks](../../../admin/config/batch_changes#forks), which push branches created by Batch Changes onto forks of the upstream repository instead than the repository itself.

#### Disable Batch Changes
- [Disable Batch Changes](../explanations/permissions_in_batch_changes.md#disabling-batch-changes).
- [Disable Batch Changes for non-site-admin users](../explanations/permissions_in_batch_changes.md#disabling-batch-changes-for-non-site-admin-users).
