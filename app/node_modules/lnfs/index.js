'use strict';
var fs = require('fs');
var path = require('path');
var mkdirp = require('mkdirp');
var rimraf = require('rimraf');

function link(src, dest, type, cb) {
	rimraf(dest, function (err) {
		if (err) {
			cb(err);
			return;
		}

		mkdirp(path.dirname(dest), function (err) {
			if (err) {
				cb(err);
				return;
			}

			fs.symlink(src, dest, type, cb);
		});
	});
}

module.exports = function (src, dest, type, cb) {
	if (typeof src !== 'string' || typeof dest !== 'string') {
		throw new Error('Source file and target required');
	}

	if (typeof type === 'function' && !cb) {
		cb = type;
		type = null;
	}

	src = path.resolve(src);
	dest = path.resolve(dest);

	fs.lstat(dest, function (err, stats) {
		if (err && err.code === 'ENOENT') {
			return link(src, dest, type, cb);
		}

		if (err) {
			cb(err);
			return;
		}

		if (!stats.isSymbolicLink()) {
			return link(src, dest, type, cb);
		}

		fs.realpath(dest, function (err, res) {
			if (err) {
				cb(err);
				return;
			}

			if (res === src) {
				cb();
				return;
			}

			link(src, dest, type, cb);
		});
	});
};

module.exports.sync = function (src, dest, type) {
	if (typeof src !== 'string' || typeof dest !== 'string') {
		throw new Error('Source file and target required');
	}

	src = path.resolve(src);
	dest = path.resolve(dest);

	try {
		var stats = fs.lstatSync(dest);
		var realpath = fs.realpathSync(dest);

		if (!stats.isSymbolicLink()) {
			rimraf.sync(dest);
			fs.symlinkSync(src, dest, type);
			return;
		}

		if (realpath === src) {
			return;
		}

		rimraf.sync(dest);
		fs.symlinkSync(src, dest, type);
	} catch (err) {
		if (err.code === 'ENOENT') {
			mkdirp.sync(path.dirname(dest));
			fs.symlinkSync(src, dest, type);
			return;
		}

		throw err;
	}
};
