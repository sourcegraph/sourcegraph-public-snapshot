"use strict";
function __export(m) {
    for (var p in m) if (!exports.hasOwnProperty(p)) exports[p] = m[p];
}
var FormData = require('form-data');
var Promise = require('any-promise');
var common_1 = require('./common');
__export(require('./common'));
function headers() {
    var common = common_1.headers();
    return function (request, next) {
        return common(request, function () {
            if (!request.get('User-Agent')) {
                request.set('User-Agent', 'https://github.com/blakeembrey/popsicle');
            }
            if (request.body instanceof FormData) {
                request.set('Content-Type', 'multipart/form-data; boundary=' + request.body.getBoundary());
                return new Promise(function (resolve, reject) {
                    request.body.getLength(function (err, length) {
                        if (err) {
                            request.set('Transfer-Encoding', 'chunked');
                        }
                        else {
                            request.set('Content-Length', String(length));
                        }
                        return resolve(next());
                    });
                });
            }
            var length = 0;
            var body = request.body;
            if (body && !request.get('Content-Length')) {
                if (Array.isArray(body)) {
                    for (var i = 0; i < body.length; i++) {
                        length += body[i].length;
                    }
                }
                else if (typeof body === 'string') {
                    length = Buffer.byteLength(body);
                }
                else {
                    length = body.length;
                }
                if (length) {
                    request.set('Content-Length', String(length));
                }
                else if (typeof body.pipe === 'function') {
                    request.set('Transfer-Encoding', 'chunked');
                }
                else {
                    return Promise.reject(request.error('Argument error, `options.body`', 'EBODY'));
                }
            }
            return next();
        });
    };
}
exports.headers = headers;
//# sourceMappingURL=index.js.map