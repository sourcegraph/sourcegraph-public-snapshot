# `eventTracker`

This guide explains what the `eventTracker` plugin is and how to integrate it into your `analytics.js` tracking implementation.

## Overview

Many website development tools and content management systems will give page authors access to modify the HTML templates and page content but not give them access to the site's JavaScript. In such cases, it's very difficult to add event listeners to track user interactions with elements on the page.

The `eventTracker` plugin solves this problem by providing declarative event binding to attributes in the HTML, making it possible to track user interactions with DOM elements without writing any JavaScript.

## Usage

To enable the `eventTracker` plugin, run the [`require`](https://developers.google.com/analytics/devguides/collection/analyticsjs/using-plugins) command, specify the plugin name `'eventTracker'`, and pass in the configuration options (if any) you want to set:

```js
ga('require', 'eventTracker', options);
```

### Modifying the HTML

To add declarative interaction tracking to a DOM element, you start by adding a `ga-on` attribute (assuming the default `'ga-'` attribute prefix) and setting its value to whatever DOM event you want to listen for (note: it must be one of the events specified in the `events` configuration option). When the specified event is detected, a hit is sent to Google Analytics with whatever field attribute values are present on the element.

Any valid [analytics.js field](https://developers.google.com/analytics/devguides/collection/analyticsjs/field-reference) can be set declaratively via an attribute. The attribute name can be determined by combining the [`attributePrefix`](#options) option with the [kebab-cased](https://en.wikipedia.org/wiki/Letter_case#Special_case_styles) version of the field name. For example, if you want to set the [`eventCategory`](https://developers.google.com/analytics/devguides/collection/analyticsjs/field-reference#eventCategory) field and you're using the default `attributePrefix` of `'ga-'`, you would use the attribute name `ga-event-category`.

Refer to the [examples](#examples) section to see what the code looks like. For a complete list of possible fields to send, refer to the [field reference](https://developers.google.com/analytics/devguides/collection/analyticsjs/field-reference) in the `analytics.js` documentation.

## Options

The following table outlines all possible configuration options for the `eventTracker` plugin. If any of the options has a default value, the default is explicitly stated:

<table>
  <tr valign="top">
    <th align="left">Name</th>
    <th align="left">Type</th>
    <th align="left">Description</th>
  </tr>
  <tr valign="top">
    <td><code>events</code></a></td>
    <td><code>Array</code></a></td>
    <td>
      A list of DOM events to listen for. Note that in order for an event set in the HTML via the <code>*-on</code> attribute to work, it must be listed in this array.<br>
      <strong>Default:</strong> <code>['click']</code>
    </td>
  </tr>
  <tr valign="top">
    <td><code>fieldsObj</code></a></td>
    <td><code>Object</code></a></td>
    <td>See the <a href="/docs/common-options.md#fieldsobj">common options guide</a> for the <code>fieldsObj</code> description.</td>
  </tr>
  <tr valign="top">
    <td><code>attributePrefix</code></a></td>
    <td><code>string</code></a></td>
    <td>
      See the <a href="/docs/common-options.md#attributeprefix">common options guide</a> for the <code>attributePrefix</code> description.<br>
      <strong>Default:</strong> <code>'ga-'</code>
    </td>
  </tr>
  <tr valign="top">
    <td><code>hitFilter</code></a></td>
    <td><code>Function</code></a></td>
    <td>See the <a href="/docs/common-options.md#hitfilter">common options guide</a> for the <code>hitFilter</code> description.</td>
  </tr>
</table>

## Default field values

The `eventTracker` plugin sets the following default field values on all hits it sends. To customize these values, use one of the [options](#options) described above, or set the field value declaratively as an attribute in the HTML.

<table>
  <tr valign="top">
    <th align="left">Field</th>
    <th align="left">Value</th>
  </tr>
  <tr valign="top">
    <td><a href="https://developers.google.com/analytics/devguides/collection/analyticsjs/field-reference#hitType"><code>hitType</code></a></td>
    <td><code>'event'</code></a></td>
  </tr>
</table>

## Methods

The following table lists all methods for the `eventTracker` plugin:

<table>
  <tr valign="top">
    <th align="left">Name</th>
    <th align="left">Description</th>
  </tr>
  <tr valign="top">
    <td><code>remove</code></a></td>
    <td>Removes the <code>eventTracker</code> plugin from the specified tracker, removes all event listeners from the DOM, and restores all modified tasks to their original state prior to the plugin being required.</td>
  </tr>
</table>

For details on how `analytics.js` plugin methods work and how to invoke them, see [calling plugin methods](https://developers.google.com/analytics/devguides/collection/analyticsjs/using-plugins#calling_plugin_methods) in the `analytics.js` documentation.

## Examples

### Basic usage

This example shows how to write the markup when not setting any configuration options:

```js
ga('require', 'eventTracker');
```

```html
<button
  ga-on="click"
  ga-event-category="Video"
  ga-event-action="play">
  Play video
</button>
```

### Customizing the `events` and `attributePrefix` options

This example customizes the `eventTracker` plugin to listen for right clicks (via the `contextmenu`  event). It also uses `'data-'` as the attribute prefix rather than the default `ga-`:

```js
ga('require', 'eventTracker', {
  events: ['contextmenu'],
  attributePrefix: 'data-'
});
```

The follow HTML will track right clicks given the above configuration:

```html
<button
  data-on="contextmenu"
  data-event-category="Info Button"
  data-event-action="right click">
  Info
</button>
```

### Tracking non-event hit types

The default `hitType` for all hits sent by the `eventTracker` plugin is `'event'`, but this can be customized either with the [`fieldsObj`](/docs/common-options.md#fieldsobj) or [`hitFilter`](/docs/common-options.md#hitfilter) options, or setting the `ga-hit-type` attribute on the element itself (assuming the default `ga-` attribute prefix).

For example, to send a [social interaction hit](https://developers.google.com/analytics/devguides/collection/analyticsjs/social-interactions) instead of an event, you could use the following HTML:

```html
<button
  ga-on="click"
  ga-hit-type="social"
  ga-social-network="Facebook"
  ga-social-action="like">
  Like us on Facebook
</button>
```
