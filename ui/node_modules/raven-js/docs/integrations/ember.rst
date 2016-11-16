Ember
=====

To use Sentry with your Ember application, you will need to use both Raven.js (Sentry's browser JavaScript SDK) and the Raven.js Ember plugin.

On its own, Raven.js will report any uncaught exceptions triggered from your application. For advanced usage examples of Raven.js, please read :doc:`Raven.js usage <../usage>`.

Additionally, the Raven.js Ember plugin will catch any Ember-specific exceptions reported through Ember's `onerror <https://guides.emberjs.com/v3.0.2/configuring-ember/debugging/#toc_implement-an-ember-onerror-hook-to-log-all-errors-in-production>`_. hook
and any `RSVP promises <https://guides.emberjs.com/v3.0.2/configuring-ember/debugging/#toc_errors-within-an-code-rsvp-promise-code>`_ that would otherwise be swallowed.

Installation
------------

Raven.js and the Raven.js Ember plugin are distributed using a few different methods.

Using our CDN
~~~~~~~~~~~~~

For convenience, our CDN serves a single, minified JavaScript file containing both Raven.js and the Raven.js Ember plugin. It should be included **after** Ember, but **before** your application code.

Example:

.. sourcecode:: html

    <script src="http://builds.emberjs.com/tags/v2.3.1/ember.prod.js"></script>
    <script src="https://cdn.ravenjs.com/3.8.1/ember/raven.min.js" crossorigin="anonymous"></script>
    <script>Raven.config('___PUBLIC_DSN___').install();</script>

Note that this CDN build auto-initializes the Ember plugin.

Using package managers
~~~~~~~~~~~~~~~~~~~~~~

Pre-built distributions of Raven.js and the Raven.js Ember plugin are available via both Bower and npm for importing in your ``ember-cli-build.js`` file.

Bower
`````

.. code

.. code-block:: sh

    $ bower install raven-js --save

.. code-block:: javascript

    app.import('bower_components/raven-js/dist/raven.js');
    app.import('bower_components/raven-js/dist/plugins/ember.js');

.. code-block:: html

    <script src="assets/vendor.js"></script>
    <script>
      Raven
        .config('___PUBLIC_DSN___')
        .addPlugin(Raven.Plugins.Ember)
        .install();
    </script>
    <script src="assets/your-app.js"></script>

npm
````

.. code-block:: sh

    $ npm install raven-js --save

.. code-block:: javascript

    app.import('bower_components/raven-js/dist/raven.js');
    app.import('bower_components/raven-js/dist/plugins/ember.js');

.. code-block:: html

    <script src="assets/vendor.js"></script>
    <script>
      Raven
        .config('___PUBLIC_DSN___')
        .addPlugin(Raven.Plugins.Ember)
        .install();
    </script>
    <script src="assets/your-app.js"></script>

These examples assume that Ember is exported globally as ``window.Ember``. You can alternatively pass a reference to the ``Ember`` object directly as the second argument to ``addPlugin``:

.. code-block:: javascript

    Raven.addPlugin(Raven.Plugins.Ember, Ember);
