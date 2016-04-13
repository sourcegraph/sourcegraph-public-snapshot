'use strict';

var File = require('vinyl');
var isTar = require('is-tar');
var stripDirs = require('strip-dirs');
var tar = require('tar-stream');
var through = require('through2');

/**
 * tar decompress plugin
 *
 * @param {Object} opts
 * @api public
 */

module.exports = function (opts) {
    opts = opts || {};
    opts.strip = +opts.strip || 0;

    return through.obj(function (file, enc, cb) {
        var extract = tar.extract();
        var self = this;

        if (file.isNull()) {
            cb(null, file);
            return;
        }

        if (file.isStream()) {
            cb(new Error('Streaming is not supported'));
            return;
        }

        if (!isTar(file.contents)) {
            cb(null, file);
            return;
        }

        extract.on('error', function (err) {
            cb(err);
            return;
        });

        extract.on('entry', function (header, stream, done) {
            var chunk = [];
            var len = 0;

            stream.on('data', function (data) {
                chunk.push(data);
                len += data.length;
            });

            stream.on('end', function () {
                if (header.type !== 'directory') {
                    self.push(new File({
                        contents: Buffer.concat(chunk, len),
                        path: stripDirs(header.name, opts.strip)
                    }));
                }

                done();
            });
        });

        extract.on('finish', function () {
            cb();
        });

        extract.end(file.contents);
    });
};
