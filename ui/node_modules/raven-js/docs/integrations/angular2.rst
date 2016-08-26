Angular 2
=========

On its own, Raven.js will report any uncaught exceptions triggered from your application. For advanced usage examples of Raven.js, please read :doc:`Raven.js usage <../usage>`.

Additionally, Raven.js can be configured to catch any Angular 2-specific exceptions reported through the `angular2/core/ExceptionHandler
<https://angular.io/docs/js/latest/api/core/index/ExceptionHandler-class.html>`_ component.


TypeScript Support
------------------

Raven.js ships with a `TypeScript declaration file
<https://github.com/getsentry/raven-js/blob/master/typescript/raven.d.ts>`_, which helps static checking correctness of
Raven.js API calls, and facilitates auto-complete in TypeScript-aware IDEs like Visual Studio Code.


Installation
------------

Raven.js should be installed via npm.

.. code-block:: sh

    $ npm install raven-js --save


Configuration
-------------

Configuration depends on which module loader/packager you are using to build your Angular 2 application.

Below are instructions for `SystemJS
<https://github.com/systemjs/systemjs>`__, followed by instructions for `Webpack
<https://webpack.github.io/>`__, Angular CLI, and other module loaders/packagers.

SystemJS
~~~~~~~~

First, configure SystemJS to locate the Raven.js package:

.. code-block:: js

    System.config({
      packages: {
        /* ... existing packages above ... */
        'raven-js': {
          main: 'dist/raven.js'
        }
      },
      paths: {
        /* ... existing paths above ... */
        'raven-js': 'node_modules/raven-js'
      }
    });

Then, in your main application file (where ``bootstrap`` is called, e.g. main.ts):

.. code-block:: js

    import Raven from 'raven-js';
    import { bootstrap } from 'angular2/platform/browser';
    import { MainApp } from './app.component';
    import { provide, ExceptionHandler } from 'angular2/core';

    Raven
      .config('___PUBLIC_DSN___')
      .install();

    class RavenExceptionHandler {
      call(err:any) {
        Raven.captureException(err.originalException);
      }
    }

    bootstrap(MainApp, [
      provide(ExceptionHandler, {useClass: RavenExceptionHandler})
    ]);

Once you've completed these two steps, you are done.

Webpack, Angular CLI, and Other Module Loaders
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

In Webpack, Angular CLI, and other module loaders/packagers, you may need to use the **require** keyword as
part of your `import` statement:

.. code-block:: js

    import Raven = require('raven-js');  // NOTE: "require" not "from"
    import { bootstrap } from 'angular2/platform/browser';
    import { MainApp } from './app.component';
    import { provide, ExceptionHandler } from 'angular2/core';

    Raven
      .config('___PUBLIC_DSN___')
      .install();

    class RavenExceptionHandler {
      call(err:any) {
        Raven.captureException(err.originalException);
      }
    }

    bootstrap(MainApp, [
      provide(ExceptionHandler, {useClass: RavenExceptionHandler})
    ]);

