Installation
============

Raven is distributed in a few different methods, and should get included after any other libraries are included, but before your own scripts.

So for example:

.. parsed-literal::

    <script src="jquery.js"></script>
    <script src="//cdn.ravenjs.com/|release|/jquery,native/raven.min.js"></script>
    <script>Raven.config('...').install();</script>
    <script src="app.js"></script>

This allows the ability for Raven's plugins to instrument themselves. If included before something like jQuery, it'd be impossible to use for example, the jquery plugin.

Using our CDN
~~~~~~~~~~~~~

We serve our own builds off of `Fastly <http://www.fastly.com/>`_. They are accessible over both http and https, so we recommend leaving the protocol off.

Our CDN distributes builds with and without :doc:`plugins </plugins/index>`.

.. parsed-literal::

    <script src="//cdn.ravenjs.com/|release|/raven.min.js"></script>

**We highly recommend trying out a plugin or two since it'll greatly improve the chances that we can collect good information.**

This version does not include any plugins. See `ravenjs.com <http://ravenjs.com/>`_ for more information about plugins and getting other builds.

Bower
~~~~~

We also provide a way to deploy Raven via `bower
<http://bower.io/>`_. Useful if you want serve your own scripts instead of depending on our CDN and mantain a ``bower.json`` with a list of dependencies and versions (adding the ``--save`` flag would automatically add it to ``bower.json``).

.. code-block:: sh

    $ bower install raven-js --save

.. code-block:: html

    <script src="/bower_components/raven-js/dist/raven.js"></script>

Also note that the file is uncompresed but is ready to pass to any decent JavaScript compressor like `uglify <https://github.com/mishoo/UglifyJS2>`_.

npm
~~~

Raven is published to npm as well. https://www.npmjs.com/package/raven-js

.. code-block:: sh

    $ npm install raven-js --save

Requirements
~~~~~~~~~~~~

Raven expects the browser to provide `window.JSON` and `window.JSON.stringify`. In Internet Explorer 8+ these are available in `standards mode <http://msdn.microsoft.com/en-us/library/cc288325(VS.85).aspx>`_.
You can also use `json2.js <https://github.com/douglascrockford/JSON-js>`_ to provide the JSON implementation in browsers/modes which doesn't support native JSON
