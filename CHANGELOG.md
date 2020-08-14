<!--
###################################### READ ME ###########################################
### This changelog should always be read on `master` branch. Its contents on version   ###
### branches do not necessarily reflect the changes that have gone into that branch.   ###
##########################################################################################
-->

# Changelog

All notable changes to `src-cli` are documented in this file.

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
