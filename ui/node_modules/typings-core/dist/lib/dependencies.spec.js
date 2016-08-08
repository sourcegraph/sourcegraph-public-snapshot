"use strict";
var test = require('blue-tape');
var path_1 = require('path');
var events_1 = require('events');
var dependencies_1 = require('./dependencies');
var RESOLVE_FIXTURE_DIR = path_1.join(__dirname, '__test__/fixtures/resolve');
var emitter = new events_1.EventEmitter();
test('dependencies', function (t) {
    t.test('resolve fixture', function (t) {
        t.test('resolve a dependency tree', function (t) {
            var expected = {
                raw: undefined,
                global: false,
                postmessage: undefined,
                name: 'foobar',
                src: path_1.join(RESOLVE_FIXTURE_DIR, 'typings.json'),
                main: 'foo.d.ts',
                files: undefined,
                version: undefined,
                browser: undefined,
                typings: undefined,
                browserTypings: undefined,
                dependencies: {},
                devDependencies: {},
                peerDependencies: {},
                globalDependencies: {},
                globalDevDependencies: {}
            };
            var bowerDep = {
                raw: 'bower:bower-dep',
                global: false,
                postmessage: undefined,
                src: path_1.join(RESOLVE_FIXTURE_DIR, 'bower_components/bower-dep/bower.json'),
                typings: 'bower-dep.d.ts',
                browserTypings: undefined,
                dependencies: {},
                devDependencies: {},
                peerDependencies: {},
                globalDependencies: {},
                globalDevDependencies: {},
                name: 'bower-dep',
                files: undefined,
                version: undefined,
                main: 'index.js',
                browser: undefined
            };
            var exampleDep = {
                raw: 'bower:example',
                global: false,
                postmessage: undefined,
                src: path_1.join(RESOLVE_FIXTURE_DIR, 'bower_components/example/bower.json'),
                main: undefined,
                browser: undefined,
                files: undefined,
                version: undefined,
                typings: undefined,
                browserTypings: undefined,
                name: 'example',
                dependencies: {},
                devDependencies: {},
                peerDependencies: {},
                globalDependencies: {},
                globalDevDependencies: {}
            };
            var typedDep = {
                raw: 'file:typings/dep.d.ts',
                global: undefined,
                postmessage: undefined,
                src: path_1.join(RESOLVE_FIXTURE_DIR, 'typings/dep.d.ts'),
                typings: path_1.join(RESOLVE_FIXTURE_DIR, 'typings/dep.d.ts'),
                main: undefined,
                browser: undefined,
                files: undefined,
                version: undefined,
                browserTypings: undefined,
                dependencies: {},
                devDependencies: {},
                peerDependencies: {},
                globalDependencies: {},
                globalDevDependencies: {}
            };
            var npmDep = {
                raw: 'npm:npm-dep',
                global: false,
                postmessage: undefined,
                src: path_1.join(RESOLVE_FIXTURE_DIR, 'node_modules/npm-dep/package.json'),
                main: './index.js',
                browser: undefined,
                files: undefined,
                version: undefined,
                typings: undefined,
                browserTypings: undefined,
                name: 'npm-dep',
                dependencies: {},
                devDependencies: {},
                peerDependencies: {},
                globalDependencies: {},
                globalDevDependencies: {}
            };
            var typedDevDep = {
                globalDependencies: {},
                globalDevDependencies: {},
                browser: undefined,
                browserTypings: undefined,
                dependencies: {},
                devDependencies: {},
                peerDependencies: {},
                main: undefined,
                name: 'dep',
                raw: 'bower:dep',
                global: false,
                postmessage: undefined,
                src: path_1.join(RESOLVE_FIXTURE_DIR, 'bower_components/dep/bower.json'),
                typings: undefined,
                files: undefined,
                version: undefined
            };
            expected.dependencies['bower-dep'] = bowerDep;
            expected.dependencies.dep = typedDep;
            expected.dependencies['npm-dep'] = npmDep;
            expected.devDependencies['dev-dep'] = typedDevDep;
            bowerDep.dependencies.example = exampleDep;
            return dependencies_1.resolveAllDependencies({
                cwd: RESOLVE_FIXTURE_DIR,
                dev: true,
                emitter: emitter
            })
                .then(function (result) {
                function removeParentReferenceFromDependencies(dependencies) {
                    Object.keys(dependencies).forEach(function (key) {
                        removeParentReference(dependencies[key]);
                    });
                }
                function removeParentReference(tree) {
                    delete tree.parent;
                    removeParentReferenceFromDependencies(tree.dependencies);
                    removeParentReferenceFromDependencies(tree.devDependencies);
                    removeParentReferenceFromDependencies(tree.peerDependencies);
                    removeParentReferenceFromDependencies(tree.globalDependencies);
                    removeParentReferenceFromDependencies(tree.globalDevDependencies);
                    return tree;
                }
                t.equal(result.parent, undefined);
                t.ok(result.dependencies.dep.parent != null);
                t.ok(result.dependencies['npm-dep'].parent != null);
                removeParentReference(result);
                t.deepEqual(result, expected);
            });
        });
    });
});
//# sourceMappingURL=dependencies.spec.js.map