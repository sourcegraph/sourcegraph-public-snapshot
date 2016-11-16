Usage
=====

By default, Raven makes a few efforts to try its best to capture
meaningful stack traces, but browsers make it pretty difficult.

The easiest solution is to prevent an error from bubbling all of the way
up the stack to ``window``.

.. _raven-js-reporting-errors:

Reporting Errors Correctly
--------------------------

There are different methods to report errors and this all depends a little
bit on circumstances.

try … catch
```````````

The simplest way, is to try and explicitly capture and report potentially
problematic code with a ``try...catch`` block and
``Raven.captureException``.

.. code-block:: javascript

    try {
        doSomething(a[0])
    } catch(e) {
        Raven.captureException(e)
    }

**Do not** throw strings! Always throw an actual ``Error`` object. For
example:

.. code-block:: javascript

    throw new Error('broken')  // good
    throw 'broken'  // bad

It's impossible to retrieve a stack trace from a string. If this happens,
Raven transmits the error as a plain message.

context/wrap
````````````

``Raven.context`` allows you to wrap any function to be immediately
executed.  Behind the scenes, Raven is just wrapping your code in a
``try...catch`` block to record the exception before re-throwing it.

.. code-block:: javascript

    Raven.context(function() {
        doSomething(a[0])
    })

``Raven.wrap`` wraps a function in a similar way to ``Raven.context``, but
instead of executing the function, it returns another function.  This is
especially useful when passing around a callback.

.. code-block:: javascript

    var doIt = function() {
        // doing cool stuff
    }

    setTimeout(Raven.wrap(doIt), 1000)

Tracking Users
--------------

While a user is logged in, you can tell Sentry to associate errors with
user data.

.. code-block:: javascript

    Raven.setUserContext({
        email: 'matt@example.com',
        id: '123'
    })

If at any point, the user becomes unauthenticated, you can call
``Raven.setUserContext()`` with no arguments to remove their data. *This
would only really be useful in a large web app where the user logs in/out
without a page reload.*

This data is generally submitted with each error or message and allows you
to figure out which errors are affected by problems.

Capturing Messages
------------------

.. code-block:: javascript

    Raven.captureMessage('Broken!')

.. _raven-js-additional-context:

Passing Additional Data
-----------------------

The ``captureMessage``, ``captureException``, ``context``, and ``wrap``
functions all allow passing additional data to be tagged onto the error.

.. describe:: level

    The log level associated with this event. Default: ``error``

    .. code-block:: javascript

        Raven.captureMessage('Something happened', {
          level: 'info' // one of 'info', 'warning', or 'error'
        });

.. describe:: logger

    The name of the logger used to record this event. Default: ``javascript``

    .. code-block:: javascript

        Raven.captureException(new Error('Oops!'), {
          logger: 'my.module'
        });

    Note that logger can also be set globally via ``Raven.config``.

.. describe:: tags

    `Tags <https://docs.sentry.io/hosted/learn/context/#tagging-events>`__ to assign to the event.

    .. code-block:: javascript

        Raven.wrap({
          tags: {git_commit: 'c0deb10c4'}
        }, function () { /* ... */ });

        // NOTE: Raven.wrap and Raven.context accept options as first argument

    You can also set tags globally to be merged in with future exceptions events via ``Raven.config``, or ``Raven.setTagsContext``:

    .. code-block:: javascript

        Raven.setTagsContext({ key: "value" });

.. describe:: extra

    Arbitrary data to associate with the event.

    .. code-block:: javascript

        Raven.context({
          extra: {planet: {name: 'Earth'}}
        }, function () { /* ... */ });

        // NOTE: Raven.wrap and Raven.context accept options as first argument

    You can also set extra data globally to be merged in with future events with ``setExtraContext``:

    .. code-block:: javascript

        Raven.setExtraContext({ foo: "bar" })


Getting Back an Event ID
------------------------

An event id is a globally unique id for the event that was just sent. This
event id can be used to find the exact event from within Sentry.

This is often used to display for the user and report an error to customer
service.

.. code-block:: javascript

    Raven.lastEventId()

``Raven.lastEventId()`` will be undefined until an event is sent. After an
event is sent, it will contain the string id.

