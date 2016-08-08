"use strict";
var test = require('blue-tape');
var path_1 = require('./path');
test('parse', function (t) {
    t.test('path from definition', function (t) {
        t.test('path', function (t) {
            t.equal(path_1.pathFromDefinition('foo/bar.d.ts'), 'foo/bar');
            t.end();
        });
        t.test('url', function (t) {
            t.equal(path_1.pathFromDefinition('http://example.com/test.d.ts'), '/test');
            t.end();
        });
    });
});
//# sourceMappingURL=path.spec.js.map