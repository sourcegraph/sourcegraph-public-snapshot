"use strict";
var test = require('blue-tape');
var Promise = require('any-promise');
var index_1 = require('./index');
test('async middleware', function (t) {
    t.test('middleware', function (t) {
        var arr = [];
        var fn = index_1.compose([
            function (req, res, next) {
                arr.push(1);
                return next().then(function (value) {
                    arr.push(5);
                    t.equal(value, 'propagate');
                    return 'done';
                });
            },
            function (req, res, next) {
                arr.push(2);
                return next().then(function (value) {
                    arr.push(4);
                    t.equal(value, 'hello');
                    return 'propagate';
                });
            }
        ]);
        return fn({}, {}, function () {
            arr.push(3);
            return 'hello';
        })
            .then(function () {
            t.deepEqual(arr, [1, 2, 3, 4, 5]);
        });
    });
    t.test('branch middleware by composing', function (t) {
        var arr = [];
        var fn = index_1.compose([
            index_1.compose([
                function (ctx, next) {
                    arr.push(1);
                    return next().catch(function (err) {
                        arr.push(3);
                    });
                },
                function (ctx, next) {
                    arr.push(2);
                    return Promise.reject(new Error('Boom!'));
                }
            ]),
            function (ctx, next) {
                arr.push(4);
                return next();
            }
        ]);
        return fn({}, function () { return undefined; })
            .then(function () {
            t.deepEqual(arr, [1, 2, 3]);
        });
    });
    t.test('throw when input is not an array', function (t) {
        t.throws(function () {
            index_1.compose('test');
        }, 'Expected middleware to be an array, got string');
        t.end();
    });
    t.test('throw when values are not functions', function (t) {
        t.throws(function () {
            index_1.compose([1, 2, 3]);
        }, 'Expected middleware to contain functions, got number');
        t.end();
    });
    t.test('throw when next is not a function', function (t) {
        var fn = index_1.compose([]);
        t.throws(function () {
            fn(true);
        }, 'Expected the last argument to be `next()`, got boolean');
        t.end();
    });
    t.test('throw when calling next() multiple times', function (t) {
        var fn = index_1.compose([
            function (value, next) {
                return next().then(next);
            }
        ]);
        t.plan(1);
        return fn({}, function () { return undefined; })
            .catch(function (err) {
            t.equal(err.message, '`next()` called multiple times');
        });
    });
});
//# sourceMappingURL=index.spec.js.map