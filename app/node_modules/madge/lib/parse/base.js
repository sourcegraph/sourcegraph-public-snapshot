'use strict';

/**
 * Module dependencies
 */
var fs = require('fs'),
	path = require('path'),
	util = require('util'),
	resolve = require('resolve'),
	EventEmitter = require('events').EventEmitter,
	commondir = require('commondir'),
	finder = require('walkdir'),
	coffee = require('coffee-script'),
	jsx = require('react-tools');

/**
 * Traversing `src` and fetches all dependencies.
 * @constructor
 * @param {Array} src
 * @param {Object} opts
 * @param {Object} parent
 */
var Base = module.exports = function(src, opts, parent) {
	if (opts.onParseFile) {
		this.on('parseFile', opts.onParseFile.bind(parent));
	}

	if (opts.onAddModule) {
		this.on('addModule', opts.onAddModule.bind(parent));
	}

	this.opts = opts;

	if (typeof this.opts.extensions === "undefined") {
		this.opts.extensions = ['.js'];
	}

	this.tree = {};
	this.extRegEx = new RegExp('\\.(coffee|jsx|' + this.opts.extensions.map(function(str) {
		return str.substring(1);
	}).join('|') + ')$', 'g');
	this.coffeeExtRegEx = /\.coffee$/;
	this.jsxExtRegEx = /\.jsx$/;
	src = this.resolveTargets(src);
	this.excludeRegex = opts.exclude ? new RegExp(opts.exclude) : false;
	this.baseDir = this.getBaseDir(src);
	this.readFiles(src);
	this.sortDependencies();
};

util.inherits(Base, EventEmitter);

/**
 * Resolve the given `id` to a filename.
 * @param  {String} dir
 * @param  {String} id
 * @return {String}
 */
Base.prototype.resolve = function (dir, id) {
	try {
		return resolve.sync(id, {
			basedir: dir,
			paths: this.opts.paths,
			extensions: this.opts.extensions
		});
	} catch (e) {
		if (this.opts.breakOnError) {
			console.log(String('\nError while resolving module from: ' + id).red);
			throw e;
		}
		return id;
	}
};

/**
 * Get the most common dir from the `src`.
 * @param  {Array} src
 * @return {String}
 */
Base.prototype.getBaseDir = function (src) {
	var dir = commondir(src);

	if (!fs.statSync(dir).isDirectory()) {
		dir = path.dirname(dir);
	}
	return dir;
};

/**
 * Resolves all paths in `sources` and ensure we have a absolute path.
 * @param  {Array} sources
 * @return {Array}
 */
Base.prototype.resolveTargets = function (sources) {
	return sources.map(function (src) {
		return path.resolve(src);
	});
};

/**
 * Normalize a module file path and return a proper identificator.
 * @param  {String} filename
 * @return {String}
 */
Base.prototype.normalize = function (filename) {
	return this.replaceBackslashInPath(path.relative(this.baseDir, filename).replace(this.extRegEx, ''));
};

/**
 * Check if module should be excluded.
 * @param  {String}
 * @return {Boolean}
 */
Base.prototype.isExcluded = function (id) {
	return this.excludeRegex && id.match(this.excludeRegex);
};

/**
 * Parse the given `filename` and add it to the module tree.
 * @param {String} filename
 */
Base.prototype.addModule = function (filename) {
	var id = this.normalize(filename);

	if (!this.isExcluded(id) && fs.existsSync(filename)) {
		this.tree[id] = this.parseFile(filename);
		this.emit("addModule", {id: id, dependencies: this.tree[id]});
	}
};

/**
 * Traverse `sources` and parse files found.
 * @param  {Array} sources
 */
Base.prototype.readFiles = function (sources) {
	sources.forEach(function (src) {
		if (fs.statSync(src).isDirectory()) {
			finder.sync(src).filter(function (filename) {
				return filename.match(this.extRegEx);
			}, this).forEach(function (filename) {
				this.addModule(filename);
			}, this);
		} else {
			this.addModule(src);
		}
	}, this);
};

/**
 * Read the given filename and compile it if necessary and return the content.
 * @param  {String} filename
 * @return {String}
 */
Base.prototype.getFileSource = function (filename) {
	var src = fs.readFileSync(filename, 'utf8');

	if (filename.match(this.coffeeExtRegEx)) {
		src = coffee.compile(src, {
			header: false,
			bare: true
		});
	} else if (filename.match(this.jsxExtRegEx)) {
		src = jsx.transform(src);
	}

	return src;
};

/**
 * Sort dependencies by name.
 */
Base.prototype.sortDependencies = function () {
	var self = this;

	this.tree = Object.keys(this.tree).sort().reduce(function (acc, id) {
		(acc[id] = self.tree[id]).sort();
		return acc;
	}, {});
};

/**
 * Replace back slashes in path (Windows) with forward slashes (*nix).
 * @param  {String} path
 * @return {String}
 */
Base.prototype.replaceBackslashInPath = function (path) {
	return path.replace(/\\/g, '/');
};