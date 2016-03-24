'use strict';

exports.__esModule = true;
exports.default = warnOnce;
/* istanbul ignore next */

var printed = {};

function warnOnce(message) {
    if (printed[message]) return;
    printed[message] = true;

    if (typeof console !== 'undefined' && console.warn) console.warn(message);
}
module.exports = exports['default'];