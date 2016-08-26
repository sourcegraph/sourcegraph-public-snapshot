# `mediaQueryTracker`

This guide explains what the `mediaQueryTracker` plugin is and how to integrate it into your `analytics.js` tracking implementation.

## Overview

Most sites today use responsive design to update the page layout based on the screen size or capabilities of the user's device. If [media queries](https://developer.mozilla.org/en-US/docs/Web/CSS/Media_Queries/Using_media_queries) are used to alter the look or functionality of a page, it's important to capture that information to better understand how usage differs when different media queries are active.

The `mediaQueryTracker` plugin allows you to register the set of media query values you're using, and those values are automatically tracked via [custom dimensions](https://support.google.com/analytics/answer/2709828) with each hit. It also sends events when those values change.

## Usage

To enable the `mediaQueryTracker` plugin, run the [`require`](https://developers.google.com/analytics/devguides/collection/analyticsjs/using-plugins) command, specify the plugin name `'mediaQueryTracker'`, and pass in the configuration options you want to set:

```js
ga('require', 'mediaQueryTracker', options);
```

The `options` object requires a list of media query definitions that must specify a [custom dimension](https://support.google.com/analytics/answer/2709828) index. Custom dimensions can be created in your property settings in Google Analytics. The following section explains how.

### Setting up custom dimensions in Google Analytics

1. Log in to Google Analytics, choose the [account and property](https://support.google.com/analytics/answer/1009618) you're sending data too, and [create a custom dimension](https://support.google.com/analytics/answer/2709829) for each set of media queries you want to track (e.g. Breakpoints, Resolution/DPI, Device Orientation)
2. Give each dimension a name (e.g. Breakpoints), select a scope of [hit](https://support.google.com/analytics/answer/2709828#example-hit), and make sure the "active" checkbox is checked.
3. In the [`definitions`](#definitions) config object, set the `name` and `dimensionIndex` values to be the same as the name and index shown in Google Analytics.

Refer to the [`definition`](#the-definition-object) object documentation for an example.

## Options

The following table outlines all possible configuration options for the `mediaQueryTracker` plugin. If any of the options has a default value, the default is explicitly stated:

<table>
  <tr valign="top">
    <th align="left">Name</th>
    <th align="left">Type</th>
    <th align="left">Description</th>
  </tr>
  <tr valign="top">
    <td><code>definitions</code></a></td>
    <td><code>Object|Array&lt;Object&gt;</code></a></td>
    <td>A <code>definition</code> object or an array of <code>definition</code> objects. See the <a href="#the-definition-object"><code>definition</code></a> object description for property details.</td>
  </tr>
  <tr valign="top">
    <td><code>changeTemplate</code></a></td>
    <td><code>Function</code></a></td>
    <td colspan="2">
    The <code>changeTemplate</code> function (via its return value) determines what the <a href="https://developers.google.com/analytics/devguides/collection/analyticsjs/field-reference#eventLabel"><code>eventLabel</code></a> field will be for event hits when the matching media definition changes. The function is invoked with the newly matching value and the previously matching value as its first and second arguments, respectively:<br>
    <strong>Default:</strong>
<pre>function(newValue, oldValue) {
  return oldValue + ' => ' + newValue;
}</pre>
    </td>
  </tr>
  <tr valign="top">
    <td><code>changeTimeout</code></a></td>
    <td><code>number</code></a></td>
    <td>The debounce timeout, i.e., the amount of time to wait before sending the change event. If multiple change events occur within the timeout period, only the last one is sent.<br>
    <strong>Default:</strong> <code>1000</code>
    </td>
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

### The `definition` object

The `definition` object allows you to group multiple different types of media queries together to be tracked by the same custom dimension.

<table>
  <tr valign="top">
    <th align="left">Name</th>
    <th align="left">Type</th>
    <th align="left">Description</th>
  </tr>
  <tr valign="top">
    <td><code>name</code></a></td>
    <td><code>string</code></a></td>
    <td>A unique name that will be used as the <a href="https://developers.google.com/analytics/devguides/collection/analyticsjs/field-reference#eventCategory"><code>eventCategory</code></a> value for media query change events.</td>
  </tr>
  <tr valign="top">
    <td><code>dimensionIndex</code></a></td>
    <td><code>number</code></a></td>
    <td>The index of the custom dimension <a href="https://support.google.com/analytics/answer/2709829">created in Google Analytics</a>.</td>
  </tr>
  <tr valign="top">
    <td><code>items</code></a></td>
    <td><code>Array</code></a></td>
    <td>An array of <code>item</code> objects. See the <a href="#the-item-object"><code>item</code></a> object description for property details.</td>
  </tr>
</table>

### The `item` object

The `item` object allows you to specify what media query values are relevant within each `definition` group.

<table>
  <tr valign="top">
    <th align="left">Name</th>
    <th align="left">Type</th>
    <th align="left">Description</th>
  </tr>
  <tr valign="top">
    <td><code>name</code></a></td>
    <td><code>string</code></a></td>
    <td>The value that will be set on the custom dimension at the specified index.</td>
  </tr>
  <tr valign="top">
    <td><code>media</code></a></td>
    <td><code>string</code></a></td>
    <td>The media query value to test for a match.</td>
  </tr>
</table>

Note: if multiple `media` values match at the same time, the one specified later in the `items` array will take precedence.

## Default field values

The `mediaQueryTracker` plugin sets the following default field values on all hits it sends. To customize these values, use one of the [options](#options) described above.

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
    <td><code>definition.name</code></a></td>
  </tr>
  <tr valign="top">
    <td><a href="https://developers.google.com/analytics/devguides/collection/analyticsjs/field-reference#eventAction"><code>eventAction</code></a></td>
    <td><code>'change'</code></a></td>
  </tr>
  <tr valign="top">
    <td><a href="https://developers.google.com/analytics/devguides/collection/analyticsjs/field-reference#eventLabel"><code>eventLabel</code></a></td>
    <td><code>options.changeTemplate()</code></td>
  </tr>
</table>

**Note:** the reference to `definition` in the table above refers to the definition the changing media value is defined in. The reference to `options` refers to passed configuration [options](#options).

## Methods

The following table lists all methods for the `mediaQueryTracker` plugin:

<table>
  <tr valign="top">
    <th align="left">Name</th>
    <th align="left">Description</th>
  </tr>
  <tr valign="top">
    <td><code>remove</code></a></td>
    <td>Removes the <code>mediaQueryTracker</code> plugin from the specified tracker, removes all media query listeners, and restores all modified tasks to their original state prior to the plugin being required.</td>
  </tr>
</table>

For details on how `analytics.js` plugin methods work and how to invoke them, see [calling plugin methods](https://developers.google.com/analytics/devguides/collection/analyticsjs/using-plugins#calling_plugin_methods) in the `analytics.js` documentation.

## Example

### Basic usage

This example requires the `mediaQueryTracker` plugin and customizes it to track breakpoint, resolution, and orientation media query data:

```js
ga('require', 'mediaQueryTracker', {
  definitions: [
    {
      name: 'Breakpoint',
      dimensionIndex: 1,
      items: [
        {name: 'sm', media: 'all'},
        {name: 'md', media: '(min-width: 30em)'},
        {name: 'lg', media: '(min-width: 48em)'}
      ]
    },
    {
      name: 'Resolution',
      dimensionIndex: 2,
      items: [
        {name: '1x',   media: 'all'},
        {name: '1.5x', media: '(min-resolution: 144dpi)'},
        {name: '2x',   media: '(min-resolution: 192dpi)'}
      ]
    },
    {
      name: 'Orientation',
      dimensionIndex: 3,
      items: [
        {name: 'landscape', media: '(orientation: landscape)'},
        {name: 'portrait',  media: '(orientation: portrait)'}
      ]
    }
  ]
});
```

### Customizing the change template and timeout

This code updates the change template to only report the new media value in the event hit. It also increases the debounce timeout amount for change events, so rapid changes have more time to settle before being reported:

```js
ga('require', 'mediaQueryTracker', {
  definitions: [
    {
      name: 'Breakpoint',
      dimensionIndex: 1,
      items: [
        {name: 'sm', media: 'all'},
        {name: 'md', media: '(min-width: 30em)'},
        {name: 'lg', media: '(min-width: 48em)'}
      ]
    }
  ],
  changeTemplate: function(newValue, oldValue) {
    return newValue;
  },
  changeTimeout: 3000
});
```
