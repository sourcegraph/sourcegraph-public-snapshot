# Modeling teams in Sourcegraph

To model your internal team structure in Sourcegraph, you can utilize Sourcegraph teams. Teams are groupings of users into a common handle. Teams are structured as a tree, so teams can have child teams.

Example:

```
engineering
├─ security
├─ code graph
│  ├─ Batch Changes
│  ├─ Code Insights
├─ source
│  ├─ repo-management
│  ├─ IAM
product
```

Teams in Sourcegraph will be usable in Sourcegraph Own [other features in the future]. Teams can be code owners and will influence the Own experience. You can search for code owned by a specific team, and in the future advanced ownership analytics will be informed by given team structures. [TODO: Link to Own docs](./teams.md)

## Configuring teams

Teams can either be defined directly in Sourcegraph by hand, or be ingested from external systems into Sourcegraph using [src-cli](https://github.com/sourcegraph/src-cli).

### From the UI

Go to **site-admin>Teams**. On this page, click "Create a new team". The team has to at least be a unique name and can optionally take a display name. Additionally, you can define a teams parent team to build a tree structure as outlined above.

After hitting create, you will be redirected to the team page where you can add Sourcegraph users as team members.

> NOTE: Teams defined from src-cli using the `-readonly` flag cannot be modified from the UI to prevent state drift from external systems ingesting the data.

### From the CLI

If you prefer a command line based approach, or would like to integrate an external system of record for teams into Sourcegraph, [src-cli](https://github.com/sourcegraph/src-cli) provides commands to manage teams as well:

```
src teams create <name> [-displayName=<displayName>] [-readonly]
src teams delete <name|ID>
src teams list [-search=<search>]
src teams add-member <name|ID> <username|userID> [<username|userID>...]
# Forcefully overwrites all members of a given team.
src teams set-members <name|ID> <username|userID> [<username|userID>...]
src teams remove-member <name|ID> <username|userID> [<username|userID>...]
```

## Common integrations

### GitHub teams

Using the GitHub CLI, you can ingest teams data from GitHub into Sourcegraph. You may want to run this process regularly.

```
TODO: Script here that scrapes the GitHub API for teams and converts them into Sourcegraph teams.
```

### GitLab teams

Using the GitLab API, you can ingest teams data from GitLab into Sourcegraph. You may want to run this process regularly.

```
TODO: Script here that scrapes the GitLab API for teams and converts them into Sourcegraph teams.
```
