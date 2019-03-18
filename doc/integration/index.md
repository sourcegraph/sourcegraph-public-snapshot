# Integrations

Sourcegraph integrates with code hosts and code review tools to power code intelligence on code files and code review diffs.

To get started, follow the documentation for [installing and configuring the browser extension](browser_extension.md).

![GitHub pull request integration](img/GitHubDiff.png.md)

## Other integrations

We also offer other integrations for Sourcegraph users, and an API to build on top of Sourcegraph:

- [Editor plugins](editor.md)
- [Browser extension](browser_extension.md)
- [Search shortcuts](browser_search_engine.md)
- [API](../api/graphql/index.md)

## Privacy

Sourcegraph integrations never send any logs, pings, usage statistics, or telemetry to Sourcegraph.com. They will only connect to Sourcegraph.com as required to provide code intelligence or other functionality on public code. As a result, no private code, private repository names, usernames, or any other specific data is sent to Sourcegraph.com.

If connected to a private, self-hosted Sourcegraph instance, Sourcegraph browser extensions will send notifications of usage to that private Sourcegraph instance only. This allows the site admins to see usage statistics.
