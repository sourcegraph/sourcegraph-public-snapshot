"use strict";
var path_1 = require('path');
var url_1 = require('url');
var isAbsolute = require('is-absolute');
var config_1 = require('./config');
exports.EOL = '\n';
function isHttp(url) {
    return /^https?\:\/\//i.test(url);
}
exports.isHttp = isHttp;
function isDefinition(path) {
    if (isHttp(path)) {
        return isDefinition(url_1.parse(path).pathname);
    }
    return /\.d\.ts$/.test(path);
}
exports.isDefinition = isDefinition;
function isModuleName(value) {
    return !isHttp(value) && !isAbsolute(value) && value.charAt(0) !== '.';
}
exports.isModuleName = isModuleName;
function normalizeSlashes(path) {
    return path.replace(/\\/g, '/');
}
exports.normalizeSlashes = normalizeSlashes;
function joinUrl(from, to) {
    return from.replace(/\/$/, '') + "/" + to.replace(/^\//, '');
}
exports.joinUrl = joinUrl;
function resolveFrom(from, to) {
    if (isHttp(to)) {
        return to;
    }
    if (isHttp(from)) {
        var url = url_1.parse(from);
        url.pathname = url_1.resolve(url.pathname, to);
        return url_1.format(url);
    }
    return path_1.resolve(path_1.dirname(from), to);
}
exports.resolveFrom = resolveFrom;
function relativeTo(from, to) {
    if (isHttp(from)) {
        var fromUrl = url_1.parse(from);
        if (isHttp(to)) {
            var toUrl = url_1.parse(to);
            if (toUrl.auth !== fromUrl.auth || toUrl.host !== fromUrl.host) {
                return to;
            }
            var relativeUrl = relativeTo(fromUrl.pathname, toUrl.pathname);
            if (toUrl.search) {
                relativeUrl += toUrl.search;
            }
            if (toUrl.hash) {
                relativeUrl += toUrl.hash;
            }
            return relativeUrl;
        }
        return relativeTo(fromUrl.pathname, to);
    }
    return path_1.relative(path_1.dirname(from), to);
}
exports.relativeTo = relativeTo;
function toDefinition(path) {
    if (isHttp(path)) {
        var url = url_1.parse(path);
        url.pathname = toDefinition(url.pathname);
        return url_1.format(url);
    }
    return path + ".d.ts";
}
exports.toDefinition = toDefinition;
function pathFromDefinition(path) {
    if (isHttp(path)) {
        return pathFromDefinition(url_1.parse(path).pathname);
    }
    return path.replace(/\.d\.ts$/, '');
}
exports.pathFromDefinition = pathFromDefinition;
function normalizeToDefinition(path) {
    if (isDefinition(path)) {
        return path;
    }
    if (isHttp(path)) {
        var url = url_1.parse(path);
        url.pathname = normalizeToDefinition(path);
        return url_1.format(url);
    }
    var ext = path_1.extname(path);
    return toDefinition(ext ? path.slice(0, -ext.length) : path);
}
exports.normalizeToDefinition = normalizeToDefinition;
function getDefinitionPath(path) {
    return path_1.join(path, 'index.d.ts');
}
exports.getDefinitionPath = getDefinitionPath;
function getDependencyPath(options) {
    var type = options.global ? 'globals' : 'modules';
    var directory = path_1.join(options.path, type, options.name);
    var definition = getDefinitionPath(directory);
    var config = path_1.join(directory, 'typings.json');
    return { directory: directory, definition: definition, config: config };
}
exports.getDependencyPath = getDependencyPath;
function getInfoFromDependencyLocation(location, bundle) {
    var parts = relativeTo(bundle, location).split(path_1.sep);
    return {
        location: location,
        global: parts[0] === 'globals',
        name: parts.slice(1, -1).join('/')
    };
}
exports.getInfoFromDependencyLocation = getInfoFromDependencyLocation;
function detectEOL(contents) {
    var match = contents.match(/\r\n|\r|\n/);
    return match ? match[0] : undefined;
}
exports.detectEOL = detectEOL;
function normalizeEOL(contents, eol) {
    return contents.replace(/\r\n|\r|\n/g, eol);
}
exports.normalizeEOL = normalizeEOL;
function normalizeResolutions(resolutions, options) {
    var resolutionMap = {};
    if (typeof resolutions === 'object') {
        for (var _i = 0, _a = Object.keys(resolutions); _i < _a.length; _i++) {
            var type = _a[_i];
            resolutionMap[type] = path_1.join(options.cwd, resolutions[type]);
        }
    }
    else if (typeof resolutions === 'string') {
        resolutionMap.main = path_1.join(options.cwd, resolutions);
    }
    else {
        resolutionMap.main = path_1.join(options.cwd, config_1.DEFAULT_TYPINGS_DIR);
    }
    return resolutionMap;
}
exports.normalizeResolutions = normalizeResolutions;
//# sourceMappingURL=path.js.map