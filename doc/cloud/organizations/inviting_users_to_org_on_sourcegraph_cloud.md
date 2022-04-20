# Invite others to join your organization on Sourcegraph Cloud

Invite others to join your organization on Sourcegraph Cloud and search across [your organization’s repositories synced to Sourcegraph Cloud](./adding_your_org_repos_to_cloud.md):

1. Go to User menu > Settings in the top navigation bar.
2. Go to your organization in the sidebar navigation.
3. Go to Members in the sidebar navigation.
4. Enter a team member’s Sourcegraph Cloud username in invite a member.
5. Press Send invitation to join.

Your team member will receive an email inviting them to join your organization on Sourcegraph Cloud. When they have accepted the invitation, they will appear in the list of organization members.

Note that during early access for organizations on Sourcegraph Cloud, only existing Sourcegraph Cloud users can be invited to an organization. They must [create a Sourcegraph Cloud account](https://sourcegraph.com/sign-up) first. 

## Troubleshooting

### Repositories from GitHub.com are missing or not showing up in search results

If your [organization’s repositories synced to Sourcegraph Cloud](./adding_your_org_repos_to_cloud.md) are missing or not showing up in search results for a member of your team, that team member may need to [add or update their personal code host connection](../../code_search/how-to/adding_repositories_to_cloud.md).

This is because Sourcegraph respects repository permissions on the code host, and [user-centric permissions](../../admin/repo/permissions.md) to determine which repositories a user has access to, from that user’s point of view. To determine which user has access to which repository on the code host, Sourcegraph relies on personal code host connections to uniquely associate a user on Sourcegraph with their account on the code host.

If you continue to have issues, please reach out to our team at [support@sourcegraph.com](mailto:support@sourcegraph.com) or post in the [Sourcegraph Community Slack](http://srcgr.ph/join-community-space).
