# How to convert repository groups to search contexts

This guide will provide steps for migrating from repository groups to [search contexts](../../code_search/explanations/features.md#search-contexts) on your Sourcegraph instance.

## Steps to convert repository groups to search contexts

1. Log in to your Sourcegraph instance.
1. Repository groups can be found in the settings configuration
    - To access your users settings, navigate to `https://your_sourcegraph_instance.com/user/settings` 
    - To access organization settings, navigate to `https://your_sourcegraph_instance.com/organizations/<organization_name>/settings` 
    - If you are a site-admin, navigate to `https://sourcegraph.com/site-admin/global-settings` to access global settings
1. Find the `search.repositoryGroups` object to view the current repository groups
1. For each of the repository groups you want to convert, do the following:
    - Navigate to `https://your_sourcegraph_instance.com/contexts`
    - Select the owner
    - Enter the repository group name as the context name
    - Optionally enter a description and choose a preferred visibility option
    - In the `Repositories and revisions` section enter the repositories from the repogroup
    - For each added repository you have to define an array of revisions to search
    - Keep a single `HEAD` revision if you only want to search the latest code on the main branch

### Converting repository group config to search contexts repositories config

For example, you have a repository group defined as: 
```json
{ "group": ["github.com/example/repo1", "github.com/example/repo2"] }
```

The equivalent search contexts repositories config would be:

```json
[
  { "repository": "github.com/example/repo1", "revisions": ["HEAD"] },
  { "repository": "github.com/example/repo2", "revisions": ["HEAD"] }
]
```

Converted search contexts can be used immediately by users on the Sourcegraph instance. The contexts selector will be shown in the search input.

## Discontinuing use of repository groups on your Sourcegraph instance

Once desired existing repository groups have been converted into search contexts, we recommend discontinuing use of repository groups.

To discontinue use of repository groups:

1. Navigate to your settings.
1. Locate the `search.repositoryGroups` object in the settings, and remove it
1. Save changes.
