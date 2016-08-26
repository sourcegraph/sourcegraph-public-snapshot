React
=====

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

At this point, Raven is ready to capture any uncaught exception.

Expanded Usage
--------------

It's likely you'll end up in situations where you want to gracefully
handle errors. A good pattern for this would be to setup a logError helper:

.. code-block:: javascript

    function logException(ex, context) {
      Raven.captureException(ex, {
        extra: context
      });
      /*eslint no-console:0*/
      window.console && console.error && console.error(ex);
    }

Now in your components (or anywhere else), you can fail gracefully:

.. code-block:: javascript

    var Component = React.createClass({
        render() {
            try {
                // ..
            } catch (ex) {
                logException(ex);
            }
        }
    });
