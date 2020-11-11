# Developing campaigns

## What are campaigns?

Before diving into the technical part of campaigns, make sure to read up on what campaigns are, what they're not and what we want them to be.

1. Start by looking at the [campaigns description on about.sourcegraph.com](https://about.sourcegraph.com).
1. Read through the [campaigns documentation](../../../campaigns/index.md).

## [Campaigns design doc](campaigns_design.md)

See "[Campaigns design doc](campaigns_design.md)".

## Starting up your environment

1. Run `./enterprise/dev/start.sh` and wait until all repositories are cloned.
1. Create [your first campaign](../../../campaigns/quickstart.md). **Remember:** If you create a campaign, you're opening real PRs on GitHub. Make sure only [testing repositories](#github-testing-account) are affected. If you create a large campaign, it takes a while to preview/create but also helps a lot with finding bugs/errors, etc.

## Glossary

The campaigns feature introduces a lot of new names, GraphQL queries and mutations and database tables. This section tries to explain the most common names and provide a mapping between the GraphQL types and their internal counterpart in the Go backend.

<!-- depends-on-source: ~/internal/campaigns/campaign.go, ~/internal/campaigns/campaign_spec.go, etc -->

| GraphQL type        | Go type              | Database table     | Description |
| ------------------- | -------------------- | -------------------| ----------- |
| `Campaign`          | `campaigns.Campaign`       | `campaigns`        | A campaign is a collection of changesets. The central entity. |
| `ChangesetSpec`     | `campaigns.ChangesetSpec`  | `changeset_specs`  | A changeset spec describes the desired state of a changeset. |
| `CampaignSpec`      | `campaigns.CampaignSpec`   | `campaign_specs`   | A campaign spec describes the desired state of a campaign. |
| `ExternalChangeset` | `campaigns.Changeset`      | `changesets`       | Changeset is the unified name for pull requests/merge requests/etc. on code hosts.        |
| `ChangesetEvent`    | `campaigns.ChangesetEvent` | `changeset_events` | A changeset event is an event on a code host, e.g. a comment or a review on a pull request on GitHub. They are created by syncing the changesets from the code host on a regular basis and by accepting webhook events and turning them into changeset events. |

## Database layout

<!-- TODO(mrnugget): Outdated
<!-- <object data="/dev/campaigns_database_layout.svg" type="image/svg+xml" style="width:100%; max-width: 800px"> -->
<!-- </object> -->
<!--  -->
<!-- (To re-generate the diagram from the `campaigns_database_layout.dot` file with Graphviz, run: `dot -Tsvg -o campaigns_database_layout.svg campaigns_database_layout.dot`.) -->

## Diving into the code as a backend developer

1. Read through `./cmd/frontend/graphqlbackend/campaigns.go` to get an overview of the campaigns GraphQL API.
1. Read through `./internal/campaigns/*.go` to see all campaigns-related type definitions.
1. Compare that with the GraphQL definitions in `./cmd/frontend/graphqlbackend/schema.graphql`.
1. Start reading through `./enterprise/internal/campaigns/resolvers/resolver.go` to see how the main mutations are implemented (look at `CreateCampaign` and `ApplyCampaign` to see how the two main operations are implemented).
1. Then start from the other end, `enterprise/cmd/repo-updater/main.go`. `enterpriseInit()` creates two sets of campaign goroutines:
  1. `campaigns.NewSyncRegistry` creates a pool of _syncers_ to pull changes from code hosts.
  2. `campaigns.RunWorkers` creates a set of _reconciler_ workers to push changes to code hosts as campaigns are applied.

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

## How src-cli executes a campaign spec

[src-cli](https://github.com/sourcegraph/src-cli) executes the `steps` in a campaign spec locally before sending up the changeset specs (which include the produced diff) and the campaign spec to the Sourcegraph instance.

Here is a short description of what `src-cli` does when you run with `src campaign [preview|apply]` and the campaign spec includes `steps` (the code for this is neatly contained in [`func runSteps`](https://github.com/sourcegraph/src-cli/blob/850eabe4c8412375de675576d780dd4058004862/internal/campaigns/run_steps.go#L20)):

### 1. Download archive and prepare

1. Download archive of repository (what it does is equivalent to: `curl -L -v -X GET -H 'Accept: application/zip' -H 'Authorization: token <THE_SRC_TOKEN>' 'http://sourcegraph.example.com/github.com/my-org/my-repo@refs/heads/master/-/raw' --output ~/tmp/my-repo.zip
`)
2. Unzip archive, e.g. into `~/Library/Caches/sourcegraph/campaigns` (see `src campaign preview -h` for default value of cache dir, overwrite with `-cache`)
3. `cd` into unzipped archive
4. In the unzipped archive directory, create a git repository:
  - Run `git init`
  - Run `git config --local user.name Sourcegraph`
  - Run `git config --local user.email campaigns@sourcegraph.com`
  - Run `git add --force --all`
  - Run `git commit --quiet --all -m sourcegraph-campaigns`

### 2. Run the steps

For each step:

1. Probe container image to see whether it has `bin/sh` or `bin/bash`
2. Write `steps.run` to a temp file, e.g. `/tmp-script`
3. Run `chmod 644 /tmp-script`
4. Run `docker run --rm --init --workdir /work --mount type=bind,source=/unzipped-archive-locally,target=/work --mount type=bind,source=/tmp-script,target=/tmp-file-in-container --entrypoint /bin/bash -- <IMAGE> /tmp-file-in-container`
5. Add all the changes, run: `git add --all`

### 3. Create final diff

In the unzipped archive:

1. Create a diff by running: `git diff --cached --no-prefix --binary`
