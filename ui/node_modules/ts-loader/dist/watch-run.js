"use strict";
var path = require('path');
var utils = require('./utils');
/**
 * Make function which will manually update changed files
 */
function makeWatchRun(instance) {
    return function (watching, cb) {
        var mtimes = watching.compiler.watchFileSystem.watcher.mtimes;
        if (null === instance.modifiedFiles) {
            instance.modifiedFiles = {};
        }
        Object.keys(mtimes)
            .filter(function (filePath) { return !!filePath.match(/\.tsx?$|\.jsx?$/); })
            .forEach(function (filePath) {
            filePath = path.normalize(filePath);
            var file = instance.files[filePath];
            if (file) {
                file.text = utils.readFile(filePath) || '';
                file.version++;
                instance.version++;
                instance.modifiedFiles[filePath] = file;
            }
        });
        cb();
    };
}
module.exports = makeWatchRun;
