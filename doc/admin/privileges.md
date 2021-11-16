# Site administrator privileges

Site administrators have [full administrative access to the Sourcegraph instance.](https://docs.sourcegraph.com/admin/config) In many cases, they also control the [deployment environment.](https://docs.sourcegraph.com/admin/install) Special privileges are granted to site-admin users.

## Access to all repositories

Site administrators are able to access all repositories on the Sourcegraph instance and [manage the settings of individual repositories.](https://docs.sourcegraph.com/admin/repo/permissions)

## Access to all GraphQL APIs

Site administrators are able to access all [GraphQL APIs](https://docs.sourcegraph.com/api/graphql) of the Sourcegraph instance, including special GraphQL APIs that require a [site-admin access token.](https://docs.sourcegraph.com/api/graphql#sudo-access-tokens)

## Impersonate regular users

Site administrators are able to impersonate regular users in two ways:

1. [View and update the settings of any user.](https://docs.sourcegraph.com/admin/config/settings#editing-global-settings-for-site-admins)
2. Initiate GraphQL API calls on behalf of any other user.

## Receive site alerts

Site administrators see update notifications and other [site-level alerts](https://docs.sourcegraph.com/admin/observability/alerting#understanding-alerts) (visible as a banner across the top of the screen) that may be invisible to non-admin users.
