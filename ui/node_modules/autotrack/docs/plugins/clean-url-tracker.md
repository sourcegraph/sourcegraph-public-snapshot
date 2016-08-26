# `cleanUrlTracker`

This guide explains what the `cleanUrlTracker` plugin is and how to integrate it into your `analytics.js` tracking implementation.

## Overview

When viewing your most visited pages in Google Analytics, it's not uncommon to see multiple different URL paths that reference the same page on your site. The following report table is a good example of this and the frustrating situation many users find themselves in today:

<table>
  <tr valign="top">
    <th align="left">Page</th>
    <th align="left">Pageviews</th>
  </tr>
  <tr valign="top">
    <td>/contact</td>
    <td>967</td>
  </tr>
  <tr valign="top">
    <td>/contact/</td>
    <td>431</td>
  </tr>
  <tr valign="top">
    <td>/contact?hl=en</td>
    <td>67</td>
  </tr>
  <tr valign="top">
    <td>/contact/index.html</td>
    <td>32</td>
  </tr>
</table>

To prevent this problem, it's best to settle on a single, canonical URL path for each page you want to track, and only ever send the canonical version to Google Analytics.

The `cleanUrlTracker` plugin helps you do this. It lets you specify a preference for whether or not to include extraneous parts of the URL path, and updates all URLs accordingly.

### How it works

The `cleanUrlPlugin` works by intercepting each hit as it's being sent and modifying the [`page`](https://developers.google.com/analytics/devguides/collection/analyticsjs/field-reference#page) field based on the rules specified by the configuration [options](#options).

If no `page` field exists, one is created based on the URL path from the [`location`](https://developers.google.com/analytics/devguides/collection/analyticsjs/field-reference#location) field.

**Note:** while the `cleanUrlTracker` plugin does modify the `page` field value for each hit, it never modifies the `location` field. This allows campaign and site search data encoded in the full URL to be preserved.

## Usage

To enable the `cleanUrlTracker` plugin, run the [`require`](https://developers.google.com/analytics/devguides/collection/analyticsjs/using-plugins) command, specify the plugin name `'cleanUrlTracker'`, and pass in the configuration options you want to set:

```js
ga('require', 'cleanUrlTracker', options);
```

## Options

The following table outlines all possible configuration options for the `cleanUrlTracker` plugin. If any of the options has a default value, the default is explicitly stated:

<table>
  <tr valign="top">
    <th align="left">Name</th>
    <th align="left">Type</th>
    <th align="left">Default</th>
  </tr>
  <tr valign="top">
    <td><code>stripQuery</code></a></td>
    <td><code>boolean</code></a></td>
    <td>
      When <code>true</code>, the query string portion of the URL will be removed.<br>
      <strong>Default:</strong> <code>false</code>
    </td>
  </tr>
  <tr valign="top">
    <td><code>queryDimensionIndex</code></a></td>
    <td><code>number</code></a></td>
    <td>
      There are cases where you want to strip the query string from the URL, but you still want to record what query string was originally there, so you can report on those values separately. You can do this by creating a new <a href="https://support.google.com/analytics/answer/2709829">custom dimension</a> in Google Analytics. Set the dimension's <a href="https://support.google.com/analytics/answer/2709828#example-hit">scope</a> to "hit", and then set the index of the newly created dimension as the <code>queryDimensionIndex</code> option. Once set, the stripped query string will be set on the custom dimension at the specified index.
    </td>
  </tr>
  <tr valign="top">
    <td><code>indexFilename</code></a></td>
    <td><code>string</code></a></td>
    <td>
      When set, the <code>indexFilename</code> value will be stripped from the end of a URL. If your server supports automatically serving index files, you should set this to whatever value your server uses (usually <code>'index.html'</code>).
    </td>
  </tr>
  <tr valign="top">
    <td><code>trailingSlash</code></a></td>
    <td><code>string</code></a></td>
    <td>
      When set to <code>'add'</code>, a trailing slash is appended to the end of all URLs (if not already present). When set to <code>'remove'</code>, a trailing slash is removed from the end of all URLs. No action is taken if any other value is used. Note: when using the <code>indexFilename</code> option, index filenames are stripped prior to the trailing slash being added or removed.
    </td>
  </tr>
</table>

## Methods

The following table lists all methods for the `cleanUrlTracker` plugin:

<table>
  <tr valign="top">
    <th align="left">Name</th>
    <th align="left">Description</th>
  </tr>
  <tr valign="top">
    <td><code>remove</code></a></td>
    <td>Removes the <code>cleanUrlTracker</code> plugin from the specified tracker and restores all modified tasks to their original state prior to the plugin being required.</td>
  </tr>
</table>

For details on how `analytics.js` plugin methods work and how to invoke them, see [calling plugin methods](https://developers.google.com/analytics/devguides/collection/analyticsjs/using-plugins#calling_plugin_methods) in the `analytics.js` documentation.

## Example

Given the four URL paths shown in the table at the beginning of this guide, the following `cleanUrlTracker` configuration would ensure that only the URL path `/contact` ever appears in your reports (assumes you've created a custom dimension for the query at index 1):

```js
ga('require', 'cleanUrlTracker', {
  stripQuery: true,
  queryDimensionIndex: 1,
  indexFilename: 'index.html',
  trailingSlash: 'remove'
});
```

And given those four URLs, the following fields would be sent to Google Analytics for each respective hit:

```
[1] {
      "location": "/contact",
      "page": "/contact"
    }

[2] {
      "location": "/contact/",
      "page": "/contact"
    }

[3] {
      "location": "/contact?hl=en",
      "page": "/contact"
      "dimension1": "hl=en"
    }

[4] {
      "location": "/contact/index.html",
      "page": "/contact"
    }
```
