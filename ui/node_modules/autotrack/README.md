# Autotrack [![Build Status](https://travis-ci.org/googleanalytics/autotrack.svg?branch=master)](https://travis-ci.org/googleanalytics/autotrack)

- [Overview](#overview)
- [Plugins](#plugins)
- [Installation and usage](#installation-and-usage)
  - [Loading autotrack via npm](#loading-autotrack-via-npm)
  - [Passing configuration options](#passing-configuration-options)
- [Advanced configuration](#advanced-configuration)
  - [Custom builds](#custom-builds)
  - [Using autotrack with multiple trackers](#using-autotrack-with-multiple-trackers)
- [Browser Support](#browser-support)
- [Translations](#translations)

## Overview

The default [JavaScript tracking snippet](https://developers.google.com/analytics/devguides/collection/analyticsjs/) for Google Analytics runs when a web page is first loaded and sends a pageview hit to Google Analytics. If you want to know about more than just pageviews (e.g. events, social interactions), you have to write code to capture that information yourself.

Since most website owners care about a lot of the same types of user interactions, web developers end up writing the same code over and over again for every new site they build.

Autotrack was created to solve this problem. It provides default tracking for the interactions most people care about, and it provides several convenience features (e.g. declarative event tracking) to make it easier than ever to understand how people are using your site.

## Plugins

The `autotrack.js` library is small (6K gzipped), and includes the following plugins. By default all plugins are bundled together, but they can be included and configured separately as well. This table includes a brief description of each plugin; you can click on the plugin name to see the full documentation and usage instructions:

<table>
  <tr>
    <th align="left">Plugin</th>
    <th align="left">Description</th>
  </tr>
  <tr>
    <td><a href="/docs/plugins/clean-url-tracker.md"><code>cleanUrlTracker</code></a></td>
    <td>Ensures consistency in the URL paths that get reported to Google Analytics; avoiding the problem where separate rows in your pages reports actually point to the same page.</td>
  </tr>
  <tr>
    <td><a href="/docs/plugins/event-tracker.md"><code>eventTracker</code></a></td>
    <td>Enables declarative event tracking, via HTML attributes in the markup.</td>
  </tr>
  <tr>
    <td><a href="/docs/plugins/impression-tracker.md"><code>impressionTracker</code></a></td>
    <td>Allows you to track when elements are visible within the viewport.</td>
  </tr>
  <tr>
    <td><a href="/docs/plugins/media-query-tracker.md"><code>mediaQueryTracker</code></a></td>
    <td>Enables tracking media query matching and media query changes.</td>
  </tr>
  <tr>
    <td><a href="/docs/plugins/outbound-form-tracker.md"><code>outboundFormTracker</code></a></td>
    <td>Automatically tracks form submits to external domains.</td>
  </tr>
  <tr>
    <td><a href="/docs/plugins/outbound-link-tracker.md"><code>outboundLinkTracker</code></a></td>
    <td>Automatically tracks link clicks to external domains.</td>
  </tr>
  <tr>
    <td><a href="/docs/plugins/page-visibility-tracker.md"><code>pageVisibilityTracker</code></a></td>
    <td>Tracks page visibility state changes, which enables much more accurate session, session duration, and pageview metrics.</td>
  </tr>
  <tr>
    <td><a href="/docs/plugins/social-widget-tracker.md"><code>socialWidgetTracker</code></a></td>
    <td>Automatically tracks user interactions with the official Facebook and Twitter widgets.</td>
  </tr>
  <tr>
    <td><a href="/docs/plugins/url-change-tracker.md"><code>urlChangeTracker</code></a></td>
    <td>Automatically tracks URL changes for single page applications.</td>
  </tr>
</table>

**Disclaimer:** autotrack is maintained by members of the Google Analytics developer platform team and is primarily intended for a developer audience. It is not an official Google Analytics product and does not qualify for Google Analytics 360 support. Developers who choose to use this library are responsible for ensuring that their implementation meets the requirements of the [Google Analytics Terms of Service](https://www.google.com/analytics/terms/us.html) and the legal obligations of their respective country.

## Installation and usage

To add autotrack to your site, you have to do two things:

1. Load the `autotrack.js` script file on your page.
2. Update your [tracking snippet](https://developers.google.com/analytics/devguides/collection/analyticsjs/tracking-snippet-reference) to [require](https://developers.google.com/analytics/devguides/collection/analyticsjs/using-plugins) the various autotrack plugins you want to use.

If your site already includes the default JavaScript tracking snippet, you can modify it too look something like this:

```html
<script>
window.ga=window.ga||function(){(ga.q=ga.q||[]).push(arguments)};ga.l=+new Date;
ga('create', 'UA-XXXXX-Y', 'auto');

// Replace the following lines with the plugins you want to use.
ga('require', 'eventTracker');
ga('require', 'outboundLinkTracker');
ga('require', 'urlChangeTracker');
// ...

ga('send', 'pageview');
</script>
<script async src='https://www.google-analytics.com/analytics.js'></script>
<script async src='path/to/autotrack.js'></script>
```

Of course, you'll have to make the following modifications to customize autotrack to your needs:

- Replace `UA-XXXXX-Y` with your [tracking ID](https://support.google.com/analytics/answer/1032385)
- Replace the sample list of plugin `require` statements with the plugins you want to use.
- Replace `path/to/autotrack.js` with the actual location of the `autotrack.js` file hosted on your server.

**Note:** the [analytics.js plugin system](https://developers.google.com/analytics/devguides/collection/analyticsjs/using-plugins) is designed to support asynchronously loaded scripts, so it doesn't matter if `autotrack.js` is loaded before or after `analytics.js`. It also doesn't matter if the `autotrack.js` library is loaded individually or bundled with the rest of your JavaScript code.

### Loading autotrack via npm

If you use npm and a module loader like [Browserify](http://browserify.org/), [Webpack](https://webpack.github.io/), or [SystemJS](https://github.com/systemjs/systemjs), you can include autotrack in your build by requiring it as you would any other npm module:

```sh
npm install autotrack
```

```js
// In your JavaScript code
require('autotrack');
```

The above code will include all autotrack plugins in your generated source file. If you only want to include a specific set of plugins, you can require them individually:

```js
// In your JavaScript code
require('autotrack/lib/plugins/clean-url-tracker');
require('autotrack/lib/plugins/outbound-link-tracker');
require('autotrack/lib/plugins/url-change-tracker');
// ...
```

The above examples show how to include the plugin source code in your final, generated JavaScript file, which accomplishes the first step of the two-step installation process.

You still have to update your tracking snippet and require the plugins you want to use:


```js
// In the analytics.js tracking snippet
ga('create', 'UA-XXXXX-Y', 'auto');

// Replace the following lines with the plugins you want to use.
ga('require', 'cleanUrlTracker');
ga('require', 'outboundLinkTracker');
ga('require', 'urlChangeTracker');
// ...

ga('send', 'pageview');
```

**Note:** be careful not to confuse the node module [`require`](https://nodejs.org/api/modules.html) statement with the `analytics.js` [`require`](https://developers.google.com/analytics/devguides/collection/analyticsjs/command-queue-reference#require) command. When loading autotrack with an npm module loader, both requires must be used.

### Passing configuration options

All autotrack plugins accept a configuration object as the third parameter to the `require` command.

Some of the plugins (e.g. `outboundLinkTracker`, `socialWidgetTracker`, `urlChangeTracker`) have a default behavior that works for most people without specifying any configuration options. Other plugins (e.g. `cleanUrlTracker`, `impressionTracker`, `mediaQueryTracker`) require certain configuration options to be set in order to work.

See the individual plugin documentation to reference what options each plugin accepts (and what the default value is, if any).

## Advanced configuration

### Custom builds

The autotrack library is built modularly and each plugin includes its own dependencies, so you can create a custom build of the library using a script bundler such as [Browserify](http://browserify.org/).

The following example shows how to create a build that only includes the `eventTracker` and `outboundLinkTracker` plugins:

```sh
browserify lib/plugins/event-tracker lib/plugins/outbound-link-tracker
```

When making a custom build, be sure to update the tracking snippet to only require plugins included in your build. Requiring a plugin that's not included in the build will create an unmet dependency, which will prevent subsequent commands from running.

If you're already using a module loader like [Browserify](http://browserify.org/), [Webpack](https://webpack.github.io/), or [SystemJS](https://github.com/systemjs/systemjs) to build your JavaScript, you can skip the above step and just require the plugins as described in the [loading autotrack via npm](#loading-autotrack-via-npm) section.

### Using autotrack with multiple trackers

All autotrack plugins support multiple trackers and work by specifying the tracker name in the `require` command. The following example creates two trackers and requires various autotrack plugins on each.

```js
// Creates two trackers, one named `tracker1` and one named `tracker2`.
ga('create', 'UA-XXXXX-Y', 'auto', 'tracker1');
ga('create', 'UA-XXXXX-Z', 'auto', 'tracker2');

// Requires plugins on tracker1.
ga('tracker1.require', 'eventTracker');
ga('tracker1.require', 'socialWidgetTracker');

// Requires plugins on tracker2.
ga('tracker2.require', 'eventTracker');
ga('tracker2.require', 'outboundLinkTracker');
ga('tracker2.require', 'pageVisibilityTracker');

// Sends the initial pageview for each tracker.
ga('tracker1.send', 'pageview');
ga('tracker2.send', 'pageview');
```

## Browser Support

Autotrack will safely run in any browser without errors, as feature detection is always used with any potentially unsupported code. However, autotrack will only track features supported in the browser running it. For example, a user running Internet Explorer 8 will not be able to track media query usage, as media queries themselves aren't supported in Internet Explorer 8.

All autotrack plugins are [tested via Sauce Labs](https://saucelabs.com/u/autotrack) in the following browsers:

<table>
  <tr>
    <td align="center">
      <img src="https://raw.github.com/alrra/browser-logos/master/chrome/chrome_48x48.png" alt="Chrome"><br>
      ✔
    </td>
    <td align="center">
      <img src="https://raw.github.com/alrra/browser-logos/master/firefox/firefox_48x48.png" alt="Firefox"><br>
      ✔
    </td>
    <td align="center">
      <img src="https://raw.github.com/alrra/browser-logos/master/safari/safari_48x48.png" alt="Safari"><br>
      6+
    </td>
    <td align="center">
      <img src="https://raw.github.com/alrra/browser-logos/master/edge/edge_48x48.png" alt="Edge"><br>
      ✔
    </td>
    <td align="center">
      <img src="https://raw.github.com/alrra/browser-logos/master/internet-explorer/internet-explorer_48x48.png" alt="Internet Explorer"><br>
      9+
    </td>
    <td align="center">
      <img src="https://raw.github.com/alrra/browser-logos/master/opera/opera_48x48.png" alt="Opera"><br>
      ✔
    </td>
  </tr>
</table>

## Translations

The following translations have been graciously provided by the community. Please note that these translations are unofficial and may be inaccurate or out of date:

* [Japanese](https://github.com/nebosuker/autotrack)
* [Chinese](https://github.com/stevezhuang/autotrack/blob/master/README.zh.md)

If you discover issues with a particular translation, please file them with the appropriate repository. To submit your own translation, follow these steps:

1. Fork this repository.
2. Update the settings of your fork to [allow issues](http://programmers.stackexchange.com/questions/179468/forking-a-repo-on-github-but-allowing-new-issues-on-the-fork).
3. Remove all non-documentation files.
4. Update the documentation files with your translated versions.
5. Submit a pull request to this repository that adds a link to your fork to the above list.
