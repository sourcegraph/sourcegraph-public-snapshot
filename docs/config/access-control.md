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

First, log in:

```
src --endpoint http://example.com login
```

Then see `src access --help` for more information on granting and
revoking user access.
