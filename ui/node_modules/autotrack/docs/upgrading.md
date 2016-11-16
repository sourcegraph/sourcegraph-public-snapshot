# Upgrade Guide

This guide outlines how to upgrade from any pre-1.0 version to version 1.0.

## Breaking changes

### Global changes

In all versions prior to 1.0, you could include all autotrack functionality with the single command `ga('require', 'autotrack')`. This was a convenient shorthand that would individually require all other plugins. You can reference the [original usage instructions](https://github.com/googleanalytics/autotrack/blob/0.6.5/README.md#usage) to see an example.

In versions 1.0+, you can no longer require all sub-plugins with this one command. Instead, you have explicitly require each plugin you want to use and pass it its own configuration options (if necessary). This change was made to avoid users accidentally enabling plugin behavior they didn't intend.

The follow example shows how to require all autotrack plugins in versions 1.0+ *(note: the configuration options are omitted for simplicity)*:

```html
<script>
window.ga=window.ga||function(){(ga.q=ga.q||[]).push(arguments)};ga.l=+new Date;
ga('create', 'UA-XXXXX-Y', 'auto');

// Plugins must be required individually.
ga('require', 'cleanUrlTracker', {...});
ga('require', 'eventTracker', {...});
ga('require', 'impressionTracker', {...});
ga('require', 'mediaQueryTracker', {...});
ga('require', 'outboundFormTracker', {...});
ga('require', 'outboundLinkTracker', {...});
ga('require', 'pageVisibilityTracker', {...});
ga('require', 'socialWidgetTracker', {...});
ga('require', 'urlChangeTracker', {...});
// ...

ga('send', 'pageview');
</script>
<script async src="https://www.google-analytics.com/analytics.js"></script>
<script async src="path/to/autotrack.js"></script>
```

In all 1.x versions, requiring the `autotrack` plugin will do nothing but log a warning to the console. In version 2.0, this warning will go away, and calls to require autotrack may prevent [subsequent commands from running](https://devsite.googleplex.com/analytics/devguides/collection/analyticsjs/using-plugins#waiting_for_plugins_to_load).

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
