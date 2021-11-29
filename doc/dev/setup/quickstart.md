# Quickstart for setting up the local enviroment

This is the quickstart guide for [developing Sourcegraph](../index.md).

> NOTE: If you run into any troubles, you can alternatively consult the [deprecated quickstart instructions without `sg`](deprecated_quickstart.md) or reach out on Slack:
>
> - [As an open source contributor](https://sourcegraph-community.slack.com/archives/C02BG0M0ZJ7)
> - [As a Sourcegraph employee](https://sourcegraph.slack.com/archives/C01N83PS4TU)
>
> You can also get help on our [developer experience discussion forum](https://github.com/sourcegraph/sourcegraph/discussions/categories/developer-experience).

<span class="virtual-br"></span>

> NOTE: Looking for how to deploy or use Sourcegraph? See our [getting started](../../index.md#getting-started) options.

<span class="virtual-br"></span>

## Install `sg`

At Sourcegraph we use [`sg`, the Sourcegraph developer tool](../background-information/sg/index.md), to manage our local development environment.

To install `sg`, run the following in your terminal:

```sh
curl --proto '=https' --tlsv1.2 -sSLf https://install.sg.dev | sh
```

See the [`sg` documentation](../background-information/sg/index.md) for more information or ask in the `#dev-experience` Slack channel.

## Run `sg setup`

Open a terminal and run the following command:

```sh
sg setup
```

Follow the printed instructions.

They will guide you through the installation of all the necessary dependencies to start the local development environment.

## Starting the databases

If you chose to run PostgreSQL and Redis **without Docker** (recommended) they should already be running. You can jump the next section.

If you chose to run Redis and PostgreSQL **with Docker** to then we need to run them:

```sh
sg run redis-postgres
```

Keep this process running in a terminal window to keep the databases running. Follow the rest of the instructions in another terminal.

## Start the server

**If you are a Sourcegraph employee**: start the local development server for Sourcegraph Enterprise with the following command:

```sh
sg start
```

**If you are not a Sourcegraph employee and don't have access to [the `dev-private` repository](https://github.com/sourcegraph/dev-private)**: you want to start Sourcegraph OSS, do this:

```sh
sg start oss
```

This will continuously compile your code and live reload your locally running instance of Sourcegraph.

Navigate your browser to https://sourcegraph.test:3443 to see if everything worked.

Congratulations on making it to the end of the quickstart guide!

## Running the server in different configurations

If you want to run the server in different configurations (with the monitoring stack, with code insights enabled, Sourcegraph OSS, ...), run the following:

```sh
sg start -help
```

That prints a list of possible configurations which you can start with `sg start`.

For example, you can start Sourcegraph in the mode it uses on Sourcegraph.com by running the following in one terminal window

```sh
sg start dotcom
```

and then, in another terminal window, start the monitoring stack:

```sh
sg start monitoring
```

### Optional: asdf

[asdf](https://github.com/asdf-vm/asdf) is a CLI tool that manages runtime versions for a number of different languages and tools. It can be likened to a language-agnostic version of [nvm](https://github.com/nvm-sh/nvm) or [pyenv](https://github.com/pyenv/pyenv).

We use asdf in buildkite to lock the versions of the tools that we use on a per-commit basis.

#### Install asdf

##### asdf binary

See the [installation instructions on the official asdf documentation](https://asdf-vm.com/#/core-manage-asdf?id=install).

##### Plugins

sourcegraph/sourcegraph uses the following plugins:

- [Go](https://github.com/kennyp/asdf-golang)

```bash
asdf plugin add golang
```

- [NodeJS](https://github.com/asdf-vm/asdf-nodejs)

```bash
asdf plugin add nodejs

# Import the Node.js release team's OpenPGP keys to main keyring
bash ~/.asdf/plugins/nodejs/bin/import-release-team-keyring

# Have asdf read .nvmrc for auto-switching between node version
## Add the following to $HOME/.asdfrc:
legacy_version_file = yes
```

- [Yarn](https://github.com/twuni/asdf-yarn)

```bash
asdf plugin add yarn
```

#### Usage instructions

[asdf](https://github.com/asdf-vm/asdf) uses versions specified in [.tool-versions](https://github.com/sourcegraph/sourcegraph/blob/main/.tool-versions) whenever a command is run from one of `sourcegraph/sourcegraph`'s subdirectories.

You can install the all the versions specified in [.tool-versions](https://github.com/sourcegraph/sourcegraph/blob/main/.tool-versions) by running `asdf install`.


## Additional resources

Here are some additional resources to help you go further:

- [How-to guides](how-to/index.md), particularly:
  - [Troubleshooting local development](troubleshooting.md)
- [Background information](../background-information/index.md) for more context on various topics.
