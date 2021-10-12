# Quickstart step 7: Start the server

## Configure `sg` to connect to databases

If you chose to run PostgreSQL and Redis **without Docker** they should already be running. You can jump the next section.

If you chose to run Redis and PostgreSQL **with Docker** to then we need to configure `sg` so it can connect to them.

1. In the `sourcegraph` folder, create a `sg.config.overwrite.yaml` file with the following contents (don't worry, `sg.config.overwrite.yaml` files are ignored by `git` and serve as a place for your local configuration):

    ```
    env:
        POSTGRES_HOST: localhost
        PGPASSWORD: sourcegraph
        PGUSER: sourcegraph
    ```

2. Start the databases:

    ```
    sg run redis-postgres
    ```

Keep this process running in a terminal window to keep the databases running. Follow the rest of the instructions in another terminal.

## Start the server

**If you are a Sourcegraph employee**: start the local development server for Sourcegraph Enterprise with the following command:

```
sg start
```

**If you are not a Sourcegraph employee and don't have access to [the `dev-private` repository](./quickstart_2_clone_repository.md)**: you want to start Sourcegraph OSS, do this:

```
sg start oss
```

This will continuously compile your code and live reload your locally running instance of Sourcegraph.

Navigate your browser to https://sourcegraph.test:3443 to see if everything worked.

If `sg` exits with errors or outputs errors, take a look at [Troubleshooting](../how-to/troubleshooting_local_development.md) or ask in the `#dev-experience` Slack channel.

## Running the server in different configurations

If you want to run the server in different configurations (with the monitoring stack, with code insights enabled, Sourcegraph OSS, ...), run the following:

```
sg start -help
```

That prints a list of possible configurations which you can start with `sg start`.

For example, you can start Sourcegraph in the mode it uses on Sourcegraph.com by running the following in one terminal window

```
sg start dotcom
```

and then, in another terminal window, start the monitoring stack:

```
sg start monitoring
```

<!-- omit in toc -->
[< Previous](quickstart_6_configure_https_reverse_proxy.md) | [Next >](quickstart_8_additional_resources.md)
