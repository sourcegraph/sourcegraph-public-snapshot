"use strict";
var Console = require('console').Console;
var stderrConsole = new Console(process.stderr);
var stdoutConsole = new Console(process.stdout);
var LogLevel;
(function (LogLevel) {
    LogLevel[LogLevel["INFO"] = 1] = "INFO";
    LogLevel[LogLevel["WARN"] = 2] = "WARN";
    LogLevel[LogLevel["ERROR"] = 3] = "ERROR";
})(LogLevel || (LogLevel = {}));
var doNothingLogger = function () {
    var messages = [];
    for (var _i = 0; _i < arguments.length; _i++) {
        messages[_i - 0] = arguments[_i];
    }
};
function makeLoggerFunc(loaderOptions) {
    return loaderOptions.silent
        ? function (whereToLog, messages) { }
        : function (whereToLog, messages) { return console.log.apply(whereToLog, messages); };
}
function makeExternalLogger(loaderOptions, logger) {
    var output = loaderOptions.logInfoToStdOut ? stdoutConsole : stderrConsole;
    return function () {
        var messages = [];
        for (var _i = 0; _i < arguments.length; _i++) {
            messages[_i - 0] = arguments[_i];
        }
        return logger(output, messages);
    };
}
function makeLogInfo(loaderOptions, logger) {
    return LogLevel[loaderOptions.logLevel] <= LogLevel.INFO
        ? function () {
            var messages = [];
            for (var _i = 0; _i < arguments.length; _i++) {
                messages[_i - 0] = arguments[_i];
            }
            return logger(loaderOptions.logInfoToStdOut ? stdoutConsole : stderrConsole, messages);
        }
        : doNothingLogger;
}
function makeLogError(loaderOptions, logger) {
    return LogLevel[loaderOptions.logLevel] <= LogLevel.ERROR
        ? function () {
            var messages = [];
            for (var _i = 0; _i < arguments.length; _i++) {
                messages[_i - 0] = arguments[_i];
            }
            return logger(stderrConsole, messages);
        }
        : doNothingLogger;
}
function makeLogWarning(loaderOptions, logger) {
    return LogLevel[loaderOptions.logLevel] <= LogLevel.WARN
        ? function () {
            var messages = [];
            for (var _i = 0; _i < arguments.length; _i++) {
                messages[_i - 0] = arguments[_i];
            }
            return logger(stderrConsole, messages);
        }
        : doNothingLogger;
}
function makeLogger(loaderOptions) {
    var logger = makeLoggerFunc(loaderOptions);
    return {
        log: makeExternalLogger(loaderOptions, logger),
        logInfo: makeLogInfo(loaderOptions, logger),
        logWarning: makeLogWarning(loaderOptions, logger),
        logError: makeLogError(loaderOptions, logger)
    };
}
exports.makeLogger = makeLogger;
