# `outboundLinkTracker`

This guide explains what the `outboundLinkTracker` plugin is and how to integrate it into your `analytics.js` tracking implementation.

## Overview

When a visitor to your site clicks a link that goes to another page on your site, you can usually see this information in Google Analytics because the page being navigated to will typically send its own pageview. However, if a visitor to your site clicks a link that points to an external domain, you'll never know unless you track that click separately.

The `outboundLinkTracker` plugin automatically detects when links are clicked to sites on different domains and sends an event hit to Google Analytics.

Historically, outbound link tracking has been tricky to implement because most browsers stop executing JavaScript on the current page once a link that requests a new page is clicked. The `outboundLinkTracker` plugin handles these complications for you.

## Usage

To enable the `outboundLinkTracker` plugin, run the [`require`](https://developers.google.com/analytics/devguides/collection/analyticsjs/using-plugins) command, specify the plugin name `'outboundLinkTracker'`, and pass in any configuration options (if any) you wish to set:

```js
ga('require', 'outboundLinkTracker', options);
```

### Determining what is an outbound link

By default a link is considered outbound if the hostname of the URL it's pointing to differs from `location.hostname`. Note that this means links pointing to different subdomains within the same higher-level domain are (by default) still considered outbound. To customize this logic, see `shouldTrackOutboundLink` in the [options](#options) section below.

## Options

The following table outlines all possible configuration options for the `outboundLinkTracker` plugin. If any of the options has a default value, the default is explicitly stated:

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
      A list of events to listen for on links. Since it's possible to navigate to a link without generating a <code>click</code> (e.g. right-clicking generates a <code>contextmenu</code> event), you can customize this option to track additional events.<br>
      <strong>Default:</strong> <code>['click']</code>
    </td>
  </tr>
  <tr valign="top">
    <td><code>linkSelector</code></a></td>
    <td><code>string</code></a></td>
    <td>
      A selector used to identify links to listen for events on.<br>
      <strong>Default:</strong> <code>'a'</code>
    </td>
  </tr>
  <tr valign="top">
    <td><code>shouldTrackOutboundLink</code></a></td>
    <td><code>Function</code></a></td>
    <td>
      A function that returns <code>true</code> if the link in question should be considered an outbound link. The function is invoked with the link element as its first argument and a <code>parseUrl</code> utility function (which returns a <a href="https://developer.mozilla.org/en-US/docs/Web/API/Location"><code>Location</code></a>-like object) as its second argument.<br>
      <strong>Default:</strong>
<pre>function shouldTrackOutboundLink(link, parseUrl) {
  var url = parseUrl(link.href);
  return url.hostname != location.hostname &amp;&amp;
      url.protocol.slice(0, 4) == 'http';
}</pre>
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

The `outboundLinkTracker` plugin sends hits with the following values. To customize these values, use one of the [options](#options) described above.

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
    <td><code>'Outbound Link'</code></a></td>
  </tr>
  <tr valign="top">
    <td><a href="https://developers.google.com/analytics/devguides/collection/analyticsjs/field-reference#eventAction"><code>eventAction</code></a></td>
    <td><code>event.type</code></a></td>
  </tr>
  <tr valign="top">
    <td><a href="https://developers.google.com/analytics/devguides/collection/analyticsjs/field-reference#eventLabel"><code>eventLabel</code></a></td>
    <td><code>link.href</code></td>
  </tr>
</table>

**Note:** the reference to `form` in the table above refers to the `<form>` element being clicked. The reference to `event` refers to the event being dispatched by the user interaction.

## Methods

The following table lists all methods for the `outboundLinkTracker` plugin:

<table>
  <tr valign="top">
    <th align="left">Name</th>
    <th align="left">Description</th>
  </tr>
  <tr valign="top">
    <td><code>remove</code></a></td>
    <td>Removes the <code>outboundLinkTracker</code> plugin from the specified tracker, removes all event listeners from the DOM, and restores all modified tasks to their original state prior to the plugin being required.</td>
  </tr>
</table>

For details on how `analytics.js` plugin methods work and how to invoke them, see [calling plugin methods](https://developers.google.com/analytics/devguides/collection/analyticsjs/using-plugins#calling_plugin_methods) in the `analytics.js` documentation.

## Examples

### Basic usage

```js
ga('require', 'outboundLinkTracker');
```

```html
<a href="https://example.com">...</a>
```

### Customizing the link selector

This code only tracks link elements with the `js-track-clicks` class.

```js
ga('require', 'outboundLinkTracker', {
  linkSelector: '.js-track-clicks'
});
```

```html
<a class="js-track-clicks" href="https://example.com">...</a>
```

### Customizing what is considered an "outbound" link

This code changes the default logic used to determine if a link is "outbound". It updates the logic to only track links that go to the `foo.com` and `bar.com` domains:


```js
ga('require', 'outboundLinkTracker', {
  shouldTrackOutboundLink: function(link, parseUrl) {
    var url = parseUrl(link.href);
    return /(foo|bar)\.com$/.test(url.hostname);
  }
});
```

With the above code, clicks on the following link won't be tracked, even though the link is pointing to an external domain:

```html
<a href="https://example.com">...</a>
```
