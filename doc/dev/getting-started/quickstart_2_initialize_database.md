# Quickstart step 2: Initialize your database

## Without Docker

You need a fresh Postgres database and a database user that has full ownership of that database.

1. Create a database for the current Unix user

    ```
    # For Linux users, first access the postgres user shell
    sudo su - postgres
    # For Mac OS users
    sudo su - _postgres
    ```

    ```
    createdb
    ```

2. Create the Sourcegraph user and password

    ```
    createuser --superuser sourcegraph
    psql -c "ALTER USER sourcegraph WITH PASSWORD 'sourcegraph';"
    ```

3. Create the Sourcegraph database

    ```
    createdb --owner=sourcegraph --encoding=UTF8 --template=template0 sourcegraph
    ```

4. Configure database settings in your environment

    The Sourcegraph server reads PostgreSQL connection configuration from the [`PG*` environment variables](http://www.postgresql.org/docs/current/static/libpq-envars.html).

    The startup script sets default values that work with the setup described here, but if you are using different values you can overwrite them, for example, in your `~/.bashrc`:

    ```
    export PGPORT=5432
    export PGHOST=localhost
    export PGUSER=sourcegraph
    export PGPASSWORD=sourcegraph
    export PGDATABASE=sourcegraph
    export PGSSLMODE=disable
    ```

    You can also use a tool like [`envdir`][envdir] or [a `.dotenv` file][dotenv] to
    source these env vars on demand when you start the server.

    [envdir]: https://cr.yp.to/daemontools/envdir.html
    [dotenv]: https://github.com/joho/godotenv

## With Docker

The Sourcegraph server reads PostgreSQL connection configuration from the [`PG*` environment variables](http://www.postgresql.org/docs/current/static/libpq-envars.html).

The development server startup script as well as the docker compose file provide default settings, so it will work out of the box.

You can overwrite these settings e.g. via your `~/.bashrc`:

```
export PGUSER=sourcegraph
export PGPASSWORD=sourcegraph
export PGDATABASE=sourcegraph
```

You can also use the `PGDATA_DIR` environment variable to specify a local folder (instead of a volume) to store the database files. See the `dev/compose.yml` file for more details.

## More info

For more information about data storage, [read our full PostgreSQL page](../background-information/postgresql.md).

Migrations are applied automatically.

[< Previous](quickstart_1_install_dependencies.md) | [Next >](quickstart_3_start_docker.md)
