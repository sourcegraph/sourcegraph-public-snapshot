# Telemetry

> NOTE: This document is a work-in-progress.

Telemetry describes the logging of user events, such as a page view or search. Telemetry data is collected by each Sourcegraph instance and is not sent to Sourcegraph.com (except in aggregate form as documented in "[Pings](../../admin/pings.md)").

## Action telemetry

An [action](../../extensions/authoring/contributions.md#actions) consists of an action ID (such as `linecounter.displayLineCountAction`), a command to invoke, and other presentation information. An action is executed when the user clicks on a button or command palette menu item for the action. Actions may be defined either by extensions or by the main application (these are called "builtin actions"). For telemetry purposes, the two types of actions are treated identically.

A telemetry event is logged for every action that is executed by the user. By default, only the action ID is logged (not the arguments). To include the arguments for an action in the telemetry event data, the caller will need to explicitly opt-in (for privacy/security). Action telemetry is implemented in the `ActionItem` React component.

Examples of builtin actions include `goToDefinition`, `goToDefinition.preloaded`, `findReferences`, and (soon) `toggleLineWrap`, `goToPermalink`, `toggleColorTheme`, etc. (i.e., the builtin buttons in the file header and global nav).

Examples of extension actions include [`git.blame.toggle` from sourcegraph/git-extras](https://sourcegraph.com/extensions/sourcegraph/git-extras/-/contributions). An extension's action IDs are specified in its extension manifest and can be viewed on the **Contributions** tab of its extension registry listing.

## Browser extension telemetry (not native integration)

#### Action telemetry

Browser extension telemetry data is sent only to the connected Sourcegraph instance URL (except in aggregate form as documented in "[Pings](../../admin/pings.md)"). 

**The telemetry is enabled if one of the following conditions are valid:**
  - if NOT in Firefox. TODO: clarify why?
  - if connected to self-hosted Sourcegraph instance URL
  - if `Send telemetry` is enabled in OptionPopupPage

**All browser extension events are triggered with:**
- `source`: `CODEHOSTINTEGRATION`
- `anonymizedUserId`: generated user ID, stored in browser local storage
- `sourcegraphURL`: connected/configured SourcegraphURL
- `platform`: detected platform name

**Following events are triggered from the browser extension:**
- `hover`: when successfully showing non-empty/non-error hover information
- `action.id`: Sourcegraph extension action ID
  - when clicking any action from the command palette
  - or when clicking from the code view toolbar (e.g., open in VsCode)

  - > NOTE: Seems like we don't pass telemetryService to Sourcegraph extensions in Bext as we do in Web.

#### UTM markers
- `utm_source={platform-name}&utm_campaign=global-search`: Github enhanced search feature (item in search dropdown)
- `utm_source=omnibox`: Browser omnibox (a.k.a 'src'). Note: Supported in Chrome and Firefox
- `/sign-in?close=true&utm_source={platform-name}`: SignIn button instead of ViewOnSourcegraph when connected to self-hosted instance. TODO: Clarify if signIn button still works properly
- `?utm_source={platform-name}`: "View File In Sourcegraph" button in file editor toolbar
- `?utm_source={platform-name}`: "View Diff In Sourcegraph" button in file editor toolbar
- `?utm_source={platform-name}`: "View Repo In Sourcegraph" button on the top


#### Error logging (Sentry)

All errors in background/content scripts are automatically logged by Sentry if `Allow Error Reporting` is enabled.
> Note: stack trace might include sensitive information stored in variables.
