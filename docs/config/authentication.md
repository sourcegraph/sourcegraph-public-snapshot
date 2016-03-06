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
