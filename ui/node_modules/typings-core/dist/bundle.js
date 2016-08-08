"use strict";
var Promise = require('any-promise');
var path_1 = require('path');
var events_1 = require('events');
var dependencies_1 = require('./lib/dependencies');
var compile_1 = require('./lib/compile');
var fs_1 = require('./utils/fs');
function bundle(options) {
    var cwd = options.cwd, global = options.global, out = options.out;
    var emitter = options.emitter || new events_1.EventEmitter();
    var resolution = options.resolution || 'main';
    if (out == null) {
        return Promise.reject(new TypeError('Out file path is required for bundle'));
    }
    return dependencies_1.resolveAllDependencies({ cwd: cwd, dev: false, global: false, emitter: emitter })
        .then(function (tree) {
        var name = options.name || tree.name;
        if (name == null) {
            return Promise.reject(new TypeError('Unable to infer typings name from project. Use the `--name` flag to specify it manually'));
        }
        return compile_1.compile(tree, [resolution], { cwd: cwd, name: name, global: global, emitter: emitter, meta: true });
    })
        .then(function (output) {
        var path = path_1.resolve(cwd, out);
        return fs_1.mkdirp(path_1.dirname(path))
            .then(function () {
            return fs_1.writeFile(path, output.results[resolution]);
        })
            .then(function () { return ({ tree: output.tree }); });
    });
}
exports.bundle = bundle;
//# sourceMappingURL=bundle.js.map