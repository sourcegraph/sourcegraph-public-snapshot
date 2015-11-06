'use strict';

/**
 * Get path to the circular dependency.
 * @param  {String} parent
 * @param  {Object} unresolved
 * @return {Array}
 */
function getPath(parent, unresolved) {
	var parentVisited = false;

	return Object.keys(unresolved).filter(function (module) {
		if (module === parent) {
			parentVisited = true;
		}
		return parentVisited && unresolved[module];
	});
}

/**
 * A circular dependency is occurring when we see a software package
 * more than once, unless that software package has all its dependencies resolved.
 * @param  {String} id
 * @param  {Object} modules
 * @param  {Object} circular
 * @param  {Object} resolved
 * @param  {Object} unresolved
 */
function resolver(id, modules, circular, resolved, unresolved) {
	unresolved[id] = true;

	if (modules[id]) {
		modules[id].forEach(function (dependency) {
			if (!resolved[dependency]) {
				if (unresolved[dependency]) {
					circular.push(getPath(dependency, unresolved));
					return;
				}
				resolver(dependency, modules, circular, resolved, unresolved);
			}
		});
	}

	resolved[id] = true;
	unresolved[id] = false;
}

/**
 * Finds all circular dependencies for the given modules.
 * @param  {Object} modules
 * @return {Object}
 */
module.exports = function (modules) {
	var circular = [],
		resolved = {},
		unresolved = {};

	Object.keys(modules).forEach(function (id) {
		resolver(id, modules, circular, resolved, unresolved);
	});

	return {
		/**
		 * Expose the circular dependency array.
		 * @return {Array}
		 */
		getArray: function () {
			return circular;
		},

		/**
		 * Check if the given module is part of a circular dependency.
		 * @return {Boolean}
		 */
		isCyclic: function (id) {
			var cyclic = false;
			circular.forEach(function (path) {
				if (path.indexOf(id) >= 0) {
					cyclic = true;
				}
			});
			return cyclic;
		}
	};
};