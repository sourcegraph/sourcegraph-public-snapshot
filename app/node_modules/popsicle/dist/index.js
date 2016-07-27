"use strict";
var http_1 = require('http');
var https_1 = require('https');
var stream_1 = require('stream');
var urlLib = require('url');
var arrify = require('arrify');
var concat = require('concat-stream');
var Promise = require('any-promise');
var zlib_1 = require('zlib');
var response_1 = require('./response');
var index_1 = require('./plugins/index');
var validTypes = ['text', 'buffer', 'array', 'uint8array', 'stream'];
function createTransport(options) {
    return {
        use: use,
        abort: abort,
        open: function (request) {
            return handle(request, options);
        }
    };
}
exports.createTransport = createTransport;
var use = [index_1.stringify(), index_1.headers()];
var REDIRECT_TYPE;
(function (REDIRECT_TYPE) {
    REDIRECT_TYPE[REDIRECT_TYPE["FOLLOW_WITH_GET"] = 0] = "FOLLOW_WITH_GET";
    REDIRECT_TYPE[REDIRECT_TYPE["FOLLOW_WITH_CONFIRMATION"] = 1] = "FOLLOW_WITH_CONFIRMATION";
})(REDIRECT_TYPE || (REDIRECT_TYPE = {}));
var REDIRECT_STATUS = {
    '300': REDIRECT_TYPE.FOLLOW_WITH_GET,
    '301': REDIRECT_TYPE.FOLLOW_WITH_GET,
    '302': REDIRECT_TYPE.FOLLOW_WITH_GET,
    '303': REDIRECT_TYPE.FOLLOW_WITH_GET,
    '305': REDIRECT_TYPE.FOLLOW_WITH_GET,
    '307': REDIRECT_TYPE.FOLLOW_WITH_CONFIRMATION,
    '308': REDIRECT_TYPE.FOLLOW_WITH_CONFIRMATION
};
function handle(request, options) {
    var followRedirects = options.followRedirects, type = options.type, unzip = options.unzip, rejectUnauthorized = options.rejectUnauthorized, ca = options.ca, key = options.key, cert = options.cert, agent = options.agent;
    var url = request.url, method = request.method, body = request.body;
    var maxRedirects = num(options.maxRedirects, 5);
    var maxBufferSize = num(options.maxBufferSize, type === 'stream' ? Infinity : 2 * 1000 * 1000);
    var storeCookies = getStoreCookies(request, options);
    var attachCookies = getAttachCookies(request, options);
    var confirmRedirect = options.confirmRedirect || falsey;
    var requestCount = 0;
    if (type && validTypes.indexOf(type) === -1) {
        return Promise.reject(request.error("Unsupported type: " + type, 'ETYPE'));
    }
    if (unzip !== false && request.get('Accept-Encoding') == null) {
        request.set('Accept-Encoding', 'gzip,deflate');
    }
    function get(url, method, body) {
        if (requestCount++ > maxRedirects) {
            return Promise.reject(request.error("Exceeded maximum of " + maxRedirects + " redirects", 'EMAXREDIRECTS'));
        }
        return attachCookies(url)
            .then(function () {
            return new Promise(function (resolve, reject) {
                var arg = urlLib.parse(url);
                var isHttp = arg.protocol !== 'https:';
                var engine = isHttp ? http_1.request : https_1.request;
                arg.method = method;
                arg.headers = request.toHeaders();
                arg.agent = agent;
                arg.rejectUnauthorized = rejectUnauthorized !== false;
                arg.ca = ca;
                arg.cert = cert;
                arg.key = key;
                var rawRequest = engine(arg);
                var requestStream = new stream_1.PassThrough();
                var responseStream = new stream_1.PassThrough();
                var uploadedBytes = 0;
                var downloadedBytes = 0;
                requestStream.on('data', function (chunk) {
                    uploadedBytes += chunk.length;
                    request._setUploadedBytes(uploadedBytes);
                });
                requestStream.on('end', function () {
                    request._setUploadedBytes(uploadedBytes, 1);
                });
                responseStream.on('data', function (chunk) {
                    downloadedBytes += chunk.length;
                    request._setDownloadedBytes(downloadedBytes);
                    if (downloadedBytes > maxBufferSize) {
                        rawRequest.abort();
                        responseStream.emit('error', request.error('Response too large', 'ETOOLARGE'));
                    }
                });
                responseStream.on('end', function () {
                    request._setDownloadedBytes(downloadedBytes, 1);
                });
                function response(incomingMessage) {
                    var headers = incomingMessage.headers, rawHeaders = incomingMessage.rawHeaders, status = incomingMessage.statusCode, statusText = incomingMessage.statusMessage;
                    var redirect = REDIRECT_STATUS[status];
                    if (followRedirects !== false && redirect != null && headers.location) {
                        var newUrl = urlLib.resolve(url, headers.location);
                        incomingMessage.resume();
                        if (redirect === REDIRECT_TYPE.FOLLOW_WITH_GET) {
                            request.set('Content-Length', '0');
                            return get(newUrl, 'GET');
                        }
                        if (redirect === REDIRECT_TYPE.FOLLOW_WITH_CONFIRMATION) {
                            if (arg.method === 'GET' || arg.method === 'HEAD') {
                                return get(newUrl, method, body);
                            }
                            if (confirmRedirect(rawRequest, incomingMessage)) {
                                return get(newUrl, method, body);
                            }
                        }
                    }
                    request.downloadLength = num(headers['content-length'], null);
                    incomingMessage.pipe(responseStream);
                    return handleResponse(request, responseStream, headers, options)
                        .then(function (body) {
                        return new response_1.default({
                            status: status,
                            headers: headers,
                            statusText: statusText,
                            rawHeaders: rawHeaders,
                            body: body,
                            url: url
                        });
                    });
                }
                function emitError(error) {
                    rawRequest.abort();
                    reject(error);
                }
                rawRequest.on('response', function (message) {
                    resolve(storeCookies(url, message.headers).then(function () { return response(message); }));
                });
                rawRequest.on('error', function (error) {
                    emitError(request.error("Unable to connect to \"" + url + "\"", 'EUNAVAILABLE', error));
                });
                request._raw = rawRequest;
                request.uploadLength = num(rawRequest.getHeader('content-length'), null);
                requestStream.pipe(rawRequest);
                requestStream.on('error', emitError);
                if (body) {
                    if (typeof body.pipe === 'function') {
                        body.pipe(requestStream);
                        body.on('error', emitError);
                    }
                    else {
                        requestStream.end(body);
                    }
                }
                else {
                    requestStream.end();
                }
            });
        });
    }
    return get(url, method, body);
}
function abort(request) {
    request._raw.abort();
}
function num(value, fallback) {
    if (value == null) {
        return fallback;
    }
    return isNaN(value) ? fallback : Number(value);
}
function falsey() {
    return false;
}
function getAttachCookies(request, options) {
    var jar = options.jar;
    var cookie = request.getAll('Cookie');
    if (!jar) {
        return function () { return Promise.resolve(); };
    }
    return function (url) {
        return new Promise(function (resolve, reject) {
            request.set('Cookie', cookie);
            options.jar.getCookies(url, function (err, cookies) {
                if (err) {
                    return reject(err);
                }
                if (cookies.length) {
                    request.append('Cookie', cookies.join('; '));
                }
                return resolve();
            });
        });
    };
}
function getStoreCookies(request, options) {
    var jar = options.jar;
    if (!jar) {
        return function () { return Promise.resolve(); };
    }
    return function (url, headers) {
        var cookies = arrify(headers['set-cookie']);
        if (!cookies.length) {
            return Promise.resolve();
        }
        var storeCookies = cookies.map(function (cookie) {
            return new Promise(function (resolve, reject) {
                jar.setCookie(cookie, url, { ignoreError: true }, function (err) {
                    return err ? reject(err) : resolve();
                });
            });
        });
        return Promise.all(storeCookies);
    };
}
function handleResponse(request, stream, headers, options) {
    var type = options.type || 'text';
    var unzip = options.unzip !== false;
    var result = new Promise(function (resolve, reject) {
        if (unzip) {
            var enc = headers['content-encoding'];
            if (enc === 'deflate' || enc === 'gzip') {
                var unzip_1 = zlib_1.createUnzip();
                stream.pipe(unzip_1);
                stream.on('error', function (err) { return unzip_1.emit('error', err); });
                stream = unzip_1;
            }
        }
        if (type === 'stream') {
            return resolve(stream);
        }
        var encoding = type === 'text' ? 'string' : type;
        var concatStream = concat({ encoding: encoding }, resolve);
        stream.on('error', reject);
        stream.pipe(concatStream);
    });
    return result;
}
//# sourceMappingURL=index.js.map