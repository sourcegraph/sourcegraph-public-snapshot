Changelog
=========

1.1.19
~~~~~~
* Use more compliant way of creating an Image in the dom. See: https://github.com/getsentry/raven-js/pull/334
* `String` objects weren't getting identified as a string. See: https://github.com/getsentry/raven-js/pull/336
* Expose getter/setter for dataCallback and shouldSendCallback
* Better handle if/when the dataCallback returns garbage
* Fix support for nodeunit. See: https://github.com/getsentry/raven-js/pull/338
* Fix `console.warn` sending as a `warning` level to server. See: https://github.com/getsentry/raven-js/issues/342
* Improve the capture of unhandled errors from promises in Ember plugin. See: https://github.com/getsentry/raven-js/pull/330

1.1.18
~~~~~~
* Fixed a trailing comma which would make IE8 cry. This affects the uncompressed builds only. Compressed builds were unaffected. See: https://github.com/getsentry/raven-js/pull/333

1.1.17
~~~~~~
* Better support for Angular errors. See: https://github.com/getsentry/raven-js/pull/238
* Allow setting truncate length through ``globalOptions.maxMessageLength``. See: https://github.com/getsentry/raven-js/pull/246
* Fixed the pattern for parsing gecko stacktraces. See: https://github.com/getsentry/raven-js/pull/252
* Browserify support. See: https://github.com/getsentry/raven-js/pull/253, https://github.com/getsentry/raven-js/pull/260, https://github.com/getsentry/raven-js/pull/261
* Start tracking ``session:duration`` automatically as metadata.
* Fix globalOptions overwrite. See: https://github.com/getsentry/raven-js/pull/264
* Better cross origin support. See: https://github.com/getsentry/raven-js/pull/276
* Better anonymous function support in Chrome stack trace parsing. See: https://github.com/getsentry/raven-js/pull/290, https://github.com/getsentry/raven-js/pull/294
* Remove deprecated ``site`` param.
* New ``Raven.isSetup()``. See: https://github.com/getsentry/raven-js/pull/309
* Better backbone.js support. See: https://github.com/getsentry/raven-js/pull/307
* ``ignoreErrors`` now also is applied to ``captureMessage()``. See: https://github.com/getsentry/raven-js/pull/312
* Capture unhandled errors from promises in Ember. See: https://github.com/getsentry/raven-js/pull/319
* Add new support for ``releases``. See: https://github.com/getsentry/raven-js/issues/325

1.1.16
~~~~~~
* Fixed a bug that was preventing stack frames from ``raven.js`` from being hidden correctly. See: https://github.com/getsentry/raven-js/pull/216
* Fixed an IE bug with the ``console`` plugin. See: https://github.com/getsentry/raven-js/issues/217
* Added support for ``chrome-extension://`` protocol in Chrome in stack traces.
* Added ``setExtraContext`` and ``setTagsContext``.  See: https://github.com/getsentry/raven-js/pull/219
* Renamed ``setUser`` to ``setUserContext`` to match. ``setUser`` still exists, but will be deprecated in a future release.
* New ``backbone.js`` plugin. See: https://github.com/getsentry/raven-js/pull/220
* Added support for ``chrome://`` protocol in Firefox in stack traces. See: https://github.com/getsentry/raven-js/pull/225
* Ignore more garbage from IE cross origin errors. See: https://github.com/getsentry/raven-js/pull/224
* Added ``Raven.debug`` to prevent logging to ``console`` when ``false``. Defaults to ``true`` for backwards compatability. See: https://github.com/getsentry/raven-js/pull/229
* Prevent calling ``Raven.config()`` or ``Raven.install()`` twice. See: https://github.com/getsentry/raven-js/pull/233

1.1.15
~~~~~~
* Fix issues if a non-string were passed to ``Raven.captureMessage`` and non-Error objects were passed to ``Raven.captureException``.

1.1.14
~~~~~~
* Only filter normal Error objects without a message, not all of them. Turns out, people throw errors like this. Ahem, Underscore.js. See: https://github.com/jashkenas/underscore/pull/1589/files

