/**
 * native plugin
 *
 * Extends support for global error handling for asynchronous browser
 * functions. Adopted from Closure Library's errorhandler.js.
 */
;(function extendToAsynchronousCallbacks(window, Raven) {
"use strict";

var _helper = function _helper(fnName) {
    var originalFn = window[fnName];
    window[fnName] = function ravenAsyncExtension() {
        // Make a copy of the arguments
        var args = [].slice.call(arguments);
        var originalCallback = args[0];
        if (typeof (originalCallback) === 'function') {
            args[0] = Raven.wrap(originalCallback);
        }
        // IE < 9 doesn't support .call/.apply on setInterval/setTimeout, but it
        // also supports only two arguments and doesn't care what this is, so we
        // can just call the original function directly.
        if (originalFn.apply) {
            return originalFn.apply(this, args);
        } else {
            return originalFn(args[0], args[1]);
        }
    };
};

_helper('setTimeout');
_helper('setInterval');

}(window, window.Raven));
