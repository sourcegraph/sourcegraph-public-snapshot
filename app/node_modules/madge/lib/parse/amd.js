'use strict';

/**
 * Module dependencies.
 */
var fs = require('fs'),
	path = require('path'),
	util = require('util'),
	amdetective = require('amdetective'),
	colors = require('colors'),
	Base = require('./base');

/**
 * This class will parse the AMD module format.
 * @see https://github.com/amdjs/amdjs-api/wiki/AMD
 * @constructor
 */
var AMD = module.exports = function () {
	Base.apply(this, arguments);
};

/**
 * Inherit from `Base`
 */
util.inherits(AMD, Base);

/**
 * Normalize a module file path and return a proper identificator.
 * @param  {String} filename
 * @return {String}
 */
AMD.prototype.normalize = function (filename) {
	return this.replaceBackslashInPath(path.relative(this.baseDir, filename).replace(this.extRegEx, ''));
};

/**
 * Parse the given file and return all found dependencies.
 * @param  {String} filename
 * @return {Array}
 */
AMD.prototype.parseFile = function (filename) {
	try {
		var dependencies = [],
			src = this.getFileSource(filename),
			fileData = {filename: filename, src: src};

     this.emit('parseFile', fileData);

		if (/define|require\s*\(/m.test(fileData.src)) {
			amdetective(fileData.src, {findNestedDependencies: this.opts.findNestedDependencies}).map(function (obj) {
				return typeof(obj) === 'string' ? [obj] : obj.deps;
			}).filter(function (deps) {
				deps.filter(function (id) {
					// Ignore RequireJS IDs and plugins
					return id !== 'require' && id !== 'exports' && id !== 'module' && !id.match(/\.?\w\!/);
				}).map(function (id) {
					// Only resolve relative module identifiers (if the first term is "." or "..")
					if (id.charAt(0) !== '.') {
						return id;
					}

					var depFilename = path.resolve(path.dirname(fileData.filename), id);

					if (depFilename) {
						return this.normalize(depFilename);
					}
				}, this).forEach(function (id) {
					if (!this.isExcluded(id) && dependencies.indexOf(id) < 0) {
						dependencies.push(id);
					}
				}, this);
			}, this);

			return dependencies;
		}
	} catch (e) {
		if (this.opts.breakOnError) {
			console.log(String('\nError while parsing file: ' + filename).red);
			throw e;
		}
	}

	return [];
};

/**
 * Get module dependencies from optimize file (r.js).
 */
AMD.prototype.addOptimizedModules = function (filename) {
	var self = this,
		anonymousRequire = [];

	amdetective(this.getFileSource(filename))
		.filter(function(obj) {
			var id = obj.name || obj;
			return id !== 'require' && id !== 'exports' && id !== 'module' && !id.match(/\.?\w\!/) && !self.isExcluded(id);
		})
		.forEach(function (obj) {
			if (typeof(obj) === 'string') {
				anonymousRequire.push(obj);
				return;
			}

			if (!self.isExcluded(obj.name)) {
				self.tree[obj.name] = obj.deps.filter(function(id) {
					return id !== 'require' && id !== 'exports' && id !== 'module' && !id.match(/\.?\w\!/) && !self.isExcluded(id);
				});
			}
	});

	if (anonymousRequire.length > 0) {
		this.tree[this.opts.mainRequireModule || ''] = anonymousRequire;
	}
};

/**
 * Parse the given `filename` and add it to the module tree.
 * @param {String} filename
 */
AMD.prototype.addModule = function (filename) {
	if (this.opts.optimized) {
		return this.addOptimizedModules(filename);
	} else {
		return Base.prototype.addModule.call(this, filename);
	}
};