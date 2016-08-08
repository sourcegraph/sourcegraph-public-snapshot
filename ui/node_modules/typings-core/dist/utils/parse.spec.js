"use strict";
var test = require('blue-tape');
var path_1 = require('path');
var parse_1 = require('./parse');
var config_1 = require('./config');
test('parse', function (t) {
    t.test('parse dependency', function (t) {
        t.test('parse filename', function (t) {
            var actual = parse_1.parseDependency('file:./foo/bar.d.ts');
            var expected = {
                raw: 'file:./foo/bar.d.ts',
                location: path_1.normalize('foo/bar.d.ts'),
                meta: {
                    name: 'bar',
                    path: path_1.normalize('foo/bar.d.ts')
                },
                type: 'file'
            };
            t.deepEqual(actual, expected);
            t.end();
        });
        t.test('parse filename relative', function (t) {
            var actual = parse_1.parseDependency('file:foo/bar.d.ts');
            var expected = {
                raw: 'file:foo/bar.d.ts',
                location: path_1.normalize('foo/bar.d.ts'),
                meta: {
                    name: 'bar',
                    path: path_1.normalize('foo/bar.d.ts')
                },
                type: 'file'
            };
            t.deepEqual(actual, expected);
            t.end();
        });
        t.test('parse npm', function (t) {
            var actual = parse_1.parseDependency('npm:foobar');
            var expected = {
                raw: 'npm:foobar',
                type: 'npm',
                meta: {
                    name: 'foobar',
                    path: 'package.json'
                },
                location: path_1.normalize('foobar/package.json')
            };
            t.deepEqual(actual, expected);
            t.end();
        });
        t.test('parse scoped npm packages', function (t) {
            var actual = parse_1.parseDependency('npm:@foo/bar');
            var expected = {
                raw: 'npm:@foo/bar',
                type: 'npm',
                meta: {
                    name: '@foo/bar',
                    path: 'package.json'
                },
                location: path_1.normalize('@foo/bar/package.json')
            };
            t.deepEqual(actual, expected);
            t.end();
        });
        t.test('parse npm filename', function (t) {
            var actual = parse_1.parseDependency('npm:typescript/bin/lib.es6.d.ts');
            var expected = {
                raw: 'npm:typescript/bin/lib.es6.d.ts',
                type: 'npm',
                meta: {
                    name: 'typescript',
                    path: path_1.normalize('bin/lib.es6.d.ts')
                },
                location: path_1.normalize('typescript/bin/lib.es6.d.ts')
            };
            t.deepEqual(actual, expected);
            t.end();
        });
        t.test('parse bower', function (t) {
            var actual = parse_1.parseDependency('bower:foobar');
            var expected = {
                raw: 'bower:foobar',
                type: 'bower',
                meta: {
                    name: 'foobar',
                    path: 'bower.json'
                },
                location: path_1.normalize('foobar/bower.json')
            };
            t.deepEqual(actual, expected);
            t.end();
        });
        t.test('parse bower filename', function (t) {
            var actual = parse_1.parseDependency('bower:foobar/' + config_1.CONFIG_FILE);
            var expected = {
                raw: 'bower:foobar/' + config_1.CONFIG_FILE,
                type: 'bower',
                meta: {
                    name: 'foobar',
                    path: config_1.CONFIG_FILE
                },
                location: path_1.normalize('foobar/' + config_1.CONFIG_FILE)
            };
            t.deepEqual(actual, expected);
            t.end();
        });
        t.test('parse github', function (t) {
            var actual = parse_1.parseDependency('github:foo/bar');
            var expected = {
                raw: 'github:foo/bar',
                type: 'github',
                meta: {
                    name: undefined,
                    org: 'foo',
                    path: config_1.CONFIG_FILE,
                    repo: 'bar',
                    sha: 'master'
                },
                location: 'https://raw.githubusercontent.com/foo/bar/master/' + config_1.CONFIG_FILE
            };
            t.deepEqual(actual, expected);
            t.end();
        });
        t.test('parse github with sha and append config file', function (t) {
            var actual = parse_1.parseDependency('github:foo/bar#test');
            var expected = {
                raw: 'github:foo/bar#test',
                type: 'github',
                meta: {
                    name: undefined,
                    org: 'foo',
                    path: config_1.CONFIG_FILE,
                    repo: 'bar',
                    sha: 'test'
                },
                location: 'https://raw.githubusercontent.com/foo/bar/test/' + config_1.CONFIG_FILE
            };
            t.deepEqual(actual, expected);
            t.end();
        });
        t.test('parse github paths to `.d.ts` files', function (t) {
            var actual = parse_1.parseDependency('github:foo/bar/typings/file.d.ts');
            var expected = {
                raw: 'github:foo/bar/typings/file.d.ts',
                type: 'github',
                meta: {
                    name: 'file',
                    org: 'foo',
                    path: 'typings/file.d.ts',
                    repo: 'bar',
                    sha: 'master'
                },
                location: 'https://raw.githubusercontent.com/foo/bar/master/typings/file.d.ts'
            };
            t.deepEqual(actual, expected);
            t.end();
        });
        t.test('parse github paths to config file', function (t) {
            var actual = parse_1.parseDependency('github:foo/bar/src/' + config_1.CONFIG_FILE);
            var expected = {
                raw: 'github:foo/bar/src/' + config_1.CONFIG_FILE,
                type: 'github',
                meta: {
                    name: undefined,
                    org: 'foo',
                    path: "src/" + config_1.CONFIG_FILE,
                    repo: 'bar',
                    sha: 'master'
                },
                location: 'https://raw.githubusercontent.com/foo/bar/master/src/' + config_1.CONFIG_FILE
            };
            t.deepEqual(actual, expected);
            t.end();
        });
        t.test('parse bitbucket', function (t) {
            var actual = parse_1.parseDependency('bitbucket:foo/bar');
            var expected = {
                raw: 'bitbucket:foo/bar',
                type: 'bitbucket',
                meta: {
                    name: undefined,
                    org: 'foo',
                    path: config_1.CONFIG_FILE,
                    repo: 'bar',
                    sha: 'master'
                },
                location: 'https://bitbucket.org/foo/bar/raw/master/' + config_1.CONFIG_FILE
            };
            t.deepEqual(actual, expected);
            t.end();
        });
        t.test('parse bitbucket and append config file to path', function (t) {
            var actual = parse_1.parseDependency('bitbucket:foo/bar/dir');
            var expected = {
                raw: 'bitbucket:foo/bar/dir',
                type: 'bitbucket',
                meta: {
                    name: undefined,
                    org: 'foo',
                    path: "dir/" + config_1.CONFIG_FILE,
                    repo: 'bar',
                    sha: 'master'
                },
                location: 'https://bitbucket.org/foo/bar/raw/master/dir/' + config_1.CONFIG_FILE
            };
            t.deepEqual(actual, expected);
            t.end();
        });
        t.test('parse bitbucket with sha', function (t) {
            var actual = parse_1.parseDependency('bitbucket:foo/bar#abc');
            var expected = {
                raw: 'bitbucket:foo/bar#abc',
                type: 'bitbucket',
                meta: {
                    name: undefined,
                    org: 'foo',
                    path: config_1.CONFIG_FILE,
                    repo: 'bar',
                    sha: 'abc'
                },
                location: 'https://bitbucket.org/foo/bar/raw/abc/' + config_1.CONFIG_FILE
            };
            t.deepEqual(actual, expected);
            t.end();
        });
        t.test('parse url', function (t) {
            var actual = parse_1.parseDependency('http://example.com/foo/' + config_1.CONFIG_FILE);
            var expected = {
                raw: 'http://example.com/foo/' + config_1.CONFIG_FILE,
                type: 'http',
                meta: {},
                location: 'http://example.com/foo/' + config_1.CONFIG_FILE
            };
            t.deepEqual(actual, expected);
            t.end();
        });
        t.test('parse registry', function (t) {
            var actual = parse_1.parseDependency('registry:dt/node');
            var expected = {
                raw: 'registry:dt/node',
                type: 'registry',
                meta: { name: 'node', source: 'dt', tag: undefined, version: undefined },
                location: 'https://api.typings.org/entries/dt/node/versions/latest'
            };
            t.deepEqual(actual, expected);
            t.end();
        });
        t.test('parse registry with scoped package', function (t) {
            var actual = parse_1.parseDependency('registry:npm/@scoped/npm');
            var expected = {
                raw: 'registry:npm/@scoped/npm',
                type: 'registry',
                meta: { name: '@scoped/npm', source: 'npm', tag: undefined, version: undefined },
                location: 'https://api.typings.org/entries/npm/%40scoped%2Fnpm/versions/latest'
            };
            t.deepEqual(actual, expected);
            t.end();
        });
        t.test('parse registry with tag', function (t) {
            var actual = parse_1.parseDependency('registry:npm/dep#3.0.0-2016');
            var expected = {
                raw: 'registry:npm/dep#3.0.0-2016',
                type: 'registry',
                meta: { name: 'dep', source: 'npm', tag: '3.0.0-2016', version: undefined },
                location: 'https://api.typings.org/entries/npm/dep/tags/3.0.0-2016'
            };
            t.deepEqual(actual, expected);
            t.end();
        });
        t.test('parse registry with version', function (t) {
            var actual = parse_1.parseDependency('registry:npm/dep@^4.0');
            var expected = {
                raw: 'registry:npm/dep@^4.0',
                type: 'registry',
                meta: { name: 'dep', source: 'npm', tag: undefined, version: '^4.0' },
                location: 'https://api.typings.org/entries/npm/dep/versions/%5E4.0/latest'
            };
            t.deepEqual(actual, expected);
            t.end();
        });
        t.test('expand registry with default source', function (t) {
            var actual = parse_1.expandRegistry('domready');
            var expected = 'registry:npm/domready';
            t.deepEqual(actual, expected);
            t.end();
        });
        t.test('expand registry with provided source', function (t) {
            var actual = parse_1.expandRegistry('env~atom');
            var expected = 'registry:env/atom';
            t.deepEqual(actual, expected);
            t.end();
        });
        t.test('unknown scheme', function (t) {
            t.throws(function () { return parse_1.parseDependency('random:fake/dep'); }, /Unknown dependency: /);
            t.end();
        });
    });
    t.test('resolve dependency', function (t) {
        t.equal(parse_1.resolveDependency('github:foo/bar/baz/x.d.ts', '../lib/test.d.ts'), 'github:foo/bar/lib/test.d.ts');
        t.equal(parse_1.resolveDependency('http://example.com/foo/bar.d.ts', 'x.d.ts'), 'http://example.com/foo/x.d.ts');
        t.end();
    });
});
//# sourceMappingURL=parse.spec.js.map