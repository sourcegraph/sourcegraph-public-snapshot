+++
title = "User access control"
+++

By default, Sourcegraph servers require authentication for all
operations, including for Web clients. All user accounts and
permissions are stored on the root server (Sourcegraph.com).

# Granting and revoking user access

The first user to register a Sourcegraph server becomes the server's
admin. Server admins can use the `src` CLI to grant and revoke other
users' permissions to access the server.

First, log in as the admin user on your Sourcegraph server:

	src --endpoint http://example.com login

Then use the `src access` commands to control access to your instance:

1. Grant read access to all specified users.
   Include the `--write` or `--admin` flags to grant write / admin access to
   all specified users:

		src access grant [--write] [--admin] <username1> <username2> ...


2. Revoke all access from all specified users:

		src access revoke <username1> <username2> ...

3. List all permitted users and their access levels on this server:

		src access list

See `src access --help` for more information.
