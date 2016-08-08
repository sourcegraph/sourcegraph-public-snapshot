"use strict";
var test = require('blue-tape');
var typings_1 = require('./typings');
var pkg = require('../package.json');
test('typings', function (t) {
    t.test('version', function (t) {
        t.equal(typings_1.VERSION, pkg.version);
        t.end();
    });
});
//# sourceMappingURL=typings.spec.js.map