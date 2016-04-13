'use strict';

var delimiter = require('path').delimiter;
var isPathInside = require('is-path-inside');

module.exports = function (str) {
	if (typeof str !== 'string') {
		throw new TypeError('Expected a string');
	}

	return process.env.PATH.split(delimiter).some(function (path) {
		return isPathInside(str, path) || str === path;
	});
};
