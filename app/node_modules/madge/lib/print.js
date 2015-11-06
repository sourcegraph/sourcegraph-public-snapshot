'use strict';

/**
 * Module dependencies.
 */
var c = require('./color');

/**
 * Return the given object as JSON.
 * @param  {Object} obj
 * @return {String}
 */
function toJSON(obj) {
	return JSON.stringify(obj, null, '  ') + '\n';
}

/**
 * Print module dependency graph as indented text (or JSON).
 * @param  {Object} modules
 * @param  {Object} opts
 */
module.exports.list = function (modules, opts) {
	opts = opts || {};

	if (opts.output === 'json') {
		return process.stdout.write(toJSON(modules));
	}

	Object.keys(modules).forEach(function (id) {
		console.log(c(id, 'cyan', opts.colors));
		modules[id].forEach(function (depId) {
			console.log(c('  ' + depId, 'grey', opts.colors));
		}, this);
	}, this);
};

/**
 * Print a summary of module dependencies.
 * @param  {Object} modules
 * @param  {Object} opts
 */
module.exports.summary = function (modules, opts) {
	var o = {};

	opts = opts || {};

	Object.keys(modules).sort(function (a, b) {
		return modules[b].length - modules[a].length;
	}).forEach(function (id) {
		if (opts.output === 'json') {
			o[id] = modules[id].length;
		} else {
			console.log(c(id + ': ', 'grey', opts.colors) + c(modules[id].length, 'cyan', opts.colors));
		}
	});

	if (opts.output === 'json') {
		return process.stdout.write(toJSON(o));
	}
};

/**
 * Print the result from Madge.circular().
 * @param  {Object} circular
 * @param  {Object} opts
 */
module.exports.circular = function (circular, opts) {
	var arr = circular.getArray();
	if (opts.output === 'json') {
		return process.stdout.write(toJSON(arr));
	}

	if (!arr.length) {
		console.log(c('No circular dependencies found!', 'green', opts.colors));
	} else {
		arr.forEach(function (path, idx) {
			path.forEach(function (module, idx) {
				if (idx) {
					process.stdout.write(c(' -> ', 'cyan', opts.colors));
				}
				process.stdout.write(c(module, 'red', opts.colors));
			});
			process.stdout.write('\n');
		});
		process.exit(1);
	}
};

/**
 * Print the result from Madge.depends().
 * @param  {Object} modules
 * @param  {Object} opts
 */
module.exports.depends = function (modules, opts) {
	if (opts.output === 'json') {
		return process.stdout.write(toJSON(modules));
	}

	modules.forEach(function (id) {
		console.log(c(id, 'grey', opts.colors));
	});
};