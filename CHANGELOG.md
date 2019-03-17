<!-- **NOTE:** this changelog should always be read on `master` branch. Its contents on version
branches do not necessarily reflect the changes that have gone into that branch. -->

# Changelog

All notable changes to Sourcegraph are documented in this file.

## Unreleased

### Added

### Changed

- The default `github.repositoryQuery` of a [GitHub external service configuration](https://docs.sourcegraph.com/admin/external_service/github#configuration) has been changed to `["none"]`. Existing configurations that had this field unset will be migrated to have the previous default explicitly set (`["affiliated", "public"]`).

### Fixed

### Removed

## 3.2.0

### Added

- Sourcegraph can now automatically use the system's theme.
  To enable, open the user menu in the top right and make sure the theme dropdown is set to "System".
  This is currently supported on macOS Mojave with Safari Technology Preview 68 and later.
- The `github.exclude` setting was added to the [GitHub external service config](https://docs.sourcegraph.com/admin/external_service/github#configuration) to allow excluding repositories yielded by `github.repos` or `github.repositoryQuery` from being synced.

### Changed

- Symbols search is much faster now. After the initial indexing, you can expect code intelligence to be nearly instant no matter the size of your repository.
- Massively reduced the number of code host API requests Sourcegraph performs, which caused rate limiting issues such as slow search result loading to appear.
- The [`corsOrigin`](https://docs.sourcegraph.com/admin/config/site_config) site config property is no longer needed for integration with GitHub, GitLab, etc., via the [Sourcegraph browser extension](https://docs.sourcegraph.com/integration/browser_extension). Only the [Phabricator extension](https://github.com/sourcegraph/phabricator-extension) requires it.

### Fixed

- Fixed a bug where adding a search scope that adds a `repogroup` filter would cause invalid queries if `repogroup:sample` was already part of the query.
- An issue where errors during displaying search results would not be displayed.

### Removed

- The `"updateScheduler2"` experiment is now the default and it's no longer possible to configure.

## 3.1.2

### Added

- The `search.contextLines` setting was added to allow configuration of the number of lines of context to be displayed around search results.

### Changed

- Massively reduced the number of code host API requests Sourcegraph performs, which caused rate limiting issues such as slow search result loading to appear.
- Improved logging in various situations where Sourcegraph would potentially hit code host API rate limits.

### Fixed

- Fixed an issue where search results loading slowly would display a `Cannot read property "lastChild" of undefined` error.

## 3.1.1

### Added

- Query builder toggle (open/closed) state is now retained.

### Fixed

- Fixed an issue where single-term values entered into the "Exact match" field in the query builder were not getting wrapped in quotes.

## 3.1.0

### Added

- Added Docker-specific help text when running the Sourcegraph docker image in an environment with an sufficient open file descriptor limit.
- Added syntax highlighting for Kotlin and Dart.
- Added a management console environment variable to disable HTTPS, see [the docs](doc/admin/management_console.md#can-i-disable-https-on-the-management-console) for more information.
- Added `auth.disableUsernameChanges` to critical configuration to prevent users from changing their usernames.
- Site admins can query a user by email address or username from the GraphQL API.
- Added a search query builder to the main search page. Click "Use search query builder" to open the query builder, which is a form with separate inputs for commonly used search keywords.

### Changed

- File match search results now show full repository name if there are results from mirrors on different code hosts (e.g. github.com/sourcegraph/sourcegraph and gitlab.com/sourcegraph/sourcegraph)
- Search queries now use "smart case" by default. Searches are case insensitive unless you use uppercase letters. To explicitly set the case, you can still use the `case` field (e.g. `case:yes`, `case:no`). To explicitly set smart case, use `case:auto`.

### Fixed

- Fixed an issue where the management console would improperly regenerate the TLS cert/key unless `CUSTOM_TLS=true` was set. See the documentation for [how to use your own TLS certificate with the management console](doc/admin/management_console.md#how-can-i-use-my-own-tls-certificates-with-the-management-console).

## 3.0.1

### Added

- Symbol search now supports Elixir, Haskell, Kotlin, Scala, and Swift

### Changed

- Significantly optimized how file search suggestions are provided when using indexed search (cluster deployments).
- Both the `sourcegraph/server` image and the [Kubernetes deployment](https://github.com/sourcegraph/deploy-sourcegraph) manifests ship with Postgres `11.1`. For maximum compatibility, however, the minimum supported version remains `9.6`. The upgrade procedure is mostly automated for existing deployments. Please refer to [this page](https://docs.sourcegraph.com/admin/postgres) for detailed instructions.

### Removed

- The deprecated `auth.disableAccessTokens` site config property was removed. Use `auth.accessTokens` instead.
- The `disableBrowserExtension` site config property was removed. [Configure nginx](https://docs.sourcegraph.com/admin/nginx) instead to block clients (if needed).

## 3.0.0

See the changelog entries for 3.0.0 beta releases and our [3.0](doc/admin/migration/3_0.md) upgrade guide if you are upgrading from 2.x.

## 3.0.0-beta.4

### Added

- Basic code intelligence for the top 10 programming languages works out of the box without any configuration. [TypeScript/JavaScript](https://sourcegraph.com/extensions/sourcegraph/typescript), [Python](https://sourcegraph.com/extensions/sourcegraph/python), [Java](https://sourcegraph.com/extensions/sourcegraph/java), [Go](https://sourcegraph.com/extensions/sourcegraph/go), [C/C++](https://sourcegraph.com/extensions/sourcegraph/cpp), [Ruby](https://sourcegraph.com/extensions/sourcegraph/ruby), [PHP](https://sourcegraph.com/extensions/sourcegraph/php), [C#](https://sourcegraph.com/extensions/sourcegraph/csharp), [Shell](https://sourcegraph.com/extensions/sourcegraph/shell), and [Scala](https://sourcegraph.com/extensions/sourcegraph/scala) are enabled by default, and you can find more in the [extension registry](https://sourcegraph.com/extensions?query=category%3A"Programming+languages").

## 3.0.0-beta.3

- Fixed an issue where the site admin is redirected to the start page instead of being redirected to the repositories overview page after deleting a repo.

## 3.0.0-beta

### Added

- Repositories can now be queried by a git clone URL through the GraphQL API.
- A new Explore area is linked from the top navigation bar (when the `localStorage.explore=true;location.reload()` feature flag is enabled).
- Authentication via GitHub is now supported. To enable, add an item to the `auth.providers` list with `type: "github"`. By default, GitHub identities must be linked to an existing Sourcegraph user account. To enable new account creation via GitHub, use the `allowSignup` option in the `GitHubConnection` config.
- Authentication via GitLab is now supported. To enable, add an item to the `auth.providers` list with `type: "gitlab"`.
- GitHub repository permissions are supported if authentication via GitHub is enabled. See the
  documentation for the `authorization` field of the `GitHubConnection` configuration.
- The repository settings mirroring page now shows when a repo is next scheduled for an update (requires experiment `"updateScheduler2": "enabled"`).
- Configured repositories are periodically scheduled for updates using a new algorithm. You can disable the new algorithm with the following site configuration: `"experimentalFeatures": { "updateScheduler2": "disabled" }`. If you do so, please file a public issue to describe why you needed to disable it.
- When using HTTP header authentication, [`stripUsernameHeaderPrefix`](https://docs.sourcegraph.com/admin/auth/#username-header-prefixes) field lets an admin specify a prefix to strip from the HTTP auth header when converting the header value to a username.
- Sourcegraph extensions whose package.json contains `"wip": true` are considered [work-in-progress extensions](https://docs.sourcegraph.com/extensions/authoring/publishing#wip-extensions) and are indicated as such to avoid users accidentally using them.
- Information about user survey submissions and a chart showing weekly active users is now displayed on the site admin Overview page.
- A new GraphQL API field `UserEmail.isPrimary` was added that indicates whether an email is the user's primary email.
- The filters bar in the search results page can now display filters from extensions.
- Extensions' `activate` functions now receive a `sourcegraph.ExtensionContext` parameter (i.e., `export function activate(ctx: sourcegraph.ExtensionContext): void { ... }`) to support deactivation and running multiple extensions in the same process.
- Users can now request an Enterprise trial license from the site init page.
- When searching, a filter button `case:yes` will now appear when relevant. This helps discovery and makes it easier to use our case-sensitive search syntax.
- Extensions can now report progress in the UI through the `withProgress()` extension API.
- When calling `editor.setDecorations()`, extensions must now provide an instance of `TextDocumentDecorationType` as first argument. This helps gracefully displaying decorations from several extensions.

### Changed

- The Postgres database backing Sourcegraph has been upgraded from 9.4 to 11.1. Existing Sourcegraph users must conduct an [upgrade procedure](https://docs.sourcegraph.com/admin/postgres_upgrade)
- Code host configuration has moved out of the site config JSON into the "External services" area of the site admin web UI. Sourcegraph instances will automatically perform a one time migration of existing data in the site config JSON. After the migration these keys can be safely deleted from the site config JSON: `awsCodeCommit`, `bitbucketServer`, `github`, `gitlab`, `gitolite`, and `phabricator`.
- Site and user usage statistics are now visible to all users. Previously only site admins (and users, for their own usage statistics) could view this information. The information consists of aggregate counts of actions such as searches, page views, etc.
- The Git blame information shown at the end of a line is now provided by the [Git extras extension](https://sourcegraph.com/extensions/sourcegraph/git-extras). You must add that extension to continue using this feature.
- The `appURL` site configuration option was renamed to `externalURL`.
- The repository and directory pages now show all entries together instead of showing files and (sub)directories separately.
- Extensions no longer can specify titles (in the `title` property in the `package.json` extension manifest). Their extension ID (such as `alice/myextension`) is used.

### Fixed

- Fixed an issue where the site admin License page showed a count of current users, rather than the max number of users over the life of the license.
- Fixed number formatting issues on site admin Overview and Survey Response pages.
- Fixed resolving of git clone URLs with `git+` prefix through the GraphQL API
- Fixed an issue where the graphql Repositories endpoint would order by a field which was not indexed. Times on Sourcegraph.com went from 10s to 200ms.
- Fixed an issue where whitespace was not handled properly in environment variable lists (`SYMBOLS_URL`, `SEARCHER_URL`).
- Fixed an issue where clicking inside the repository popover or clicking "Show more" would dismiss the popover.

### Removed

- The `siteID` site configuration option was removed because it is no longer needed. If you previously specified this in site configuration, a new, random site ID will be generated upon server startup. You can safely remove the existing `siteID` value from your site configuration after upgrading.
- The **Info** panel was removed. The information it presented can be viewed in the hover.
- The top-level `repos.list` site configuration was removed in favour of each code-host's equivalent options,
  now configured via the new _External Services UI_ available at `/site-admin/external-services`. Equivalent options in code hosts configuration:
  - Github via [`github.repos`](https://docs.sourcegraph.com/admin/site_config/all#repos-array)
  - Gitlab via [`gitlab.projectQuery`](https://docs.sourcegraph.com/admin/site_config/all#projectquery-array)
  - Phabricator via [`phabricator.repos`](https://docs.sourcegraph.com/admin/site_config/all#phabricator-array)
  - [Other external services](https://docs.sourcegraph.com/admin/repo/add_from_other_external_services)
- Removed the `httpStrictTransportSecurity` site configuration option. Use [nginx configuration](https://docs.sourcegraph.com/admin/nginx) for this instead.
- Removed the `tls.letsencrypt` site configuration option. Use [nginx configuration](https://docs.sourcegraph.com/admin/nginx) for this instead.
- Removed the `tls.cert` and `tls.key` site configuration options. Use [nginx configuration](https://docs.sourcegraph.com/admin/nginx) for this instead.
- Removed the `httpToHttpsRedirect` and `experimentalFeatures.canonicalURLRedireect` site configuration options. Use [nginx configuration](https://docs.sourcegraph.com/admin/nginx) for these instead.
- Sourcegraph no longer requires access to `/var/run/docker.sock`.

## 2.13.6

### Added

- The `/-/editor` endpoint now accepts a `hostname_patterns` URL parameter, which specifies a JSON
  object mapping from hostname to repository name pattern. This serves as a hint to Sourcegraph when
  resolving git clone URLs to repository names. The name pattern is the same style as is used in
  code host configurations. The default value is `{hostname}/{path}`.

## 2.13.5

### Fixed

- Fixed another issue where Sourcegraph would try to fetch more than the allowed number of repositories from AWS CodeCommit.

## 2.13.4

### Changed

- The default for `experimentalFeatures.canonicalURLRedirect` in site config was changed back to `disabled` (to avoid [#807](https://github.com/sourcegraph/sourcegraph/issues/807)).

## 2.13.3

### Fixed

- Fixed an issue that would cause the frontend health check endpoint `/healthz` to not respond. This only impacts Kubernetes deployments.
- Fixed a CORS policy issue that caused requests to be rejected when they come from origins not in our [manifest.json](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/client/browser/src/extension/manifest.spec.json#L72) (i.e. requested via optional permissions by the user).
- Fixed an issue that prevented `repositoryQuery` from working correctly on GitHub enterprise instances.

## 2.13.2

### Fixed

- Fixed an issue where Sourcegraph would try to fetch more than the allowed number of repositories from AWS CodeCommit.

## 2.13.1

### Changed

- The timeout when running `git ls-remote` to determine if a remote url is cloneable has been increased from 5s to 30s.
- Git commands now use [version 2 of the Git wire protocol](https://opensource.googleblog.com/2018/05/introducing-git-protocol-version-2.html), which should speed up certain operations (e.g. `git ls-remote`, `git fetch`) when communicating with a v2 enabled server.

## 2.13.0

### Added

- A new site config option `search.index.enabled` allows toggling on indexed search.
- Search now uses [Sourcegraph extensions](https://docs.sourcegraph.com/extensions) that register `queryTransformer`s.
- GitLab repository permissions are now supported. To enable this, you will need to set the `authz`
  field in the `GitLabConnection` configuration object and ensure that the access token set in the
  `token` field has both `sudo` and `api` scope.

### Changed

- When the `DEPLOY_TYPE` environment variable is incorrectly specified, Sourcegraph now shuts down and logs an error message.
- The `experimentalFeatures.canonicalURLRedirect` site config property now defaults to `enabled`. Set it to `disabled` to disable redirection to the `appURL` from other hosts.
- Updating `maxReposToSearch` site config no longer requires a server restart to take effect.
- The update check page no longer shows an error if you are using an insiders build. Insiders builds will now notify site administrators that updates are available 40 days after the release date of the installed build.
- The `github.repositoryQuery` site config property now accepts arbitrary GitHub repository searches.

### Fixed

- The user account sidebar "Password" link (to the change-password form) is now shown correctly.
- Fixed an issue where GitHub rate limits were underutilized if the remaining
  rate limit dropped below 150.
- Fixed an issue where GraphQL field `elapsedMilliseconds` returned invalid value on empty searches
- Editor extensions now properly search the selection as a literal string, instead of incorrectly using regexp.
- Fixed a bug where editing and deleting global saved searches was not possible.
- In index search, if the search regex produces multiline matches, search results are still processed per line and highlighted correctly.
- Go-To-GitHub and Go-To-GitLab buttons now link to the right branch, line and commit range.
- Go-to-GitHub button links to default branch when no rev is given.
- The close button in the panel header stays located on the top.
- The Phabricator icon is now displayed correctly.
- The view mode button in the BlobPage now shows the correct view mode to switch to.

### Removed

- The experimental feature flag to disable the new repo update scheduler has been removed.
- The `experimentalFeatures.configVars` feature flag was removed.
- The `experimentalFeatures.multipleAuthProviders` feature flag was removed because the feature is now always enabled.
- The following deprecated auth provider configuration properties were removed: `auth.provider`, `auth.saml`, `auth.openIDConnect`, `auth.userIdentityHTTPHeader`, and `auth.allowSignup`. Use `auth.providers` for all auth provider configuration. (If you were still using the deprecated properties and had no `auth.providers` set, all access to your instance will be rejected until you manually set `auth.providers`.)
- The deprecated site configuration properties `search.scopes` and `settings` were removed. Define search scopes and settings in global settings in the site admin area instead of in site configuration.
- The `pendingContents` property has been removed from our GraphQL schema.
- The **Explore** page was replaced with a **Repositories** search link in the top navigation bar.

## 2.12.3

### Fixed

- Fixed an error that prevented users without emails from submitting satisfaction surveys.

## 2.12.2

### Fixed

- Fixed an issue where private GitHub Enterprise repositories were not fetched.

## 2.12.1

### Fixed

- We use GitHub's REST API to query affliated repositories. This API has wider support on older GitHub enterprise versions.
- Fixed an issue that prevented users without email addresses from signing in (https://github.com/sourcegraph/sourcegraph/issues/426).

## 2.12.0

### Changed

- Reduced the size of in-memory data structured used for storing search results. This should reduce the backend memory usage of large result sets.
- Code intelligence is now provided by [Sourcegraph extensions](https://docs.sourcegraph.com/extensions). The extension for each language in the site configuration `langservers` property is automatically enabled.
- Support for multiple authentication providers is now enabled by default. To disable it, set the `experimentalFeatures.multipleAuthProviders` site config option to `"disabled"`. This only applies to Sourcegraph Enterprise.
- When using the `http-header` auth provider, valid auth cookies (from other auth providers that are currently configured or were previously configured) are now respected and will be used for authentication. These auth cookies also take precedence over the `http-header` auth. Previously, the `http-header` auth took precedence.
- Bitbucket Server username configuration is now used to clone repositories if the Bitbucket Server API does not set a username.
- Code discussions: On Sourcegraph.com / when `discussions.abuseProtection` is enabled in the site config, rate limits to thread creation, comment creation, and @mentions are now applied.

### Added

- Search syntax for filtering archived repositories. `archived:no` will exclude archived repositories from search results, `archived:only` will search over archived repositories only. This applies for GitHub and GitLab repositories.
- A Bitbucket Server option to exclude personal repositories in the event that you decide to give an admin-level Bitbucket access token to Sourcegraph and do not want to create a bot account. See https://docs.sourcegraph.com/integration/bitbucket_server#excluding-personal-repositories for more information.
- Site admins can now see when users of their Sourcegraph instance last used it via a code host integration (e.g. Sourcegraph browser extensions). Visit the site admin Analytics page (e.g. https://sourcegraph.example.com/site-admin/analytics) to view this information.
- A new site config option `extensions.allowRemoteExtensions` lets you explicitly specify the remote extensions (from, e.g., Sourcegraph.com) that are allowed.
- Pings now include a total count of user accounts.

### Fixed

- Files with the gitattribute `export-ignore` are no longer excluded for language analysis and search.
- "Discard changes?" confirmation popup doesn't pop up every single time you try to navigate to a new page after editting something in the site settings page anymore.
- Fixed an issue where Git repository URLs would sometimes be logged, potentially containing e.g. basic auth tokens.
- Fixed date formatting on the site admin Analytics page.
- File names of binary and large files are included in search results.

### Removed

- The deprecated environment variables `SRC_SESSION_STORE_REDIS` and `REDIS_MASTER_ENDPOINT` are no longer used to configure alternative redis endpoints. For more information, see "[Using external databases with Sourcegraph](https://docs.sourcegraph.com/admin/external_database)".

## 2.11.1

### Added

- A new site config option `git.cloneURLToRepositoryName` specifies manual mapping from Git clone URLs to Sourcegraph repository names. This is useful, for example, for Git submodules that have local clone URLs.

### Fixed

- Slack notifications for saved searches have been fixed.

## 2.11.0

### Changed

### Added

- Support for ACME "tls-alpn-01" challenges to obtain LetsEncrypt certificates. Previously Sourcegraph only supported ACME "http-01" challenges which required port 80 to be accessible.
- gitserver periodically removes stale lock files that git can leave behind.
- Commits with empty trees no longer return 404.
- Clients (browser/editor extensions) can now query configuration details from the `ClientConfiguration` GraphQL API.
- The config field `auth.accessTokens.allow` allows or restricts use of access tokens. It can be set to one of three values: "all-users-create" (the default), "none" (all access tokens are disabled), and "site-admin-create" (access tokens are enabled, but only site admins can create new access tokens). The field `auth.disableAccessTokens` is now deprecated in favor of this new field.
- A webhook endpoint now exists to trigger repository updates. For example, `curl -XPOST -H 'Authorization: token $ACCESS_TOKEN' $SOURCEGRAPH_ORIGIN/.api/repos/$REPO_URI/-/refresh`.
- Git submodules entries in the file tree now link to the submodule repository.

### Fixed

- An issue / edge case where the Code Intelligence management admin page would incorrectly show language servers as `Running` when they had been removed from Docker.
- Log level is respected in lsp-proxy logs.
- Fixed an error where text searches could be routed to a faulty search worker.
- Gitolite integration should correctly detect names which Gitolite would consider to be patterns, and not treat them as repositories.
- repo-updater backs off fetches on a repo that's failing to fetch.
- Attempts to add a repo with an empty string for the name are checked for and ignored.
- Fixed an issue where non-site-admin authenticated users could modify global settings (not site configuration), other organizations' settings, and other users' settings.
- Search results are rendered more eagerly, resulting in fewer blank file previews
- An issue where automatic code intelligence would fail to connect to the underlying `lsp` network, leading to `dial tcp: lookup lang on 0.0.0.0:53: no such host` errors.
- More useful error messages from lsp-proxy when a language server can't get a requested revision of a repository.
- Creation of a new user with the same name as an existing organization (and vice versa) is prevented.

### Removed

## 2.10.5

### Fixed

- Slack notifications for saved searches have been fixed.

## 2.10.4

### Fixed

- Fixed an issue that caused the frontend to return a HTTP 500 and log an error message like:
  ```
  lvl=eror msg="ui HTTP handler error response" method=GET status_code=500 error="Post http://127.0.0.1:3182/repo-lookup: context canceled"
  ```

## 2.10.3

### Fixed

- The SAML AuthnRequest signature when using HTTP redirect binding is now computed using a URL query string with correct ordering of parameters. Previously, the ordering was incorrect and caused errors when the IdP was configured to check the signature in the AuthnRequest.

## 2.10.2

### Fixed

- SAML IdP-initiated login previously failed with the IdP set a RelayState value. This now works.

## 2.10.1

### Changed

- Most `experimentalFeatures` in the site configuration now respond to configuration changes live, without requiring a server restart. As usual, you will be prompted for a restart after saving your configuration changes if one is required.
- Gravatar image avatars are no longer displayed for committers.

## 2.10.0

### Changed

- In the file tree, if a directory that contains only a single directory is expanded, its child directory is now expanded automatically.

### Fixed

- Fixed an issue where `sourcegraph/server` would not start code intelligence containers properly when the `sourcegraph/server` container was shut down non-gracefully.
- Fixed an issue where the file tree would return an error when navigating between repositories.

## 2.9.4

### Changed

- Repo-updater has a new and improved scheduler for periodic repo fetches. If you have problems with it, you can revert to the old behavior by adding `"experimentalFeatures": { "updateScheduler": "disabled" }` to your `config.json`.
- A once-off migration will run changing the layout of cloned repos on disk. This should only affect installations created January 2018 or before. There should be no user visible changes.
- Experimental feature flag "updateScheduler" enables a smarter and less spammy algorithm for automatic repository updates.
- It is no longer possible to disable code intelligence by unsetting the LSP_PROXY environment variable. Instead, code intelligence can be disabled per language on the site admin page (e.g. https://sourcegraph.example.com/site-admin/code-intelligence).
- Bitbucket API requests made by Sourcegraph are now under a self-enforced API rate limit (since Bitbucket Server does not have a concept of rate limiting yet). This will reduce any chance of Sourcegraph slowing down or causing trouble for Bitbucket Server instances connected to it. The limits are: 7,200 total requests/hr, with a bucket size / maximum burst size of 500 requests.
- Global, org, and user settings are now validated against the schema, so invalid settings will be shown in the settings editor with a red squiggly line.
- The `http-header` auth provider now supports being used with other auth providers (still only when `experimentalFeatures.multipleAuthProviders` is `true`).
- Periodic fetches of Gitolite-hosted repositories are now handled internally by repo-updater.

### Added

- The `log.sentry.dsn` field in the site config makes Sourcegraph log application errors to a Sentry instance.
- Two new repository page hotkeys were added: <kbd>r</kbd> to open the repositories menu and <kbd>v</kbd> to open the revision selector.
- Repositories are periodically (~45 days) recloned from the codehost. The codehost can be relied on to give an efficient packing. This is an alternative to running a memory and CPU intensive git gc and git prune.
- The `auth.sessionExpiry` field sets the session expiration age in seconds (defaults to 90 days).

### Fixed

- Fixed a bug in the API console that caused it to display as a blank page in some cases.
- Fixed cases where GitHub rate limit wasn't being respected.
- Fixed a bug where scrolling in references, history, etc. file panels was not possible in Firefox.
- Fixed cases where gitserver directory structure migration could fail/crash.
- Fixed "Generate access token" link on user settings page. Previously, this link would 404.
- Fixed a bug where the search query was not updated in the search bar when searching from the homepage.
- Fixed a possible crash in github-proxy.
- Fixed a bug where file matching for diff search was case sensitive by default.

### Removed

- `SOURCEGRAPH_CONFIG` environment variable has been removed. Site configuration is always read from and written to disk. You can configure the location by providing `SOURCEGRAPH_CONFIG_FILE`. The default path is `/etc/sourcegraph/config.json`.

## 2.9.3

### Changed

- The search results page will merge duplicated lines of context.
- The following deprecated site configuration properties have been removed: `github[].preemptivelyClone`, `gitOriginMap`, `phabricatorURL`, `githubPersonalAccessToken`, `githubEnterpriseURL`, `githubEnterpriseCert`, and `githubEnterpriseAccessToken`.
- The `settings` field in the site config file is deprecated and will not be supported in a future release. Site admins should move those settings (if any) to global settings (in the site admin UI). Global settings are preferred to site config file settings because the former can be applied without needing to restart/redeploy the Sourcegraph server or cluster.

### Fixed

- Fixed a goroutine leak which occurs when search requests are canceled.
- Console output should have fewer spurious line breaks.
- Fixed an issue where it was not possible to override the `StrictHostKeyChecking` SSH option in the SSH configuration.
- Cross-repository code intelligence indexing for non-Go languages is now working again (originally broken in 2.9.2).

## 2.9.1

### Fixed

- Fixed an issue where saving an organization's configuration would hang indefinitely.

## 2.9.0

### Changed

- Hover tooltips were rewritten to fix a couple of issues and are now much more robust, received a new design and show more information.
- The `max:` search flag was renamed to `count:` in 2.8.8, but for backward compatibility `max:` has been added back as a deprecated alias for `count:`.
- Drastically improved the performance / load time of the Code Intelligence site admin page.

### Added

- The site admin code intelligence page now displays an error or reason whenever language servers are unable to be managed from the UI or Sourcegraph API.
- The ability to directly specify the root import path of a repository via `.sourcegraph/config.json` in the repo root, instead of relying on the heuristics of the Go language server to detect it.

### Fixed

- Configuring Bitbucket Server now correctly suppresses the the toast message "Configure repositories and code hosts to add to Sourcegraph."
- A bug where canonical import path comments would not be detected by the Go language server's heuristics under `cmd/` folders.
- Fixed an issue where a repository would only be refreshed on demand by certain user actions (such as a page reload) and would otherwise not be updated when expected.
- If a code host returned a repository-not-found or unauthorized error (to `repo-updater`) for a repository that previously was known to Sourcegraph, then in some cases a misleading "Empty repository" screen was shown. Now the repository is displayed as though it still existed, using cached data; site admins must explicitly delete repositories on Sourcegraph after they have been deleted on the code host.
- Improved handling of GitHub API rate limit exhaustion cases. Cached repository metadata and Git data will be used to provide full functionality during this time, and log messages are more informative. Previously, in some cases, repositories would become inaccessible.
- Fixed an issue where indexed search would sometimes not indicate that there were more results to show for a given file.
- Fixed an issue where the code intelligence admin page would never finish loading language servers.

## 2.9.0-pre0

### Changed

- Search scopes have been consolidated into the "Filters" bar on the search results page.
- Usernames and organization names of up to 255 characters are allowed. Previously the max length was 38.

### Fixed

- The target commit ID of a Git tag object (i.e., not lightweight Git tag refs) is now dereferenced correctly. Previously the tag object's OID was given.
- Fixed an issue where AWS Code Commit would hit the rate limit.
- Fixed an issue where dismissing the search suggestions dropdown did not unfocus previously highlighted suggestions.
- Fixed an issue where search suggestions would appear twice.
- Indexed searches now return partial results if they timeout.
- Git repositories with files whose paths contain `.git` path components are now usable (via indexed and non-indexed search and code intelligence). These corrupt repositories are rare and generally were created by converting some other VCS repository to Git (the Git CLI will forbid creation of such paths).
- Various diff search performance improvements and bug fixes.
- New Phabricator extension versions would used cached stylesheets instead of the upgraded version.
- Fixed an issue where hovers would show an error for Rust and C/C++ files.

### Added

- The `sourcegraph/server` container now emits the most recent log message when redis terminates to make it easier to debug why redis stopped.
- Organization invites (which allow users to invite other users to join organizations) are significantly improved. A new accept-invitation page was added.
- The new help popover allows users to easily file issues in the Sourcegraph public issue tracker and view documentation.
- An issue where Java files would be highlighted incorrectly if they contained JavaDoc blocks with an uneven number of opening/closing `*`s.

### Removed

- The `secretKey` site configuration value is no longer needed. It was only used for generating tokens for inviting a user to an organization. The invitation is now stored in the database associated with the recipient, so a secret token is no longer needed.
- The `experimentalFeatures.searchTimeoutParameter` site configuration value has been removed. It defaulted to `enabled` in 2.8 and it is no longer possible to disable.

### Added

- Syntax highlighting for:
  - TOML files (including Go `Gopkg.lock` and Rust `Cargo.lock` files).
  - Rust files.
  - GraphQL files.
  - Protobuf files.
  - `.editorconfig` files.

## 2.8.9

### Changed

- The "invite user" site admin page was moved to a sub-page of the users page (`/site-admin/users/new`).
- It is now possible for a site admin to create a new user without providing an email address.

### Fixed

- Checks for whether a repo is cloned will no longer exhaust open file pools over time.

### Added

- The Phabricator extension shows code intelligence status and supports enabling / disabling code intelligence for files.

## 2.8.8

### Changed

- Queries for repositories (in the explore, site admin repositories, and repository header dropdown) are matched on case-insensitive substrings, not using fuzzy matching logic.
- HTTP Authorization headers with an unrecognized scheme are ignored; they no longer cause the HTTP request to be rejected with HTTP 401 Unauthorized and an "Invalid Authorization header." error.
- Renamed the `max` search flag to `count`. Searches that specify `count:` will fetch at least that number of results, or the full result set.
- Bumped `lsp-proxy`'s `initialize` timeout to 3 minutes for every language.
- Search results are now sorted by repository and file name.
- More easily accessible "Show more" button at the top of the search results page.
- Results from user satisfaction surveys are now always hosted locally and visible to admins. The `"experimentalFeatures": { "hostSurveysLocally" }` config option has been deprecated.
- If the OpenID Connect authentication provider reports that a user's email address is not verified, the authentication attempt will fail.

### Fixed

- Fixed an issue where the search results page would not update its title.
- The session cookie name is now `sgs` (not `sg-session`) so that Sourcegraph 2.7 and Sourcegraph 2.8 can be run side-by-side temporarily during a rolling update without clearing each other's session cookies.
- Fixed the default hostnames of the C# and R language servers
- Fixed an issue where deleting an organization prevented the creation of organizations with the name of the deleted organization.
- Non-UTF8 encoded files (e.g. ISO-8859-1/Latin1, UTF16, etc) are now displayed as text properly rather than being detected as binary files.
- Improved error message when lsp-proxy's initalize timeout occurs
- Fixed compatibility issues and added [instructions for using Microsoft ADFS 2.1 and 3.0 for SAML authentication](https://docs.sourcegraph.com/admin/auth/saml_with_microsoft_adfs).
- Fixed an issue where external accounts associated with deleted user accounts would still be returned by the GraphQL API. This caused the site admin external accounts page to fail to render in some cases.
- Significantly reduced the number of code host requests for non github.com or gitlab.com repositories.

### Added

- The repository revisions popover now shows the target commit's last-committed/authored date for branches and tags.
- Setting the env var `INSECURE_SAML_LOG_TRACES=1` on the server (or the `sourcegraph-frontend` pod in Kubernetes) causes all SAML requests and responses to be logged, which helps with debugging SAML.
- Site admins can now view user satisfaction surveys grouped by user, in addition to chronological order, and aggregate summary values (including the average score and the net promoter score over the last 30 days) are now displayed.
- The site admin overview page displays the site ID, the primary admin email, and premium feature usage information.
- Added Haskell as an experimental language server on the code intelligence admin page.

## 2.8.0

### Changed

- `gitMaxConcurrentClones` now also limits the concurrency of updates to repos in addition to the initial clone.
- In the GraphQL API, `site.users` has been renamed to `users`, `site.orgs` has been renamed to `organizations`, and `site.repositories` has been renamed to `repositories`.
- An authentication provider must be set in site configuration (see [authentication provider documentation](https://docs.sourcegraph.com/admin/auth)). Previously the server defaulted to builtin auth if none was set.
- If a process dies inside the Sourcegraph container the whole container will shut down. We suggest operators configure a [Docker Restart Policy](https://docs.docker.com/config/containers/start-containers-automatically/#restart-policy-details) or a [Kubernetes Restart Policy](https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/#restart-policy). Previously the container would operate in a degraded mode if a process died.
- Changes to the `auth.public` site config are applied immediately in `sourcegraph/server` (no restart needed).
- The new search timeout behavior is now enabled by default. Set `"experimentalFeatures": {"searchTimeoutParameter": "disabled"}` in site config to disable it.
- Search includes files up to 1MB (previous limit was 512KB for unindexed search and 128KB for indexed search).
- Usernames and email addresses reported by OpenID Connect and SAML auth providers are now trusted, and users will sign into existing Sourcegraph accounts that match on the auth provider's reported username or email.
- The repository sidebar file tree is much, much faster on massive repositories (200,000+ files)
- The SAML authentication provider was significantly improved. Users who were signed in using SAML previously will need to reauthenticate via SAML next time they visit Sourcegraph.
- The SAML `serviceProviderCertificate` and `serviceProviderPrivateKey` site config properties are now optional.

### Fixed

- Fixed an issue where Index Search status page failed to render.
- User data on the site admin Analytics page is now paginated, filterable by a user's recent activity, and searchable.
- The link to the root of a repository in the repository header now preserves the revision you're currently viewing.
- When using the `http-header` auth provider, signin/signup/signout links are now hidden.
- Repository paths beginning with `go/` are no longer reservered by Sourcegraph.
- Interpret `X-Forwarded-Proto` HTTP header when `httpToHttpsRedirect` is set to `load-balanced`.
- Deleting a user account no longer prevents the creation of a new user account with the same username and/or association with authentication provider account (SAML/OpenID/etc.)
- It is now possible for a user to verify an email address that was previously associated with now-deleted user account.
- Diff searches over empty repositories no longer fail (this was not an issue for Sourcegraph cluster deployments).
- Stray `tmp_pack_*` files from interrupted fetches should now go away.
- When multiple `repo:` tokens match the same repo, process @revspec requirements from all of them, not just the first one in the search.

### Removed

- The `ssoUserHeader` site config property (deprecated since January 2018) has been removed. The functionality was moved to the `http-header` authentication provider.
- The experiment flag `showMissingReposEnabled`, which defaulted to enabled, has been removed so it is no longer possible to disable this feature.
- Event-level telemetry has been completely removed from self-hosted Sourcegraph instances. As a result, the `disableTelemetry` site configuration option has been deprecated. The new site-admin Pings page clarifies the only high-level telemetry being sent to Sourcegraph.com.
- The deprecated `adminUsernames` site config property (deprecated since January 2018) has been removed because it is no longer necessary. Site admins can designate other users as site admins in the site admin area, and the first user to sign into a new instance always becomes a site admin (even when using an external authentication provider).

### Added

- The new repository contributors page (linked from the repository homepage) displays the top Git commit authors in a repository, with filtering options.
- Custom language servers in the site config may now specify a `metadata` property containing things like homepage/docs/issues URLs for the language server project, as well as whether or not the language server should be considered experimental (not ready for prime-time). This `metadata` will be displayed in the UI to better communicate the status of a language server project.
- Access tokens now have scopes (which define the set of operations they permit). All access tokens still provide full control of all resources associated with the user account (the `user:all` scope, which is now explicitly displayed).
- The new access token scope `site-admin:sudo` allows the holder to perform any action as any other user. Only site admins may create this token.
- Links to Sourcegraph's changelog have been added to the site admin Updates page and update alert.
- If the site configuration is invalid or uses deprecated properties, a global alert will be shown to all site admins.
- There is now a code intelligence status indicator when viewing files. It contains information about the capabailities of the language server that is providing code intelligence for the file.
- Java code intelligence can now be enabled for repositories that aren't automatically supported using a
  `javaconfig.json` file. For Gradle plugins, this file can be generated using
  the [Javaconfig Gradle plugin](https://docs.sourcegraph.com/extensions/language_servers/java#gradle-execution).
- The new `auth.providers` site config is an array of authentication provider objects. Currently only 1 auth provider is supported. The singular `auth.provider` is deprecated.
- Users authenticated with OpenID Connect are now able to sign out of Sourcegraph (if the provider supports token revocation or the end-session endpoint).
- Users can now specify the number of days, weeks, and months of site activity to query through the GraphQL API.
- Added 14 new experimental language servers on the code intelligence admin page.
- Added `httpStrictTransportSecurity` site configuration option to customize the Strict-Transport-Security HTTP header. It defaults to `max-age=31536000` (one year).
- Added `nameIDFormat` in the `saml` auth provider to set the SAML NameID format. The default changed from transient to persistent.
- (This feature has been removed.) Experimental env var expansion in site config JSON: set `SOURCEGRAPH_EXPAND_CONFIG_VARS=1` to replace `${var}` or `$var` (based on environment variables) in any string value in site config JSON (except for JSON object property names).
- The new (optional) SAML `serviceProviderIssuer` site config property (in an `auth.providers` array entry with `{"type":"saml", ...}`) allows customizing the SAML Service Provider issuer name.
- The site admin area now has an "Auth" section that shows the enabled authentication provider(s) and users' external accounts.

## 2.7.6

### Fixed

- If a user's account is deleted, session cookies for that user are no longer considered valid.

## 2.7.5

### Changed

- When deploying Sourcegraph to Kubernetes, RBAC is now used by default. Most Kubernetes clusters require it. See the Kubernetes installation instructions for more information (including disabling if needed).
- Increased git ssh connection timeout to 30s from 7s.
- The Phabricator integration no longer requires staging areas, but using them is still recommended because it improves performance.

### Fixed

- Fixed an issue where language servers that were not enabled would display the "Restart" button in the Code Intelligence management panel.
- Fixed an issue where the "Update" button in the Code Intelligence management panel would be displayed inconsistently.
- Fixed an issue where toggling a dynamic search scope would not also remove `@rev` (if specified)
- Fixed an issue where where modes that can only be determined by the full filename (not just the file extension) of a path weren't supported (Dockerfiles are the first example of this).
- Fixed an issue where the GraphiQL console failed when variables are specified.
- Indexed search no longer maintains its own git clones. For Kubernetes cluster deployments, this significantly reduces disk size requirements for the indexed-search pod.
- Fixed an issue where language server Docker containers would not be automatically restarted if they crashed (`sourcegraph/server` only).
- Fixed an issue where if the first user on a site authenticated via SSO, the site would remain stuck in uninitialized mode.

### Added

- More detailed progress information is displayed on pages that are waiting for repositories to clone.
- Admins can now see charts with daily, weekly, and monthly unique user counts by visiting the site-admin Analytics page.
- Admins can now host and see results from Sourcegraph user satisfaction surveys locally by setting the `"experimentalFeatures": { "hostSurveysLocally": "enabled"}` site config option. This feature will be enabled for all instances once stable.
- Access tokens are now supported for all authentication providers (including OpenID Connect and SAML, which were previously not supported).
- The new `motd` setting (in global, organization, and user settings) displays specified messages at the top of all pages.
- Site admins may now view all access tokens site-wide (for all users) and revoke tokens from the new access tokens page in the site admin area.

## 2.7.0

### Changed

- Missing repositories no longer appear as search results. Instead, a count of repositories that were not found is displayed above the search results. Hovering over the count will reveal the names of the missing repositories.
- "Show more" on the search results page will now reveal results that have already been fetched (if such results exist) without needing to do a new query.
- The bottom panel (on a file) now shows more tabs, including docstrings, multiple definitions, references (as before), external references grouped by repository, implementations (if supported by the language server), and file history.
- The repository sidebar file tree is much faster on massive repositories (200,000+ files)

### Fixed

- Searches no longer block if the index is unavailable (e.g. after the index pod restarts). Instead, it respects the normal search timeout and reports the situation to the user if the index is not yet available.
- Repository results are no longer returned for filters that are not supported (e.g. if `file:` is part of the search query)
- Fixed an issue where file tree elements may be scrolled out of view on page load.
- Fixed an issue that caused "Could not ensure repository updated" log messages when trying to update a large number of repositories from gitolite.
- When using an HTTP authentication proxy (`"auth.provider": "http-header"`), usernames are now properly normalized (special characters including `.` replaced with `-`). This fixes an issue preventing users from signing in if their username contained these special characters.
- Fixed an issue where the site-admin Updates page would incorrectly report that update checking was turned off when `telemetryDisabled` was set, even as it continued to report new updates.
- `repo:` filters that match multiple repositories and contain a revision specifier now correctly return partial results even if some of the matching repositories don't have a matching revision.
- Removed hardcoded list of supported languages for code intelligence. Any language can work now and support is determined from the server response.
- Fixed an issue where modifying `config.json` on disk would not correctly mark the server as needing a restart.
- Fixed an issue where certain diff searches (with very sparse matches in a repository's history) would incorrectly report no results found.
- Fixed an issue where the `langservers` field in the site-configuration didn't require both the `language` and `address` field to be specified for each entry

### Added

- Users (and site admins) may now create and manage access tokens to authenticate API clients. The site config `auth.disableAccessTokens` (renamed to `auth.accessTokens` in 2.11) disables this new feature. Access tokens are currently only supported when using the `builtin` and `http-header` authentication providers (not OpenID Connect or SAML).
- User and site admin management capabilities for user email addresses are improved.
- The user and organization management UI has been greatly improved. Site admins may now administer all organizations (even those they aren't a member of) and may edit profile info and configuration for all users.
- If SSO is enabled (via OpenID Connect or SAML) and the SSO system provides user avatar images and/or display names, those are now used by Sourcegraph.
- Enable new search timeout behavior by setting `"experimentalFeatures": { "searchTimeoutParameter": "enabled"}` in your site config.
  - Adds a new `timeout:` parameter to customize the timeout for searches. It defaults to 10s and may not be set higher than 1m.
  - The value of the `timeout:` parameter is a string that can be parsed by [time.Duration](https://golang.org/pkg/time/#ParseDuration) (e.g. "100ms", "2s").
  - When `timeout:` is not provided, search optimizes for retuning results as soon as possible and will include slower kinds of results (e.g. symbols) only if they are found quickly.
  - When `timeout:` is provided, all result kinds are given the full timeout to complete.
- A new user settings tokens page was added that allows users to obtain a token that they can use to authenticate to the Sourcegraph API.
- Code intelligence indexes are now built for all repositories in the background, regardless of whether or not they are visited directly by a user.
- Language servers are now automatically enabled when visiting a repository. For example, visiting a Go repository will now automatically download and run the relevant Docker container for Go code intelligence.
  - This change only affects when Sourcegraph is deployed using the `sourcegraph/server` Docker image (not using Kubernetes).
  - You will need to use the new `docker run` command at https://docs.sourcegraph.com/#quickstart in order for this feature to be enabled. Otherwise, you will receive errors in the log about `/var/run/docker.sock` and things will work just as they did before. See https://docs.sourcegraph.com/extensions/language_servers for more information.
- The site admin Analytics page will now display the number of "Code Intelligence" actions each user has made, including hovers, jump to definitions, and find references, on the Sourcegraph webapp or in a code host integration or extension.
- An experimental cross repository jump to definition which consults the OSS index on Sourcegraph.com. This is disabled by default; use `"experimentalFeatures": { "jumpToDefOSSIndex": "enabled" }` in your site configuration to enable it.
- Users can now view Git branches, tags, and commits, and compare Git branches and revisions on Sourcegraph. (The code host icon in the header takes you to the commit on the code host.)
- A new admin panel allows you to view and manage language servers. For Docker deployments, it allows you to enable/disable/update/restart language servers at the click of a button. For cluster deployments, it shows the current status of language servers.
- Users can now tweet their feedback about Sourcegraph when clicking on the feedback smiley located in the navbar and filling out a Twitter feedback form.
- A new button in the repository header toggles on/off the Git history panel for the current file.

## 2.6.8

### Bug fixes

- Searches of `type:repo` now work correctly with "Show more" and the `max` parameter.
- Fixes an issue where the server would crash if the DB was not available upon startup.

## 2.6.7

### Added

- The duration that the frontend waits for the PostgreSQL database to become available is now configurable with the `DB_STARTUP_TIMEOUT` env var (the value is any valid Go duration string).
- Dynamic search filters now suggest exclusions of Go test files, vendored files and node_modules files.

## 2.6.6

### Added

- Authentication to Bitbucket Server using username-password credentials is now supported (in the `bitbucketServer` site config `username`/`password` options), for servers running Bitbucket Server version 2.4 and older (which don't support personal access tokens).

## 2.6.5

### Added

- The externally accessible URL path `/healthz` performs a basic application health check, returning HTTP 200 on success and HTTP 500 on failure.

### Behavior changes

- Read-only forks on GitHub are no longer synced by default. If you want to add a readonly fork, navigate directly to the repository page on Sourcegraph to add it (e.g. https://sourcegraph.mycompany.internal/github.com/owner/repo). This prevents your repositories list from being cluttered with a large number of private forks of a private repository that you have access to. One notable example is https://github.com/EpicGames/UnrealEngine.
- SAML cookies now expire after 90 days. The previous behavior was every 1 hour, which was unintentionally low.

## 2.6.4

### Added

- Improve search timeout error messages
- Performance improvements for searching regular expressions that do not start with a literal.

## 2.6.3

### Bug fixes

- Symbol results are now only returned for searches that contain `type:symbol`

## 2.6.2

### Added

- More detailed logging to help diagnose errors with third-party authentication providers.
- Anchors (such as `#my-section`) in rendered Markdown files are now supported.
- Instrumentation section for admins. For each service we expose pprof, prometheus metrics and traces.

### Bug fixes

- Applies a 1s timeout to symbol search if invoked without specifying `type:` to not block plain text results. No change of behaviour if `type:symbol` is given explicitly.
- Only show line wrap toggle for code-view-rendered files.

## 2.6.1

### Bug fixes

- Fixes a bug where typing in the search query field would modify the expanded state of file search results.
- Fixes a bug where new logins via OpenID Connect would fail with the error `SSO error: ID Token verification failed`.

## 2.6.0

### Added

- Support for [Bitbucket Server](https://www.atlassian.com/software/bitbucket/server) as a codehost. Configure via the `bitbucketServer` site config field.
- Prometheus gauges for git clone queue depth (`src_gitserver_clone_queue`) and git ls-remote queue depth (`src_gitserver_lsremote_queue`).
- Slack notifications for saved searches may now be added for individual users (not just organizations).
- The new search filter `lang:` filters results by programming language (example: `foo lang:go` or `foo -lang:clojure`).
- Dynamic filters: filters generated from your search results to help refine your results.
- Search queries that consist only of `file:` now show files whose path matches the filters (instead of no results).
- Sourcegraph now automatically detects basic `$GOPATH` configurations found in `.envrc` files in the root of repositories.
- You can now configure the effective `$GOPATH`s of a repository by adding a `.sourcegraph/config.json` file to your repository with the contents `{"go": {"GOPATH": ["mygopath"]}}`.
- A new `"blacklistGoGet": ["mydomain.org,myseconddomain.com"]` offers users a quick escape hatch in the event that Sourcegraph is making unwanted `go get` or `git clone` requests to their website due to incorrectly-configured monorepos. Most users will never use this option.
- Search suggestions and results now include symbol results. The new filter `type:symbol` causes only symbol results to be shown.
  Additionally, symbols for a repository can be browsed in the new symbols sidebar.
- You can now expand and collapse all items on a search results page or selectively expand and collapse individual items.

### Configuration changes

- Reduced the `gitMaxConcurrentClones` site config option's default value from 100 to 5, to help prevent too many concurrent clones from causing issues on code hosts.
- Changes to some site configuration options are now automatically detected and no longer require a server restart. After hitting Save in the UI, you will be informed if a server restart is required, per usual.
- Saved search notifications are now only sent to the owner of a saved search (all of an organization's members for an organization-level saved search, or a single user for a user-level saved search). The `notifyUsers` and `notifyOrganizations` properties underneath `search.savedQueries` have been removed.
- Slack webhook URLs are now defined in user/organization JSON settings, not on the organization profile page. Previously defined organization Slack webhook URLs are automatically migrated to the organization's JSON settings.
- The "unlimited" value for `maxReposToSearch` is now `-1` instead of `0`, and `0` now means to use the default.
- `auth.provider` must be set (`builtin`, `openidconnect`, `saml`, `http-header`, etc.) to configure an authentication provider. Previously you could just set the detailed configuration property (`"auth.openIDConnect": {...}`, etc.) and it would implicitly enable that authentication provider.
- The `autoRepoAdd` site configuration property was removed. Site admins can add repositories via site configuration.

### Bug fixes

- Only cross reference index enabled repositories.
- Fixed an issue where search would return results with empty file contents for matches in submodules with indexing enabled. Searching over submodules is not supported yet, so these (empty) results have been removed.
- Fixed an issue where match highlighting would be incorrect on lines that contained multibyte characters.
- Fixed an issue where search suggestions would always link to master (and 404) even if the file only existed on a branch. Now suggestions always link to the revision that is being searched over.
- Fixed an issue where all file and repository links on the search results page (for all search results types) would always link to master branch, even if the results only existed in another branch. Now search results links always link to the revision that is being searched over.
- The first user to sign up for a (not-yet-initialized) server is made the site admin, even if they signed up using SSO. Previously if the first user signed up using SSO, they would not be a site admin and no site admin could be created.
- Fixed an issue where our code intelligence archive cache (in `lsp-proxy`) would not evict items from the disk. This would lead to disks running out of free space.

## 2.5.16, 2.5.17

- Version bump to keep deployment variants in sync.

## 2.5.15

### Bug fixes

- Fixed issue where a Sourcegraph cluster would incorrectly show "An update is available".
- Fixed Phabricator links to repositories
- Searches over a single repository are now less likely to immediately time out the first time they are searched.
- Fixed a bug where `auth.provider == "http-header"` would incorrectly require builtin authentication / block site access when `auth.public == "false"`.

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

- Code Intelligence support
- Custom links to code hosts with the `links:` config options in `repos.list`

### Changed

- Search by file path enabled by default

## 2.4.2

### Added

- Repository settings mirror/cloning diagnostics page

### Changed

- Repositories added from GitHub are no longer enabled by default. The site admin UI for enabling/disabling repositories is improved.

## 2.4.0

### Added

- Search files by name by including `type:path` in a search query
- Global alerts for configuration-needed and cloning-in-progress
- Better list interfaces for repositories, users, organizations, and threads
- Users can change their own password in settings
- Repository groups can now be specified in settings by site admins, organizations, and users. Then `repogroup:foo` in a search query will search over only those repositories specified for the `foo` repository group.

### Changed

- Log messages are much quieter by default

## 2.3.11

### Added

- Added site admin updates page and update checking
- Added site admin telemetry page

### Changed

- Enhanced site admin panel
- Changed repo- and SSO-related site config property names to be consistent, updated documentation

## 2.3.10

### Added

- Online site configuration editing and reloading

### Changed

- Site admins are now configured in the site admin area instead of in the `adminUsernames` config key or `ADMIN_USERNAMES` env var. Users specified in those deprecated configs will be designated as site admins in the database upon server startup until those configs are removed in a future release.

## 2.3.9

### Fixed

- An issue that prevented creation and deletion of saved queries

## 2.3.8

### Added

- Built-in authentication: you can now sign up without an SSO provider.
- Faster default branch code search via indexing.

### Fixed

- Many performance improvements to search.
- Much log spam has been eliminated.

### Changed

- We optionally read `SOURCEGRAPH_CONFIG` from `$DATA_DIR/config.json`.
- SSH key required to clone repositories from GitHub Enterprise when using a self-signed certificate.

## 0.3 - 13 December 2017

The last version without a CHANGELOG.
