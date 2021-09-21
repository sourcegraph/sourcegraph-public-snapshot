# Quickstart step 6: Start the server

## Install `sg`, the Sourcegraph developer tool

At Sourcegraph we use [`sg`](https://github.com/sourcegraph/sourcegraph/tree/main/dev/sg) to manage our local developer environment.

To install `sg`, do the following:

1. Navigate to the sourcegraph source code folder

    ```
    cd sourcegraph
    ```

2. Install `sg` and follow printed instructions

    ```
    ./dev/sg/install.sh
    ```

## For Sourcegraph employees: clone shared configuration

Before you can start the local environment, you'll need to clone [`sourcegraph/dev-private`](https://github.com/sourcegraph/dev-private) (which has convenient preconfigured settings and external services on an enterprise account) alongside the `sourcegraph/sourcegraph` repository, for example:

```
/dir
 |-- dev-private
 +-- sourcegraph
```

Note: Ensure that you have the latest changes from [`sourcegraph/dev-private`](https://github.com/sourcegraph/dev-private) as the secrets are updated from time to time.

```
[repo-updater] ERROR source.list-repos, error: 1 error occurred:
[repo-updater] 	* UnrecognizedClientException: The security token included in the request is invalid.
[repo-updater] 	status code: 400, request id: 1aa331d7-42ff-4d21-9465-2483409f86b7
```

## Start databases

If you do **not** use Docker, PostgreSQL and Redis should already be running. You can jump the next section.

If you **do use Docker for Redis & PostgreSQL** to then we need to configure `sg` so it can connect to them.

1. In the `sourcegraph` folder, create a `sg.config.overwrite.yaml` file with the following contents (don't worry, `sg.config.overwrite.yaml` files are ignored by `git` and serve as a place for your local configuration):

    ```
    env:
        POSTGRES_HOST: localhost
        PGPASSWORD: sourcegraph
        PGUSER: sourcegraph
    ```

2. Start the databases:

    ```
    sg start redis-postgres
    ```

## Start the server 

Start the local development server with the following command:

```
sg start
```

This will continuously compile your code and live reload your locally running instance of Sourcegraph.

Navigate your browser to https://sourcegraph.test:3443 to see if everything worked.

If `sg` exits with errors or outputs errors, have a look at [Troubleshooting](../how-to/troubleshooting_local_development.md) or ask in the `#dev-experience` Slack channel.

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
[< Previous](quickstart_5_configure_https_reverse_proxy.md) | [Next >](quickstart_7_additional_resources.md)
