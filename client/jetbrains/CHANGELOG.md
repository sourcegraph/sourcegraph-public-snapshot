# Sourcegraph Changelog

## [Unreleased]

### Highlights

### Merged PRs

## [3.3.0]

### Highlights

- New: setting to accept non-trusted SSL certificates
- New: button to stop generation of a chat reply
- New: Cody icon in the status bar to indicate you're not logged in

### Merged PRs

- Total [6 PRs](https://github.com/sourcegraph/sourcegraph/milestone/234?closed=1) merged since last release
- JetBrains: add the client value to telemetry events,
  by [@chwarwick](https://github.com/chwarwick) ([#57015](https://github.com/sourcegraph/sourcegraph/pull/57015))
- JetBrains: show icon & notification when user is not logged in,
  by [@Strum355](https://github.com/Strum355) ([#57006](https://github.com/sourcegraph/sourcegraph/pull/57006))
- JetBrains: Cody: Only enable the verbose debug setting if regular debug is also enabled,
  by [@Gedochao](https://github.com/Gedochao) ([#57000](https://github.com/sourcegraph/sourcegraph/pull/57000))
- JetBrains: Cody: Add a setting for accepting non-trusted certificates automatically,
  by [@Gedochao](https://github.com/Gedochao) ([#56999](https://github.com/sourcegraph/sourcegraph/pull/56999))
- JetBrains: add "Stop" button to cancel pending chat message,
  by [@szymonprz](https://github.com/szymonprz) ([#56998](https://github.com/sourcegraph/sourcegraph/pull/56998))
- JetBrains: Cody: Improve autocomplete custom color setting,
  by [@Gedochao](https://github.com/Gedochao) ([#56997](https://github.com/sourcegraph/sourcegraph/pull/56997))

## [3.2.0]

### Highlights

- Bugfix: autocomplete was not working in some cases
- Bugfix: chat was not working in some cases
- New: improved onboarding in the Cody sidebar after signing in for the first time
- New: simpler login to sourcegraph.com through settings UI
- Changed: Cody sidebar now defaults to the right instead of left allowing you to keep the "Project" file tree open
  alongside the Cody sidebar. You can drag the sidebar to the left if you don't like the new default.
- Changed: Cody icon in the sidebar is now monochromatic (grey)

### Merged PRs

- Total [17 PRs](https://github.com/sourcegraph/sourcegraph/milestone/233?closed=1) merged since the last release
- JetBrains: Cody: Fix the rendering of language table header in Cody settings,
  by [@Gedochao](https://github.com/Gedochao) ([#56956](https://github.com/sourcegraph/sourcegraph/pull/56956))
- JetBrains: add restarting logic for agent,
  by [@olafurpg](https://github.com/olafurpg) ([#56954](https://github.com/sourcegraph/sourcegraph/pull/56954))
- JetBrains: Cody: Check if the project wasn't disposed when getting the active account,
  by [@Gedochao](https://github.com/Gedochao) ([#56945](https://github.com/sourcegraph/sourcegraph/pull/56945))
- JetBrains: Fix: don't try to display inlays on not supported editor instances,
  by [@szymonprz](https://github.com/szymonprz) ([#56951](https://github.com/sourcegraph/sourcegraph/pull/56951))
- JetBrains: Cody: Improve the languages table in Cody settings,
  by [@Gedochao](https://github.com/Gedochao) ([#56939](https://github.com/sourcegraph/sourcegraph/pull/56939))
- JetBrains: Sign in to sourcegraph.com with a webflow from action on settings page,
  by [@szymonprz](https://github.com/szymonprz) ([#56941](https://github.com/sourcegraph/sourcegraph/pull/56941))
- JetBrains: fix bug for access token description,
  by [@olafurpg](https://github.com/olafurpg) ([#56897](https://github.com/sourcegraph/sourcegraph/pull/56897))
- nix: wrap gradle with x11 libs for intellij plugin,
  by [@Strum355](https://github.com/Strum355) ([#56890](https://github.com/sourcegraph/sourcegraph/pull/56890))
- JetBrains: display icon only instead of a button for embedding status,
  by [@szymonprz](https://github.com/szymonprz) ([#56884](https://github.com/sourcegraph/sourcegraph/pull/56884))
- JetBrains: Cody: Fix/remove code at risk of compatibility issues with the JetBrains platform,
  by [@Gedochao](https://github.com/Gedochao) ([#56885](https://github.com/sourcegraph/sourcegraph/pull/56885))
- call it "Cody" not "Cody AI",
  by [@sqs](https://github.com/sqs) ([#56863](https://github.com/sourcegraph/sourcegraph/pull/56863))
- JetBrains: use monochromatic icon in sidebar,
  by [@szymonprz](https://github.com/szymonprz) ([#56881](https://github.com/sourcegraph/sourcegraph/pull/56881))
- JetBrains: bump kotlin to 1.9.10,
  by [@Strum355](https://github.com/Strum355) ([#56828](https://github.com/sourcegraph/sourcegraph/pull/56828))
- JetBrains: improve design for redirect page when authorizing through web flow,
  by [@danielmarquespt](https://github.com/danielmarquespt) ([#56829](https://github.com/sourcegraph/sourcegraph/pull/56829))
- JetBrains: Display onboarding guidance after successful account creation,
  by [@szymonprz](https://github.com/szymonprz) ([#56824](https://github.com/sourcegraph/sourcegraph/pull/56824))
- JetBrains: Cody: Placing cody sidebar on the right side by default,
  by [@Sa1to](https://github.com/Sa1to) ([#56822](https://github.com/sourcegraph/sourcegraph/pull/56822))

## [3.1.0]

The Cody functionality of this plugin remains **experimental**. Don't hesitate to report
a [new issue](https://github.com/sourcegraph/sourcegraph/issues/new/?title=JetBrains:) on GitHub if you encounter
problems.

### Highlights

- Overall improved quality and performance for Cody autocomplete and chat
- New: simpler login flow for both sourcegraph.com and enterprise instances
- New: Cody icon in the status bar
- New: alt-backslash shortcut to trigger autocomplete
- New: setting to enable/disable autocomplete per language
- New: setting to customize the color of autocomplete hints
- New: ability to customize codebase context for Cody chat
- Fixed: compatibility issue with IdeaVIM
- Changed: settings are now separated by Sourcegraph, Cody, and Code Search
- Changed: recipes are now called "commands"
- Compatibility: this plugin is now only compatible with IDEA 2022.1+ (was previously 2021.2+)

### Merged PRs

- Total [99 PRs](https://github.com/sourcegraph/sourcegraph/milestone/232?closed=1) merged since the last release
- JetBrains: Cody: Bump JetBrains platform compat to `221.5080.210`,
  by [@Gedochao](https://github.com/Gedochao) ([#56625](https://github.com/sourcegraph/sourcegraph/pull/56625))
- JetBrains: fix capitalization,
  by [@vdavid](https://github.com/vdavid) ([#56576](https://github.com/sourcegraph/sourcegraph/pull/56576))
- JetBrains: unify onboarding notifications,
  by [@szymonprz](https://github.com/szymonprz) ([#56610](https://github.com/sourcegraph/sourcegraph/pull/56610))
- JetBrains: refine autocomplete trigger,
  by [@olafurpg](https://github.com/olafurpg) ([#56588](https://github.com/sourcegraph/sourcegraph/pull/56588))
- JetBrains: document `auth` package ,
  by [@szymonprz](https://github.com/szymonprz) ([#56541](https://github.com/sourcegraph/sourcegraph/pull/56541))
- JetBrains: Fix formatting,
  by [@olafurpg](https://github.com/olafurpg) ([#56540](https://github.com/sourcegraph/sourcegraph/pull/56540))
- JetBrains: Just use \n,
  by [@vdavid](https://github.com/vdavid) ([#56531](https://github.com/sourcegraph/sourcegraph/pull/56531))
- JetBrains: update contributing docs,
  by [@olafurpg](https://github.com/olafurpg) ([#56513](https://github.com/sourcegraph/sourcegraph/pull/56513))
- JetBrains: correctly manage chat transcript,
  by [@olafurpg](https://github.com/olafurpg) ([#56671](https://github.com/sourcegraph/sourcegraph/pull/56671))
- JetBrains: refine embedding status indicator,
  by [@olafurpg](https://github.com/olafurpg) ([#56636](https://github.com/sourcegraph/sourcegraph/pull/56636))
- JetBrains: Cody: Verify plugin compatibility,
  by [@Gedochao](https://github.com/Gedochao) ([#56634](https://github.com/sourcegraph/sourcegraph/pull/56634))
- JetBrains: Update README,
  by [@iskyOS](https://github.com/iskyOS) ([#56437](https://github.com/sourcegraph/sourcegraph/pull/56437))
- JetBrains: removes code search onboarding,
  by [@szymonprz](https://github.com/szymonprz) ([#56564](https://github.com/sourcegraph/sourcegraph/pull/56564))
- JetBrains: fix onboarding bug with web flow,
  by [@olafurpg](https://github.com/olafurpg) ([#56753](https://github.com/sourcegraph/sourcegraph/pull/56753))
- JetBrains: Move referrer setting to site-wide,
  by [@vdavid](https://github.com/vdavid) ([#56722](https://github.com/sourcegraph/sourcegraph/pull/56722))
- JetBrains: make accounts panel smaller,
  by [@szymonprz](https://github.com/szymonprz) ([#56454](https://github.com/sourcegraph/sourcegraph/pull/56454))
- JetBrains: Return to IDE after token generation,
  by [@vdavid](https://github.com/vdavid) ([#56681](https://github.com/sourcegraph/sourcegraph/pull/56681))
- JetBrains: add button to change codebase context,
  by [@olafurpg](https://github.com/olafurpg) ([#56683](https://github.com/sourcegraph/sourcegraph/pull/56683))
- JetBrains: ensure active account is set,
  by [@szymonprz](https://github.com/szymonprz) ([#56670](https://github.com/sourcegraph/sourcegraph/pull/56670))
- JetBrains: release 3.1.0-alpha.8,
  by [@olafurpg](https://github.com/olafurpg) ([#56614](https://github.com/sourcegraph/sourcegraph/pull/56614))
- JetBrains: fix ktfmt Gradle/Spotless config,
  by [@olafurpg](https://github.com/olafurpg) ([#56613](https://github.com/sourcegraph/sourcegraph/pull/56613))
- JetBrains: Cody: Add a status bar action for enabling/disabling autocomplete for a language,
  by [@Gedochao](https://github.com/Gedochao) ([#56724](https://github.com/sourcegraph/sourcegraph/pull/56724))
- JetBrains: Cody: Add actions for opening settings from the status bar,
  by [@Gedochao](https://github.com/Gedochao) ([#56754](https://github.com/sourcegraph/sourcegraph/pull/56754))
- JetBrains: Detect line separator,
  by [@vdavid](https://github.com/vdavid) ([#56053](https://github.com/sourcegraph/sourcegraph/pull/56053))
- JetBrains: automatically generate access token description,
  by [@olafurpg](https://github.com/olafurpg) ([#56570](https://github.com/sourcegraph/sourcegraph/pull/56570))
- JetBrains: Cody: Fixing manual autocomplete trigger,
  by [@Sa1to](https://github.com/Sa1to) ([#56575](https://github.com/sourcegraph/sourcegraph/pull/56575))
- JetBrains: release 3.1.0-alpha.7,
  by [@olafurpg](https://github.com/olafurpg) ([#56565](https://github.com/sourcegraph/sourcegraph/pull/56565))
- JetBrains: Cody: Switch to master-layout-based settings UI,
  by [@Gedochao](https://github.com/Gedochao) ([#56579](https://github.com/sourcegraph/sourcegraph/pull/56579))
- JetBrains: fix unstable IntelliJ public API changes,
  by [@szymonprz](https://github.com/szymonprz) ([#56539](https://github.com/sourcegraph/sourcegraph/pull/56539))
- JetBrains: add link to generate new token,
  by [@szymonprz](https://github.com/szymonprz) ([#56532](https://github.com/sourcegraph/sourcegraph/pull/56532))
- JetBrains: release 3.1.0-alpha.4,
  by [@olafurpg](https://github.com/olafurpg) ([#56533](https://github.com/sourcegraph/sourcegraph/pull/56533))
- JetBrains: display initial sign in with sourcegraph panel,
  by [@szymonprz](https://github.com/szymonprz) ([#56633](https://github.com/sourcegraph/sourcegraph/pull/56633))
- JetBrains: fix repo embedding status in chat sidebar,
  by [@olafurpg](https://github.com/olafurpg) ([#56584](https://github.com/sourcegraph/sourcegraph/pull/56584))
- JetBrains: clear last autocomplete candidate after acceptance,
  by [@chwarwick](https://github.com/chwarwick) ([#56505](https://github.com/sourcegraph/sourcegraph/pull/56505))
- JetBrains: Rename plugin,
  by [@vdavid](https://github.com/vdavid) ([#56230](https://github.com/sourcegraph/sourcegraph/pull/56230))
- JetBrains: move caching graphql query loader to our sources from IntelliJ,
  by [@szymonprz](https://github.com/szymonprz) ([#56542](https://github.com/sourcegraph/sourcegraph/pull/56542))
- JetBrains: Cody: Allow manual autocomplete to be triggered even when implicit completions are disabled ,
  by [@Gedochao](https://github.com/Gedochao) ([#56473](https://github.com/sourcegraph/sourcegraph/pull/56473))
- JetBrains: stop logging `completion:started` events,
  by [@olafurpg](https://github.com/olafurpg) ([#56452](https://github.com/sourcegraph/sourcegraph/pull/56452))
- JetBrains: add ability to edit account from context menu,
  by [@szymonprz](https://github.com/szymonprz) ([#56493](https://github.com/sourcegraph/sourcegraph/pull/56493))
- JetBrains: improve and update plugin setup guide,
  by [@MaedahBatool](https://github.com/MaedahBatool) ([#56280](https://github.com/sourcegraph/sourcegraph/pull/56280))
- JetBrains: add cancelation support,
  by [@olafurpg](https://github.com/olafurpg) ([#56119](https://github.com/sourcegraph/sourcegraph/pull/56119))
- JetBrains: customize autocomplete color new branch,
  by [@Sa1to](https://github.com/Sa1to) ([#56391](https://github.com/sourcegraph/sourcegraph/pull/56391))
- JetBrains: Cody: Allow to enable/disable autocomplete per-language,
  by [@Gedochao](https://github.com/Gedochao) ([#56411](https://github.com/sourcegraph/sourcegraph/pull/56411))
- JetBrains: redesign settings to accounts model,
  by [@szymonprz](https://github.com/szymonprz) ([#56362](https://github.com/sourcegraph/sourcegraph/pull/56362))
- JetBrains: update autocomplete provider list and default to null,
  by [@chwarwick](https://github.com/chwarwick) ([#56515](https://github.com/sourcegraph/sourcegraph/pull/56515))
- JetBrains: Cody: Remove inter-configurable links from settings,
  by [@Gedochao](https://github.com/Gedochao) ([#56680](https://github.com/sourcegraph/sourcegraph/pull/56680))
- JetBrains: Cody: Make implicit autocomplete trigger enabled by default,
  by [@Gedochao](https://github.com/Gedochao) ([#56617](https://github.com/sourcegraph/sourcegraph/pull/56617))
- JetBrains: forward SOCKS `proxy` setting to agent,
  by [@cbart](https://github.com/cbart) ([#56264](https://github.com/sourcegraph/sourcegraph/pull/56264))
- JetBrains: use "Sourcegraph & Cody" title in settings UI,
  by [@olafurpg](https://github.com/olafurpg) ([#56672](https://github.com/sourcegraph/sourcegraph/pull/56672))
- JetBrains: Use agent for recipes,
  by [@vdavid](https://github.com/vdavid) ([#56196](https://github.com/sourcegraph/sourcegraph/pull/56196))
- JetBrains: manually construct server url to fix any issues with it,
  by [@szymonprz](https://github.com/szymonprz) ([#56419](https://github.com/sourcegraph/sourcegraph/pull/56419))
- gh: give a proper name to JetBrains test job,
  by [@jhchabran](https://github.com/jhchabran) ([#56622](https://github.com/sourcegraph/sourcegraph/pull/56622))
- JetBrains: require restart after installation,
  by [@olafurpg](https://github.com/olafurpg) ([#56174](https://github.com/sourcegraph/sourcegraph/pull/56174))
- JetBrains: fix `CodyAgent.isConnected` logic,
  by [@olafurpg](https://github.com/olafurpg) ([#56173](https://github.com/sourcegraph/sourcegraph/pull/56173))
- JetBrains: Cody: disable autocomplete with IdeaVIM visual mode,
  by [@Sa1to](https://github.com/Sa1to) ([#56506](https://github.com/sourcegraph/sourcegraph/pull/56506))
- JetBrains: Cody: Removing autocomplete when esc pressed using IdeaVIM…,
  by [@Gedochao](https://github.com/Gedochao) ([#56347](https://github.com/sourcegraph/sourcegraph/pull/56347))
- JetBrains: Cody: Disable autocomplete when navigating with arrow keys,
  by [@Sa1to](https://github.com/Sa1to) ([#56543](https://github.com/sourcegraph/sourcegraph/pull/56543))
- JetBrains: improve settings UI,
  by [@olafurpg](https://github.com/olafurpg) ([#55876](https://github.com/sourcegraph/sourcegraph/pull/55876))
- JetBrains: catch exception - fix #56032,
  by [@olafurpg](https://github.com/olafurpg) ([#56048](https://github.com/sourcegraph/sourcegraph/pull/56048))
- JetBrains: Add extension details to logged events,
  by [@chwarwick](https://github.com/chwarwick) ([#56267](https://github.com/sourcegraph/sourcegraph/pull/56267))
- JetBrains: Cody: Add a status bar toggle to enable/disable autocomplete,
  by [@Gedochao](https://github.com/Gedochao) ([#56310](https://github.com/sourcegraph/sourcegraph/pull/56310))
- JetBrains: Cody: Disabling autocomplete in readonly files,
  by [@olafurpg](https://github.com/olafurpg) ([#56256](https://github.com/sourcegraph/sourcegraph/pull/56256))
- JetBrains: Add remaining autocomplete telemetry parameters ,
  by [@chwarwick](https://github.com/chwarwick) ([#56242](https://github.com/sourcegraph/sourcegraph/pull/56242))
- JetBrains: Detect other completion plugins,
  by [@vdavid](https://github.com/vdavid) ([#55932](https://github.com/sourcegraph/sourcegraph/pull/55932))
- JetBrains: Add context summary to completion:suggested,
  by [@vdavid](https://github.com/vdavid) ([#56203](https://github.com/sourcegraph/sourcegraph/pull/56203))
- JetBrains: log `contextSummary` property for autocomplete telemetry,
  by [@olafurpg](https://github.com/olafurpg) ([#56198](https://github.com/sourcegraph/sourcegraph/pull/56198))
- JetBrains: release 3.1.0-alpha.3,
  by [@olafurpg](https://github.com/olafurpg) ([#56193](https://github.com/sourcegraph/sourcegraph/pull/56193))
- JetBrains: prevent null pointer exception,
  by [@olafurpg](https://github.com/olafurpg) ([#55869](https://github.com/sourcegraph/sourcegraph/pull/55869))
- JetBrains: don't warn about missing context with empty context files,
  by [@olafurpg](https://github.com/olafurpg) ([#56628](https://github.com/sourcegraph/sourcegraph/pull/56628))
- JetBrains: release 3.1.0-alpha.2,
  by [@olafurpg](https://github.com/olafurpg) ([#56166](https://github.com/sourcegraph/sourcegraph/pull/56166))
- JetBrains: consistently use `Autocomplete` instead of `AutoComplete`,
  by [@olafurpg](https://github.com/olafurpg) ([#56106](https://github.com/sourcegraph/sourcegraph/pull/56106))
- JetBrains: Cody: Hotfix: notify agent of configuration change on account settings change,
  by [@Gedochao](https://github.com/Gedochao) ([#56581](https://github.com/sourcegraph/sourcegraph/pull/56581))
- JetBrains: Fix agent binary name for Windows,
  by [@vdavid](https://github.com/vdavid) ([#56055](https://github.com/sourcegraph/sourcegraph/pull/56055))
- JetBrains: format build.gradle.kts with spotless,
  by [@olafurpg](https://github.com/olafurpg) ([#56047](https://github.com/sourcegraph/sourcegraph/pull/56047))
- nix: initial client/intellij nix flake,
  by [@Strum355](https://github.com/Strum355) ([#56319](https://github.com/sourcegraph/sourcegraph/pull/56319))
- JetBrains: display visual hint about cody autocomplete status in the status bar,
  by [@szymonprz](https://github.com/szymonprz) ([#56046](https://github.com/sourcegraph/sourcegraph/pull/56046))
- JetBrains: remove non-agent autocomplete and chat,
  by [@olafurpg](https://github.com/olafurpg) ([#55997](https://github.com/sourcegraph/sourcegraph/pull/55997))
- JetBrains: release 3.1.0-alpha.1,
  by [@olafurpg](https://github.com/olafurpg) ([#56008](https://github.com/sourcegraph/sourcegraph/pull/56008))
- JetBrains: move all network traffic to agent,
  by [@olafurpg](https://github.com/olafurpg) ([#56001](https://github.com/sourcegraph/sourcegraph/pull/56001))
- JetBrains: release 3.0.10-alpha.3,
  by [@olafurpg](https://github.com/olafurpg) ([#55971](https://github.com/sourcegraph/sourcegraph/pull/55971))
- JetBrains: hide autocomplete with ESC keyboard shortcut,
  by [@szymonprz](https://github.com/szymonprz) ([#55955](https://github.com/sourcegraph/sourcegraph/pull/55955))
- JetBrains: Rename "recipes" to "commands" in docs and UI,
  by [@vdavid](https://github.com/vdavid) ([#56229](https://github.com/sourcegraph/sourcegraph/pull/56229))
- JetBrains: release 3.0.10-alpha,
  by [@olafurpg](https://github.com/olafurpg) ([#55888](https://github.com/sourcegraph/sourcegraph/pull/55888))
- JetBrains: Send config changed event when codebase changes,
  by [@chwarwick](https://github.com/chwarwick) ([#55900](https://github.com/sourcegraph/sourcegraph/pull/55900))
- JetBrains: ensure nonnull line separator method can't return a null ,
  by [@chwarwick](https://github.com/chwarwick) ([#56334](https://github.com/sourcegraph/sourcegraph/pull/56334))
- JetBrains: add alt-backslash shortcut to trigger autocomplete,
  by [@olafurpg](https://github.com/olafurpg) ([#55926](https://github.com/sourcegraph/sourcegraph/pull/55926))
- JetBrains: update CONTRIBUTING.md docs to mention steps to use unstable-codegen through a proxy,
  by [@cbart](https://github.com/cbart) ([#56263](https://github.com/sourcegraph/sourcegraph/pull/56263))
- JetBrains: don't start/stop agent for disposed projects,
  by [@olafurpg](https://github.com/olafurpg) ([#56189](https://github.com/sourcegraph/sourcegraph/pull/56189))
- JetBrains: start/stop agent on install/unistall events,
  by [@olafurpg](https://github.com/olafurpg) ([#56116](https://github.com/sourcegraph/sourcegraph/pull/56116))
- JetBrains: fix path of agent binaries in plugin zip,
  by [@olafurpg](https://github.com/olafurpg) ([#55867](https://github.com/sourcegraph/sourcegraph/pull/55867))
- JetBrains: add additional checks to ensure agent is connected before communicating,
  by [@chwarwick](https://github.com/chwarwick) ([#55999](https://github.com/sourcegraph/sourcegraph/pull/55999))
- JetBrains: await on agent server before sending telemetry,
  by [@olafurpg](https://github.com/olafurpg) ([#56007](https://github.com/sourcegraph/sourcegraph/pull/56007))
- JetBrains: add options to accept non-trusted certificates,
  by [@olafurpg](https://github.com/olafurpg) ([#56004](https://github.com/sourcegraph/sourcegraph/pull/56004))
- JetBrains: make sure caret is visible after accepting autocomplete,
  by [@olafurpg](https://github.com/olafurpg) ([#55924](https://github.com/sourcegraph/sourcegraph/pull/55924))
- Delve for all packages,
  by [@rafax](https://github.com/rafax) ([#56535](https://github.com/sourcegraph/sourcegraph/pull/56535))
- Access tokens: add URL search parameter to control default description,
  by [@olafurpg](https://github.com/olafurpg) ([#56536](https://github.com/sourcegraph/sourcegraph/pull/56536))
- search jobs: implement Resolver.CancelSearchJob,
  by [@stefanhengl](https://github.com/stefanhengl) ([#56365](https://github.com/sourcegraph/sourcegraph/pull/56365))
- ci: skip `sg lint` and `bazel` operations when only client/jetbrains changed,
  by [@burmudar](https://github.com/burmudar) ([#56061](https://github.com/sourcegraph/sourcegraph/pull/56061))
- ci: fix pipeline,
  by [@jhchabran](https://github.com/jhchabran) ([#56274](https://github.com/sourcegraph/sourcegraph/pull/56274))

## [3.0.9]

### Changed

- Store application level access tokens in a safe way [#55251](https://github.com/sourcegraph/sourcegraph/pull/55251)
- Autocomplete is now powered by the agent when enabled (off by
  default) [#55638](https://github.com/sourcegraph/sourcegraph/pull/55638), [#55826](https://github.com/sourcegraph/sourcegraph/pull/55826)

### Fixed

- Removed jumping text effect from the chat when generating
  response [#55357](https://github.com/sourcegraph/sourcegraph/pull/55357)
- Chat message doesn't jump after finished response
  generation [#55390](https://github.com/sourcegraph/sourcegraph/pull/55390)
- Removed jumping text effect from the chat when generating
  response [#55357](https://github.com/sourcegraph/sourcegraph/pull/55357)

## [3.0.8]

### Fixed

- Improved the auto-scrolling of the Cody chat [#55150](https://github.com/sourcegraph/sourcegraph/pull/55150)
- Fixed mouse wheel and mouse drag scrolling in the Cody
  chat [#55199](https://github.com/sourcegraph/sourcegraph/pull/55199)

## [3.0.7]

### Added

- New menu item in the toolbar cogwheel menu to open the Cody app
  settings [#55146](https://github.com/sourcegraph/sourcegraph/pull/55146)

### Changed

- Improved UI of the onboarding widgets [#55090](https://github.com/sourcegraph/sourcegraph/pull/55090)
- Improved perceived autocomplete performance [#55098](https://github.com/sourcegraph/sourcegraph/pull/55098)

### Fixed

- Enable/disable Cody automatically based on the
  settings [#55138](https://github.com/sourcegraph/sourcegraph/pull/55138)

## [3.0.6]

### Added

- Automatic detection of Cody app status in the settings
  window [#54955](https://github.com/sourcegraph/sourcegraph/pull/54955)
- Add "Enable Cody" option to settings [#55004](https://github.com/sourcegraph/sourcegraph/pull/55004)

### Changed

- Disable "summarize recent code changes" button if git repository is not
  available [#54859](https://github.com/sourcegraph/sourcegraph/pull/54859)
- Get the chat model max tokens value from the instance when
  available [#54954](https://github.com/sourcegraph/sourcegraph/pull/54954)

### Fixed

- Downgraded connection errors for invalid or inaccessible enterprise instances to
  warnings [#54916](https://github.com/sourcegraph/sourcegraph/pull/54916)
- Try to log error stack traces and recover from them, rather than re-throw the
  exception [#54917](https://github.com/sourcegraph/sourcegraph/pull/54917)
- Show only one informative message in case of invalid access
  token [#54951](https://github.com/sourcegraph/sourcegraph/pull/54951)
- Don't display `<br />` tag in the chat message when trying to insert new line in the code
  block [#55007](https://github.com/sourcegraph/sourcegraph/pull/55007)

## [3.0.5]

### Added

- Added embeddings status in footer [#54575](https://github.com/sourcegraph/sourcegraph/pull/54575)
- Added currently opened file name in footer [#54610](https://github.com/sourcegraph/sourcegraph/pull/54610)
- Auto-growing prompt input [#53594](https://github.com/sourcegraph/sourcegraph/pull/53594)
- Added "stop generating" button [#54710](https://github.com/sourcegraph/sourcegraph/pull/54710)
- Copy code block button added to editor in the chat message to copy the text to
  clipboard [#54783](https://github.com/sourcegraph/sourcegraph/pull/54783)
- Insert at Cursor button added to editor in the chat message to insert the text form the editor to main
  editor [#54815](https://github.com/sourcegraph/sourcegraph/pull/54815)
- Added support for multiline autocomplete [#54848](https://github.com/sourcegraph/sourcegraph/pull/54848)

### Fixed

- Fixed telemetry for Sourcegraph.com [#54885](https://github.com/sourcegraph/sourcegraph/pull/54885)

## [3.0.4]

### Added

- Added embeddings status in footer [#54575](https://github.com/sourcegraph/sourcegraph/pull/54575)
- Added currently opened file name in footer [#54610](https://github.com/sourcegraph/sourcegraph/pull/54610)
- Added "stop generating" button [#54710](https://github.com/sourcegraph/sourcegraph/pull/54710)
- Made prompt input grow automatically [#53594](https://github.com/sourcegraph/sourcegraph/pull/53594)

### Changed

- Fixed logging to use JetBrains api + other minor fixes [#54579](https://github.com/sourcegraph/sourcegraph/pull/54579)
- Enabled editor recipe context menu items when working with Cody app only when Cody app is
  running [#54583](https://github.com/sourcegraph/sourcegraph/pull/54583)
- Renamed `completion` to `autocomplete` in both the UI and
  code [#54606](https://github.com/sourcegraph/sourcegraph/pull/54606)
- Increased minimum rows of prompt input form 2 to 3 [#54733](https://github.com/sourcegraph/sourcegraph/pull/54733)
- Improved completion prompt with changes from the VS Code
  plugin [#54668](https://github.com/sourcegraph/sourcegraph/pull/54668)
- Displayed more informative message when no context has been
  found [#54480](https://github.com/sourcegraph/sourcegraph/pull/54480)

### Fixed

- Now avoiding NullPointerException in an edge case when the chat doesn't
  exist [#54785](https://github.com/sourcegraph/sourcegraph/pull/54785)

## [3.0.3]

### Added

- Added recipes to editor context menu [#54430](https://github.com/sourcegraph/sourcegraph/pull/54430)
- Figure out default repository when no files are opened in the
  editor [#54476](https://github.com/sourcegraph/sourcegraph/pull/54476)
- Added `unstable-codegen` completions support [#54435](https://github.com/sourcegraph/sourcegraph/pull/54435)

### Changed

- Use smaller Cody logo in toolbar and editor context
  menu [#54481](https://github.com/sourcegraph/sourcegraph/pull/54481)
- Sourcegraph link sharing and opening file in browser actions are disabled when working with Cody
  app [#54473](https://github.com/sourcegraph/sourcegraph/pull/54473)

### Fixed

- Preserve new lines in the human chat message [#54417](https://github.com/sourcegraph/sourcegraph/pull/54417)
- JetBrains: Handle response == null case when checking for
  embeddings [#54492](https://github.com/sourcegraph/sourcegraph/pull/54492)

## [3.0.2]

### Fixed

- Repositories with http/https remotes are now available for
  Cody [#54372](https://github.com/sourcegraph/sourcegraph/pull/54372)

## [3.0.1]

### Changed

- Sending message on Enter rather than Ctrl/Cmd+Enter [#54331](https://github.com/sourcegraph/sourcegraph/pull/54331)
- Updated name to Cody AI app [#54360](https://github.com/sourcegraph/sourcegraph/pull/54360)

### Removed

- Sourcegraph CLI's SRC_ENDPOINT and SRC_ACCESS_TOKEN env variables overrides for the local config got
  removed [#54369](https://github.com/sourcegraph/sourcegraph/pull/54369)

### Fixed

- telemetry is now being sent to both the current instance & dotcom (unless the current instance is dotcom, then just
  that) [#54347](https://github.com/sourcegraph/sourcegraph/pull/54347)
- Don't display doubled messages about the error when trying to load
  context [#54345](https://github.com/sourcegraph/sourcegraph/pull/54345)
- Now handling Null error messages in error logging
  properly [#54351](https://github.com/sourcegraph/sourcegraph/pull/54351)
- Made sidebar refresh work for non-internal builds [#54348](https://github.com/sourcegraph/sourcegraph/pull/54358)
- Don't display duplicated files in the "Read" section in the
  chat [#54363](https://github.com/sourcegraph/sourcegraph/pull/54363)
- Repositories without configured git remotes are now available for
  Cody [#54370](https://github.com/sourcegraph/sourcegraph/pull/54370)
- Repositories with http/https remotes are now available for
  Cody [#54372](https://github.com/sourcegraph/sourcegraph/pull/54372)

## [3.0.0]

### Added

- Background color and font of inline code blocks differs from regular text in
  message [#53761](https://github.com/sourcegraph/sourcegraph/pull/53761)
- Autofocus Cody chat prompt input [#53836](https://github.com/sourcegraph/sourcegraph/pull/53836)
- Basic integration with the local Cody App [#54061](https://github.com/sourcegraph/sourcegraph/pull/54061)
- Background color and font of inline code blocks differs from regular text in
  message [#53761](https://github.com/sourcegraph/sourcegraph/pull/53761)
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
- A separate setting for enabling/disabling Cody
  completions. [pull/53302](https://github.com/sourcegraph/sourcegraph/pull/53302)
- Debounce for inline Cody completions [pull/53447](https://github.com/sourcegraph/sourcegraph/pull/53447)
- Enabled "Translate to different language" recipe [#53393](https://github.com/sourcegraph/sourcegraph/pull/53393)
- Skip Cody completions if there is code in line suffix or in the middle of a word in
  prefix [#53476](https://github.com/sourcegraph/sourcegraph/pull/53476)
- Enabled "Summarize recent code changes" recipe [#53534](https://github.com/sourcegraph/sourcegraph/pull/53534)

### Changed

- Convert `\t` to spaces in leading whitespace for autocomplete suggestions (according to
  settings) [#53743](https://github.com/sourcegraph/sourcegraph/pull/53743)
- Disabled line highlighting in code blocks in chat [#53829](https://github.com/sourcegraph/sourcegraph/pull/53829)
- Parallelized completion API calls and reduced debounce down to
  20ms [#53592](https://github.com/sourcegraph/sourcegraph/pull/53592)

### Fixed

- Fixed the y position at which autocomplete suggestions are
  rendered [#53677](https://github.com/sourcegraph/sourcegraph/pull/53677)
- Fixed rendered completions being cleared after disabling them in
  settings [#53758](https://github.com/sourcegraph/sourcegraph/pull/53758)
- Wrap long words in the chat message [#54244](https://github.com/sourcegraph/sourcegraph/pull/54244)
- Reset conversation button re-enables "Send"
  button [#53669](https://github.com/sourcegraph/sourcegraph/pull/53669)
- Fixed font on the chat ui [#53540](https://github.com/sourcegraph/sourcegraph/pull/53540)
- Fixed line breaks in the chat ui [#53543](https://github.com/sourcegraph/sourcegraph/pull/53543)
- Reset prompt input on message send [#53543](https://github.com/sourcegraph/sourcegraph/pull/53543)
- Fixed UI of the prompt input [#53548](https://github.com/sourcegraph/sourcegraph/pull/53548)
- Fixed zero-width spaces popping up in inline
  autocomplete [#53599](https://github.com/sourcegraph/sourcegraph/pull/53599)
- Reset conversation button re-enables "Send" button [#53669](https://github.com/sourcegraph/sourcegraph/pull/53669)
- Fixed displaying message about invalid access token on any 401 error from
  backend [#53674](https://github.com/sourcegraph/sourcegraph/pull/53674)

## [3.0.0-alpha.9]

### Added

- Background color and font of inline code blocks differs from regular text in
  message [#53761](https://github.com/sourcegraph/sourcegraph/pull/53761)
- Autofocus Cody chat prompt input [#53836](https://github.com/sourcegraph/sourcegraph/pull/53836)
- Basic integration with the local Cody App [#54061](https://github.com/sourcegraph/sourcegraph/pull/54061)
- Onboarding of the user when using local Cody App [#54298](https://github.com/sourcegraph/sourcegraph/pull/54298)

## [3.0.0-alpha.7]

### Added

- Background color and font of inline code blocks differs from regular text in
  message [#53761](https://github.com/sourcegraph/sourcegraph/pull/53761)
- Autofocus Cody chat prompt input [#53836](https://github.com/sourcegraph/sourcegraph/pull/53836)
- Cody Agent [#53370](https://github.com/sourcegraph/sourcegraph/pull/53370)

### Changed

- Convert `\t` to spaces in leading whitespace for autocomplete suggestions (according to
  settings) [#53743](https://github.com/sourcegraph/sourcegraph/pull/53743)
- Disabled line highlighting in code blocks in chat [#53829](https://github.com/sourcegraph/sourcegraph/pull/53829)

### Fixed

- Fixed the y position at which autocomplete suggestions are
  rendered [#53677](https://github.com/sourcegraph/sourcegraph/pull/53677)
- Fixed rendered completions being cleared after disabling them in
  settings [#53758](https://github.com/sourcegraph/sourcegraph/pull/53758)
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
- A separate setting for enabling/disabling Cody
  completions. [pull/53302](https://github.com/sourcegraph/sourcegraph/pull/53302)
- Debounce for inline Cody completions [pull/53447](https://github.com/sourcegraph/sourcegraph/pull/53447)
- Enabled "Translate to different language" recipe [#53393](https://github.com/sourcegraph/sourcegraph/pull/53393)
- Skip Cody completions if there is code in line suffix or in the middle of a word in
  prefix [#53476](https://github.com/sourcegraph/sourcegraph/pull/53476)
- Enabled "Summarize recent code changes" recipe [#53534](https://github.com/sourcegraph/sourcegraph/pull/53534)

### Changed

- Parallelized completion API calls and reduced debounce down to
  20ms [#53592](https://github.com/sourcegraph/sourcegraph/pull/53592)

### Fixed

- Fixed font on the chat ui [#53540](https://github.com/sourcegraph/sourcegraph/pull/53540)
- Fixed line breaks in the chat ui [#53543](https://github.com/sourcegraph/sourcegraph/pull/53543)
- Reset prompt input on message send [#53543](https://github.com/sourcegraph/sourcegraph/pull/53543)
- Fixed UI of the prompt input [#53548](https://github.com/sourcegraph/sourcegraph/pull/53548)
- Fixed zero-width spaces popping up in inline
  autocomplete [#53599](https://github.com/sourcegraph/sourcegraph/pull/53599)
- Reset conversation button re-enables "Send" button [#53669](https://github.com/sourcegraph/sourcegraph/pull/53669)
- Fixed displaying message about invalid access token on any 401 error from
  backend [#53674](https://github.com/sourcegraph/sourcegraph/pull/53674)

## [3.0.0-alpha.1]

### Added

- Alpha-quality Cody chat, not ready yet for internal dogfooding.
- Alpha-quality Cody code completions, not ready yet for internal dogfooding.

## [2.1.4]

### Added

- Add `extensionDetails` to `public_argument` on logger [#51321](https://github.com/sourcegraph/sourcegraph/pull/51321)

### Fixed

- Handle case when remote for local branch != sourcegraph
  remote [#52172](https://github.com/sourcegraph/sourcegraph/pull/52172)

## [2.1.3]

### Added

- Compatibility with IntelliJ 2023.1

### Fixed

- Fixed a backward-compatibility issue with Sourcegraph versions prior to
  4.3 [#50080](https://github.com/sourcegraph/sourcegraph/issues/50080)

## [2.1.2]

### Added

- Compatibility with IntelliJ 2022.3

## [2.1.1]

### Added

- Now the name of the remote can contain slashes

### Fixed

- “Open in Browser” and “Copy Link” features now open the correct branch when it exists on the
  remote. [pull/44739](https://github.com/sourcegraph/sourcegraph/pull/44739)
- Fixed a bug where if the tracked branch had a different name from the local branch, the local branch name was used in
  the URL, incorrectly

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

- Removed tracking parameters from all shareable
  URLs [pull/42022](https://github.com/sourcegraph/sourcegraph/pull/42022)

### Fixed

- Remove pointer cursor in the web view. [pull/41845](https://github.com/sourcegraph/sourcegraph/pull/41845)
- Updated “Learn more” URL to link the blog post in the update
  notification [pull/41846](https://github.com/sourcegraph/sourcegraph/pull/41846)
- Made the plugin compatible with versions 3.42.0 and
  below [pull/42105](https://github.com/sourcegraph/sourcegraph/pull/42105)

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

- It is now possible to configure the plugin per-repository using a `.idea/sourcegraph.xml` file. See the README for
  details.
- Special thanks: @oliviernotteghem for contributing the new features in this release!
- Fixed bugs where Open in Sourcegraph from the git menu does not work for repos with ssh url as their remote url

## [1.2.1] - Open Revision in Sourcegraph

- Added "Open In Sourcegraph" action to VCS History and Git Log to open a revision in the Sourcegraph diff view.
- Added "defaultBranch" configuration option that allows opening files in a specific branch on Sourcegraph.
- Added "remoteUrlReplacements" configuration option that allow users to replace specified values in the remote url with
  new strings.

## [1.2.0] - Copy link to file, search in repository, per-repository configuration, bug fixes & more

- The search menu entry is now no longer present when no text has been selected.
- When on a branch that does not exist remotely, `master` will now be used instead.
- Menu entries (Open file, etc.) are now under a Sourcegraph sub-menu.
- Added a "Copy link to file" action (alt+c / opt+c).
- Added a "Search in repository" action (alt+r / opt+r).
- It is now possible to configure the plugin per-repository using a `.idea/sourcegraph.xml` file. See the README for
  details.
- Special thanks: @oliviernotteghem for contributing the new features in this release!

## [1.1.2] - Minor bug fixes around searching.

- Fixed an error that occurred when trying to search with no selection.
- The git remote used for repository detection is now `sourcegraph` and then `origin`, instead of the previously poor
  choice of just the first git remote.

## [1.1.1] - Fixed search shortcut

- Updated the search URL to reflect a recent Sourcegraph.com change.

## [1.1.0] - Configurable Sourcegraph URL

- Added support for using the plugin with on-premises Sourcegraph instances.

## [1.0.0] - Initial Release

- Basic Open File & Search functionality.
