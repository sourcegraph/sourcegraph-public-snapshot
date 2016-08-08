"use strict";
var test = require('blue-tape');
var makeErrorCause = require('./index');
test('make error cause', function (t) {
    var TestError = makeErrorCause('TestError');
    t.test('render the cause', function (t) {
        var cause = new Error('boom!');
        var error = new TestError('something bad', cause);
        var again = new TestError('more bad', error);
        t.equal(error.cause, cause);
        t.equal(error.toString(), 'TestError: something bad\nCaused by: Error: boom!');
        t.equal(again.cause, error);
        t.equal(again.toString(), 'TestError: more bad\nCaused by: TestError: something bad\nCaused by: Error: boom!');
        t.end();
    });
});
//# sourceMappingURL=index.spec.js.map