# Trying out `codemod-ui`

1. Check out the `plain-prop-names` branch of https://github.com/sourcegraph/extension-api-classes, run `yarn link`, and run `yarn run build` (all subsequent steps are in the `sourcegraph` repository)
1. Check out the `a8n/wip` branch in the `sourcegraph` repository
1. Run `yarn link @sourcegraph/extension-api-classes`
1. Rerun `enterprise/dev/start.sh` (there are DB migrations, new `yarn` dependencies, and a new Procfile entry)
1. Add the following `github.com/` repositories for best results:

   ```
   ["lyft/amundsenfrontendlibrary", "lyft/pipelines", "sourcegraph/codeintellify", "sourcegraph/react-loading-spinner", "sourcegraph/about"]
   ```
1. [Sideload](https://docs.sourcegraph.com/extensions/authoring/local_development) the extension at http://localhost:1235. (This is the extension in `./extensions/enterprise/sandbox`, and the main Procfile now starts a Parcel server for it.)
1. In the UI, go to **Campaigns** in the global navbar.

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

- An auto-changeset means applying the same code action to all current and future instances of a diagnostic. The set of diagnostics is specified as a query mentioning the tags and other criteria (repo/file/etc.) of the diagnostic.

- Diagnostic type: let extensions register diagnostic types, which specify a common set of actions. Should this be registerDiagnosticType (no because that would mean the eslint extension would need to call that too many times, once per eslint rule id); registerDiagnosticActionTemplate (complex, kind of duplicates provideCodeActions); provideBatchCodeActions(takes multiple docs/locations/diagnostics plus a query) which returns a plan? Also it would let you set up notifications. TODO!(sqs): define and simplify the new concepts being introduced.

** The idea is that the heavy lifting is done on the backend beforehand, and the extensions provide the crucial UX flows to make use of that data.


## Notification

A notification consists of a message, actions, and contextual information about the current state. It is derived from a diagnostics query.


--------

A status contains one or more checks, acts as a container instance for configuring them, plus has the ability to roll-up their results into a single status.

A check defines a class of diagnostics and common actions that can be taken on them:

- ESLint: is configured, runs in CI, all errors are fixed, all recommendations are applied
- Up-to-date npm dependencies: lockfiles (eg yarn.lock) exist, latest dependency version is in use
- TypeScript build config: uses Yarn, has standard set of scripts, has standard set of prettierignore/prettierrc/eslintrc/tsconfig/etc. files
- API usage review: notify me of any new users of a specific package
- Deployment config: TODO
- Upgrade a specific library and all call sites: TODO

DiagnosticQuery - some have actions associated with them (eg up-to-date npm deps), some actions are for the entire set of checks (eg TS build config, wouldn't want to do any single step independently...although that probably only applies to auto-changesets, when this is still manual you may want to run some of the steps manually to (eg) create a PR to standardize prettierrc)

** Calling them Status is weird because it adds a new layer/concept above Checks, but it's nice because it can also be used to communicate the status of other things like language analysis, what people are doing, etc. (although that can go in "activity")


.

---

The extension API for diagnostics is "push" instead of "pull", which is the opposite of the extension APIs for hovers, references, etc. Extensions monitor the workspace state on their own and update diagnostics on their own, instead of registering a diagnostic provider that is invoked per file. The diagnostics API is "push" because the client has no way of knowing which actions might trigger an update of diagnostics (e.g., a change to one file might cause errors in hundreds of other files), so it needs to rely on the extension to listen for its own triggers. TODO!(sqs): Maybe the simple reason is that in VS Code, there is no streaming (ie only Promises, not Observables), so they needed to do it this way (although VS Code's current way has a benefit of being able to send partial updates and not re-sending the entire diagnostics set each time)?

Why have both annotations/diagnostics *and* decorations? They serve different purposes. Annotations/diagnostics are for things that are permanent/long-lived, derived from the code itself and not per-user state, are viewed in aggregate/summary or in a list, and are semantically meaningful to other consumers. Decorations are purely visual and for interactive consumption by a user.

Make diagnostics into a provider

DiagnosticQuery


Pipelines, rules, statuses, workflows, policies? They are a diagnostic query (or null) -> action? In general, they are event -> action.

For ESLint:

- TS/JS code that is not checked by ESLint -> set of diagnostics to show where, then null action (or fail build, notify $USER, submit PR to add ESLint)
- New ESLint rules -> action to notify someone

For updating a dep:

- New version published -> open or modify existing changeset

Changeset:

- Base branch changed -> rerun actions against base branch

Default rules in Check for ESLint:

- Any TS/JS code that ESLint is not running on -> notify all, fail check
- Any rule failures -> notify all, propose fix

A notification is versioned but has a persistent identifier (eg "this is the persistent notification about the rule that TS/JS code without ESLint should trigger a notification") and can pertain to any number of repositories.

Default rules for package.json standardization:

- Publisher is sourcegraph (diagnostic + code action)
- License is MIT (diagnostic + code action)

A check shows the diagnostic groups (with related rules, if any) plus the 1-time and automated actions that can be taken on them.

---

In the check area, as a user, I want to:

- See the overall check status
- Understand and configure the check settings
- See what automation is currently active for this check
- See individual problems
- Fix an individual problem
- Fix a batch of problems (and preview what will happen)

---

Show all diagnostics individually, but make it easy to batch them when fixing:

1. Set diagnostic query
1. Default action = ignore, can be changed (batch > apply to all 37)
1. Choose actions for each diagnostic
1. 

- "Fix all"

TODO - show a bar above "Filter v" or at the bottom that shows (1) the number of actions chosen for diagnostics, (2) the total diffstat and number of repos affected, (3) "Batch" > "Apply $TITLE to all (N)", (4) "New changeset"

---

An operation is an action applied to a set of diagnostics, or just an action. 


--------


- Changeset = (repository, branch, PR link, operation[], rule[])
- Changeset campaign = (name, changeset[], rule[]) - cross-repository

What about something that finds a lot of diagnostics across many repositories and requires action on them, but where that action isn't opening a diff? Is a changeset just a special case of an issue?

Do we need a campaign, or can it just be a label?

- Issue = (repository, issue link, diagnostic[], comment[]) -- upon 1st diagnostic action that performs edit, upgrade the issue into a changeset
- Changeset = Issue & (branch, PR link, comment[], operation[])
- Campaign = (name, issue[], changeset[], rule[])

* An issue seems heavyweight for assigning diagnostics.

* A campaign's rule can say how to group the actions by changeset/issue (eg by code owner)

For fixing invalid codeowners:

- First make an issue on all repositories listing the problems and linking to Sourcegraph to select the fix (TODO or if you wanted to bypass that step and just pick the autofix, how would that look?)
- Then when the fix was chosen on each issue, create a changeset whose commit closes the issue

== Rules have conditions (which is the same as search syntax) + actions --- this reuses search stuff and makes it easy to go from a search to a rule

TODO think about how this would work for "i want all instances of ___ to be reviewed and approved" eg call sites

Campaign rules:

- Campaign = (name, changeset[],

>

Codemod workflow:

- Find a bunch of candidates to fix
- Make changesets, grouped by (repository, code owner)

Review API consumers workflow:

- Find all instances of ee
- Create issues, grouped by (repository, code owner)


-------

Diagnostic providers are responsible for querying the diagnostics GraphQL API (by whatever criteria they use for canonicalizing/deduping diagnostics) to see if a diagnostic is currently contained in any threads.

Diagnostic providers can also 'resolve' a diagnostic given the data stored in the database. This is faster than running the provider over the entire workspace again.


-------

Key insight is that there is no meaningful distinction between issues and changesets in a world with automation? When you aren't reviewing the diff line-by-line, you might "approve" and "merge" changes that you only see a high-level description of (the intent of), not the actual changes. In that world, things that look like issues (with no code diff) have "approve" and "merge" buttons.


-------

# Demos

## Monitor existing campaign

## Deprecate lodash

I'm Quinn, the CEO of Sourcegraph. We're building the standard developer platform for engineering teams. We started with code search and navigation, and now many companies you know have every engineer using Sourcegraph for code search and navigation (see our site for customer logos).

With tons of developers using our product, we get to learn how so many different teams build software and the problems they face. One problem kept coming up again and again: it's super painful to make large-scale code changes that touch hundreds or thousands of internal projects and code owners.

We're excited to preview our solution to this problem. We're calling it automation, and when it's released, the Sourcegraph standard developer platform will be search plus automation.

Let me set the stage for the demo now. Large engineering teams need to make large-scale code changes to pay down tech debt, remove legacy code, keep dependencies up-to-date, and address critical security issues. This is painful in 3 ways:

- It's hard to write and test the code that performs the edits.
- It's hard to create and update branches and pull requests across thousands of teammates and repositories (or trees in a monorepo).
- It's hard to track the progress of the campaign and help code owners merge changes ASAP.

Because of this pain, way too few large-scale code changes are made. Automatically fixable tech debt like lint issues persists, legacy code remains and nobody knows if it's called, dependencies diverge and get stale, and you're never really sure if everyone has upgraded (and stayed upgraded) to dependencies with critical security fixes.

We want to give you the power to automate and manage large-scale code changes so you can focus on the coding tasks that truly engage your brain, without all that background noise. Let's jump in and see how you can deprecate a dependency across your entire company with Sourcegraph.

Sourcegraph wants to make this much easier for you, so you can 

1. Go to Campaigns
1. Select the organization that will own the campaign
1. Select the type of campaign

Campaign rules

## Eliminate code duplication


----

make CreateCampaignInput take all the stuff, including the rules and threads, and also make a one-shot preview endpoint that lets us show live stats on the new-campaign page

----

demo todos:

- add <PageTitle /> components everywhere
