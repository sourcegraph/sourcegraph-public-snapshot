# Troubleshooting user authentication (SSO)

## Basic principles

As of 3.20, Sourcegraph supports 6 authentication methods as listed in our [User authentication (SSO)](https://docs.sourcegraph.com/admin/auth) page, details about how to configure those methods should refer to the relevant document pages.

Among these authentication methods, there are 4 of them that require Sourcegraph to send requests to the authentication provider, including GitHub OAuth, GitLab OAuth, OpenID Connect and SAML.

If a GitHub OAuth is configured in Sourcegraph, users should be redirected to the GitHub instance (GitHub.com, GitHub Enterprise, etc.) for authorizing the Sourcegraph OAuth application upon sign in. Once the user is authenticated on the GitHub instance, then authorized the application, the GitHub instance should redirect the user straight back to the Sourcegraph instance but not anywhere else. 

The reason is that such redirection contains critical and confidential information that Sourcegraph consumes (as an OAuth consumer), and information may be stripped or expired if there is a third-party getting in the way.

Therefore, the **Callback URL/Single Sign On URL** is very important to get right on the authentication provider side. Here is an example of callback URL in a GitHub OAuth application settings:

![GitHub OAuth callback URL example](https://storage.googleapis.com/sourcegraph-assets/docs/images/auth/troubleshooting_callback_url_example.png)

For an OAuth application against GitHub.com (GitHub OAuth), the authentication flow should look like this (-> indicates a browser redirection):

```
   Sourcegraph (choose to sign in with GitHub.com)
-> GitHub.com (authorize the application)
-> Sourcegraph 
```

### The authentication provider always redirects back to its direct consumer

All GitHub OAuth, GitLab OAuth, OpenID Connect and SAML require the admin to configure a **Callback URL/Single Sign On URL** on the authentication provider side, which means the redirection back to the Sourcegraph (the consumer) is critical to make the user authentication work (as explained in the last section).

In a complex authentication set up, there is a chain of authentication. For example, Sourcegraph syncs users from the GitHub instance via GitHub OAuth, and the GitHub instance is in turn syncing users from a company central SSO, say, Okta, via SAML.

In such a setup, the Sourcegraph instance only talks to the GitHub instance, and is not aware of the existence of the Okta. The same is true that Okta does not know that the Sourcegraph is in fact syncing users from it (through GitHub).

Therefore, it makes no sense to have Okta redirect users directly back to Sourcegraph which effectively bypasses the GitHub instance. In a correct set up, the authentication provider should always redirect the user to its direct consumer. Thus, *Okta should only redirect users back to the GitHub instance, and the GitHub instance in turn redirects the user back to Sourcegraph with any information needed for the consumer.*

## Debugging playbook

### Make sure the client ID and client secret are actually correct

Invalid pair of client ID and client secret does not prevent the user from completing the OAuth flow until a confusing error, which gives the user a wrong intuitive that they’re correct.

### Test in incognito mode

It is possible that a browser extension that the user uses modifies/strips a HTTP header or cookie for the authentication. Ask the user to perform the test in incognito mode without any browser extension is preferred. The incognito mode also helps to ensure that the user logs out from all involved parties.

### Preserve log in DevTools

To keep a record of redirection in a user's browser is very helpful to audit the auth flow. Here is how to do it in Chrome:

![Preserve log in Chrome](https://storage.googleapis.com/sourcegraph-assets/docs/images/auth/troubleshooting_preserve_log_in_chrome.png)

It would be easier to eyeball logs if the user does a “clear” before any tests:

![Clear log in Chrome](https://storage.googleapis.com/sourcegraph-assets/docs/images/auth/troubleshooting_clear_log_in_chromejpg)

### Configure SAML in Okta

Things to note:

1. Okta has a weird UX that many options (e.g. SAML templates) are not available in the so-called **Developer Console** UI, make sure first changin to the **Classic UI** on the top-left.

    ![Okta classic UI](https://storage.googleapis.com/sourcegraph-assets/docs/images/auth/troubleshooting_okta_classic_ui.png)

2. The Okta docs must be opened directly from the application settings because it contains a special URL parameter “baseAdminUrl” for displaying confidential information in the docs:

    ![Okta view instruction](https://storage.googleapis.com/sourcegraph-assets/docs/images/auth/troubleshooting_okta_view_instruction.png)
    ![Okta view instruction 2](https://storage.googleapis.com/sourcegraph-assets/docs/images/auth/troubleshooting_okta_view_instruction_2.png)

    Without the URL parameter `baseAdminUrl`, it displays a static warning:

    ![Okta static warning](https://storage.googleapis.com/sourcegraph-assets/docs/images/auth/troubleshooting_okta_static_warning.png)

### Sourcegraph specific cookies

For GitHub OAuth, Sourcegraph uses a special cookie called `github-state-cookie`, which should be sent by browser automatically upon the user is being redirected back to Sourcegraph (the callback URL):

![Sourcegraph state cookie](https://storage.googleapis.com/sourcegraph-assets/docs/images/auth/troubleshooting_sourcegraph_state_cookie.png)

If the cookie is not found in the **Request Headers**, then there is a problem. Maybe due to proxy server stripping out any unrecognized cookies (so the browser won’t be able to get it back from the Sourcegraph instance in the first place).

### Things to collect for async analyzes

- `"auth.providers"` portion of the Sourcegraph site configuration. Example:

    !["auth.providers" site config](https://storage.googleapis.com/sourcegraph-assets/docs/images/auth/troubleshooting_sourcegraph_auth_providers.png)

- A screenshot of the Sourcegraph OAuth application on the GitHub instance. Example:

    ![OAuth application settings](https://storage.googleapis.com/sourcegraph-assets/docs/images/auth/troubleshooting_oauth_application_settings.png)

- A screenshot of the GitHub authentication settings.
  - For GitHub Enterprise Server, it is in the management console's **Settings > Authentication**. Example:

        ![GitHub Enterprise Server SAML settings](https://storage.googleapis.com/sourcegraph-assets/docs/images/auth/troubleshooting_ghe_saml_settings.png)

  - For GitHub Enterprise Cloud, it is in the organization's **Settings > Organization security > SAML single sign-on**. Example:

        ![GitHub Enterprise Cloud SAML settings](https://storage.googleapis.com/sourcegraph-assets/docs/images/auth/troubleshooting_github_enterprise_cloud_saml_settings.png)

- Screenshots of the Okta application settings on both General and Sign On tabs, with all options from top to bottom.
- Full browser logs (https://toolbox.googleapps.com/apps/har_analyzer/)
