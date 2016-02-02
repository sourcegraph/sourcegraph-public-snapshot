# dev

- Removed the config option `--restrict-write-access`/`RestrictWriteAccess`.
  This flag would limit write access to admin users only. This is no longer
  necessary now that our ACL model includes write access.

# 0.13.1

- Promote Notification Center feature to default on (it previously had to be toggled on via a flag).
  It allows receiving notifications in a centralized manner in a web UI. New notifications
  appear as a red badge over the icon in top right. Supported apps (Tracker) will subscribe you to
  updates for threads you create, where you participate, or get @mentioned.

# 0.13.0

- Fix issues serving /robots.txt and /favicon.ico.

# 0.12.1

- Fix issue blocking changesets from being created from the UI.

# 0.12.0

- Reduce errors of Drone connecting to SSH server that
is not (yet) listening.

# 0.11.10

- Added an SSH client to the sourcegraph docker container.
- Only try to convert localhost to an absolute IP in
  Drone build configuration if the given clone
  URL is localhost.

# 0.11.9

- Added a number of commits to improve the resilience of the build
  process.

# 0.11.8

- Added a new event logger in `util/eventsutil`, which is similar to
  `util/metricsutil` but enables more precise logging of user actions
  on the Sourcegraph instance.
- Enforce minimum `ulimit` value of 10000 to prevent errors caused by insufficient file
  descriptors.

# 0.11.7

- Improve the process of running a Sourcegraph server in Docker. The
  docker image is simplified, takes less disk space, and no longer uses
  an external config file.
- A regression that caused expanding the top and bottom of diffs (for more
  context) in Changesets has been fixed.
- Users may now commit a `.sourcegraph-merge-template` file to their repository
  root which will be used to generate the merge commit message when merging a
  changeset via the web UI or src CLI. See [the docs](https://src.sourcegraph.com/sourcegraph/.docs/features/custom_merge_template)
  for more info.
- Merging a changeset via the web UI or CLI with the squash option on a
  Sourcegraph-hosted or GitHub mirror repository will now correctly mark the
  changeset as merged and automatically close it.
- Clicking the "Merge changeset" button will now correctly show users the
  updated changeset instead of a blank page.
- Merging a changeset on a Sourcegraph-hosted repository via the web UI or src
  CLI will now properly mark the changeset as merged and automatically close it.
- Merging a changeset on a Sourcegraph-hosted repository via the web UI or src
  CLI will now cause the changeset to properly persist and not lose its data.

# 0.11.6

- Fixes panic on empty commit author.

# 0.11.5

- Fix cloud install automation by providing appropriate default config.ini
to Linux .deb and .rpm packages.

# 0.11.4

- Improve git event parser to be significantly faster and correct for large git pushes.
- Fixed a regression that prevented users from being able to add new SSH keys
  via the CLI or new web UI altogether.

# 0.11.3

- CI is now integrated into Sourcegraph with drone.io. Docker is now a
  dependency for a working environment. Run `src info` to check your system
  requirements.
- Several flags related to workers have been removed:
  `-n`/`--num-workers`/`NumWorkers`, `--build-root`/`BuildRoot`,
  `--clean`/`Clean`. They are no longer relevant due to relying on docker. To
  control build concurrency specify `--parallel`.
- Builds have a new build: `BuilderConfig`. **PostgreSQL backend:**
  Run `alter table repo_build add column builder_config text default
  '';` to perform this migration.
- Build tasks have another 2 additional fields: `Skipped` and
  `Warnings`.  **PostgreSQL backend:** Run `alter table
  repo_build_task add column skipped bool default false; alter table
  repo_build_task add column warnings bool default false;` to perform
  this migration.
- Build tasks now have a new field, `ParentID`. **PostgreSQL
  backend:** Run `alter table repo_build_task add column parent_id
  bigint default 0;` to perform this migration.
- A new UI to add and remove multiple SSH public keys is provided under the user
  settings page. Users will need to re-add their SSH public key (via UI or CLI)
  once more after upgrading due to a change in the storage backend, this is a
  one-time process.

# 0.11.0

- Alongside Tracker and Changes applications now sits a new "API Docs" tab which
  automatically generates an API documentation overview for a given directory of
  code.

# 0.10.0

- Builds are now identified by a numeric build ID, instead of by
  commit ID. After this update, build metadata from pre-update builds
  will not be available, but build results (source code analysis) will
  remain. **PostgreSQL backend:** Run `drop table repo_build; drop
  table repo_build_task;` and then run the CLI command `src pgsql
  create` to perform this migration.
- Two unnecessary fields on build objects have been
  removed. **PostgreSQL backend:** Run `alter table repo_build drop
  column "import"; alter table repo_build drop column usecache;` to
  perform this migration.
- The `src push` command no longer guesses the current repository. You
  must specify it with `src push --repo my/repo`.


# 0.9.0

- Display the latest commit that touched each file and directory on
  the repository tree listing. This may cause degraded performance on
  extremely large Git repositories; use the
  `--app.disable-tree-entry-commits`/`DisableTreeEntryCommits` config
  to disable this feature.
  - Perform an inventory of repositories to determine what languages,
  etc., are in use, by walking their directory tree. This occurs after
  each push. This operation may be slow for extremely large
  repositories; use the
  `--local.disable-repo-inventory`/`DisableRepoInventory` configs to
  disable this functionality.
- Allow enabling/disabling apps on a per-repository basis (with `src
  repo config app`). **PostgreSQL backend:** Run `alter table
  repo_config add column apps text[];` to perform this migration.
- Remove the concept of enabling/disabling a repository, which was no
  longer used. **PostgreSQL backend:** Run `alter table repo_config
  drop column enabled; alter table repo_config drop column admin_uid;`
  to perform this migration.
- Add a new revision syntax `REV^{srclib}`, which refers to the the
  nearest ancestor to REV that has srclib Code Intelligence data.
- Remove the `--app.show-latest-built-commit`/`ShowLatestBuiltCommit`
  configs. This functionality now occurs automatically without the
  negative tradeoffs associated with the previous scheme. An old
  file's contents are annotated with Code Intelligence if it has not
  changed since the last build. See docs/config/repos.md for
  information on obtaining the old ShowLatestBuiltCommit behavior by
  using `REV^{srclib}` as a repo's default branch.
- Eliminate the display of definition names on the repository
  directory view (beside directory entries). This feature was not very
  useful and removing it improves performance. This means the
  `--app.disable-dir-defs` and `DisableDirDefs` configs no longer
  exist.

# 0.8.15

- Change user authentication to be managed locally (not via OAuth2 to Sourcegraph.com)
	- Store user data and access controls on Sourcegraph server filesystem or database
	- Add support for generating invitation links for teammates
	- Change access control CLI commands (see [updated docs](https://src.sourcegraph.com/sourcegraph/.docs/management/access-control/))
	- [Read more about Sourcegraph authentication](https://src.sourcegraph.com/sourcegraph/.docs/config/authentication/)
- Add support for creating and merging changesets for mirrored GitHub repositories
- Improve UX for external host repository mirrors
	- Enable and clone repos automatically
	- Improve performance and fix bugs for refreshing stale data
- Fix JIRA integration bug for mirrored repositories (see [updated docs](https://src.sourcegraph.com/sourcegraph/.docs/integrations/JIRA/))
- Improve performance of loading changesets
- Add various UI enhancements
- Fix various UI bugs
