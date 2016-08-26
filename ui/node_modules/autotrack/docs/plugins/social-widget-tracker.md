# `socialWidgetTracker`

This guide explains what the `socialWidgetTracker` plugin is and how to integrate it into your `analytics.js` tracking implementation.

## Overview

The `socialWidgetTracker` automatically adds social tracking for the official Twitter tweet/follow buttons and the Facebook like button. If you have the official Twitter or Facebook buttons on your page and you've enabled the `socialWidgetTracker` plugin, user interactions with those buttons will be automatically tracked.

## Usage

To enable the `socialWidgetTracker` plugin, run the [`require`](https://developers.google.com/analytics/devguides/collection/analyticsjs/using-plugins) command, specify the plugin name `'socialWidgetTracker'`, and pass in the configuration options (if any) you want to set:

```js
ga('require', 'socialWidgetTracker', options);
```

## Options

The following table outlines all possible configuration options for the `socialWidgetTracker` plugin. If any of the options has a default value, the default is explicitly stated:

<table>
  <tr valign="top">
    <th align="left">Name</th>
    <th align="left">Type</th>
    <th align="left">Description</th>
  </tr>
  <tr valign="top">
    <td><code>fieldsObj</code></a></td>
    <td><code>Object</code></a></td>
    <td>See the <a href="/docs/common-options.md#fieldsobj">common options guide</a> for the <code>fieldsObj</code> description.</td>
  </tr>
  <tr valign="top">
    <td><code>hitFilter</code></a></td>
    <td><code>Function</code></a></td>
    <td>See the <a href="/docs/common-options.md#hitfilter">common options guide</a> for the <code>hitFilter</code> description.</td>
  </tr>
</table>

## Default field values

The `socialWidgetTracker` plugin sets the following default field values on all hits it sends. To customize these values, use one of the [options](#options) described above.

### Facebook like button

<table>
  <tr valign="top">
    <th align="left">Field</th>
    <th align="left">Value</th>
  </tr>
  <tr valign="top">
    <td><a href="https://developers.google.com/analytics/devguides/collection/analyticsjs/field-reference#hitType"><code>hitType</code></a></td>
    <td><code>'social'</code></td>
  </tr>
  <tr valign="top">
    <td><a href="https://developers.google.com/analytics/devguides/collection/analyticsjs/field-reference#socialNetwork"><code>socialNetwork</code></a></td>
    <td><code>'Facebook'</code></a></td>
  </tr>
  <tr valign="top">
    <td><a href="https://developers.google.com/analytics/devguides/collection/analyticsjs/field-reference#socialAction"><code>socialAction</code></a></td>
    <td><code>'like'</code> or <code>'unlike'</code> (depending on the button state)</td>
  </tr>
  <tr valign="top">
    <td><a href="https://developers.google.com/analytics/devguides/collection/analyticsjs/field-reference#socialTarget"><code>socialTarget</code></a></td>
    <td>The URL the button was registered with.</td>
  </tr>
</table>

### Twitter tweet button

<table>
  <tr valign="top">
    <th align="left">Field</th>
    <th align="left">Value</th>
  </tr>
  <tr valign="top">
    <td><a href="https://developers.google.com/analytics/devguides/collection/analyticsjs/field-reference#hitType"><code>hitType</code></a></td>
    <td><code>'social'</code></td>
  </tr>
  <tr valign="top">
    <td><a href="https://developers.google.com/analytics/devguides/collection/analyticsjs/field-reference#socialNetwork"><code>socialNetwork</code></a></td>
    <td><code>'Twitter'</code></a></td>
  </tr>
  <tr valign="top">
    <td><a href="https://developers.google.com/analytics/devguides/collection/analyticsjs/field-reference#socialAction"><code>socialAction</code></a></td>
    <td><code>'tweet'</code></td>
  </tr>
  <tr valign="top">
    <td><a href="https://developers.google.com/analytics/devguides/collection/analyticsjs/field-reference#socialTarget"><code>socialTarget</code></a></td>
    <td>The widget's <code>data-url</code> attribute or the URL of the current page.</td>
  </tr>
</table>

### Twitter follow button

<table>
  <tr valign="top">
    <th align="left">Field</th>
    <th align="left">Value</th>
  </tr>
  <tr valign="top">
    <td><a href="https://developers.google.com/analytics/devguides/collection/analyticsjs/field-reference#hitType"><code>hitType</code></a></td>
    <td><code>'social'</code></td>
  </tr>
  <tr valign="top">
    <td><a href="https://developers.google.com/analytics/devguides/collection/analyticsjs/field-reference#socialNetwork"><code>socialNetwork</code></a></td>
    <td><code>'Twitter'</code></a></td>
  </tr>
  <tr valign="top">
    <td><a href="https://developers.google.com/analytics/devguides/collection/analyticsjs/field-reference#socialAction"><code>socialAction</code></a></td>
    <td><code>'follow'</code></td>
  </tr>
  <tr valign="top">
    <td><a href="https://developers.google.com/analytics/devguides/collection/analyticsjs/field-reference#socialTarget"><code>socialTarget</code></a></td>
    <td>The widget's <code>data-screen-name</code> attribute.</td>
  </tr>
</table>

## Methods

The following table lists all methods for the `socialWidgetTracker` plugin:

<table>
  <tr valign="top">
    <th align="left">Name</th>
    <th align="left">Description</th>
  </tr>
  <tr valign="top">
    <td><code>remove</code></a></td>
    <td>Removes the <code>socialWidgetTracker</code> plugin from the specified tracker, removes all event listeners registered with the social SDKs, and restores all modified tasks to their original state prior to the plugin being required.</td>
  </tr>
</table>

For details on how `analytics.js` plugin methods work and how to invoke them, see [calling plugin methods](https://developers.google.com/analytics/devguides/collection/analyticsjs/using-plugins#calling_plugin_methods) in the `analytics.js` documentation.

## Examples

### Basic usage

In most cases, this plugin needs no customization:

```js
ga('require', 'socialWidgetTracker');
```

### Sending events instead of social interaction hits

If you want to send events instead of social interaction hits, you can map the hit field values via the [`hitFilter`](#options) option:

```js
ga('require', 'socialWidgetTracker', {
  hitFilter: function(model) {
    // Changes the hit type from `social` to `event`.
    model.set('hitType', 'event');

    // Maps the social values to event values.
    model.set('eventCategory', model.get('socialNetwork'));
    model.set('eventAction', model.get('socialAction'));
    model.set('eventLabel', model.get('socialTarget'));

    // Unsets the social values.
    model.set('socialNetwork', null);
    model.set('socialAction', null);
    model.set('socialTarget', null);
  }
});
```
