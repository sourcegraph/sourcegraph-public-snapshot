.. _raven-js-sourcemaps:

Source Maps
===========

Sentry supports un-minifying JavaScript via `Source Maps
<http://blog.getsentry.com/2015/10/29/debuggable-javascript-with-source-maps.html>`_. This lets you
view source code context obtained from stack traces in their original untransformed form, which is
is particularly useful for debugging minified code (e.g. UglifyJS), or transpiled code from a higher-level
language (e.g. TypeScript, ES6).

Generating a Source Map
-----------------------

Most modern JavaScript transpilers support source maps. Below are instructions for some common tools.

UglifyJS
~~~~~~~~

UglifyJS is a popular tool for minifying your source code for production. It can dramatically
reduce the size of your files by eliminating whitespace, rewriting variable names, removing dead code branches,
and more.

If you are using UglifyJS to minify your source code, the following command will additionally generate a source map
that maps the minified code back to the original source:

::

    node_modules/uglify-js/bin/uglifyjs {input} \
      --source-map-root={relroot}/ \
      --source-map-url={name}.map.js \
      --source-map={relpath}/{name}.map.js \
      --source-map-include-sources \
      -o {output}


Webpack
~~~~~~~

Webpack is a powerful build tool that resolves and bundles your JavaScript modules into files fit for running in the
browser. It also supports many different "loaders" which can convert higher-level languages like TypeScript and
ES6/ES2015 into browser-compatible JavaScript.

Webpack can be configured to output source maps by editing webpack.config.js.

::

    module.exports = {
        // ... other config above ...
        entry: {
          "app": "src/app.js"
        },
        output: {
          path: path.join(__dirname, 'dist'),
          filename: "[name].js",
          sourceMapFilename: "[name].map.js",
        }
    };

SystemJS
~~~~~~~~

SystemJS is the default module loader for Angular 2 projects. The `SystemJS build tool
<https://github.com/systemjs/builder>`_ can be used to bundle, transpile, and minify source code for
use in production environments, and can be configured to output source maps.

::

    builder.bundle('src/app.js', 'dist/app.min.js', {
        minify: true,
        sourceMaps: true,
        sourceMapContents: true
    });

.. admonition:: Inline Sources

    All of the example configurations above inline your original, un-transformed source files into the generated source map file. Sentry
    requires both source map(s) **and** your original source files in order to perform reverse transformations. If you
    choose NOT to inline your source files, you must make those source files available to Sentry in *addition* to your
    source maps (see below).


Making Source Maps Available to Sentry
--------------------------------------

Source maps can be either:

1) Served publicly over HTTP alongside your source files.

2) Uploaded directly to Sentry (**recommended**).

Hosting Source Map Files
~~~~~~~~~~~~~~~~~~~~~~~~

By default, Sentry will look for source map directives in your compiled JavaScript files, which are located
on the last line and have the following format:

.. code-block:: javascript

    //# sourceMappingURL=<url>

When Sentry encounters such a directive, it will resolve the source map URL relative the source file in which
it is found, and attempt an HTTP request to fetch it.

So for example if you have a minified JavaScript file located at ``http://example.org/js/app.min.js``. And in that file,
on the last line, the following directive is found:

.. code-block:: javascript

    //# sourceMappingURL=app.map.js

Sentry will attempt to fetch ``app.map.js`` from http://example.org/js/app.map.js.

Alternatively, during source map generation you can specify a fully qualified URL where your source maps are located:

.. code-block:: javascript

    //# sourceMappingURL=http://example.org/js/app.map.js

While making source maps available to Sentry from your servers is the easiest integration, it is not always advisable:

* Sentry may not always be able to reach your servers.
* If you do not specify versions in your asset URLs, there may be a version mismatch
* The additional latency may mean that source mappings are not available for all errors.

For these reasons, it is recommended to upload source maps to Sentry beforehand (see below).

.. admonition:: Working Behind a Firewall

    While the recommended solution is to upload your source artifacts to Sentry, sometimes it’s necessary to allow communication from Sentry’s internal IPs. For more information on Sentry’s public IPs, :ref:`ip-ranges`.

Uploading Source Maps to Sentry
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

In many cases your application may sit behind firewalls or you simply
can't expose source code to the public. Sentry provides an abstraction
called **Releases** which you can attach source artifacts to.

