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
1. Changesets, code discussions, issues, etc. are all stored inside the Git
   repository itself (see the `refs/src/*` and `refs/changesets/*` Git refs).
1. User accounts are OAuth2 accounts and are stored on Sourcegraph.com, so you
   never need to worry about backing these up.
1. Your local machine's authentication info (i.e., your `src login`) is stored in
   `~/.src-auth`. This is your personal login information, not server-wide user
   data.
1. srclib build caches and language toolchains are in `~/.srclib`,
   but you should never need to backup these as they contain no user data.
