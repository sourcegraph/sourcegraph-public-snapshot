# Settings

## Global settings

Settings provide the ability to customize and control the Sourcegraph UI and user-specific features. They do not configure operational aspects of the instance (which are set in the [site configuration](site_config.md)).

Settings can be set at the global level (by site admins), the organization level (by organization members), and at the individual user level.

<div class="text-center">
  <object data="settings-cascade.svg" type="image/svg+xml" style="width:80%;"></object>
</div>

<div class="text-center small">
  Non-sighted users can view a <a href="settings-cascade">text-representation of this diagram.</a>
</div>

### Editing global settings (for site admins)

Global settings are found in **Site admin > Global settings** while links to organization and user settings are found in the user dropdown menu.

After setting or changing certain values in **Site admin > Global settings** the frontend will restart automatically or
you might be asked to restart the frontend for the changes to take effect.
In case of a Kubernetes deployment this can be done as follows:

```shell script
kubectl delete pods -l app=sourcegraph-frontend
``` 

## Site config

Site configuration defines how various Sourcegraph features behave. See the [full reference](../reference/site_config.md) below for a comprehensive list of site configuration options.

### View and edit site configuration

Site admins can view and edit site configuration on a Sourcegraph instance:

1. Go to **User menu > Site admin**.
1. Open the **Configuration** page. (The URL is `https://sourcegraph.example.com/site-admin/configuration`.)
