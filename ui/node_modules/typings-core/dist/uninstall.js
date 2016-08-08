"use strict";
var extend = require('xtend');
var Promise = require('any-promise');
var events_1 = require('events');
var path_1 = require('path');
var fs_1 = require('./utils/fs');
var find_1 = require('./utils/find');
var path_2 = require('./utils/path');
function uninstallDependency(name, options) {
    return uninstallDependencies([name], options);
}
exports.uninstallDependency = uninstallDependency;
function uninstallDependencies(names, options) {
    var emitter = options.emitter || new events_1.EventEmitter();
    return find_1.findConfigFile(options.cwd)
        .then(function (configFile) {
        var cwd = path_1.dirname(configFile);
        return fs_1.readConfig(configFile)
            .then(function (config) {
            var resolutions = path_2.normalizeResolutions(config.resolution, options);
            return extend(options, { resolutions: resolutions, cwd: cwd, emitter: emitter });
        });
    }, function () {
        var resolutions = path_2.normalizeResolutions(undefined, options);
        return extend(options, { emitter: emitter, resolutions: resolutions });
    })
        .then(function (options) {
        return Promise.all(names.map(function (x) { return uninstallFrom(x, options); }))
            .then(function () { return writeBundle(names, options); })
            .then(function () { return writeToConfig(names, options); })
            .then(function () { return undefined; });
    });
}
exports.uninstallDependencies = uninstallDependencies;
function uninstallFrom(name, options) {
    var cwd = options.cwd, global = options.global, emitter = options.emitter, resolutions = options.resolutions;
    return Promise.all(Object.keys(resolutions).map(function (type) {
        var path = resolutions[type];
        var _a = path_2.getDependencyPath({ path: path, name: name, global: global }), directory = _a.directory, definition = _a.definition, config = _a.config;
        return fs_1.isFile(definition)
            .then(function (exists) {
            if (!exists) {
                emitter.emit('enoent', { path: definition });
                return;
            }
            return Promise.all([
                fs_1.unlink(definition),
                fs_1.unlink(config)
            ])
                .then(function () { return fs_1.rmdirUntil(directory, cwd); });
        });
    }));
}
function writeToConfig(names, options) {
    if (options.save || options.saveDev || options.savePeer) {
        return fs_1.transformConfig(options.cwd, function (config) {
            for (var _i = 0, names_1 = names; _i < names_1.length; _i++) {
                var name = names_1[_i];
                if (options.save) {
                    if (options.global) {
                        if (config.globalDependencies && config.globalDependencies[name]) {
                            delete config.globalDependencies[name];
                        }
                        else {
                            return Promise.reject(new TypeError("Typings for \"" + name + "\" are not listed in global dependencies"));
                        }
                    }
                    else {
                        if (config.dependencies && config.dependencies[name]) {
                            delete config.dependencies[name];
                        }
                        else {
                            return Promise.reject(new TypeError("Typings for \"" + name + "\" are not listed in dependencies"));
                        }
                    }
                }
                if (options.saveDev) {
                    if (options.global) {
                        if (config.globalDevDependencies && config.globalDevDependencies[name]) {
                            delete config.globalDevDependencies[name];
                        }
                        else {
                            return Promise.reject(new TypeError("Typings for \"" + name + "\" are not listed in global dev dependencies"));
                        }
                    }
                    else {
                        if (config.devDependencies && config.devDependencies[name]) {
                            delete config.devDependencies[name];
                        }
                        else {
                            return Promise.reject(new TypeError("Typings for \"" + name + "\" are not listed in dev dependencies"));
                        }
                    }
                }
                if (options.savePeer) {
                    if (config.peerDependencies && config.peerDependencies[name]) {
                        delete config.peerDependencies[name];
                    }
                    else {
                        return Promise.reject(new TypeError("Typings for \"" + name + "\" are not listed in peer dependencies"));
                    }
                }
            }
            return config;
        });
    }
}
function writeBundle(names, options) {
    var global = options.global, resolutions = options.resolutions;
    return Promise.all(Object.keys(resolutions).map(function (type) {
        var path = resolutions[type];
        var bundle = path_2.getDefinitionPath(path);
        var paths = names.map(function (name) { return path_2.getDependencyPath({ path: path, name: name, global: global }).definition; });
        return fs_1.transformDtsFile(bundle, function (references) {
            return references.filter(function (x) { return paths.indexOf(x) === -1; });
        });
    }));
}
//# sourceMappingURL=uninstall.js.map