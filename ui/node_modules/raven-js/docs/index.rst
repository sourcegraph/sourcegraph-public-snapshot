.. sentry:edition:: self

    Raven.js
    ========

.. sentry:edition:: hosted, on-premise

    .. class:: platform-js

    JavaScript
    ==========

Raven.js is a tiny standalone JavaScript client for Sentry. It can be
used to report errors from a web browser. The quality of reporting will
heavily depend on the environment the data is submitted from.

**Note**: If you're using Node on the server, you'll need
`raven-node <https://github.com/getsentry/raven-node>`_.

Installation
------------

Raven.js is distributed in a few different methods, and should get
included after any other libraries are included, but before your own
scripts.  For all details see :doc:`install`.

.. sourcecode:: html

    <script src="https://cdn.ravenjs.com/3.8.1/raven.min.js"></script>


Configuring the Client
----------------------

Now you need to set up Raven.js to use your `Sentry DSN
<https://docs.sentry.io/hosted/quickstart/#configure-the-dsn>`_:

.. code-block:: javascript

    Raven.config('___PUBLIC_DSN___').install()

At this point, Raven is ready to capture any uncaught exception.

Once you have Raven up and running, it is highly recommended to check out :doc:`config`
and :doc:`usage` to improve your results.

Reporting Errors
----------------

By default, Raven makes a best effort to capture any uncaught exception.

To report errors manually, wrap potentially problematic code with a ``try...catch``
block and call ``Raven.captureException``:

.. code-block:: javascript

    try {
        doSomething(a[0])
    } catch(e) {
        Raven.captureException(e)
    }

There are more ways to report errors.  For a complete guide on this see
:ref:`raven-js-reporting-errors`.

Adding Context
--------------

While a user is logged in, you can tell Sentry to associate errors with
user data.  This data is then submitted with each error which allows you
to figure out which users are affected.

.. code-block:: javascript

    Raven.setUserContext({
        email: 'matt@example.com',
        id: '123'
    })

If at any point, the user becomes unauthenticated, you can call
``Raven.setUserContext()`` with no arguments to remove their data.

Other similar methods are ``Raven.setExtraContext`` and
``Raven.setTagsContext`` as well as ``Raven.context``.  See
:ref:`raven-js-additional-context` for more info.

Dealing with Minified Source Code
---------------------------------

Raven and Sentry support `Source Maps
<http://www.html5rocks.com/en/tutorials/developertools/sourcemaps/>`_.  If
you provide source maps in addition to your minified files that data
becomes available in Sentry.  For more information see
:ref:`raven-js-sourcemaps`.

Browser Compatibility
---------------------

Raven.js supports all major browsers. In older browsers, error reports collected
by Raven.js may have a degraded level of detail – for example, missing stack trace data
or missing source code column numbers.

The table below describes what features are available in each supported browser:

+-------------------------+--------------+----------------+-------------+
| Browser                 | Line numbers | Column numbers | Stack trace |
+=========================+==============+================+=============+
| Chrome                  | ✓            | ✓              | ✓           |
+-------------------------+--------------+----------------+-------------+
| Firefox                 | ✓            | ✓              | ✓           |
+-------------------------+--------------+----------------+-------------+
| Edge                    | ✓            | ✓              | ✓           |
+-------------------------+--------------+----------------+-------------+
| IE 11                   | ✓            | ✓              | ✓           |
+-------------------------+--------------+----------------+-------------+
| IE 10                   | ✓            | ✓              | ✓           |
+-------------------------+--------------+----------------+-------------+
| IE 9                    | ✓            | ✓              |             |
+-------------------------+--------------+----------------+-------------+
| IE 8                    | ✓            |                |             |
+-------------------------+--------------+----------------+-------------+
| Safari 6+               | ✓            | ✓              | ✓           |
+-------------------------+--------------+----------------+-------------+
| iOS Safari 6+           | ✓            | ✓              | ✓           |
+-------------------------+--------------+----------------+-------------+
| Opera 15+               | ✓            | ✓              | ✓           |
+-------------------------+--------------+----------------+-------------+
| Android Browser 4.4     | ✓            | ✓              | ✓           |
+-------------------------+--------------+----------------+-------------+
| Android Browser 4 - 4.3 | ✓            |                |             |
+-------------------------+--------------+----------------+-------------+

For unlisted browsers (e.g. IE7), Raven.js is designed to fail gracefully. Including
it on your page should have no effect on your page; it will just not collect
and report uncaught exceptions.

Deep Dive
---------

For more detailed information about how to get most out of Raven.js there
is additional documentation available that covers all the rest:

.. toctree::
   :maxdepth: 2
   :titlesonly:

   install
   config
   usage
   integrations/index
   sourcemaps
   tips

.. sentry:edition:: self

   Development info:

    .. toctree::
       :maxdepth: 2
       :titlesonly:

       contributing

Resources:

* `Downloads and CDN <http://ravenjs.com/>`_
* `Bug Tracker <http://github.com/getsentry/raven-js/issues>`_
* `Github Project <http://github.com/getsentry/raven-js>`_
