# Configuring credentials

Interacting with a code host (such as creating, updating, or syncing changesets) is made possible by configuring an access token for that code host. Sourcegraph uses these tokens to manage changesets on your behalf, and with your specific permissions.

## Requirements

- Sourcegraph instance with repositories in it. See the "[Quickstart](../../index.md#quick-install)" guide on how to setup a Sourcegraph instance.
- Account on the code host with access to the repositories you wish to target with your batch changes.

## Types of access tokens used by Batch Changes

There are two types of access token that can be configured for use with Batch Changes:

1. [**Personal access token**](#personal-access-tokens) - A token set by an individual Batch Changes user for their personal code host user account.
1. [**Global service account token**](#global-service-account-tokens) (*Admins only*) - A token that can be used by any Batch Changes user who does not have a personal access token configured. These are also required for [importing changesets](./tracking_existing_changesets.md) and syncing changeset state from the code host when webhooks are not configured.

Different tokens are used for different types of operations, as is illustrated in the hierarchy table below.

🟢  **Preferred** - Sourcegraph will prefer to use this token for this operation, if it is configured.

🟡  **Fallback** - Sourcegraph will fall back to use this token for this operation, if it is configured.

🔴  **Unsupported** - Sourcegraph cannot use this token for this operation.

Operation | [Personal Access Token](#personal-access-tokens) | [Global Service Account Token](#global-service-account-tokens)
--------- | :-: | :-:
Pushing a branch with the changes | 🟢 | 🟡
[Publishing a changeset](./publishing_changesets.md) | 🟢 | 🟡
Updating a changeset | 🟢 | 🟡
Closing a changeset | 🟢 | 🟡
[Importing a changeset](./tracking_existing_changesets.md) | 🔴 | 🟢
Syncing a changeset | 🔴 | 🟢

When writing a changeset to the code host, the author will reflect the token used (e.g., on GitHub, the pull request author will be you). It is for this reason that a personal access token is preferred for most operations.

## Personal access tokens

### Do I need to add a personal access token?

Personal access tokens are not strictly required if a global access token has also been configured, but users should add one if they want Sourcegraph to create changesets under their name.

> NOTE: Commit author is determined by your spec file or local git config at the time of running `src batch [apply|preview]`, completely independent from code host credentials.

### Adding a token

Adding a personal access token is done through the Batch Changes section of your user settings:

1. From any Sourcegraph page, click on your avatar at the top right of the page.
1. Select **Settings** from the dropdown menu.
1. Click **Batch Changes** on the sidebar menu.
1. Click **Add credentials** and follow the steps to [create a new token](#creating-a-code-host-token) for the code host

<video width="1920" height="1080" autoplay loop muted playsinline controls style="width: 100%; height: auto; max-width: 50rem">
  <source src="https://sourcegraphstatic.com/docs/videos/batch_changes/user-tokens.webm" type="video/webm">
  <source src="https://sourcegraphstatic.com/docs/videos/batch_changes/user-tokens.mp4" type="video/mp4">
</video>

Code hosts with tokens configured are indicated by a green tick to the left of the code host name, while code hosts without tokens have an empty red circle next to them.

### Removing a token

To remove a token, navigate back to the same section of your user settings, then click **Remove**:

<video width="1920" height="1080" autoplay loop muted playsinline controls style="width: 100%; height: auto; max-width: 50rem">
  <source src="https://sourcegraphstatic.com/docs/videos/batch_changes/removing-user-token.webm" type="video/webm">
  <source src="https://sourcegraphstatic.com/docs/videos/batch_changes/removing-user-token.mp4" type="video/mp4">
</video>

The code host's indicator should revert to the empty red circle once the token is removed.

## Global service account tokens

### Do I need to add a global service account token?

Global credentials are usable by all users of the Sourcegraph instance who have not added their own personal access tokens for Batch Changes. This makes them a handy fallback, but not strictly required if users are adding their own tokens for publishing changesets.

However, currently a global service account token is required for [importing existing changesets](./tracking_existing_changesets.md) on your code hosts into batch changes.

Additionally, if you have not [configured webhooks](../../admin/config/batch_changes.md#incoming-webhooks) from your code host,  Sourcegraph requires a global service account keep changesets up to date.

If [forks are enabled](../../admin/config/batch_changes.md#forks), then note that repositories will also be forked into the service account.

### Adding a token

Adding a global service account token is done through the Batch Changes section of the site admin area:

1. From any Sourcegraph page, click on your avatar at the top right of the page.
1. Select **Site admin** from the dropdown menu.
1. Click **Batch Changes** on the sidebar menu.
1. Click **Add credentials** and follow the steps to [create a new token](#creating-a-code-host-token) for the code host

Code hosts with tokens configured are indicated by a green tick to the left of the code host name, while code hosts without tokens have an empty red circle next to them.

### Removing a token

To remove a token, navigate back to the same section of the site admin area, then click **Remove**. The code host's indicator should revert to the empty red circle once the token is removed.

## Creating a code host token

To finish configuring the new credentials, you will need to create a new personal access token on your code host and paste it into the input field on the **Add credentials** modal:

<img class="screenshot" src="https://sourcegraphstatic.com/docs/images/batch_changes/user-token-input.png" alt="An input dialog, titled &quot;Batch Changes credentials: GitHub&quot;, with an input box to type or paste a token and a list of scopes that must be enabled for this type of code host token">

### GitHub

#### GitHub.com

On GitHub.com, [you can create a code host token with the correct scopes at this link](https://github.com/settings/tokens/new?scopes=repo,read:org,user:email,read:discussion,workflow).

When working with organizations that have SAML SSO (Single Sign On) enabled, configuring credentials requires an additional step that [involves white-listing the token for use in that organization](https://docs.github.com/en/enterprise-cloud@latest/authentication/authenticating-with-saml-single-sign-on/authorizing-a-personal-access-token-for-use-with-saml-single-sign-on).

#### GitHub Enterprise

Follow the steps to [create a personal access token](https://docs.github.com/en/free-pro-team@latest/github/authenticating-to-github/creating-a-personal-access-token) on GitHub Enterprise. Batch Changes requires the following scopes:

- `repo`
- `read:org`
- `user:email`
- `read:discussion`
- `workflow`

This is done by selecting the relevant checkboxes when creating the token:

<img class="screenshot" src="https://sourcegraphstatic.com/docs/images/batch_changes/github-token.png" alt="The GitHub token creation page, with the repo scope selected">

> NOTE: `workflow` is technically only required if your batch changes modify files in the `.github` directory of a repository, but we recommend enabling it regardless to avoid confusing errors at a later time.

When working with organizations that have SAML SSO (Single Sign On) enabled, configuring credentials requires an additional step that [involves white-listing the token for use in that organization](https://docs.github.com/en/enterprise-cloud@latest/authentication/authenticating-with-saml-single-sign-on/authorizing-a-personal-access-token-for-use-with-saml-single-sign-on).

### GitLab

Follow the steps to [create a personal access token](https://docs.gitlab.com/ee/user/profile/personal_access_tokens.html#creating-a-personal-access-token) on GitLab. Batch Changes requires the following scopes:

- `api`
- `read_repository`
- `write_repository`

This is done by selecting the relevant checkboxes when creating the token:

<img class="screenshot" src="https://sourcegraphstatic.com/docs/images/batch_changes/gitlab-token.png" alt="The GitLab token creation page, with the api, read_repository, and write_repository scopes selected">

### Bitbucket Server / Bitbucket Data Center

Follow the steps to [create a personal access token](https://confluence.atlassian.com/bitbucketserver0516/personal-access-tokens-966061199.html?utm_campaign=in-app-help&utm_medium=in-app-help&utm_source=stash#Personalaccesstokens-Generatingpersonalaccesstokens) on Bitbucket.

Batch Changes requires the access token to have the `write` permission on both projects and repositories. This is done by selecting the **Write** level in the **Projects** dropdown, and letting it be inherited by repositories:

<img class="screenshot" src="https://sourcegraphstatic.com/docs/images/batch_changes/bb-token.png" alt="The Bitbucket Server / Bitbucket Data Center token creation page, with Write permissions selected on both the Project and Repository dropdowns">

### Bitbucket Cloud

Follow the steps to [create an app password](https://support.atlassian.com/bitbucket-cloud/docs/app-passwords/) on Bitbucket. Batch Changes requires the following scopes:

- `account:read`
- `repo:read`
- `repo:write`
- `pr:write`
- `pipeline:read`

<img class="screenshot" src="https://sourcegraphstatic.com/docs/images/batch_changes/bb-cloud-app-password.png" alt="The Bitbucket Cloud app password creation page">

### SSH access to code host

When Sourcegraph is configured to [clone repositories using SSH via the `gitURLType` setting](../../admin/repo/auth.md), an SSH keypair will be generated for you and the public key needs to be added to the code host to allow push access. In the process of adding your personal access token you will be given that public key. You can also come back later and copy it to paste it in your code hosts SSH access settings page.

<img class="screenshot" src="https://sourcegraphstatic.com/docs/images/batch_changes/create-credential-ssh-key.png" alt="Credentials setup process, showing the SSH public key to be copied">


