# Sourcegraph for JetBrains IDEs [![JetBrains Plugin](https://img.shields.io/badge/JetBrains-Sourcegraph-green.svg)](https://plugins.jetbrains.com/plugin/9682-sourcegraph)

- Search snippets of code on Sourcegraph.
- Copy and share a link to code on Sourcegraph.
- Quickly go from files in your editor to Sourcegraph.

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

The plugin is configurable _globally_ by creating a `sourcegraph-jetbrains.properties` in your home directory. For example, modify the following URL to match your on-premises Sourcegraph instance URL:

```
url = https://sourcegraph.example.com
```

You may also choose to configure it _per repository_ using a `.idea/sourcegraph.xml` file in your repository like so:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<project version="4">
    <component name="Config">
        <option name="url" value="https://sourcegraph.example.com" />
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
- Develop as you would normally (hit Debug icon in top right of IntelliJ).
- To create `sourcegraph.jar`:
  1. Update `plugin.xml` (change version AND describe changes in change notes).
  2. Update `Util.java` (change `VERSION` constant).
  3. Update `README.md` (copy changelog from plugin.xml).
  4. choose `Build` -> `Prepare Plugin Module 'sourcegraph' For Deployment`
  5. `git commit -m "all: release v<THE VERSION>"` and `git push` and `git tag v<THE VERSION>` and `git push --tags`
  6. Upload the jar to the releases tab of this repository.
  7. Publish according to http://www.jetbrains.org/intellij/sdk/docs/basics/getting_started/publishing_plugin.html (note: it takes ~2 business days for JetBrains support team to review the plugin).

## Version History

#### v1.2.0

- The search menu entry is now no longer present when no text has been selected.
- When on a branch that does not exist remotely, `master` will now be used instead.
- Menu entries (Open file, etc.) are now under a Sourcegraph sub-menu.
- Added a "Copy link to file" action (alt+c / opt+c).
- Added a "Search in repository" action (alt+r / opt+r).
- It is now possible to configure the plugin per-repository using a `.idea/sourcegraph.xml` file. See the README for details.
- Special thanks: @oliviernotteghem for contributing the new features in this release!

#### v1.1.2

- Fixed an error that occurred when trying to search with no selection.
- The git remote used for repository detection is now `sourcegraph` and then `origin`, instead of the previously poor choice of just the first git remote.

#### v1.1.1

- Fixed search shortcut; Updated the search URL to reflect a recent Sourcegraph.com change.

#### v1.1.0

- Added support for using the plugin with on-premises Sourcegraph instances.

#### v1.0.0

- Initial Release; basic Open File & Search functionality.
