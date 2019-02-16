# Site configuration

Site configuration defines how various Sourcegraph features behave. See the [full reference](#reference) below for a comprehensive list of site configuration options.

## View and edit site configuration

Site admins can view and edit site configuration on a Sourcegraph instance:

1. Go to **User menu > Site admin**.
1. Open the **Configuration** page. (The URL is `https://sourcegraph.example.com/site-admin/configuration`.)

Some options, such as the external URL and user authentication, are considered [critical configuration](critical_config.md) and must be edited in the [management console](../management_console.md).

## Reference

All site configuration options and their default values are shown below.

> NOTE: Not finding the option you're looking for? It may be a [critical configuration](critical_config.md) option, which means it must be set in the [management console](../management_console.md).

<div markdown-func=jsonschemadoc jsonschemadoc:path="admin/config/site.schema.json">[View page on docs.sourcegraph.com](https://docs.sourcegraph.com/admin/config/site_config) to see rendered content.</div>
