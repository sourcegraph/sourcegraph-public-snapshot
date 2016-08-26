Contributing
============

This part of the documentation gives you a basic overview of how to help
with the development of Raven.js.

Setting up an Environment
-------------------------

To run the test suite and run our code linter, node.js and npm are
required. If you don't have node installed, `get it here
<http://nodejs.org/download/>`_ first.

Installing all other dependencies is as simple as:

.. code-block:: sh

    $ npm install

And if you don't have `Grunt <http://gruntjs.com/>`_ already, feel free to
install that globally:

.. code-block:: sh

    $ npm install -g grunt-cli

Running the Test Suite
~~~~~~~~~~~~~~~~~~~~~~

The test suite is powered by `Mocha
<http://visionmedia.github.com/mocha/>`_ and can both run from the command
line, or in the browser.

From the command line:

.. code-block:: sh

    $ grunt test

From your browser:

.. code-block:: sh

    $ grunt run:test

Then visit: http://localhost:8000/test/

Compiling Raven.js
~~~~~~~~~~~~~~~~~~

The simplest way to compile your own version of Raven.js is with the
supplied grunt command:

.. code-block:: sh

    $ grunt build

By default, this will compile raven.js and all of the included plugins.

If you only want to compile the core raven.js:

.. code-block:: sh

    $ grunt build.core

Files are compiled into ``build/``.

Contributing Back Code
~~~~~~~~~~~~~~~~~~~~~~

Please, send over suggestions and bug fixes in the form of pull requests
on `GitHub <https://github.com/getsentry/raven-js>`_. Any nontrivial
fixes/features should include tests.  Do not include any changes to the
``dist/`` folder or bump version numbers yourself.

Documentation
-------------

The documentation is written using `reStructuredText
<http://en.wikipedia.org/wiki/ReStructuredText>`_, and compiled using
`Sphinx <http://sphinx-doc.org/>`_. If you don't have Sphinx installed,
you can do it using following command (assuming you have Python already
installed in your system):

.. code-block:: sh

    $ pip install sphinx

Documentation can be then compiled by running:

.. code-block:: sh

    $ make docs

Afterwards you can view it in your browser by running following command
and than pointing your browser to http://127.0.0.1:8000/:

.. code-block:: sh

    $ grunt run:docs


Releasing New Version
~~~~~~~~~~~~~~~~~~~~~

* Bump version numbers in ``package.json``, ``bower.json``, and ``src/raven.js``.
* ``$ grunt dist`` This will compile a new version and update it in the
  ``dist/`` folder.
* Confirm that build was fine, etc.
* Commit new version, create a tag. Push to GitHub.
* ``$ grunt publish`` to recompile all plugins and all permutations and
  upload to S3.
* ``$ npm publish`` to push to npm.
* Confirm that the new version exists behind ``cdn.ravenjs.com``
* Update version in the ``gh-pages`` branch specifically for
  http://ravenjs.com/.
* glhf
