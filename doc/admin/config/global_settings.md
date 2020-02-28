# Global settings

Settings provide the ability to customize and control the Sourcegraph UI and user specific features and do not configure operational aspects of the instance.

Settings are also different to [site configuration](site_config.md) as they are set the global level by site admins, the organization level by organization members, and at the individual user level.

<div class="text-center">
  <object data="settings-cascade.svg" type="image/svg+xml" style="width:80%;"></object>
</div>

<div class="text-center small">
  Non-sighted users can view a <a href="settings-cascade">text-representation of this diagram.</a>
</div>

## View and edit settings

Global settings are found in **Site admin > Global settings** while links to organization and user settings are found in the user dropdown menu.

## Reference

Settings options and their default values are shown below.

<div markdown-func=jsonschemadoc jsonschemadoc:path="admin/config/settings.schema.json">[View page on docs.sourcegraph.com](https://docs.sourcegraph.com/doc/admin/config/settings) to see rendered content.</div>
