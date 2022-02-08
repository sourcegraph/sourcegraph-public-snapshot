# Administration and security of Code Insights

## Code Insights enforce user permissions 

Users can only create code insights that include repositories they have access to. Moreover, when creating code insights, the repository field will *not* validate nor show users repositories they would not have access to otherwise.

When a user is viewing an insight, any repositories they do not have access to will not be included in the total counts.

## Security of native Sourcegraph Code Insights (Search-based and Language Insights)

Sourcegraph search-based and language insights run natively on a Sourcegraph instance using the instance's Sourcegraph search API. This means they don't send any information about your code to third-party servers. 

## Security of Sourcegraph extension-provided Code Insights

Sourcegraph extension-provided insights adhere to the same security standards as any other Sourcegraph extension. Refer to [Security and privacy of Sourcegraph extensions](../../extensions/security.md). 

If you are concerned about the security of extension-provided insights, then you can: 

## Disable Sourcegraph extension-provided Code Insights 

If you want to disable Sourcegraph-extension-provided code insights, you can do so the same way you would disable any other extension. Refer to [Disabling remote extensions](../../admin/extensions.md#use-extensions-from-sourcegraph-com-or-disable-remote-extensions) and [Allow only specific extensions](../../admin/extensions.md#use-extensions-from-sourcegraph-com-or-disable-remote-extensions).

## Insight and Dashboard permissions

Note: there are no separate read/write permissions. If a user can view an insight or dashboard, they can also edit it.

A user can view an insight if at least one of the following is true:

1. The user created the insight.
2. The user has permission to view a dashboard that the insight is on.

Except for the singular, non-transferable creator's permission noted in case 1 above, permissions can be thought of as belonging to a dashboard. Dashboards have 3 permission levels:

- User: only this specific user can view this dashboard.
- Organization: only users within this organization can view this dashboard.
- Global: any user on the Sourcegraph instance can view this dashboard.

### Changing permission levels

Because there are no separate read/write permissions and no dashboard owners, any user who can view a dashboard can change its permission level or add/remove insights from the dashboard. The only way to guarantee continued access to an insight that you did not create is to add it to a private dashboard.

If a user gets deleted, any insights they created will still be visible to other users via the dashboards they appear on. However, if one of these insights is removed from all dashboards, it will no longer be accessible.
