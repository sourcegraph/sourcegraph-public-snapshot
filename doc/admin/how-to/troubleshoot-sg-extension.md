# How to troubleshoot a Sourcegraph extension

This guide gives specific instructions for troubleshooting [extensions](https://docs.sourcegraph.com/extensions) developed by Sourcegraph.

## FAQs

#### How do I know if a Sourcegraph extension is running?

1. Right click on the Sourcegraph website and click Inspect (Chrome/Firefox) / Inspect Element (Safari) to open `Developer Tools`
2. You should see a console message `Activating Sourcegraph extension:` follows by the names of all the running extensions in the `Console` tab
3. If you don't see the expected extension running, please go to the User Menu on your Sourcegraph instance and click on Extensions to make sure the extension is enabled

#### A Sourcegraph extension is not working, what should I do?

1. First of all, please make sure the extension in question is running by following the steps from above
2. Look for error messages in your browser's `Developer Console`
3. Look for error messages in your browser's  `Network panel`

#### Why is the extension icon on my sidebar is shown as inactive / greyed out?

This happens if the extension is disabled or if you're visiting a page where an extension can't have any actions. For example, Open-in-Editor extensions do not work on top level folders [(example)](https://sourcegraph.com/github.com/sourcegraph/sourcegraph) because you cannot open a repo.


#### How do I upgrade the extensions in my private extension registry?

You can upgrade an extension in your private extension registry by simpily running the same `src extensions copy -extension-id=... -current-user=...` command as you would when you first [publish the extension](https://docs.sourcegraph.com/admin/extensions#publish-extensions-to-a-private-extension-registry).


#### What does it mean when a red dot shows up on the Sourcegraph browser extension icon?

The red dot indicates that either the Sourcegraph URL entered is invalid, or you are currently on a private repository. Visit our [browser extension docs](https://docs.sourcegraph.com/integration/browser_extension#make-it-work-for-private-code) for more information about enabling Sourcegraph to work with private repositories.

## Extension Specific

### VS Code Extension

#### Unsupported features by Sourcegraph version

Here is a list of known limitations to the VS Code Extension that we are looking into for future releases:

1. Only work with instances that support stream search
2. Search does not work across instances on version 3.31.x
3. Searches performed within the extension are not logged in Cloud for instance below version 3.34.0
4. Search context are not fetched correctly for version below 3.36.0 (fixed in v2.0.9)
5. Web extension supports instances using 3.36.0+ officially.

#### How to use the VS Code Extension with your private Sourcegraph instance
The extension is connected to the [Sourcegraph public instance](https://sourcegraph.com/) by default. You can also add the following settings in your [VS Code User Setting](https://code.visualstudio.com/docs/getstarted/settings#_settings-file-locations) to connect the extension to your private instance: 

1. `sourcegraph.url`: the instance url of your private instance 

2. `sourcegraph.accessToken`: an access token created by your private Sourcegraph instance

Note: If only an access token is configured, the extension will try to run searches on our public instance using the token instead of the corresponding instance. 

#### How to update the Sourcegraph VS Code Extension to the latest version
![image](https://user-images.githubusercontent.com/68532117/153280003-df575725-22c2-4a5a-b94b-2137790da039.png)
1. Search for `Sourcegraph` in your VS Code Extensions Marketplace. 
2. From there you can check if an update is available for the extension.
3. The version number next to the extension name indicates the version that you are currently running.

#### Sign-up Banner remains visible when a valid access token has been provided in V2

A fix has been implemented in v2.0.6. Please update your extension to the latest version and restart VS Code by clicking on `Code` > `Quit Visual Studio Code`. This is to restart VS Code using the updated version of the extension.

#### The `sourcegraph.defaultBranch` and `sourcegraph.remoteUrlReplacements`settings are not working in V2

A fix has been implemented in v2.0.7. Please update your extension to the latest version and restart VS Code by clicking on `Code` > `Quit Visual Studio Code`. This is to restart VS Code using the updated version of the extension.

#### Restarting vs Reloading VS Code
Please note that reloading VS Code does not have the same effect as restarting (`Code` > `Quit Visual Studio Code`) the application. You must restart the application after upgrading the extension for VS Code to run in the newest version.

#### Error: Could not register service workers: InvalidStateError: Failed to register a ServiceWorker: The document is in an invalid state.

This error message comes from VS Code. Restarting the editor should resolve the issue as reported by a VS Code user [here](https://github.com/microsoft/vscode/issues/128649).

#### Error: The connection was closed before your search was completed. This may be due to a problem with a firewall, VPN or proxy, or a failure with the Sourcegraph server.

1. It is possible that the provided Access Token is not valid for the instance that your VS Code is connected to. Please try updating both the url and access token in your [VS Code User Setting](https://code.visualstudio.com/docs/getstarted/settings#_settings-file-locations) to see if the issue persists.
2. If the issue persists, try connecting using a CORS proxy or turning off your VPN settings. 
3. Add custom headers using the `sourcegraph.requestHeaders` setting (added in v2.0.9) if a specific header is required to make connection to your private instance.
4. A CORS proxy is required to run the extension on VS Code Web for instances running in version below [3.26.0](https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/CHANGELOG.md#3-36-0). 

#### Unable to verify your access token for sourcegraph.example.com. Please try again with a new access token or restart VS Code if the instance URL has been updated.
This error can arise from the URL and/or access token not being applied properly when entered in the UI. 

Check `settings.json` by going to the extension's settings and then clicking “Edit in settings.json” under *Sourcegraph: Access Token*. Ensure that values for `sourcegraph.accessToken` and `sourcegraph.url` are present and correct.

If the error persists, please [contact us](mailto:support@sourcegraph.com). 

#### The search results are not displayed using my VS Code color

The extension currently supports the following VS Code Color Theme:

- Dark (Visual Studio)
- Light+ (default light)
- Light (Visual Studio)
- High Contrast
- Monokai
- Monokai Pro
- One Dark Pro
- Dracula
- Dracula Soft
- Atom One Dark
- Cobalt2
- Panda Syntax
- Night Owl
- Hack The Box
- Solarized Light
- Solarized Dark

### Sourcegraph Extensions

#### Sonarqube: Error fetching Sonarqube data: Error: Forbidden
1. Look for error messages in your browser's  `Network panel`
2. If there is an error message indicates that the cors-anywhere request was being denied and you are using `"sonarqube.corsAnywhereUrl": "https://cors-anywhere.herokuapp.com"` in your configuration, please visit [https://cors-anywhere.herokuapp.com/corsdemo](https://cors-anywhere.herokuapp.com/corsdemo) to opt-in for temporary access by clicking on the `Request temporary access to the demo server` button
3. Alternatively you may remove this configuration option to use the default Sourcegraph's CORS proxy

#### Git-extras: Git blame is not working even though it is displayed as enabled
The extension is running if you can locate the Author in the bottom status bar. The plugin has 3 modes which can be activated by clicking on the extension icon on the extension sidebar on the right:

1. All but status bar are hidden
2. Author will be shown on the selected line
3. Author will show up on all lines (where changes are made)

#### Git-extras: Git blame is slow to load
The extension is expected to work slow when there are issues with the gitserver (eg. running out of resources) because the extension is dependent on the gitserver.

#### ESlint: The extension is not working on an instance.
The ESlint extension requires the eslint.insight.repository and eslint.insight.step to be configured in either the global settings or in the user settings for each insight using ESLint.

#### Open-in-intellij: Sourcegraph fails to load a file when trying to open the file from intellij Plugin.
This is most likely due to the file being opened in a Sourcegraph instance that does not have access to your files. You must first configure the plugin in order to use it with your private instance. See the plugin docs for more information.

#### Search-export: Can I export search results?
1. You can export search results by enabling the [Sourcegraph search results CSV export extension](https://sourcegraph.com/extensions/sourcegraph/search-export)
2. Once it is enabled, you will find an `Export to CSV` button in the Search-Results page
 
#### Search-export: Network Error when downloading CSV

It's likely that the CSV file exceeds the browser's limit for data URI size. Users can limit the size of search match preview size through their user settings (see ["contributions"](https://sourcegraph.com/extensions/sourcegraph/search-export/-/contributions) for search-export). If decreasing the size of search match previews doesn't resolve the issue, users can decrease the amount of search results exported with the `count:` filter in their search query.

#### Search-export: The number of exported results does not match the number of results displayed on Sourcegraph

The Sourcegraph [Streaming API](../../api/stream_api/index.md) determines the number of results in the Sourcegraph UI. However, our Search-export extension runs a query on our GraphQL API and will only export the complete list of results if the search query includes the `count:all` keyword.

If one file has two matches, the file will be listed only once in the exported csv. Hence number of lines in csv Might be lower than number of results.


___

See the [Sourcegraph browser extension docs](https://docs.sourcegraph.com/integration/browser_extension#troubleshooting) for more troubleshooting tips.
