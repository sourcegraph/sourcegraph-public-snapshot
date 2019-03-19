# Sourcegraph browser extension

The [open-source](https://github.com/sourcegraph/sourcegraph/tree/master/client/browser) Sourcegraph
browser extension adds code intelligence to files and diffs on GitHub, GitHub
Enterprise, GitLab, Phabricator, and Bitbucket Server.

<p>
  <a target="_blank" href="https://chrome.google.com/webstore/detail/sourcegraph/dgjhfomjieaadpoljlnidmbgkdffpack" style="display:flex;align-items:center">
  <img src="img/chrome.svg" width="24" height="24" style="margin-right:5px" /> <strong>Install Sourcegraph for Chrome</strong>
  </a>
</p>

<p>
  <a target="_blank" href="https://storage.googleapis.com/sourcegraph-for-firefox/latest.xpi" style="display:flex;align-items:center">
  <img src="img/firefox.svg" width="24" height="24" style="margin-right:5px" /> <strong>Install Sourcegraph for Firefox</strong>
  </a>
</p>

![Sourcegraph browser extension](img/BrowserExtension.gif)

## Features

### Code intelligence

When you hover your mouse over code in files, diffs, pull requests, etc., the Sourcegraph extension displays a tooltip with:

- Documentation and the type signature for the hovered token
- **Go to definition** button
- **Find references** button

### Search shortcut in location bar

The Sourcegraph extension adds a search engine shortcut to your web browser that performs a search on your Sourcegraph instance. After you've installed it (see above), use the search shortcut it provides to perform a search:

1.  In the Chrome or Firefox location bar, type <kbd>src</kbd> <kbd>Space</kbd>.
1.  Start typing your search query.
1.  Select an instant search suggestion or press <kbd>Enter</kbd> to see all results.

To install this search engine shortcut manually, and for more information, see "[Browser search engine shortcuts](browser_search_engine.md)".

## Configuration

### Configuring the Sourcegraph instance to use

By default, the browser extension communicates with [Sourcegraph.com](https://sourcegraph.com), which has only public code.

To use the browser extension with a different Sourcegraph instance:

1.  Click the Sourcegraph extension icon in the browser toolbar to open the settings page.
1.  Click **Update** and enter the URL of a Sourcegraph instance (such as `https://sourcegraph.example.com` or `https://sourcegraph.com`).
1.  Click **Save**.

### Enabling the browser extension on your code host

By default, the Sourcegraph browser extension will only provide code intelligence on [github.com](https://github.com/). It needs additional permissions in order to run on other code hosts.

To grant these permissions:

1.  Navigate to any page on your code host.
1.  Right-click the Sourcegraph icon in the browser extension toolbar.
1.  Click "Enable Sourcegraph on this domain".
1.  Click "Allow" in the permissions request popup.

### Troubleshooting

The most common problem is:

#### No code intelligence or buttons ("View repository", "View file", etc.) are displayed on the code host.

Try the following:

1.  Click the Sourcegraph extension icon in the browser toolbar to open the settings page.
    - Ensure that the Sourcegraph URL is correct. It must point to your own Sourcegraph instance to work on private code.
    - Check whether any permissions must be granted. If so, the settings page will offer you to "grant the Sourcegraph browser extension additional permissions".
1. On some code hosts, you need to be signed in (to the code host) to use the browser extension. Try signing in.

## Privacy

Sourcegraph integrations never send any logs, pings, usage statistics, or telemetry to Sourcegraph.com. They will only connect to Sourcegraph.com as required to provide code intelligence or other functionality on public code. As a result, no private code, private repository names, usernames, or any other specific data is sent to Sourcegraph.com.

If connected to a private, self-hosted Sourcegraph instance, Sourcegraph browser extensions will send notifications of usage to that private Sourcegraph instance only. This allows the site admins to see usage statistics.
