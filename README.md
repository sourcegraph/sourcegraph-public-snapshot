# Sourcegraph CLI [![Build Status](https://travis-ci.org/sourcegraph/src-cli.svg)](https://travis-ci.org/sourcegraph/src-cli) [![Build status](https://ci.appveyor.com/api/projects/status/fwa1bkd198hyim8a?svg=true)](https://ci.appveyor.com/project/sourcegraph/src-cli) [![Go Report Card](https://goreportcard.com/badge/sourcegraph/src-cli)](https://goreportcard.com/report/sourcegraph/src-cli)

**Quick links**: [Installation](#installation), [Setup](#setup) ([Authentication](#authentication)), [Usage](#usage)

The Sourcegraph `src` CLI provides access to [Sourcegraph](https://sourcegraph.com) via a command-line interface.

![image](https://user-images.githubusercontent.com/3173176/43567326-3db5f31c-95e6-11e8-9e74-4c04079c01b0.png)

It currently provides the ability to:

- **Execute search queries** from the command line and get nice colorized output back (or JSON, optionally).
- **Execute GraphQL queries** against a Sourcegraph instance, and get JSON results back (`src api`).
  - You can provide your API access token via an environment variable or file on disk.
  - You can easily convert a `src api` command into a curl command with `src api -get-curl`.
- **Manage repositories, users, and organizations** using the `src repos`, `src users`, and `src orgs` commands.

If there is something you'd like to see Sourcegraph be able to do from the CLI, let us know! :)

## Installation

For Sourcegraph 3.13 and newer, the preferred method of installation is to ask _your_ Sourcegraph instance for the latest compatible version. To do this, replace `https://sourcegraph.com` in the commands below with the address of your instance.

For Sourcegraph 3.12 and older, run the following commands verbatim (against sourcegraph.com) or install from one of the published [releases on GitHub](https://github.com/sourcegraph/src-cli/releases).

```
https://github.com/sourcegraph/src-cli/releases/download/{version}/{binary}
````

#### Mac OS

```bash
curl -L https://sourcegraph.com/.api/src-cli/src_darwin_amd64 -o /usr/local/bin/src
chmod +x /usr/local/bin/src
```

#### Linux

```bash
curl -L https://sourcegraph.com/.api/src-cli/src_linux_amd64 -o /usr/local/bin/src
chmod +x /usr/local/bin/src
```

#### Windows

**NOTE:** Windows support is still rough around the edges, but is available. If you encounter issues, please let us know by filing an issue :)

Run in PowerShell as administrator:

```powershell
New-Item -ItemType Directory 'C:\Program Files\Sourcegraph'
Invoke-WebRequest https://sourcegraph.com/.api/src-cli/src_windows_amd64.exe -OutFile 'C:\Program Files\Sourcegraph\src.exe'
[Environment]::SetEnvironmentVariable('Path', [Environment]::GetEnvironmentVariable('Path', [EnvironmentVariableTarget]::Machine) + ';C:\Program Files\Sourcegraph', [EnvironmentVariableTarget]::Machine)
$env:Path += ';C:\Program Files\Sourcegraph'
```

Or manually:

- Download the latest src_windows_amd64.exe: https://sourcegraph.com/.api/src-cli/src_windows_amd64.exe and rename to `src.exe`.
- Place the file under e.g. `C:\Program Files\Sourcegraph\src.exe`
- Add that directory to your system path to access it from any command prompt

#### Renaming `src` (optional)

If you have a naming conflict with the `src` command, such as a Bash alias, you can rename the static binary. For example, on Linux / Mac OS:

```sh
mv /usr/local/bin/src /usr/local/bin/sourcegraph-cli
```

You can then invoke it via `sourcegraph-cli`.

## Setup

If you want to use `src` with your own Sourcegraph instance set the `SRC_ENDPOINT` environment variable:

```sh
SRC_ENDPOINT=https://sourcegraph.example.com src search
```

### Authentication

Some Sourcegraph instances will be configured to require authentication. You can do so via the environment:

```sh
SRC_ENDPOINT=https://sourcegraph.example.com SRC_ACCESS_TOKEN="secret" src ...
```

Or via the configuration file (`~/src-config.json`):

```sh
	{"accessToken": "secret", "endpoint": "https://sourcegraph.example.com"}
```

See `src -h` for more information on specifying access tokens.

To acquire the access token, visit your Sourcegraph instance (or https://sourcegraph.com), click your username in the top right to open the user menu, select **Settings**, and then select **Access tokens** in the left hand menu.

## Usage

Consult `src -h` and `src api -h` for usage information.

## Development

If you want to develop the CLI, you can install it with `go get`:

```
go get -u github.com/sourcegraph/src-cli/cmd/src
```

## Releasing

1.  Find the latest version (either via the releases tab on GitHub or via git tags) to determine which version you are releasing.
2.  `VERSION=9.9.9 ./release.sh` (replace `9.9.9` with the version you are releasing)
3.  Travis will automatically perform the release. Once it has finished, **confirm that the curl commands fetch the latest version above**.
4.  Update the `MinimumVersion` constant in the [src-cli package](https://github.com/sourcegraph/sourcegraph/tree/master/internal/src-cli/consts.go).

### Patch releases

If a backwards-compatible change is made _after_ a backwards-incompatible one, the backwards-compatible one should be re-released to older instances that support it.

A Sourcegraph instance returns the highest patch version with the same major and minor version as `MinimumVersion` as defined in the instance. Patch versions are reserved solely for non-breaking changes and minor bug fixes. This allows us to dynamically release fixes for older versions of `src-cli` without having to update the instance.

To release a bug fix or a new feature that is backwards compatible with one of the previous two minor version of Sourcegraph, cherry-pick the changes into a patch branch and re-releases with a new patch version. 

For example, suppose we have the the recommended versions.

| Sourcegraph version | Recommended src-cli version |
| ------------------- | --------------------------- | 
| `3.100`             | `3.90.5`                    |
| `3.99`              | `3.85.7`                    |

If a new feature is added to a new `3.91.6` release of src-cli and this change requires only features available in Sourcegraph `3.99`, then this feature should also be present in a new `3.85.8` release of src-cli. Because a Sourcegraph instance will automatically select the highest patch version, all non-breaking changes should increment only the patch version. 

Note that if instead the recommended src-cli version for Sourcegraph `3.99` was `3.90.4` in the example above, there is no additional step required, and the new patch version of src-cli will be available to both Sourcegraph versions.
