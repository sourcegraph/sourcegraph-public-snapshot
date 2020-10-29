<!--
###################################### READ ME ###########################################
### This changelog should always be read on `main` branch. Its contents on version     ###
### branches do not necessarily reflect the changes that have gone into that branch.   ###
##########################################################################################
-->

# Changelog

All notable changes to `src-cli` are documented in this file.

## Unreleased changes

### Added

- EXPERIMENTAL: Templated campaign specs and file mounting. The campaign specs evaluated by `src campaign [preview|apply]` can now include template variables in `steps.run`, `steps.env`, and the new `steps.files` property, which allows users to create files inside the container in which the step is executed. The feature is marked as EXPERIMENTAL because it might change in the near future until we deem it non-experimental. See [#361](https://github.com/sourcegraph/src-cli/pull/361) for details.

### Changed

### Fixed

## 3.21.4

### Added

### Changed

### Fixed

- The `src lsif upload` command now respects `SRC_HEADER_` environment variables for multipart uploads. These environment variables are described [here](AUTH_PROXY.md). [#360](https://github.com/sourcegraph/src-cli/pull/360)

## 3.21.3

### Added

### Changed

- The progress bar in `src campaign [preview|apply]` now shows when executing a step failed in a repository by styling the line red and displaying standard error output. [#355](https://github.com/sourcegraph/src-cli/pull/355)
- The `src lsif upload` command will give more informative output when an unexpected payload (non-JSON or non-unmarshallable) is received from the target endpoint. [#359](https://github.com/sourcegraph/src-cli/pull/359)

### Fixed

## 3.21.2

### Fixed

- The cache dir used by `src campaign [preview|apply]` is now created before trying to create files in it, fixing a bug where the first run of the command could fail with a "file doesn't exist" error message. [#352](https://github.com/sourcegraph/src-cli/pull/352)

## 3.21.1

### Added

- The `published` flag in campaign specs may now be an array, which allows only specific changesets within a campaign to be published based on the repository name. [#294](https://github.com/sourcegraph/src-cli/pull/294)
- A new `src campaign new` command creates a campaign spec YAML file with common values prefilled to make it easier to create a new campaign. [#339](https://github.com/sourcegraph/src-cli/pull/339)
- New experimental command [`src validate`](https://docs.sourcegraph.com/admin/validation) validates a Sourcegraph installation. [#200](https://github.com/sourcegraph/src-cli/pull/200)

### Changed

- Error reporting by `src campaign [preview|apply]` has been improved and now includes more information about which step failed in which repository. [#325](https://github.com/sourcegraph/src-cli/pull/325)
- The default behaviour of `src campaigns [preview|apply]` has been changed to retain downloaded archives of repositories for better performance across re-runs of the command. To use the old behaviour and delete the archives use the `-clean-archives` flag. Repository archives are also not stored in the directory for temp data (see `-tmp` flag) anymore but in the cache directory, which can be configured with the `-cache` flag. To manually delete archives between runs, delete the `*.zip` files in the `-cache` directory (see `src campaigns -help` for its default location).
- `src campaign [preview|apply]` now check whether `git` and `docker` are available before trying to execute a campaign spec's steps. [#326](https://github.com/sourcegraph/src-cli/pull/326)
- The progress bar displayed by `src campaign [preview|apply]` has been extended by status bars that show which steps are currently being executed for each repository. [#338](https://github.com/sourcegraph/src-cli/pull/338)
- `src campaign [preview|apply]` now shows a warning when no changeset specs have been created.
- Requests sent to Sourcegraph by the `src campaign` commands now use gzip compression for the body when talking to Sourcegraph 3.21.0 and later. [#336](https://github.com/sourcegraph/src-cli/pull/336) and [#343](https://github.com/sourcegraph/src-cli/pull/343)

### Fixed

- Log files created by `src campaigns [preview|apply]` are deleted again after successful execution. This was a regression and is not new behaviour. If steps failed to execute or the `-keep-logs` flag is set the log files are not cleaned up.
- `src campaign [preview|apply]` now correctly handle the interrupt signal (emitted in a terminal with Ctrl-C) and abort execution of campaign steps, cleaning up running Docker containers.

## 3.21.0

### Added

- The new `src login` subcommand helps you authenticate `src` to access your Sourcegraph instance with your user credentials. [#317](https://github.com/sourcegraph/src-cli/pull/312)

## 3.20.0

### Added

- Campaigns specs now include an optional `author` property. (If not included, `src campaigns` generates default values for author name and email.) `src campaigns` now includes the name and email in all changeset specs that it generates.
- The campaigns temp directory can now be overwritten by using the `-tmp` flag with `src campaigns [apply|preview]` or by setting `SRC_CAMPAIGNS_TMP_DIR`. The directory is used to, for example, store log files and unzipped repository archives when executing campaign specs.

### Changed

- Repositories without a default branch are skipped when applying/previewing a campaign spec. [#312](https://github.com/sourcegraph/src-cli/pull/312)
- Log files produced when applying/previewing a campaign spec now have the `.log` file extension for easier opening. [#315](https://github.com/sourcegraph/src-cli/pull/315)
- Campaign specs that apply to unsupported repositories will no longer generate an error. Instead, those repositories will be skipped by default and the campaign will be applied to the supported repositories only. [#314](https://github.com/sourcegraph/src-cli/pull/314)

### Fixed

- Empty changeset specs without a diff are no longer uploaded as part of a campaign spec. [#313](https://github.com/sourcegraph/src-cli/pull/313)

## 3.19.0

### Changed

- The default branch for the `src-cli` project has been changed to `main`. [#262](https://github.com/sourcegraph/src-cli/pull/262)

### Fixed

- `src campaigns` output has been improved in the Windows console. [#274](https://github.com/sourcegraph/src-cli/pull/274)
- `src campaigns` will no longer generate warnings if `user.name` or `user.email` have not been set in the global Git configuration. [#277](https://github.com/sourcegraph/src-cli/pull/277)

## 3.18.0

### Added

- Add `-dump-requests` as an option to all commands that interact with the Sourcegraph API. [#266](https://github.com/sourcegraph/src-cli/pull/266)

### Changed

- Reworked the `src campaigns` family of commands to [align with the new spec-based workflow](https://docs.sourcegraph.com/user/campaigns). Most notably, campaigns are now created and applied using the new `src campaigns apply` command, and use [the new YAML spec format](https://docs.sourcegraph.com/user/campaigns#creating-a-campaign). [#260](https://github.com/sourcegraph/src-cli/pull/260)

## 3.17.1

### Added

- Add -upload-route to the lsif upload command.

## 3.17.0

### Added

- New command `src serve-git` which can serve local repositories for Sourcegraph to clone. This was previously in a command called `src-expose`. See [serving local repositories](https://docs.sourcegraph.com/admin/external_service/src_serve_git) in our documentation to find out more. [#12363](https://github.com/sourcegraph/sourcegraph/issues/12363)
- When used with Sourcegraph 3.18 or later, campaigns can now be created on GitLab. [#231](https://github.com/sourcegraph/src-cli/pull/231)

### Changed

### Fixed

## 3.16.1

### Fixed

- Fix inferred root for lsif upload command. [#248](https://github.com/sourcegraph/src-cli/pull/248)

### Removed

- Removed `clone-in-progress` flag. [#246](https://github.com/sourcegraph/src-cli/pull/246)

## 3.16

### Added

- Add `--no-progress` flag to the `lsif upload` command to disable verbose output in non-TTY environments.
- `SRC_HEADER_AUTHORIZATION="Bearer $(...)"` is now supported for authenticating `src` with custom auth proxies. See [auth proxy configuration docs](AUTH_PROXY.md) for more information. [#239](https://github.com/sourcegraph/src-cli/pull/239)
- Pull missing docker images automatically. [#191](https://github.com/sourcegraph/src-cli/pull/191)
- Searches that result in errors will now display any alerts returned by Sourcegraph, including suggestions for how the search could be corrected. [#221](https://github.com/sourcegraph/src-cli/pull/221)

### Changed

- The terminal UI has been replaced by the logger-based UI that was previously only visible in verbose-mode (`-v`). [#228](https://github.com/sourcegraph/src-cli/pull/228)
- Deprecated the `-endpoint` flag. Instead, use the `SRC_ENDPOINT` environment variable. [#235](https://github.com/sourcegraph/src-cli/pull/235)

### Fixed

### Removed
