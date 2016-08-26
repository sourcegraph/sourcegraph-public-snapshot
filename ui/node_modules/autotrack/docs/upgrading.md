# Upgrade Guide

This guide outlines how to upgrade from any pre-1.0 version to version 1.0.

## Breaking changes

### Global changes

In version 1.0, you can no longer require all plugins with the command `ga('require', 'autotrack')`. This change was made to avoid users accidentally enabling plugin behavior they didn't intend.

Going forward, all autotrack plugins must be individually required, and their options individually specified.

```html
<script>
window.ga=window.ga||function(){(ga.q=ga.q||[]).push(arguments)};ga.l=+new Date;
ga('create', 'UA-XXXXX-Y', 'auto');

// Plugins must be required individually.
ga('require', 'eventTracker');
ga('require', 'outboundLinkTracker');
ga('require', 'urlChangeTracker');
// ...

ga('send', 'pageview');
</script>
<script async src="https://www.google-analytics.com/analytics.js"></script>
<script async src="path/to/autotrack.js"></script>
```

In all 1.x versions, a warning will be logged to the console if you require the `autotrack` plugin. In version 2.0, this warning will go away.

### Individual plugin changes

#### [`mediaQueryTracker`](/docs/plugins/media-query-tracker.md)

- The `mediaQueryDefinitions` option has been renamed to `definitions`.
- The `mediaQueryChangeTemplate` option has been renamed to `changeTemplate`.
- The `mediaQueryChangeTimeout` option has been renamed to `changeTimeout`.

#### `socialTracker`

- The `socialTracker` plugin has been renamed to [`socialWidgetTracker`](/docs/plugins/social-widget-tracker.md) and no longer supports declarative social interaction tracking (since that can now be handled entirely via the [`eventTracker`](/docs/plugins/event-tracker.md) plugin).

## Plugin enhancements

### Global enhancement

- All plugins that send hits accept both [`fieldsObj`](/docs/common-options.md#fieldsobj) and [`hitFilter`](/docs/common-options.md#hitfilter) options. These options can be used to set or change any valid analytics.js field prior to the hit being sent.
- All plugins that send hits as a result of user interaction with a DOM element support [setting field values declaratively](/docs/common-options.md#attributeprefix).

### Individual plugin enhancements

#### [`eventTracker`](/docs/plugins/event-tracker.md)

- Added support for declarative tracking of any DOM event, not just click events (e.g. `submit`, `contextmenu`, etc.)

#### [`outboundFormTracker`](/docs/plugins/outbound-form-tracker.md)

- Added support for tracking forms within shadow DOM subtrees.
- Added the ability to customize the selector used to identify forms.
- Added a `parseUrl` utility function to the `shouldTrackOutboundForm` method to more easily identify or exclude outbound forms.

#### [`outboundLinkTracker`](/docs/plugins/outbound-link-tracker.md)

- Added support for DOM events other than `click` (e.g. `contextmenu`, `touchend`, etc.)
- Added support for tracking links within shadow DOM subtrees.
- Added the ability to customize the selector used to identify links.
- Added a `parseUrl` utility function to the `shouldTrackOutboundLink` method to more easily identify or exclude outbound links.

## New plugins

The following new plugins have been added. See their individual documentation pages for usage details.

- [`cleanUrlTracker`](/docs/plugins/clean-url-tracker.md)
- [`impressionTracker`](/docs/plugins/impression-tracker.md)
- [`pageVisibilityTracker`](/docs/plugins/page-visibility-tracker.md)
