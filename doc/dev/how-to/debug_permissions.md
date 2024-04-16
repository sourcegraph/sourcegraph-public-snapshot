# Debugging Repository Permissions

This document provides a list of options to try when debugging permissions issues.

## Check the user permissions screen

1. As a site administrator, visit the **Site admin** > **User administration** page and search for the user in question.
![](https://storage.googleapis.com/sourcegraph-assets/Docs/how-to/debug-permissions/user_search.png)
2. Click on their username and navigate to their **Settings** > **Permissions** page.
   This page gives an overview of a user's permission sync jobs, as well as the repositories they have access to and why.
![](https://storage.googleapis.com/sourcegraph-assets/Docs/how-to/debug-permissions/permissions_page.png)
3. At the bottom of the **Permissions** page, you'll find a list of repositories the user can access and the reason for their access. Possible access reasons are:
    - **Public** - The repository is marked as public on the code host, so all users on the Sourcegraph instance can access this repository.
    - **Unrestricted** - The code host connection that this repository belongs to is not enforcing authorization, so all users on the Sourcegraph instance can access this repository.
    - **Permissions Sync** - The user has access to this repository because they have connected an external account with access to it on the code host.
    - **Explicit API** - The user has been granted explicit access to this repository using the explicit permissions API.
![](https://storage.googleapis.com/sourcegraph-assets/Docs/how-to/debug-permissions/accessible_repos.png)
4. If the user should have access to a repository, but the repository is not at all present in this list, check if the user has connected their code host account to their user profile.

## Check the user's Account security page

1. As a site administrator, navigate to the user's settings and select **Account security**.
![](https://storage.googleapis.com/sourcegraph-assets/Docs/how-to/debug-permissions/account_security_page.png)
    - This page shows a user's connected external accounts. An external account is required to sync user permissions from code hosts.
2. If a user has not connected their code host account, the user will have to do that by visiting their Account security settings and clicking on Add. Without a connected code host account, Sourcegraph is not able to associate their Sourcegraph identity with a code host identity to map permissions. If you don’t see the provider for your code host here, you have to [configure it first](https://sourcegraph.com/docs/admin/auth).
3. If the account is connected, you can revisit the Permissions page, schedule another permissions sync, and see if anything changes. If nothing changes when a permission sync is complete, check the outbound requests log to verify that code host API calls are succeeding.

## Check the outbound requests log

1. On the Site admin settings page, click on the Outbound requests option near the bottom.
![](https://storage.googleapis.com/sourcegraph-assets/Docs/how-to/debug-permissions/outbound_requests_option.png)
2. If outbound requests aren’t enabled, you’ll need to enable it first. In the site configuration, add `"outboundRequestLogLimit": 50`, (50 indicates the number of most recent requests to store. Adjust as needed).
3. The Outbound requests page will show a list of all the HTTP requests Sourcegraph is making to outside services. Calls made to do user/repo permissions syncs will also show up here. You can try to catch a user permission sync "in the act" by scheduling a sync and visiting the Outbound requests page.
![](https://storage.googleapis.com/sourcegraph-assets/Docs/how-to/debug-permissions/outbound_requests.png)
4. Here, we can see a bunch of API calls being made to the GitLab API, and they seem to be succeeding. To stop the list from refreshing every 5 seconds, you can click on the Pause updating button, and to see more details, you can click on the More info option.
    - You'll need to use your judgment to determine if anything looks out of place. Ideally, you'd like to see calls to the code host. You want to see those calls succeed with a 200 OK status code. Look at the URLs to see if they make sense. Does anything seem weird?

## Configure authorization for a code host connection

1. Select the Code host connections section on the Site admin settings page. You should see a list of code host connections.
![](https://storage.googleapis.com/sourcegraph-assets/Docs/how-to/debug-permissions/code_host_connections.png)
2. Select the code host that you’d like to view, which should show you its configuration page.
![](https://storage.googleapis.com/sourcegraph-assets/Docs/how-to/debug-permissions/gitlab_connection.png)
    - Here you can see that this GitLab connection is configured with authorization. If the authorization field is missing completely then the code host connection will be marked as “unrestricted” and all repositories synced by that code host connection will be accessible to everyone on Sourcegraph.
    - Not all authorization configurations are the same. Visit the [code hosts docs page](https://sourcegraph.com/docs/admin/external_service#configure-a-code-host-connection) for more detailed information on configuring authorization for the specific code host you’re dealing with.
3. Once you’ve adjusted and saved the config, a sync should be triggered for the code host. If no sync was triggered, hit the Trigger manual sync button.
4. Next, check the user permissions screen to see if things are behaving as expected now, and if nothing has changed, try scheduling a permissions sync.

## Is the repo public on the code host but not public on Sourcegraph or vice versa?

1. Confirm that the repository on the code host has the visibility you expect it to have.
    - Does the code host even have a concept of visibility? Some enterprise code hosts don’t have a concept of “publically available” repositories.
2. Confirm that the visibility of the repository on Sourcegraph differs from what you saw on the code host itself.
![](https://storage.googleapis.com/sourcegraph-assets/Docs/how-to/debug-permissions/repo_search.png)
    - Private repositories on Sourcegraph will have the Private label as displayed above. If that label is not there, consider the repo to be public.
3. If there is a discrepancy between the visibility of the repository on the code host vs. Sourcegraph, it could point to a bug on Sourcegraph’s side.

## Is the repository actually present on Sourcegraph?

1. Navigate to the **Site admin** settings and select **Repositories** under the **Repositories** section in the navigation menu.
![](https://storage.googleapis.com/sourcegraph-assets/Docs/how-to/debug-permissions/repositories_option.png)
2. Search for the repository in question.
![](https://storage.googleapis.com/sourcegraph-assets/Docs/how-to/debug-permissions/repo_search.png)
3. If the repository does not show up, but you expect it to, the instance may be configured to enforce repository permissions even for site admins. In this case, the site admin can only see the repository if they have explicit access. It may be worth temporarily turning off this enforcement for debugging purposes.
4. To disable permissions enforcement for site administrators, navigate to the Site Configuration section.
![](https://storage.googleapis.com/sourcegraph-assets/Docs/how-to/debug-permissions/site_config_option.png)
5. In the site configuration, there will be a line that reads `"authz.enforceForSiteAdmins": true` . Set it to false. Attempt to search for the repository again.
    - If there is no such entry, or if the setting is already set to false, the repository is not synced to Sourcegraph at all, and the fault is most likely in the code host connection.
