+++
title = "DB and storage"
description = "Manage the PostgreSQL database and srclib data storage backend used by Sourcegraph"
+++

Sourcegraph stores most data in a
[PostgreSQL database](http://www.postgresql.org). Git repositories,
build log files, srclib data, and uploaded user content (e.g., image
attachments in issues) are stored on the filesystem.


# Configuring PostgreSQL

The Sourcegraph server reads PostgreSQL connection configuration from
the
[`PG*` environment variables](http://www.postgresql.org/docs/current/static/libpq-envars.html);
for example:

```
PGHOST=pgsql.example.com
PGPORT=5432
PGUSER=pguser
PGPASSWORD=pgpass
PGDATABASE=sourcegraph
PGSSLMODE=disable
```

To test the environment's credentials, run `psql` (the PostgreSQL CLI
client) with the `PG*` environment variables set. If you see a
database prompt, then the environment's credentials are valid.

If you installed Sourcegraph using the standard installation script on
Ubuntu Linux, these values live in `/etc/sourcegraph/config.env`.

## Initialization and management

Prior to running Sourcegraph for the first time you will need to run `src pgsql create` which will initialize the database and tables.

The `src pgsql` command provides subcommands to drop, reset and truncate the database. See `src pgsql -h` for more information.


# srclib code analysis data

By default, srclib data is stored on the local filesystem.

To use S3 as a storage backend for `srclib` code analysis data, set the `graphstore.root` option:

```
[serve]
graphstore.root = s3://bucketname
```

Tell `src` how to find and authenticate to your S3 bucket by setting environment variables:

```
AWS_REGION (e.g. "us-west-2")
AWS_ACCESS_KEY_ID
AWS_SECRET_KEY
```

For safety, you may use IAM to create an access key that can only read and write to the created bucket.
