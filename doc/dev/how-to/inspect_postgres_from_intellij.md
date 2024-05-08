# Inspecting the Postgres database with IntelliJ

This guide teaches you step by step how to use the Database plugin of IntelliJ to set up a connection
to your local sourcegraph database, and how to run SQL queries through it.

If you haven't done so yet, run `sg setup` and `sg start` to make sure your database is up and running.

Below we'll assume the default config for postgres. It is defined in [`sg.config.yaml`](/sg.config.yaml) (or your overrides),
and has `env` keys like `PGPASSWORD` and `PGPORT`.

## 1. Open the database plugin

It's usually on the right edge of your IDE.

![](https://storage.googleapis.com/sourcegraph-assets/docs/images/dev/how-to/intellij-database-database-plugin.png)

## 2. Create a new PostgreSQL connection

Click the `+` sign, then Data Source, and then find PostgreSQL.

![](https://storage.googleapis.com/sourcegraph-assets/docs/images/dev/how-to/intellij-database-create-data-source.png)

## 3. Credentials and Database

Enter the credentials as you see them in the [`sg.config.yaml`](/sg.config.yaml) (or your overrides) mentioned above.

Also enter `sourcegraph` as the database name (or the override that you're using).

![](https://storage.googleapis.com/sourcegraph-assets/docs/images/dev/how-to/intellij-database-config.png)

Below the connection settings you can click on "Test Connection" to make sure that your drivers are working.
IntelliJ may prompt you to download the drivers then.

![](https://storage.googleapis.com/sourcegraph-assets/docs/images/dev/how-to/intellij-database-test-connection.png)

## 4. Select the sourcegraph public schemas

In the `Schema` tab of the connection settings, select the `sourcegraph` database, and its `public` schema. Extend
the `sourcegraph` database if you don't see the `public` schema.

![](https://storage.googleapis.com/sourcegraph-assets/docs/images/dev/how-to/intellij-database-select-schema.png)

Click on Ok to wrap it up.

## Select a table

Open the datasource on the right, and drill down to a table. You can double-click a table to see its content.

![](https://storage.googleapis.com/sourcegraph-assets/docs/images/dev/how-to/intellij-database-select-table.png)

## Query

By click `+` and then Query Console or Shift+CMD+L (on Mac), you can get a regular SQL query editor.

![](https://storage.googleapis.com/sourcegraph-assets/docs/images/dev/how-to/intellij-database-query.png)
