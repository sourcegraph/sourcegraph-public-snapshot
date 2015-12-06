+++
title = "Persistence backends"
description = "Choose a scheme for persisting user data"
+++

Sourcegraph supports choosing between two storage backends. User data can either be stored on the filesystem using the default `fs` backend, or it can be configured to store data in a [PostgreSQL database](http://www.postgresql.org) for high scalability & availability.

# Using the PostgreSQL backend

With this configuration Sourcegraph will store all user data, Tracker threads, Changesets, etc. inside of PostgreSQL with the exception of:

- Repository data (stored in `$SGPATH/repos`, i.e. `~/.sourcegraph/repos`).
- Build Data (automatically rebuilds if lost).

## Enable PostgreSQL Backend

In order to use the `pgsql` backend you will need to use the `src serve --local-store=pgsql` CLI option or add the following to your configuration file:

```
[serve.Local store]
Store = pgsql
```

Note: If you installed Sourcegraph using one of the standard distribution or cloud provider packages,
Sourcegraph will run with configuration found at `/etc/sourcegraph/config.ini`.

## Providing authentication details

You will need to export some information about your PostgreSQL database as environment variables:

```
PGHOST=pgsql.example.com
PGPORT=5432
PGUSER=pguser
PGPASSWORD=pgpass
PGDATABASE=sourcegraph
PGSSLMODE=disable
```

If you are running Sourcegraph on Ubuntu Linux or one of the supported cloud providers, you can edit `/etc/sourcegraph/config.env` to export these variables in the server’s environment.

## Initialization and management

Prior to running Sourcegraph for the first time you will need to run `src pgsql create` which will initialize the database and tables.

The `src pgsql` command provides subcommands to drop, reset and truncate the database. See `src pgsql -h` for more information.

# graphstore

To use S3 as a storage backend for `srclib` Code Intelligence data, set the `graphstore.root` option:

```
--graphstore.root=s3://bucketname
```

Tell `src` how to authenticate to S3 by setting environment variables:

```
AWS_ACCESS_KEY_ID
AWS_SECRET_KEY
```

For safety, you may use IAM to create an access key that can only read and write to the created bucket.
