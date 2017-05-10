# PostgreSQL

Sourcegraph stores most data in a
[PostgreSQL database](http://www.postgresql.org). Git repositories,
uploaded user content (e.g., image attachments in issues) are stored
on the filesystem.

# Initializing PostgreSQL

After installing PostgreSQL, set up up a `sourcegraph` user and database:

```
sudo su - postgres # this line only needed for Linux
createdb
createuser --superuser sourcegraph
psql -c "ALTER USER sourcegraph WITH PASSWORD 'sourcegraph';"
createdb --owner=sourcegraph --encoding=UTF8 --template=template0 sourcegraph
```

Then update your `postgresql.conf` default timezone to UTC. Determine the location
of your `postgresql.conf` by running `psql -c 'show config_file;'`. Update the line beginning
with `timezone =` to the following:

```
timezone = 'UTC'
```

Finally, restart your database server (mac: `brew services restart postgresql`)

# Configuring PostgreSQL

The Sourcegraph server reads PostgreSQL connection configuration from
the
[`PG*` environment variables](http://www.postgresql.org/docs/current/static/libpq-envars.html);
for example:

```
PGPORT=5432
PGUSER=sourcegraph
PGPASSWORD=sourcegraph
PGDATABASE=sourcegraph
PGSSLMODE=disable
```

To test the environment's credentials, run `psql` (the PostgreSQL CLI
client) with the `PG*` environment variables set. If you see a
database prompt, then the environment's credentials are valid.
