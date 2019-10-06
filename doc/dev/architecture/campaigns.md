# Campaign architecture

## Action

An **action** is a reusable function that accepts code-related inputs and produce outputs consisting of diagnostics, diffs, or plans describing external state mutations. It is implemented by a JavaScript callback or Docker image.

Examples:

- `findPackageJsonDependencies`: Find package.json packages that depend on `$packageName` at a version satisfying `$packageVersion`.
- `upgradePackageJsonDependencies`: Generate diffs for upgrading all dependencies on `$packageName@$packageVersion` to `$newVersion`.

## Run

A **run** of an action runs the action with a specific set of arguments.

## Workflow

A **workflow** describes the triggers and arguments for invoking one or more actions.

Examples:

- For upgrading a dependency:
  - On all code, run `findPackageJsonDependencies` and show diagnostics for instances of the deprecated dependency.
  - On the default branch, run `upgradePackageJsonDependencies` and create/update branches and pull requests with the diffs.

## Campaign

A **campaign** executes a workflow over a specific set of repositories on Sourcegraph and tracks all associated invocations, branches, pull requests, and other external state mutations.
