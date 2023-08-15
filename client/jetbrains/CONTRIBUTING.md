# Contributing to Sourcegraph JetBrains Extension

Thank you for your interest in contributing to Sourcegraph!
The goal of this document is to provide a high-level overview of how you can contribute to the Sourcegraph JetBrains Extension.
Please refer to our [main CONTRIBUTING](https://github.com/sourcegraph/sourcegraph/blob/main/CONTRIBUTING.md) docs for general information regarding contributing to any Sourcegraph feature.

## License

Apache

## Feedback

Your feedback is important to us and is greatly appreciated. Please do not hesitate to submit your ideas or suggestions about how we can improve the extension to our [JetBrains Plugin Feedback Discussion Thread](https://github.com/sourcegraph/sourcegraph/discussions/43930) on GitHub.

## Issues / Bugs

New issues and feature requests can be filed through our [issue tracker](https://github.com/sourcegraph/sourcegraph/issues/new?labels=team/integrations,jetbrains-ide&title=JetBrains:+) using the `jetbrains-ide` & `team/integrations` labels.

## Development

- Clone `https://github.com/sourcegraph/sourcegraph` (on Windows, you'll need to use WSL2)
- Run `pnpm install` in the root directory to get all dependencies
- Run `pnpm generate` in the root directory to generate graphql files
- Go to `client/jetbrains/` and run `pnpm build` to generate the JS files, or `pnpm watch` to watch for changes and regenerate on the fly
- You can test the “Find with Sourcegraph” window by running `pnpm standalone` in the `client/jetbrains/` directory and opening [http://localhost:3000/](http://localhost:3000/) in your browser.
- Make sure you have Java 11 installed. Two ways to do that:
  1. CLI: [SDKMAN!](https://github.com/sdkman/homebrew-tap):
     - `brew tap sdkman/tap`
     - `brew install sdkman-cli`
     - Add this to your `.zprofile`:
       ```sh
       export SDKMAN_DIR=$(brew --prefix sdkman-cli)/libexec
       [[ -s "${SDKMAN_DIR}/bin/sdkman-init.sh" ]] && source "${SDKMAN_DIR}/bin/sdkman-init.sh"
       ```
     - Try it with `sdk version` in a new terminal. It should work.
     - Run `sdk install java 11.0.20-amzn`
  2. GUI:
     - Open your clone of the repo in IntelliJ
     - Go to Project Structure (`⌘;`) | Platform Settings | SDK | Plus sign | Download JDK... | set version to 11 | pick Amazon Corretto aarch64
- Run the plugin in a sandboxed IDE. Two ways to do that:
  1. CLI: run `./gradlew :runIde`. This will start the platform with the versions defined in `gradle.properties`, [here](https://github.com/sourcegraph/sourcegraph/blob/main/client/jetbrains/gradle.properties#L14-L16).
  2. Run | Run... (`⌃⌥R`) | Edit Configurations... | Plus sign | Gradle | set Tasks to `runIde` | set Gradle project to `jetbrains` | name it `runIde` | OK | Run it with the green play button at the top right of the IDE.
  - Note: IntelliJ version 2021.3 or later is required for Macs with Apple Silicon chips.
- Build a deployable plugin artifact by running `./gradlew buildPlugin`. The output file is `build/distributions/Sourcegraph.zip`.
- Reformat the codebase with `./gradlew spotlessApply`.
- Install the google-java-format plugin
  https://plugins.jetbrains.com/plugin/8527-google-java-format and configure
  IntelliJ's file save actions to format.
- Set the environment variable `CODY_COMPLETIONS_ENABLED=true` to enable inline code completions.
- Ensure `src login` is logged into your sourcegraph.com account. This avoids
  the need to manually configure the access token in the UI every time you run
  `./gradlew :runIde`.
- If you are using an M1 MacBook and get a JCEF-related error using the "Find with Sourcegraph" command, try
  running `./gradlew -PplatformVersion=221.5080.210 :runIde` instead.
  See https://youtrack.jetbrains.com/issue/IDEA-291946 for more details.
- To debug communication between the IntelliJ plugin and Cody agent, it's useful to keep an open terminal tab that's
  running the command `fail -f build/sourcegraph/cody-agent-trace.json`.
- The Cody agent is a JSON-RPC server that implements the prompt logic for Cody. The JetBrains plugin needs access to the
  agent binary to function properly. This agent binary is automatically built from source if it does not exist. To
  speed up edit/test/debug feedback loops, the agent binary does not get rebuilt unless you provide the
  `-PforceAgentBuild=true` flag when running Gradle. For example, `./gradlew :runIde -PforceAgentBuild=true`.
- The Cody agent is disabled by default for local `./gradlew :runIde` task only.
  Use the `-PenableAgent=true` property to enable the Cody agent. For example, `./gradlew :runIde -PenableAgent=true`.
  When the agent is disabled, the plugin falls back to the non-agent-based implementation.

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
