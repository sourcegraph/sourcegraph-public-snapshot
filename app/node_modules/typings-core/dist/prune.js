"use strict";
var path_1 = require('path');
var Promise = require('any-promise');
var extend = require('xtend');
var events_1 = require('events');
var find_1 = require('./utils/find');
var fs_1 = require('./utils/fs');
var path_2 = require('./utils/path');
function prune(options) {
    var cwd = options.cwd, production = options.production;
    var emitter = options.emitter || new events_1.EventEmitter();
    return find_1.findConfigFile(cwd)
        .then(function (path) {
        var cwd = path_1.dirname(path);
        return fs_1.readConfig(path)
            .then(function (config) {
            return transformBundles(config, { cwd: cwd, production: production, emitter: emitter });
        });
    });
}
exports.prune = prune;
function transformBundles(config, options) {
    var production = options.production;
    var resolutions = path_2.normalizeResolutions(config.resolution, options);
    var dependencies = extend(config.dependencies, config.peerDependencies, production ? {} : config.devDependencies);
    var globalDependencies = extend(config.globalDependencies, production ? {} : config.globalDevDependencies);
    return Promise.all(Object.keys(resolutions).map(function (type) {
        return transformBundle(resolutions[type], type, dependencies, globalDependencies, options);
    })).then(function () { return undefined; });
}
function transformBundle(path, resolution, dependencies, globalDependencies, options) {
    var emitter = options.emitter;
    var rmQueue = [];
    var bundle = path_2.getDefinitionPath(path);
    return fs_1.isFile(bundle)
        .then(function (exists) {
        if (!exists) {
            return;
        }
        return fs_1.transformDtsFile(bundle, function (typings) {
            var infos = typings.map(function (x) { return path_2.getInfoFromDependencyLocation(x, bundle); });
            var validPaths = [];
            for (var _i = 0, infos_1 = infos; _i < infos_1.length; _i++) {
                var _a = infos_1[_i], name = _a.name, global_1 = _a.global, location = _a.location;
                if (global_1) {
                    if (!globalDependencies.hasOwnProperty(name)) {
                        emitter.emit('prune', { name: name, global: global_1, resolution: resolution });
                        rmQueue.push(rmDependency({ name: name, global: global_1, path: path, emitter: emitter }));
                    }
                    else {
                        validPaths.push(location);
                    }
                }
                else {
                    if (!dependencies.hasOwnProperty(name)) {
                        emitter.emit('prune', { name: name, global: global_1, resolution: resolution });
                        rmQueue.push(rmDependency({ name: name, global: global_1, path: path, emitter: emitter }));
                    }
                    else {
                        validPaths.push(location);
                    }
                }
            }
            return validPaths;
        });
    })
        .then(function () { return Promise.all(rmQueue); })
        .then(function () { return undefined; });
}
function rmDependency(options) {
    var path = options.path, emitter = options.emitter;
    var _a = path_2.getDependencyPath(options), directory = _a.directory, definition = _a.definition, config = _a.config;
    function remove(path) {
        return fs_1.isFile(path)
            .then(function (exists) {
            if (!exists) {
                emitter.emit('enoent', { path: path });
                return;
            }
            return fs_1.unlink(path);
        });
    }
    return Promise.all([
        remove(config),
        remove(definition)
    ])
        .then(function () { return fs_1.rmdirUntil(directory, path); });
}
exports.rmDependency = rmDependency;
//# sourceMappingURL=prune.js.map