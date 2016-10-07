/*---------------------------------------------------------------------------------------------
 *  Copyright (c) Microsoft Corporation. All rights reserved.
 *  Licensed under the MIT License. See License.txt in the project root for license information.
 *--------------------------------------------------------------------------------------------*/
'use strict';
// --- THIS FILE IS TEMPORARY UNTIL ENV.TS IS CLEANED UP. IT CAN SAFELY BE USED IN ALL TARGET EXECUTION ENVIRONMENTS (node & dom) ---
var _isWindows = false;
var _isMacintosh = false;
var _isLinux = false;
var _isRootUser = false;
var _isNative = false;
var _isWeb = false;
var _isQunit = false;
var _locale = undefined;
var _language = undefined;
exports.LANGUAGE_DEFAULT = 'en';
// OS detection
if (typeof process === 'object') {
    _isWindows = (process.platform === 'win32');
    _isMacintosh = (process.platform === 'darwin');
    _isLinux = (process.platform === 'linux');
    _isRootUser = !_isWindows && (process.getuid() === 0);
    var rawNlsConfig = process.env['VSCODE_NLS_CONFIG'];
    if (rawNlsConfig) {
        try {
            var nlsConfig = JSON.parse(rawNlsConfig);
            var resolved = nlsConfig.availableLanguages['*'];
            _locale = nlsConfig.locale;
            // VSCode's default language is 'en'
            _language = resolved ? resolved : exports.LANGUAGE_DEFAULT;
        }
        catch (e) {
        }
    }
    _isNative = true;
}
else if (typeof navigator === 'object') {
    var userAgent = navigator.userAgent;
    _isWindows = userAgent.indexOf('Windows') >= 0;
    _isMacintosh = userAgent.indexOf('Macintosh') >= 0;
    _isLinux = userAgent.indexOf('Linux') >= 0;
    _isWeb = true;
    _locale = navigator.language;
    _language = _locale;
    _isQunit = !!self.QUnit;
}
(function (Platform) {
    Platform[Platform["Web"] = 0] = "Web";
    Platform[Platform["Mac"] = 1] = "Mac";
    Platform[Platform["Linux"] = 2] = "Linux";
    Platform[Platform["Windows"] = 3] = "Windows";
})(exports.Platform || (exports.Platform = {}));
var Platform = exports.Platform;
exports._platform = Platform.Web;
if (_isNative) {
    if (_isMacintosh) {
        exports._platform = Platform.Mac;
    }
    else if (_isWindows) {
        exports._platform = Platform.Windows;
    }
    else if (_isLinux) {
        exports._platform = Platform.Linux;
    }
}
exports.isWindows = _isWindows;
exports.isMacintosh = _isMacintosh;
exports.isLinux = _isLinux;
exports.isRootUser = _isRootUser;
exports.isNative = _isNative;
exports.isWeb = _isWeb;
exports.isQunit = _isQunit;
exports.platform = exports._platform;
/**
 * The language used for the user interface. The format of
 * the string is all lower case (e.g. zh-tw for Traditional
 * Chinese)
 */
exports.language = _language;
/**
 * The OS locale or the locale specified by --locale. The format of
 * the string is all lower case (e.g. zh-tw for Traditional
 * Chinese). The UI is not necessarily shown in the provided locale.
 */
exports.locale = _locale;
var _globals = (typeof self === 'object' ? self : global);
exports.globals = _globals;
function hasWebWorkerSupport() {
    return typeof _globals.Worker !== 'undefined';
}
exports.hasWebWorkerSupport = hasWebWorkerSupport;
exports.setTimeout = _globals.setTimeout.bind(_globals);
exports.clearTimeout = _globals.clearTimeout.bind(_globals);
exports.setInterval = _globals.setInterval.bind(_globals);
exports.clearInterval = _globals.clearInterval.bind(_globals);
