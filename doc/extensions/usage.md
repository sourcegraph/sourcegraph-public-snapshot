# Using Sourcegraph extensions

> NOTE: Sourcegraph extensions are being deprecated with the upcoming Sourcegraph September release. [Learn more](./deprecation.md).

## Usage

To view all available extensions on your Sourcegraph instance, click **User menu > Extensions** in the top navigation bar.

> Don't see an extension you need? You can [share your idea](https://github.com/sourcegraph/sourcegraph/issues?q=is%3Aopen+is%3Aissue+label%3Aextension-request) with an issue labeled `extension-request`.

To enable/disable an extension for yourself, click **User menu > Extensions**, find the extension, and toggle the slider.

After enabling a Sourcegraph extension, it is immediately ready to use. Of course, some extensions only activate for certain files (e.g., the Python extension only adds code navigation for `.py` files).

### On your code host

Install the [browser extension](../integration/browser_extension.md) and point it to your Sourcegraph instance (or Sourcegraph.com) in its options menu. It will consult your Sourcegraph user settings and activate the extensions you've enabled whenever you view code or diffs on your code host or review tool.

### For organizations

To enable/disable an extension for all organization members, add it to the `extensions` object in organization settings (as shown below).

```json
{
  ...,
  "extensions": {
    ...,
    "alice/myextension": true, // or false to disable
    ...
  },
  ...
}
```

### For all users

On a self-hosted Sourcegraph instance, add the same JSON above to global settings (in **Site admin > Global settings**).

<div style="text-align:center;margin:20px 0;display:flex">
<img src="https://sourcegraphstatic.com/docs/images/extensions/all-users-global-settings.png" style="padding:15px"></a>
</div>
