# Adding your organization’s repositories to Sourcegraph Cloud

Search across your code with your team on Sourcegraph Cloud by adding your organization’s repositories to Sourcegraph from GitHub.com or GitLab.com.

## Adding a code host connection

To add your organization’s repositories to Sourcegraph Cloud from GitHub.com or GitLab.com, you must first create a code host connection:

1. Go to User menu > Settings in the top navigation bar.
2. Go to your organization in the sidebar navigation.
3. Go to Code host connections in the sidebar navigation.
4. Connect with GitHub and/or GitLab.

## Security recommendations for access tokens

Both GitHub and GitLab code host connections require an access token. There are two types of tokens you can supply:

- **Machine user token (recommended):** This is a personal access token generated for a “machine user” that is only affiliated with an organization. This gives the code host connection access to only the repositories the machine user is granted access to.
- **Personal access token (not recommended):** This gives the code host connection the same level of access to repositories as the account that created the token, including all public and private repositories associated with the account.

We recommend setting up a machine user to configure your organization’s code host connections.

**Using your own personal access token in your organization’s code host connection may reveal your public and private repositories to other members of your organization.** This is because during early access for organizations on Sourcegraph Cloud, all members of your organization have administration access to the organization settings, including which repositories are synced to Sourcegraph Cloud. The list of repositories available to be synced to Sourcegraph Cloud is determined through the access token associated with the code host connection.

For further instructions and information about personal access tokens and setting up a machine user, please see:

- GitHub: [Machine users in GitHub docs](https://developer.github.com/v3/guides/managing-deploy-keys/#machine-users)
- GitHub: [Personal access tokens](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/creating-a-personal-access-token)
- GitLab: [Personal access tokens](https://docs.gitlab.com/ee/security/token_overview.html#security-considerations)

## Access token scopes

### GitHub.com

Access token requires the following scopes: repo, read:org, and user:email.

### GitLab.com

Access token requires the following scopes: read_user, read_api, and read_repository.

## Adding repositories

To add repositories to sync with Sourcegraph Cloud, follow these steps:

1. Go to User menu > Settings in the top navigation bar.
2. Go to your organization in the sidebar navigation.
3. Go to Repositories in the sidebar navigation.
4. Press the Manage Repositories button.
5. Select public or private repositories to sync with Sourcegraph. This selection can be changed at any time.
6. Press the Save button.

## Searching across your code with your team

After adding your organization’s repositories to Sourcegraph Cloud, those repositories are available for all members of your organization. All members of your organization can use your [organization’s automatic search context](./searching_org_repo_sourcegraph_cloud.md) to search across only your organization’s repositories. 

## Who can see your organization’s code on Sourcegraph Cloud

Please see [who can see your code on Sourcegraph Cloud](./code_visibility_teams_sourcegraph_cloud.md).

## Troubleshooting

### Repositories from GitHub.com are missing or not showing up in search results

If you’ve added your organization’s repositories to Sourcegraph Cloud, and those repositories are missing or not showing up in search results for you or members of your organization, there are two possible causes that must both be resolved.

**GitHub.com organization has not granted access to the Sourcegraph.com OAuth app** If you’ve added a code host connection for GitHub.com and repositories you expect to find are missing or not showing up in search results or in the list of repositories in Organization settings > Repositories, your organization may need to [grant access to the Sourcegraph OAuth app via GitHub.com](https://docs.github.com/en/organizations/restricting-access-to-your-organizations-data/approving-oauth-apps-for-your-organization). You can re-request access to a particular organization using from [your GitHub Settings](https://github.com/settings/connections/applications/e917b2b7fa9040e1edd4). If you do not see your organization listed, you may need to ask your organization admin. 

### Personal code host connection is missing or outdated
If a specific member of your organization finds that private repositories are missing or not showing up in search results, that member may need to add or update their [personal code host connection](../../code_search/how-to/adding_repositories_to_cloud.md).

This is because Sourcegraph respects repository permissions on the code host, and uses [user-centric permissions](../../admin/repo/permissions.md) to determine which repositories a user has access to, from that user’s point of view. To determine which user has access to which repository on the code host, Sourcegraph relies on personal code host connections to uniquely associate a user on Sourcegraph with their account on the code host.

If you continue to have issues, please reach out to our team at [support@sourcegraph.com](mailto:support@sourcegraph.com) or post in the [Sourcegraph Community Slack](http://srcgr.ph/join-community-space).
