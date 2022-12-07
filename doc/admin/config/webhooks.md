# Webhooks

Webhook receivers can be configured on a Sourcegraph instance in order to receive webhook events from code hosts. This allows Sourcegraph to react more quickly to events that occur outside of the instance. 

We currently support webhooks for the following:

Code host | [Batch changes](../../batch_changes/index.md) | Repository push
--------- | :-: | :-:
GitHub | 🟢 | 🟢 
GitLab | 🟢 | 🔴
Bitbucket Server / Datacenter | 🟢 | 🔴 
Bitbucket Cloud | 🟢 | 🔴

Webhooks need to be configured both on the sending side, the code host and receiveing side, Sourcegraph.


## Deprecation notice

As of Sourcegraph 4.3.0 webhooks added via code host configuration are deprecated and support will be removed in release 4.6.0.

This includes any webhooks pointed at URLs starting with the following:

* `.api/github-webhooks`
* `.api//gitlab-webhooks`
* `.api/bitbucket-server-webhooks`
* `.api/bitbucket-cloud-webhooks`

In order to continue using webhooks you need to follow the steps below to [add a receiver](#adding-a-receiver) and then update the webhook configured on your code host with the new webhook url which will look something like `https://sourcegraph-instance/.api/webhooks/{UUID}`

## Adding a receiver

Before adding a webhook receiver you should ensure that you have at least one [code host connection](../external_service) configured. 

In order to receive webhook events you need to add a receiver. The receiver will be configured to accept events from a specific code host connection based on it's type and URN.

1. Navigate to Site Admin > Incoming webhooks
1. Click `Add webhook`
1. You'll be presented with a form to create a new webhook receiver:
   1. Webhook name: Optional descriptive name for the webhook
   1. Code host type: Select from the dropdown. This will be filtered based on code host connections added on your instance. 
   1. Code host URN: The URN for the code host. Again, this will be filtered by code host connections added on your instance.
   1. Secret: If supported by the code host, this is an arbitrary shared secret between Sourcegraph and the code host. A default value is provided but you are free to change it.
1. Click `Create`

The receiver will now be created and you will be redirected to a page showing more details of the created webhook.

Most importantly, you will be presented with the unique URL for this webhook which is required when configuring the webhook on your code host.

## Webhook logging

Sourcegraph can track incoming webhooks from code hosts to more easily debug issues with webhook delivery. These webhooks can be viewed in two places depending on how they were added:

1. Via code host connection: **Site Admin > Batch Changes > Incoming webhooks**.
1. Via **Site Admin > Incoming webhooks**:  In the details are for each added receiver.

By default, sites without [database encryption](encryption.md) enabled will retain three days of webhook logs. Sites with encryption will not retain webhook logs by default, as webhooks may include sensitive information; these sites can enable webhook logging and optionally configure encryption for them by using the settings below.

### Enabling webhook logging

Webhook logging is controlled by the `webhook.logging` site configuration
option. This option is an object with the following keys:

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `enabled` | `boolean` | If `true`, incoming webhooks will be stored. | `true` if no site encryption is enabled; `false` otherwise. |
| `retention` | `string` | The length of time to retain the webhooks, expressed as a valid [Go duration](https://pkg.go.dev/time#ParseDuration). | `72h` |

#### Examples

To disable webhook logging:

```json
{
  "webhook.logging": {"enabled": false}
}
```

To retain webhook logs for one day:

```json
{
  "webhook.logging": {
    "enabled": false,
    "retention": "24h"
  }
}
```

### Encrypting webhook logs

Webhook logs can be encrypted by specifying a `webhookLogKey` in the [on-disk database encryption site configuration](encryption.md).



Recent received webhook payloads can be seen on the webhook details page for each receiver.
  
 > NOTE: Deprecated webhooks added via code host configuration can be viewed in **Site Admin > Batch Changes > Incoming webhooks**. Webhook logging can be configured through the [incoming webhooks site configuration](../../admin/config/batch_changes.md#incoming-webhooks).

## Configuring webhooks on the code host

The instructions for setting up webhooks on the code host are specific to each code host type.

### GitHub

#### Batch changes

1. Copy the webhook URL displayed after adding the receiver as mentioned [above](#adding-a-receiver)
1. On GitHub, go to the settings page of your organization. From there, click **Settings**, then **Webhooks**, then **Add webhook**.
1. Fill in the webhook form:
   * **Payload URL**: the URL you copied above from Sourcegraph.
   * **Content type**: this must be set to `application/json`.
   * **Secret**: the secret token you configured Sourcegraph to use above.
   * **Which events**: select **Let me select individual events**, and then enable:
     - Issue comments
     - Pull requests
     - Pull request reviews
     - Pull request review comments
     - Check runs
     - Check suites
     - Statuses
   * **Active**: ensure this is enabled.
1. Click **Add webhook**.
1. Confirm that the new webhook is listed.

Done! Sourcegraph will now receive webhook events from GitHub and use them to sync pull request events, used by [batch changes](../../batch_changes/index.md), faster and more efficiently.

#### Repository push

Follow the same steps as above, but ensure you include the `push` event under **Let me select individual events**

### GitLab

#### Batch changes

1. Copy the webhook URL displayed after adding the receiver as mentioned [above](#adding-a-receiver)
1. On GitLab, go to your project, and then **Settings > Webhooks** (or **Settings > Integration** on older GitLab versions that don't have the **Webhooks** option).
1. Fill in the webhook form:
   * **URL**: the URL you copied above from Sourcegraph.
   * **Secret token**: the secret token you configured Sourcegraph to use above.
   * **Trigger**: select **Merge request events** and **Pipeline events**.
   * **Enable SSL verification**: ensure this is enabled if you have configured SSL with a valid certificate in your Sourcegraph instance.
1. Click **Add webhook**.
1. Confirm that the new webhook is listed below **Project Hooks**.

Done! Sourcegraph will now receive webhook events from GitLab and use them to sync merge request events, used by [batch changes](../../batch_changes/index.md), faster and more efficiently.

**NOTE:** We currently do not support [system webhooks](https://docs.gitlab.com/ee/administration/system_hooks.html) as these provide a different set of payloads.

### Bitbucket server

#### Batch changes

The [Sourcegraph Bitbucket Server plugin](../../integration/bitbucket_server.md#sourcegraph-bitbucket-server-plugin) enables the Bitbucket Server / Bitbucket Data Center instance to send webhooks to Sourcegraph.

1. Install the [Sourcegraph Bitbucket Server plugin](../../integration/bitbucket_server.md#sourcegraph-bitbucket-server-plugin) on your Bitbucket Server / Bitbucket Data Center instance.
1. On your Bitbucket Server / Bitbucket Data Center instance, go to **Administration > Add-ons > Sourcegraph**
1. Fill in the **Add a webhook** form
   * **Name**: A unique name representing your Sourcegraph instance
   * **Scope**: `global`
   * **Endpoint**: The URL found after creating a webhook receiver
   * **Events**: `repo:build_status`, `pr:activity:status`, `pr:activity:event`, `pr:activity:rescope`, `pr:activity:merge`, `pr:activity:comment`, `pr:activity:reviewers`, `pr:participant:status`
   * **Secret**: The secret you configured when creating the webhook receiver
1. Confirm that the new webhook is listed under **All webhooks** with a timestamp in the **Last successful** column.

Done! Sourcegraph will now receive webhook events from Bitbucket Server / Bitbucket Data Center and use them to sync pull request events, used by [batch changes](../../batch_changes/index.md), faster and more efficiently.

### Bitbucket cloud

#### Batch changes

> NOTE: Experimental webhook support for Bitbucket Cloud was added in Sourcegraph 3.40. Please <a href="https://about.sourcegraph.com/contact">contact us</a> with any issues found while using webhooks.

1. On Bitbucket Cloud, go to each repository, and then **Repository settings > Webhooks**.
1. Click **Add webhook**.
1. Fill in the webhook form:
   * **Title**: any title.
   * **URL**: the URL found after creating a webhokk receiver
   * **Triggers**: select **Build status created** and **Build status updated** under **Repository**, and every item under **Pull request**.
1. Click **Save**.
1. Confirm that the new webhook is listed below **Repository hooks**.

Done! Sourcegraph will now receive webhook events from Bitbucket Cloud and use them to sync pull request events, used by [batch changes](../../batch_changes/index.md), faster and more efficiently.
