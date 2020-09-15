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

### Changed

### Fixed

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
