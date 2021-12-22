# Administration and security of Code Insights

## Code Insights enforce user permissions 

Users can only create code insights that include repositories they have access to. Moreover, when creating code insights, the repository field will *not* validate nor show users repositories they would not have access to otherwise. 

## Security of native Sourcegraph Code Insights (Search-based and Language Insights)

Sourcegraph search-based and language insights run natively on a Sourcegraph instance using the instance's Sourcegraph search API. This means they don't send any information about your code to third-party servers. 

## Security of Sourcegraph extension-provided Code Insights

Sourcegraph extension-provided insights adhere to the same security standards as any other Sourcegraph extension. Refer to [Security and privacy of Sourcegraph extensions](../../extensions/security.md). 

If you are concerned about the security of extension-provided insights, then you can: 

## Disable Sourcegraph extension-provided Code Insights 

If you want to disable Sourcegraph-extension-provided code insights, you can do so the same way you would disable any other extension. Refer to [Disabling remote extensions](../../admin/extensions.md#use-extensions-from-sourcegraph-com-or-disable-remote-extensions) and [Allow only specific extensions](../../admin/extensions.md#use-extensions-from-sourcegraph-com-or-disable-remote-extensions). 