# Common Options

Many of the autotrack plugins accept options that are common to multiple different plugins. The following common options are documented in this guide:

- [`fieldsObj`](#fieldsobj)
- [`attributePrefix`](#attributeprefix)
- [`hitFilter`](#hitfilter)

## `fieldsObj`

Some of the autotrack plugins send hits with default [analytics.js field values](https://developers.google.com/analytics/devguides/collection/analyticsjs/field-reference) set. These plugins accept a `fieldsObj` option, which allows you to customize those values for each plugin. It also allows you to set any fields that aren't set by default.

The `fieldsObj` option is an `Object` whose properties can be any [analytics.js field name](https://developers.google.com/analytics/devguides/collection/analyticsjs/field-reference), and whose values will be used as the corresponding field value for all hits sent by the plugin.

### Examples

#### `mediaQueryTracker`

This configuration ensures all events sent by the `mediaQueryTracker` plugin are non-interaction events:

```js
ga('require', 'mediaQueryTracker', {
  definitions: [...],
  fieldsObj: {
    nonInteraction: true
  }
});
```

#### `urlChangeTracker`

This configuration sets a [custom dimension](https://support.google.com/analytics/answer/2709828) at index 1 for all pageview hits sent by the `urlChangeTracker`. This would allow you to differentiate between the initial pageview and "virtual" pageviews sent after loading new pages via AJAX:

```js
ga('require', 'urlChangeTracker', {
  fieldsObj: {
    dimension1: 'virtual'
  }
});
ga('send', 'pageview', {
  dimension1: 'pageload'
});
```

## `attributePrefix`

All plugins that send hits to Google Analytics as a result of user interactions with DOM elements support declarative attribute binding (see the [`eventTracker`](/docs/plugins/event-tracker.md) plugin to see how this works). As such, each of these plugins accept an `attributePrefix` option to customize what attribute prefix to use.

By default, the `attributePrefix` value used by each plugin is the string `'ga-'`, though that value can be customized on a per-plugin basis.

**Note:** when setting the same field in both the `fieldsObj` option as well as via a DOM element attribute, the attribute's value will override the `fieldsObj` value.

### Examples

#### `eventTracker`

```js
ga('require', 'eventTracker', {
  attributePrefix: 'data-'
});
```

```html
<button
  data-on="click"
  data-event-category="Video"
  data-event-action="play">
  Play video
</button>
```

#### `impressionTracker`

```js
ga('require', 'impressionTracker', {
  elements: ['cta'],
  attributePrefix: 'data-ga'
});
```

```html
<div
  id="cta"
  data-ga-event-category="Call to action"
  data-ga-event-action="seen">
  Call to action
</a>
```

#### `outboundLinkTracker`

```js
ga('require', 'outboundLinkTracker', {
  attributePrefix: ''
});
```

```html
<a href="https://example.com" event-category="External Link">Click</a>
```

## `hitFilter`

The `hitFilter` option is useful when you need to make more advanced modifications to a hit, or when you need to abort the hit altogether. `hitFilter` is a function that gets invoked with the tracker's [model object](https://developers.google.com/analytics/devguides/collection/analyticsjs/model-object-reference) as its first argument, and (if the hit was initiated by a user interaction with a DOM element) the DOM element as the second argument.

Within the `hitFilter` function you can get the value of any of the model object's fields using the [`get`](https://developers.google.com/analytics/devguides/collection/analyticsjs/model-object-reference#get) method on the `model` argument. And you can set a new value using the [`set`](https://developers.google.com/analytics/devguides/collection/analyticsjs/model-object-reference#set) method on the `model` argument. To abort the hit, throw an error.

To modify the model for the current hit only (and not all subsequent hits), make sure to set the third argument ([`temporary`](https://developers.google.com/analytics/devguides/collection/analyticsjs/model-object-reference#set)) to `true`.

### How it works

The `hitFilter` option works by overriding the tracker's [`buildHitTask`](https://developers.google.com/analytics/devguides/collection/analyticsjs/tasks). The passed `hitFilter` function runs after the `fieldsObj` values and attribute fields have been set on the tracker but before running the original `buildHitTask`. Refer to the guide on [analytics.js tasks](https://developers.google.com/analytics/devguides/collection/analyticsjs/tasks) to learn more.

### Examples

#### `pageVisibilityTracker`

This configuration sets custom dimension 1 to whatever the `eventValue` field is set to for all `visibilitychange` event hits. It specifies `true` as the third argument to the `set` method, so this change affects the current hit only:

```js
ga('require', 'pageVisibilityTracker', {
  hitFilter: function(model) {
    model.set('dimension1', String(model.get('eventValue')), true);
  }
});
```

#### `impressionTracker`

This configuration prevents hits from being sent for impressions on elements with the `is-invisible` class.

```js
ga('require', 'impressionTracker', {
  hitFilter: function(model, element) {
    if (element.className.indexOf('is-invisible') > -1) {
      throw new Error('Aborting hit');
    }
  }
});
```
