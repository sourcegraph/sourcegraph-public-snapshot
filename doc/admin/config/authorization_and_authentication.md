# Authentication and authorization in Sourcegraph

Sourcegraph has two authentication concepts:

1. The system through which your users log in (SAML IdP, OAuth with a code host, username/password, OpenID Connect, Authentication Proxy)
2. The system which controls user permissions (Code host, explicit permissions API in Sourcegraph)

We suggest configuring both when using Sourcegraph Enterprise. If you do not configure permissions, all users will be able to see all of the code in the instance.

## Authentication in Sourcegraph

Sourcegraph supports username/password auth by default and SAML, OAuth, HTTP Proxy auth, and OpenID Connect if configured. Changing a username in Sourcegraph will allow the user to escalate permissions, so if you are syncing permissions, you will need to add the following to your site config at `https://sourcegraph.yourdomain.com/siteadmin/configuration` ([Learn more about viewing and editing your site configuration.](./site_config.md#view-and-edit-site-configuration))

```
auth.enableUsernameChanges: false
```

For users using any of the other authentication mechanisms, removing `builtin` as an authentication mechanism is best practice. (Customers in a managed instance environment will need to leave `builtin` enabled for Sourcegraph employee access. Consult with your Customer Engineer for more info.)

> NOTE: If Sourcegraph is running on a free license all users will be created as site admins. Learn more about license settings on our [pricing page](https://about.sourcegraph.com/pricing).

## Authorization in Sourcegraph

If you use GitHub, GitLab, or Bitbucket Server / Bitbucket Data Center, you can sync access permissions directly from the code host:

- [GitHub](#github-enterprise-or-github-cloud-authentication-and-authorization)
- [GitLab](#gitlab-enterprise-or-gilab-cloud-authentication-and-authorization)
- [Bitbucket Server / Bitbucket Data Center](#bitbucket-server-authorization)

If you do not use one of those listed code hosts, you will need to control access using our [explicit permissions API](#explicit-permissions-api-authorization).

## Username normalization

Sourcegraph does not support email-style usernames. In contexts in which you are providing the username to Sourcegraph, it will be normalized, which can impact username matching where that is relevant. See [full documentation for username normalization](../auth/index.md#username-normalization).

## Instructions for individual authentication/authorization pathways

We recommend that you start at the instructions for your code host, if listed, for a complete explanation of the authentication/authorization options available to you when using that code host.

### Built-in username/password authentication

Built-in username/password authentication is Sourcegraph’s default authentication option. To enable it, add the following to your site config:

```
{
  // Other config settings go here
  "auth.providers": [
    {
      "type": "builtin",
      "allowSignup": true 
    }
  ]
}
```

Set `allowSignup` to `false` if you want to create user accounts instead of allowing the user to create their own.

Learn more about [built-in password authentication](../auth/index.md#builtin-password-authentication). 

### GitHub Enterprise or GitHub Cloud authentication and authorization

We support both authentication and permissions syncing (through OAuth) for GitHub. If you use GitHub as your code host, we do not officially support using another authentication mechanism (SAML, etc.). Instead, you’ll need to follow this flow:

1. Use SAML (or another auth mechanism) to log in to GitHub
2. Use GitHub OAuth to log in to Sourcegraph

In this way, access to Sourcegraph will still be managed by your identity provider, using the code host as a middle step.

Follow these steps to [configure authentication with GitHub via OAuth](../auth/index.md#github). 

Once authentication with GitHub via OAuth is configured, follow [these steps to configure access permissions](../repo/permissions.md#github). Users will log into Sourcegraph using Github OAuth, and permissions will be synced in the background.

### GitLab Enterprise or GitLab Cloud authentication and authorization

We support both authentication and permissions syncing (through OAuth) for GitLab. If you use GitLab as your code host, you have two available authentication flows:

#### Option 1

1. Use SAML (or another auth mechanism) to log in to GitLab
2. Use GitLab OAuth to log in to Sourcegraph

In this way, access to Sourcegraph will still be managed by your identity provider, using the code host as a middle step. This option is the simplest to configure. To do so, [set up GitLab as an authentication option](../auth/index.md#gitlab), and then [enable permissions syncing](../repo/permissions.md#oauth-application).

#### Option 2

Alternatively, you can configure SAML authentication in Sourcegraph, and use GitLab permissions syncing in the background to control access permissions. To implement this method, you will need to make sure that GitLab is able to return a value in `identities.provider` for the `GET /users` endpoint ([GitLab documentation](https://docs.gitlab.com/ee/api/users.html#for-admins)) that your identity provider is able to pass as the `nameID` in the SAML response. If that isn’t possible, you will need to use the first option. 

To configure SAML auth with GitLab permissions, you will need to first [configure permissions from GitLab](../repo/permissions.md#administrator-sudo-level-access-token). Then, [configure SAML authentication](../auth/saml/index.md). The `nameID` passed by the identity provider will need to match the value of `identities.provider`. 

For example, if the GitLab API returns:

```json
"identities": [
   {"provider": "saml", "extern_uid": "email@domain.com"}
]
```

Then you will need to configure permission in Sourcegraph as:

```
{
  "url": "https://gitlab.com",
  "token": "$PERSONAL_ACCESS_TOKEN",
  "authorization": {
    "identityProvider": {
      "type": "external",
      "authProviderID": "$AUTH_PROVIDER_ID",
      "authProviderType": "$AUTH_PROVIDER_TYPE",
      "gitlabProvider": "saml"
    }
  }
}
```

And configure the identity provider to pass the email address as the `nameID`. 

### Bitbucket Server / Bitbucket Data Center authorization

We do not currently support OAuth for Bitbucket Server or Bitbucket Data Center. You will need to combine permissions syncing from Bitbucket Server / Bitbucket Data Center with another authentication mechanism (SAML, built-in auth, HTTP authentication proxies). Bitbucket Server and Bitbucket Data Center only pass usernames to Sourcegraph, so you’ll need to make sure that those usernames are matched by whatever mechanism you choose to use for access.

Follow the steps to [sync Bitbucket Server / Bitbucket Data Center permissions](../repo/permissions.md#bitbucket-server). Then, do one of the following:

1. Create the user accounts in Sourcegraph with matching usernames. (Access using `builtin` auth.)
2. [Configure SAML authentication](../auth/saml/index.md). If you are using Bitbucket Server / Bitbucket Data Center, the `login` attribute is *not* optional. You need to pass the Bitbucket Server username as the `login` attribute. 
3. [Configure an HTTP authentication proxy](../auth/index.md#http-authentication-proxies), passing the Bitbucket Server username value as the `usernameHeader`. 

### Explicit Permissions API authorization

With any authentication mechanism, you can use our GraphQL API to set permissions for all repositories. If you choose to do this, this is the *only* mechanism that can be used for permissions—all others will be ignored. Follow the instructions for the [mutations needed within the GraphQL API](../repo/permissions.md#explicit-permissions-api) to configure access.

### OpenID Connect authentication

Use OpenID Connect authentication if accessing using OpenID Connect, such as when logging in through a Google Workspace, or if other authentication methods aren’t an option. See [how to set up OpenID Connect authentication](../auth/index.md#openid-connect).

### HTTP authentication proxy authentication

HTTP authentication proxy authentication is not generally recommended unless it's a common authentication process within your organization. See [how to configure HTTP authentication proxies](../auth/index.md#http-authentication-proxies).
