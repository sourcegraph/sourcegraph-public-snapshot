/*global ErrorUtils:false*/

/**
 * react-native plugin for Raven
 *
 * Usage:
 *   var Raven = require('raven-js');
 *   Raven.addPlugin(require('raven-js/plugins/react-native'));
 *
 * Options:
 *
 *   pathStrip: A RegExp that matches the portions of a file URI that should be
 *     removed from stacks prior to submission.
 *
 *   onInitialize: A callback that fires once the plugin has fully initialized
 *     and checked for any previously thrown fatals.  If there was a fatal, its
 *     data payload will be passed as the first argument of the callback.
 *
 */
'use strict';

// Example React Native path format (iOS):
// /var/containers/Bundle/Application/{DEVICE_ID}/HelloWorld.app/main.jsbundle

var PATH_STRIP_RE = /^.*\/[^\.]+(\.app|CodePush)/;

var FATAL_ERROR_KEY = '--rn-fatal--';
var ASYNC_STORAGE_KEY = '--raven-js-global-error-payload--';

/**
 * Strip device-specific IDs from React Native file:// paths
 */
function normalizeUrl(url, pathStripRe) {
    return url
        .replace(/^file\:\/\//, '')
        .replace(pathStripRe, '');
}

/**
 * Extract key/value pairs from an object and encode them for
 * use in a query string
 */
function urlencode(obj) {
    var pairs = [];
    for (var key in obj) {
        if ({}.hasOwnProperty.call(obj, key))
            pairs.push(encodeURIComponent(key) + '=' + encodeURIComponent(obj[key]));
    }
    return pairs.join('&');
}

/**
 * Initializes React Native plugin
 */
function reactNativePlugin(Raven, options) {
    options = options || {};

    // react-native doesn't have a document, so can't use default Image
    // transport - use XMLHttpRequest instead
    Raven.setTransport(reactNativePlugin._transport);

    // Use data callback to strip device-specific paths from stack traces
    Raven.setDataCallback(function(data) {
        reactNativePlugin._normalizeData(data, options.pathStrip)
    });

    // Check for a previously persisted payload, and report it.
    reactNativePlugin._restorePayload()
        .then(function(payload) {
            options.onInitialize && options.onInitialize(payload);
            if (!payload) return;
            Raven._sendProcessedPayload(payload, function(error) {
                if (error) return; // Try again next launch.
                reactNativePlugin._clearPayload();
            });
        })
        ['catch'](function() {});

    // Make sure that if multiple fatals occur, we only persist the first one.
    //
    // The first error is probably the most important/interesting error, and we
    // want to crash ASAP, rather than potentially queueing up multiple errors.
    var handlingFatal = false;

    var defaultHandler = ErrorUtils.getGlobalHandler && ErrorUtils.getGlobalHandler() || ErrorUtils._globalHandler;

    Raven.setShouldSendCallback(function(data, originalCallback) {
        if (!(FATAL_ERROR_KEY in data)) {
            // not a fatal (will not crash runtime), continue as planned
            return originalCallback ? originalCallback.call(this, data) : true;
        }

        var origError = data[FATAL_ERROR_KEY];
        delete data[FATAL_ERROR_KEY];

        reactNativePlugin._persistPayload(data)
            .then(function() {
                defaultHandler(origError, true);
                handlingFatal = false; // In case it isn't configured to crash.
                return null;
            })
            ['catch'](function() {});

        return false; // Do not continue.
    });

    ErrorUtils.setGlobalHandler(function(error, isFatal) {
        var captureOptions = {
            timestamp: new Date() / 1000
        };
        var error = arguments[0];
        // We want to handle fatals, but only in production mode.
        var shouldHandleFatal = isFatal && !global.__DEV__;
        if (shouldHandleFatal) {
            if (handlingFatal) {
                console.log('Encountered multiple fatals in a row. The latest:', error);
                return;
            }
            handlingFatal = true;
            // We need to preserve the original error so that it can be rethrown
            // after it is persisted (see our shouldSendCallback above).
            captureOptions[FATAL_ERROR_KEY] = error;
        }
        Raven.captureException(error, captureOptions);
        // Handle non-fatals regularly.
        if (!shouldHandleFatal) {
            defaultHandler(error);
        }
    });
}

/**
 * Saves the payload for a globally-thrown error, so that we can report it on
 * next launch.
 *
 * Returns a promise that guarantees never to reject.
 */
reactNativePlugin._persistPayload = function(payload) {
    var AsyncStorage = require('react-native').AsyncStorage;
    return AsyncStorage.setItem(ASYNC_STORAGE_KEY, JSON.stringify(payload))
        ['catch'](function() { return null; });
}

/**
 * Checks for any previously persisted errors (e.g. from last crash)
 *
 * Returns a promise that guarantees never to reject.
 */
reactNativePlugin._restorePayload = function() {
    var AsyncStorage = require('react-native').AsyncStorage;
    var promise = AsyncStorage.getItem(ASYNC_STORAGE_KEY)
        .then(function(payload) { return JSON.parse(payload); })
        ['catch'](function() { return null; });
    // Make sure that we fetch ASAP.
    AsyncStorage.flushGetRequests();

    return promise;
};

/**
 * Clears any persisted payloads.
 */
reactNativePlugin._clearPayload = function() {
    var AsyncStorage = require('react-native').AsyncStorage;
    return AsyncStorage.removeItem(ASYNC_STORAGE_KEY)
        ['catch'](function() { return null; });
}

/**
 * Custom HTTP transport for use with React Native applications.
 */
reactNativePlugin._transport = function (options) {
    var request = new XMLHttpRequest();
    request.onreadystatechange = function (e) {
        if (request.readyState !== 4) {
            return;
        }

        if (request.status === 200) {
            if (options.onSuccess) {
                options.onSuccess();
            }
        } else {
            if (options.onError) {
                options.onError(new Error('Sentry error code: ' + request.status));
            }
        }
    };

    request.open('POST', options.url + '?' + urlencode(options.auth));

    // NOTE: React Native ignores CORS and will NOT send a preflight
    //       request for application/json.
    // See: https://facebook.github.io/react-native/docs/network.html#xmlhttprequest
    request.setRequestHeader('Content-type', 'application/json');

    // Sentry expects an Origin header when using HTTP POST w/ public DSN.
    // Just set a phony Origin value; only matters if Sentry Project is configured
    // to whitelist specific origins.
    request.setRequestHeader('Origin', 'react-native://');
    request.send(JSON.stringify(options.data));
};

/**
 * Strip device-specific IDs found in culprit and frame filenames
 * when running React Native applications on a physical device.
 */
reactNativePlugin._normalizeData = function (data, pathStripRe) {
    if (!pathStripRe) {
        pathStripRe = PATH_STRIP_RE;
    }

    if (data.culprit) {
        data.culprit = normalizeUrl(data.culprit, pathStripRe);
    }

    if (data.exception) {
        // if data.exception exists, all of the other keys are guaranteed to exist
        data.exception.values[0].stacktrace.frames.forEach(function (frame) {
            frame.filename = normalizeUrl(frame.filename, pathStripRe);
        });
    }
};

module.exports = reactNativePlugin;
