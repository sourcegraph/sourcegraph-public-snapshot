# Developing campaigns

## What are campaigns?

Before diving into the technical part of campaigns, make sure to read up on what campaigns are, what they're not and what we want them to be.

1. Start by looking at the product page for [code change management](https://about.sourcegraph.com/product/code-change-management)
1. Read through the first page of the [campaigns documentation](https://docs.sourcegraph.com/user/campaigns/) **IMPORTANT:** Watch the video!

## Starting up your environment

1. Run `./enterprise/dev/start.sh` — Wait until all repositories are cloned.
2. Follow the [campaigns "Getting started" guide](../user/campaigns/getting_started.md) to setup campaigns.
3. Create [your first campaign](../user/campaigns/creating_campaign_from_patches.md). **Remember:** If you create a campaign, you're opening real PRs on GitHub. Make sure only [testing repositories](#github-testing-account) are affected. If you create a large campaign, it takes a while to preview/create but also helps a lot with finding bugs/errors, etc.

## Glossary

The campaigns feature introduces a lot of new names, GraphQL queries and mutations and database tables. This section tries to explain the most common names and provide a mapping between the GraphQL types and their internal counterpart in the Go backend.

| GraphQL type        | Go type              | Database table     | Description |
| ------------------- | -------------------- | -------------------| ----------- |
| `Campaign`          | `campaigns.Campaign`       | `campaigns`        | A campaign is a collection of changesets. The central entity. |
| `ExternalChangeset` | `campaigns.Changeset`      | `changesets`       | Changeset is the unified name for pull requests/merge requests/etc. on code hosts.        |
| `PatchSet`          | `campaigns.PatchSet`       | `patch_sets`       | A patch set is a collection of patches that will be applied by creating and publishing a campaign. A campaign *has one* patch set. |
| `Patch`             | `campaigns.Patch`          | `patches`          | A patch for a repository that *can* be turned into a changeset on a code host. It belongs to a patch set, which has multiple patches, one per repository. |
| -                   | `campaigns.ChangesetJob`   | `changeset_jobs`   | It represents the process of turning a `Patch` (GraphQL)/`campaigns.Patch` (Go) into a `Changeset` on the code host. It is executed asynchronously in the background when a campaign is created with a patch set. |
| `ChangesetEvent`    | `campaigns.ChangesetEvent` | `changeset_events` | A changeset event is an event on a code host, e.g. a comment or a review on a pull request on GitHub. They are created by syncing the changesets from the code host on a regular basis and by accepting webhook events and turning them into changeset events. |

## Database layout

<object data="/dev/campaigns_database_layout.svg" type="image/svg+xml" style="width:100%; max-width: 800px">
</object>

(To re-generate the diagram from the `campaigns_database_layout.dot` file with Graphviz, run: `dot -Tsvg -o campaigns_database_layout.svg campaigns_database_layout.dot`.)

## Diving into the code as a backend developer

1. Read through `./cmd/frontend/graphqlbackend/campaigns.go` to get an overview of the campaigns GraphQL API.
1. Read through `./internal/campaigns/types.go` to see all campaigns-related type definitions.
1. Compare that with the GraphQL definitions in `./cmd/frontend/graphqlbackend/schema.graphql`.
1. Start reading through `./enterprise/internal/campaigns/resolvers/resolver.go` to see how the main mutation are implemented (look at `createPatchSetFromPatches` and `createCampaign` to see how the two main operations are implemented).
1. Then start from the other end, `enterprise/cmd/repo-updater/main.go`, and see how the enterprise `repo-updater` uses `campaigns.Syncer` to sync `Changesets`.

## GitHub testing account

Campaigns create changesets (PRs) on code hosts. If you are not part of the Sourcegraph organization, we recommend you create dummy projects to safely test changes on so you do not spam real repositories with your tests. If you _are_ part of the Sourcegraph organization, we have an account set up for this purpose.

To use this account, follow these steps:

1. Find the GitHub `sd9` user in 1Password
2. Copy the `Campaigns Testing Token`
3. Change your `dev-private/enterprise/dev/external-services-config.json` to only contain a GitHub config with the token, like this:

```json
{
  "GITHUB": [
    {
      "authorization": {},
      "url": "https://github.com",
      "token": "<TOKEN>",
      "repositoryQuery": ["affiliated"]
    }
  ]
}
```