The release API is intended to allow you to store source files (and
sourcemaps) within Sentry. This removes the requirement for them to be
web-accessible, and also removes any inconsistency that could come from
network flakiness (on either your end, or Sentry's end).

You can either interact with the API directly or you can upload sourcemaps
with the help of the Sentry CLI (:ref:`upload-sourcemaps-with-cli`).

* Start by creating a new authentication token under **[Account] > API**.
* Ensure you you have ``project:write`` selected under scopes.
* You'll use the Authorization header with the value of ``Bearer: {TOKEN}``
  with API requests.

Now you need to setup your build system to create a release, and attach
the various source files. You will want to upload all dist files (i.e. the
minified/shipped JS), the referenced sourcemaps, and the files that those
sourcemaps point to.

.. code-block:: bash

    # Create a new release
    $ curl https://app.getsentry.com/api/0/projects/:organization_slug/:project_slug/releases/ \
      -X POST \
      -H 'Authorization: Bearer {TOKEN}' \
      -H 'Content-Type: application/json'
      -d '{"version": "2da95dfb052f477380608d59d32b4ab9"}' \

    {
      "dateCreated": "2015-03-06T04:51:32.723Z",
      "version": "2da95dfb052f477380608d59d32b4ab9"
    }

When uploading the file, you'll need to reference it just as it would be referenced
if a browser (or filesystem) had to resolve its path. So for example, if your sourcemap
reference is just a relative path, it's **relative to the location of the referencing file**.

So for example, if you have ``http://example.com/app.min.js``, and the file contains the
reference to ``app.map.js``, the name of the uploaded file should be ``http://example.com/app.map.js``.

.. code-block:: bash

    # Upload a file for the given release
    $ curl https://app.getsentry.com/api/0/projects/:organization_slug/:project_slug/releases/2da95dfb052f477380608d59d32b4ab9/files/ \
      -X POST \
      -H 'Authorization: Bearer {TOKEN}' \
      -F file=@app.js.map \
      -F name="http://example.com/app.js.map"

    {
      "dateCreated": "2015-03-06T04:53:00.308Z",
      "headers": {
        "Content-Type": "application/octet-stream"
      },
      "id": "1",
      "name": "http://example.com/app.js.map",
      "sha1": "22591348ed129fe016c535654f6493737f0f9df6",
      "size": 452
    }

.. code-block:: bash

    # If you make a mistake, you can also simply clear out the release
    $ curl https://app.getsentry.com/api/0/projects/:organization_slug/:project_slug/releases/2da95dfb052f477380608d59d32b4ab9/ \
      -H 'Authorization: Bearer {TOKEN}' \
      -X DELETE

Additionally, you'll need to configure the client to send the ``release``:

.. code-block:: javascript

    Raven.config({
        release: '2da95dfb052f477380608d59d32b4ab9'
    });

Note: You dont *have* to upload the source files (ref'd by sourcemaps),
but without them the grouping algorithm will not be as strong, and the UI
will not show any contextual source.

Additional information can be found in the `Releases API documentation
<https://docs.getsentry.com/hosted/api/releases/>`_.


.. _assets_multiple_origins:

.. admonition:: Assets Accessible at Multiple Origins

    It's not uncommon for a web application to be accessible at multiple
    origins. For example:

    * Website is operable over both ``https`` and ``http``
    * Geolocated web addresses: e.g. ``https://us.example.com``, ``https://eu.example.com``
    * Multiple static CDNs: e.g. ``https://static1.example.com``, ``https://static2.example.com``
    * Customer-specific domains/subdomains

    In this situation, **identical** JavaScript and source map files may be located
    at two or more distinct origins. If you are dealing with such a deployment, you have
    two choices for naming your uploaded artifacts:

    1. Upload the same artifact multiple times with each possible URL where it appears, for example:

        * https://static1.example.com/js/app.js
        * https://static2.example.com/js/app.js

    2. Alternatively, you can omit the protocol + host and use a special tilde (~) prefixed path like so:

        ~/js/app.js

    The ~ prefix tells Sentry that for a given URL, **any** combination of protocol and hostname whose path is
    ``/js/app.js`` should use this artifact. **ONLY** use this method if your source/sourcemap files
    are identical at all possible protocol/hostname combinations. Note that Sentry will prioritize
    full URLs over tilde prefixed paths if found.


.. _upload-sourcemaps-with-cli:

Using Sentry CLI
----------------

You can also use the Sentry :ref:`sentry-cli` to manage releases and
sourcemaps on Sentry.  If you have it installed you can create releases
with the following command::

    $ sentry-cli releases -o MY_ORG -p MY_PROJECT new 2da95dfb052f477380608d59d32b4ab9

After you have run this, you can use the `files` command to automatically
add all javascript files and sourcemaps below a folder.  They are
automatically prefixed with a URL or your choice::

    $ sentry-cli releases -o MY_ORG -p MY_PROJECT files \
      2da95dfb052f477380608d59d32b4ab9 upload-sourcemaps --url-prefix \
      https://mydomain.invalid/static /path/to/assets

.. admonition:: Assets Accessible at Multiple Origins

    If you leave out the ``--url-prefix`` parameter the paths will be
    prefixed with ``~/`` automatically to support multi origin behavior.

All files that end with `.js` and `.map` below `/path/to/assets` are
automatically uploaded to the release `2da95dfb052f477380608d59d32b4ab9`
in this case.  If you want to use other extensions you can provide it with
the ``--ext`` parameter.

.. admonition:: Validating Sourcemaps with Sentry CLI

    Unfortunately it can be quite challenging to ensure that sourcemaps
    are actually valid themselves and uploaded correctly.  To ensure
    that everything is working as intended you can use the `--validate`
    flag when uploading sourcemaps which will attempt to locally parse the
    sourcemap and look up the references.  Note that there are known cases
    where the validate flag will indicate failures when the setup is
    correct (if you have references to external sourcemaps then the
    validation tool will indicate a failure).

    Here are some things you can check in addition to the validation step:

    *   Make sure that the URL prefix is correct for your files.  This is
        easy to get wrong.
    *   Make sure you upload the matching sourcemaps for your minimized
        files.
    *   Make sure that your minified files you have on your servers
        actually have references to your files.


.. sentry:edition:: hosted

Troubleshooting
---------------

Source maps can sometimes be tricky to get going. If you're having trouble, try the following tips.


Verify your source maps are built correctly
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

We maintain an online validation tool that can be used to test your source
(and sourcemaps) against: `sourcemaps.io <http://sourcemaps.io>`_.

Alternatively, if you are using Sentry CLI to upload source maps to Sentry, you can use the `--validate`
command line option to verify your source maps are correct.


Verify sourceMappingURL is present
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

Some CDNs automatically strip comments from static files, including JavaScript files. This can have the effect of stripping your JavaScript file of its ``sourceMappingURL`` directive, because it is considered a comment. For example, CloudFlare has a feature called `Auto-Minify
<https://blog.cloudflare.com/an-all-new-and-improved-autominify/>`_ which will strip ``sourceMappingURL`` if it is enabled.

Double-check that your deployed, final JavaScript files have ``sourceMappingURL`` present.

Verify artifact names match sourceMappingURL
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

When `uploading source maps to Sentry
<#uploading-source-maps-to-sentry>`_, you must name your source map files with the same name found in ``sourceMappingURL``.

For example, if you have the following in a minified application file, ``app.min.js``:

.. code-block:: javascript

    //-- end app.min.js
    //# sourceMappingURL=https://example.com/dist/js/app.min.js.map

Sentry will look for a matching artifact named exactly ``https://example.com/dist/js/app.min.js.map``.

Note also that Sentry will resolve relative paths. For example, if you have the following:

.. code-block:: JavaScript

    // -- end app.min.js (located at https://example.com/dist/js/app.min.js)
    //# sourceMappingURL=app.min.js.map

Sentry will resolve ``sourceMappingURL`` relative to ``https://example.com/dist/js/`` (the root path from which ``app.min.js`` was served). You will again need to name your source map with the full URL: ``https://example.com/dist/js/app.min.js.map``.

If you serve the same assets from multiple origins, you can also alternatively use our tilde (~) path prefix to ignore
matching against protocol + hostname. In which case, ``~/dist/js/app.min.js.map``, will also work. See: `Assets Accessible at Multiple Origins
<#assets-multiple-origins>`_.

Verify artifacts are uploaded before errors occur
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

Sentry expects that source code and source maps in a given release are uploaded to Sentry **before** errors occur in that release.

If you upload artifacts **after** an error is captured by Sentry, Sentry will not go back and retroactively apply any source annotations to those errors. Only new errors triggered after the artifact was uploaded will be affected.

Verify your source files are not too large
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

For an individual artifact, Sentry accepts a max filesize of **20 MB**.

Often users hit this limit because they are transmitting source files at an interim build stage. For example, after Webpack/Browserify has combined all
your source files, but before minification has taken place. If possible, send the original source files.
