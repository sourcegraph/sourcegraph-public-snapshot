Tips and Tricks
===============

These are some general recommendations and tips for how to get the most
out of Raven.js and Sentry.

Decluttering Sentry
-------------------

The first thing to do is to consider constructing a whitelist of domains
in which might raise acceptable exceptions.

If your scripts are loaded from ``cdn.example.com`` and your site is
``example.com`` it'd be reasonable to set ``whitelistUrls`` to:

.. code-block:: javascript

    whitelistUrls: [
      /https?:\/\/((cdn|www)\.)?example\.com/
    ]

Since this accepts a regular expression, that would catch anything
\*.example.com or example.com exactly. See also: :ref:`Config:
whitelistUrls<config-whitelist-urls>`.

Next, checkout the list of :doc:`integrations <integrations/index>` we provide and see
which are applicable to you.

The community has compiled a list of common ignore rules for common
things, like Facebook, Chrome extensions, etc. So it's recommended to at
least check these out and see if they apply to you. `Check out the
original gist <https://gist.github.com/impressiver/5092952>`_.

.. code-block:: javascript

    var ravenOptions = {
        ignoreErrors: [
          // Random plugins/extensions
          'top.GLOBALS',
          // See: http://blog.errorception.com/2012/03/tale-of-unfindable-js-error. html
          'originalCreateNotification',
          'canvas.contentDocument',
          'MyApp_RemoveAllHighlights',
          'http://tt.epicplay.com',
          'Can\'t find variable: ZiteReader',
          'jigsaw is not defined',
          'ComboSearch is not defined',
          'http://loading.retry.widdit.com/',
          'atomicFindClose',
          // Facebook borked
          'fb_xd_fragment',
          // ISP "optimizing" proxy - `Cache-Control: no-transform` seems to
          // reduce this. (thanks @acdha)
          // See http://stackoverflow.com/questions/4113268
          'bmi_SafeAddOnload',
          'EBCallBackMessageReceived',
          // See http://toolbar.conduit.com/Developer/HtmlAndGadget/Methods/JSInjection.aspx
          'conduitPage'
        ],
        ignoreUrls: [
          // Facebook flakiness
          /graph\.facebook\.com/i,
          // Facebook blocked
          /connect\.facebook\.net\/en_US\/all\.js/i,
          // Woopra flakiness
          /eatdifferent\.com\.woopra-ns\.com/i,
          /static\.woopra\.com\/js\/woopra\.js/i,
          // Chrome extensions
          /extensions\//i,
          /^chrome:\/\//i,
          // Other plugins
          /127\.0\.0\.1:4001\/isrunning/i,  // Cacaoweb
          /webappstoolbarba\.texthelp\.com\//i,
          /metrics\.itunes\.apple\.com\.edgesuite\.net\//i
        ]
    };


Sampling Data
-------------

It happens frequently that errors sent from your frontend can be
overwhelming. One solution here is to only send a sample of the events
that happen. You can do this via the ``shouldSendCallback`` setting:

.. code-block:: javascript

    shouldSendCallback: function(data) {
        // only send 10% of errors
        var sampleRate = 10;
        return (Math.random() * 100 <= sampleRate);
    }

jQuery AJAX Error Reporting
---------------------------

For automatically reporting AJAX errors from jQuery, the following tips
might be helpful, however depending on the type of request you might
have to do slightly different things.

Same Origin
-----------

Whenever an Ajax request completes with an error, jQuery triggers the
``ajaxError`` event, passing the ``event`` object, the ``jqXHR`` object
(prior to jQuery 1.5, the ``XHR`` object), and the ``settings`` object
that was used in the creation of the request. When an HTTP error occurs,
the fourth argument (``thrownError``) receives the textual portion of the
HTTP status, such as "Not Found" or "Internal Server Error."

You can use this event to globally handle Ajax errors:

.. code-block:: javascript

    $(document).ajaxError(function(event, jqXHR, ajaxSettings, thrownError) {
        Raven.captureMessage(thrownError || jqXHR.statusText, {
            extra: {
                type: ajaxSettings.type,
                url: ajaxSettings.url,
                data: ajaxSettings.data,
                status: jqXHR.status,
                error: thrownError || jqXHR.statusText,
                response: jqXHR.responseText.substring(0, 100)
            }
        });
    });


**Note:**

* This handler is not called for cross-domain script and cross-domain
  JSONP requests.
* If ``$.ajax()`` or ``$.ajaxSetup()`` is called with the ``global``
  option set to ``false``, the ``.ajaxError()`` method will not fire.
* As of jQuery 1.8, the ``.ajaxError()`` method should only be attached to
  document.


Cross Origin
------------

Due to security reasons most web browsers are not giving permissions to
access error messages for cross domain scripts. This is not jQuery issue
but an overall javascript limitation.

Depending on your situation you have different options now:

When you control the backend
````````````````````````````

If you have access to the backend system you are calling, you can set
response headers to allow a cross domain call:

.. code-block:: yaml

    Access-Control-Allow-Origin: *

Script tags have now got a new non-standard attribute called
``crossorigin`` (`read more
<https://developer.mozilla.org/en-US/docs/Web/HTML/Element/script#attr-crossorigin>`_).
The most secure value for this would be ``anonymous``. So, you'll have to
modify your script tags to look like the following:

.. code-block:: html

    <script src="http://sub.domain.com/script.js" crossorigin="anonymous"></script>

When you have no access to the backend
``````````````````````````````````````

If you have no access to the backend, you could try a workaround, which is
basically adding a timeout on the Ajax call. This is however very dirty,
and will fail on slow connection or long response time:

.. code-block:: javascript

    $.ajax({
        url: 'http:/mysite/leaflet.js',
        success: function() { ... },
        error: function() { ... },
        timeout: 2000, // 2 seconds timeout before error function will be called
        dataType: 'script',
        crossDomain: true
    });
