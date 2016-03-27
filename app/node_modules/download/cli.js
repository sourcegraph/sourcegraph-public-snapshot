#!/usr/bin/env node
'use strict';

var Download = require('./');
var meow = require('meow');
var progress = require('download-status');
var stdin = require('get-stdin');

/**
 * Options
 */

var cli = meow({
	help: [
		'Usage',
		'  download <url>',
		'  download <url> > <file>',
		'  download --out <directory> <url>',
		'  cat <file> | download --out <directory>',
		'',
		'Example',
		'  download http://foo.com/file.zip',
		'  download http://foo.com/cat.png > dog.png',
		'  download --extract --strip 1 --out dest http://foo.com/file.zip',
		'  cat urls.txt | download --out dest',
		'',
		'Options',
		'  -e, --extract           Try decompressing the file',
		'  -o, --out               Where to place the downloaded files',
		'  -s, --strip <number>    Strip leading paths from file names on extraction'
	].join('\n')
}, {
	boolean: [
		'extract'
	],
	string: [
		'out',
		'strip'
	],
	alias: {
		e: 'extract',
		o: 'out',
		s: 'strip'
	},
	default: {
		out: process.cwd()
	}
});

/**
 * Run
 *
 * @param {Array} src
 * @param {String} dest
 * @api private
 */

function run(src, dest) {
	var download = new Download(cli.flags);

	src.forEach(download.get.bind(download));

	if (process.stdout.isTTY) {
		download.use(progress());
		download.dest(dest);
	}

	download.run(function (err, files) {
		if (err) {
			console.error(err.message);
			process.exit(1);
		}

		if (!process.stdout.isTTY) {
			files.forEach(function (file) {
				process.stdout.write(file.contents);
			});
		}
	});
}

/**
 * Apply arguments
 */

if (process.stdin.isTTY) {
	var src = cli.input;
	var dest = cli.flags.out;

	if (!src.length){
		console.error([
			'Specify a URL to fetch',
			'',
			'Example',
			'  download http://foo.com/file.zip',
			'  download http://foo.com/cat.png > dog.png',
			'  download --extract --strip 1 --out dest http://foo.com/file.zip',
			'  cat urls.txt | download --out dest'
		].join('\n'));

		process.exit(1);
	}

	run(src, dest);
} else {
	stdin(function (data) {
		var src = cli.input;
		var dest = cli.flags.out;

		[].push.apply(src, data.trim().split('\n'));
		run(src, dest);
	});
}
