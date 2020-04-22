# Critical configuration

> NOTE: Since Sourcegraph v3.11, critical configuration has been removed. All critical configuration options are now in the site configuration. See the [migration notes for Sourcegraph v3.11+](../migration/3_11.md) for more information.

Critical configuration defines how critical Sourcegraph components behave, such as the external URL and user authentication. Unlike normal [site configuration](site_config.md), incorrect critical configuration can make Sourcegraph's web interface unreachable. Therefore, critical configuration must be edited on the failsafe [management console](../management_console.md), to allow recovery from misconfiguration.

## View and edit critical configuration

Critical configuration must be edited on the [management console](../management_console.md). See "[Accessing the management console](../management_console.md#accessing-the-management-console)" for more information.

## Reference

All critical configuration options and their default values are shown below.

> NOTE: Not finding the option you're looking for? It may be a normal [site configuration](site_config.md) option, which means it must be set in **Site admin > Configuration** on the main Sourcegraph web interface.

<div markdown-func=jsonschemadoc jsonschemadoc:path="admin/config/critical.schema.json">[View page on docs.sourcegraph.com](https://docs.sourcegraph.com/doc/admin/config/critical_config) to see rendered content.</div>

## Authentication providers `auth.providers`

The `auth.providers` critical configuration property defines how users can authenticate to the Sourcegraph instance. It is an array of values, each of which defines an authentication provider. If the array has more than one element, users may use any of the configured authentication methods.

All authentication providers support the (optional) `displayName` property, which is used to distinguish the authentication provider when there are multiple providers configured.

## Builtin password authentication

Defines an authentication provider that stores and validates passwords for each user account. It also allows users (and site admins) to reset passwords.

To use this authentication method, add an element to the `auth.providers` array with the following shape:

<div markdown-func=jsonschemadoc jsonschemadoc:path="admin/config/critical.schema.json" jsonschemadoc:ref="#/definitions/BuiltinAuthProvider">[View page on docs.sourcegraph.com](https://docs.sourcegraph.com/doc/admin/config/critical_config) to see rendered content.</div>

## SAML

Defines an authentication provider backed by SAML.

Note: if you are using IdP-initiated login, you must have _at most one_ SAML authentication provider in the `auth.providers` array.

To use this authentication method, add an element to the `auth.providers` array with the following shape:

<div markdown-func=jsonschemadoc jsonschemadoc:path="admin/config/critical.schema.json" jsonschemadoc:ref="#/definitions/SAMLAuthProvider">[View page on docs.sourcegraph.com](https://docs.sourcegraph.com/doc/admin/config/critical_config) to see rendered content.</div>

## OpenID Connect (including G Suite)

Defines an authentication provider backed by OpenID Connect. The most common case is G Suite (Google) authentication.

To use this authentication method, add an element to the `auth.providers` array with the following shape:

<div markdown-func=jsonschemadoc jsonschemadoc:path="admin/config/critical.schema.json" jsonschemadoc:ref="#/definitions/OpenIDConnectAuthProvider">[View page on docs.sourcegraph.com](https://docs.sourcegraph.com/doc/admin/config/critical_config) to see rendered content.</div>

## HTTP authentication proxy

Defines an authentication provider that authenticates users by consulting an HTTP request header set by an authentication proxy, such as https://github.com/bitly/oauth2_proxy.

To use this authentication method, add an element to the `auth.providers` array with the following shape:

<div markdown-func=jsonschemadoc jsonschemadoc:path="admin/config/critical.schema.json" jsonschemadoc:ref="#/definitions/HTTPHeaderAuthProvider">[View page on docs.sourcegraph.com](https://docs.sourcegraph.com/doc/admin/config/critical_config) to see rendered content.</div>

## Known bugs

The following critical configuration options require the server to be restarted for the changes to take effect:

```
lightstepAccessToken
lightstepProject
auth.userOrgMap
auth.providers
update.channel
useJaeger
```
