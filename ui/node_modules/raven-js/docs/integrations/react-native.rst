React Native
============

React Native for Raven.js is a pure JavaScript error reporting solution. The plugin will report errors originating from React Native's
JavaScript engine (e.g. programming errors like "x is undefined"), but might not catch errors that originate from the underlying
operating system (iOS / Android) unless they happen to be transmitted to React Native's global error handler.

Errors caught via the React Native plugin include stack traces, breadcrumbs, and allow for unminification via source maps.

Installation
------------

In the root of your React Native project, install raven-js via npm:

.. sourcecode:: bash

    $ npm install --save raven-js

At the top of your main application file (e.g. index.ios.js and/or index.android.js), add the following code:

.. sourcecode:: javascript

    var React = require('react');

    var Raven = require('raven-js');
    require('raven-js/plugins/react-native')(Raven);

Configuring the Client
----------------------

Now we need to set up Raven.js to use your Sentry DSN:

.. code-block:: javascript

    Raven
      .config('___PUBLIC_DSN___', { release: RELEASE_ID })
      .install();

RELEASE_ID is a string representing the “version” of the build you are
about to distribute. This can be the SHA of your Git repository’s HEAD. It
can also be a semantic version number (e.g. “1.1.2”), pulled from your
project’s package.json file. More below.

About Releases
--------------

Every time you build and distribute a new version of your React Native
app, you’ll want to create a new release inside Sentry.  This is for two
important reasons:

- You can associate errors tracked by Sentry with a particular build
- You can store your source files/source maps generated for each build inside Sentry

Unlike a normal web application where your JavaScript files (and source
maps) are served and hosted from a web server, your React Native code is
being served from the target device’s filesystem. So you’ll need to upload
both your **source code** AND **source maps** directly to Sentry, so that
we can generate handy stack traces for you to browse when examining
exceptions triggered by your application.


Generating and Uploading Source Files/Source Maps
-------------------------------------------------

To generate both an application JavaScript file (main.jsbundle) and source map for your project (main.jsbundle.map), use the react-native CLI tool:

.. code-block:: bash

    react-native bundle \
      --dev false \
      --platform ios \
      --entry-file index.ios.js \
      --bundle-output main.jsbundle \
      --sourcemap-output main.jsbundle.map

This will write both main.jsbundle and main.jsbundle.map to the current directory. Next, you'll need to `create a new release and upload these files as release artifacts
<https://docs.sentry.io/hosted/clients/javascript/sourcemaps/#uploading-source-maps-to-sentry>`__.

Naming your Artifacts
~~~~~~~~~~~~~~~~~~~~~

In Sentry, artifacts are designed to be "named" using the full URL or path at which that artifact is located (e.g. `https://example.org/app.js` or `/path/to/file.js/`).
Since React Native applications are installed to a user's device, on a path that includes unique device identifiers (and thus different for every user),
the React Native plugin strips the entire path leading up to your application root.

This means that although your code may live at the following path:

.. code::

    /var/containers/Bundle/Application/{DEVICE_ID}/HelloWorld.app/main.jsbundle

The React Native plugin will reduce this to:

.. code::

    /main.jsbundle

Therefore in this example, you should name your artifacts as "/main.jsbundle" and "/main.jsbundle.map".

Source Maps with the Simulator
------------------------------

When developing with the simulator, it is not necessary to build source maps manually, as they are generated automatically on-demand.

Note however that artifact names are completely different when using the simulator. This is because instead of those files existing
on a path on a device, they are served over HTTP via the `React Native packager
<https://github.com/facebook/react-native/tree/master/packager>`__.

Typically, simulator assets are served at the following URLs:

- Bundle: http://localhost:8081/index.ios.bundle?platform=ios&dev=true
- Source map: http://localhost:8081/index.ios.map?platform=ios&dev=true

If you want to evaluate Sentry's source map support using the simulator, you will need to fetch these assets at these URLs (while the React Native
packager is running), and upload them to Sentry as artifacts. They should be named using the full URL at which they are located, including
the query string.


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
