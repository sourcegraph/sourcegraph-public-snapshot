# Administration and security of Code Insights

## Enable Code Insights for all users 

To enable code insights for all users – which adds an "Insights" link in the Sourcegraph navigation bar and lets others create their own insights – add the following to your organization settings at `sourcegraph.example.com/organizations/[your_org]/settings`:

`"experimentalFeatures": { "codeInsights": true }`

## Code Insights enforce user permissions 

Users can only create code insights that include repositories they have access to. Moreover, when creating code insights, the repository "autosuggest" field will *not* show users repositories they would not have access to otherwise. TODO TODO confirm

## Security of native Sourcegraph Code Insights

Sourcegraph search-based and language insights run natively on your Sourcegraph instance using your instance's Sourcegraph search API. 

## Security of Sourcegraph extension-provided Code Insights

Sourcegraph extension-provided insights adhere to the same security standards as any other Sourcegraph extension. Refer to [Security and privacy of Sourcegraph extensions](../../extensions/security.md). 

## Disable Sourcegraph-extension-provided Code Insights 

If you want to disable Sourcegraph-extension-provided code insights, you can do so the same way you would disable any other extension. Refer to [Disabling remote extensions](../../admin/extensions.md#use-extensions-from-sourcegraph-com-or-disable-remote-extensions) and [Allow only specific extensions](../../admin/extensions.md#use-extensions-from-sourcegraph-com-or-disable-remote-extensions)



