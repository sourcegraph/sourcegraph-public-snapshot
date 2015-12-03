+++
title = "Authentication"
description = "Manage user authentication on your Sourcegraph instance"
+++

## OAuth

The default authentication mechanism for Sourcegraph is using OAuth2 via Sourcegraph.com. Users can log into Sourcegraph instances with their Sourcegraph.com account credentials. This requires no configuration from the admin.

See [OAuth2.md]({{< relref "dev/OAuth2.md" >}}) for details of the OAuth2 implementation.

## LDAP

Sourcegraph supports authentication via a configured LDAP server. The LDAP server can be internal to your company's network, but it must be accessible from your local Sourcegraph instance.

To use LDAP for authentication, the user entries in the directory service must contain email addresses. These email addresses will used to link the LDAP account to a Sourcegraph.com account. If the email field in the directory service can be modified by the user, it is not recommended to use LDAP authentication with Sourcegraph as that is a security vulnerability.

To set up LDAP auth, add the following fields to the config file `/etc/sourcegraph/config.ini`:

```
[serve.Authentication]
Source = ldap

[serve.LDAP]
# the host name or IP of your LDAP server
Host = ldap.mycompany.org
# the port on which your LDAP server is listening on
Port = 389
# TLS is optional
TLS = true
# a service account used to make LDAP search requests
SearchUser = cn=serviceuser,ou=svcaccts,dc=mycompany,dc=org
# service account password
SearchPassword = mysecret
# The point in the LDAP tree where users are searched from (optional)
DomainBase = dc=mycompany,dc=org
# Filter the search query to this subtree (optional)
Filter = ou=useraccounts
# The LDAP field mapped to the user's login (usually uid or cn) (required)
UserIDField = uid
# The LDAP field mapped to the user's email (required)
EmailField = mail
```

Restart the server `sudo restart src` for these settings to take effect. If the configured LDAP server is unreachable, Sourcegraph will throw a fatal error and exit. You can view the error log in the file `/var/log/upstart/src.log`.

### Access control

Currently, LDAP authenticated Sourcegraph instances have limited support for controlling access to the Sourcegraph server

1. Admin users: Configure a user as admin by first logging in to your Sourcegraph via the web interface with your LDAP credentials, and then modifying the users file `$SGPATH/db/users.json` on the Sourcegraph server, to add a `Admin = true` field in your user entry.

2. Use the Filter field (described above) to restrict access to subset of LDAP users. The string specified in the Filter will be ANDed with the search term. For instance, with `Filter = ou=devs` the search query will be `(&(ou=devs)(uid=username))`. If the Filter is malformed, the LDAP search requests will not succeed and users will not be able to log in to Sourcegraph.

3. By default, users who can sign in with their LDAP credentials to your Sourcegraph will have read+write access. To restrict write access to only those users that you specified as admins, add the following flag in the config file and restart the server:

```
[serve.Auth]
RestrictWriteAccess = true
```
