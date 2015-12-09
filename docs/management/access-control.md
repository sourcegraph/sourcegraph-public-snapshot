+++
title = "User access control"
+++

By default, Sourcegraph servers require authentication for all
operations, including for Web clients. All user accounts and
permissions are stored locally on the Sourcegraph server.

The first user to register the Sourcegraph server becomes the server's
admin. Server admins can use the `src` CLI to grant and revoke other
users' permissions to access the server.

# Inviting new users
Admins can create invite links from their Sourcegraph homepage,
or via the CLI via this command:

	src user invite person1@domain.com person2@domain.com

This will generate invite links for each email address, which must be shared
with the user to allow them to create an account.

By default, new users have only `read` access on the server, which allows them
to read and clone code, and create issues. To grant additional privileges to
the new users, pass the `--write` or `--admin` flag in the command:

	src user invite --write person1@domain.com person2@domain.com

This will grant write access to the accounts created from the generated invite links.

## Anonymous readers and open logins
You can set your Sourcegraph server to be publicly readable, for example if your
code is open or fair source. To configure this, pass the flag `--auth.allow-anon-readers`
to `src serve`, or set this in your config file `/etc/sourcegraph/config.ini`:

	[serve.Auth]
	AllowAnonymousReaders = true

However, anonymous readers cannot post issues. To enable anyone to create or comment on
tracker threads on your Sourcegraph server, you must allow anyone to create a read-only
account on your server, by setting the flag `--auth.allow-all-logins` on the CLI, or by
adding to your config file:

	[serve.Auth]
	AllowAllLogins = true

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
