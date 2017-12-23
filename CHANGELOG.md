# Changelog

All notable changes to Sourcegraph Server's [docker
image](https://hub.docker.com/r/sourcegraph/server/tags/) will be documented
in this file.

For information on writing a good changelog see
http://keepachangelog.com/en/1.0.0/ and its changelog
https://github.com/olivierlacan/keep-a-changelog/blob/master/CHANGELOG.md

Before cutting a new release, please:

1. Build a new version of the image. Test it
2. Test that all the features mentioned work
3. Ensure the documentation is ready
4. Tag and push a new version. Update this document.

## 2.3.9

### Fixed

* An issue that prevented creation and deletion of saved queries

## 2.3.8

### Added

* Native authentication: you can now sign up without an SSO provider or Auth0.
* Faster default branch code search via indexing.

### Fixed

* Many performance improvements to search.
* Much log spam has been eliminated.

### Changed

* We optionally read `SOURCEGRAPH_CONFIG` from `$DATA_DIR/config.json`.
* SSH key required to clone repos from GitHub Enterprise when using a self-signed certificate.

## 0.3 - 13 December 2017

The last version without a CHANGELOG.
