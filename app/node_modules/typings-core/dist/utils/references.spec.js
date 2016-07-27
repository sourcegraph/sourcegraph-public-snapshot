"use strict";
var test = require('blue-tape');
var path_1 = require('path');
var references = require('./references');
test('references', function (t) {
    t.test('parse references from string', function (t) {
        var file = "\n/// <reference path=\"foobar.d.ts\" />\n\n///\t<reference\t path=\"example.d.ts\"/>\n";
        var actual = references.extractReferences(file, __dirname);
        var expected = [
            {
                start: 1,
                end: 38,
                path: path_1.join(__dirname, 'foobar.d.ts')
            },
            {
                start: 39,
                end: 77,
                path: path_1.join(__dirname, 'example.d.ts')
            }
        ];
        t.deepEqual(actual, expected);
        t.end();
    });
    t.test('compile a path to reference string', function (t) {
        var actual = references.toReference('foobar.d.ts', __dirname);
        var expected = '/// <reference path="foobar.d.ts" />';
        t.equal(actual, expected);
        t.end();
    });
});
//# sourceMappingURL=references.spec.js.map