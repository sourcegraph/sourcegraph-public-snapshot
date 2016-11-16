Vue.js (2.0)
============

.. sentry:support-warning::

    This plugin only works with Vue 2.0 or greater.


To use Sentry with your Vue application, you will need to use both Raven.js (Sentry's browser JavaScript SDK) and the Raven.js Vue plugin.

On its own, Raven.js will report any uncaught exceptions triggered from your application. For advanced usage examples of Raven.js, please read :doc:`Raven.js usage <../usage>`.

Additionally, the Raven.js Vue plugin will capture the name and props state of the active component where the error was thrown. This is reported via Vue's `config.errorHandler` hook.

Installation
------------

Raven.js and the Raven.js Vue plugin are distributed using a few different methods.

Using our CDN
~~~~~~~~~~~~~

For convenience, our CDN serves a single, minified JavaScript file containing both Raven.js and the Raven.js Vue plugin. It should be included **after** Vue, but **before** your application code.

Example:

.. sourcecode:: html

    <script src="https://cdn.jsdelivr.net/vue/2.0.0-rc/vue.min.js"></script>
    <script src="https://cdn.ravenjs.com/3.8.1/vue/raven.min.js"
        crossorigin="anonymous"></script>
    <script>Raven.config('___PUBLIC_DSN___').install();</script>

Note that this CDN build auto-initializes the Vue plugin.

Using package managers
~~~~~~~~~~~~~~~~~~~~~~

Both Raven.js and the Raven.js Vue plugin can be installed via npm and Bower.

npm
````

.. code-block:: sh

    $ npm install raven-js --save


Bower
`````

.. code-block:: sh

    $ bower install raven-js --save

In your main application file, import and configure both Raven.js and the Raven.js Vue plugin as follows:

.. code-block:: js

    import Vue from 'vue';
    import Raven from 'raven-js';
    import RavenVue from 'raven-js/plugins/vue';

    Raven
        .config('___PUBLIC_DSN___')
        .addPlugin(RavenVue, Vue)
        .install();
