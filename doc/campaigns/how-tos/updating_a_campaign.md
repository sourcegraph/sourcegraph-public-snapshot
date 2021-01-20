# Updating a campaign

Updating a campaign works by applying a campaign spec to an **existing** campaign in the same namespace.

Since campaigns are uniquely identified by their name (the `name` property in the campaign spec) and the namespace in which they were created, you can edit any other part of a campaign spec and apply it again.

When a new campaign spec is applied to an existing campaign the existing campaign is updated and its changesets are updated to match the new desired state.

## Requirements 

To update a changeset, you need:

1. [admin permissions for the campaign](../explanations/permissions_in_campaigns.md#permission-levels-for-campaigns),
1. write access to the changeset's repository (on the code host), and
1. a personal access token [configured in Sourcegraph for your code host(s)](configuring_user_credentials.md).

For more information, see [Code host interactions in campaigns](../explanations/permissions_in_campaigns.md#code-host-interactions-in-campaigns).

## Preview and apply a new campaign spec

In order to update a campaign after previewing the changes, do the following:

1. Edit the [campaign spec](../references/campaign_spec_yaml_reference.md) with which you created the campaign to include the changes you want to make to the campaign. For example, change the [`description`](../references/campaign_spec_yaml_reference.md#description) of the campaign or change [the commit message in the `changesetTemplate`](../references/campaign_spec_yaml_reference.md#changesettemplate-commit-message).
1. Use the [Sourcegraph CLI (`src`)](https://github.com/sourcegraph/src-cli) to execute and upload the campaign spec.

    <pre><code>src campaign preview -f <em>YOUR_CAMPAIGN_SPEC.campaign.yaml</em></code></pre>
1. Open on the URL that's printed to preview the changes that will be made by applying the new campaign spec.
1. Click **Apply spec** to update the campaign.

All of the changesets on your code host will be updated to the desired state that was shown in the preview.

## Apply a new campaign spec directly

In order to update a campaign directly, without preview, do the following:

1. Edit the [campaign spec](../references/campaign_spec_yaml_reference.md) with which you created the campaign to include the changes you want to make to the campaign.
1. Use the [Sourcegraph CLI (`src`)](https://github.com/sourcegraph/src-cli) to execute, upload, and the campaign spec.

    <pre><code>src campaign apply -f <em>YOUR_CAMPAIGN_SPEC.campaign.yaml</em></code></pre>

The new campaign spec will be applied directly and the campaign and its changesets will be updated.

## How campaign updates are processed

Changes in the campaign spec that that affect campaign itself, such as the [`description`](../references/campaign_spec_yaml_reference.md#description), are applied directly when you apply the new campaign spec.

Changes that affect the changesets are processed asynchronously to update the changeset to its new desired state. Different fields are processed differently.

Here are some examples:

- When the diff or attributes that affect the resulting commit of a changeset directly (such as the [`changesetTemplate.commit.message`](../references/campaign_spec_yaml_reference.md#changesettemplate-commit-message) or the [`changesetTemplate.commit.author`](../references/campaign_spec_yaml_reference.md#changesettemplate-commit-author)) and the changeset has been published, the commit on the code host will be overwritten by a new commit that includes the updated diff.
- When the the [`changesetTemplate.title`](../references/campaign_spec_yaml_reference.md#changesettemplate-title) or the [`changesetTemplate.body`](../references/campaign_spec_yaml_reference.md#changesettemplate-commit-author) are changed and the changeset has been published, the changeset on the code host will be updated accordingly.
- When the [`changesetTemplate.branch`](../references/campaign_spec_yaml_reference.md#changesettemplate-title) is changed after the changeset has been published on the code host, the existing changeset will be closed on the code host and new one, with the new branch, will be created.
- When the campaign spec is changed in such a way that no diff is produced in a repository in which the campaign already created and published a changeset, the existing changeset will be closed on the code host and detached from the campaign.

See the "[Campaigns design](../explanations/campaigns_design.md)" doc for more information on the declarative nature of the campaigns system.

## Updating a campaign to change its scope

### Increasing the number of changesets

You can gradually increase the number of repositories to which a campaign applies by modifying the entries in the [`on`](../references/campaign_spec_yaml_reference.md#on) property of the campaign spec.

That means you can start with an `on` property like this in your campaign spec:

```yaml
# [...]

# Find all repositories that contain a README.md file, in the GitHub my-company org.
on:
  - repositoriesMatchingQuery: file:README.md repo:github.com/my-company

# [...]
```

After you applied that campaign spec, you can extend the scope of campaign by changing the `on` property to result in more repositories:

```yaml
# [...]

# Find all repositories that contain a README.md file, in the GitHub my-company and my-company-ci org.
on:
  - repositoriesMatchingQuery: file:README.md repo:github.com/my-company|github.com/my-company-ci

# [...]
```

The updated [`repo:` keyword](../../code_search/reference/queries.md#keywords-all-searches) in the search query will result in more repositories being returned by the search.

If you apply the updated campaign spec new changesets will be created for each additional repository.

### Decreasing the number of changesets

You can also decrease the number of repositories to which a campaign applies the by modifying the entries in the [`on`](../references/campaign_spec_yaml_reference.md#on) property.

If you, for example, started with this campaign spec

```yaml
# [...]

# Find all repositories that contain a README.md file, in the GitHub my-company org.
on:
  - repositoriesMatchingQuery: file:README.md repo:github.com/my-company

# [...]
```

and applied it and [published changesets](publishing_changesets.md) and then change it to this

```yaml
# [...]

# Find all repositories that contain a README.md file, in the GitHub my-company org.
on:
  - repositoriesMatchingQuery: file:README.md repo:github.com/my-company/my-one-repository

# [...]
```

and apply it, then all the changesets that were published in repositories other than `my-one-repository` _will be closed on the code host and detached from the campaign_.
