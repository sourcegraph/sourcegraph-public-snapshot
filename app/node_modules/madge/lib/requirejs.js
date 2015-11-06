'use strict';

/**
 * Module dependencies.
 */
var fs = require('fs'),
	parse = require('./parse/parse'),
	path = require('path');

/**
 * Read shim dependencies from RequireJS config.
 * @param  {String} filename
 * @param  {String} [exclude]
 * @return {Object}
 */
module.exports.getShimDepsFromConfig = function (filename, exclude) {
	var deps = {},
		config = parse.findConfig(filename, fs.readFileSync(filename, 'utf8')),
		excludeRegex = exclude ? new RegExp(exclude) : false,
		isIncluded = function (key) {
			return !(excludeRegex && key.match(excludeRegex));
		};

	if (config.shim) {
		Object.keys(config.shim).filter(isIncluded).forEach(function (key) {
			if (config.shim[key].deps) {
				deps[key] = config.shim[key].deps.filter(isIncluded);
			} else {
				deps[key] = [];
			}
		});
	}

	return deps;
};

/**
* Read path definitions from RequireJS config.
* @param  {String} filename
* @param  {String} [exclude]
* @return {Object}
*/
module.exports.getPathsFromConfig = function (filename, exclude) {
	var paths = {},
		config = parse.findConfig(filename, fs.readFileSync(filename, 'utf8')),
		excludeRegex = exclude ? new RegExp(exclude) : false,
		isIncluded = function (key) {
			return !(excludeRegex && key.match(excludeRegex));
		};

	if (config.paths) {
		Object.keys(config.paths).filter(isIncluded).forEach(function (key) {
			paths[key] = config.paths[key];
		});
	}

	return paths;
};

/**
 * Read baseUrl from RequireJS config.
 * @param  {String} filename
 * @param  {String} srcBaseDir
 */
module.exports.getBaseUrlFromConfig = function (filename, srcBaseDir) {
	var config = parse.findConfig(filename, fs.readFileSync(filename, 'utf8'));
	return config.baseUrl ? path.relative(srcBaseDir, config.baseUrl) : '';
};