+++
title = "Authentication"
description = "Manage user authentication on your Sourcegraph instance"
+++

## Local auth

The default authentication mechanism for Sourcegraph is via username & password.
Passwords are hashed and stored on your file system or database, depending
on your [persistence configuration]({{< relref "config/localstore.md" >}}).

The first user to create an account on a Sourcegraph instance becomes the instance
admin. The admin may [invite other users and manage access controls]
({{< relref "management/access-control.md" >}}).

### Resetting passwords

If an SMTP server is not configured for your Sourcegraph server, admins must generate password reset links for users who lose or need to change their password, by running this command:

	src user reset-password -e <email@domain.com>

or

	src user reset-password -l <login>

where `<email@domain.com>` is the email address and `<login>` is the login name associated with the registered user account. This will output a reset link which should be given to the user.


## LDAP

Sourcegraph supports authentication via a configured LDAP server. The LDAP server can be internal to your company's network, but it must be accessible from your local Sourcegraph instance.

To use LDAP for authentication, the user entries in the directory service must contain email addresses. These email addresses will used to link the LDAP account to a Sourcegraph.com account. If the email field in the directory service can be modified by the user, it is not recommended to use LDAP authentication with Sourcegraph as that is a security vulnerability.

To set up LDAP auth, add the following fields to the config file `/etc/sourcegraph/config.ini`:

```
[serve]
auth.source = ldap

# the host name or IP of your LDAP server
ldap.host = ldap.mycompany.org
# the port on which your LDAP server is listening on
ldap.port = 389
# TLS is optional
ldap.tls = true
# a service account used to make LDAP search requests
ldap.search-user = cn=serviceuser,ou=svcaccts,dc=mycompany,dc=org
# service account password
ldap.search-password = mysecret
# The point in the LDAP tree where users are searched from (optional)
ldap.domain-base = dc=mycompany,dc=org
# Filter the search query to this subtree (optional)
ldap.filter = ou=useraccounts
# The LDAP field mapped to the user's login (usually uid or cn) (required)
ldap.user-id = uid
# The LDAP field mapped to the user's email (required)
ldap.email = mail
```

Restart the server `sudo restart src` for these settings to take effect. If the configured LDAP server is unreachable, Sourcegraph will throw a fatal error and exit. You can view the error log in the file `/var/log/upstart/src.log`.

### Access control

Currently, LDAP authenticated Sourcegraph instances have limited support for controlling access to the Sourcegraph server

1. Admin users: Configure a user as admin by first logging in to your Sourcegraph via the web interface with your LDAP credentials, and then modifying the users file `$SGPATH/db/users.json` on the Sourcegraph server, to add a `Admin = true` field in your user entry.

2. Use the Filter field (described above) to restrict access to subset of LDAP users. The string specified in the Filter will be ANDed with the search term. For instance, with `ldap.filter = ou=devs` the search query will be `(&(ou=devs)(uid=username))`. If the Filter is malformed, the LDAP search requests will not succeed and users will not be able to log in to Sourcegraph.

3. By default, users who can sign in with their LDAP credentials to your Sourcegraph will have read+write access. To restrict write access to only those users that you specified as admins, add the following flag in the config file and restart the server:

```
[serve]
auth.restrict-write-access = true
```
