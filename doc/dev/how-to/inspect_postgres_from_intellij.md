# Inspecting the Postgres database with IntelliJ

This guide teaches you step by step how to use the Database plugin of IntelliJ to set up a connection
to your local sourcegraph database, and how to run SQL queries through it.

If you haven't done so yet, run `sg setup` and `sg start` to make sure your database is up and running.

Below we'll assume the default config for postgres. It is defined in [`sg.config.yaml`](/sg.config.yaml) (or your overrides),
and has `env` keys like `PGPASSWORD` and `PGPORT`.

## 1. Open the database plugin

It's usually on the right edge of your IDE.

![](https://github.com/sourcegraph/sourcegraph/assets/1830132/064d1970-c9d5-4605-8242-1776b76a1ec2)

## 2. Create a new PostgreSQL connection

Click the `+` sign, then Data Source, and then find PostgreSQL.

![](https://github.com/sourcegraph/sourcegraph/assets/1830132/652af44e-f7d3-47f7-8782-3a17a319e203)

## 3. Credentials and Database

Enter the credentials as you see them in the [`sg.config.yaml`](/sg.config.yaml) (or your overrides) mentioned above.

Also enter `sourcegraph` as the database name (or the override that you're using).

![](https://github.com/sourcegraph/sourcegraph/assets/1830132/260acade-720b-478e-bb29-56c3095c9ebe)

Below the connection settings you can click on "Test Connection" to make sure that your drivers are working.
IntelliJ may prompt you to download the drivers then.

![](https://github.com/sourcegraph/sourcegraph/assets/1830132/5f684d02-05b6-46d9-9ab4-cae158cdf9cd)

## 4. Select the sourcegraph public schemas

In the `Schema` tab of the connection settings, select the `sourcegraph` database, and its `public` schema. Extend
the `sourcegraph` database if you don't see the `public` schema.

![](https://github.com/sourcegraph/sourcegraph/assets/1830132/b7a6cc51-d638-4f97-a5ef-8f5b20589d9b)

Click on Ok to wrap it up.

## Select a table

Open the datasource on the right, and drill down to a table. You can double-click a table to see its content.

![](https://github.com/sourcegraph/sourcegraph/assets/1830132/e1d2b6ad-0c5a-46ec-92c4-44f54011cb00)

## Query

By click `+` and then Query Console or Shift+CMD+L (on Mac), you can get a regular SQL query editor.

![](https://github.com/sourcegraph/sourcegraph/assets/1830132/1021255f-0677-43d3-91cb-ecd26b521164)
