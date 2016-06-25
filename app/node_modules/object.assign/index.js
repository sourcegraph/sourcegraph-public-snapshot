'use strict';

var defineProperties = require('define-properties');

var implementation = require('./implementation');
var getPolyfill = require('./polyfill');
var shim = require('./shim');

defineProperties(implementation, {
	implementation: implementation,
	getPolyfill: getPolyfill,
	shim: shim
});

module.exports = implementation;
