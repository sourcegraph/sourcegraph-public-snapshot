# `outboundFormTracker`

This guide explains what the `outboundFormTracker` plugin is and how to integrate it into your `analytics.js` tracking implementation.

## Overview

When a visitor to your site submits a form that goes to another page on your site, you can usually see this information in Google Analytics because the page being navigated to will typically send its own pageview. However, if a visitor to your site submits a form that points to an external domain, you'll never know unless you track that submit separately.

The `outboundFormTracker` plugin automatically detects when forms are submitted to sites on different domains and sends an event hit to Google Analytics.

Historically, outbound form tracking has been tricky to implement because most browsers stop executing JavaScript on the current page once a form that requests a new page is submitted. The `outboundFormTracker` plugin handles these complications for you.

## Usage

To enable the `outboundFormTracker` plugin, run the [`require`](https://developers.google.com/analytics/devguides/collection/analyticsjs/using-plugins) command, specify the plugin name `'outboundFormTracker'`, and pass in any configuration options (if any) you wish to set:

```js
ga('require', 'outboundFormTracker', options);
```

### Determining what is an outbound form

By default a form is considered outbound if the hostname of the URL it's pointing to differs from `location.hostname`. Note that this means forms pointing to different subdomains within the same higher-level domain are (by default) still considered outbound. To customize this logic, see `shouldTrackOutboundForm` in the [options](#options) section below.

## Options

The following table outlines all possible configuration options for the `outboundFormTracker` plugin. If any of the options has a default value, the default is explicitly stated:

<table>
  <tr valign="top">
    <th align="left">Name</th>
    <th align="left">Type</th>
    <th align="left">Description</th>
  </tr>
  <tr valign="top">
    <td><code>formSelector</code></a></td>
    <td><code>string</code></a></td>
    <td>
      A selector used to identify forms to listen for submit events on.<br>
      <strong>Default:</strong> <code>'form'</code>
    </td>
  </tr>
  <tr valign="top">
    <td><code>shouldTrackOutboundForm</code></a></td>
    <td><code>Function</code></a></td>
    <td>
      A function that returns <code>true</code> if the form in question should be considered an outbound form. The function is invoked with the form element as its first argument and a <code>parseUrl</code> utility function (which returns a <a href="https://developer.mozilla.org/en-US/docs/Web/API/Location"><code>Location</code></a>-like object) as its second argument.<br>
      <strong>Default:</strong>
<pre>function shouldTrackOutboundForm(form, parseUrl) {
  var url = parseUrl(form.action);
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

The `outboundFormTracker` plugin sends hits with the following values. To customize these values, use one of the [options](#options) described above.

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
    <td><code>'Outbound Form'</code></a></td>
  </tr>
  <tr valign="top">
    <td><a href="https://developers.google.com/analytics/devguides/collection/analyticsjs/field-reference#eventAction"><code>eventAction</code></a></td>
    <td><code>'submit'</code></a></td>
  </tr>
  <tr valign="top">
    <td><a href="https://developers.google.com/analytics/devguides/collection/analyticsjs/field-reference#eventLabel"><code>eventLabel</code></a></td>
    <td><code>form.action</code></td>
  </tr>
</table>

**Note:** the reference to `form` in the table above refers to the `<form>` element being submitted.

## Methods

The following table lists all methods for the `outboundFormTracker` plugin:

<table>
  <tr valign="top">
    <th align="left">Name</th>
    <th align="left">Description</th>
  </tr>
  <tr valign="top">
    <td><code>remove</code></a></td>
    <td>Removes the <code>outboundFormTracker</code> plugin from the specified tracker, removes all event listeners from the DOM, and restores all modified tasks to their original state prior to the plugin being required.</td>
  </tr>
</table>

For details on how `analytics.js` plugin methods work and how to invoke them, see [calling plugin methods](https://developers.google.com/analytics/devguides/collection/analyticsjs/using-plugins#calling_plugin_methods) in the `analytics.js` documentation.

## Examples

### Basic usage

```js
ga('require', 'outboundFormTracker');
```

```html
<form action="https://example.com">...</form>
```

### Customizing the form selector

This code only tracks form elements with the `js-track-submits` class.

```js
ga('require', 'outboundFormTracker', {
  formSelector: '.js-track-submits'
});
```

```html
<form class="js-track-submits" action="https://example.com">...</form>
```

### Customizing what is considered an "outbound" form

This code changes the default logic used to determine if a form is "outbound". It updates the logic to only track forms that submit to the `foo.com` and `bar.com` domains:


```js
ga('require', 'outboundFormTracker', {
  shouldTrackOutboundForm: function(form, parseUrl) {
    var url = parseUrl(form.action);
    return /(foo|bar)\.com$/.test(url.hostname);
  }
});
```

With the above code, submits from the following form won't be tracked, even though the form is submitting to an external domain:

```html
<form action="https://example.com">...</form>
```
