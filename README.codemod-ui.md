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
