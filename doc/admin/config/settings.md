# Settings

Settings provide the ability to customize and control the Sourcegraph UI and user-specific features. They do not configure operational aspects of the instance (which are set in the [site configuration](site_config.md)).

Settings can be set at the global level (by site admins), the organization level (by organization members), and at the individual user level.

<div class="text-center">
  <object data="settings-cascade.svg" type="image/svg+xml" style="width:80%;"></object>
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

<div markdown-func=jsonschemadoc jsonschemadoc:path="admin/config/settings.schema.json">[View page on docs.sourcegraph.com](https://docs.sourcegraph.com/doc/admin/config/settings) to see rendered content.</div>

## Additional details on settings

### Notices

Notices can be added in global, organization, or user settings. The `notices` setting is a list of configuration consisting of the following elements:

1. `message`: the markdown copy to be displayed in the banner
1. `location`: where the banner will be shown. Either on the home page with `"home"` or at the top of the page with `"top"`
1. `dismissible (optional)`: boolean (`true` or `false`). If true, users will be able to close the notice and not see it again. If false, it will persist on the instance until the configuration is removed.
1. `variant (optional)`: one of `"primary"`, `"secondary"`, `"success"`, `"danger"`, `"warning"`, `"info"`, `"note"` the style of the notice. Although specifics such as color depend on the theme in general `danger` or `primary` will draw more attention than `secondary` or `note`.
The default style depends on the location of the notice.
1. `styleOverrides (optional)`: a configuration object with the following elements:
   1. `backgroundColor (optional)`: a hexadecimal color code for forcing a specific background color.
   1. `textColor (optional)`: a hexadecimal color code for forcing a specific text color.
   1. `textCentered (optional)`: boolean (`true` or `false`). If true, the text will be centered in the banner.

### Example settings:

```json
"notices": [
  {
    "message": "Your important message here! [Include a link for more information](http://example.com).",
    "location": "top",
    "dismissible": true,
    "variant": "danger",
    "styleOverrides": {
      "styleOverrides": {
        "backgroundColor": "#7f1d1d",
        "textColor": "#fecaca",
        "textCentered": true
      }
    }
  }
]
```
