'use strict';

/**
 * Module dependencies.
 */
var colors = require('colors');

/**
 * Return colored string (or not).
 * @param  {String} str
 * @param  {String} name
 * @param  {Boolean} use
 * @return {String}
 */
module.exports = function (str, name, use) {
	if (!use) {
		return str;
	}
	return String(str)[name];
};