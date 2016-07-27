"use strict";
var extend = require('xtend');
var Promise = require('any-promise');
var path_1 = require('path');
var events_1 = require('events');
var dependencies_1 = require('./lib/dependencies');
var compile_1 = require('./lib/compile');
var find_1 = require('./utils/find');
var fs_1 = require('./utils/fs');
var path_2 = require('./utils/path');
var parse_1 = require('./utils/parse');
exports.parseDependencyExpression = parse_1.parseDependencyExpression;
exports.buildDependencyExpression = parse_1.buildDependencyExpression;
var config_1 = require('./utils/config');
function install(options) {
    var cwd = options.cwd, production = options.production;
    var emitter = options.emitter || new events_1.EventEmitter();
    return find_1.findConfigFile(cwd)
        .then(function (configFile) {
        var cwd = path_1.dirname(configFile);
        return fs_1.readConfig(configFile)
            .then(function (config) {
            var resolutions = path_2.normalizeResolutions(config.resolution, options);
            return dependencies_1.resolveTypeDependencies({
                cwd: cwd,
                emitter: emitter,
                global: true,
                peer: true,
                dev: !production
            })
                .then(function (tree) {
                var cwd = path_1.dirname(tree.src);
                var queue = [];
                function addToQueue(deps, global) {
                    for (var _i = 0, _a = Object.keys(deps); _i < _a.length; _i++) {
                        var name = _a[_i];
                        var tree_1 = deps[name];
                        queue.push(compile_1.compile(tree_1, Object.keys(resolutions), {
                            cwd: cwd,
                            name: name,
                            global: global,
                            emitter: emitter,
                            meta: true
                        }));
                    }
                }
                addToQueue(tree.dependencies, false);
                addToQueue(tree.devDependencies, false);
                addToQueue(tree.peerDependencies, false);
                addToQueue(tree.globalDependencies, true);
                addToQueue(tree.globalDevDependencies, true);
                return Promise.all(queue)
                    .then(function (results) {
                    return Promise.all(results.map(function (x) { return writeResult(x, { resolutions: resolutions }); }))
                        .then(function () { return writeBundle(results, { resolutions: resolutions }); })
                        .then(function () { return ({ tree: tree }); });
                });
            });
        });
    }, function () {
        emitter.emit('enoent', { path: path_1.join(cwd, config_1.CONFIG_FILE) });
        return { tree: extend(dependencies_1.DEFAULT_DEPENDENCY) };
    });
}
exports.install = install;
function installDependencyRaw(raw, options) {
    return installDependenciesRaw([raw], options).then(function (x) { return x[0]; });
}
exports.installDependencyRaw = installDependencyRaw;
function installDependenciesRaw(raw, options) {
    return new Promise(function (resolve) {
        var expressions = raw.map(function (x) { return parse_1.parseDependencyExpression(x, options); });
        return resolve(installDependencies(expressions, options));
    });
}
exports.installDependenciesRaw = installDependenciesRaw;
function installDependency(expression, options) {
    return installDependencies([expression], options).then(function (x) { return x[0]; });
}
exports.installDependency = installDependency;
function installDependencies(expressions, options) {
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
        return Promise.all(expressions.map(function (x) { return compileDependency(x, options); }))
            .then(function (results) {
            return Promise.all(results.map(function (x) { return writeResult(x, options); }))
                .then(function () { return writeBundle(results, options); })
                .then(function () { return writeToConfig(results, options); })
                .then(function () { return results.map(function (_a) {
                var name = _a.name, tree = _a.tree;
                return ({ name: name, tree: tree });
            }); });
        });
    });
}
exports.installDependencies = installDependencies;
function compileDependency(expression, options) {
    var dependency = parse_1.parseDependency(expression.location);
    var cwd = options.cwd, global = options.global, resolutions = options.resolutions;
    var emitter = options.emitter || new events_1.EventEmitter();
    var expName = expression.name || dependency.meta.name;
    return checkTypings(dependency, options)
        .then(function () {
        return dependencies_1.resolveDependency(dependency, { cwd: cwd, emitter: emitter, name: expName, dev: false, peer: false, global: false });
    })
        .then(function (tree) {
        var name = expName || tree.name;
        if (!name) {
            return Promise.reject(new TypeError("Unable to install dependency from \"" + tree.raw + "\" without a name"));
        }
        if (tree.postmessage) {
            emitter.emit('postmessage', { name: name, message: tree.postmessage });
        }
        return compile_1.compile(tree, Object.keys(resolutions), {
            cwd: cwd,
            name: name,
            global: global,
            emitter: emitter,
            meta: true
        });
    });
}
function writeToConfig(results, options) {
    if (options.save || options.saveDev || options.savePeer) {
        return fs_1.transformConfig(options.cwd, function (config) {
            for (var _i = 0, results_1 = results; _i < results_1.length; _i++) {
                var _a = results_1[_i], name = _a.name, tree = _a.tree;
                var raw = tree.raw;
                if (options.save) {
                    if (options.global) {
                        config.globalDependencies = extend(config.globalDependencies, (_b = {}, _b[name] = raw, _b));
                    }
                    else {
                        config.dependencies = extend(config.dependencies, (_c = {}, _c[name] = raw, _c));
                    }
                }
                else if (options.saveDev) {
                    if (options.global) {
                        config.globalDevDependencies = extend(config.globalDevDependencies, (_d = {}, _d[name] = raw, _d));
                    }
                    else {
                        config.devDependencies = extend(config.devDependencies, (_e = {}, _e[name] = raw, _e));
                    }
                }
                else if (options.savePeer) {
                    if (options.global) {
                        throw new TypeError('Unable to use `savePeer` with the `global` flag');
                    }
                    else {
                        config.peerDependencies = extend(config.peerDependencies, (_f = {}, _f[name] = raw, _f));
                    }
                }
            }
            return config;
            var _b, _c, _d, _e, _f;
        });
    }
    return Promise.resolve();
}
function writeBundle(results, options) {
    var resolutions = options.resolutions;
    return Promise.all(Object.keys(resolutions).map(function (resolution) {
        var path = resolutions[resolution];
        var paths = results.map(function (_a) {
            var name = _a.name, global = _a.global;
            return path_2.getDependencyPath({ path: path, name: name, global: global }).definition;
        });
        return fs_1.mkdirp(path)
            .then(function () {
            var bundle = path_2.getDefinitionPath(path);
            if (paths.length === 0) {
                return fs_1.touch(bundle);
            }
            return fs_1.transformDtsFile(bundle, function (x) { return x.concat(paths); });
        });
    }));
}
function writeResult(result, options) {
    var name = result.name, global = result.global, tree = result.tree, results = result.results;
    var resolutions = options.resolutions;
    return Promise.all(Object.keys(resolutions).map(function (resolution) {
        var path = resolutions[resolution];
        var contents = results[resolution];
        var _a = path_2.getDependencyPath({ name: name, global: global, path: path }), directory = _a.directory, config = _a.config, definition = _a.definition;
        return fs_1.mkdirp(directory)
            .then(function () {
            return Promise.all([
                fs_1.writeJson(config, { resolution: resolution, tree: fs_1.treeToJson(tree) }),
                fs_1.writeFile(definition, contents)
            ]);
        });
    }));
}
function checkTypings(dependency, options) {
    var type = dependency.type, meta = dependency.meta;
    if (type === 'registry' && meta.source === 'npm') {
        return find_1.findUp(options.cwd, path_1.join('node_modules', meta.name, 'package.json'))
            .then(function (path) {
            return fs_1.readJson(path)
                .then(function (packageJson) {
                if (packageJson && typeof packageJson.typings === 'string') {
                    options.emitter.emit('hastypings', {
                        name: meta.name,
                        source: meta.source,
                        path: path,
                        typings: path_2.resolveFrom(path, packageJson.typings)
                    });
                }
            });
        })
            .catch(function (err) { return undefined; });
    }
    return Promise.resolve();
}
//# sourceMappingURL=install.js.map