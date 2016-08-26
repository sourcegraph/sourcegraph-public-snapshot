# Changelog

## 3.5.1
* BUGFIX: Fix non-fatals crashing React Native plugin unless `shouldSendCallback` is specified. See: https://github.com/getsentry/raven-js/pull/694

## 3.5.0
* NEW: Can now disable automatic collection of breadcrumbs via `autoBreadcrumbs` config option. See: https://github.com/getsentry/raven-js/pull/686
* NEW: Can now configure max number of breadcrumbs to collect via `maxBreadcrumbs`. See: https://github.com/getsentry/raven-js/pull/685
* NEW: Added Vue.js plugin. See: https://github.com/getsentry/raven-js/pull/688
* CHANGE: Raven.js now collects 100 breadcrumbs by default. See: https://github.com/getsentry/raven-js/pull/685
* CHANGE: React Native plugin now also normalizes paths from CodePush. See: https://github.com/getsentry/raven-js/pull/683

## 3.4.1
* BUGFIX: Fix exception breadcrumbs having "undefined" for exception value. See: https://github.com/getsentry/raven-js/pull/681

## 3.4.0
* CHANGE: React Native plugin now stores errors in AsyncStorage and sends error data on app init. See: https://github.com/getsentry/raven-js/pull/626
* BUGFIX: React Native path normalization regex has been updated. See: https://github.com/getsentry/raven-js/pull/666
* BUGFIX: Angular 1 plugin now extracts errors from minified exception strings. See: https://github.com/getsentry/raven-js/pull/667

## 3.3.0
* NEW: Can now specify `environment` configuration option. See: https://github.com/getsentry/raven-js/pull/661
* CHANGE: Raven.js now serializes data payload w/ json-stringify-safe to avoid circular references. See: https://github.com/getsentry/raven-js/pull/652
* BUGFIX: Angular 1.x plugin no longer clobbers user-specified `dataCallback`. See: https://github.com/getsentry/raven-js/pull/658

## 3.2.1
* BUGFIX: Fixed error when manually calling captureException with Error objects w/ maxMessageLength > 0. See: https://github.com/getsentry/raven-js/pull/647
* BUGFIX: Fixed TypeScript language declaration file for compatibility w/ Webpack loaders. See: https://github.com/getsentry/raven-js/pull/645
* BUGFIX: Fixed Raven dropping file:/// frames from Phantom 1.x. See: https://github.com/getsentry/raven-js/pull/642

## 3.2.0
* CHANGE: Callbacks set via `setDataCallback`, `setShouldSendCallback` now receive any prior-set callback as the 2nd argument. See: https://github.com/getsentry/raven-js/pull/636
* CHANGE: Raven.js no longer passes a 'message' interface for exceptions. See: https://github.com/getsentry/raven-js/pull/632
* CHANGE: Log level now recorded for "sentry" breadcrumbs. See: https://github.com/getsentry/raven-js/pull/633

## 3.1.1
* BUGFIX: Fix message truncation occurring before dataCallback is invoked. See: https://github.com/getsentry/raven-js/issues/605
* BUGFIX: Fix pushState error in Chrome Apps. See: https://github.com/getsentry/raven-js/issues/601
* BUGFIX: Fix error in addEventListener call affecting very old Firefox versions. See: https://github.com/getsentry/raven-js/issues/603

## 3.1.0
* NEW: Added TypeScript declaration file for compatibility with TypeScript projects. See: https://github.com/getsentry/raven-js/pull/610

## 3.0.5
* BUGFIX: Fix breadcrumb instrumentation failing in IE8. See: https://github.com/getsentry/raven-js/issues/594

