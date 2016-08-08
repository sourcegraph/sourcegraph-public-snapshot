/**
 * jQuery plugin
 *
 * Patches event handler callbacks and ajax callbacks.
 */
;(function(window, Raven, $) {
'use strict';

// quit if jQuery isn't on the page
if (!$) {
    return;
}

var _oldEventAdd = $.event.add;
$.event.add = function ravenEventAdd(elem, types, handler, data, selector) {
    var _handler;

    if (handler && handler.handler) {
        _handler = handler.handler;
        handler.handler = Raven.wrap(handler.handler);
    } else {
        _handler = handler;
        handler = Raven.wrap(handler);
    }

    // If the handler we are attaching doesn’t have the same guid as
    // the original, it will never be removed when someone tries to
    // unbind the original function later. Technically as a result of
    // this our guids are no longer globally unique, but whatever, that
    // never hurt anybody RIGHT?!
    if (_handler.guid) {
        handler.guid = _handler.guid;
    } else {
        handler.guid = _handler.guid = $.guid++;
    }

    return _oldEventAdd.call(this, elem, types, handler, data, selector);
};

var _oldReady = $.fn.ready;
$.fn.ready = function ravenjQueryReadyWrapper(fn) {
    return _oldReady.call(this, Raven.wrap(fn));
};

var _oldAjax = $.ajax;
$.ajax = function ravenAjaxWrapper(url, options) {
    var keys = ['complete', 'error', 'success'], key;

    // Taken from https://github.com/jquery/jquery/blob/eee2eaf1d7a189d99106423a4206c224ebd5b848/src/ajax.js#L311-L318
    // If url is an object, simulate pre-1.5 signature
    if (typeof url === 'object') {
        options = url;
        url = undefined;
    }

    // Force options to be an object
    options = options || {};

    /*jshint -W084*/
    while(key = keys.pop()) {
        if ($.isFunction(options[key])) {
            options[key] = Raven.wrap(options[key]);
        }
    }
    /*jshint +W084*/

    try {
        return _oldAjax.call(this, url, options);
    } catch (e) {
        Raven.captureException(e);
        throw e;
    }
};

}(window, window.Raven, window.jQuery));
