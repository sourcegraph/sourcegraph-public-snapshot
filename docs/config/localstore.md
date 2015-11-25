+++
title = "Local Store Backends"
+++

Sourcegraph supports choosing between two local storage backends. User data can either be stored on the filesystem using the default `fs` backend, or it can be configured to store data in a [PostgreSQL database](http://www.postgresql.org) for high scalability & availability.

# Using the PostgreSQL backend

With this configuration, Sourcegraph will store all user data, Tracker threads, Changesets, etc. inside of PostgreSQL with the exception of:

- Repository data (stored in `$SGPATH/repos`, i.e. `~/.sourcegraph/repos`).
- Build Data (automatically rebuilds if lost).

## Enable PostgreSQL Backend

In order to use the `pgsql` backend you will need to use the `src serve --local-store=...` CLI option or modify your configuration file with the following:

```
[serve.Local store]
Store = pgsql
```

Note: If you installed Sourcegraph using one of the standard distribution or cloud provider packages,
Sourcegraph will run with configuration found at `/etc/sourcegraph/config.ini`.

## Providing authentication details

Additionally, you will need to export some information about your PostgreSQL database as environment variables:

```
PGHOST=pgsql.example.com
PGPORT=5432
PGUSER=pguser
PGPASSWORD=pgpass
PGDATABASE=sourcegraph
PGSSLMODE=disable
```

If you are running Sourcegraph on Ubuntu Linux or one of the supported cloud providers, you can edit the ``/etc/sourcegraph/config.env` file to export these variables in the server’s environment.

## Initialization

Prior to running for the first time, you will need to run `src pgsql create` which will initialize the database and tables. See the Management section below for more information.c

## Management

The `src pgsql` command allows you to manage the database simply. It uses the same environment variables mentioned above. The very first time you'll need to create the database using:

```
src pgsql create
```

More subcommands are provided to drop, reset and truncate the database: see `src pgsql -h` for more information.
