"use strict";
var extend = require('xtend');
var Promise = require('any-promise');
var popsicle_1 = require('popsicle');
function popsicleRetry(retries) {
    if (retries === void 0) { retries = popsicleRetry.retries(); }
    var iter = 0;
    return function retry(request, next) {
        function attempt(error, response, result) {
            var delay = retries(error, response, ++iter);
            if (delay <= 0) {
                return result;
            }
            return new Promise(function (resolve) {
                setTimeout(function () {
                    var options = extend(request.toOptions(), {
                        use: request.middleware.slice(request.middleware.indexOf(retry))
                    });
                    return resolve(new popsicle_1.Request(options));
                }, delay);
            });
        }
        return next()
            .then(function (response) {
            return attempt(null, response, Promise.resolve(response));
        }, function (error) {
            return attempt(error, null, Promise.reject(error));
        });
    };
}
var popsicleRetry;
(function (popsicleRetry) {
    function retryAllowed(error, response) {
        if (error) {
            if (error.code === 'EUNAVAILABLE') {
                if (process.browser) {
                    return navigator.onLine !== false;
                }
                var code = error.cause.code;
                return (code === 'ECONNREFUSED' ||
                    code === 'ECONNRESET' ||
                    code === 'ETIMEDOUT' ||
                    code === 'EPIPE');
            }
            return false;
        }
        if (response) {
            return response.statusType() === 5;
        }
        return false;
    }
    popsicleRetry.retryAllowed = retryAllowed;
    function retries(count, isRetryAllowed) {
        if (count === void 0) { count = 5; }
        if (isRetryAllowed === void 0) { isRetryAllowed = retryAllowed; }
        return function (error, response, iter) {
            if (iter > count || !isRetryAllowed(error, response)) {
                return -1;
            }
            var noise = Math.random() * 100;
            return (1 << iter) * 1000 + noise;
        };
    }
    popsicleRetry.retries = retries;
})(popsicleRetry || (popsicleRetry = {}));
module.exports = popsicleRetry;
//# sourceMappingURL=index.js.map