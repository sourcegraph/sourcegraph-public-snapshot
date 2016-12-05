Backbone
========

Installation
------------

Start by adding the ``raven.js`` script tag to your page. It should be loaded as early as possible.

.. sourcecode:: html

    <script src="https://cdn.ravenjs.com/3.8.1/raven.min.js"
        crossorigin="anonymous"></script>

Configuring the Client
----------------------

Next configure Raven.js to use your Sentry DSN:

.. code-block:: javascript

    Raven.config('___PUBLIC_DSN___').install()

At this point, Raven is ready to capture any uncaught exception.
