Configuration
=============

We must first configure Sentry to allow certain hosts to report errors. This prevents abuse so somebody else couldn't start sending errors to your account from their site.

**Note**: Without setting this, all messages will be rejected!

This can be found under the *Project Details* page in Sentry.

.. image:: https://i.imgur.com/S09MeSM.png

Now need to set up Raven.js to use your Sentry DSN.

.. code-block:: javascript

    Raven.config('https://public@getsentry.com/1').install()

At this point, Raven is ready to capture any uncaught exception.

Although, this technically works, this is not going to yield the greatest results. It's highly recommended to next check out :doc:`/usage/index`.

Optional settings
~~~~~~~~~~~~~~~~~

``Raven.config()`` can be passed an optional object for extra configuration.

logger
------

The name of the logger used by Sentry. Default: ``javascript``

.. code-block:: javascript

    {
      logger: 'javascript'
    }

release
-------

Track the version of your application in Sentry.

.. code-block:: javascript

    {
      release: '721e41770371db95eee98ca2707686226b993eda'
    }

Can also be defined with ``Raven.setReleaseContext('721e41770371db95eee98ca2707686226b993eda')``.

.. _config-whitelist-urls:

tags
----

Additional `tags <https://www.getsentry.com/docs/tags/>`__ to assign to each event.

.. code-block:: javascript

    {
      tags: {git_commit: 'c0deb10c4'}
    }

whitelistUrls
-------------

The inverse of ``ignoreUrls``. Only report errors from whole urls matching a regex pattern or an exact string. ``whitelistUrls`` should match the url of your actual JavaScript files. It should match the url of your site if and only if you are inlining code inside ``<script>`` tags.

Does not affect captureMessage or when non-error object is passed in as argument to captureException.

.. code-block:: javascript

    {
      whitelistUrls: [/getsentry\.com/, /cdn\.getsentry\.com/]
    }

ignoreErrors
------------

Very often, you will come across specific errors that are a result of something other than your application, or errors that you're completely not interested in. `ignoreErrors` is a list of these messages to be filtered out before being sent to Sentry as either regular expressions or strings.

Does not affect captureMessage or when non-error object is passed in as argument to captureException.

.. code-block:: javascript

    {
      ignoreErrors: ['fb_xd_fragment']
    }

ignoreUrls
----------

The inverse of ``whitelistUrls`` and similar to ``ignoreErrors``, but will ignore errors from whole urls matching a regex pattern or an exact string.

.. code-block:: javascript

    {
      ignoreUrls: [/graph\.facebook\.com/, 'http://example.com/script2.js']
    }

Does not affect captureMessage or when non-error object is passed in as argument to captureException.

includePaths
------------

An array of regex patterns to indicate which urls are a part of your app in the stack trace. All other frames will appear collapsed inside Sentry to make it easier to discern between frames that happened in your code vs other code. It'd be suggested to add the current page url, and the host for your CDN.

.. code-block:: javascript

    {
        includePaths: [/https?:\/\/getsentry\.com/, /https?:\/\/cdn\.getsentry\.com/]
    }

dataCallback
------------

A function that allows mutation of the data payload right before being sent to Sentry.

.. code-block:: javascript

    {
        dataCallback: function(data) {
          // do something to data
          return data;
        }
    }

Can also be set at runtime with `Raven.setDataCallback(function(data) { ... })`.

shouldSendCallback
------------------

A callback function that allows you to apply your own filters to determine if the message should be sent to Sentry.

.. code-block:: javascript

    {
        shouldSendCallback: function(data) {
          return false;
        }
    }

Can also be set at runtime with `Raven.setShouldSendCallback(function(data) { ... })`.

maxMessageLength
------------------

By default, raven truncates messages to a max length of 100 characters. You can customize the max length with this parameter.


Putting it all together
~~~~~~~~~~~~~~~~~~~~~~~

.. parsed-literal::

    <!DOCTYPE html>
    <html>
    <head>
        <title>Awesome stuff happening here</title>
    </head>
    <body>
        ...
        <script src="jquery.min.js"></script>
        <script src="//cdn.ravenjs.com/|release|/jquery,native/raven.min.js"></script>
        <script>
            var options = {
                logger: 'my-logger',
                whitelistUrls: [
                    /disqus\\.com/, /getsentry\\.com/
                ],
                ignoreErrors: [
                    'fb_xd_fragment', /ReferenceError:.*/
                ],
                includePaths: [
                    /https?:\\/\\/(www\\.)?getsentry\\.com/
                ]
            };
            Raven.config('https://public@app.getsentry.com/1', options).install();
        </script>
        <script src="myapp.js"></script>
    </body>
    </html>

TraceKit specific optional settings
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

Usually there is no need to touch these settings, but they exist in case you need to tweak something.

fetchContext
------------

Enable TraceKit to attempt to fetch source files to look up anonymous function names, this can be useful to enable if you don't get the context for some entries in the stack trace. Default value is ``false``.

.. code-block:: javascript

    {
        fetchContext: true
    }

linesOfContext
--------------

The count of lines surrounding the error line that should be used as context in the stack trace, default value is ``11``. Only applicable when ``fetchContext` is enabled.

.. code-block:: javascript

    {
        linesOfContext: 11
    }

collectWindowErrors
-------------------

Enable or disable the TraceKit ``window.onerror`` handler, default value is ``true``.

.. code-block:: javascript

    {
        collectWindowErrors: true
    }
