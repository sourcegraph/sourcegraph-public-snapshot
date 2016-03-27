'use strict';
var objectAssign = require('object-assign');
var chalk = require('chalk');
var lpadAlign = require('lpad-align');
var Progress = require('progress');

module.exports = function (opts) {
	opts = opts || {};
	opts.stream = opts.stream || process.stderr;
	opts.indent = opts.indent || 2;

	var words = [
		'fetch',
		'progress'
	];

	var progress = chalk.cyan(lpadAlign('progress', words, opts.indent));
	var str = progress + ' : [:bar] :percent :etas';

	var bar = new Progress(str, objectAssign({
		complete: '=',
		incomplete: ' ',
		width: 20,
		total: 0
	}, opts));

	var streams = 0;

	return function (res, url, cb) {
		if (!res.headers['content-length'] || !opts.stream.isTTY) {
			cb();
			return;
		}

		streams ++;
		bar.total += parseInt(res.headers['content-length'], 10);

		var fetch = chalk.cyan(lpadAlign('fetch', words, opts.indent));

		opts.stream.clearLine();
		opts.stream.cursorTo(0);
		opts.stream.write(fetch + ' : ' + url + '\n');

		res.on('data', function (data) {
			bar.tick(data.length);
		});

		res.on('end', function () {
			if (--streams === 0) {
				opts.stream.write('\n');
			}

			cb();
		});
	};
};
