"use strict";
var Promise = require('any-promise');
var path_1 = require('./utils/path');
var fs_1 = require('./utils/fs');
var parse_1 = require('./utils/parse');
var rc_1 = require('./utils/rc');
function viewEntry(raw, options) {
    return new Promise(function (resolve) {
        var meta = parse_1.parseDependency(parse_1.expandRegistry(raw)).meta;
        var path = "entries/" + encodeURIComponent(meta.source) + "/" + encodeURIComponent(meta.name);
        return resolve(fs_1.readJsonFrom(path_1.joinUrl(rc_1.default.registryURL, path)));
    });
}
exports.viewEntry = viewEntry;
function viewVersions(raw, options) {
    return new Promise(function (resolve) {
        var meta = parse_1.parseDependency(parse_1.expandRegistry(raw)).meta;
        var path = "entries/" + encodeURIComponent(meta.source) + "/" + encodeURIComponent(meta.name) + "/versions";
        if (meta.version) {
            path += "/" + encodeURIComponent(meta.version);
        }
        return resolve(fs_1.readJsonFrom(path_1.joinUrl(rc_1.default.registryURL, path)));
    });
}
exports.viewVersions = viewVersions;
//# sourceMappingURL=view.js.map