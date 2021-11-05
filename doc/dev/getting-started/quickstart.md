# Quickstart 

>  NOTE: If you run into any troubles please consult the [deprecated quickstart instructions without `sg`](deprecated_quickstart.md) or reach out on Slack:
>  - [As an open source contributor](https://sourcegraph-community.slack.com/archives/C02BG0M0ZJ7)
>  - [As a Sourcegraph employee](https://sourcegraph.slack.com/archives/C01N83PS4TU)

## Install `sg`

At Sourcegraph we use [`sg`](https://github.com/sourcegraph/sourcegraph/tree/main/dev/sg), the Sourcegraph developer tool, to manage our local development environment.

To install `sg`, run the following in your terminal:

```
curl --proto '=https' --tlsv1.2 -sSLf https://install.sg.dev | sh
```

See the [`sg` documentation](../background-information/sg/index.md) for more information or ask in the `#dev-experience` Slack channel.

## Run `sg setup`

Open a terminal and run the following comming:

```
sg setup
```

Follow the printed instructions.

They will guide you through the installation of all the necessary dependencies to start the local development environment.

## Starting the databases

If you chose to run PostgreSQL and Redis **without Docker** they should already be running. You can jump the next section.

If you chose to run Redis and PostgreSQL **with Docker** to then we need to run them:

```
sg run redis-postgres
```

Keep this process running in a terminal window to keep the databases running. Follow the rest of the instructions in another terminal.

## Start the server

**If you are a Sourcegraph employee**: start the local development server for Sourcegraph Enterprise with the following command:

```
sg start
```

**If you are not a Sourcegraph employee and don't have access to [the `dev-private` repository](https://github.com/sourcegraph/dev-private)**: you want to start Sourcegraph OSS, do this:

```
sg start oss
```

This will continuously compile your code and live reload your locally running instance of Sourcegraph.

Navigate your browser to https://sourcegraph.test:3443 to see if everything worked.

Congratulations on making it to the end of the quickstart guide!

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

## Additional resources

Here are some additional resources to help you go further:

- [How-to guides](../how-to/index.md), particularly:
  - [Troubleshooting local development](../how-to/troubleshooting_local_development.md)
- [Background information](../background-information/index.md) for more context on various topics.
