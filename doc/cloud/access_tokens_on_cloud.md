# Security recommendations for code host access tokens on Sourcegraph Cloud

Both GitHub and GitLab code host connections for organizations require an access token. There are two types of tokens you can supply:

- **Machine user token** (recommended):  
This is a personal access token generated for a “machine user” that is only affiliated with an organization. This gives the code host connection access to only the repositories the machine user is granted access to.
- **Personal access token** (not recommended):  
This gives the code host connection the same level of access to repositories as the account that created the token, including all public and private repositories associated with the account.

We recommend setting up a machine user on your code host to configure your organization’s code host connections.

**Using your own personal access token from your code host in your organization’s code host connection may reveal your public and private repositories to other members of your organization.** This is because during early access for organizations on Sourcegraph Cloud, all members of your organization have administration access to the organization settings, including which repositories are synced to Sourcegraph Cloud. The list of repositories available to be synced to Sourcegraph Cloud is determined through the access token associated with the code host connection.

For further instructions and information about personal access tokens and setting up a machine user, please see:

- GitHub: [Machine users in GitHub docs](https://developer.github.com/v3/guides/managing-deploy-keys/#machine-users)
- GitHub: [Personal access tokens](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/creating-a-personal-access-token)
- GitLab: [Personal access tokens](https://docs.gitlab.com/ee/security/token_overview.html#security-considerations)

Looking for Sourcegraph API access tokens? See [creating a Sourcegraph API access token](../cli/how-tos/creating_an_access_token.md)
