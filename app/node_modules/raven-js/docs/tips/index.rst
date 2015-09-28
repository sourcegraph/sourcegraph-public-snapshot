Pro Tips™
=========


Decluttering Sentry
~~~~~~~~~~~~~~~~~~~

The first thing to do is to consider constructing a whitelist of domains in which might raise acceptable exceptions.

If your scripts are loaded from ``cdn.example.com`` and your site is ``example.com`` it'd be reasonable to set ``whitelistUrls`` to:

.. code-block:: javascript

    whitelistUrls: [
      /https?:\/\/((cdn|www)\.)?example\.com/
    ]

Since this accepts a regular expression, that would catch anything \*.example.com or example.com exactly. See also: :ref:`Config: whitelistUrls<config-whitelist-urls>`.

Next, checkout the list of :doc:`plugins </plugins/index>` we provide and see which are applicable to you.

The community has compiled a list of common ignore rules for common things, like Facebook, Chrome extensions, etc. So it's recommended to at least check these out and see if they apply to you. `Check out the original gist <https://gist.github.com/impressiver/5092952>`_.

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
          // ISP "optimizing" proxy - `Cache-Control: no-transform` seems to reduce this. (thanks @acdha)
          // See http://stackoverflow.com/questions/4113268/how-to-stop-javascript-injection-from-vodafone-proxy
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
~~~~~~~~~~~~~

It happens frequently that errors sent from your frontend can be overwhelming. One solution here is to only send a sample of the events that happen. You can do this via the ``shouldSendCallback`` setting:

.. code-block:: javascript

    shouldSendCallback: function(data) {
        // only send 10% of errors
        var sampleRate = 10;
        return (Math.random() * 100 <= sampleRate);
    }
