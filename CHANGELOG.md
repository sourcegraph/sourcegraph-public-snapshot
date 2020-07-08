<!--
###################################### READ ME ###########################################
### This changelog should always be read on `master` branch. Its contents on version   ###
### branches do not necessarily reflect the changes that have gone into that branch.   ###
##########################################################################################
-->

# Changelog

All notable changes to `src-cli` are documented in this file.

## Unreleased

### Added

- `SRC_HEADER_AUTHORIZATION="Bearer $(...)"` is now supported for authenticating `src` with custom auth proxies. See [auth proxy configuration docs](AUTH_PROXY.md) for more information.
- Pull missing docker images automatically. [#191](https://github.com/sourcegraph/src-cli/pull/191)
- Searches that result in errors will now display any alerts returned by Sourcegraph, including suggestions for how the search could be corrected. [#221](https://github.com/sourcegraph/src-cli/pull/221)

### Changed

- The terminal UI has been replaced by the logger-based UI that was previously only visible in verbose-mode (`-v`). [#228](https://github.com/sourcegraph/src-cli/pull/228)
- Deprecated the `-endpoint` flag. Instead, use the `SRC_ENDPOINT` environment variable. [#235](https://github.com/sourcegraph/src-cli/pull/235)

### Fixed

### Removed