## 3.0.4
* BUGFIX: Navigation breadcrumbs now include query strings and document fragment (#). See: https://github.com/getsentry/raven-js/issues/573
* BUGFIX: Remove errant `throw` call in _makeRequest affecting some Raven configs. See: https://github.com/getsentry/raven-js/pull/572

## 3.0.3
* BUGFIX: Fix pushState instrumentation breaking on non-string URL args. See: https://github.com/getsentry/raven-js/issues/569  

## 3.0.2
* BUGFIX: Fix XMLHttpRequest.prototype.open breaking on non-string `url` arguments. See: https://github.com/getsentry/raven-js/issues/567

## 3.0.1
* BUGFIX: Fix broken CDN builds. See: https://github.com/getsentry/raven-js/pull/566

## 3.0.0
* NEW: Raven.js now collects breadcrumbs from XMLHttpRequest objects, URL changes (pushState), console log calls, UI clicks, and errors.
* BUGFIX: Fix parsing error messages from Opera Mini. See: https://github.com/getsentry/raven-js/pull/554
* REMOVED: Fallback Image transport (HTTP GET) has been removed. See: https://github.com/getsentry/raven-js/pull/545
* REMOVED: TraceKit client-side source fetching has been removed. See: https://github.com/getsentry/raven-js/pull/542

## 2.3.0
* NEW: `pathStrip` option now available in React Native plugin. See: https://github.com/getsentry/raven-js/pull/515
* BUGFIX: Handle stacks from internal exceptions sometimes thrown by Firefox. See: https://github.com/getsentry/raven-js/pull/536
* BUGFIX: Better error message strings in browsers w/ limited onerror implementations. See: https://github.com/getsentry/raven-js/pull/538

## 2.2.1
* BUGFIX: Fix HTTP requests not sending with React Native on Android devices. See: https://github.com/getsentry/raven-js/issues/526
* BUGFIX: Raven.js now captures stack traces caused by Firefox internal errors. See: https://github.com/getsentry/raven-js/pull/528

## 2.2.0
* NEW: `allowSecretKey` configuration option. See: https://github.com/getsentry/raven-js/pull/525
* NEW: Console plugin can be configured to capture specific log levels. See: https://github.com/getsentry/raven-js/pull/514
* CHANGE: React Native plugin now calls default exception handler. See: https://github.com/getsentry/raven-js/pull/492
* CHANGE: React Native plugin now uses HTTP POST transport. See: https://github.com/getsentry/raven-js/pull/494
* BUGFIX: Fix Raven throwing exception when run via Webdriver. See: https://github.com/getsentry/raven-js/issues/495

## 2.1.1
* BUGFIX: Fixed IE8 regression introduced in 2.1.0. See: https://github.com/getsentry/raven-js/issues/498
* BUGFIX: Fixed initialization error when run via Selenium. See: https://github.com/getsentry/raven-js/issues/495

## 2.1.0
* BUGFIX: Fixed Raven.js rejecting frames w/ blob URLs. See: https://github.com/getsentry/raven-js/issues/463
* BUGFIX: Fixed plugin files not consumable without module loader. See: https://github.com/getsentry/raven-js/issues/446
* BUGFIX: Fixed bug in console.js plugin where `level` wasn't passed. See: https://github.com/getsentry/raven-js/pull/474
* BUGFIX: Fixed broken debug logging in IE9 and below. See: https://github.com/getsentry/raven-js/pull/473
* BUGFIX: Fixed `XMLHttpRequest` wrapper not capturing all event handlers. See: https://github.com/getsentry/raven-js/issues/453
* CHANGE: `Raven.uninstall` now restores original builtin functions (e.g. setTimeout). See: https://github.com/getsentry/raven-js/issues/228
* CHANGE: `maxMessageLength` now defaults to 0 (no limit). See: https://github.com/getsentry/raven-js/pull/441
* NEW: New `stackTraceLimit` config option (default 50 in supported browsers). See: https://github.com/getsentry/raven-js/pull/419/files
* NEW: `Raven.showReportDialog` (experimental). See: https://github.com/getsentry/raven-js/pull/456

## 2.0.5
* BUGFIX: Fixed exception thrown by React Native plugin. See: https://github.com/getsentry/raven-js/issues/468
* BUGFIX: Fixed "pre-built JavaScript" warning when loading Raven.js via Webpack. See: https://github.com/getsentry/raven-js/issues/465

## 2.0.4
* BUGFIX: Fixed bug where Raven.VERSION was not set when required as a CommonJS module.

## 2.0.2
* BUGFIX: Fixed bug where wrapped requestAnimationFrame didn't return callback ID. See: https://github.com/getsentry/raven-js/pull/460

## 2.0.1
* BUGFIX: Fixed bug where unwrapped errors might be suppressed. See: https://github.com/getsentry/raven-js/pull/447

## 2.0.0

* CHANGE: Raven.js now wraps functions passed to timer functions, event listeners, and XMLHttpRequest handlers
* CHANGE: Removed jQuery, Backbone, and native plugins (now handled inside raven.js)
* CHANGE: Default HTTP transport changed from `Image` GET to `XMLHttpRequest` POST (w/ CORS)
* CHANGE: When using CommonJS, plugins are initialized via `Raven.addPlugin(require('raven-js/plugins/ember'))`
* CHANGE: Raven builds are generated using Browserify
* NEW: Integration tests (/test/integration/index.html)

## 1.3.0
* CHANGE: `console` plugin will now send all arguments as an `extra` value. See: https://github.com/getsentry/raven-js/pull/398
* CHANGE: Bump to v7 of the Sentry API spec. This now requires a Sentry 7.7.0+ https://github.com/getsentry/raven-js/pull/403
* CHANGE: Revamp of AngularJS plugin. Please see documentation. See: https://github.com/getsentry/raven-js/pull/405
* CHANGE: `Raven.debug` now defaults to `false`. https://github.com/getsentry/raven-js/commit/dc142b88f0c4953f54cb3754f9015d95ada55ba0
* BUGFIX: `Raven.wrap` now correctly preserves `prototype`. See: https://github.com/getsentry/raven-js/issues/401 and https://github.com/getsentry/raven-js/pull/402
* NEW: `serverName` config option. https://github.com/getsentry/raven-js/pull/404
* NEW: Experimental support for React Native added.

## 1.2.0
* BUGFIX: Error in cases where a `document` context doesn't exist. See: https://github.com/getsentry/raven-js/pull/383
* BUGFIX: Trailing comma when using unminified dist which affects IE9. See: https://github.com/getsentry/raven-js/pull/385
* NEW: Add ability to swap in a custom transport. Adds `Raven.setTransport`, and `transport` option to config. Docs: https://docs.getsentry.com/hosted/clients/javascript/config/
* CHANGE: Always expose `Raven` to `window`. Please call `Raven.noConflict()` if you want it restored to what it was. See: https://github.com/getsentry/raven-js/pull/393
* DEPRECATED: Replace `Raven.setReleaseContext` with `Raven.setRelease`.
* NEW: Add `Raven.clearContext()` to empty all of the context.
* NEW: Add `Raven.getContext()` to get a copy of the current context.
* NEW: `Raven.set{Extra,Tags}Context(ctx)` now merges with existing values instead of overwriting.
* NEW: Add `Raven.addPlugin()` to register a plugin to be initialized when installed.
* NEW: Plugins are now initialized and loaded when calling `Raven.install()`. This avoid some race conditions with load order.

## 1.1.22

* Fix another outstanding bug related to https://github.com/getsentry/raven-js/issues/377 that wasn't fully resolved with 1.1.21
* Laid groundwork for pluggable transports, but not ready for public consumption yet

## 1.1.21

* Fix a bug where calling `captureException` before calling `Raven.config()` would trigger it's own exception. See: https://github.com/getsentry/raven-js/issues/377

## 1.1.20

* Wrap jquery's deferred[ resolveWith | rejectWith | notifyWith ] See: https://github.com/getsentry/raven-js/pull/268
* Use window.crypto for uuid4 if present. See: https://github.com/getsentry/raven-js/pull/349
* Add winjs support. See: https://github.com/getsentry/raven-js/commit/b9a1292cbc9275fc9f9f1108ff3698cbd5ce2180
* Fix calling `Raven.captureException` from browser console. See: https://github.com/getsentry/raven-js/issues/358
* guard against document.location being null or undefined. See: https://github.com/getsentry/raven-js/pull/357
* Change error message format to match other clients. See: https://github.com/getsentry/raven-js/commit/64ca528b1b066ec7cdb5ef38e755c445f16ebef7
* Don't require a user in the DSN. See: https://github.com/getsentry/raven-js/pull/361
* Add `crossOrigin` option. See: https://github.com/getsentry/raven-js/pull/362
* Avoid recursing when using the `console` plugin. See: https://github.com/getsentry/raven-js/commit/f92ff9de538f331a291af4a7d302202e587aaae5

## 1.1.19

* Use more compliant way of creating an Image in the dom. See: https://github.com/getsentry/raven-js/pull/334
* `String` objects weren't getting identified as a string. See: https://github.com/getsentry/raven-js/pull/336
* Expose getter/setter for dataCallback and shouldSendCallback
* Better handle if/when the dataCallback returns garbage
* Fix support for nodeunit. See: https://github.com/getsentry/raven-js/pull/338
* Fix `console.warn` sending as a `warning` level to server. See: https://github.com/getsentry/raven-js/issues/342
* Improve the capture of unhandled errors from promises in Ember plugin. See: https://github.com/getsentry/raven-js/pull/330

## 1.1.18

* Fixed a trailing comma which would make IE8 cry. This affects the uncompressed builds only. Compressed builds were unaffected. See: https://github.com/getsentry/raven-js/pull/333

## 1.1.17

* Better support for Angular errors. See: https://github.com/getsentry/raven-js/pull/238
* Allow setting truncate length through `globalOptions.maxMessageLength`. See: https://github.com/getsentry/raven-js/pull/246
* Fixed the pattern for parsing gecko stacktraces. See: https://github.com/getsentry/raven-js/pull/252
* Browserify support. See: https://github.com/getsentry/raven-js/pull/253, https://github.com/getsentry/raven-js/pull/260, https://github.com/getsentry/raven-js/pull/261
* Start tracking `session:duration` automatically as metadata.
* Fix globalOptions overwrite. See: https://github.com/getsentry/raven-js/pull/264
* Better cross origin support. See: https://github.com/getsentry/raven-js/pull/276
* Better anonymous function support in Chrome stack trace parsing. See: https://github.com/getsentry/raven-js/pull/290, https://github.com/getsentry/raven-js/pull/294
* Remove deprecated `site` param.
* New `Raven.isSetup()`. See: https://github.com/getsentry/raven-js/pull/309
* Better backbone.js support. See: https://github.com/getsentry/raven-js/pull/307
* `ignoreErrors` now also is applied to `captureMessage()`. See: https://github.com/getsentry/raven-js/pull/312
* Capture unhandled errors from promises in Ember. See: https://github.com/getsentry/raven-js/pull/319
* Add new support for `releases`. See: https://github.com/getsentry/raven-js/issues/325

## 1.1.16

* Fixed a bug that was preventing stack frames from `raven.js` from being hidden correctly. See: https://github.com/getsentry/raven-js/pull/216
* Fixed an IE bug with the `console` plugin. See: https://github.com/getsentry/raven-js/issues/217
* Added support for `chrome-extension://` protocol in Chrome in stack traces.
* Added `setExtraContext` and `setTagsContext`.  See: https://github.com/getsentry/raven-js/pull/219
* Renamed `setUser` to `setUserContext` to match. `setUser` still exists, but will be deprecated in a future release.
* New `backbone.js` plugin. See: https://github.com/getsentry/raven-js/pull/220
* Added support for `chrome://` protocol in Firefox in stack traces. See: https://github.com/getsentry/raven-js/pull/225
* Ignore more garbage from IE cross origin errors. See: https://github.com/getsentry/raven-js/pull/224
* Added `Raven.debug` to prevent logging to `console` when `false`. Defaults to `true` for backwards compatability. See: https://github.com/getsentry/raven-js/pull/229
* Prevent calling `Raven.config()` or `Raven.install()` twice. See: https://github.com/getsentry/raven-js/pull/233

## 1.1.15

* Fix issues if a non-string were passed to `Raven.captureMessage` and non-Error objects were passed to `Raven.captureException`.

## 1.1.14

* Only filter normal Error objects without a message, not all of them. Turns out, people throw errors like this. Ahem, Underscore.js. See: https://github.com/jashkenas/underscore/pull/1589/files

## 1.1.13

* Fixed a unicode issue in the previous release.

## 1.1.12

* Fix a bug using the `console` plugin with older IE. See: https://github.com/getsentry/raven-js/pull/192
* Added initial `ember.js` plugin for early testing and feedback.
* Added initial `angular.js` plugin for early testing and feedback.
* Fixed an issue with the `require.js` plugin basically not working at all. See: https://github.com/getsentry/raven-js/commit/c2a2e2672a2a61a5a07e88f24a9c885f6dba57ae
* Got rid of `Raven.afterLoad` and made it internal only.
* `Raven.TraceKit` is now internal only.
* Truncate message length to a max of 100 characters becasue angular.js sucks and generates stupidly large error messages.

## 1.1.11

* Capture column number from FireFox
* Fix propagation of extra options through `captureException`, see: https://github.com/getsentry/raven-js/pull/189
* Fix a minor bug that causes TraceKit to blow up of someone passes something dumb through `window.onerror`

## 1.1.10

* A falsey DSN value disables Raven without yelling about an invalid DSN.

## 1.1.9

* Added `Raven.lastEventId()` to get back the Sentry event id. See: http://raven-js.readthedocs.org/en/latest/usage/index.html#getting-back-an-event-id
* Fixed a bug in the `console` plugin. See: https://github.com/getsentry/raven-js/pull/181
* Provide a way out of deep wrapping arguments. See: https://github.com/getsentry/raven-js/pull/182
* `Raven.uninstall()` actually removes the patched `window.onerror`.
* No more globally exposed `TraceKit`!

## 1.1.8

* Fixed a bug in IE8. See: https://github.com/getsentry/raven-js/pull/179

## 1.1.4-1.1.7

These were a bunch of super small incremental updates trying to get better integration and better support inside Sentry itself.

* Culprit determined from the src url of the offending script, not the url of the page.
* Send Sentry the frames in the right order. They were being sent in reverse. Somehow nobody noticed this.
* Support for Chrome's new window.onerror api. See: https://github.com/getsentry/raven-js/issues/172

## 1.1.3

* When loading with an AMD loader present, do not automatically call `Raven.noConflict()`. This was causing issues with using plugins. See: https://github.com/getsentry/raven-js/pull/165
* https://github.com/getsentry/raven-js/pull/168

## 1.1.2

* An invalid DSN will now raise a RavenConfigError instead of some cryptic error
* Will raise a RavenConfigError when supplying the private key part of the DSN since this isn't applicable for raven.js and is harmful to include
* https://github.com/getsentry/raven-js/issues/128

## 1.1.1

* Fixed a bug in parsing some DSNs. See: https://github.com/getsentry/raven-js/issues/160

## 1.1.0

### Plugins
If you're upgrading from 1.0.x, 2 "plugins" were included with the package. These 2 plugins are now stripped out of core and included as the `jquery` and `native` plugins. If you'd like to start using 1.1.0 and maintain existing functionality, you'll want to use: http://cdn.ravenjs.com/1.1.0/jquery,native/raven.min.js For a list of other plugins, checkout http://ravenjs.com

### ravenjs.com
A new website dedicated to helping you compile a custom build of raven.js

### whitelistUrls
`whitelistUrls` are recommended over `ignoreUrls`. `whitelistUrls` drastically helps cut out noisy error messages from other scripts running on your site.

### Misc
* `ignoreUrls`, `ignoreErrors`, `includePaths` have all been unified to accept both a regular expression and strings to avoid confusion and backwards compatability
* `Raven.wrap` recursively wraps arguments
* Events are dispatched when an exception is received, recorded or failed sending to Sentry
* Support newer Sentry protocol which allows smaller packets
* Allow loading raven async with RavenConfig
* Entirely new build system with Grunt
* `options.collectWindowErrors` to tell Raven to ignore window.onerror
