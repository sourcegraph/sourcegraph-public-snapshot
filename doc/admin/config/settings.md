# Settings

Settings provide the ability to customize and control the Sourcegraph UI and user-specific features. They do not configure operational aspects of the instance (which are set in the [site configuration](site_config.md)).

Settings can be set at the global level (by site admins), the organization level (by organization members), and at the individual user level.

<div class="text-center">
  <object data="settings-cascade.svg" type="image/svg+xml" style="width:80%;"></object>
</div>

<div class="text-center small">
  Non-sighted users can view a <a href="settings-cascade">text-representation of this diagram.</a>
</div>

## Editing global settings (for site admins)

Global settings are found in **Site admin > Global settings** while links to organization and user settings are found in the user dropdown menu.

After setting or changing certain values in **Site admin > Global settings** the frontend will restart automatically or
you might be asked to restart the frontend for the changes to take effect.
In case of a Kubernetes deployment this can be done as follows:

```
bash
kubectl delete pods -l app=sourcegraph-frontend
``` 

## Reference

Settings options and their default values are shown below.

[Sourcegraph extensions](../../extensions/index.md) can also define new settings properties. Check the documentation of an extension to see what settings it defines, or consult the `contributes.configuration` object in the extension's `package.json` (e.g., for the [Codecov extension](https://sourcegraph.com/github.com/codecov/sourcegraph-codecov@560595f0dab5dfb54f5da8be95e685dd2d88c2cf/-/blob/package.json#L178)).

## Additional details on settings

### Notices

Notices can be added in global, organization, or user settings. The `notices` setting is a list of configuration consisting of three elements:

1. `message`: the markdown copy to be displayed in the banner
1. `location`: where the banner will be shown. Either on the home page with `"home"` or at the top of the page with `"top"`
1. `dismissable`: boolean (`true` or `false`). If true, users will be able to close the notice and not see it again. If false, it will persist on the instance until the configuration is removed.

### Example settings:

```
"notices": [
      {
        "message": "Your message here! [Include a link for more information]().",
        "location": "top",
        "dismissible": true
      }
    ]
```
