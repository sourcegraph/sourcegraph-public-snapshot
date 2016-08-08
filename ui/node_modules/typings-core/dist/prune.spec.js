"use strict";
var test = require('blue-tape');
var Promise = require('any-promise');
var path_1 = require('path');
var prune_1 = require('./prune');
var fs_1 = require('./utils/fs');
var BROWSER_GLOBAL_TYPINGS = 'typings/browser/globals/test/index.d.ts';
var BROWSER_TYPINGS = 'typings/browser/modules/test/index.d.ts';
var MAIN_GLOBAL_TYPINGS = 'typings/main/globals/test/index.d.ts';
var MAIN_TYPINGS = 'typings/main/modules/test/index.d.ts';
var BROWSER_INDEX = 'typings/browser/index.d.ts';
var MAIN_INDEX = 'typings/main/index.d.ts';
test('prune', function (t) {
    t.test('remove extraneous typings', function (t) {
        var FIXTURE_DIR = path_1.join(__dirname, '__test__/prune-extraneous');
        return generateTestDefinitions(FIXTURE_DIR)
            .then(function () {
            return prune_1.prune({ cwd: FIXTURE_DIR });
        })
            .then(function () {
            return Promise.all([
                fs_1.readFile(path_1.join(FIXTURE_DIR, BROWSER_INDEX), 'utf8'),
                fs_1.readFile(path_1.join(FIXTURE_DIR, MAIN_INDEX), 'utf8'),
                fs_1.isFile(path_1.join(FIXTURE_DIR, BROWSER_GLOBAL_TYPINGS)),
                fs_1.isFile(path_1.join(FIXTURE_DIR, BROWSER_TYPINGS)),
                fs_1.isFile(path_1.join(FIXTURE_DIR, MAIN_GLOBAL_TYPINGS)),
                fs_1.isFile(path_1.join(FIXTURE_DIR, MAIN_TYPINGS))
            ]);
        })
            .then(function (_a) {
            var browserDts = _a[0], mainDts = _a[1], hasBrowserGlobalDefinition = _a[2], hasBrowserDefinition = _a[3], hasMainGlobalDefinition = _a[4], hasMainDefinition = _a[5];
            t.equal(browserDts, [
                "/// <reference path=\"globals/test/index.d.ts\" />",
                "/// <reference path=\"modules/test/index.d.ts\" />",
                ""
            ].join('\n'));
            t.equal(mainDts, [
                "/// <reference path=\"globals/test/index.d.ts\" />",
                "/// <reference path=\"modules/test/index.d.ts\" />",
                ""
            ].join('\n'));
            t.equal(hasBrowserGlobalDefinition, true);
            t.equal(hasBrowserDefinition, true);
            t.equal(hasMainGlobalDefinition, true);
            t.equal(hasMainDefinition, true);
        });
    });
    t.test('remove all dev dependencies', function (t) {
        var FIXTURE_DIR = path_1.join(__dirname, '__test__/prune-production');
        return generateTestDefinitions(FIXTURE_DIR)
            .then(function () {
            return prune_1.prune({
                cwd: FIXTURE_DIR,
                production: true
            });
        })
            .then(function () {
            return Promise.all([
                fs_1.readFile(path_1.join(FIXTURE_DIR, BROWSER_INDEX), 'utf8'),
                fs_1.readFile(path_1.join(FIXTURE_DIR, MAIN_INDEX), 'utf8'),
                fs_1.isFile(path_1.join(FIXTURE_DIR, BROWSER_GLOBAL_TYPINGS)),
                fs_1.isFile(path_1.join(FIXTURE_DIR, BROWSER_TYPINGS)),
                fs_1.isFile(path_1.join(FIXTURE_DIR, MAIN_GLOBAL_TYPINGS)),
                fs_1.isFile(path_1.join(FIXTURE_DIR, MAIN_TYPINGS))
            ]);
        })
            .then(function (_a) {
            var browserDts = _a[0], mainDts = _a[1], hasBrowserGlobalDefinition = _a[2], hasBrowserDefinition = _a[3], hasMainGlobalDefinition = _a[4], hasMainDefinition = _a[5];
            t.equal(browserDts, "\n");
            t.equal(mainDts, "\n");
            t.equal(hasBrowserGlobalDefinition, false);
            t.equal(hasBrowserDefinition, false);
            t.equal(hasMainGlobalDefinition, false);
            t.equal(hasMainDefinition, false);
        });
    });
});
function generateTestDefinitions(directory) {
    var FAKE_GLOBAL_MODULE = "declare module 'test' {}\n";
    var FAKE_MODULE = "export function test (): boolean;\n";
    var dirs = [
        path_1.join(directory, 'typings/main/globals/test'),
        path_1.join(directory, 'typings/browser/globals/test'),
        path_1.join(directory, 'typings/main/modules/test'),
        path_1.join(directory, 'typings/browser/modules/test')
    ];
    return Promise.all(dirs.map(function (dir) { return fs_1.mkdirp(dir); }))
        .then(function () {
        return Promise.all([
            fs_1.writeFile(path_1.join(directory, BROWSER_GLOBAL_TYPINGS), FAKE_GLOBAL_MODULE),
            fs_1.writeFile(path_1.join(directory, BROWSER_TYPINGS), FAKE_MODULE),
            fs_1.writeFile(path_1.join(directory, MAIN_GLOBAL_TYPINGS), FAKE_GLOBAL_MODULE),
            fs_1.writeFile(path_1.join(directory, MAIN_TYPINGS), FAKE_MODULE),
            fs_1.writeFile(path_1.join(directory, BROWSER_INDEX), [
                "/// <reference path=\"globals/test/index.d.ts\" />",
                "/// <reference path=\"modules/test/index.d.ts\" />",
                "/// <reference path=\"modules/extraneous/index.d.ts\" />",
                ""
            ].join('\n')),
            fs_1.writeFile(path_1.join(directory, MAIN_INDEX), [
                "/// <reference path=\"globals/test/index.d.ts\" />",
                "/// <reference path=\"modules/test/index.d.ts\" />",
                "/// <reference path=\"modules/extraneous/index.d.ts\" />",
                ""
            ].join('\n'))
        ]);
    });
}
//# sourceMappingURL=prune.spec.js.map