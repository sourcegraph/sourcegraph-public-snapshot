+++
title = "User access control"
+++

By default, Sourcegraph servers require authentication for all
operations, including for Web clients. All user accounts and
permissions are stored locally on the Sourcegraph server.

The first user to register the Sourcegraph server becomes the server's
admin. Server admins can use the `src` CLI to grant and revoke other
users' permissions to access the server.

# Listing current users
Admins can list the existing users and their access levels by running:

	src user list

# Granting and revoking user access

Admins can use the following command to update a user's access level on the server:

	src user update --access=<level> <login>

`<login>` must be an existing username and `<level>` must be one of `read`,
`write` or `admin`.

# Deleting user accounts

Admins can delete a user account from their Sourcegraph server by running any of the following commands:

	src user delete -l <login>
	src user delete -e <email@domain.com>
	src user delete -i <uid>

where `<login>`, `<email@domain.com>` and `<uid>` are the login name, email address and UID, respectively, associated with the user account.
