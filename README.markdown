# [Sourcegraph](https://sourcegraph.com) CLI [![Build Status](https://github.com/sourcegraph/src-cli/workflows/Go%20CI/badge.svg)](https://github.com/sourcegraph/src-cli/actions?query=workflow%3A%22Go+CI%22) [![Go Report Card](https://goreportcard.com/badge/sourcegraph/src-cli)](https://goreportcard.com/report/sourcegraph/src-cli)

<img src="https://user-images.githubusercontent.com/3173176/43567326-3db5f31c-95e6-11e8-9e74-4c04079c01b0.png" width=500 align=right>

`src` is a command line interface to Sourcegraph:

- **Search & get results in your terminal**
- **Search & get JSON** for programmatic consumption
- Make **GraphQL API requests** with auth easily & get JSON back fast
- Execute **[batch changes](https://docs.sourcegraph.com/batch_changes)**
- **Manage & administrate** repositories, users, and more
- **Easily convert src-CLI commands to equivalent curl commands**, just add --get-curl!

**Note:** Using Sourcegraph 3.12 or earlier? [See the older README](https://github.com/sourcegraph/src-cli/tree/3.11.2).

## Installation

Binary downloads are available on the [releases tab](https://github.com/sourcegraph/src-cli/releases), and through Sourcegraph.com. _If the latest version does not work for you,_ consider using the version compatible with your Sourcegraph instance instead.

### Installation: Mac OS

#### Latest version

```bash
curl -L https://sourcegraph.com/.api/src-cli/src_darwin_amd64 -o /usr/local/bin/src
chmod +x /usr/local/bin/src
```

or

```bash
brew install sourcegraph/src-cli/src-cli
```

#### Version compatible with your Sourcegraph instance

Replace `sourcegraph.example.com` with your Sourcegraph instance URL:

```bash
curl -L https://sourcegraph.example.com/.api/src-cli/src_darwin_amd64 -o /usr/local/bin/src
chmod +x /usr/local/bin/src
```

### Installation: Linux

#### Latest version

```bash
curl -L https://sourcegraph.com/.api/src-cli/src_linux_amd64 -o /usr/local/bin/src
chmod +x /usr/local/bin/src
```

#### Version compatible with your Sourcegraph instance

Replace `sourcegraph.example.com` with your Sourcegraph instance URL:

```bash
curl -L https://sourcegraph.example.com/.api/src-cli/src_linux_amd64 -o /usr/local/bin/src
chmod +x /usr/local/bin/src
```

### Installation: Windows

See [Sourcegraph CLI for Windows](WINDOWS.md).

## Log into your Sourcegraph instance

Run <code><strong>src login <i>SOURCEGRAPH-URL</i></strong></code> to authenticate `src` to access your Sourcegraph instance with your user credentials.

<blockquote>

**Examples**
`src login https://sourcegraph.example.com`
`src login https://sourcegraph.com`

</blockquote>

`src` consults the following environment variables:

- `SRC_ENDPOINT`: the URL to your Sourcegraph instance (such as `https://sourcegraph.example.com`)
- `SRC_ACCESS_TOKEN`: your Sourcegraph access token (on your Sourcegraph instance, click your user menu in the top right, then select **Settings > Access tokens** to create one)

For convenience, you can export these environment variables in your shell profile. To configure them as environment variables, add the following to your terminal profile file, typically accessible at `~/.bash_profile` (if using Bash) or `~/.zprofile` (if using Zsh):

```
export SRC_ACCESS_TOKEN=my-token
export SRC_ENDPOINT=https://sourcegraph.example.com 
```

Replace `my-token` and `https://sourcegraph.example.com` with the appropriate values for your account and instance.

You can also inline them in a single command with:

```sh
SRC_ENDPOINT=https://sourcegraph.example.com SRC_ACCESS_TOKEN=my-token src search 'foo'
```

Is your Sourcegraph instance behind a custom auth proxy? See [auth proxy configuration](./AUTH_PROXY.md) docs.

## Usage

`src` provides different subcommands to interact with different parts of Sourcegraph:

 - `src login` - authenticate to a Sourcegraph instance with your user credentials
 - `src search` - perform searches and get results in your terminal or as JSON
 - `src api` - run Sourcegraph GraphQL API requests
 - `src batch` - execute and manage [batch changes](https://docs.sourcegraph.com/batch_changes)
 - `src repos` - manage repositories
 - `src users` - manage users
 - `src orgs` - manages organization
 - `src config` - manage global, org, and user settings
 - `src extsvc` - manage external services (repository configuration)
 - `src extensions` - manage extensions
 - `src lsif` - manages LSIF data
 - `src serve-git` - serves your local git repositories over HTTP for Sourcegraph to pull
 - `src version` - check version and guaranteed-compatible version for your Sourcegraph instance

Run `src -h` and `src <subcommand> -h` for more detailed usage information.

#### Optional: Renaming `src`

If you have a naming conflict with the `src` command, such as a Bash alias, you can rename the static binary. For example, on Linux / Mac OS:

```sh
mv /usr/local/bin/src /usr/local/bin/src-cli
```

You can then invoke it via `src-cli`.
