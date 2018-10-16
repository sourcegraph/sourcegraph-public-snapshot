# Sourcegraph browser extension

The [open-source](https://github.com/sourcegraph/browser-extensions) Sourcegraph
browser extension adds code intelligence to files and diffs on GitHub, GitHub
Enterprise, GitLab, Phabricator, and Bitbucket Server.

<a target="_blank" href="https://chrome.google.com/webstore/detail/sourcegraph/dgjhfomjieaadpoljlnidmbgkdffpack" style="display:flex;align-items:center">
<img src="img/chrome.svg" width="24" height="24" style="margin-right:5px" /> <strong>Sourcegraph for Chrome</strong>
</a>
<div style="margin-top:5px"/>
<a target="_blank" href="https://addons.mozilla.org/en-US/firefox/addon/sourcegraph/" style="display:flex;align-items:center">
<img src="img/firefox.svg" width="24" height="24" style="margin-right:5px" /> <strong>Sourcegraph for Firefox</strong>
</a>

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

## Configuring the Sourcegraph instance to use

By default, the browser extension communicates with [Sourcegraph.com](https://sourcegraph.com), which has only public code.

To use the browser extension with a different Sourcegraph instance:

1.  Click the Sourcegraph extension icon in the browser toolbar to open the settings page.
1.  Click **Update** and enter the URL of a Sourcegraph instance (such as `https://sourcegraph.example.com` or `https://sourcegraph.com`).
1.  Click **Save**.

> NOTE: The Sourcegraph instance's site admin must [update the `corsOrigin` site config property](../admin/site_config/index.md) to allow the extension to communicate with it from all of the code hosts and other sites it will be used on. For example:

```json
{
  // ...
  "corsOrigin":
    "https://github.example.com https://gitlab.example.com https://bitbucket.example.org https://phabricator.example.com"
  // ...
}
```

### Troubleshooting

The most common problem is:

#### No code intelligence or buttons ("View repository", "View file", etc.) are displayed on the code host.

Try the following:

1.  Click the Sourcegraph extension icon in the browser toolbar to open the settings page.
    - Ensure that the Sourcegraph URL is correct. It must point to your own Sourcegraph instance to work on private code.
    - Check whether any permissions must be granted. If so, the settings page will display an alert with a **Grant permissions** button.
    - Confirm with your Sourcegraph instance's site admin that the site config `corsOrigin` property contains the URL of the external site on which you are trying to use the browser extension.
