# Sourcegraph Changelog

## [Unreleased]

### Added

### Changed

### Deprecated

### Removed

### Fixed

### Security

## [2.0.1]

- Improve Fedora Linux compatibility: Using `BrowserUtil.browse()` rather than `Desktop.getDesktop().browse()` to open
  links in the browser.

## [2.0.0]

- Added a new UI to search with Sourcegraph from inside the IDE. Open it with <kbd>Alt+S</kbd> (<kbd>⌥S</kbd> on Mac) by
  default.
- Added a settings UI to conveniently configure the plugin
- General revamp on the existing features
- Source code is now
  at [https://github.com/sourcegraph/sourcegraph/tree/main/client/jetbrains](https://github.com/sourcegraph/sourcegraph/tree/main/client/jetbrains)

## [1.2.4]

- Fixed an issue that prevent the latest version of the plugin to work with JetBrains 2022.1 products.

## [1.2.3]

- Upgrade JetBrains IntelliJ shell to 1.3.1 and modernize the build and release pipeline.

## [1.2.2] - Minor bug fixes

- It is now possible to configure the plugin per-repository using a `.idea/sourcegraph.xml` file. See the README for details.
- Special thanks: @oliviernotteghem for contributing the new features in this release!
- Fixed bugs where Open in Sourcegraph from the git menu does not work for repos with ssh url as their remote url

## [1.2.1] - Open Revision in Sourcegraph

- Added "Open In Sourcegraph" action to VCS History and Git Log to open a revision in the Sourcegraph diff view.
- Added "defaultBranch" configuration option that allows opening files in a specific branch on Sourcegraph.
- Added "remoteUrlReplacements" configuration option that allow users to replace specified values in the remote url with new strings.

## [1.2.0] - Copy link to file, search in repository, per-repository configuration, bug fixes & more

- The search menu entry is now no longer present when no text has been selected.
- When on a branch that does not exist remotely, `master` will now be used instead.
- Menu entries (Open file, etc.) are now under a Sourcegraph sub-menu.
- Added a "Copy link to file" action (alt+c / opt+c).
- Added a "Search in repository" action (alt+r / opt+r).
- It is now possible to configure the plugin per-repository using a `.idea/sourcegraph.xml` file. See the README for details.
- Special thanks: @oliviernotteghem for contributing the new features in this release!

## [1.1.2] - Minor bug fixes around searching.

- Fixed an error that occurred when trying to search with no selection.
- The git remote used for repository detection is now `sourcegraph` and then `origin`, instead of the previously poor choice of just the first git remote.

## [1.1.1] - Fixed search shortcut

- Updated the search URL to reflect a recent Sourcegraph.com change.

## [1.1.0] - Configurable Sourcegraph URL

- Added support for using the plugin with on-premises Sourcegraph instances.

## [1.0.0] - Initial Release

- Basic Open File & Search functionality.
