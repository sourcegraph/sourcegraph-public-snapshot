# Contributing to Sourcegraph JetBrains Extension

Thank you for your interest in contributing to Sourcegraph!
The goal of this document is to provide a high-level overview of how you can contribute to the Sourcegraph JetBrains Extension.
Please refer to our [main CONTRIBUTING](https://github.com/sourcegraph/sourcegraph/blob/main/CONTRIBUTING.md) docs for general information regarding contributing to any Sourcegraph feature.

## License

Apache

## Feedback

Your feedback is important to us and is greatly appreciated. Please do not hesitate to submit your ideas or suggestions about how we can improve the extension to our [JetBrains Plugin Feedback Discussion Thread](https://github.com/sourcegraph/sourcegraph/discussions/43930) on GitHub.

## Issues / Bugs

New issues and feature requests can be filed through our [issue tracker](https://github.com/sourcegraph/sourcegraph/issues/new?labels=team/integrations,jetbrains-ide&title=JetBrains:+&projects=Integrations%20Project%20Board) using the `jetbrains-ide` & `team/integrations` labels.

## Development

- Clone `https://github.com/sourcegraph/sourcegraph` (on Windows, you'll need to use WSL2)
- Run `pnpm install` in the root directory to get all dependencies
- Run `pnpm generate` in the root directory to generate graphql files
- Go to `client/jetbrains/` and run `pnpm build` to generate the JS files, or `pnpm watch` to watch for changes and regenerate on the fly
- You can test the “Find with Sourcegraph” window by running `pnpm standalone` in the `client/jetbrains/` directory and opening [http://localhost:3000/](http://localhost:3000/) in your browser.
- Run the plugin in a sandboxed IDE by running `./gradlew runIde`. This will start the platform with the versions defined in `gradle.properties`, [here](https://github.com/sourcegraph/sourcegraph/blob/main/client/jetbrains/gradle.properties#L14-L16).
  - Note: 2021.3 or later is required for Macs with Apple Silicon chips.
- Build a deployable plugin artifact by running `./gradlew buildPlugin`. The output file is `build/distributions/Sourcegraph.zip`.

## Publishing a new version

The publishing process is based on the [intellij-platform-plugin-template](https://github.com/JetBrains/intellij-platform-plugin-template).

### Publishing from your local machine

1. Update `pluginVersion` in `gradle.properties`

- To create pre-release builds with the same version as a previous one, append `.{N}`. For example, `1.0.0-alpha`,
  then `1.0.0-alpha.1`, `1.0.0-alpha.2`, and so on.

2. Describe the changes in the `[Unreleased]` section of `client/jetbrains/CHANGELOG.md` then remove any empty headers
3. Go through
   the [manual test cases](https://docs.sourcegraph.com/integration/jetbrains/manual_testing)
4. Make sure `runIde` is not running
5. Commit your changes
6. Run `PUBLISH_TOKEN=<YOUR TOKEN HERE> ./scripts/release.sh` from inside the `client/jetbrains` directory (You
   can [generate tokens on the JetBrains marketplace](https://plugins.jetbrains.com/author/me/tokens)).
7. Commit changes and create PR

## Retrying a release

It happened in the past that we had compatibility issues and the version got rejected.
Here is what to do in this case:

1. Go to the [versions](https://plugins.jetbrains.com/plugin/9682-sourcegraph/versions/) page logged in with a JetBrains
   plugin admin account
2. Go to the latest (failed) version, and click the Trash icon to delete it.
3. Fix the problem in the code
4. **Important:** Don't forget to revert the info in CHANGELOG.md to the pre-release state. Then you'll need to commit.
5. Publish the version again
6. (Here, of course, you'll need to commit CHANGELOG.md again)
7. You don't need to wait for JetBrains' email—compatibility checks are visible after a few minutes on the same page.
8. If the version is still rejected, repeat the process.

## Enabling web view debugging

Parts of this extension rely on the [JCEF](https://plugins.jetbrains.com/docs/intellij/jcef.html) web view features
built into the JetBrains platform. To enable debugging tools for this view, please follow these steps:

1. [Enable JetBrains internal mode](https://plugins.jetbrains.com/docs/intellij/enabling-internal.html)
2. Open Find Actions: (<kbd>Ctrl+Shift+A</kbd> / <kbd>⌘⇧A</kbd>)
3. Search for "Registry..." and open it
4. Find option `ide.browser.jcef.debug.port`
5. Change the default value to an open port (we use `9222`)
6. Restart IDE
7. Open the “Find with Sourcegraph” window (<kbd>Alt+A</kbd> / <kbd>⌥A</kbd>)
8. Switch to a browser window, go to [`localhost:9222`](http://localhost:9222), and select the Sourcegraph window. Sometimes it needs some back and forth to focus the external browser with the JCEF component also focused—you may need to move the popup out of the way and click the external browser rather than using <kbd>Alt+Tab</kbd> / <kbd>⌘Tab</kbd>.
