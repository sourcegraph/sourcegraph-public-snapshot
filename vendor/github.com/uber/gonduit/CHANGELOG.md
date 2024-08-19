# Changelog

All notable changes to this project will be documented in this file.

## [0.13.0] - 2021-06-11
### Added
- Support for `harbormaster.build.search` method.

## [0.12.0] - 2020-12-10
### Added
- Support for `project.search` method.
- Added missing fields to RepositoryURI struct

### Fixed
- `go fmt` executed

## [0.11.0] - 2020-12-02
### Added
- Support for `transaction.search` method.
- Support for `differential.diff.search` method.

## [0.10.1] - 2020-11-26
### Added
- Added missing edge type `revision.parent`.

## [0.10.0] - 2020-11-26
### Added
- Support for `edge.search` method.

## [0.9.1] - 2020-11-17
### Fixed
- `diffusion.repository.search` was not sending attachments correctly.

## [0.9.0] - 2020-11-10
### Added
- Support for `harbormaster.buildable.search` method.
- Support for `diffusion.repository.search` method.

### Changed
- `responses.SearchResponse` struct was renamed to `ResponseObject` and embeded
  `SearchCursor` was removed from it because not every object has a search
  cursor.

## [0.8.1] - 2020-11-09
### Changed
- Restored  test server 404 response to keep backwards compatibility with
  earlier versions.

## [0.8.0] - 2020-11-08
### Changed
- Refactored test server by removing `gin` as dependency.
- Rewritten tests to not depend on `gin`.

### Removed
- Gonduit does not depend on `gin` library anymore.

## [0.7.0] - 2020-11-05
### Added
- Support for `differential.revision.search` endpoint.

### Changed
- Migrated from Glide to go modules.

## [0.6.1] - 2019-11-26
- Dummy release

## [0.6.0] - 2019-11-26
### Changed
- `DifferentialRevision.Reviewers` now points not to map but
  `DifferentialRevisionReviewers` struct (which is map also).

### Fixed
- `differential.query` method does not fail anymore if revision has no
  reviewers.

## [0.5.0] - 2019-10-14
### Added
- Support for differential.getcommitmessage.

## [0.4.1] - 2019-10-08
### Added
- Support for differential.getcommitpaths.

## [0.4.0] - 2019-07-12
### Added
- Support for differential.querydiffs.
- Timeout field to code.ClientOptions.
- DifferentialStatusLegacy with int representations of statuses.
- Client interface to pass own http.Client instance.
- Introduced basic context.Context compatability.

### Changed
- Changed fields on entities.DifferentialRevision to match actual response
  returned from Phabricator. This is breaking change.

### Removed
- DifferentialStatus struct as it is not used anymore.

## [0.3.3] - 2019-06-07
### Added
- Added support for `maniphest.search` endpoint.

## [0.3.2] - 2019-01-31
### Fixed
- Return `ConduitError` with proper status code when Phabricator fails with
  HTML output and client can not parse JSON.

## [0.3.1] - 2019-01-08
### Added
- Added `Email` value to `entities.User` struct for response to `user.query`
  endpoint.

## [0.3.0] - 2018-11-19
### Changed
- Changed import paths from `etcinit` to `uber`.
- Updated vesions of dependencies.
