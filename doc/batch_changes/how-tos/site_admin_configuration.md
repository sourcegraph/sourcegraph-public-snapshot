# Site admin configuration for Batch Changes



#### Setup Batch Changes 
1. Using Batch Changes requires a [code host connection](../../../admin/external_service/index.md) to a supported code host (currently GitHub, Bitbucket Server, and GitLab).
1. [Configure repository permissions](../../../admin/repo/permissions.md), which Batch Changes will respect.
1. [Configure credentials](configuring_credentials.md).
1. Setup webhooks to make sure changesets sync fast. See [Batch Changes effect on codehost rate limits](../references/requirements.md#batch-changes-effect-on-code-host-rate-limits).
1. (Optional) [Control the rate at which Batch Changes will publish changesets on code hosts](../../../admin/config/batch_changes.md#rollout-windows).

#### Disable Batch Changes
- [Disable Batch Changes](../explanations/permissions_in_batch_changes.md#disabling-batch-changes).
- [Disable Batch Changes for non-site-admin users](../explanations/permissions_in_batch_changes.md#disabling-batch-changes-for-non-site-admin-users).

