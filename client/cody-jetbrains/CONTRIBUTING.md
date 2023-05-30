# Contributing to Cody JetBrains Extension

Thank you for your interest in contributing to Cody!
The goal of this document is to provide a high-level overview of how you can contribute to the Cody JetBrains Extension.
Please refer to our [main CONTRIBUTING](https://github.com/sourcegraph/sourcegraph/blob/main/CONTRIBUTING.md) docs for general information regarding contributing to any Cody feature.

## License

Apache

## Feedback

Your feedback is important to us and is greatly appreciated. Please do not hesitate to submit your ideas or suggestions about how we can improve the extension to our [JetBrains Plugin Feedback Discussion Thread](https://github.com/sourcegraph/sourcegraph/discussions/51210) on GitHub.

## Issues / Bugs

New issues and feature requests can be filed through our [issue tracker](https://github.com/sourcegraph/sourcegraph/issues/new?labels=team/cody,cody/jetbrains&title=Cody:+) using the `cody/jetbrains` & `team/cody` labels.

## Development

- Clone `https://github.com/sourcegraph/sourcegraph` (on Windows, you'll need to use WSL2)
- Go to `client/cody-jetbrains/` and run the plugin in a sandboxed IDE by running `./gradlew :runIde`. This will start the platform with the versions defined in `gradle.properties`, [here](https://github.com/sourcegraph/sourcegraph/blob/main/client/cody-jetbrains/gradle.properties#L14-L16).
- Build a deployable plugin artifact by running `./gradlew buildPlugin`. The output file is `build/distributions/Cody.zip`.
- Reformat the codebase with `./gradlew spotlessApply`.
- Install the google-java-format plugin
  https://plugins.jetbrains.com/plugin/8527-google-java-format and configure
  IntelliJ's file save actions to format.
- Ensure `src login` is logged into your sourcegraph.com account. This avoids
  the need to manually configure the access token in the UI every time you run
  `./gradlew :runIde`.

## Publishing a new version

The publishing process is based on the [intellij-platform-plugin-template](https://github.com/JetBrains/intellij-platform-plugin-template).

### Publishing from your local machine

1. Update `pluginVersion` in `gradle.properties`
   - To create pre-release builds with the same version as a previous one, append `.{N}`.
     For example, `1.0.0-alpha`, then `1.0.0-alpha.1`, `1.0.0-alpha.2`, and so on.
2. Describe the changes in the `[Unreleased]` section of `client/cody-jetbrains/CHANGELOG.md` then remove any empty headers
3. Make sure `runIde` is not running
4. Commit your changes
5. Run `PUBLISH_TOKEN=<YOUR TOKEN HERE> ./scripts/release.sh` from inside the `client/cody-jetbrains` directory (You can [generate tokens on the JetBrains marketplace](https://plugins.jetbrains.com/author/me/tokens)).
6. Commit changes and create PR
