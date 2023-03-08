# Azure DevOps

Site admins can sync Git repositories hosted on [Azure DevOps](https://dev.azure.com) with Sourcegraph so that users can search and navigate the repositories.

To connect Azure DevOps to Sourcegraph:

1. Go to **Site admin > Manage code hosts > Add repositories**.
2. Select **Azure DevOps**.
3. Next you will have to provide a [configuration](#configuration) for the Azure DevOps code host connection. Here is an example configuration:
```json
{
  "url": "https://dev.azure.com/",
  "username": "<admin username>",
  "token": "<admin token>",
  "projects": [
    "org1/project1"
  ],
  "orgs": [
    "org2"
  ]
}
```
4. Press **Add repositories**.

## Repository syncing

Currently, all repositories belonging to the configured organizations/projects will be synced.

In addition, there is one more field for configuring which repositories are mirrored:

- [`exclude`](azuredevops.md#configuration)<br>A list of repositories to exclude.

### HTTPS cloning

Sourcegraph clones repositories from Azure DevOps via HTTP(S), using the [`username`](azuredevops.md#configuration) and [`token`](azuredevops.md#configuration) required fields you provide in the configuration.

## Configuration

Azure DevOps connections support the following configuration options, which are specified in the JSON editor in the site admin "Manage code hosts" area.

<div markdown-func=jsonschemadoc jsonschemadoc:path="admin/external_service/azuredevops.schema.json">[View page on docs.sourcegraph.com](https://docs.sourcegraph.com/admin/external_service/azuredevops) to see rendered content.</div>

## Webhooks

Please consult [this page](../config/webhooks.md) in order to configure webhooks.
