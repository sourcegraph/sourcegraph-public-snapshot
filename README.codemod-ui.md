# Trying out `codemod-ui`

1. Check out the `plain-prop-names` branch of https://github.com/sourcegraph/extension-api-classes, run `yarn link`, and run `yarn run build` (all subsequent steps are in the `sourcegraph` repository)
1. Check out the `codemod-ui` branch in the `sourcegraph` repository
1. Run `yarn link @sourcegraph/extension-api-classes`
1. Rerun `enterprise/dev/start.sh` (there are DB migrations and new `yarn` dependencies)
1. Add the following `github.com/` repositories for best results:

   ```
   ["lyft/amundsenfrontendlibrary", "lyft/pipelines", "sourcegraph/codeintellify", "sourcegraph/react-loading-spinner", "sourcegraph/about"]
   ```
1. Run the extension that provides the checks shown in the demo video (https://sourcegraph.slack.com/archives/CHEKCRWKV/p1559132679011300): `cd extensions/enterprise/check-search && yarn && yarn run serve` and keep this running
1. In your browser, [sideload the `check-search` extension](https://docs.sourcegraph.com/extensions/authoring/local_development) (it is on `http://localhost:1234`)
1. In the UI, create an organization, then visit the organization's **Projects** tab and create a project (eg `myproject`).
1. Visit the newly created project (eg `http://localhost:3080/p/1`).
1. Click on **Checks** on the left-hand sidebar
1. Click **New check**
1. Create any check. They're actually all the same (they all are just hard-coded to use the `check-search` extension).
1. Visit the check's **Items** tab, etc., as shown in the demo video above.

## Downgrading back to `master`

There are DB migrations, so if you downgrade to `master`, it may complain. You can use a separate PostgreSQL DB to avoid needing to wipe your DB each time you switch back and forth.

1. Run `sudo -u postgres createdb -O ${PGUSER-$USER} sg-codemod-ui` to create a new database owned by your PostgreSQL user.
1. When using `codemod-ui`, use `PGDATABASE=sg-codemod-ui enterprise/dev/start.sh` to start the server in the separate DB.

You *might* see some weirdness because Redis isn't similarly isolated. I don't know the solution because I haven't encountered any problems, but there's probably an easy way to use a separate Redis namespace, too.


---

# Concepts

## Changeset

A **changeset** consists of:

- a set of changes to code across one or more repositories
- optionally, the **plan** for how those changes were made, so the changes can be recomputed against an updated base branch

You can create a changeset in 3 ways:

- clicking "Create changeset" on a notification about a suggested change
- performing a search-and-replace
- clicking a code action in a code file (e.g., "Rename identifier")

## Plan

A changeset's **plan** is the sequence of operations to compute the changeset's change. Unlike a patch (which only has the line-level changes), a changeset with a plan also includes the programmatic steps to compute those changes.

> Example: You want to upgrade all repositories' `lodash` dependencies to version `^3.0.0`. If another change to `package.json` or `yarn.lock` is made before the changeset is merged, it knows how to recompute its updates to those files. If a new repository is added that depends on an older version of `lodash`, it knows to add the new repository to the changeset.

The plan is stored in the changeset. Each operation in the sequence consists of:

- parameters, specified as a JSON object whose schema is defined by the plan
- a command to invoke (with the parameters and diagnostics passed as arguments)

## Status

A status consists of:

- a group of related diagnostics
- aggregated user-facing container for related diagnostics and actions, plus the configuration used to generate and compute these diagnostics and actions.

goal is to allow changesets to granularly select high-level actions using diagnostics+codeactions, and also to not require status providers to reimplement a lot of custom stuff for notifications.

- idea: make notificationprovider be a function of the diagnostics that exist? assumption is that the diagnostics are the slow part to compute and can be precomputed on the backend, and the notificationprovider/status can then be frontend-only.

TODO: how to cover new usages of your code, active people on this repo, diagnostics, etc.? rename diagnostics to annotations/notes?

- other idea: statuses are just diagnostics (multiple categories thereof, like "wrong go version in CI" and "missing .travis.yml") and actions. there can be "fix all" actions for an entire category of diagnostic that creates a changeset using the default code action for all instances of that diagnostic category. the user can then go diagnostic-by-diagnostic and customize the code action if desired (including ignoring a diagnostic, ie the null code action). any new instances of that diagnostic found are added to the changeset (if it's an auto-changeset) with the default code action. **TODO!(sqs): this seems the most promising and simple**


## Notification

A notification consists of a message, actions, and contextual information about the current state.
