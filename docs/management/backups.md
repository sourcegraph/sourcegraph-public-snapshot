+++
title = "Data Backups"
+++

# Creating a backup

To create a backup of all your Sourcegraph data and source code repositories,
simply archive your `.sourcegraph` directory:

```bash
$ cd /backups
$ tar -zcvf archive-12-11-2015.tar.gz ~/.sourcegraph/
```

This will create an entire backup of all repositories, changesets, code
discussions, issues, etc in a `/backups/archive-12-11-2015.tar.gz` tarball.

# Dataset Locations

Sourcegraph stores nearly all of it's data in a directory called _the
sourcegraph path_. This is configured by setting the `SGPATH` environment
variable prior to running `src serve`. On Mac OSX this defaults to `~/.sourcegraph`
and on Linux/Ubuntu this defaults to `/home/sourcegraph/.sourcegraph`.

1. Source code repositories are standard git repositories and are located at
   `$SGPATH/repos`.
1. Changesets, Code Discussions, Issues etc. are all stored inside the Git
   repository itself (see the `refs/src/*` and `refs/changesets/*` Git refs).
1. User accounts are OAuth2 accounts and are stored at sourcegraph.com, so you
   never need to worry about backing these up.
1. Your local machine's authentication info (i.e. your `src login`) is stored at
   `~/.src-auth`. This is your personal login information, not server-wide user
   data.
1. Srclib build caches and language toolchains are at `~/.srclib` directories,
   but you should never need to backup these as they contain no user data.
