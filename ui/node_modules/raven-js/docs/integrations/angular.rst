Angular 1
=========

To use Sentry with your Angular 1.x application, you will need to use both Raven.js (Sentry's browser JavaScript SDK) and the Raven.js Angular plugin.

On its own, Raven.js will report any uncaught exceptions triggered from your application. For advanced usage examples of Raven.js, please read :doc:`Raven.js usage <../usage>`.

Additionally, the Raven.js Angular plugin will catch any Angular-specific exceptions reported through Angular's ``$exceptionHandler`` interface.

**Note**: This documentation is for Angular 1.x. See also: :doc:`Angular 2.x <angular2>`

Installation
------------

Raven.js and the Raven.js Angular plugin are distributed using a few different methods.

Using our CDN
~~~~~~~~~~~~~

For convenience, our CDN serves a single, minified JavaScript file containing both Raven.js and the Raven.js Angular plugin. It should be included **after** Angular, but **before** your application code.

Example:

.. sourcecode:: html

    <script src="https://ajax.googleapis.com/ajax/libs/angularjs/1.4.5/angular.min.js"></script>
    <script src="https://cdn.ravenjs.com/3.5.1/angular/raven.min.js"></script>
    <script>Raven.config('___PUBLIC_DSN___').install();</script>

Note that this CDN build auto-initializes the Angular plugin.

Using package managers
~~~~~~~~~~~~~~~~~~~~~~

Pre-built distributions of Raven.js and the Raven.js Angular plugin are available via both Bower and npm.

Bower
`````

.. code

.. code-block:: sh

    $ bower install raven-js --save

.. code-block:: html

    <script src="/bower_components/angular/angular.js"></script>
    <script src="/bower_components/raven-js/dist/raven.js"></script>
    <script src="/bower_components/raven-js/dist/plugins/angular.js"></script>
    <script>
      Raven
        .config('___PUBLIC_DSN___')
        .addPlugin(Raven.Plugins.Angular)
        .install();
    </script>

npm
````

.. code-block:: sh

    $ npm install raven-js --save

.. code-block:: html

    <script src="/node_modules/angular/angular.js"></script>
    <script src="/node_modules/raven-js/dist/raven.js"></script>
    <script src="/node_modules/raven-js/dist/plugins/angular.js"></script>
    <script>
      Raven
        .config('___PUBLIC_DSN___')
        .addPlugin(Raven.Plugins.Angular)
        .install();
    </script>

These examples assume that Angular is exported globally as `window.angular`. You can alternatively pass a reference to the `angular` object directly as the second argument to `addPlugin`:

.. code-block:: javascript

  Raven.addPlugin(Raven.Plugins.Angular, angular);

Module loaders (CommonJS)
~~~~~~~~~~~~~~~~~~~~~~~~~

Raven and the Raven Angular plugin can be loaded using a module loader like Browserify or Webpack.

.. code-block:: javascript

    var angular = require('angular');
    var Raven = require('raven-js');

    Raven
      .config('___PUBLIC_DSN___')
      .addPlugin(require('raven-js/plugins/angular'), angular)
      .install();

Note that when using CommonJS-style imports, you must pass a reference to the `angular` as the second argument to `addPlugin`.

Angular Configuration
---------------------

Inside your main Angular application module, you need to declare `ngRaven` as a module dependency:

.. code-block:: javascript

    var myApp = angular.module('myApp', [
      'ngRaven',
      'ngRoute',
      'myAppControllers',
      'myAppFilters'
    ]);
