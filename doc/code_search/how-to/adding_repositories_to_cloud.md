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

## Searching across repositories you've added to Sourcegraph with search contexts

Once you've added repositories to Sourcegraph, you can search across those repositories by default using search contexts.

Sourcegraph Cloud supports two search contexts: 

- Your personal context, `context:@username`, which automatically includes all repositories you add to Sourcegraph.
- The global context, `context:global`, which includes all repositories on Sourcegraph Cloud.

Coming soon: create your own search contexts that include the repositories you choose. Want early access to custom search contexts? [Let us know](mailto:feedback@sourcegraph.com).
