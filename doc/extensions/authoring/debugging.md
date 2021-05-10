# Debugging a Sourcegraph extension

You can use your browser's devtools to debug Sourcegraph extensions running on a Sourcegraph instance (including Sourcegraph.com). Most browsers ship with an advanced JavaScript debugger, so this is a great way to find the cause of bugs in your extension.

## Setup

Before debugging your extension, make sure of the following things:

- You've enabled the extension on your Sourcegraph instance.
- You're on a page that triggers your extension's `activationEvents`. (For example, extensions with `*` are always activated. Extensions with `onLanguage:python` are only activated when you view `.py` files.)
- You published a sourcemap along with your extension's JavaScript file. The `src extensions publish` command expects a `.map` file next to your extension's `.js` file (e.g., `extension.js` and `extension.map`). If you used the [extension creator](https://github.com/sourcegraph/create-extension), this is already set up for you. (If you can't use sourcemaps, debugging the transpiled JavaScript code is still possible.)

## Use console.log

For quick debugging, you can use `console.log` (or [other `console` methods](https://developer.mozilla.org/en-US/docs/Web/API/console)). Sourcegraph extensions just consist of JavaScript code running in your browser in a Web Worker, so all of your favorite JavaScript debugging tricks work.

If debugging on a code host e.g. GitHub, the extension runs as a background script. To view your `console.log` statements, go to the background page in chrome://extensions. If you are unable to see a "Background page" link under the Sourcegraph extension then you need to enable Developer mode.

## Use the JavaScript debugger

Because Sourcegraph extensions just consist of JavaScript code that runs in a Web Worker in your browser, you can use your browser's JavaScript debugger to set breakpoints and step through execution.

To set breakpoints and step through execution in your Sourcegraph extension:

1. Open your browser's devtools.
1. In the **Sources** tab (or wherever your browser's devtools shows source files), open a source file from your extension. There are 3 ways to do this:
   - Press <kbd>Ctrl</kbd><kbd>P</kbd> or <kbd>Cmd</kbd><kbd>P</kbd> and search for the file by name. If you used the [extension creator](https://github.com/sourcegraph/create-extension), the main source filename is the extension name followed by `.ts` (e.g., `my-extension.ts`).
   - Add a `console.log` to your extension code and reload. Look in the devtools console for the log message, and then click on the filename where the log line was emitted (which is usually shown on the far right).
   - Add a `debugger;` statement to your extension code and reload (with your browser devtools open). Execution will stop when your browser encounters that statement, and you'll be dropped into the source file.
1. Set breakpoints and step through execution as you would for any other JavaScript file in your browser's devtools.

## `sourcegraph.app.log`

`console.log` can be helpful to debug Sourcegraph extensions locally, but sometimes you'll want to see logs from users of a published extension, which can help with debugging issues that are hard to reproduce. If you use `console.log` for these kinds of logs in a published extension, extension may pollute the console with unnecessary logs when things are working as expected. Using `sourcegraph.app.log`, you can comfortably log all the information you need while giving end users the ability to toggle your extension's logs on and off as needed.

> Note: This feature was introduced in Sourcegraph version 3.28. Extensions should check if `sourcegraph.app.log` is defined to prevent errors on older versions of Sourcegraph.

### Extension API

In your extension, call [`sourcegraph.app.log`]() anywhere you'd call `console.log`.

<!-- TODO: link to extension API in main -->


### User settings

Users can enable an extension's logs by adding its ID to the [`extensions.activeLoggers`]() setting.

<!-- TODO: link to settings schema in main -->
