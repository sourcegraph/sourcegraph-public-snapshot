"use strict";
var test = require('blue-tape');
var thenify = require('thenify');
var fs_1 = require('fs');
var path_1 = require('path');
var init_1 = require('./init');
var config_1 = require('./utils/config');
var fs_2 = require('./utils/fs');
test('init', function (t) {
    t.test('init an empty file', function (t) {
        var FIXTURE_DIR = path_1.join(__dirname, '__test__/init');
        var path = path_1.join(FIXTURE_DIR, config_1.CONFIG_FILE);
        return init_1.init({ cwd: FIXTURE_DIR })
            .then(function () {
            return fs_2.readJson(path);
        })
            .then(function (config) {
            t.ok(typeof config === 'object');
        })
            .then(function () {
            return thenify(fs_1.unlink)(path);
        });
    });
    t.test('upgrade from tsd', function (t) {
        var FIXTURE_DIR = path_1.join(__dirname, '__test__/init-upgrade');
        var path = path_1.join(FIXTURE_DIR, config_1.CONFIG_FILE);
        return init_1.init({ cwd: FIXTURE_DIR, upgrade: true })
            .then(function () {
            return fs_2.readJson(path);
        })
            .then(function (config) {
            t.deepEqual(config, {
                globalDependencies: {
                    codemirror: 'github:DefinitelyTyped/DefinitelyTyped/codemirror' +
                        '/codemirror.d.ts#01ce3ccf7f071514ff5057ef32a4550bf0b81dfe',
                    jquery: 'github:DefinitelyTyped/DefinitelyTyped/jquery' +
                        '/jquery.d.ts#01ce3ccf7f071514ff5057ef32a4550bf0b81dfe',
                    node: 'github:DefinitelyTyped/DefinitelyTyped/node/node.d.ts#3b2ed809b9e8f7dc4fcc1d80199129a0b73fb277'
                }
            });
        })
            .then(function () {
            return thenify(fs_1.unlink)(path);
        });
    });
    t.test('guess project name', function (t) {
        var FIXTURE_DIR = path_1.join(__dirname, '__test__/init-guess-name');
        var path = path_1.join(FIXTURE_DIR, config_1.CONFIG_FILE);
        return init_1.init({ cwd: FIXTURE_DIR })
            .then(function () {
            return fs_2.readJson(path);
        })
            .then(function (config) {
            t.equals(config.name, 'typings-test');
        })
            .then(function () {
            return thenify(fs_1.unlink)(path);
        });
    });
});
//# sourceMappingURL=init.spec.js.map