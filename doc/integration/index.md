# Integrations

Sourcegraph integrates with your other tools so you can search and browse code efficiently from any part of your workflow:

- [Browser extension](browser_extension.md): go-to-definitions and hovers in your code host and code reviews
- [Editor plugins](editor.md): jump to Sourcegraph from your editor
- [Search shortcuts](browser_search_engine.md): quickly search from your browser
- [GraphQL API](../api/graphql/index.md)

![GitHub pull request integration](img/GitHubDiff.png.md)

## Privacy

Sourcegraph integrations never send any logs, pings, usage statistics, or telemetry to Sourcegraph.com. They will only connect to Sourcegraph.com as required to provide code intelligence or other functionality on public code. As a result, no private code, private repository names, usernames, or any other specific data is sent to Sourcegraph.com.

If connected to a private, self-hosted Sourcegraph instance, Sourcegraph browser extensions will send notifications of usage to that private Sourcegraph instance only. This allows the site admins to see usage statistics.
