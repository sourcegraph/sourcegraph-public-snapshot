'use strict';

var combine = require('stream-combiner');
var concat = require('concat-stream');
var conf = require('rc')('npm');
var each = require('each-async');
var File = require('vinyl');
var fs = require('vinyl-fs');
var path = require('path');
var rename = require('gulp-rename');
var tar = require('decompress-tar');
var tarBz2 = require('decompress-tarbz2');
var tarGz = require('decompress-targz');
var through = require('through2');
var urlRegex = require('url-regex');
var Ware = require('ware');
var zip = require('decompress-unzip');

/**
 * Initialize a new `Download`
 *
 * @param {Object} opts
 * @api public
 */

function Download(opts) {
	if (!(this instanceof Download)) {
		return new Download(opts);
	}

	this.opts = opts || {};
	this.opts.strictSSL = conf['strict-ssl'];
	this.opts.proxy = conf['https-proxy'] ||
					  conf['http-proxy'] ||
					  conf.proxy ||
					  process.env.HTTPS_PROXY ||
					  process.env.https_proxy ||
					  process.env.HTTP_PROXY ||
					  process.env.http_proxy ||
					  this.opts.proxy;

	this.tasks = [];
	this.ware = new Ware();
	this._get = [];
}

/**
 * Get or set URL to download
 *
 * @param {String} url
 * @api public
 */

Download.prototype.get = function (url) {
	if (!arguments.length) {
		return this._get;
	}

	this._get.push(url);
	return this;
};

/**
 * Get or set the destination folder
 *
 * @param {String} dir
 * @api public
 */

Download.prototype.dest = function (dir) {
	if (!arguments.length) {
		return this._dest;
	}

	this._dest = dir;
	return this;
};

/**
 * Rename the downloaded file
 *
 * @param {Function|String} name
 * @api public
 */

Download.prototype.rename = function (name) {
	if (!arguments.length) {
		return this._name;
	}

	this._name = name;
	return this;
};

/**
 * Add a plugin to the middleware stack
 *
 * @param {Function} plugin
 * @api public
 */

Download.prototype.use = function (plugin) {
	this.ware.use(plugin);
	return this;
};

/**
 * Add a task to the middleware stack
 *
 * @param {Function} task
 * @api public
 */

Download.prototype.pipe = function (task) {
	this.tasks.push(task);
	return this;
};

/**
 * Run
 *
 * @param {Function} cb
 * @api public
 */

Download.prototype.run = function (cb) {
	cb = cb || function () {};

	var request = require('request');
	var self = this;
	var files = [];

	each(this.get(), function (url, i, done) {
		if (!urlRegex().test(url)) {
			done(new Error('Specify a valid URL'));
			return;
		}

		request.get(url, self.opts)
			.on('response', function (res) {
				self.res(url, res, function (err, ret) {
					if (err) {
						done(err);
						return;
					}

					files.push(ret);
					done();
				});
			})
			.on('error', done);
	}, function (err) {
		if (err) {
			cb(err);
			return;
		}

		var pipe = self.construct(files);
		var end = concat(function (files) {
			cb(null, files, pipe);
		});

		pipe.on('error', cb);
		pipe.pipe(end);
	});
};

/**
 * Handle response
 *
 * @param {String} url
 * @param {Object} res
 * @param {Function} cb
 * @api private
 */

Download.prototype.res = function (url, res, cb) {
	var ret = [];
	var len = 0;

	if (res.statusCode < 200 || res.statusCode >= 300) {
		var err = new Error([
			'Couldn\'t connect to ' + url,
			'(' + res.statusCode + ')'
		].join(' '));

		err.code = res.statusCode;
		res.destroy();
		cb(err);
		return;
	}

	res.on('error', cb);
	res.on('data', function (data) {
		ret.push(data);
		len += data.length;
	});

	this.ware.run(res, url);

	res.on('end', function () {
		cb(null, {
			path: path.basename(url),
			contents: Buffer.concat(ret, len),
			url: url
		});
	});
};

/**
 * Construct stream
 *
 * @param {Array} files
 * @api private
 */

Download.prototype.construct = function (files) {
	var stream = through.obj();

	files.forEach(function (file) {
		var obj = new File(file);
		obj.url = file.url;
		stream.write(obj);
	});

	stream.end();

	if (this.opts.extract) {
		this.tasks.unshift(tar(this.opts), tarBz2(this.opts), tarGz(this.opts), zip(this.opts));
	}

	this.tasks.unshift(stream);

	if (this.rename()) {
		this.tasks.push(rename(this.rename()));
	}

	if (this.dest()) {
		this.tasks.push(fs.dest(this.dest(), this.opts));
	}

	return combine(this.tasks);
};

/**
 * Module exports
 */

module.exports = Download;
