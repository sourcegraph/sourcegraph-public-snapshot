# dev

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
