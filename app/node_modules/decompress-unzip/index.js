'use strict';

var File = require('vinyl');
var fs = require('fs');
var isZip = require('is-zip');
var stripDirs = require('strip-dirs');
var through = require('through2');
var yauzl = require('yauzl');

module.exports = function (opts) {
	opts = opts || {};
	opts.strip = +opts.strip || 0;

	return through.obj(function (file, enc, cb) {
		var self = this;

		if (file.isNull()) {
			cb(null, file);
			return;
		}

		if (file.isStream()) {
			cb(new Error('Streaming is not supported'));
			return;
		}

		if (!isZip(file.contents)) {
			cb(null, file);
			return;
		}

		yauzl.fromBuffer(file.contents, function (err, zipFile) {
			var count = 0;

			if (err) {
				cb(err);
				return;
			}

			zipFile
				.on('error', cb)
				.on('entry', function (entry) {
					if (entry.fileName.charAt(entry.fileName.length - 1) === '/') {
						if (++count === zipFile.entryCount) {
							cb();
						}

						return;
					}

					zipFile.openReadStream(entry, function (err, readStream) {
						if (err) {
							cb(err);
							return;
						}

						var chunks = [];
						var len = 0;

						readStream
							.on('error', cb)
							.on('data', function (data) {
								chunks.push(data);
								len += data.length;
							})
							.on('end', function () {
								var stat;
								var mode = (entry.externalFileAttributes >> 16) & 0xFFFF;

								if (mode) {
									stat = new fs.Stats();
									stat.mode = mode;
								} else {
									stat = null;
								}

								self.push(new File({
									stat: stat,
									contents: Buffer.concat(chunks, len),
									path: stripDirs(entry.fileName, opts.strip)
								}));

								if (++count === zipFile.entryCount) {
									cb();
								}
							});
					});
			});
		});
	});
};
