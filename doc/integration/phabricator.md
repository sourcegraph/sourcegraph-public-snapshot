# Phabricator integration with Sourcegraph

## Linking and syncing Phabricator repositories

If you mirror your source repositories on Phabricator, Sourcegraph can provide users with links to various Phabricator pages.

The `phabricator` configuration option takes in an array of Phabricator configurations. A Phabricator configuration consists of the following fields:

- `url` field that maps to the url of the Phabricator host
- `token` an optional Conduit API token, which you may generate from the Phabricator web interface. The token is used to fetch the list of repos available on the Phabricator installation
- `repos` if your Phabricator installation mirrors repositories from a different origin than Sourcegraph, you must specify a list of repository `path`s (as displayed on Sourcegraph) and their corresponding Phabricator `callsign`s. For example: `[{ path: 'gitolite.example.org/foobar', callsign: 'FOO'}]`. _Note that the `callsign` is case sensitive._

At least one of token and repos should be provided.

For example:

```json
{
  // ...
  "phabricator": [
    {
      "url": "https://phabricator.example.com",
      "token": "api-abcdefghijklmnop",
      "repos": [{ "path": "gitolite.example.com/mux", "callsign": "MUX" }]
    }
  ]
  // ...
}
```

### Troubleshooting

If your outbound links to Phabricator are not present or not working, verify your Sourcegraph repository path matches the "normalized" URI output by Phabricator's `diffusion.repository.search` conduit API.

For example, if you have a repository on Sourcegraph whose URL is `https://sourcegraph.example.com/path/to/repo` then you should see a URI returned from `diffusion.repository.search` whose `normalized` field is `path/to/repo`. Check this by navigating to `$PHABRICATOR_URL/conduit/method/diffusion.repository.search/` and use the "Call Method" form with `attachments` field set to `{ "uris": true }` and `constraints` field set to `{ "callsigns": ["$CALLSIGN_FOR_REPO_ON_SOURCEGRAPH"]}`. In the generated output, verify that the first URI has a normalized path equal to `path/to/repo`.

## Browser extension

The [Sourcegraph browser extension](browser_extension.md) supports Phabricator. When installed in your web browser, it adds hover tooltips, go-to-definition, find-references, and code search to files and diffs viewed on Phabricator.

1.  Install the [Sourcegraph browser extension](browser_extension.md).
1.  [Configure the browser extension](browser_extension.md#configuring-the-sourcegraph-instance-to-use) to use your Sourcegraph instance (where you've added the `phabricator` site config property).
1.  Click the Sourcegraph icon in the browser toolbar to open the settings page. If a permissions notice is displayed, click **Grant permissions** to allow the browser extension to work on your Phabricator instance.
1.  Visit any file or diff on Phabricator. Hover over code or click the "View file" and "View repository" buttons.

## Phabricator extension

For production usage, we strongly recommend installing the Sourcegraph Phabricator extension for all users (so that each user doesn't need to install the browser extension individually). This involves adding a new extension to the extension directory of your Phabricator instance.

See the [phabricator-extension](https://github.com/sourcegraph/phabricator-extension) repository for installation instructions and configuration settings.
