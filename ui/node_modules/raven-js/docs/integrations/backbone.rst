Backbone
========

Installation
------------

Start by adding the ``raven.js`` script tag to our page. You'll want to position it
after you load all other external libraries (like jQuery), but before your code.

.. sourcecode:: html

    <script src="https://cdn.ravenjs.com/3.5.1/raven.min.js"></script>

Configuring the Client
----------------------

Now need to set up Raven.js to use your Sentry DSN:

.. code-block:: javascript

    Raven.config('___PUBLIC_DSN___').install()

At this point, Raven is ready to capture any uncaught exception via standard hooks
in addition to Backbone specific hooks.
