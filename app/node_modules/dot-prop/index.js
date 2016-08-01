'use strict';
var isObj = require('is-obj');

module.exports.get = function (obj, path) {
	if (!isObj(obj) || typeof path !== 'string') {
		return obj;
	}

	var pathArr = getPathSegments(path);

	for (var i = 0; i < pathArr.length; i++) {
		obj = obj[pathArr[i]];

		if (obj === undefined) {
			break;
		}
	}

	return obj;
};

module.exports.set = function (obj, path, value) {
	if (!isObj(obj) || typeof path !== 'string') {
		return;
	}

	var pathArr = getPathSegments(path);

	for (var i = 0; i < pathArr.length; i++) {
		var p = pathArr[i];

		if (!isObj(obj[p])) {
			obj[p] = {};
		}

		if (i === pathArr.length - 1) {
			obj[p] = value;
		}

		obj = obj[p];
	}
};

module.exports.delete = function (obj, path) {
	if (!isObj(obj) || typeof path !== 'string') {
		return;
	}

	var pathArr = getPathSegments(path);

	for (var i = 0; i < pathArr.length; i++) {
		var p = pathArr[i];

		if (i === pathArr.length - 1) {
			delete obj[p];
			return;
		}

		obj = obj[p];
	}
};

module.exports.has = function (obj, path) {
	if (!isObj(obj) || typeof path !== 'string') {
		return false;
	}

	var pathArr = getPathSegments(path);

	for (var i = 0; i < pathArr.length; i++) {
		obj = obj[pathArr[i]];

		if (obj === undefined) {
			return false;
		}
	}

	return true;
};

function getPathSegments(path) {
	var pathArr = path.split('.');
	var parts = [];

	for (var i = 0; i < pathArr.length; i++) {
		var p = pathArr[i];

		while (p[p.length - 1] === '\\') {
			p = p.slice(0, -1) + '.';
			p += pathArr[++i];
		}

		parts.push(p);
	}

	return parts;
}
