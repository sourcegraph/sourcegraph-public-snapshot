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