.. code-block:: javascript

    Raven.captureMessage('Broken!')
    alert(Raven.lastEventId())

.. _javascript-user-feedback:

User Feedback
-------------

Often you might find yourself wanting to collect additional feedback from
the user. Sentry supports this via an embeddable widget.

.. sourcecode:: javascript

    try {
        handleRouteChange(...)
    } catch (err) {
        Raven.captureException(err);
        Raven.showReportDialog();
    }

For more details on this feature, see the :doc:`User Feedback guide <../../learn/user-feedback>`.


Verify Raven Setup
------------------

If you need to conditionally check if raven needs to be initialized or
not, you can use the `isSetup` function.  It will return `true` if Raven
is already initialized:

.. code-block:: javascript

    Raven.isSetup()


.. _raven-js-source-maps:

Dealing with Minified Source Code
---------------------------------

Raven and Sentry support `Source Maps
<http://www.html5rocks.com/en/tutorials/developertools/sourcemaps/>`_.

We have provided some instructions to creating Source Maps over at
https://docs.sentry.io/hosted/clients/javascript/sourcemaps/. Also, checkout our `Gruntfile
<https://github.com/getsentry/raven-js/blob/master/Gruntfile.js>`_ for a
good example of what we're doing.

You can use `Source Map Validator
<https://sourcemaps.io/>`_ to help verify that things
are correct.

CORS
----

If you're hosting your scripts on another domain and things don't get
caught by Raven, it's likely that the error will bubble up to
``window.onerror``. If this happens, the error will report some ugly
``Script error`` and Raven will drop it on the floor since this is a
useless error for everybody.

To help mitigate this, we can tell the browser that these scripts are safe
and we're allowing them to expose their errors to us.

In your ``<script>`` tag, specify the ``crossorigin`` attribute:

.. code-block:: html

    <script src="//cdn.example.com/script.js" crossorigin="anonymous"></script>

And set an ``Access-Control-Allow-Origin`` HTTP header on that file.

.. code-block:: console

  Access-Control-Allow-Origin: *

.. note:: both of these steps need to be done or your scripts might not
   even get executed

Promises
--------

By default, Raven.js does not capture unhandled promise rejections.

Most Promise libraries have a global hook for capturing unhandled errors. You will need to
manually hook into such an event handler and call ``Raven.captureException`` or ``Raven.captureMessage``
directly.

For example, the `RSVP.js library
<https://github.com/tildeio/rsvp.js/>`_ (used by Ember.js) allows you to bind an event handler to a `global error event
<https://github.com/tildeio/rsvp.js#error-handling>`_:

.. code-block:: javascript

    RSVP.on('error', function(reason) {
        Raven.captureException(reason);
    });

`Bluebird
<http://bluebirdjs.com/>`_ and other promise libraries report unhandled rejections to a global DOM event, ``unhandledrejection``:

.. code-block:: javascript

    window.onunhandledrejection = function(evt) {
        Raven.captureException(evt.reason);
    };

Please consult your promise library documentation on how to hook into its global unhandled rejection handler, if it exposes one.

Custom Grouping Behavior
------------------------

In some cases you may see issues where Sentry groups multiple events together
when they should be separate entities. In other cases, Sentry simply doesn't
group events together because they're so sporadic that they never look the same.

Both of these problems can be addressed by specifying the ``fingerprint``
attribute.

For example, if you have HTTP 404 (page not found) errors, and you'd prefer they
deduplicate by taking into account the URL:

.. code-block:: javascript

    Raven.captureException(ex, {fingerprint: ['{{ default }}', 'http://my-url/']});

.. sentry:edition:: hosted, on-premise

    For more information, see :ref:`custom-grouping`.

Preventing Abuse
----------------

By default, the Sentry server accepts errors from any host. This can lead to an abuse
scenario where a malicious party triggers JavaScript errors from a different website that are
accepted by your Sentry Project. To prevent this, it is recommended to whitelist known hosts where your
JavaScript code is operating.

This setting can be found under the **Project Settings** page in Sentry. You'll need
to add each domain that you plan to report from into the **Allowed Domains**
box. When an error is collected by Raven.js and transmitted to Sentry, Sentry will verify the ``Origin`` and/or
``Referer`` headers of the HTTP request to verify that it matches one of your allowed hosts.
