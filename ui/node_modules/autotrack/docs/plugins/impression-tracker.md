# `impressionTracker`

This guide explains what the `impressionTracker` plugin is and how to integrate it into your `analytics.js` tracking implementation.

## Overview

The `impressionTracker` plugin allows you to specify a list of elements and then track whether any of those elements are visible within the browser viewport. If any of the elements are not visible, an event is sent to Google Analytics as soon as they become visible.

Impression tracking is useful for getting a more accurate sense of whether particular advertisements or call-to-action elements were seen by the user.

## Usage

To enable the `impressionTracker` plugin, run the [`require`](https://developers.google.com/analytics/devguides/collection/analyticsjs/using-plugins) command, specify the plugin name `'impressionTracker'`, and pass in the configuration options you want to set:

```js
ga('require', 'impressionTracker', options);
```

### Browser support

The `impressionTracker` plugin takes advantage of a new browser API called [`IntersectionObserver`](https://developer.mozilla.org/en-US/docs/Web/API/Intersection_Observer_API), which allows you to register a callback that gets invoked whenever a target element is visible within the viewport.

The `IntersectionObserver` API is supported natively in Chrome 51+, and with an [`IntersectionObserver` polyfill](https://github.com/WICG/IntersectionObserver/tree/gh-pages/polyfill), it can be used in all other browsers as well.

To use the polyfill, add it to your page prior to requiring the `impressionTracker` plugin:

The `IntersectionObsever` polyfill is available on npm, and can be installed by running the following command:

```
npm install intersection-observer
```

Then link to the `intersection-observer.js` script file in your HTML page:


```html
<script src="path/to/intersection-observer.js"></script>
```

The `IntersectionObsever` polyfill is also available from [polyfill.io](https://polyfill.io). An advantage of using `polyfill.io` is the service automatically detects browser support and only delivers the polyfill (and any needed dependencies) if the requesting brower lacks native support.

You can link to the CDN version on `polyfill.io` with the following script:

```html
<script src="https://cdn.polyfill.io/v2/polyfill.min.js?features=IntersectionObserver"></script>
```

See the [`polyfill.io` documentation](https://polyfill.io/v2/docs/) for more information on using the service.

## Options

The following table outlines all possible configuration options for the `impressionTracker` plugin. If any of the options has a default value, the default is explicitly stated:

<table>
  <tr valign="top">
    <th align="left">Name</th>
    <th align="left">Type</th>
    <th align="left">Description</th>
  </tr>
  <tr valign="top">
    <td><code>elements</code></a></td>
    <td><code>Array&lt;string|Object&gt;</code></a></td>
    <td>
      A list of element IDs or element objects. See the <a href="#element-object-properties"><code>element</code> object properties</a> section below for details.</td>
  </tr>
  <tr valign="top">
    <td><code>rootMargin</code></a></td>
    <td><code>string</code></a></td>
    <td>A CSS margin string accepting pixel or percentage values. It is passed as the <a href="https://developer.mozilla.org/en-US/docs/Web/API/IntersectionObserver#Properties"><code>rootMargin</code></a> option to the <code>IntersectionObserver</code> instance, which is used to expand or contract the viewport area to change when an element is considered visible. For example: the string <code>'-20px 0'</code> would contract the viewport by 20 pixels on the top and bottom sides, and all element visibility calculations would be based on that rather than the full viewport dimensions.</td>
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

### `element` object properties

<table>
  <tr valign="top">
    <th align="left">Name</th>
    <th align="left">Type</th>
    <th align="left">Description</th>
  </tr>
  <tr valign="top">
    <td><code>id</code></a></td>
    <td><code>string</code></a></td>
    <td>The ID attribute of the element to track.</td>
  </tr>
  <tr valign="top">
    <td><code>threshold</code></a></td>
    <td><code>number</code></a></td>
    <td>
      A percentage of the element's area that must be within the viewport in order for the element to be considered visible. This value is used as one of the <a href="https://developer.mozilla.org/en-US/docs/Web/API/IntersectionObserver#Properties"><code>thresholds</code></a> passed to the <code>IntersectionObserver</code> instance. A threshold of <code>1</code> means the element must be entirely within the viewport to be considered visible. A threshold of <code>0.5</code> means half of the element must be within the viewport. And a threshold of <code>0</code> means if any of the element intersects with the viewport at all (including just the border), it is considered visible.<br>
      <strong>Default:</strong> <code>0</code>
    </td>
  </tr>
  <tr valign="top">
    <td><code>trackFirstImpressionOnly</code></a></td>
    <td><code>boolean</code></a></td>
    <td>
      When <code>true</code>, an impression for this element is only tracked the first time it is visible within the viewport. Set this to <code>false</code> if you want to track subsequent impressions.<br>
      <strong>Default:</strong> <code>true</code>
    </td>
  </tr>

</table>

## Default field values

The `impressionTracker` plugin sets the following default field values on all hits it sends. To customize these values, use one of the [options](#options) described above, or set the field value declaratively as an attribute in the HTML.

<table>
  <tr valign="top">
    <th align="left">Field</th>
    <th align="left">Value</th>
  </tr>
  <tr valign="top">
    <td><a href="https://developers.google.com/analytics/devguides/collection/analyticsjs/field-reference#hitType"><code>hitType</code></a></td>
    <td><code>'event'</code></td>
  </tr>
  <tr valign="top">
    <td><a href="https://developers.google.com/analytics/devguides/collection/analyticsjs/field-reference#eventCategory"><code>eventCategory</code></a></td>
    <td><code>'Viewport'</code></a></td>
  </tr>
  <tr valign="top">
    <td><a href="https://developers.google.com/analytics/devguides/collection/analyticsjs/field-reference#eventAction"><code>eventAction</code></a></td>
    <td><code>'impression'</code></a></td>
  </tr>
  <tr valign="top">
    <td><a href="https://developers.google.com/analytics/devguides/collection/analyticsjs/field-reference#eventLabel"><code>eventLabel</code></a></td>
    <td><code>element.id</code></td>
  </tr>
</table>

**Note:** the reference to `element` in the table above refers to the element being observed.

## Methods

The following table lists all methods for the `impressionTracker` plugin:

<table>
  <tr valign="top">
    <th align="left">Name</th>
    <th align="left">Description</th>
  </tr>
  <tr valign="top">
    <td><code>remove</code></a></td>
    <td>Removes the <code>impressionTracker</code> plugin from the specified tracker, disconnects all observers, and restores all modified tasks to their original state prior to the plugin being required.</td>
  </tr>
</table>

For details on how `analytics.js` plugin methods work and how to invoke them, see [calling plugin methods](https://developers.google.com/analytics/devguides/collection/analyticsjs/using-plugins#calling_plugin_methods) in the `analytics.js` documentation.

## Examples

### Basic tracking of when elements are visible

This example sends an event when any part of the `#foo` and/or `#bar` elements become visible in the viewport:

```js
ga('require', 'impressionTracker', {
  elements: ['foo', 'bar']
});
```

### Using threshold values

This example only sends events when the `#foo` and/or `#bar` element are at least half-way visible in the viewport:

```js
ga('require', 'impressionTracker', {
  elements: [
    {
      id: 'foo',
      threshold: 0.5
    },
    {
      id: 'bar',
      threshold: 0.5
    }
  ]
});
```

### Tracking multiple impressions for the same element

This example sends events anytime the `#foo` and/or `#bar` elements become visible in the viewport. Then, if the `#foo` and/or `#bar` elements leave the viewport and become visible again later, another event is sent for each occurrence.

```js
ga('require', 'impressionTracker', {
  elements: [
    {
      id: 'foo',
      trackFirstImpressionOnly: false
    },
    {
      id: 'bar',
      trackFirstImpressionOnly: false
    }
  ]
});
```

### Change the `rootMargin` value

This example sends events anytime the `#foo` and/or `#bar` elements appear at least within 20 pixels of the top and bottom edges of the viewport.

```js
ga('require', 'impressionTracker', {
  elements: ['foo', 'bar'],
  rootMargin: '-20px 0'
});
```
