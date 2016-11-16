Angular 2
=========

On its own, Raven.js will report any uncaught exceptions triggered from your application. For advanced usage examples of Raven.js, please read :doc:`Raven.js usage <../usage>`.

Additionally, Raven.js can be configured to catch any Angular 2-specific exceptions reported through the `angular2/core/ErrorHandler
<https://angular.io/docs/js/latest/api/core/index/ErrorHandler-class.html>`_ component.


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

Then, in your main module file (where ``@NgModule`` is called, e.g. app.module.ts):

.. code-block:: js

    import Raven = require('raven-js');
    import { BrowserModule } from '@angular/platform-browser';
    import { NgModule, ErrorHandler } from '@angular/core';
    import { AppComponent } from './app.component';

    Raven
      .config('___PUBLIC_DSN___')
      .install();

    class RavenErrorHandler implements ErrorHandler {
      handleError(err:any) : void {
        Raven.captureException(err.originalError);
      }
    }

    @NgModule({
      imports: [ BrowserModule ],
      declarations: [ AppComponent ],
      bootstrap: [ AppComponent ],
      providers: [ { provide: ErrorHandler, useClass: RavenErrorHandler } ]
    })
    export class AppModule { }

Once you've completed these two steps, you are done.
