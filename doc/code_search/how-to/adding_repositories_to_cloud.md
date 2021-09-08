# Adding repositories to Sourcegraph cloud

Sourcegraph cloud lets you add repositories to search with Sourcegraph without needing to self-host. You can add repositories you own or collaborate on from GitHub.com or GitLab.com.

## Adding a code host connection

> NOTE: If you're adding organization repositories from GitHub.com, the Sourcegraph.com OAuth application needs to [get approval from the organization owner for granting access](https://docs.github.com/en/organizations/restricting-access-to-your-organizations-data/approving-oauth-apps-for-your-organization).

To add repositories you own or collaborate on from GitHub or GitLab, you must first create a code host connection:

1. Go to **User menu > Settings** in the top navigation bar.
1. Go to **Code host connections** in the sidebar navigation.
1. Connect with GitHub and/or Gitlab using OAuth.

## Adding repositories by selecting Sync All (Recommended)

If you want to sync all the repositories you have access to and have Sourcegraph automatically sync all new repositories on the code host in the future, follow the below steps: 

1. Go to **User menu > Settings** in the top navigation bar.
1. Go to **Repositories** in the sidebar navigation.
1. Press the **Manage Repositories** button.
1. Choose the 'Sync All' option.
1. Press the **Save** button.


## Adding specific repositories  

If you only want to add specific repositories to Sourcegraph cloud, follow the below steps:

1. Go to **User menu > Settings** in the top navigation bar.
1. Go to **Repositories** in the sidebar navigation.
1. Press the **Manage Repositories** button.
1. Choose the 'Sync select repositories' option. This will display a list of all repositories, public or private, which you have access to. Sourcegraph will only sync the repositories you select. This can be changed at a later time. 
1. Press the **Save** button.

## Adding other public repositories from GitHub.com or GitLab.com

To add public or private repositories from GitHub.com or Gitlab.com without creating a code host connection:

1. Go to **User menu > Settings** in the top navigation bar.
1. Go to **Repositories** in the sidebar navigation.
1. Press the **Manage Repositories** button.
1. Check **Sync specific public repositories by URL** under **Other public repositories**
1. Specify public repositories on GitHub and GitLab using complete URLs to the repositories. One repository per line.
1. Press the **Save** button.

## Who can see your code on Sourcegraph cloud.

Please see [who can see your code on Sourcegraph cloud](../explanations/code_visibility_on_sourcegraph_cloud.md).

## Troubleshooting

### Repositories from code hosts are missing or not showing up while adding repositories

If you've connected with a code host and repositories you expect to find are not shown while adding repositories in **User menu > Settings > Repositories > Manage repositories**, you may not have permission on the remote code host to add those repositories to Sourcegraph.

If the missing repositories belong to an organization on GitHub, [review the Sourcegraph OAuth access](https://github.com/settings/connections/applications/e917b2b7fa9040e1edd4). If necessary, grant or request access for the given organization. If you are not an owner, you will receive an email notification when an owner grants access. Then, you can [follow the steps to add repositories](#adding-repositories-you-own-or-collaborate-on-from-github-or-gitlab).

If you continue to have issues, [reach out to our team](mailto:support@sourcegraph.com).
