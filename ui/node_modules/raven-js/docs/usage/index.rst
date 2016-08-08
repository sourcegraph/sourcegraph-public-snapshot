Usage
=====

By default, Raven makes a few efforts to try its best to capture meaningful stack traces, but browsers make it pretty difficult.

The easiest solution is to prevent an error from bubbling all of the way up the stack to ``window``.

How to actually capture an error correctly
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

try...catch
-----------

The simplest way, is to try and explicitly capture and report potentially problematic code with a ``try...catch`` block and ``Raven.captureException``.

.. code-block:: javascript

    try {
        doSomething(a[0])
    } catch(e) {
        Raven.captureException(e)
    }

**Do not** throw strings! Always throw an actual ``Error`` object. For example:

.. code-block:: javascript

    throw new Error('broken')  // good
    throw 'broken'  // bad

It's impossible to retrieve a stack trace from a string. If this happens, Raven transmits the error as a plain message.

context/wrap
------------

``Raven.context`` allows you to wrap any function to be immediately executed. Behind the scenes, Raven is just wrapping your code in a ``try...catch`` block to record the exception before re-throwing it.

.. code-block:: javascript

    Raven.context(function() {
        doSomething(a[0])
    })

``Raven.wrap`` wraps a function in a similar way to ``Raven.context``, but instead of executing the function, it returns another function. This is totally awesome for use when passing around a callback.

.. code-block:: javascript

    var doIt = function() {
        // doing cool stuff
    }

    setTimeout(Raven.wrap(doIt), 1000)

Tracking authenticated users
~~~~~~~~~~~~~~~~~~~~~~~~~~~~

While a user is logged in, you can tell Sentry to associate errors with user data.

.. code-block:: javascript

    Raven.setUserContext({
        email: 'matt@example.com',
        id: '123'
    })

If at any point, the user becomes unauthenticated, you can call ``Raven.setUserContext()`` with no arguments to remove their data. *This would only really be useful in a large web app where the user logs in/out without a page reload.*

Capturing a specific message
~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: javascript

    Raven.captureMessage('Broken!')

Passing additional data
~~~~~~~~~~~~~~~~~~~~~~~

``captureException``, ``context``, ``wrap``, and ``captureMessage`` functions all allow passing additional data to be tagged onto the error, such as ``tags`` or ``extra`` for additional context.

.. code-block:: javascript

    Raven.captureException(e, {tags: { key: "value" }})

    Raven.captureMessage('Broken!', {tags: { key: "value" }})

    Raven.context({tags: { key: "value" }}, function(){ ... })

    Raven.wrap({logger: "my.module"}, function(){ ... })

    Raven.captureException(e, {extra: { foo: "bar" }})

You can also set context variables globally to be merged in with future exceptions with ``setExtraContext`` and ``setTagsContext``.

.. code-block:: javascript

    Raven.setExtraContext({ foo: "bar" })
    Raven.setTagsContext({ key: "value" })


Getting back an event id
~~~~~~~~~~~~~~~~~~~~~~~~

An event id is a globally unique id for the event that was just sent. This event id can be used to find the exact event from within Sentry.

This is often used to display for the user and report an error to customer service.

.. code-block:: javascript

    Raven.lastEventId()

``Raven.lastEventId()`` will be undefined until an event is sent. After an event is sent, it will contain the string id.

.. code-block:: javascript

    Raven.captureMessage('Broken!')
    alert(Raven.lastEventId())


Check if Raven is setup and ready to go
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: javascript

    Raven.isSetup()


Dealing with minified source code
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

Raven and Sentry now support `Source Maps <http://www.html5rocks.com/en/tutorials/developertools/sourcemaps/>`_.

We have provided some instructions to creating Source Maps over at https://www.getsentry.com/docs/sourcemaps/. Also, checkout our `Gruntfile <https://github.com/getsentry/raven-js/blob/master/Gruntfile.js>`_ for a good example of what we're doing.

You can use `Source Map Validator <http://sourcemap-validator.herokuapp.com/>`_ to help verify that things are correct.

CORS
~~~~

If you're hosting your scripts on another domain and things don't get caught by Raven, it's likely that the error will bubble up to ``window.onerror``. If this happens, the error will report some ugly ``Script error`` and Raven will drop it on the floor
since this is a useless error for everybody.

To help mitigate this, we can tell the browser that these scripts are safe and we're allowing them to expose their errors to us.

In your ``<script>`` tag, specify the ``crossorigin`` attribute:

.. code-block:: html

    <script src="//cdn.example.com/script.js" crossorigin="anonymous"></script>

And set an ``Access-Control-Allow-Origin`` HTTP header on that file.

.. code-block:: console

  Access-Control-Allow-Origin: *

**Note: both of these steps need to be done or your scripts might not even get executed**
