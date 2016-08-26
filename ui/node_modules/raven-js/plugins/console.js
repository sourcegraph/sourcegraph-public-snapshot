/**
 * console plugin
 *
 * Monkey patches console.* calls into Sentry messages with
 * their appropriate log levels. (Experimental)
 *
 * Options:
 *
 *   `levels`: An array of levels (methods on `console`) to report to Sentry.
 *     Defaults to debug, info, warn, and error.
 */
'use strict';

var wrapConsoleMethod = require('../src/console').wrapMethod;

function consolePlugin(Raven, console, pluginOptions) {
    console = console || window.console || {};
    pluginOptions = pluginOptions || {};

    var logLevels = pluginOptions.levels || ['debug', 'info', 'warn', 'error'],
        level = logLevels.pop();

    var callback = function (msg, data) {
        Raven.captureMessage(msg, data);
    };

    while(level) {
        wrapConsoleMethod(console, level, callback);
        level = logLevels.pop();
    }
}

module.exports = consolePlugin;
