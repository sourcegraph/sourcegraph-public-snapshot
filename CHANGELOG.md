# Changelog

All notable changes to Sourcegraph Server and Data Center are documented in this file.

## Unreleased changes (add your changes here!)

* Improve search timeout error messages

## 2.6.3

Bug fixes

* Symbol results are now only returned for searches that contain `type:symbol`

## 2.6.2

### Added

* More detailed logging to help diagnose errors with third-party authentication providers.
* Anchors (such as `#my-section`) in rendered Markdown files are now supported.
* Instrumentation section for admins. For each service we expose pprof, prometheus metrics and traces.

### Bug fixes

* Applies a 1s timeout to symbol search if invoked without specifying `type:` to not block plain text results. No change of behaviour if `type:symbol` is given explicitely.
* Only show line wrap toggle for code-view-rendered files.

## 2.6.1

### Bug fixes

* Fixes a bug where typing in the search query field would modify the expanded state of file search results.
* Fixes a bug where new logins via OpenID Connect would fail with the error `SSO error: ID Token verification failed`.

## 2.6.0

### Added

* Support for [Bitbucket Server](https://www.atlassian.com/software/bitbucket/server) as a codehost. Configure via the `bitbucketServer` site config field.
* Prometheus gauges for git clone queue depth (`src_gitserver_clone_queue`) and git ls-remote queue depth (`src_gitserver_lsremote_queue`).
* Slack notifications for saved searches may now be added for individual users (not just organizations).
* The new search filter `lang:` filters results by programming language (example: `foo lang:go` or `foo -lang:clojure`).
* Dynamic filters: filters generated from your search results to help refine your results.
* Search queries that consist only of `file:` now show files whose path matches the filters (instead of no results).
* Sourcegraph now automatically detects basic `$GOPATH` configurations found in `.envrc` files in the root of repositories.
* You can now configure the effective `$GOPATH`s of a repository by adding a `.sourcegraph/config.json` file to your repository with the contents `{"go": {"GOPATH": ["mygopath"]}}`.
* A new `"blacklistGoGet": ["mydomain.org,myseconddomain.com"]` offers users a quick escape hatch in the event that Sourcegraph is making unwanted `go get` or `git clone` requests to their website due to incorrectly-configured monorepos. Most users will never use this option.
* Search suggestions and results now include symbol results. The new filter `type:symbol` causes only symbol results to be shown.
  Additionally, symbols for a repository can be browsed in the new symbols sidebar.
* You can now expand and collapse all items on a search results page or selectively expand and collapse individual items.

### Configuration changes

* Reduced the `gitMaxConcurrentClones` site config option's default value from 100 to 5, to help prevent too many concurrent clones from causing issues on code hosts.
* Changes to some site configuration options are now automatically detected and no longer require a server restart. After hitting Save in the UI, you will be informed if a server restart is required, per usual.
* Saved search notifications are now only sent to the owner of a saved search (all of an organization's members for an organization-level saved search, or a single user for a user-level saved search). The `notifyUsers` and `notifyOrganizations` properties underneath `search.savedQueries` have been removed.
* Slack webhook URLs are now defined in user/organization JSON settings, not on the organization profile page. Previously defined organization Slack webhook URLs are automatically migrated to the organization's JSON settings.
* The "unlimited" value for `maxReposToSearch` is now `-1` instead of `0`, and `0` now means to use the default.
* `auth.provider` must be set (`builtin`, `openidconnect`, `saml`, `http-header`, etc.) to configure an authentication provider. Previously you could just set the detailed configuration property (`"auth.openIDConnect": {...}`, etc.) and it would implicitly enable that authentication provider.
* The `autoRepoAdd` site configuration property was removed. Site admins can add repositories via site configuration.

### Bug fixes

* Only cross reference index enabled repositories.
* Fixed an issue where search would return results with empty file contents for matches in submodules with indexing enabled. Searching over submodules is not supported yet, so these (empty) results have been removed.
* Fixed an issue where match highlighting would be incorrect on lines that contained multibyte characters.
* Fixed an issue where search suggestions would always link to master (and 404) even if the file only existed on a branch. Now suggestions always link to the revision that is being searched over.
* Fixed an issue where all file and repo links on the search results page (for all search results types) would always link to master branch, even if the results only existed in another branch. Now search results links always link to the revision that is being searched over.
* The first user to sign up for a (not-yet-initialized) server is made the site admin, even if they signed up using SSO. Previously if the first user signed up using SSO, they would not be a site admin and no site admin could be created.
* Fixed an issue where our code intelligence archive cache (in `lsp-proxy`) would not evict items from the disk. This would lead to disks running out of free space.

## 2.5.16, 2.5.17

* Version bump to keep Data Center and Server versions in sync.

## 2.5.15

### Bug fixes

* Fixed issue where Sourcegraph Data Center would incorrectly show "An update is available".
* Fixed Phabricator links to repos
* Searches over a single repo are now less likely to immediately time out the first time they are searched.
* Fixed a bug where `auth.provider == "http-header"` would incorrectly require builtin authentication / block site access when `auth.public == "false"`.

### Phabricator Integration Changes

We now display a "View on Phabricator" link rather than a "View on other code host" link if you are using Phabricator and hosting on Github or another code host with a UI. Commit links also will point to Phabricator.

### Improvements to SAML authentication

You may now optionally provide the SAML Identity Provider metadata XML file contents directly, with the `auth.saml` `identityProviderMetadata` site configuration property. (Previously, you needed to specify the URL where that XML file was available; that is still possible and is more common.) The new option is useful for organizations whose SAML metadata is not web-accessible or while testing SAML metadata configuration changes.

## 2.5.13

### Improvements to builtin authentication

When using `auth.provider == "builtin"`, two new important changes mean that a Sourcegraph server will be locked down and only accessible to users who are invited by an admin user (previously, we advised users to place their own auth proxy in front of Sourcegraph servers).

1.  When `auth.provider == "builtin"` Sourcegraph will now by default require an admin to invite users instead of allowing anyone who can visit the site to sign up. Set `auth.allowSignup == true` to retain the old behavior of allowing anyone who can access the site to signup.
2.  When `auth.provider == "builtin"`, Sourcegraph will now respects a new `auth.public` site configuration option (default value: `false`). When `auth.public == false`, Sourcegraph will not allow anyone to access the site unless they have an account and are signed in.

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
