[![Circle CI](https://circleci.com/gh/amplitude/Amplitude-Javascript.svg?style=badge&circle-token=80de0dbb7632b2db13f76ccb20a79bbdfc50c215)](https://circleci.com/gh/amplitude/Amplitude-Javascript)

Amplitude-Javascript
====================

This Readme will guide you through using Amplitude's Javascript SDK to track users and events. See our [SDK documentation](https://rawgit.com/amplitude/Amplitude-Javascript/master/documentation/Amplitude.html) for a description of all available SDK methods.

# Setup #
1. If you haven't already, go to http://amplitude.com and register for an account. You will receive an API Key.
2. On every page that uses analytics, paste the following Javascript code between the `<head>` and `</head>` tags:

    ```html
        <script type="text/javascript">
          (function(e,t){var n=e.amplitude||{_q:[]};var r=t.createElement("script");r.type="text/javascript";
          r.async=true;r.src="https://d24n15hnbwhuhn.cloudfront.net/libs/amplitude-2.12.1-min.gz.js";
          r.onload=function(){e.amplitude.runQueuedFunctions()};var s=t.getElementsByTagName("script")[0];
          s.parentNode.insertBefore(r,s);function i(e,t){e.prototype[t]=function(){this._q.push([t].concat(Array.prototype.slice.call(arguments,0)));
          return this}}var o=function(){this._q=[];return this};var a=["add","append","clearAll","prepend","set","setOnce","unset"];
          for(var u=0;u<a.length;u++){i(o,a[u])}n.Identify=o;var c=function(){this._q=[];return this;
          };var p=["setProductId","setQuantity","setPrice","setRevenueType","setEventProperties"];
          for(var l=0;l<p.length;l++){i(c,p[l])}n.Revenue=c;var d=["init","logEvent","logRevenue","setUserId","setUserProperties","setOptOut","setVersionName","setDomain","setDeviceId","setGlobalUserProperties","identify","clearUserProperties","setGroup","logRevenueV2","regenerateDeviceId"];
          function v(e){function t(t){e[t]=function(){e._q.push([t].concat(Array.prototype.slice.call(arguments,0)));
          }}for(var n=0;n<d.length;n++){t(d[n])}}v(n);e.amplitude=n})(window,document);

          amplitude.init("YOUR_API_KEY_HERE");
        </script>
    ```

    Note: if you are using [RequireJS](http://requirejs.org/), follow these [alternate instructions](https://github.com/amplitude/Amplitude-Javascript#loading-with-requirejs) for Step 2.

3. Replace `YOUR_API_KEY_HERE` with the API Key given to you.
4. To track an event anywhere on the page, call:

    ```javascript
    amplitude.logEvent('EVENT_IDENTIFIER_HERE');
    ```

5. Events are uploaded immediately and saved to the browser's local storage until the server confirms the upload. After calling logEvent in your app, you will immediately see data appear on Amplitude.

# Tracking Events #

It's important to think about what types of events you care about as a developer. You should aim to track between 20 and 200 types of events on your site. Common event types are actions the user initiates (such as pressing a button) and events you want the user to complete (such as filling out a form, completing a level, or making a payment).

Here are some resources to help you with your instrumentation planning:
  * [Event Tracking Quick Start Guide](https://amplitude.zendesk.com/hc/en-us/articles/207108137).
  * [Event Taxonomy and Best Practices](https://amplitude.zendesk.com/hc/en-us/articles/211988918).

Having large amounts of distinct event types, event properties and user properties, however, can make visualizing and searching of the data very confusing. By default we only show the first:
  * 1000 distinct event types
  * 2000 distinct event properties
  * 1000 distinct user properties

Anything past the above thresholds will not be visualized. **Note that the raw data is not impacted by this in any way, meaning you can still see the values in the raw data, but they will not be visualized on the platform.** We have put in very conservative estimates for the event and property caps which we don’t expect to be exceeded in any practical use case. If you feel that your use case will go above those limits please reach out to support@amplitude.com.

# Settings Custom User IDs #

If your app has its own login system that you want to track users with, you can call `setUserId` at any time:

```javascript
amplitude.setUserId('USER_ID_HERE');
```

You can also add the user ID as an argument to the `init` call:

```javascript
amplitude.init('YOUR_API_KEY_HERE', 'USER_ID_HERE');
```

### Logging Out and Anonymous Users ###

A user's data will be merged on the backend so that any events up to that point from the same browser will be tracked under the same user. Note: if a user logs out, or you want to log the events under an anonymous user, you need to do 2 things: 1) set the userId to `null` 2) regenerate a new deviceId. After doing that, events coming from the current user will appear as a brand new user in Amplitude dashboards. Note if you choose to do this, then you won't be able to see that the 2 users were using the same browser/device.

```javascript
amplitude.setUserId(null); // not string 'null'
amplitude.regenerateDeviceId();
```

# Setting Event Properties #

You can attach additional data to any event by passing a Javascript object as the second argument to `logEvent`. The Javascript object should be in the form of key + value pairs that can be JSON serialized. The keys should be string values. The values can be booleans, strings, numbers, arrays of strings/numbers/booleans, nested Javascript objects, and errors (note you cannot nest arrays or Javascript objects inside array values). The SDK will validate the event properties that you set and will log any errors or warnings to console if there are any issues. Here is an example:

```javascript
var eventProperties = {};
eventProperties.key = 'value';
amplitude.logEvent('EVENT_IDENTIFIER_HERE', eventProperties);
```

Alternatively, you can set multiple event properties like this:
```javascript
var eventProperties = {
    'color': 'blue',
    'age': 20,
    'key': 'value'
};
amplitude.logEvent('EVENT_IDENTIFIER_HERE', eventProperties);
```

# User Properties and User Property Operations #

The SDK supports the operations `set`, `setOnce`, `unset`, and `add` on individual user properties. The operations are declared via a provided `Identify` interface. Multiple operations can be chained together in a single `Identify` object. The `Identify` object is then passed to the Amplitude client to send to the server. The results of the operations will be visible immediately in the dashboard, and take effect for events logged after.

1. `set`: this sets the value of a user property.

    ```javascript
    var identify = new amplitude.Identify().set('gender', 'female').set('age', 20);
    amplitude.identify(identify);
    ```

2. `setOnce`: this sets the value of a user property only once. Subsequent `setOnce` operations on that user property will be ignored. In the following example, `sign_up_date` will be set once to `08/24/2015`, and the following setOnce to `09/14/2015` will be ignored:

    ```javascript
    var identify = new amplitude.Identify().setOnce('sign_up_date', '08/24/2015');
    amplitude.identify(identify);

    var identify = new amplitude.Identify().setOnce('sign_up_date', '09/14/2015');
    amplitude.identify(identify);
    ```

3. `unset`: this will unset and remove a user property.

    ```javascript
    var identify = new amplitude.Identify().unset('gender').unset('age');
    amplitude.identify(identify);
    ```

4. `add`: this will increment a user property by some numerical value. If the user property does not have a value set yet, it will be initialized to 0 before being incremented.

    ```javascript
    var identify = new amplitude.Identify().add('karma', 1).add('friends', 1);
    amplitude.identify(identify);
    ```

5. `append`: this will append a value or values to a user property. If the user property does not have a value set yet, it will be initialized to an empty list before the new values are appended. If the user property has an existing value and it is not a list, it will be converted into a list with the new value appended.

    ```javascript
    var identify = new amplitude.Identify().append('ab-tests', 'new-user-test').append('some_list', [1, 2, 3, 4, 'values']);
    amplitude.identify(identify);
    ```

6. `prepend`: this will prepend a value or values to a user property. Prepend means inserting the value(s) at the front of a given list. If the user property does not have a value set yet, it will be initialized to an empty list before the new values are prepended. If the user property has an existing value and it is not a list, it will be converted into a list with the new value prepended.

    ```javascript
    var identify = new amplitude.Identify().prepend('ab-tests', 'new-user-test').prepend('some_list', [1, 2, 3, 4, 'values']);
    amplitude.identify(identify);
    ```

Note: if a user property is used in multiple operations on the same `Identify` object, only the first operation will be saved, and the rest will be ignored. In this example, only the set operation will be saved, and the add and unset will be ignored:

```javascript
var identify = new amplitude.Identify()
    .set('karma', 10)
    .add('karma', 1)
    .unset('karma');
amplitude.identify(identify);
```

### Arrays in User Properties ###

The SDK supports arrays in user properties. Any of the user property operations above (with the exception of `add`) can accept a Javascript array. You can directly `set` arrays, or use `append` to generate an array.

```javascript
var identify = new amplitude.Identify()
    .set('colors', ['rose', 'gold'])
    .append('ab-tests', 'campaign_a')
    .append('existing_list', [4, 5]);
amplitude.identify(identify);
```

### Setting Multiple Properties with `setUserProperties` ###

You may use `setUserProperties` shorthand to set multiple user properties at once. This method is simply a wrapper around `Identify.set` and `identify`.

```javascript
var userProperties = {
    gender: 'female',
    age: 20
};
amplitude.setUserProperties(userProperties);
```

### Clearing User Properties ###

You may use `clearUserProperties` to clear all user properties at once. Note: the result is irreversible!

```javascript
amplitude.clearUserProperties();
```

# Tracking Revenue #

The preferred method of tracking revenue for a user now is to use `logRevenueV2()` in conjunction with the provided `Revenue` interface. `Revenue` instances will store each revenue transaction and allow you to define several special revenue properties (such as revenueType, productId, etc) that are used in Amplitude dashboard's Revenue tab. You can now also add event properties to the revenue event, via the eventProperties field. These `Revenue` instance objects are then passed into `logRevenueV2` to send as revenue events to Amplitude servers. This allows us to automatically display data relevant to revenue on the Amplitude website, including average revenue per daily active user (ARPDAU), 1, 7, 14, 30, 60, and 90 day revenue, lifetime value (LTV) estimates, and revenue by advertising campaign cohort and daily/weekly/monthly cohorts.

Each time a user generates revenue, you create a `Revenue` object and fill out the revenue properties:
```javascript
var revenue = new amplitude.Revenue().setProductId('com.company.productId').setPrice(3.99).setQuantity(3);
amplitude.logRevenueV2(revenue);
```

`productId` and `price` are required fields. `quantity` defaults to 1 if not specified. Each field has a corresponding `set` method (for example `setProductId`, `setQuantity`, etc). This table describes the different fields available:

| Name               | Type       | Description                                                                                              | default |
|--------------------|------------|----------------------------------------------------------------------------------------------------------|---------|
| productId          | String     | Required: an identifier for the product (we recommend something like the Google Play Store product Id)   | null    |
| quantity           | Integer    | Required: the quantity of products purchased. Defaults to 1 if not specified. Revenue = quantity * price | 1       |
| price              | Double     | Required: the price of the products purchased (can be negative). Revenue = quantity * price              | null    |
| revenueType        | String     | Optional: the type of revenue (ex: tax, refund, income)                                                  | null    |
| eventProperties    | Object     | Optional: an object of event properties to include in the revenue event                                  | null    |

Note: the price can be negative, which might be useful for tracking revenue lost, for example refunds or costs. Also note, you can set event properties on the revenue event just as you would with logEvent by passing in a object of string key value pairs. These event properties, however, will only appear in the Event Segmentation tab, not in the Revenue tab.

### Backwards compatibility ###

The existing `logRevenue` methods still work but are deprecated. Fields such as `revenueType` will be missing from events logged with the old methods, so Revenue segmentation on those events will be limited in Amplitude dashboards.

# Opting User Out of Logging #

You can turn off logging for a given user:

```javascript
amplitude.setOptOut(true);
```

No events will be saved or sent to the server while opt out is enabled. The opt out
setting will persist across page loads. Calling

```javascript
amplitude.setOptOut(false);
```

will reenable logging.

# Configuration Options #

You can configure Amplitude by passing an object as the third argument to the `init`:

```javascript
amplitude.init('YOUR_API_KEY_HERE', null, {
    // optional configuration options
    saveEvents: true,
    includeUtm: true,
    includeReferrer: true,
    batchEvents: true,
    eventUploadThreshold: 50
});
```

| option | type | description | default |
|------------|----------|------------------------------------------------------------------------|-----------|
| batchEvents | boolean | If `true`, events are batched together and uploaded only when the number of unsent events is greater than or equal to `eventUploadThreshold` or after `eventUploadPeriodMillis` milliseconds have passed since the first unsent event was logged. | `false` |
| cookieExpiration | number | The number of days after which the Amplitude cookie will expire | 365\*10 (10 years) |
| cookieName | string | Custom name for the Amplitude cookie | 'amplitude_id' |
| deviceId | string | Custom device ID to set. Note this is not recommended unless you really know what you are doing (like if you have your own system for tracking user devices) | Randomly generated UUID |
| domain | string | Custom cookie domain | The top domain of the current page's url |
| eventUploadPeriodMillis | number | Amount of time in milliseconds that the SDK waits before uploading events if `batchEvents` is `true`. | 30\*1000 (30 sec) |
| eventUploadThreshold | number | Minimum number of events to batch together per request if `batchEvents` is `true`. | 30 |
| includeReferrer | boolean | If `true`, captures the `referrer` and `referring_domain` for each session, as well as the user's `initial_referrer` and `initial_referring_domain` via a set once operation. | `false` |
| includeUtm | boolean | If `true`, finds utm parameters in the query string or the __utmz cookie, parses, and includes them as user propeties on all events uploaded. | `false` |
| language | string | Custom language to set | Language determined by browser |
| optOut | boolean | Whether to disable tracking for the current user | `false` |
| platform | string | Custom platform to set | 'Web' |
| saveEvents | boolean | If `true`, saves events to localStorage and removes them upon successful upload.<br><i>NOTE:</i> Without saving events, events may be lost if the user navigates to another page before events are uploaded. | `true` |
| savedMaxCount | number | Maximum number of events to save in localStorage. If more events are logged while offline, old events are removed. | 1000 |
| sessionTimeout | number | Time between logged events before a new session starts in milliseconds | 30\*60\*1000 (30 min) |
| uploadBatchSize | number | Maximum number of events to send to the server per request. | 100 |

# Advanced #
This SDK automatically grabs useful data about the browser, including browser type and operating system version.

### Setting Groups ###

Amplitude supports assigning users to groups, and performing queries such as Count by Distinct on those groups. An example would be if you want to group your users based on what organization they are in by using an orgId. You can designate Joe to be in orgId 10, while Sue is in orgId 15. When performing an event segmentation query, you can then select Count by Distinct orgIds to query the number of different orgIds that have performed a specific event. As long as at least one member of that group has performed the specific event, that group will be included in the count. See our help article on [Count By Distinct]() for more information.

When setting groups you need to define a `groupType` and `groupName`(s). In the above example, 'orgId' is a `groupType`, and the value 10 or 15 is the `groupName`. Another example of a `groupType` could be 'sport' with `groupNames` like 'tennis', 'baseball', etc.

You can use `setGroup(groupType, groupName)` to designate which groups a user belongs to. Note: this will also set the `groupType`: `groupName` as a user property. **This will overwrite any existing groupName value set for that user's groupType, as well as the corresponding user property value.** `groupType` is a string, and `groupName` can be either a string or an array of strings to indicate a user being in multiple groups (for example Joe is in orgId 10 and 16, so the `groupName` would be [10, 16]).

```javascript
amplitude.setGroup('orgId', '15');
amplitude.setGroup('sport', ['soccer', 'tennis']);
```

You can also use `logEventWithGroups` to set event-level groups, meaning the group designation only applies for the specific event being logged and does not persist on the user (unless you explicitly set it with `setGroup`).

```javascript
var eventProperties = {
  'key': 'value'
}

amplitude.logEventWithGroups('initialize_game', eventProperties, {'sport': 'soccer'});
```

### Setting Version Name ###
By default, no version name is set. You can specify a version name to distinguish between different versions of your site by calling `setVersionName`:

```javascript
amplitude.setVersionName('VERSION_NAME_HERE');
```

### Custom Device Ids ###
Device IDs are generated randomly, although you can define a custom device ID setting it as a configuration option or by calling:

```javascript
amplitude.setDeviceId('CUSTOM_DEVICE_ID');
```

**Note: this is not recommended unless you really know what you are doing** (like if you have your own system for tracking user devices). Make sure the deviceId you set is sufficiently unique (we recommend something like a UUID - see `src/uuid.js` for an example of how to generate) to prevent conflicts with other devices in our system.

### Callbacks for LogEvent, Identify, and Redirect ###
You can pass a callback function to logEvent and identify, which will get called after receiving a response from the server:

```javascript
amplitude.logEvent("EVENT_IDENTIFIER_HERE", null, callback_function);
```

```javascript
var identify = new amplitude.Identify().set('key', 'value');
amplitude.identify(identify, callback_function);
```

The status and response body from the server are passed to the callback function, which you might find useful. An example of a callback function which redirects the browser to another site after a response:

```javascript
var callback_function = function(status, response) {
    if (status === 200 && response === 'success') {
        // do something here
    }
    window.location.replace('URL_OF_OTHER_SITE');
};
```

You can also use this to track outbound links on your website. For example you would have a link like this:

```html
<a href="javascript:trackClickLinkA();">Link A</a>
```

And then you would define a function that is called when the link is clicked like this:

```javascript
var trackClickLinkA = function() {
    amplitude.logEvent('Clicked Link A', null, function() {
        window.location='LINK_A_URL';
    });
};
```

In the case that `optOut` is true, then no event will be logged, but the callback will be called. In the case that `batchEvents` is true, if the batch requirements `eventUploadThreshold` and `eventUploadPeriodMillis` are not met when `logEvent` is called, then no request is sent, but the callback is still called. In these cases, the callback will be called with an input status of 0 and response 'No request sent'.

### Init Callbacks ###
You can also pass a callback function to init, which will get called after the SDK finishes its asynchronous loading. *Note: no values are passed to the init callback function*:

```javascript
amplitude.init('YOUR_API_KEY_HERE', 'USER_ID_HERE', null, function() {
  console.log(amplitude.options.deviceId);  // access Amplitude's deviceId after initialization
});
```

### Loading with RequireJS ###
If you are using [RequireJS](http://requirejs.org/) to load your Javascript files, you can also use it to load the Amplitude Javascript SDK script directly instead of using our loading snippet. On every page that uses analytics, paste the following Javascript code between the `<head>` and `</head>` tags:

```html
  <script src='scripts/require.js'></script>  <!-- loading RequireJS -->
  <script>
    require(['https://d24n15hnbwhuhn.cloudfront.net/libs/amplitude-2.12.1-min.gz.js'], function(amplitude) {
      amplitude.init('YOUR_API_KEY_HERE'); // replace YOUR_API_KEY_HERE with your Amplitude api key.
      window.amplitude = amplitude;  // You can bind the amplitude object to window if you want to use it directly.
      amplitude.logEvent('Clicked Link A');
    });
  </script>
```

You can also define the path in your RequireJS configuration like so:
```html
  <script src='scripts/require.js'></script>  <!-- loading RequireJS -->
  <script>
    requirejs.config({
      paths: {
        'amplitude': 'https://d24n15hnbwhuhn.cloudfront.net/libs/amplitude-2.12.1-min.gz'
      }
    });

    require(['amplitude'], function(amplitude) {
      amplitude.init('YOUR_API_KEY_HERE'); // replace YOUR_API_KEY_HERE with your Amplitude api key.
      window.amplitude = amplitude;  // You can bind the amplitude object to window if you want to use it directly.
      amplitude.logEvent('Clicked Link A');
    });
  </script>
  <script>
    require(['amplitude'], function(amplitude) {
      amplitude.logEvent('Page loaded');
    });
  </script>
```
