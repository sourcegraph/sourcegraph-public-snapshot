# Sourcegraph Changelog

## [Unreleased]

### Added

- Added embeddings status in footer [#54575](https://github.com/sourcegraph/sourcegraph/pull/54575)

### Changed

- fix logging to use JetBrains api + other minor fixes [#54579](https://github.com/sourcegraph/sourcegraph/pull/54579)
- Enable editor recipe context menu items when working with Cody app only when Cody app is running [#54583](https://github.com/sourcegraph/sourcegraph/pull/54583)

### Deprecated

### Removed

### Fixed

### Security

## [3.0.3]

### Added

- Added recipes to editor context menu [#54430](https://github.com/sourcegraph/sourcegraph/pull/54430)
- Figure out default repository when no files are opened in the editor [#54476](https://github.com/sourcegraph/sourcegraph/pull/54476)
- Added `unstable-codegen` completions support [#54435](https://github.com/sourcegraph/sourcegraph/pull/54435)

### Changed

- Use smaller Cody logo in toolbar and editor context menu [#54481](https://github.com/sourcegraph/sourcegraph/pull/54481)
- Sourcegraph link sharing and opening file in browser actions are disabled when working with Cody app [#54473](https://github.com/sourcegraph/sourcegraph/pull/54473)
- Display more informative message when no context has been found [#54480](https://github.com/sourcegraph/sourcegraph/pull/54480)

### Deprecated

### Removed

### Fixed

- Preserve new lines in the human chat message [#54417](https://github.com/sourcegraph/sourcegraph/pull/54417)
- JetBrains: Handle response == null case when checking for embeddings [#54492](https://github.com/sourcegraph/sourcegraph/pull/54492)

### Security

## [3.0.2]

### Fixed

- Repositories with http/https remotes are now available for Cody [#54372](https://github.com/sourcegraph/sourcegraph/pull/54372)

## [3.0.1]

### Changed

- Sending message on Enter rather than Ctrl/Cmd+Enter [#54331](https://github.com/sourcegraph/sourcegraph/pull/54331)
- Updated name to Cody AI app [#54360](https://github.com/sourcegraph/sourcegraph/pull/54360)

### Removed

- Sourcegraph CLI's SRC_ENDPOINT and SRC_ACCESS_TOKEN env variables overrides for the local config got removed [#54369](https://github.com/sourcegraph/sourcegraph/pull/54369)

### Fixed

- telemetry is now being sent to both the current instance & dotcom (unless the current instance is dotcom, then just that) [#54347](https://github.com/sourcegraph/sourcegraph/pull/54347)
- Don't display doubled messages about the error when trying to load context [#54345](https://github.com/sourcegraph/sourcegraph/pull/54345)
- Now handling Null error messages in error logging properly [#54351](https://github.com/sourcegraph/sourcegraph/pull/54351)
- Made sidebar refresh work for non-internal builds [#54348](https://github.com/sourcegraph/sourcegraph/pull/54358)
- Don't display duplicated files in the "Read" section in the chat [#54363](https://github.com/sourcegraph/sourcegraph/pull/54363)
- Repositories without configured git remotes are now available for Cody [#54370](https://github.com/sourcegraph/sourcegraph/pull/54370)
- Repositories with http/https remotes are now available for Cody [#54372](https://github.com/sourcegraph/sourcegraph/pull/54372)

## [3.0.0]

### Added

- Background color and font of inline code blocks differs from regular text in message [#53761](https://github.com/sourcegraph/sourcegraph/pull/53761)
- Autofocus Cody chat prompt input [#53836](https://github.com/sourcegraph/sourcegraph/pull/53836)
- Basic integration with the local Cody App [#54061](https://github.com/sourcegraph/sourcegraph/pull/54061)
- Background color and font of inline code blocks differs from regular text in message [#53761](https://github.com/sourcegraph/sourcegraph/pull/53761)
- Autofocus Cody chat prompt input [#53836](https://github.com/sourcegraph/sourcegraph/pull/53836)
- Cody Agent [#53370](https://github.com/sourcegraph/sourcegraph/pull/53370)
- Chat message when access token is invalid or not
  configured [#53659](https://github.com/sourcegraph/sourcegraph/pull/53659)
- A separate setting for the (optional) dotcom access
  token. [pull/53018](https://github.com/sourcegraph/sourcegraph/pull/53018)
- Enabled "Explain selected code (detailed)"
  recipe [#53080](https://github.com/sourcegraph/sourcegraph/pull/53080)
- Enabled multiple recipes [#53299](https://github.com/sourcegraph/sourcegraph/pull/53299)
  - Explain selected code (high level)
  - Generate a unit test
  - Generate a docstring
  - Improve variable names
  - Smell code
  - Optimize code
- A separate setting for enabling/disabling Cody completions. [pull/53302](https://github.com/sourcegraph/sourcegraph/pull/53302)
- Debounce for inline Cody completions [pull/53447](https://github.com/sourcegraph/sourcegraph/pull/53447)
- Enabled "Translate to different language" recipe [#53393](https://github.com/sourcegraph/sourcegraph/pull/53393)
- Skip Cody completions if there is code in line suffix or in the middle of a word in prefix [#53476](https://github.com/sourcegraph/sourcegraph/pull/53476)
- Enabled "Summarize recent code changes" recipe [#53534](https://github.com/sourcegraph/sourcegraph/pull/53534)

### Changed

- Convert `\t` to spaces in leading whitespace for autocomplete suggestions (according to settings) [#53743](https://github.com/sourcegraph/sourcegraph/pull/53743)
- Disabled line highlighting in code blocks in chat [#53829](https://github.com/sourcegraph/sourcegraph/pull/53829)
- Parallelized completion API calls and reduced debounce down to 20ms [#53592](https://github.com/sourcegraph/sourcegraph/pull/53592)

### Fixed

- Fixed the y position at which autocomplete suggestions are rendered [#53677](https://github.com/sourcegraph/sourcegraph/pull/53677)
- Fixed rendered completions being cleared after disabling them in settings [#53758](https://github.com/sourcegraph/sourcegraph/pull/53758)
- Wrap long words in the chat message [#54244](https://github.com/sourcegraph/sourcegraph/pull/54244)
- Reset conversation button re-enables "Send"
  button [#53669](https://github.com/sourcegraph/sourcegraph/pull/53669)
- Fixed font on the chat ui [#53540](https://github.com/sourcegraph/sourcegraph/pull/53540)
- Fixed line breaks in the chat ui [#53543](https://github.com/sourcegraph/sourcegraph/pull/53543)
- Reset prompt input on message send [#53543](https://github.com/sourcegraph/sourcegraph/pull/53543)
- Fixed UI of the prompt input [#53548](https://github.com/sourcegraph/sourcegraph/pull/53548)
- Fixed zero-width spaces popping up in inline autocomplete [#53599](https://github.com/sourcegraph/sourcegraph/pull/53599)
- Reset conversation button re-enables "Send" button [#53669](https://github.com/sourcegraph/sourcegraph/pull/53669)
- Fixed displaying message about invalid access token on any 401 error from backend [#53674](https://github.com/sourcegraph/sourcegraph/pull/53674)

## [3.0.0-alpha.9]

### Added

- Background color and font of inline code blocks differs from regular text in message [#53761](https://github.com/sourcegraph/sourcegraph/pull/53761)
- Autofocus Cody chat prompt input [#53836](https://github.com/sourcegraph/sourcegraph/pull/53836)
- Basic integration with the local Cody App [#54061](https://github.com/sourcegraph/sourcegraph/pull/54061)
- Onboarding of the user when using local Cody App [#54298](https://github.com/sourcegraph/sourcegraph/pull/54298)

## [3.0.0-alpha.7]

### Added

- Background color and font of inline code blocks differs from regular text in message [#53761](https://github.com/sourcegraph/sourcegraph/pull/53761)
- Autofocus Cody chat prompt input [#53836](https://github.com/sourcegraph/sourcegraph/pull/53836)
- Cody Agent [#53370](https://github.com/sourcegraph/sourcegraph/pull/53370)

### Changed

- Convert `\t` to spaces in leading whitespace for autocomplete suggestions (according to settings) [#53743](https://github.com/sourcegraph/sourcegraph/pull/53743)
- Disabled line highlighting in code blocks in chat [#53829](https://github.com/sourcegraph/sourcegraph/pull/53829)

### Fixed

- Fixed the y position at which autocomplete suggestions are rendered [#53677](https://github.com/sourcegraph/sourcegraph/pull/53677)
- Fixed rendered completions being cleared after disabling them in settings [#53758](https://github.com/sourcegraph/sourcegraph/pull/53758)
- Wrap long words in the chat message [#54244](https://github.com/sourcegraph/sourcegraph/pull/54244)
- Reset conversation button re-enables "Send"
  button [#53669](https://github.com/sourcegraph/sourcegraph/pull/53669)

## [3.0.0-alpha.6]

### Added

- Chat message when access token is invalid or not
  configured [#53659](https://github.com/sourcegraph/sourcegraph/pull/53659)

## [3.0.0-alpha.5]

### Added

- A separate setting for the (optional) dotcom access
  token. [pull/53018](https://github.com/sourcegraph/sourcegraph/pull/53018)
- Enabled "Explain selected code (detailed)"
  recipe [#53080](https://github.com/sourcegraph/sourcegraph/pull/53080)
- Enabled multiple recipes [#53299](https://github.com/sourcegraph/sourcegraph/pull/53299)
  - Explain selected code (high level)
  - Generate a unit test
  - Generate a docstring
  - Improve variable names
  - Smell code
  - Optimize code
- A separate setting for enabling/disabling Cody completions. [pull/53302](https://github.com/sourcegraph/sourcegraph/pull/53302)
- Debounce for inline Cody completions [pull/53447](https://github.com/sourcegraph/sourcegraph/pull/53447)
- Enabled "Translate to different language" recipe [#53393](https://github.com/sourcegraph/sourcegraph/pull/53393)
- Skip Cody completions if there is code in line suffix or in the middle of a word in prefix [#53476](https://github.com/sourcegraph/sourcegraph/pull/53476)
- Enabled "Summarize recent code changes" recipe [#53534](https://github.com/sourcegraph/sourcegraph/pull/53534)

### Changed

- Parallelized completion API calls and reduced debounce down to 20ms [#53592](https://github.com/sourcegraph/sourcegraph/pull/53592)

### Fixed

- Fixed font on the chat ui [#53540](https://github.com/sourcegraph/sourcegraph/pull/53540)
- Fixed line breaks in the chat ui [#53543](https://github.com/sourcegraph/sourcegraph/pull/53543)
- Reset prompt input on message send [#53543](https://github.com/sourcegraph/sourcegraph/pull/53543)
- Fixed UI of the prompt input [#53548](https://github.com/sourcegraph/sourcegraph/pull/53548)
- Fixed zero-width spaces popping up in inline autocomplete [#53599](https://github.com/sourcegraph/sourcegraph/pull/53599)
- Reset conversation button re-enables "Send" button [#53669](https://github.com/sourcegraph/sourcegraph/pull/53669)
- Fixed displaying message about invalid access token on any 401 error from backend [#53674](https://github.com/sourcegraph/sourcegraph/pull/53674)

## [3.0.0-alpha.1]

### Added

- Alpha-quality Cody chat, not ready yet for internal dogfooding.
- Alpha-quality Cody code completions, not ready yet for internal dogfooding.

## [2.1.4]

### Added

- Add `extensionDetails` to `public_argument` on logger [#51321](https://github.com/sourcegraph/sourcegraph/pull/51321)

### Fixed

- Handle case when remote for local branch != sourcegraph remote [#52172](https://github.com/sourcegraph/sourcegraph/pull/52172)

## [2.1.3]

### Added

- Compatibility with IntelliJ 2023.1

### Fixed

- Fixed a backward-compatibility issue with Sourcegraph versions prior to 4.3 [#50080](https://github.com/sourcegraph/sourcegraph/issues/50080)

## [2.1.2]

### Added

- Compatibility with IntelliJ 2022.3

## [2.1.1]

### Added

- Now the name of the remote can contain slashes

### Fixed

- “Open in Browser” and “Copy Link” features now open the correct branch when it exists on the remote. [pull/44739](https://github.com/sourcegraph/sourcegraph/pull/44739)
- Fixed a bug where if the tracked branch had a different name from the local branch, the local branch name was used in the URL, incorrectly

## [2.1.0]

### Added

- Perforce support [pull/43807](https://github.com/sourcegraph/sourcegraph/pull/43807)
- Multi-repo project support [pull/43807](https://github.com/sourcegraph/sourcegraph/pull/43807)

### Changed

- Now using the VCS API bundled with the IDE rather than relying on the `git`
  command [pull/43807](https://github.com/sourcegraph/sourcegraph/pull/43807)

## [2.0.2]

### Added

- Added feature to specify auth headers [pull/42692](https://github.com/sourcegraph/sourcegraph/pull/42692)

### Removed

- Removed tracking parameters from all shareable URLs [pull/42022](https://github.com/sourcegraph/sourcegraph/pull/42022)

### Fixed

- Remove pointer cursor in the web view. [pull/41845](https://github.com/sourcegraph/sourcegraph/pull/41845)
- Updated “Learn more” URL to link the blog post in the update notification [pull/41846](https://github.com/sourcegraph/sourcegraph/pull/41846)
- Made the plugin compatible with versions 3.42.0 and below [pull/42105](https://github.com/sourcegraph/sourcegraph/pull/42105)

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
