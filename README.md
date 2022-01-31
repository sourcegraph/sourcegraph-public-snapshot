<!-- Plugin description -->

# Sourcegraph for JetBrains IDEs [![JetBrains Plugin](https://img.shields.io/badge/JetBrains-Sourcegraph-green.svg)](https://plugins.jetbrains.com/plugin/9682-sourcegraph)

- Search snippets of code on Sourcegraph.
- Copy and share a link to code on Sourcegraph.
- Quickly go from files in your editor to Sourcegraph.
<!-- Plugin description end -->

The plugin works with all JetBrains IDEs including:

- IntelliJ IDEA
- IntelliJ IDEA Community Edition
- PhpStorm
- WebStorm
- PyCharm
- PyCharm Community Edition
- RubyMine
- AppCode
- CLion
- GoLand
- DataGrip
- Rider
- Android Studio

## Installation

- Select `IntelliJ IDEA` then `Preferences` (or use <kbd>⌘,</kbd>)
- Click `Plugins` in the left-hand pane.
- Choose `Browse repositories...`
- Search for `Sourcegraph` -> `Install`
- Restart your IDE if needed, then select some code and choose `Sourcegraph` in the right-click context menu to see actions and keyboard shortcuts.

## Configuring for use with a private Sourcegraph instance

The plugin is configurable _globally_ by creating a `.sourcegraph-jetbrains.properties` (or `sourcegraph-jetbrains.properties` pre-v1.2.2) in your home directory. For example, modify the following URL to match your on-premises Sourcegraph instance URL:

```
url = https://sourcegraph.example.com
defaultBranch = example-branch
remoteUrlReplacements = git.example.com, git-web.example.com
```

You may also choose to configure it _per repository_ using a `.idea/sourcegraph.xml` (or `idea/sourcegraph.xml` pre-v1.2.2) file in your repository like so:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<project version="4">
    <component name="Config">
        <option name="url" value="https://sourcegraph.example.com" />
        <option name="defaultBranch" value="example-branch" />
        <option name="remoteUrlReplacements" value="git.example.com, git-web.example.com" />
    </component>
</project>
```

By default, the plugin will use the `origin` git remote to determine which repository on Sourcegraph corresponds to your local repository. If your `origin` remote doesn't match Sourcegraph, you may instead configure a `sourcegraph` Git remote which will take priority.

## Questions & Feedback

Please file an issue: https://github.com/sourcegraph/sourcegraph-jetbrains/issues/new

## Uninstallation

- Select `IntelliJ IDEA` then `Preferences` (or use <kbd>⌘,</kbd>)
- Click `Plugins` in the left-hand pane.
- Search for `Sourcegraph` -> Right click -> `Uninstall` (or uncheck to disable)

## Development

- Start IntelliJ and choose `Check out from Version Control` -> `Git` -> `https://github.com/sourcegraph/sourcegraph-jetbrains`
- Develop as you would normally (hit Debug icon in top right of IntelliJ) or using gradlew commands:
  1. `./gradlew runIde` to run an IDE instance with sourcegraph plugin installed. This will start the platform with the versions defined in [`gradle.properties`](https://github.com/sourcegraph/sourcegraph-jetbrains/blob/main/gradle.properties#L14-L16). _Note: 2021.3 is required for M1 Macs._
  2. `./gradlew buildPlugin` to build plugin artifact (`build/distributions/Sourcegraph.zip`)

## Publishing a new version

The publishing process is based on the actions outlined in the [`intellij-platform-plugin-template`](https://github.com/JetBrains/intellij-platform-plugin-template).

1. Update `gradle.properties` and set the version number for this release (e.g. `1.2.3`).
2. Create a [new release](https://github.com/sourcegraph/sourcegraph-jetbrains/releases/new) on GitHub.
3. Pick the new version number as the git tag (e.g. `v1.2.3`).
4. Copy/paste the `[Unreleased]` section of the [`CHANGELOG.md`](https://github.com/sourcegraph/sourcegraph-jetbrains/blob/main/CHANGELOG.md) into the GitHub release text.
5. Once published, a GitHub action is triggered that will publish the release automatically and create a PR to update the changelog and version text. You may need to manually fix the content.

## Version History

See [`CHANGELOG.md`](https://github.com/sourcegraph/sourcegraph-jetbrains/blob/main/CHANGELOG.md).
