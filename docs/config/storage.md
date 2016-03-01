+++
title = "DB and storage"
description = "Manage the PostgreSQL database and srclib data storage backend used by Sourcegraph"
+++

Sourcegraph stores most data in a
[PostgreSQL database](http://www.postgresql.org). Git repositories,
build log files, srclib data, and uploaded user content (e.g., image
attachments in issues) are stored on the filesystem.

# Initializing PostgreSQL

After installing PostgreSQL, setup up a `sourcegraph` user and database:

```
# Linux only: sudo su - postgres
createuser --superuser sourcegraph
sql -c "ALTER USER sourcegraph WITH PASSWORD 'sourcegraph';"
createdb --owner=sourcegraph --encoding=UTF8 --template=template0 sourcegraph
```

Then update your `postgresql.conf` default timezone to UTC. Determine the location
of your `postgresql.conf` by running `psql -c 'show config_file;'`. Update the line beginning
with `timezone =` to the following:

```
timezone = 'UTC'
```

Finally, restart your database server (e.g. `sudo /etc/init.d/posgresql restart`).

# Configuring PostgreSQL

The Sourcegraph server reads PostgreSQL connection configuration from
the
[`PG*` environment variables](http://www.postgresql.org/docs/current/static/libpq-envars.html);
for example:

```
PGHOST=pgsql.example.com
PGPORT=5432
PGUSER=sourcegraph
PGPASSWORD=sourcegraph
PGDATABASE=sourcegraph
PGSSLMODE=disable
```

To test the environment's credentials, run `psql` (the PostgreSQL CLI
client) with the `PG*` environment variables set. If you see a
database prompt, then the environment's credentials are valid.

# Sourcegraph database management

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
