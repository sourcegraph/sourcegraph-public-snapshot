"use strict";
var Promise = require('any-promise');
function compose(middleware) {
    if (!Array.isArray(middleware)) {
        throw new TypeError("Expected middleware to be an array, got " + typeof middleware);
    }
    for (var _i = 0, middleware_1 = middleware; _i < middleware_1.length; _i++) {
        var fn = middleware_1[_i];
        if (typeof fn !== 'function') {
            throw new TypeError("Expected middleware to contain functions, got " + typeof fn);
        }
    }
    return function () {
        var args = [];
        for (var _i = 0; _i < arguments.length; _i++) {
            args[_i - 0] = arguments[_i];
        }
        var index = -1;
        var done = args.pop();
        if (typeof done !== 'function') {
            throw new TypeError("Expected the last argument to be `next()`, got " + typeof done);
        }
        function dispatch(pos) {
            if (pos <= index) {
                throw new TypeError('`next()` called multiple times');
            }
            index = pos;
            var fn = middleware[pos] || done;
            return new Promise(function (resolve) {
                return resolve(fn.apply(void 0, args.concat([function next() {
                    return dispatch(pos + 1);
                }])));
            });
        }
        return dispatch(0);
    };
}
exports.compose = compose;
//# sourceMappingURL=index.js.map