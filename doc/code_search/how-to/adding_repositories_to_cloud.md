# Adding repositories to Sourcegraph Cloud

Sourcegraph Cloud lets you add repositories to search with Sourcegraph. You can add repositories you own or collaborate on from GitHub or GitLab, or add other public repositories from GitHub or Gitlab.

> NOTE: Sourcegraph Cloud currently only supports adding public repositories. Soon, you will be able to search private repositories with Sourcegraph Cloud. [Get updated when this feature launches](https://share.hsforms.com/1copeCYh-R8uVYGCpq3s4nw1n7ku).

## Adding repositories you own or collaborate on from GitHub or GitLab

To add repositories you own or collaborate on from GitHub or GitLab, you must first create a code host connection:

1. Go to **User menu > Settings** in the top navigation bar.
1. Go to **Code host connections** in the sidebar navigation.
1. Connect with GitHub and/or Gitlab using OAuth, or choose "Connect with access token" to enter a manually-created access token for the given code host.

Once you've connected with a code host, you can add repositories:

1. Go to **User menu > Settings** in the top navigation bar.
1. Go to **Repositories** in the sidebar navigation.
1. Press the **Manage Repositories** button.
1. Choose the public repositories you own or collaborate on from your connected code hosts that you want to search with Sourcegraph.
1. Press the **Save** button.

## Adding other public repositories from GitHub or GitLab

To add public repositories from GitHub or Gitlab without creating a code host connection:

1. Go to **User menu > Settings** in the top navigation bar.
1. Go to **Repositories** in the sidebar navigation.
1. Press the **Manage Repositories** button.
1. Check **Sync specific public repositories by URL** under **Other public repositories**
1. Specify public repositories on GitHub and GitLab using complete URLs to the repositories. One repository per line.
1. Press the **Save** button.

## Who can see your code on Sourcegraph Cloud.

Please see [who can see your code on Sourcegraph Cloud](../explanations/code_visibility_on_sourcegraph_cloud.md).

## Troubleshooting

### Repositories from code hosts are missing or not showing up while adding repositories

If you've connected with a code host and repositories you expect to find are not shown while adding repositories in **User menu > Settings > Repositories > Manage repositories**, you may not have permission on the remote code host to add those repositories to Sourcegraph.

If the missing repositories belong to an organization on GitHub, [review the Sourcegraph OAuth access](https://github.com/settings/connections/applications/e917b2b7fa9040e1edd4). If necessary, grant or request access for the given organization. If you are not an owner, you will receive an email notification when an owner grants access. Then, you can [follow the steps to add repositories](#adding-repositories-you-own-or-collaborate-on-from-github-or-gitlab).
