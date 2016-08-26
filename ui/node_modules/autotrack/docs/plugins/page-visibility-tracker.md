# `pageVisibilityTracker`

This guide explains what the `pageVisibilityTracker` plugin is and how to integrate it into your `analytics.js` tracking implementation.

## Overview

It's becoming increasingly common for users to visit your site, and then leave it open in a browser tab for hours or days. And with rise in popularity of single page applications, some tabs almost never get closed.

Because of this shift, the traditional model of pageviews and sessions simply does not apply in a growing number of cases.

The `pageVisibilityTracker` plugin changes this paradigm by shifting from pageload being the primary indicator to [Page Visibility](https://developer.mozilla.org/en-US/docs/Web/API/Page_Visibility_API). To put that another way, if a user visits your application and interacts with it, switches to a different tab, and then comes back to your application hours or days later. Whether they reloaded the page or not, with the `pageVisibilityTracker` enabled, this will be considered a new pageview and a new session.

### How it works

The `pageVisibilityTracker` plugin listens for [`visibilitychange`](https://developer.mozilla.org/en-US/docs/Web/Events/visibilitychange) events on the current page and sends hits to Google Analytics capturing how long the page was in each state. It also programmatically starts new sessions and sends new pageviews when the visibility state changes from hidden to visible (if the previous session has timed out).

### Impact on session and pageview counts

When using the `pageVisibilityTracker` plugin, you'll probably notice an increase in your session and pageview counts. This is not an error, the reality is your current implementation (based just on pageloads) is likely underreporting.

## Usage

To enable the `pageVisibilityTracker` plugin, run the [`require`](https://developers.google.com/analytics/devguides/collection/analyticsjs/using-plugins) command, specify the plugin name `'pageVisibilityTracker'`, and pass in any configuration options (if any) you wish to set:

```js
ga('require', 'pageVisibilityTracker', options);
```

## Options

The following table outlines all possible configuration options for the `pageVisibilityTracker` plugin. If any of the options has a default value, the default is explicitly stated:

<table>
  <tr valign="top">
    <th align="left">Name</th>
    <th align="left">Type</th>
    <th align="left">Description</th>
  </tr>
  <tr valign="top">
    <td><code>sessionTimeout</code></a></td>
    <td><code>number</code></a></td>
    <td>
      The <a href="https://support.google.com/analytics/answer/2795871">session timeout</a> amount (in minutes) of the Google Analytics property. By default this value is 30 minutes, which is the same default used for new Google Analytics properties. The value set for this plugin should always be the same as the property setting in Google Analytics.<br>
      <strong>Default:</strong> <code>30</code>
  </td>
  </tr>
  <tr valign="top">
    <td><code>changeTemplate</code></a></td>
    <td><code>Function</code></a></td>
    <td>
      A function that accepts the old and new values and returns a string to be used as the <a href="https://developers.google.com/analytics/devguides/collection/analyticsjs/field-reference#eventLabel"><code>eventLabel</code></a> field for change events.<br>
      <strong>Default:</strong>
<pre><code>function(oldValue, newValue) {
  return oldValue + ' => ' + newValue;
};</code></pre></td>
    </td>
  </tr>
  <tr valign="top">
    <td><code>hiddenMetricIndex</code></a></td>
    <td><code>number</code></a></td>
    <td>If set, a <a href="https://support.google.com/analytics/answer/2709828">custom metric</a> at the index provided is sent when the page's visibility state changes from hidden to visible. The metric value is the amount of time (in seconds) the page was in the hidden state.</td>
  </tr>
  <tr valign="top">
    <td><code>visibleMetricIndex</code></a></td>
    <td><code>number</code></a></td>
    <td>If set, a <a href="https://support.google.com/analytics/answer/2709828">custom metric</a> at the index provided is sent when the page's visibility state changes from visible to hidden. The metric value is the amount of time (in seconds) the page was in the visible state.</td>
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

### Visibility state change events

The `pageVisibilityTracker` plugin sets the following default field values on event hits it sends. To customize these values, use one of the [options](#options) described above.

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
    <td><code>'Page Visibility'</code></a></td>
  </tr>
  <tr valign="top">
    <td><a href="https://developers.google.com/analytics/devguides/collection/analyticsjs/field-reference#eventAction"><code>eventAction</code></a></td>
    <td><code>'change'</code></a></td>
  </tr>
  <tr valign="top">
    <td><a href="https://developers.google.com/analytics/devguides/collection/analyticsjs/field-reference#eventLabel"><code>eventLabel</code></a></td>
    <td><code>options.changeTemplate()</code></td>
  </tr>
  <tr valign="top">
    <td><a href="https://developers.google.com/analytics/devguides/collection/analyticsjs/field-reference#eventValue"><code>eventValue</code></a></td>
    <td>The elapsed time since the session start or the previous change event.</td>
  </tr>
  <tr valign="top">
    <td><a href="https://developers.google.com/analytics/devguides/collection/analyticsjs/field-reference#nonInteraction"><code>nonInteraction</code></a></td>
    <td><code>true</code> if the visibility state is <code>hidden</code>, <code<>false</code<> otherwise.</td>
  </tr>
</table>

**Note:** the reference to `options` refers to passed configuration [options](#options).

### New pageview events

If the page's visibility state changes from `hidden` to `visible` and the session has timed out. A pageview is sent instead of a change event.

<table>
  <tr valign="top">
    <th align="left">Field</th>
    <th align="left">Value</th>
  </tr>
  <tr valign="top">
    <td><a href="https://developers.google.com/analytics/devguides/collection/analyticsjs/field-reference#hitType"><code>hitType</code></a></td>
    <td><code>'pageview'</code></td>
  </tr>
</table>


## Improving session duration calculations

While not a feature enabled by default, the `pageVisibilityTracker` plugin is also capable of dramatically improving the accuracy of session duration calculations.

Session duration in Google Analytics is defined as the amount of time between the first interaction hit and the last interaction hit within a single session. That means that if a session only has one interaction hit, the duration is zero, even if the user was on the page for hours!

The `pageVisibilityTracker` plugin defaults to sending visibility state change events as [non-interaction](https://developers.google.com/analytics/devguides/collection/analyticsjs/events#non-interaction_events) hits (for the `hidden` event), but you can customize this to make all hits interactive and since most sessions will contain visibility state change events, your sessions durations will become much more accurate.

The downside of this approach is it will alter your bounce rate. However, for many users, this trade-off is worth it since bounce rate can be calculated in other ways (e.g. with a custom segment).

The examples section below includes a code sample showing [how to make all events interaction events](#ensuring-all-events-are-interaction-events).

## Methods

The following table lists all methods for the `pageVisibilityTracker` plugin:

<table>
  <tr valign="top">
    <th align="left">Name</th>
    <th align="left">Description</th>
  </tr>
  <tr valign="top">
    <td><code>remove</code></a></td>
    <td>Removes the <code>pageVisibilityTracker</code> plugin from the specified tracker, removes all event listeners from the DOM, and restores all modified tasks to their original state prior to the plugin being required.</td>
  </tr>
</table>

For details on how `analytics.js` plugin methods work and how to invoke them, see [calling plugin methods](https://developers.google.com/analytics/devguides/collection/analyticsjs/using-plugins#calling_plugin_methods) in the `analytics.js` documentation.

## Examples

### Ensuring all events are interaction events

This example uses the `fieldsObj` option to set the `nonInteraction` value for all hits to `null`, which overrides the plugin's default `nonInteraction` value of `true` for hidden events. Making this change will dramatically [improve session duration calculations](#improving-session-duration-calculations) as described above.

```js
ga('require', 'pageVisibilityTracker', {
  fieldsObj: {
    nonInteraction: null
  }
});
```

### Setting custom metrics to track time spent in the hidden and visible states

If you want to track the total (or average) time a user spends in each visibility state via a [custom metric](https://support.google.com/analytics/answer/2709828) (by default it's tracked as the `eventValue`), you can use the hidden and visible metric indexes.

```js
ga('require', 'pageVisibilityTracker', {
  hiddenMetricIndex: 1,
  visibleMetricIndex: 2,
});
```

**Note:** this requires [creating custom metrics](https://support.google.com/analytics/answer/2709829) in your Google Analytics property settings.