1.1.13
~~~~~~
* Fixed a unicode issue in the previous release.

1.1.12
~~~~~~
* Fix a bug using the ``console`` plugin with older IE. See: https://github.com/getsentry/raven-js/pull/192
* Added initial ``ember.js`` plugin for early testing and feedback.
* Added initial ``angular.js`` plugin for early testing and feedback.
* Fixed an issue with the ``require.js`` plugin basically not working at all. See: https://github.com/getsentry/raven-js/commit/c2a2e2672a2a61a5a07e88f24a9c885f6dba57ae
* Got rid of ``Raven.afterLoad`` and made it internal only.
* ``Raven.TraceKit`` is now internal only.
* Truncate message length to a max of 100 characters becasue angular.js sucks and generates stupidly large error messages.

1.1.11
~~~~~~
* Capture column number from FireFox
* Fix propagation of extra options through ``captureException``, see: https://github.com/getsentry/raven-js/pull/189
* Fix a minor bug that causes TraceKit to blow up of someone passes something dumb through ``window.onerror``

1.1.10
~~~~~~
* A falsey DSN value disables Raven without yelling about an invalid DSN.

1.1.9
~~~~~
* Added ``Raven.lastEventId()`` to get back the Sentry event id. See: http://raven-js.readthedocs.org/en/latest/usage/index.html#getting-back-an-event-id
* Fixed a bug in the ``console`` plugin. See: https://github.com/getsentry/raven-js/pull/181
* Provide a way out of deep wrapping arguments. See: https://github.com/getsentry/raven-js/pull/182
* ``Raven.uninstall()`` actually removes the patched ``window.onerror``.
* No more globally exposed ``TraceKit``!

1.1.8
~~~~~
* Fixed a bug in IE8. See: https://github.com/getsentry/raven-js/pull/179

1.1.4-1.1.7
~~~~~~~~~~~
These were a bunch of super small incremental updates trying to get better integration and better support inside Sentry itself.

* Culprit determined from the src url of the offending script, not the url of the page.
* Send Sentry the frames in the right order. They were being sent in reverse. Somehow nobody noticed this.
* Support for Chrome's new window.onerror api. See: https://github.com/getsentry/raven-js/issues/172

1.1.3
~~~~~
* When loading with an AMD loader present, do not automatically call ``Raven.noConflict()``. This was causing issues with using plugins. See: https://github.com/getsentry/raven-js/pull/165
* https://github.com/getsentry/raven-js/pull/168

1.1.2
~~~~~
* An invalid DSN will now raise a RavenConfigError instead of some cryptic error
* Will raise a RavenConfigError when supplying the private key part of the DSN since this isn't applicable for raven.js and is harmful to include
* https://github.com/getsentry/raven-js/issues/128

1.1.1
~~~~~
* Fixed a bug in parsing some DSNs. See: https://github.com/getsentry/raven-js/issues/160

1.1.0
~~~~~

Plugins
-------
If you're upgrading from 1.0.x, 2 "plugins" were included with the package. These 2 plugins are now stripped out of core and included as the ``jquery`` and ``native`` plugins. If you'd like to start using 1.1.0 and maintain existing functionality, you'll want to use: http://cdn.ravenjs.com/1.1.0/jquery,native/raven.min.js For a list of other plugins, checkout http://ravenjs.com

ravenjs.com
-----------
A new website dedicated to helping you compile a custom build of raven.js

whitelistUrls
-------------
``whitelistUrls`` are recommended over ``ignoreUrls``. ``whitelistUrls`` drastically helps cut out noisy error messages from other scripts running on your site.

Misc
----
* ``ignoreUrls``, ``ignoreErrors``, ``includePaths`` have all been unified to accept both a regular expression and strings to avoid confusion and backwards compatability
* ``Raven.wrap`` recursively wraps arguments
* Events are dispatched when an exception is received, recorded or failed sending to Sentry
* Support newer Sentry protocol which allows smaller packets
* Allow loading raven async with RavenConfig
* Entirely new build system with Grunt
* ``options.collectWindowErrors`` to tell Raven to ignore window.onerror
