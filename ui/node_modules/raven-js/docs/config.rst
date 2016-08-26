Configuration
=============

To get started, you need to configure Raven.js to use your Sentry DSN:

.. sourcecode:: javascript

    Raven.config('___PUBLIC_DSN___').install()

At this point, Raven is ready to capture any uncaught exception.

Optional settings
-----------------

``Raven.config()`` can optionally be passed an additional argument for extra configuration:

.. sourcecode:: javascript

    Raven.config('___PUBLIC_DSN___', {
        release: '1.3.0'
    }).install()

Those configuration options are documented below:

.. describe:: logger

    The name of the logger used by Sentry. Default: ``javascript``

    .. code-block:: javascript

        {
          logger: 'javascript'
        }

.. describe:: release

    Track the version of your application in Sentry.

    .. code-block:: javascript

        {
          release: '721e41770371db95eee98ca2707686226b993eda'
        }

    Can also be defined with ``Raven.setRelease('721e41770371db95eee98ca2707686226b993eda')``.

.. describe:: environment

    Track the environment name inside Sentry.

    .. code-block:: javascript

        {
          environment: 'production'
        }

.. describe:: serverName

    .. versionadded:: 1.3.0

    Typically this would be the server name, but that doesn’t exist on
    all platforms. Instead you may use something like the device ID, as it
    indicates the host which the client is running on.

    .. code-block:: javascript

        {
          serverName: device.uuid
        }


.. describe:: tags

    Additional `tags <https://www.getsentry.com/docs/tags/>`__ to assign to each event.

    .. code-block:: javascript

        {
          tags: {git_commit: 'c0deb10c4'}
        }

.. _config-whitelist-urls:

.. describe:: whitelistUrls

    The inverse of ``ignoreUrls``. Only report errors from whole urls
    matching a regex pattern or an exact string. ``whitelistUrls`` should
    match the url of your actual JavaScript files. It should match the url
    of your site if and only if you are inlining code inside ``<script>``
    tags. Not setting this value is equivalent to a catch-all and will not
    filter out any values.

    Does not affect ``captureMessage`` or when non-error object is passed in
    as argument to captureException.

    .. code-block:: javascript

        {
          whitelistUrls: [/getsentry\.com/, /cdn\.getsentry\.com/]
        }

.. describe:: ignoreErrors

    Very often, you will come across specific errors that are a result of
    something other than your application, or errors that you're
    completely not interested in. `ignoreErrors` is a list of these
    messages to be filtered out before being sent to Sentry as either
    regular expressions or strings.

    Does not affect captureMessage or when non-error object is passed in
    as argument to captureException.

    .. code-block:: javascript

        {
          ignoreErrors: ['fb_xd_fragment']
        }

.. describe:: ignoreUrls

    The inverse of ``whitelistUrls`` and similar to ``ignoreErrors``, but
    will ignore errors from whole urls matching a regex pattern or an
    exact string.

    .. code-block:: javascript

        {
          ignoreUrls: [/graph\.facebook\.com/, 'http://example.com/script2.js']
        }

    Does not affect captureMessage or when non-error object is passed in
    as argument to ``captureException``.

.. describe:: includePaths

    An array of regex patterns to indicate which urls are a part of your
    app in the stack trace. All other frames will appear collapsed inside
    Sentry to make it easier to discern between frames that happened in
    your code vs other code. It'd be suggested to add the current page
    url, and the host for your CDN.

    .. code-block:: javascript

        {
            includePaths: [/https?:\/\/getsentry\.com/, /https?:\/\/cdn\.getsentry\.com/]
        }

.. describe:: dataCallback

    A function that allows mutation of the data payload right before being
    sent to Sentry.

    .. code-block:: javascript

        {
            dataCallback: function(data) {
              // do something to data
              return data;
            }
        }

.. describe:: shouldSendCallback

    A callback function that allows you to apply your own filters to
    determine if the message should be sent to Sentry.

    .. code-block:: javascript

        {
            shouldSendCallback: function(data) {
              return false;
            }
        }

.. describe:: maxMessageLength

    By default, Raven does not truncate messages. If you need to truncate
    characters for whatever reason, you may set this to limit the length.

.. describe:: autoBreadcrumbs

    Enables/disables automatic collection of breadcrumbs. Possible values are:

    * `true` (default)
    * `false` - all automatic breadcrumb collection disabled
    * A dictionary of individual breadcrumb types that can be enabled/disabled:

    .. code-block:: javascript

        autoBreadcrumbs: {
            'xhr': false,      // XMLHttpRequest
            'console': false,  // console logging
            'dom': true,       // DOM interactions, i.e. clicks/typing
            'location': false  // url changes, including pushState/popState
        }

.. describe:: maxBreadcrumbs

    By default, Raven captures as many as 100 breadcrumb entries. If you find this too noisy, you can reduce this
    number by setting `maxBreadcrumbs`. Note that this number cannot be set higher than the default of 100.

.. describe:: transport

    Override the default HTTP data transport handler.

    Alternatively, can be specified using ``Raven.setTransport(myTransportFunction)``.

    .. code-block:: javascript

        {
            transport: function (options) {
                // send data
            }
        }

    The provided function receives a single argument, ``options``, with the following properties:

    url
        The target url where the data is sent.

    data
        The outbound exception data.

        For POST requests, this should be JSON-encoded and set as the
        HTTP body (and transferred as Content-type: application/json).

        For GET requests, this should be JSON-encoded and passed over the
        query string with key ``sentry_data``.

    auth
        An object representing authentication data. This should be converted to urlencoded key/value pairs
        and passed as part of the query string, for both GET and POST requests.

        The auth object has the following properties:

        sentry_version
            The API version of the Sentry server.
        sentry_client
            The name and version of the Sentry client of the form ``client/version``.
            In this case, ``raven-js/${Raven.VERSION}``.
        sentry_key
            Your public client key (DSN).

    onSuccess
        Callback to be invoked upon a successful request.

    onFailure
        Callback to be invoked upon a failed request.

.. describe:: allowSecretKey

    By default, Raven.js will throw an error if configured with a Sentry DSN that contains a secret key.
    When using Raven.js with a web application accessed via a browser over the web, you should
    only use your public DSN. But if you are using Raven.js in an environment like React Native or Electron,
    where your application is running "natively" on a device and not accessed at a web address, you may need
    to use your secret DSN string. To do so, set ``allowPrivateKey: true`` during configuration.


Putting it all together
-----------------------

.. code-block:: html

    <!doctype html>
    <html>
    <head>
        <title>Awesome stuff happening here</title>
    </head>
    <body>
        ...
        <script src="jquery.min.js"></script>
        <script src="https://cdn.ravenjs.com/3.5.1/raven.min.js"></script>
        <script>
            Raven.config('___PUBLIC_DSN___', {
                logger: 'my-logger',
                whitelistUrls: [
                    /disqus\.com/,
                    /getsentry\.com/
                ],
                ignoreErrors: [
                    'fb_xd_fragment',
                    /ReferenceError:.*/
                ],
                includePaths: [
                    /https?:\/\/(www\.)?getsentry\.com/
                ]
            }).install();
        </script>
        <script src="myapp.js"></script>
    </body>
    </html>
