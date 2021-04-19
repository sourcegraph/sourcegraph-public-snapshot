# Configuring credentials

> NOTE: This page describes functionality added in Sourcegraph 3.22. Older Sourcegraph versions only allow batch changes to be applied and managed by site admins.

In order to [publish changesets with Batch Changes](publishing_changesets.md), you need to add a personal access token for each code host that your batch change interacts with. These tokens are used by Sourcegraph to create and manage changesets on behalf of yourself, and with your specific permissions, on the code host. Since Sourcegraph 3.27, it is also possible to configure a global service account per code host to be used when the user doesn't have credentials configured.

## Requirements

- Sourcegraph instance with repositories in it. See the "[Quickstart](../../index.md#quickstart)" guide on how to setup a Sourcegraph instance.

## Adding a personal access token

Access tokens can be configured either for your user account, or globally, if you're a site admin of the Sourcegraph instance.

### For yourself

Adding personal access tokens is done through the Batch Changes section of your user settings:

1. From any Sourcegraph page, click on your avatar at the top right of the page.
1. Select **Settings** from the dropdown menu.
1. Click **Batch Changes** on the sidebar menu.

You should now see a list of the code hosts that are configured on Sourcegraph. Code hosts with tokens configured are indicated by a green tick, while code hosts without tokens have an empty red circle next to them. If a global access token has been configured, it is not required (but you can still do it, to create the changesets under your name) to do this. The UI will inform you if that's the case.

### Global service account

Configuring a global service account is done through the Batch Changes section of the site admin area: _(Site admins only)_

1. From any Sourcegraph page, click on your avatar at the top right of the page.
1. Select **Site admin** from the dropdown menu.
1. Click **Batch Changes** on the sidebar menu.

You should now see a list of the code hosts that are configured on Sourcegraph. Code hosts with tokens configured are indicated by a green tick, while code hosts without tokens have an empty red circle next to them. Credentials that are configured here will be usable by all users of the Sourcegraph instance for publishing and updating changesets on the code host.

### Configuring a code host

<video width="1920" height="1080" autoplay loop muted playsinline controls style="width: 100%; height: auto; max-width: 50rem">
  <source src="https://sourcegraphstatic.com/docs/videos/batch_changes/user-tokens.webm" type="video/webm">
  <source src="https://sourcegraphstatic.com/docs/videos/batch_changes/user-tokens.mp4" type="video/mp4">
</video>

To add a token for a code host, click on the **Add token** button next to its name. This will display an input modal like the following:

<img class="screenshot" src="https://sourcegraphstatic.com/docs/images/batch_changes/user-token-input.png" alt="An input dialog, titled &quot;Github Batch Changes token for https://github.com&quot;, with an input box to type or paste a token and a list of scopes that must be enabled on the token, which are repo, read:org, user:email, and read:discussion">

To create a personal access token for a specific code host provider, please refer to the relevant section for "[GitHub](#github)", "[GitLab](#gitlab)", or "[Bitbucket Server](#bitbucket-server)". Once you have a token, you should paste it into the Sourcegraph input shown above, and click **Add token**.

> NOTE: See ["Code host interactions in Batch Changes"](../explanations/permissions_in_batch_changes.md#code-host-interactions-in-batch-changes) for details on what the permissions are used for.

Once this is done, Sourcegraph should indicate that you have a token with a green tick:

<img class="screenshot" src="https://sourcegraphstatic.com/docs/images/batch_changes/one-token.png" alt="A list of code hosts, with GitHub indicating that it has a token and the other hosts indicating that they do not">

### SSH access to code host

When Sourcegraph is configured to [clone repositories using SSH via the `gitURLType` setting](../../admin/repo/auth.md), an SSH keypair will be generated for you and the public key needs to be added to the code host to allow push access. In the process of adding your personal access token you will be given that public key. You can also come back later and copy it to paste it in your code hosts SSH access settings page.

<img class="screenshot" src="https://sourcegraphstatic.com/docs/images/batch_changes/create-credential-ssh-key.png" alt="Credentials setup process, showing the SSH public key to be copied">

### GitHub

In addition to the below, you should refer to [GitHub's documentation on creating a personal access token](https://docs.github.com/en/free-pro-team@latest/github/authenticating-to-github/creating-a-personal-access-token).

Sourcegraph requires the `repo`, `read:org`, `user:email`, and `read:discussion` scopes to be enabled on the user token. This is done by selecting the relevant checkboxes when creating the token:

<img class="screenshot" src="https://sourcegraphstatic.com/docs/images/batch_changes/github-token.png" alt="The GitHub token creation page, with the repo scope selected">

### GitLab

In addition to the below, you should refer to [GitLab's documentation on creating a personal access token](https://docs.gitlab.com/ee/user/profile/personal_access_tokens.html#creating-a-personal-access-token).

Sourcegraph requires the `api`, `read_repository`, and `write_repository` scopes to be enabled on the user token. This is done by selecting the relevant checkboxes when creating the token:

<img class="screenshot" src="https://sourcegraphstatic.com/docs/images/batch_changes/gitlab-token.png" alt="The GitLab token creation page, with the api, read_repository, and write_repository scopes selected">

### Bitbucket Server

In addition to the below, you should refer to [Bitbucket Server's documentation on creating a personal access token](https://confluence.atlassian.com/bitbucketserver0516/personal-access-tokens-966061199.html?utm_campaign=in-app-help&utm_medium=in-app-help&utm_source=stash#Personalaccesstokens-Generatingpersonalaccesstokens).

Sourcegraph requires the access token to have the `write` permission on both projects and repositories. This is done by selecting the **Write** level in the **Projects** dropdown, and letting it be inherited by repositories:

<img class="screenshot" src="https://sourcegraphstatic.com/docs/images/batch_changes/bb-token.png" alt="The Bitbucket Server token creation page, with Write permissions selected on both the Project and Repository dropdowns">

## Removing a personal access token

Removing personal access tokens is done through the the Batch Changes section of your user settings. To access this page, follow these instructions (also shown in the video below):

1. From any Sourcegraph page, click on your avatar at the top right of the page.
1. Select **Settings** from the dropdown menu.
1. Click **Batch Changes** on the sidebar menu.

You should now see a list of the code hosts that are configured on Sourcegraph. Code hosts with tokens configured are indicated by a green tick, while code hosts without tokens have an empty red circle next to them.

<video width="1920" height="1080" autoplay loop muted playsinline controls style="width: 100%; height: auto; max-width: 50rem">
  <source src="https://sourcegraphstatic.com/docs/videos/batch_changes/removing-user-token.webm" type="video/webm">
  <source src="https://sourcegraphstatic.com/docs/videos/batch_changes/removing-user-token.mp4" type="video/mp4">
</video>

To remove a personal access token for a code host, click **Remove** next to that code host. The code host's indicator will change to an empty red circle to indicate that no token is configured for that code host:

<img class="screenshot" src="https://sourcegraphstatic.com/docs/images/batch_changes/no-tokens.png" alt="A list of code hosts, with all code hosts indicating that they do not have a token">
