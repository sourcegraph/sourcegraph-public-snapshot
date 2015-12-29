+++
title = "Backups"
+++

# Creating a backup

To create a backup of all your Sourcegraph data and source code
repositories, archive your `.sourcegraph` directory:

```bash
$ tar -czvf archive-YYYY-MM-DD.tar.gz ~/.sourcegraph
```

This will create a backup of all repositories, changesets, code
discussions, issues, etc. in a `/backups/archive-YYYY-MM-DD.tar.gz`
tarball.

# Storage locations

Sourcegraph stores data in a directory given by the `SGPATH`
environment variable (or `$HOME/.sourcegraph` by default). If you
installed Sourcegraph on a Linux server, it may be running as user
`sourcegraph`, in which case the default directory would be
`/home/sourcegraph/.sourcegraph`.

1. Source code repositories are standard git repositories and are located at
   `$SGPATH/repos`.
1. Changesets, issues, user accounts, etc., are all also stored inside `$SGPATH`.
1. Your local machine's authentication info (i.e., your `src login`) is stored in
   `~/.src-auth`. This is your personal login information, not server-wide user
   data.
