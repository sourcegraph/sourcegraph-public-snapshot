# Developing Automation

## What is Automation?

Before diving into the technical part of Automation, make sure to read up on what Automation is, what it's not and what we want it to be:

1. Start by reading through the [Automation product landingpage](https://about.sourcegraph.com/product/automation/)
1. **IMPORTANT:** Watch the Automation videos! At the bottom of the landingpage, you'll find two demo videos that showcase the original prototype for Automation. A lot of our work aims to reproduce what you can see in these videos — in a scalable way that supports multiple code hosts. **Make sure to watch these videos!**
1. Take a look at the [Automation sequence of milestones](https://docs.google.com/document/d/1TDsjrCy55UTZA_NyofVssnBotTyPP6Hvbrsp__aUCfM/edit#heading=h.go9qqwdnhiyu) to get a high-level overview of what we did so far and what still needs to be done
1. Read through the [user documentation on Automation](../user/automation.md).

## Starting up your environment

1. Run `./enterprise/dev/start.sh` — Wait until all repositories are cloned.
2. Follow the [user guide on creating campaigns](../user/automation.md). **Remember:** If you create a campaign, you're opening real PRs on GitHub. Make sure only [testing repositories](#github-testing-account) are affected. If you create a large campaign, it takes a while to preview/create but also helps a lot with finding bugs/errors, etc.

## Glossary

Automation introduces a lot of new names, GraphQL queries and mutations and database tables. This section tries to explain the most common names and provide a mapping between the GraphQL types and their internal counterpart in the Go backend.

| GraphQL type        | Go type              | Database table     | Description |
| ------------------- | -------------------- | -------------------| ----------- |
| `Campaign`          | `a8n.Campaign`       | `campaigns`        | A campaign is a collection of changesets on code hosts. The central entity in Automation. |
| `ExternalChangeset` | `a8n.Changeset`      | `changesets`       | Changeset is the unified name for pull requests/merge requests/etc. on code hosts.        |
| `CampaignPlan`      | `a8n.CampaignPlan`   | `campaign_plans`   | A campaign plan is a collection of changes (think: patches/diffs) that will be applied by running a Campaign. A campaign *has one* campaign plan. |
| `ChangesetPlan`     | `a8n.CampaignJob`    | `campaign_jobs`    | A *plan* for a changeset. It represents a patch per repository that *can* be a changeset. It belongs to a campaign plan, which has multiple changeset plans, one per repository. |
| -                   | `a8n.ChangesetJob`   | `changeset_jobs`   | It represents the process of turning a `ChangesetPlan` (GraphQL)/`a8n.CampaignJob` (Go) into a `Changeset` on the code host. It is executed asynchronously in the background when a campaign is created with a campaign plan. |
| `ChangesetEvent`    | `a8n.ChangesetEvent` | `changeset_events` | A changeset event is an event on a code host, e.g. a comment or a review on a pull request on GitHub. They are created by syncing the changesets from the code host on a regular basis and by accepting webhook events and turning them into changeset events. |

## Diving into the code as a backend developer

1. Read through `./cmd/frontend/graphqlbackend/a8n.go` to get an overview of the Automation GraphQL API.
1. Read through `./internal/a8n/types.go` to see all Automation related type definitions.
1. Compare that with the GraphQL definitions in `./cmd/frontend/graphqlbackend/schema.graphql`.
1. Start reading through `./enterprise/internal/a8n/resolvers/resolver.go` to see how the main mutation are implemented (look at `createCampaignPlanFromPatches` and `createCampaign` to see how the two main operations are implemented).
1. Then start from the other end, `enterprise/cmd/repo-updater/main.go`, and see how the enterprise `repo-updater` uses `a8n.Syncer` to sync `Changesets`.

## GitHub testing account

Automation features require creating changesets (PRs) on code hosts. If you are not part of the Sourcegraph organization, we recommend you create dummy projects to safely test changes on so you do not spam real repositories with your tests. If you _are_ part of the Sourcegraph organization, we have an account set up for this purpose.

To use this account, follow these steps:

1. Find the GitHub `sd9` user in 1Password
2. Copy the Automation Testing Token
3. Change your `dev-private/enterprise/dev/external-services-config.json` to only contain a GitHub external service config with the token, like this:

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
