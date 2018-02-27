# Changelog

All notable changes to Sourcegraph Server and Data Center are documented in this file.

## Unreleased changes (add your changes here!)

* Fixed issue where Sourcegraph Data Center would incorrectly show "An update is available".

### Phabricator Integration Changes

We now display a "View on Phabricator" link rather than a "View on other code host" link if you are using Phabricator and hosting on Github or another code host with a UI. Commit links also will point to Phabricator.

## 2.5.13

### Improvements to builtin authentication

When using `auth.provider == "builtin"`, two new important changes mean that a Sourcegraph server will be locked down and only accessible to users who are invited by an admin user (previously, we advised users to place their own auth proxy in front of Sourcegraph servers).

1. When `auth.provider == "builtin"` Sourcegraph will now by default require an admin to invite users instead of allowing anyone who can visit the site to sign up. Set `auth.allowSignup == true` to retain the old behavior of allowing anyone who can access the site to signup.
2. When `auth.provider == "builtin"`, Sourcegraph will now respects a new `auth.public` site configuration option (default value: `false`). When `auth.public == false`, Sourcegraph will not allow anyone to access the site unless they have an account and are signed in.

## 2.4.3

### Added

* Code Intelligence support
* Custom links to code hosts with the `links:` config options in `repos.list`

### Changed

* Search by file path enabled by default

## 2.4.2

### Added

* Repository settings mirror/cloning diagnostics page

### Changed

* Repositories added from GitHub are no longer enabled by default. The site admin UI for enabling/disabling repositories is improved.

## 2.4.0

### Added

* Search files by name by including `type:path` in a search query
* Global alerts for configuration-needed and cloning-in-progress
* Better list interfaces for repositories, users, organizations, and threads
* Users can change their own password in settings
* Repository groups can now be specified in settings by site admins, organizations, and users. Then `repogroup:foo` in a search query will search over only those repositories specified for the `foo` repository group.

### Changed

* Server log messages are much quieter by default

## 2.3.11

### Added

* Added site admin updates page and update checking
* Added site admin telemetry page

### Changed

* Enhanced site admin panel
* Changed repo- and SSO-related site config property names to be consistent, updated documentation

## 2.3.10

### Added

* Online site configuration editing and reloading

### Changed

* Site admins are now configured in the site admin area instead of in the `adminUsernames` config key or `ADMIN_USERNAMES` env var. Users specified in those deprecated configs will be designated as site admins in the database upon server startup until those configs are removed in a future release.

## 2.3.9

### Fixed

* An issue that prevented creation and deletion of saved queries

## 2.3.8

### Added

* Built-in authentication: you can now sign up without an SSO provider.
* Faster default branch code search via indexing.

### Fixed

* Many performance improvements to search.
* Much log spam has been eliminated.

### Changed

* We optionally read `SOURCEGRAPH_CONFIG` from `$DATA_DIR/config.json`.
* SSH key required to clone repos from GitHub Enterprise when using a self-signed certificate.

## 0.3 - 13 December 2017

The last version without a CHANGELOG.
