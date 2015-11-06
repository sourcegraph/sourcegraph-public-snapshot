'use strict';

/**
 * Module dependencies.
 */
var	util = require('util'),
	requirejs = require('./requirejs'),
	cyclic = require('./cyclic'),
	CJS = require('./parse/cjs'),
	AMD = require('./parse/amd'),
	ES6 = require('./parse/es6'),
	graph = require('./graph');

/**
 * Merge the two given trees.
 * @param  {Object} a
 * @param  {Object} b
 */
function mergeTrees(a, b) {
	Object.keys(b).forEach(function (id) {
		if (!a[id]) {
			a[id] = [];
		}

		b[id].forEach(function (dep) {
			if (a[id].indexOf(dep) < 0) {
				a[id].push(dep);
			}
		});
	});
}

/**
 * Helper for re-mapping path-refs to id-refs that are specified in RequireJS' path config.
 * @param {Object} deps (dependency-list)
 * @param {Object} pathDefs (path-definitions from requirejs-config)
 * @param {String} baseDir (base directory of source files)
 */
function convertPathsToIds(deps, pathDefs, baseDir) {
	var path, pathDeps, i1, len1, i2, len2, isContained;

	if (baseDir){
		baseDir += '/';
	} else {
		baseDir  = '';
	}

	Object.keys(pathDefs).forEach(function (id) {
		path = pathDefs[id];

		//if path does not start with / or a protocol: prepend with baseDir
		if (!/^[^\/]+:\/\/|^\//m.test(path) ){
			path = baseDir + path;
		}

		if (path !== id && deps[path]) {
			if (deps[id] && deps[id].length > 0){
				pathDeps = deps[path].slice(0, deps[path].length-1);

				//remove entries from <path-ref>, if already contained in <id-ref>
				for (i1=0, len1 = pathDeps.length; i1 < len1; ++i1){
					for (i2=0, len2 = deps[id].length; i2 < len2; ++i2){
						if (pathDeps[i1] === deps[id][i2]){
							pathDeps.splice(i1--, 1);
							break;
						}
					}
				}
				deps[id] = deps[id].concat(pathDeps);
			} else {
				deps[id] = deps[path];
			}

			delete deps[path];
		} else if (!deps[id]) {
			deps[id] = [];
		}

		//normalize entries within deps-arrays (i.e. replace path-refs with id-refs)
		Object.keys(pathDefs).forEach(function (id) {
			path = baseDir + pathDefs[id];
			if (deps[id]){
				for (i1=0, len1 = deps[id].length; i1 < len1; ++i1){
					//replace path-ref with id-ref (if necessary)
					if (deps[id][i1] === path){
						deps[id][i1] = id;
					}
				}
			}
		});
	});
}

/**
 * Expose factory function.
 * @api public
 * @param {String|Array|Object} src
 * @param {Object} opts
 * @return {Madge}
 */
module.exports = function (src, opts) {
	return new Madge(src, opts);
};

/**
 * Class constructor.
 * @constructor
 * @api public
 * @param {String|Array|Object} src
 * @param {Object} opts
 */
function Madge(src, opts) {
	var tree = [];

	this.opts = opts || {};
	this.opts.format = String(this.opts.format || 'cjs').toLowerCase();

	if (typeof(src) === 'object' && !Array.isArray(src)) {
		this.tree = src;
		return;
	}

	if (typeof(src) === 'string') {
		src = [src];
	}

	if (src && src.length) {
		tree = this.parse(src);
	}

	if (this.opts.requireConfig) {
		var baseDir = src.length ? src[0].replace(/\\/g, '/') : '';
		baseDir = requirejs.getBaseUrlFromConfig(this.opts.requireConfig, baseDir);
		convertPathsToIds(tree, requirejs.getPathsFromConfig(this.opts.requireConfig, this.opts.exclude), baseDir);
		mergeTrees(tree, requirejs.getShimDepsFromConfig(this.opts.requireConfig, this.opts.exclude));
	}

	this.tree = tree;
}

/**
 * Parse the given source folder(s).
 * @param  {Array|Object} src
 * @return {Object}
 */
Madge.prototype.parse = function(src) {
	if (this.opts.format === 'cjs') {
		return new CJS(src, this.opts, this).tree;
	} else if (this.opts.format === 'amd') {
		return new AMD(src, this.opts, this).tree;
	} else if (this.opts.format === 'es6') {
		return new ES6(src, this.opts, this).tree;
	} else {
		throw new Error('invalid module format "' + this.opts.format + '"');
	}
};

/**
 * Return the module dependency graph as an object.
 * @api public
 * @return {Object}
 */
Madge.prototype.obj = function () {
	return this.tree;
};

/**
 * Return the modules that has circular dependencies.
 * @api public
 * @return {Object}
 */
Madge.prototype.circular = function () {
	return cyclic(this.tree);
};

/**
 * Return a list of modules that depends on the given module.
 * @api public
 * @param  {String} id
 * @return {Array|Object}
 */
Madge.prototype.depends = function (id) {
	return Object.keys(this.tree).filter(function (module) {
		if (this.tree[module]) {
			return this.tree[module].reduce(function (acc, dependency) {
				if (dependency === id) {
					acc = module;
				}
				return acc;
			}, false);
		}
	}, this);
};

/**
 * Return the module dependency graph as DOT output.
 * @api public
 * @return {String}
 */
Madge.prototype.dot = function () {
	return graph.dot(this.tree);
};

/**
 * Return the module dependency graph as a PNG image.
 * @api public
 * @param  {Object}   opts
 * @param  {Function} callback
 */
Madge.prototype.image = function (opts, callback) {
	graph.image(this.tree, opts, callback);
};
